# Replay das partidas — manual do projeto

Este é o manual prático do plano 05. Ele explica, em linguagem simples, o que
mudou no jogo, como uma partida vira um arquivo `.hwr`, como as versões do motor
são mantidas e o que fazer quando algo diverge. Todos os caminhos abaixo são
relativos a `server/replay/`.

A ideia central: **um replay não é um vídeo**. O servidor guarda a semente do
sorteio e a lista de entradas (jogadas, ticks, conexões). Quem assiste roda o
mesmo motor, da mesma versão, com a mesma semente, e chega ao mesmo estado.
É por isso que uma partida inteira cabe em algumas centenas de bytes.

Mapa rápido do código:

| Peça | Onde está |
| --- | --- |
| Arquivo `.hwr` (escrever/ler) | `format/format.go` |
| Reprodução em Go | `runner/runner.go` |
| Motores congelados por versão | `engines/vN/` e `engines/registry.go` |
| Lista de versões (sem importar os motores) | `versions/versions.go` |
| Ponte com o navegador | `cmd/wasm/main.go` e `cmd/wasm/session.go` |
| Teste de 200 partidas + golden | `e2e/e2e_test.go` e `engines/v1/testdata/` |
| Gravação no jogo | `../gameServer/redisFuncs/withRoom.go`, `replay.go` e `../gameServer/services/replayRecorder.go` |
| Assinatura e armazenamento | `../backend/services/replayService.go` e `../backend/routes/handlers/replaysHandler.go` |
| Loader, worker e tela | `../../client/web/src/replay/` e `../../client/web/src/features/replay/` |

## 1. O que mudou no motor e por quê

Antes, o motor era uma função do mundo: chamava Redis, mandava mensagem pelo
WebSocket e lia `time.Now` e o `math/rand` global. Duas execuções com o mesmo
estado podiam terminar diferentes — o segundo sorteio já não era o mesmo. Para
reproduzir uma partida, o motor precisava virar uma função de **estado +
entradas**.

O que mudou:

- **`Room` virou o estado completo da partida.** Em
  [`../gameServer/game/structs/rooms/room.go`](../gameServer/game/structs/rooms/room.go)
  ela guarda, além do que já existia, `Seed` (semente de 128 bits), `Rng`
  (estado do PCG), `Connected` (quem está conectado), `NextMoveTime` (ordem de
  turno que nunca repete) e `NonReplayable`. O relógio fixo da chamada fica em
  um campo privado, preenchido pelo `WithRoom`.
- **Semente, sorteio e relógio ficaram determinísticos.** Em
  [`../gameServer/game/structs/rooms/determinism.go`](../gameServer/game/structs/rooms/determinism.go)
  estão `NewSeed` (usa `crypto/rand`), o PCG com `MarshalBinary` no JSON da sala,
  `Room.Now()`/`Room.SetNow()` e `Room.Rand()`. O tipo `Seed` esconde o valor em
  `String`, `GoString` e `LogValue`, então a semente não aparece em log nem em
  erro.
- **O motor deixou de conversar com o mundo.** A interface `Sink`
  ([`../gameServer/game/structs/rooms/sink.go`](../gameServer/game/structs/rooms/sink.go))
  só tem `Broadcast`, `RegistryTimeout` e `ClearTimeout`. Em produção quem
  implementa é o `RedisSink`
  ([`../gameServer/redisFuncs/persistence.go`](../gameServer/redisFuncs/persistence.go));
  no replay e nos testes é um no-op. Salvar não é mais papel do motor.
- **`WithRoom` virou a fronteira de toda mudança.** Em
  [`../gameServer/redisFuncs/withRoom.go`](../gameServer/redisFuncs/withRoom.go)
  ele pega o mutex e o lock distribuído, carrega a sala do Redis, captura o
  relógio **uma vez**, preenche `Connected` a partir do registro de presença e
  chama o `fn`. O `fn` devolve `aplicado`; só quando é `true` a sala é
  persistida. Quando há save, a entrada do replay vai no mesmo `MULTI`.
- **`ApplyPlay` ficou puro.** Em
  [`../gameServer/game/engine/play.go`](../gameServer/game/engine/play.go) ele
  valida e aplica a jogada, usando os protocolos e as validações de sempre, e
  devolve `true`/`false`. Não salva, não transmite, não consulta Redis. O
  `ProcessPlay` é o adapter fino que chama `ApplyPlay` e, se aceitou, faz o
  `Broadcast`.
- **`Tick` ficou puro.** Em
  [`../gameServer/game/engine/gameProcessor.go`](../gameServer/game/engine/gameProcessor.go)
  ele fecha a janela atual e avança um passo: próximo evento de apresentação,
  próximo efeito lógico ou próximo turno. Com o sink no-op, roda só em memória.
- **Os protocolos continuam os mesmos.** Os arquivos de
  [`../gameServer/handlers/websocket/messages/protocols/`](../gameServer/handlers/websocket/messages/protocols/)
  só trocaram as fontes de tempo e aleatoriedade: `time.Now` virou `room.Now()`
  e `rand.*` virou `room.Rand()`. As regras também saíram do disco e passaram a
  ser embutidas (`//go:embed`) em `game/rules`.
- **O que ficou fora do motor:** o processador (`game/runtime`), o `redisFuncs`,
  o WebSocket/`ConnManager` e o chat. O `make engine-check` falha se o fecho do
  motor voltar a importar `go-redis`, `gorilla` ou `redisFuncs`.

O comportamento do jogo não mudou de propósito. A suíte antiga continua verde e
o teste `TestDeterministicReplaySameSeed` (em
`../gameServer/tests/scenario/determinism_test.go`) joga 200 partidas duas
vezes com a mesma semente, com `Clone()` entre os passos, e compara o hash do
estado a cada passo.

## 2. Como uma partida vira replay

O caminho é este:

```text
WebSocket -> WithRoom -> Redis (sala + lista replay:<matchId>)
          -> fim de jogo -> replayRecorder -> POST /internal/replays
          -> backend assina e grava no Postgres -> GET /replays/{matchId}
```

Passo a passo:

1. **A partida nasce com identidade e semente.** `CreateNewRoom`
   ([`../gameServer/services/roomService.go`](../gameServer/services/roomService.go))
   gera `MatchID` (UUID v7, público) e `Seed` (128 bits de `crypto/rand`). A
   semente é serializada no JSON da sala no Redis — que já guarda baralho e mãos
   — mas não entra no JSON público, não vai para o cliente e não aparece em log.
2. **Cada mudança passa pelo `WithRoom` com uma causa.** As causas usadas hoje
   são `play`, `tick`, `connect`, `disconnect` e `leave`; existem constantes
   para `join` e `start`, mas o início de partida é detectado pela passagem de
   `StartTime` de zero para não-zero. O `WithRoom` captura `now` uma única vez e
   fixa esse relógio durante o `fn`.
3. **O recorder decide o que vira entrada.** Em
   [`../gameServer/redisFuncs/replay.go`](../gameServer/redisFuncs/replay.go):
   - a transição de “não iniciado” para “iniciado” grava a entrada **0**,
     tipo `start`, com os conectados daquele momento;
   - toda jogada grava, **inclusive a recusada** (com `applied=false`), porque o
     replay também precisa reproduzir a tentativa e descartá-la;
   - `tick` só grava quando o `fn` devolve `aplicado=true`; tick que não rodou
     (janela ainda aberta, lock perdido) não grava nada;
   - `connect`, `disconnect`, `join` e `leave` só gravam quando aplicados;
   - partida com comando de debug (`NonReplayable`) ou com
     `REPLAY_ENABLED=false` não grava nada.
4. **A gravação é atômica com o save.** Quando o `fn` aplicou e há entrada para
   gravar, o `WithRoom` chama `SaveRoomAndReplay`
   ([`../gameServer/redisFuncs/persistence.go`](../gameServer/redisFuncs/persistence.go)),
   que faz `SET room` + `RPUSH replay:<matchId>` + `EXPIRE 6h` no mesmo `MULTI`.
   Sem entrada para gravar, faz só `SaveRoom`. Jogada recusada grava apenas a
   entrada, sem `SET room`. Cada entrada tem `at` (Unix ms), `type`, `applied`,
   `player`, a foto dos conectados, o payload JSON original da jogada e um
   checkpoint de 8 bytes.
5. **No fim, o arquivo é montado.** Quando a sala termina em `GameOver` e não é
   `NonReplayable`, o `WithRoom` dispara `ReplayFinalize` em goroutine. Em
   [`../gameServer/services/replayRecorder.go`](../gameServer/services/replayRecorder.go),
   `FinalizeReplay`/`buildReplay` leem a lista, trocam UUID por índice
   (a ordem lexicográfica usada pelos laços do motor), reescrevem os IDs dentro
   do payload, montam o bitmap de conectados, acrescentam a entrada final com o
   vencedor e os checkpoints, e chamam `format.Encode` **sem assinatura**.
6. **O envio é assíncrono e só limpa depois do “ok”.** O `.hwr` vai em base64
   para `POST /internal/replays` com `x-api-key` e timeout de 3 s. Se o backend
   responde 2xx, a lista `replay:<matchId>` é apagada. Se falhar, a lista fica
   até o TTL de 6 h e o replay **não** é gerado — não há retry.
7. **O backend valida, assina e guarda.** Em
   [`../backend/routes/handlers/replaysHandler.go`](../backend/routes/handlers/replaysHandler.go)
   ele decodifica com os limites do formato, exige versão conhecida, entrada
   final e vencedor, assina com `REPLAY_SIGNING_KEY` e grava o arquivo assinado
   com `ON CONFLICT DO NOTHING`. Depois `GET /replays/{matchId}` entrega o
   arquivo já assinado.

O `.hwr` é: `[u32 com o tamanho do gzip][corpo gzipado][64 bytes de assinatura,
opcional]`. O corpo tem magic `HWR1`, versão do formato, versão do motor, `kid`,
`matchId`, semente, início em Unix ms, modo, jogadores (índice + nome até 32
bytes), flags, entradas e checkpoints. O layout completo está documentado em
[`format/format.go`](format/format.go).

Tamanho real: no teste de 200 partidas a mediana ficou em centenas de bytes, e
o golden da v1 tem **368 bytes** (`engines/v1/testdata/golden_1.hwr`). O alvo do
formato é ≤ 4 KB.

## 3. Versões e CI: por que um replay antigo nunca muda

O motor é versionado por um número inteiro simples. A versão do **jogo**
(semver em `VERSION`, raiz do repo) é outra coisa e não entra neste controle:
ver [`../../VERSIONING.md`](../../VERSIONING.md). Existem dois arquivos no
gameServer:

- [`../gameServer/game/ENGINE_VERSION`](../gameServer/game/ENGINE_VERSION):
  a versão atual, hoje `1`.
- [`../gameServer/game/engine.sum`](../gameServer/game/engine.sum): o sha256 do
  **fecho do motor**, não do repositório inteiro.

O hash do `engine.sum` cobre:

- o conteúdo normalizado (LF) dos arquivos `.go` de todos os pacotes internos que
  `go list -deps ./game/engine` encontra;
- os YAML de `game/rules/modes/*.yaml`;
- as linhas do `go.sum` das dependências de terceiros usadas por esse fecho.

Por isso, **qualquer byte** do motor — inclusive uma regra em
`ClassicRules.yaml` — muda o hash. Já uma mudança no `redisFuncs`, no handler de
WebSocket (`handlers/http/websocket.go`) ou em `game/runtime` não muda, porque
essas partes ficam fora do fecho (os protocolos de carta, que decidem a partida,
continuam dentro).

O [`../tools/enginecheck/main.go`](../tools/enginecheck/main.go) roda no
`make engine-check` e falha quando:

1. o hash atual difere do `engine.sum` (motor mudou sem subir a versão);
2. existe `time.Now` ou `rand.` no fecho fora de `determinism.go`;
3. o fecho importa `go-redis`, `gorilla` ou `redisFuncs`;
4. uma cópia congelada `engines/vN` foi editada (o hash dela não bate com
   `vN/engine.sum`);
5. `services.EngineVersion` (a versão que vai no cabeçalho do `.hwr`) diverge de
   `game/ENGINE_VERSION`.

O [`../tools/enginebump/main.go`](../tools/enginebump/main.go) roda no
`make engine-bump`:

- se o fecho não mudou desde a última gravação, é no-op;
- senão, sobe a versão, copia o fecho para `engines/vN` reescrevendo os imports
  internos para `Hoodwink-replay/engines/vN/...`, copia os YAML de regras, grava
  o hash da cópia em `vN/engine.sum`;
- regenera `versions/versions.go` e `engines/registry.go`;
- atualiza `game/ENGINE_VERSION` e `game/engine.sum`.

`engines/vN` é **congelado**: não edite nada dentro dele. O `versions.go` é a
lista usada pelo backend para aceitar ou recusar uma versão; o `registry.go`
liga o runner ao pacote daquela versão. O runner escolhe a versão lida do
cabeçalho do `.hwr`; se não existir, devolve um erro claro.

O CI tem um job `engine-check` que roda `make engine-check` em `server/`, e o
job de deploy depende de `web`, `server` e `engine-check`. Ou seja: se você
mexeu no motor e não rodou o bump, o hash não bate, o CI fica vermelho e o
deploy não acontece. Isso é intencional — é o que obriga a congelar exatamente
o código que gerou as próximas partidas.

A toolchain também é fixa: `toolchain go1.26.5` no `go.mod` do gameServer e do
replay; os jobs `replay-wasm` e `engine-check` fixam `go-version: "1.26.5"`, e o
job `server` usa o `toolchain` do próprio `go.mod`.

Por que o replay antigo continua igual: cada arquivo diz a versão do motor que o
gerou; o runner abre a cópia congelada `engines/vN` daquela versão, que nunca
muda. Uma mudança futura cria a v(N+1), mas não toca na vN. O golden
`engines/v1/testdata/golden_1.hwr` (hash final
`1a453dd51cde6e05206a58247eb8c303b6fb8274ac05dbf81a636b5ef81bb627`) prova que
a v1 continua reproduzindo a mesma partida; ele prova regressão, não a
corretude das regras — isso é papel do teste de 200 partidas.

## 4. Como fazer o bump (`make engine-bump`)

### O passo a passo

Os comandos abaixo assumem que você está na raiz do repositório.

1. Faça a mudança no motor e rode a suíte do gameServer:
   ```bash
   (cd server/gameServer && go vet ./... && go test ./...)
   (cd server/replay && go test ./...)
   ```
2. Confira que o check ficou vermelho pelo motivo certo:
   ```bash
   (cd server && make engine-check)
   # deve dizer: "o hash do motor mudou sem subir a versão"
   ```
3. Congele a nova versão:
   ```bash
   (cd server && make engine-bump)
   ```
   Isso cria `server/replay/engines/vN/`, atualiza `versions/versions.go`,
   `engines/registry.go`, `game/ENGINE_VERSION`, `game/engine.sum` e sincroniza
   `services.EngineVersion` em `../gameServer/services/replayRecorder.go` (a
   versão que vai no cabeçalho do `.hwr`).
4. Não edite `services.EngineVersion` à mão: o `make engine-check` falha se ela
   divergir de `game/ENGINE_VERSION`. O `checkHeader` do e2e também compara com
   `services.EngineVersion`, então não há asserção de versão para ajustar.
5. Confirme o verde:
   ```bash
   (cd server && make engine-check)
   (cd server/replay && go test ./...)
   ```
6. Gere o WASM de **todas** as versões e o manifest do bundle:
   ```bash
   (cd server && make replay-wasm && make replay-wasm-check)
   ```
   `replay-wasm` apaga os `replay-v*.wasm*` antigos de
   `client/web/public/replay/`, compila um binário por versão listada em
   `versions.go`, comprime com brotli q11, copia o `wasm_exec.js` da toolchain e
   regenera `client/web/src/replay/manifest.gen.ts`.
7. Rode o front:
   ```bash
   (cd client/web && npm test && npm run build)
   ```
8. Faça o commit de tudo que o bump gerou:
   - `server/gameServer/game/ENGINE_VERSION` e
     `server/gameServer/game/engine.sum`;
   - `server/gameServer/services/replayRecorder.go` (versão do cabeçalho);
   - `server/replay/engines/vN/` inteiro;
   - `server/replay/versions/versions.go` e `server/replay/engines/registry.go`;
   - `client/web/src/replay/manifest.gen.ts`;
   - `client/web/public/replay/` (o `.wasm`, o `.wasm.br` e o `wasm_exec.js`).

### Deploy na ordem certa

O backend precisa conhecer a nova versão antes de receber um replay vN (senão
responde “versão do motor não suportada”). O web precisa ter o `.wasm` e o
manifest da vN antes de alguém tentar assistir. A ordem ideal é:

```text
backend (com o novo versions.All)
  -> web (com o novo WASM/manifest)
  -> gameServer (começa a gerar partidas vN)
```

O script de deploy atual faz um único `helm upgrade` com as três imagens; a
janela costuma ser curta, mas em partida longa ainda pode existir. Se puder,
faça o bump em horário sem partidas iniciadas.

### Se uma partida atravessar o deploy

O plano previa gravar a versão do motor na sala e marcar a partida como “não
replayável” se ela mudasse no meio. **O código atual ainda não faz isso**: a
sala não guarda a versão com que começou, e `NonReplayable` hoje só é ligado
pelo comando de debug
([`../gameServer/handlers/websocket/connection/processDebugCommand.go`](../gameServer/handlers/websocket/connection/processDebugCommand.go)).
Na prática:

- evite subir um bump com partida iniciada; espere a sala terminar (o TTL de
  sala é `ROOM_AFK_TTL_LIMIT=180` segundos e a partida tem teto de 6 h);
- se uma partida atravessar, parte das entradas foi produzida pelo motor
  antigo e parte pelo novo. O arquivo pode sair com a versão nova, mas o runner
  vN vai replicar entradas antigas com regras novas e provavelmente acusar
  `Divergence`;
- trate esse replay como suspeito: não tente “consertar” regerando com o motor
  novo. O caminho existente é marcar a sala como `NonReplayable` se ela ainda
  estiver em andamento, ou descartar a linha de `match_replays` se o arquivo já
  foi gerado. A assinatura continua válida (o backend emitiu o arquivo), mas ele
  não é uma reprodução confiável.

### Se você mudou só o runner ou o formato

Não há bump de motor se o fecho do gameServer não mudou. Ainda assim, rode
`make replay-wasm` e `make replay-wasm-check`, porque o `.wasm` é compilado a
partir do runner atual: o hash e o conteúdo do arquivo mudam mesmo com a mesma
`ENGINE_VERSION`.

## 5. Chaves, `kid` e rotação

- O backend assina os replays com `REPLAY_SIGNING_KEY`, que é **base64 de um PEM
  PKCS#8 Ed25519 em uma linha**. A validação acontece na subida
  ([`../backend/services/replayService.go`](../backend/services/replayService.go));
  sem chave ou com chave inválida, o backend não sobe.
- O `kid` é o nome público da chave. Hoje o backend usa a constante
  `replaySigningKid = "replay-v1"`; ele vai no cabeçalho do `.hwr` e permite que
  o cliente escolha a chave pública certa.
- O cliente embute a lista de chaves públicas em
  [`../../client/web/src/replay/keys.ts`](../../client/web/src/replay/keys.ts).
  Cada item é `{ kid, publicKeyHex }`. O `kid` não é segredo.
- A assinatura cobre `"HWR1-SIG" || corpo gzipado` e fica **fora** do gzip, no
  fim do arquivo. Trocar um byte do arquivo invalida a assinatura.
- A chave pública que está no `keys.ts` hoje é a de desenvolvimento, derivada da
  chave de exemplo do `server/backend/.env.example`. **Em produção é preciso
  trocar pela pública da chave real** antes de publicar; senão todo replay
  oficial abre com a faixa “replay não oficial”.
- O backend assina na gravação, não no download. O arquivo guardado no Postgres
  já é o arquivo assinado.

Como rotacionar sem quebrar replays antigos:

1. Gere um novo par Ed25519 e escolha um novo `kid` (por exemplo,
   `replay-v2`). O `kid` deve ter no máximo 32 bytes.
2. Adicione a **nova pública** em `keys.ts` **mantendo** as antigas. O formato e
   o cliente já aceitam mais de um `kid`.
3. Publique o web primeiro. A partir daí o bundle conhece a chave nova e a
   antiga.
4. Troque `REPLAY_SIGNING_KEY` no secret do deploy e a constante
   `replaySigningKid` no backend para o novo `kid`; publique o backend em
   seguida. Novos replays passam a sair assinados com a chave nova.
5. Mantenha a chave pública antiga no bundle enquanto existir replay assinado
   com ela. A retenção apaga os replays do servidor depois de 30 dias; passado
   esse prazo (e depois de confirmar que não há replay vivo com o `kid` antigo),
   remova a pública antiga e a privada antiga do secret.

Enquanto a pública antiga estiver no bundle, os arquivos antigos continuam
oficiais. Depois que ela for removida, um `.hwr` antigo baixado no disco só
continua reproduzível se o motor daquela versão ainda estiver no manifest, e
mesmo assim sai como “não oficial” — por isso a remoção só deve acontecer depois
da retenção.

A assinatura prova que o backend emitiu aquele arquivo. Ela **não** prova que o
motor consegue reproduzi-lo; isso é papel dos checkpoints e dos testes.

## 6. Custo de manter versões e como aposentar uma `vN`

Cada `engines/vN` não é o repositório inteiro: é só o fecho do motor, os YAML de
regras e o `engine.sum`. Mesmo assim, cada versão:

- ocupa espaço no repositório e entra no build do módulo `server/replay`;
- gera um `.wasm` e um `.wasm.br` no bundle (hoje ~5,4 MiB cru e ~1,1 MiB
  brotli por versão);
- é importada pelo `engines/registry.go`, então o runner e o WASM carregam o
  registry de todas as versões. Quanto mais versões, maior o binário e o tempo
  de build.

Regra de aposentadoria: **só apague uma versão quando não houver replay vivo
que a use**. O backend apaga replays com `ended_at < now() - interval '30 days'`
no job de retenção
([`../backend/services/requestLogRetention.go`](../backend/services/requestLogRetention.go)).
Além disso, não pode haver partida em andamento com aquela versão (o teto de
partida é 6 h). Na prática, 30 dias é uma folga confortável.

Para decidir, consulte o Postgres:

```sql
SELECT engine_version, count(*)
FROM match_replays
GROUP BY 1
ORDER BY 1;
```

Se a versão não aparece (ou só aparece com replays que já saíram da retenção),
não é a versão atual de `ENGINE_VERSION` e não há partida em andamento, ela pode
ser aposentada. O que apagar:

1. a pasta `server/replay/engines/vN/` inteira (incluindo `testdata/`, se
   houver);
2. os artefatos `client/web/public/replay/replay-vN-*.wasm` e `.wasm.br`;
3. a entrada da versão em `server/replay/versions/versions.go` e os imports,
   tipos e a entrada do registry em `server/replay/engines/registry.go`;
4. se existir um golden da versão, remova/ajuste o teste correspondente (hoje
   `TestGoldenV1` lê `engines/v1/testdata/golden_1.hwr`).

Atenção: **não existe hoje um `make engine-prune`**. O `engine-bump` só regenera
`versions.go` e `registry.go` quando o fecho mudou; se você só apagar a pasta, o
`registry.go` continua importando um pacote que não existe e o build quebra.
Depois de remover a pasta, ajuste os dois arquivos gerados (ou rode o bump
depois de uma mudança real) e confira:

```bash
cd server/replay && go build ./...
cd .. && make engine-check
make replay-wasm
make replay-wasm-check
```

O `make replay-wasm` começa apagando todos os `replay-v*.wasm*` e reconstrói
apenas as versões que sobraram em `versions.go`; é ele que limpa os artefatos da
versão aposentada e regenera o manifest. Um `.hwr` antigo baixado no disco deixa
de abrir no bundle novo se a versão dele foi removida (o loader diz “versão do
motor de replay indisponível”). Isso é esperado depois da retenção.

## 7. Como desligar (`REPLAY_ENABLED=false`)

O gameServer lê `REPLAY_ENABLED` em
[`../gameServer/redisFuncs/replay.go`](../gameServer/redisFuncs/replay.go). O
padrão é **ligado**; só o valor exatamente `false` desliga. `FALSE`, `0` ou
`off` não desligam.

No Kubernetes, a variável está em
[`../../k8s/charts/hoodwink/values.yaml`](../../k8s/charts/hoodwink/values.yaml)
(`gameserver.config.REPLAY_ENABLED: "true"`). Para desligar, mude para
`"false"`, faça o deploy/rollout do gameServer e reinicie os pods. Não é preciso
deploy de código. Localmente, mude no `gameServer/.env`.

O que deixa de funcionar:

- nenhuma entrada nova é gravada (`replayEntryToRecord` retorna `false`);
- não há `replay:<matchId>` no Redis para partidas novas;
- no fim, `ReplayFinalize` não é chamado, então não há `.hwr`, nem envio ao
  backend, nem linha em `match_replays`;
- o histórico e o botão “Assistir” não ganham partidas novas.

O que continua funcionando:

- salvar a sala e jogar normalmente (a gravação de replay é separada do save);
- baixar e assistir replays já gravados;
- o backend, a assinatura e a retenção dos replays antigos.

Cuidados:

- desligar no meio de uma partida deixa uma lista parcial no Redis; ela expira
  em 6 h e não é finalizada. Ligue/desligue entre partidas;
- religar no meio de uma partida faz o recorder começar sem a entrada `start`;
  o runner exige que a primeira entrada seja `start`, então essa partida não gera
  um `.hwr` reproduzível. O certo é religar antes da próxima partida;
- o backend continua exigindo `REPLAY_SIGNING_KEY` na subida, mesmo com a
  gravação desligada.

## Ver o teste ao vivo (opcional)

O e2e de 200 partidas fica silencioso por padrão (CI). Para acompanhar no terminal:

```bash
cd server/replay/e2e
E2E_VISUAL=1 go test -v ./... -count=1
# 200 partidas é longo; para um smoke rápido:
E2E_VISUAL=1 go test -short -v ./...
```

A cada ~150 ms (ou 25 passos) ele reescreve uma linha no `stderr` com partida, passo,
turno, jogadores vivos e a última ação (`jogada TAKE_COINS`, `tick`, `desconectou...`).
No fim de cada partida imprime entradas, tamanho do `.hwr` e hash. Sem `E2E_VISUAL`,
os helpers retornam `nil` e não há escrita nem formatação — o único custo é um branch
por passo. O `-v` é necessário porque o `go test` captura a saída dos testes que passam.

## Quando algo diverge

O replay confere o estado a cada 20 entradas e no fim. O checkpoint é composto
pelos **8 primeiros bytes do sha256** do JSON canônico da sala. A
canonicalização existe para o hash não depender de UUID, nome de exibição ou
precisão de relógio; ela está duplicada de propósito nos dois lados:

- no servidor, em
  [`../gameServer/redisFuncs/replayCanonical.go`](../gameServer/redisFuncs/replayCanonical.go);
- no runner, em `normalizedForHash`/`checkpointOf`
  ([`runner/runner.go`](runner/runner.go)).

Os dois precisam normalizar igual: ID do jogador vira índice na ordem
lexicográfica (`SortedPlayerIDs`), `MoveTime` vira ordem densa, nome/senha/ID da
sala/chat são zerados, e tempos são truncados em milissegundos.

O runner:

- escolhe o motor pelo `EngineVersion` do cabeçalho; versão desconhecida é erro
  claro (`versão do motor desconhecida`);
- injeta a semente, o relógio (`StartedAtMs` + deltas) e os conectados de cada
  entrada;
- faz `room.Clone()` antes de cada passo (ida e volta JSON, como o `WithRoom` de
  produção);
- aplica jogada recusada em um clone descartado, igual ao servidor;
- confere os checkpoints e devolve `*Divergence` com `Step`, `Want` e `Got` no
  primeiro passo que não bate;
- recupera panic e devolve `*StepError` com o passo;
- para em 5.000 passos.

Quando o runner acusa divergência, o primeiro passo apontado é o suspeito: as
entradas anteriores bateram. O que investigar, nesta ordem:

1. **A versão está certa?** Confira `EngineVersion` no cabeçalho do `.hwr`, o
   `engineVersion` do `replayRecorder.go` e se o manifest do web tem o `.wasm`
   dessa versão. Um bump sem atualizar a constante do recorder é a causa mais
   comum.
2. **A partida atravessou um deploy?** Se sim, houve mistura de motores; veja a
   seção 4 e trate o arquivo como suspeito.
3. **O motor foi editado sem bump?** `(cd server && make engine-check)` aponta
   o hash e as cópias congeladas.
4. **A cópia `engines/vN` foi editada?** O mesmo check acusa; restaure a cópia
   do git.
5. **A canonicalização mudou de um lado só?** Compare
   `replayCanonical.go` com `normalizedForHash`; os dois são um contrato.
6. **O payload da entrada mudou?** Veja a entrada no passo indicado
   (`Type`, `Player`, `Payload`); o runner usa o mesmo `ParsePlayerPlay` do
   websocket.
7. **O relógio/conectados?** O delta e o bitmap da entrada precisam ser os
   mesmos que o servidor viu. Um `now` capturado duas vezes ou um `Connected`
   diferente muda o estado.

Ferramentas de investigação:

```bash
cd server/replay

# golden da v1: reproduz do zero e confere o hash final
cd e2e && go test ./... -run TestGoldenV1 -v

# o teste que adultera um checkpoint mostra o passo exato
go test ./runner/ -run TestCheckpointAdulteradoAcusaOPasso -v

# ida e volta do formato e limites
go test ./format/ -v

# e2e completo (200 partidas, desconexões e jogadas inválidas)
cd e2e && go test ./... -count=2
```

Para um arquivo específico, o caminho é o mesmo dos testes: `format.Decode` para
abrir, `runner.New` para escolher o motor, `Step` até o erro, `Position()` para
saber o passo e `Hash()` para comparar. Não existe um CLI de replay; os testes
`runner_test.go` e `e2e_test.go` são o modelo.

O golden da v1 (`engines/v1/testdata/golden_1.hwr` + `golden_1.sha256`) é uma
prova de regressão. **Nunca rode `UPDATE_GOLDEN=1` para “consertar” a v1**: se
ele falhar, ou o motor v1 foi alterado, ou o golden foi corrompido. Cuidado: o
`TestGoldenV1` grava sempre em `engines/v1/testdata/`; rodar `UPDATE_GOLDEN=1`
depois de um bump gravaria um replay da versão nova no lugar do golden v1.
Regenerar só vale quando o teste aponta para a versão/golden que você está
criando — nunca para esconder uma diferença na v1.

Uma divergência não é a mesma coisa que “replay não oficial”. “Não oficial” é o
selo de assinatura: arquivo sem assinatura, `kid` desconhecido ou assinatura
errada. Um arquivo assinado pode divergir se o motor/versão não for o mesmo; ele
é autêntico, mas não é uma reprodução confiável. Não edite o `.hwr` para tentar
resolver — a assinatura cobre o gzip inteiro.

## 9. Limites e segurança do arquivo

O `.hwr` vem de fora (do disco ou da rede), então o leitor é hostil. Os limites
estão em [`format/format.go`](format/format.go):

| Limite | Valor | Constante |
| --- | --- | --- |
| Tamanho do arquivo | 256 KiB, recusado antes de ler | `MaxFileSize` |
| Corpo depois do gunzip | 1 MiB | `MaxDecompressedSize` |
| Jogadores | 2 a 10 | `MinPlayers`/`MaxPlayers` |
| Nome de jogador | 32 bytes no arquivo; 32 code points na tela | `MaxPlayerNameBytes`/`sanitizeText` |
| Modo | 32 bytes | `MaxModeBytes` |
| `kid` | 32 bytes | `MaxKidBytes` |
| Entradas | 1 a 5.000 | `MaxEntries` |
| Payload de uma jogada | 2 KiB | `MaxPayloadBytes` |
| Duração total | 6 horas | `MaxDurationMs` |
| Checkpoint | 8 bytes a cada 20 entradas e no fim | `CheckpointInterval`/`CheckpointSize` |
| Assinatura | 64 bytes | `SignatureSize` |

O leitor também recusa: delta negativo, tipo de entrada fora da lista fechada,
índice de jogador inválido, bitmap do tamanho errado, número de checkpoints
diferente do esperado, bytes sobrando no fim, magic/versão errados e gzip
multistream (bomba de gzip). Todo erro estrutural vira a mensagem fixa
`replay inválido`, sem ecoar bytes e sem panic.

O front repete o teto antes de ler
([`../../client/web/src/replay/loader.ts`](../../client/web/src/replay/loader.ts)):
`MAX_REPLAY_BYTES = 256 * 1024`, conferido em `assertReplayFileSize` antes do
`arrayBuffer`; o envelope só aceita resto 0 (sem assinatura) ou 64 bytes
(assinatura); para descobrir a versão ele descomprime só um prefixo de 32 bytes,
e o teto de 1 MiB do corpo inteiro é aplicado pelo leitor Go dentro do WASM;
cada pedido ao Worker tem 10 s e o Worker é encerrado se estourar.

Assinatura:

- sem assinatura: `Official=false`, status `unsigned`, o arquivo abre com a
  faixa “replay não oficial”;
- `kid` desconhecido: `ErrUnknownKey`, status `unknown-kid`, abre como não
  oficial;
- assinatura errada: `ErrInvalidSignature`, status `invalid`, abre como não
  oficial;
- assinatura válida: `Official=true`, status `valid`.

O WASM também é tratado como hostil:

- o `sha256` dos bytes **descomprimidos** é conferido contra o manifest antes de
  `WebAssembly.instantiate`; nunca se usa `instantiateStreaming` direto da rede;
- hash errado apaga o Cache Storage e tenta uma vez; se o servidor entregar o
  `.br` sem `Content-Encoding: br`, o loader detecta e avisa em vez de instanciar
  lixo;
- o nginx serve o `.wasm.br` com `application/wasm`, `Content-Encoding: br` e
  `immutable` apenas no nome com hash; o `wasm_exec.js` é `no-cache` e tem o
  sha256 no manifest;
- a CSP do web inclui `wasm-unsafe-eval`, `worker-src 'self' blob:`,
  `object-src 'none'` e `base-uri 'none'` (o arquivo atual também libera as
  origens do Google Login).

Nomes e textos vindos do arquivo entram como texto do React, nunca como HTML;
`sanitizeText`
([`../../client/web/src/features/replay/sanitize.ts`](../../client/web/src/features/replay/sanitize.ts))
corta em 32 code points e remove caracteres de controle e bidi. Chat não entra no
replay.

A semente é o único segredo do arquivo. Durante a partida ela não vai ao
cliente, não vai a log e não vai a erro; a lista `replay:<matchId>` guarda só
entradas, sem semente. Ela só aparece no cabeçalho do `.hwr` depois do fim, que é
quando o arquivo é montado. O motor não guarda nenhuma chave secreta.

O backend ainda aplica a própria camada:

- `POST /internal/replays` exige API key, aceita no máximo 256 KiB, decodifica
  com os mesmos limites, exige versão em `versions.All`, entrada final e
  vencedor; a versão e o vencedor vêm do cabeçalho, e os UUIDs dos jogadores vêm
  do `player_ids` do corpo (o backend confere se as quantidades batem);
- `ON CONFLICT DO NOTHING` torna o envio idempotente e nunca sobrescreve um
  replay já gravado;
- `GET /replays/{matchId}` exige JWT, devolve o mesmo 404 para inexistente e
  para quem não jogou, e usa `Cache-Control: private, no-store`;
- a retenção apaga depois de 30 dias.

Criptografar o WASM não protegeria nada: o navegador precisa decifrar e a chave
iria junto. O hash detecta arquivo corrompido ou cache envenenado; contra um
servidor invadido, a defesa é o CSP e o processo de deploy. O motor não tem
segredo para roubar.

## 10. Estado atual do deploy

### O que o `make replay-wasm` gera

O alvo em [`../makefile`](../makefile):

- usa **Go padrão 1.26.5** (`GOOS=js GOARCH=wasm`), com
  `-trimpath -ldflags="-s -w -X main.buildEngineVersion=N"`;
- compila um binário por versão de `versions/versions.go`;
- grava `client/web/public/replay/replay-vN-<8 hex do sha256>.wasm` e o
  `.wasm.br` (brotli q11);
- copia o `wasm_exec.js` da toolchain para `public/replay/`;
- gera `client/web/src/replay/manifest.gen.ts` com `version`, `file`, `sha256`
  (bytes descomprimidos), `sha256Br` (arquivo `.br`) e `wasmExecSha256`;
- apaga antes todos os `replay-v*.wasm*` antigos, então não deixa órfãos.

O `make replay-wasm-check` (via `cmd/wasm/check_manifest.mjs`) confere que cada
arquivo existe, que o sha256 do `.wasm` e do `.wasm.br` bate com o manifest, que
o `wasm_exec.js` bate, que o `.wasm` começa com `\0asm` e que não há arquivo
órfão.

No estado atual da branch, a v1 gera aproximadamente:

- `replay-v1-f9e92d0e.wasm`: ~5,4 MiB;
- `replay-v1-f9e92d0e.wasm.br`: ~1,1 MiB;
- `wasm_exec.js`: ~17 KiB;
- sha256 do wasm `f9e92d0e...` e do `.br` `da62d191...`.

**TinyGo e `wasm-opt` não existem no ambiente e não são usados.** O build é Go
padrão; por isso o `.wasm` é maior. `brotli` é obrigatório (`make replay-wasm`
falha sem ele), e `node` é obrigatório para o `replay-wasm-check`.

**Regenerar depois de mexer no runner.** O `.wasm` é compilado a partir de
`runner/`, `format/`, `cmd/wasm/` e do motor. Se você mudar o runner e não rodar
`make replay-wasm`, o bundle continua com o motor antigo mesmo que o servidor
esteja gravando partidas novas. Rode também depois de qualquer bump (a nova
versão precisa do próprio `.wasm` e de uma entrada no manifest).

### CI

O workflow [`../../.github/workflows/ci.yml`](../../.github/workflows/ci.yml)
tem:

- job `replay-wasm`: instala Go 1.26.5 e brotli, roda `make -C server
  replay-wasm`, confere os artefatos e os envia como artifact;
- job `web`: depende do `replay-wasm`, baixa o artifact para `client/web` e roda
  `npm ci`, `npm test` e `npm run build`;
- job `server`: `go vet` + `go test` em `server/backend` e `server/gameServer`;
- job `engine-check`: roda `make engine-check` em `server/`;
- job `deploy`: só em push para `production`, e só depois de `web`, `server` e
  `engine-check` verdes. Ele builda as imagens na VM a partir do checkout e faz
  `helm upgrade --atomic`.

Consequência importante: o CI **regenera** o WASM no job `replay-wasm` e o job
`web` usa esse artifact. Mas o deploy na VM **não roda** `make replay-wasm`: ele
builda o Dockerfile a partir do repositório. O
[`../../client/web/Dockerfile`](../../client/web/Dockerfile) exige
`public/replay` e `src/replay/manifest.gen.ts` e falha cedo se não existirem.
Portanto, os artefatos gerados precisam estar **commitados**. Se ficarem
desatualizados, o CI ainda passa (ele gera os seus), mas a produção recebe o
WASM antigo. O `replay-wasm-check` do CI confere o artifact gerado, não o
arquivo commitado.

### nginx, CSP e cache

O [`../../client/web/nginx.conf`](../../client/web/nginx.conf):

- declara `types { application/wasm wasm; }`;
- serve `replay-vN-<hash>.wasm.br` com `Content-Encoding: br`,
  `application/wasm` e `Cache-Control: public, max-age=31536000, immutable`
  apenas no padrão com hash;
- serve `wasm_exec.js` com `no-cache`;
- repete a CSP no `location = /index.html` (o `add_header` não é herdado por
  location que declara outro `add_header`).

O app guarda o `.wasm` no Cache Storage (`hoodwink-replay-wasm-v1`) e o fetch usa
`force-cache`. O segundo carregamento não baixa o mesmo arquivo; um bump gera
nome/hash novo, então a URL muda e o cache antigo não é usado.

### Chaves e migrations no deploy

- `REPLAY_SIGNING_KEY` precisa estar no secret do deploy (o chart mapeia
  `secrets.keys.replaySigningKey`). O backend não sobe sem ela. O texto de erro
  do script de deploy lista outras chaves e não menciona essa; confira o
  `k8s/.env.production` manualmente.
- A chave pública de produção precisa estar em `keys.ts`; a atual é a de
  desenvolvimento.
- A migration `server/backend/db/migrations/202611070003_match_replays.sql`
  cria `match_replays` (e o configmap tem o mesmo SQL idempotente). Sem ela, o
  upload de replay falha.
- A retenção de 30 dias roda no job do plano 04
  (`requestLogRetention.go`), no mesmo advisory lock.

### Limitações conhecidas deste estado

- O `engineVersion` do recorder é uma constante manual; o `engine-bump` não a
  atualiza nem atualiza asserções de teste que fixam a versão.
- A sala não guarda a versão do motor; uma partida que atravessa deploy não é
  marcada automaticamente como não replayável.
- Não existe `make engine-prune`; aposentar uma versão exige ajustar
  `versions.go` e `registry.go` à mão (ou rodar o bump depois de uma mudança
  real).
- O CI não compara o WASM commitado com o WASM que ele gera; o deploy usa o
  commitado.

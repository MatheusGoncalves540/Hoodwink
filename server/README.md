# 🎭 Hoodwink

> Um jogo de blefe estratégico e imprevisível, onde cartas são vidas, moedas são poder, e a mentira pode ser sua maior arma.

## 📌 Descrição

**Hoodwink** é um jogo de cartas baseado em blefe, onde os jogadores utilizam habilidades especiais para manipular a partida, eliminar oponentes e tentar ser o último sobrevivente. As cartas não são jogadas como em um baralho comum — elas representam as "vidas" dos jogadores, e cada uma possui habilidades únicas que podem ser utilizadas com ou sem possuí-las de fato.

---

## 🧠 Conceitos Principais

- Cada jogador começa com **2 cartas** (vidas) e **3 moedas** (definido em `gameServer/game/rules/modes/ClassicRules.yaml`).
- É possível **usar habilidades de cartas sem necessariamente tê-las**, blefando.
- Os outros jogadores podem **contestar** ações suspeitas.
- **Moedas** são utilizadas para ações básicas e ativação de habilidades.
- O objetivo do jogo é ser o **último jogador com cartas vivas na mão**.

---

## ▶️ Ações Básicas por Turno

Em cada turno, o jogador pode realizar **uma das ações abaixo**:

1. Passar o turno (deixar o tempo acabar).
2. **Coletar 1 moeda + taxas** (nada pode impedir).
3. **Pagar 6 + (taxas / 2, arredondado para baixo)** para matar 1 carta de um adversário (pode ser bloqueado pela Guardiã).
4. **Usar uma habilidade de carta** (mesmo sem possuí-la, blefando).
5. **Trocar carta(s) com o baralho**: pagando metade das suas moedas (arredondada para cima) por 1 carta, ou todas as moedas para trocar as 2. É preciso ter no mínimo **6 moedas**.

---

## ⚠️ Contestações

- Jogadores vivos podem contestar o uso de uma carta.
- Se o jogador **possuía a carta**: ela é revelada, o jogador escolhe uma carta do contestador para morrer e, depois, descarta a carta revelada e compra outra do baralho.
- Se o jogador **não possuía a carta**: o contestador escolhe uma carta do jogador para morrer e o efeito da carta é cancelado (as moedas gastas não são devolvidas).

---

## 💰 Limite de Moedas

- O limite máximo de moedas por jogador é **20**.
- Ao receber mais que isso, o excedente é **desconsiderado** e o jogador **perde automaticamente uma carta** (nada pode impedir essa perda).

---

## 📈 Sistema de Taxas

- As **taxas** aumentam o custo de uso das habilidades de cartas e são usadas em algumas ações básicas.
- No modo clássico elas não mudam aleatoriamente: só sobem com o **Político** e descem com o **Rebelde**.
- As taxas variam de **-5 a +5** e todos os cálculos são arredondados para baixo.
- Kamikaze, Político e Rebelde não são afetados pelas taxas.

---

## 🃏 Habilidades das Cartas

- Cada carta possui **uma habilidade especial**, com interações variadas entre elas.
- Algumas cartas possuem efeitos passivos, outras reativas, e outras ofensivas ou defensivas.

Exemplos:
- **Político**: aumenta as taxas em 1 (efeito acumulativo).
- **Assassino**: mata uma carta por 3 moedas (+ taxas).
- **Guardiã**: impede ataques.
- **Trapaceiro**: rouba moedas.
- **Egoísta**: impede que o adversário receba moedas ou cartas.
- **Kamikaze**: leva uma carta inimiga ao morrer.

Os valores e limites de cada carta ficam em `gameServer/game/rules/modes/ClassicRules.yaml`, que é a fonte de verdade do modo clássico.

---

## 🏁 Condições de Vitória

- Um jogador é **eliminado** ao perder todas as suas cartas.
- O jogo termina quando **restar apenas um jogador com cartas vivas**.

---

## 🛠️ Desenvolvimento

O ambiente de dev sobe o `backend` (API REST + PostgreSQL) e o `gameServer` (WebSocket + Redis) juntos a partir do `makefile` da raiz.

Pré-requisitos: Go 1.24+, Podman (máquina iniciada com `podman machine start`) e, opcionalmente, `goose` para migrations e `air` para live reload. No Windows, o `makefile` detecta e usa automaticamente o `sh.exe` do Git for Windows (instalação padrão em `C:\Program Files\Git`), então os alvos funcionam direto do PowerShell.

```bash
# Primeira execução: dependências Go + containers
make setup

# Sobe o ambiente completo (Ctrl+C encerra os dois serviços)
make dev
```

Cada serviço carrega o próprio `.env` (`backend/.env` e `gameServer/.env`), então mantenha-os criados a partir dos respectivos `.env.example`.

Outros alvos úteis:

- `make backend` / `make game` — sobe apenas um dos serviços
- `make infra-up` / `make infra-down` / `make infra-clean` — gerencia PostgreSQL e Redis
- `make migrate` — aplica as migrations pendentes (goose)
- `make build` / `make test` — compila ou testa os dois módulos
- `make engine-check` — verifica a versão/fecho do motor de replay; `make engine-bump` — congela o motor em `replay/engines/vN` (toda mudança no motor exige o bump)
- `make help` — lista todos os alvos

---

## 📂 Estrutura do Projeto

> Este repositório contém o código-fonte, regras completas e eventuais assets do jogo Hoodwink.

---

## IP do jogador (JWT e ticket)

Os tokens de login (HS256) e o ticket de sala (Ed25519) levam, além das claims
existentes, três claims novas — `ip`, `ua` (User-Agent cortado em 256) e `iat`
(Unix). Elas são só registro para rastrear trapaça: nada no sistema recusa um
token porque o IP/navegador mudou. A leitura é sempre comma-ok — um token antigo
sem essas claims continua válido.

A regra de IP é única nos dois módulos (`utils.ClientIP(r)`):

| Caminho | IP gravado |
| --- | --- |
| Produção pela Cloudflare | nginx troca `$remote_addr` pelo `CF-Connecting-IP` (só quando vem de faixa da Cloudflare) e repassa como `X-Real-IP` pela loopback → usa o `X-Real-IP` |
| Direto no 443 com `CF-Connecting-IP` falso | não vem de faixa da Cloudflare, o realip ignora → vale o IP real de quem bateu |
| Direto no 8088/4444 | não é loopback, `X-Real-IP` ignorado → vale o IP real |
| `make dev` local | loopback sem cabeçalho → `127.0.0.1` |
| Amigos pela Radmin | falam direto no 8088/4444 → `RemoteAddr` (26.x), cabeçalho ignorado |
| gameServer → backend (`/internal/*`) | loopback sem cabeçalho → `127.0.0.1`; o IP do jogador vai no corpo de `/internal/tickets` |
| Probes do k8s / smoke test | não são gravados |

`ClientIP` só confia em `X-Real-IP` quando `RemoteAddr` é loopback e o header é
um IP válido; senão usa `RemoteAddr` sem porta; nunca lê `X-Forwarded-For` nem
`CF-Connecting-IP`.

### Renovação do login (`/auth/refresh`)

`POST /auth/refresh` (grupo JWT) recarrega o usuário pelo `id` do token (401 se
não existir), recusa token `temp`, copia `username`/`email` do banco e devolve
`{token}` com `exp`/`iat`/`ip`/`ua` novos.

O cliente renova quando já passou da metade da vida
(`now > iat + (exp - iat) / 2`); sem `iat`, quando faltam menos de 12h para o
`exp`. Renova ao abrir o app e antes de `/newRoom` e `/getTicket`, com uma
Promise por vez. Erro de rede mantém o token; 401 limpa a sessão e abre o login;
token vencido não é enviado.

---

## Coletor de requests (plano 04)

O backend e o gameServer gravam toda request HTTP em `request_logs`, no mesmo
Postgres. O gameServer junta as linhas e manda em `POST /internal/requestLogs`
com a API key; o backend mascara de novo antes de gravar. Ficam de fora `/alive`,
tudo em `/internal/*` e mensagens de WebSocket (do gameServer só entra o upgrade
`/game`). A query string nunca é gravada: só `r.URL.Path`.

### Virar super-admin

A tela confere `users.is_super_admin` no banco a cada request. O valor não vai no
JWT, então mudar no banco vale na próxima chamada. Para liberar uma conta:

```sql
UPDATE users SET is_super_admin = true WHERE username = 'seu-usuario';
```

Usuário sem a coluna ligada recebe 403 em `GET /admin/requestLogs`.

### Tela `/#admin`

Abra `http://localhost:5173/#admin` em dev ou `https://<seu-dominio>/#admin` em
produção. Não há link no menu. A tela lista as requests e filtra por rota,
período (24h, 7d, 30d ou personalizado), status (`4xx` ou um inteiro de 100 a
599), nome do usuário, ID do usuário (UUID), serviço e arquivados. O detalhe
mostra corpo, erro, User-Agent e IP como texto puro.

- 401: sessão inválida ou expirada.
- 403: a conta não é super-admin.

Limites da API: intervalo de até 31 dias nos ativos e 183 dias nos arquivados,
`page` de 1 a 1000 e `limit` de até 200.

### Campos gravados

| Campo | Conteúdo |
| --- | --- |
| `at` | horário UTC da request |
| `service` | `backend` ou `gameServer` |
| `method` | método HTTP |
| `path` | caminho da URL, sem query string (até 512) |
| `route` | padrão do chi, por exemplo `/getTicket/{RoomId}` |
| `status` | código HTTP |
| `duration_ms` | duração em milissegundos |
| `ip` | IP do jogador, pela regra `ClientIP` acima |
| `user_agent` | navegador, cortado em 256 |
| `user_id` | UUID do usuário, quando autenticado |
| `username` | nome do usuário, cortado em 64 |
| `request_id` | identificador da request |
| `body` | JSON mascarado, até 4 KB. JSON inválido ou não-JSON vira `<N bytes>` |
| `error` | texto do erro, mascarado e cortado em 1024 |

### Máscara

O cabeçalho `Authorization` nunca é gravado. No corpo, qualquer chave que
contenha, sem diferenciar maiúsculas, `token`, `ticket`, `password`, `senha`,
`secret`, `authorization`, `apikey`, `api-key`, `credential` ou `jwt` vira `"***"`
em qualquer profundidade. O campo `error` passa pela mesma máscara.

A lista existe nos dois módulos. Mudou em um, mude no outro:
`server/backend/utils/maskBody.go` e `server/gameServer/utils/maskBody.go`.

### Retenção

Um job roda a cada hora no backend. Numa transação com
`pg_try_advisory_xact_lock`, ele move para `request_logs_archive` o que tem mais
de 30 dias, em lotes de 10.000, e preserva o `id` original. Depois apaga do
arquivo o que passou de 6 meses.

Para rodar uma vez e sair:

```bash
cd server/backend && go run . -run-retention-once
```

### Dado pessoal no corpo

O login de visitante não manda `email` no corpo. Ele manda `username`, e o
backend monta o email `<username>+<uuid>@guest.hoodwink`. Como o corpo fica até
6 meses no banco, esse username (que compõe o email) também fica. Qualquer email
que apareça em outro corpo segue a mesma retenção. Trate o histórico como dado
pessoal: a máscara cobre segredos, não emails.

---

## Relatos de bug (plano 06)

O jogador marca **Algo deu errado** durante a partida e detalha no fim; fora
da partida, reporta pelo Lobby. O relato guarda o `match_id` público do plano
05 e os turnos marcados, mas não abre o replay, não anexa print e não espera
resposta. Marcar sem detalhar já vale como relato: ele fica com `match_id` e
`turns` para localizar o momento depois.

### Fluxo de marcação na partida

O botão **Algo deu errado** aparece na mesa (`RoomMenu` e `DesktopRail`) somente
quando a sala conhece o `matchId` do plano 05. O fluxo é:

1. Clicar no botão abre a confirmação **Marcar que algo deu errado agora?**
   com **Marcar** e **Cancelar**. Cancelar não chama a API e não grava nada.
2. **Marcar** chama `POST /bugReports/flag` com `{matchId, turn}`. A rota exige
   login válido; 401 abre o login.
3. O backend cria o relato (`201`) ou acrescenta o turno ao relato que já
   existe para aquele jogador e partida (`200`). Repetir o mesmo turno é
   idempotente (`200`, sem duplicar); o 11º turno distinto vira `400`.
4. Só depois de `200`/`201` o front grava o rascunho no `localStorage`, na
   chave `hoodwink-bug-drafts`, como `{reportId, matchId, turns, at, userId}`.
   Erro de rede ou `429` mostra o aviso e **não** grava o rascunho.
5. Flags da mesma partida e do mesmo usuário viram um rascunho só (turnos
   únicos, até 10). O rascunho guarda o `userId` e é ignorado se outro usuário
   estiver logado.

No fim da partida, o **Finale** procura um rascunho daquela partida e mostra o
cartão **Você marcou que algo deu errado no turno X. O que aconteceu?**, com as
seis categorias em botões grandes, texto opcional e **Enviar** / **Pular**:

- **Enviar** faz `PATCH /bugReports/{id}` (só o dono) com categoria, descrição
  e contexto.
- **Pular** apaga só o rascunho local; o relato marcado continua no banco.

Se o jogo for fechado antes do fim, na próxima abertura o rascunho com menos de
7 dias aparece **uma vez** no Lobby com o mesmo cartão. `PATCH` `404`/`403`
apaga o rascunho e avisa **Esse relato não existe mais**; `429` ou falha de
rede mantêm o texto e o rascunho.

No Lobby, **Reportar bug** abre o mesmo cartão sem partida. Nesse caso a
categoria é obrigatória e a descrição precisa ter de 10 a 2000 caracteres.
Aceita anônimo: Bearer ausente, inválido ou vencido grava como anônimo e nunca
devolve `401`.

### Onde ficam os relatos

Ficam na tabela `bug_reports`, no mesmo Postgres do backend. A migration é
`server/backend/db/migrations/202611070004_bug_reports.sql`; o mesmo SQL
idempotente está em
`k8s/charts/hoodwink/templates/migrations-configmap.yaml`, e o model em
`server/backend/db/models/bugReport.go`.

| Coluna | Conteúdo |
| --- | --- |
| `id` | `bigserial` |
| `created_at`, `updated_at` | horários UTC |
| `user_id` | UUID do usuário; nulo em relato anônimo; `on delete set null` |
| `username` | nome do usuário, quando logado |
| `source` | `partida` ou `lobby` |
| `match_id` | UUID da partida (só em `partida`) |
| `turns` | turnos marcados, `int[]`, no máximo 10 |
| `category` | uma das 6 categorias ou nulo |
| `description` | texto livre até 2000 caracteres ou nulo |
| `context` | JSONB do contexto automático, até 8 KB |
| `ip` | IP do jogador pela regra `ClientIP` |
| `user_agent` | navegador |
| `status` | `novo` (padrão), `visto` ou `resolvido` |

`source`, `user_id`, `username`, `ip` e `user_agent` saem do servidor/token,
nunca do corpo. Há índices em `created_at desc`, `status` e `match_id`, além do
índice único parcial em `(user_id, match_id)` para relatos `partida` — é ele que
faz o `INSERT ... ON CONFLICT` somar turnos em vez de criar outra linha. O
backend não confere se o jogador estava mesmo na partida; o admin vê o
`match_id` e cruza depois.

### Aba "Bugs" em `/#admin`

A aba **Bugs** fica ao lado de **Requests** em `/#admin` e só aparece para
super-admin (a conta precisa de `users.is_super_admin = true`; usuário comum
recebe `403`). A lista vem em `created_at desc`, com contador de `novos`,
filtros de `status` e `source`, paginação Anterior/Próxima e detalhe com turnos,
categoria, descrição e contexto. O status pode ser trocado entre `novo`,
`visto` e `resolvido`. Tudo é renderizado como texto puro, nunca HTML.

| Rota | Uso |
| --- | --- |
| `POST /bugReports` | cria relato do Lobby; auth opcional |
| `POST /bugReports/flag` | marca um turno na partida; exige login |
| `PATCH /bugReports/{id}` | detalha o relato; só o dono e só `source=partida` |
| `GET /admin/bugReports` | lista para o super-admin; filtros `status`, `source`, `matchId`, `page` e `limit` (padrão 50, máximo 200) |
| `PATCH /admin/bugReports/{id}` | troca o `status`; devolve o relato atualizado |

O `GET /admin/bugReports` devolve `{items, total, novos, page, limit}`. `novos`
conta apenas `status='novo'` e não é escondido pelo filtro de status.

### Limites

| Limite | Valor | Onde fica |
| --- | --- | --- |
| Corpo HTTP | 16 KB (`413`) | `maxBugReportBody` e `http.MaxBytesReader` em `decodeBugBody` |
| Contexto | 8 KB (`400`) | `maxBugContextBytes` |
| Turnos por partida | 10 | `bug_reports_turns_max` no banco, `upsertTurnSQL` e `MAX_TURNS` em `drafts.ts` |
| Relatos `partida` por usuário | 50 em 24h (`429`) | `maxPartidaReports24h` e `partidaWindow` |
| Chamadas logadas | 10 por usuário a cada 10 min (`429`) | `bugUserRateLimit` e `bugRateWindow` |
| Chamadas anônimas | 5 por IP a cada 10 min (`429`) | `bugIPRateLimit` e `bugRateWindow` |
| `matchId` | UUID v7 com horário entre agora − 24h e agora + 5 min (`400`) | `validMatchID`, `partidaWindow` e `matchFutureTolerance` |
| Turno | inteiro de 0 a 10.000 (`400`) | `maxTurn` |
| Descrição | até 2000; no Lobby, no mínimo 10 | `maxBugDescription` e `minLobbyDescription` |

As constantes do backend ficam no bloco `const` de
`server/backend/routes/handlers/bugReportHandler.go`. Os limitadores são
criados em `server/backend/routes/handlers/handler.go` e implementados em
`server/backend/routes/handlers/rateLimit.go`: são em memória, com varredura e
teto de 10.000 chaves, valem **por pod** (`ponytail:`) e agrupam IPv6 por `/64`.
Se o jogo escalar para vários pods, troque por Redis. Os limites da listagem do
admin (`defaultBugLimit = 50` e `maxBugLimit = 200`) ficam em
`server/backend/routes/handlers/adminBugReportsHandler.go`.

### Contexto automático

O cartão tem a opção **Enviar dados do aparelho**, ligada por padrão; desmarcar
manda o contexto como nulo. Quando ligada, o front coleta:

- `scene`: cena/âncora atual (`data-scene`, hash, `cena` ou `sala`/`lobby`).
- `buildVersion`: `VITE_BUILD_VERSION` ou `VITE_APP_VERSION`; sem valor, `dev`.
- `screen`: tamanho da janela, por exemplo `390x844`.
- `errors`: os últimos 10 erros de JS de `window.onerror` e
  `unhandledrejection`, cada um cortado em 500 caracteres.

As URLs dos erros perdem query e hash, e `Bearer ...`, `token=` e `ticket=` são
apagados da mensagem e do stack. As mensagens da partida não vão, porque o
replay já guarda tudo.

O backend não confia nesse contexto: `sanitizeBugContext` passa o JSON por
`utils.MaskContext` antes de gravar, usando a mesma lista de chaves sensíveis
do `maskBody` do plano 04 (`server/backend/utils/maskBody.go`). No banco, o
contexto fica mascarado de novo, mesmo que o front tenha falhado.

### Categorias

As seis categorias são `moedas`, `carta`, `travou`, `regra`, `tela` e `outro`.
A lista fechada existe em três lugares e precisa mudar nos três:

- Front: `BUG_CATEGORIES` em `client/web/src/game/copy.ts` (o tipo
  `BugCategory` e os rótulos derivam dela).
- Handler: `bugCategories` e `validBugCategory` em
  `server/backend/routes/handlers/bugReportHandler.go`.
- Banco: o `CHECK (category IN (...))` da migration. O SQL do configmap repete
  o mesmo `check`, então atualize os dois.

### Apagar via SQL

Não há job de retenção para `bug_reports`; diferentemente de `request_logs`,
a limpeza é manual. No `make dev`, entre no Postgres com:

```bash
podman exec -it postgresdb psql -U postgres -d hoodwink
```

E apague o que precisar:

```sql
-- um relato específico
DELETE FROM bug_reports WHERE id = 42;

-- todos os relatos de uma partida
DELETE FROM bug_reports WHERE match_id = '0190f4f2-0000-7000-8000-000000000000';

-- retenção manual: mais de 30 dias
DELETE FROM bug_reports WHERE created_at < now() - interval '30 days';
```

Para conferir os mais recentes:

```sql
SELECT id, created_at, source, match_id, turns, category, status
FROM bug_reports
ORDER BY created_at DESC
LIMIT 20;
```

Apagar um usuário não apaga os relatos: a FK usa `on delete set null`, então a
linha continua com `user_id` nulo.

---

## 👨‍💻 Contato

Para dúvidas, sugestões ou contribuições, entre em contato com o mantenedor do projeto.

---


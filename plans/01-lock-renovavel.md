# PLANO 01: Lock da sala renovado enquanto a jogada processa
Status: concluído (validação manual pendente em make dev/VM)
Objetivo: o lock `lock:room:<id>` no Redis fica com o pod enquanto ele processa e é renovado em segundo plano; se o pod morre, o lock vence em poucos segundos e outro pod assume.
Fora do escopo: trocar o mutex local por sala (`roomMutexes`); fila de jogadas entre pods; Redlock com vários Redis; lock por player (`lock:player:*`).
Premissas e decisões: Clientes: não se aplica (repo sem `task clientes`; módulo tocado só o gameServer) · o lock passa a ser pego dentro de `redisFuncs.WithRoom`, ANTES do `LoadRoom` (hoje é pego depois de carregar a sala, então outro pod pode ler um estado velho) · todos os callers de WithRoom (jogada, gameProcessor, connect, disconnect) ganham o lock juntos, e saem os `AcquireRoomLock` soltos de `ProcessPlay` e `processRoomWithLock` · TTL 3s, renovado a cada 1s por goroutine (env `ROOM_LOCK_TTL_MS`, padrão 3000; intervalo = TTL/3) · renovar e soltar usam Lua com compare-and-set do dono (`GET==dono → PEXPIRE` / `GET==dono → DEL`); hoje o Get+Del separado pode apagar o lock de outro pod · se a renovação falhar (o lock foi perdido), o ctx passado ao fn é cancelado e o `SaveRoom` seguinte falha em vez de gravar por cima de outro pod · se o lock está com outro pod, WithRoom tenta de novo a cada 50ms por até 2s e só então desiste (hoje a jogada é descartada sem aviso) · `SaveRoomWithTTL` deixa de pegar lock próprio quando chamado dentro do WithRoom (o dono já é a instância) · testes usam `github.com/alicebob/miniredis/v2` só como dependência de teste (roda Lua e avança o relógio do TTL)
Pré-requisitos: Go 1.24+; `go get github.com/alicebob/miniredis/v2` no módulo gameServer

## Tarefas
- [x] T1 [Backend] Criar `AcquireRoomLock`/`renew`/`release` com Lua compare-and-set e goroutine de renovação que cancela o ctx ao perder o lock · arquivos: `server/gameServer/game/structs/rooms/roomPersistence.go`, `server/gameServer/game/structs/rooms/roomLock_test.go`, `server/gameServer/go.mod` · depende: — · verificar: `cd server/gameServer && go test ./game/structs/rooms/ -run Lock` cobre: renova além do TTL, outro dono não consegue pegar, release de não-dono não apaga, perda do lock cancela o ctx
- [x] T2 [Backend] Mover o lock para dentro de `WithRoom` (pegar antes do LoadRoom, retry 50ms até 2s, soltar depois do fn) e remover os locks soltos de `ProcessPlay` e `processRoomWithLock` · arquivos: `server/gameServer/redisFuncs/withRoom.go`, `server/gameServer/handlers/websocket/messages/processWebsocketMessage.go`, `server/gameServer/game/engine/gameProcessor.go`, `server/gameServer/game/structs/rooms/roomPersistence.go` · depende: T1 · verificar: `cd server/gameServer && go vet ./... && go test ./...` verde e `grep -rn "AcquireRoomLock(" server/gameServer --include=*.go` só aparece em withRoom.go, roomPersistence.go e testes
- [x] T3 [QA] Teste de WithRoom com miniredis: dois "pods" (instanceIDs diferentes) na mesma sala, o segundo espera o primeiro terminar um fn que dura 5s (> TTL) e vê o estado gravado por ele; pod que "morre" (fn nunca volta, renovação parada) perde o lock em até 3s · arquivos: `server/gameServer/redisFuncs/withRoom_test.go` · depende: T2 · verificar: `cd server/gameServer && go test ./redisFuncs/ -run WithRoom -count=3`
- [x] T4 [Docs] Documentar `ROOM_LOCK_TTL_MS` no `.env.example` e no values do chart · arquivos: `server/gameServer/.env.example`, `k8s/charts/hoodwink/values.yaml` · depende: T2 · verificar: `grep ROOM_LOCK_TTL_MS server/gameServer/.env.example k8s/charts/hoodwink/values.yaml` acha as duas linhas

## Aceite final
- [x] `cd server/gameServer && go vet ./... && go test ./...` verde
- [x] Teste do T3 prova que um processamento de 5s não perde o lock e que um pod morto libera a sala em até 3s (`go test ./redisFuncs/ -run WithRoom -v`)
- [ ] `make dev` + uma partida local de 3 jogadores até o fim sem log "falha ao adquirir lock da sala"

## Riscos
- Retry de 2s segura a goroutine de leitura do WebSocket daquele jogador → o gameProcessor processa no máximo 5 salas por tick, então a espera real é curta; logar quando passar de 500ms
- Cancelar o ctx no meio do Tick deixa a sala com efeito aplicado só em memória → o SaveRoom falha e nada é gravado; o próximo dono recarrega o estado anterior consistente
- Relógio do miniredis ≠ relógio real → os testes usam `FastForward` para o TTL e nunca `time.Sleep` longo

Esforço: M

## Revisão de implementação (2026-11-06)

- Entregue: lock `lock:room:<id>` adquirido dentro do `WithRoom` (antes do `LoadRoom`), TTL 3s renovado a cada 1s por Lua compare-and-set; retry de 50ms por até 2s; perda do lock cancela o ctx e o `SaveRoom` falha.
- `ROOM_LOCK_TTL_MS` documentado no `.env.example` do gameServer e no values do chart.
- Testes com `miniredis` (dependência só de teste) cobrem renovação além do TTL, release de não-dono, dois pods na mesma sala e pod morto liberando em até 3s.
- Verdes: `cd server/gameServer && go vet ./... && go test ./...` e `go test ./redisFuncs/ -run WithRoom -count=3`.
- Pendente (manual): `make dev` com partida local de 3 jogadores até o fim, sem log de falha ao adquirir lock.

# Relatório — 12-arrays-postgres-producao

Plano: `plans/12-arrays-postgres-producao.md` · branch `plano/12-arrays-postgres-producao` · status: concluído

## Concluído
- T1/T2: `StringArray`/`Int32Array` (`db/models/array.go`) com `Scan`/`Value`; models passam a usá-los e os stores/DTOs convertem para `[]string`/`[]int32`.
- T2: `gormBugReportStore.Create` virou SQL cru (`insertLobbyBugSQL`) com `turns` no default e `RETURNING id, created_at, updated_at`.
- T3: `routes/handlers/postgresArrays_test.go` cobre SELECT de `uuid[]`/`int[]`, Create do Lobby com `turns = {}` e `UpsertTurn` repetido.
- T4: seção de arrays no `server/README.md` (sintoma, cast, como rodar o teste).

## Validação
- `go build ./... && go vet ./...` ok; `go test ./...` verde.
- `HW_TEST_PG=<dsn local> go test ./routes/handlers -run TestPostgresArrays` verde; sem a variável o teste é pulado.

## Premissas e ressalvas
- Repo sem `Taskfile`/`task`: `task modules:generate` não existe e `task clientes` não se aplica (já registrado nas Premissas do plano).
- Postgres local tinha só `users`: as migrations 003/004 foram aplicadas à mão com um programa descartável (sem `goose` instalado). O programa foi removido antes do commit.
- Conferência em produção continua manual, depende do deploy da branch `production`.

## Erros corrigidos
- `unsupported Scan ... into type *[]string`/`*[]int32`: o driver devolve arrays como literal `{a,b}`; resolvido com `sql.Scanner` nos models.
- `Create` do Lobby por `db.Create` gravaria `turns` como `text`; agora é SQL cru sem a coluna de array.

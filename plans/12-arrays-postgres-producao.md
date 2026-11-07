# PLANO 12: Corrige leitura de arrays do Postgres (relato de bug e replay)
Status: pendente
Objetivo: Fazer o backend ler `uuid[]` e `int[]` corretamente, para `/me/matches` e `/bugReports/flag` pararem de responder 500 e o relato já gravado voltar a aparecer.
Fora do escopo: mudar o schema ou as migrations; alterar as regras de relato; mudar o formato `.hwr`; redesenhar as telas de perfil e de bug.
Premissas e decisões: Clientes: não se aplica (repo sem `task clientes`; módulos tocados: backend) · causa confirmada em produção: o driver devolve `player_ids` e `turns` como `string` e o GORM falha em `[]string` e `[]int32` · o relato e o replay são gravados; o 500 acontece na leitura de volta (`RETURNING` e `SELECT`) · a correção é um tipo de slice com `sql.Scanner` e `driver.Valuer` para `uuid[]` e outro para `int[]` · `Value()` devolve `nil` para slice nil, `{}` para vazio e elementos entre aspas quando tiverem vírgula/aspas; `Scan(nil)` deixa o slice nil · só os campos dos models mudam de tipo; `bugCreateStore.Create`, `bugFlagStore.UpsertTurn` e os DTOs continuam com `[]int32`/`[]string` e convertem no store · nenhuma escrita de `player_ids`/`turns` pode passar por `db.Create`, porque o `driver.Valuer` devolve `text` e o Postgres recusa em `uuid[]`/`int[]`; `gormBugReportStore.Create` passa a ser SQL cru com `turns` no default · nada de variável de ambiente nova nem dependência nova · o `Insert` do replay continua usando o literal `{a,b}` com `?::uuid[]` · o relato idempotente passa a devolver 201/200 em vez de 500 · a verificação em produção fica manual após o deploy · Revisado em 2026-11-07: 9 correções aplicadas.
Pré-requisitos: Go 1.26.5; Postgres local (`make infra-up` + `make -C server migrate`) para o teste de integração; `HW_TEST_PG` é o DSN do `server/backend/.env`, passado inline no comando e nunca exportado, para o `go test ./...` continuar pulando o teste; acesso ao cluster para a conferência manual final.

## Tarefas
- [x] T1 [Backend] Criar `StringArray` e `Int32Array` com `Scan`/`Value` tolerantes a `string`, `[]byte` e `nil`, com o contrato de nil/vazio/aspas · arquivos: `server/backend/db/models/array.go`, `server/backend/db/models/array_test.go` · depende: — · verificar: `cd server/backend && go test ./db/models`
- [x] T2 [Backend] Usar os tipos novos só nos models, converter nos stores/DTOs, trocar `Create` do relato por SQL cru e ajustar testes · arquivos: `server/backend/db/models/matchReplay.go`, `server/backend/db/models/bugReport.go`, `server/backend/routes/handlers/replaysHandler.go`, `server/backend/routes/handlers/bugReportHandler.go`, `server/backend/routes/handlers/replaysHandler_test.go`, `server/backend/routes/handlers/bugReportHandler_test.go` · depende: T1 · verificar: `cd server/backend && go test ./...`
- [ ] T3 [QA] Adicionar teste de integração com Postgres real que cobre scan de `uuid[]`/`int[]`, `Create` do lobby com turns vazio e `UpsertTurn` repetido · arquivos: `server/backend/routes/handlers/postgresArrays_test.go` · depende: T2 · verificar: `cd server/backend && HW_TEST_PG=<dsn> go test ./routes/handlers -run TestPostgresArrays`
- [ ] T4 [Docs] Registrar no README do backend o sintoma `unsupported Scan`, o SQL cru com cast e como rodar o teste · arquivos: `server/README.md` · depende: T3 · verificar: `grep -n "unsupported Scan" server/README.md`

## Aceite final
- [ ] `cd server/backend && go test ./...` verde.
- [ ] `cd server/backend && HW_TEST_PG=<dsn> go test ./routes/handlers -run TestPostgresArrays` verde e, sem `HW_TEST_PG`, o teste é pulado.
- [ ] O teste de integração lê `'{...}'::uuid[]` e `'{1,2}'::int[]` pelo `SELECT` do GORM e confere o conteúdo; o scan quebra se o campo voltar a slice puro.
- [ ] O teste de integração cria um relato do lobby com turns vazio e confirma que a linha volta com `turns = {}`.
- [ ] `UpsertTurn` chamado duas vezes com o mesmo turno devolve a mesma linha, sem duplicar.
- [ ] Conferência manual do dono, após o deploy: o perfil lista as partidas e "Algo deu errado" no mesmo turno devolve `200` sem duplicar.

## Riscos
- O parser caseiro quebrar com elemento entre aspas ou array vazio: cobrir `{}`, `{a,b}`, `{"a","b"}`, `nil` e lixo no teste unitário.
- Nenhum caminho pode gravar `turns`/`player_ids` por `db.Create`; toda escrita usa SQL cru com cast (`?::uuid[]`, `ARRAY[?]::int[]`).
- O teste de integração gravar em banco compartilhado: usar match e usuário de teste, semear com SQL cru e apagar no `t.Cleanup`.
- A produção continuar com a imagem antiga: a conferência final é manual e depende do deploy da branch `production`.

Esforço: M

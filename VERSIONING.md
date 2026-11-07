# Versionamento do jogo

A versão do **jogo** é independente da versão do **motor de replay**.

| Versão | Fonte | Formato | Sobe quando |
| --- | --- | --- | --- |
| Jogo | `VERSION` (raiz) | semver `MAJOR.MINOR.PATCH` | muda o produto (regras, telas, backend, deploy) |
| Motor de replay | `server/gameServer/game/ENGINE_VERSION` | inteiro `1, 2, 3…` | muda qualquer coisa que altere a reprodução de uma partida |

Uma partida grava as duas: `engine_version` (motor usado para reproduzir) e
`game_version` (build que estava no ar). A versão do jogo não precisa subir a
cada mudança no motor, e vice-versa.

## Fonte da verdade

`VERSION` na raiz do repositório (`0.1.0`). Nada mais é editado à mão: o build
injeta o valor nos binários e no bundle web. `client/web/package.json` espelha o
mesmo número para o npm e o script de bump mantém os dois juntos.

## Como subir a versão

```bash
make -C server game-bump level=patch   # padrão: patch
make -C server game-bump level=minor
make -C server game-bump level=major
make -C server version                  # mostra a versão atual
make -C server version-check            # valida VERSION + package.json
```

Para um release, a tag precisa bater com o `VERSION`:

```bash
git tag v0.1.0
make -C server version-check tag=v0.1.0
git push origin v0.1.0
```

## Para onde a versão flui

- **Go (backend e gameServer)**: injetada por `-ldflags -X .../version.Version`
  no Dockerfile e no `make build`. Sem ldflags (`go run`, `go test`) fica `dev`.
- **API**: `GET /version` no backend e no gameServer responde
  `{ "version": "0.1.0", "commit": "<sha>" }`.
- **Web**: `VITE_GAME_VERSION` (Vite lê o `VERSION` da raiz no dev/CI; o Docker
  recebe por build-arg). Aparece em Preferências no Lobby e vai no contexto dos
  relatos de bug (`buildVersion`).
- **Partidas**: o gameServer envia `game_version` no `POST /internal/replays`;
  o backend grava em `match_replays.game_version` e devolve em `/me/matches`.
  Partidas antigas ficam `unknown`.
- **Deploy**: o CI lê `VERSION`, valida a tag e passa a versão para as três
  imagens (backend, gameServer, web).

A versão do **chart** (`k8s/charts/hoodwink/Chart.yaml`) e a tag das imagens
continuam sendo coisas do deploy (tag = SHA do commit), não a versão do jogo.

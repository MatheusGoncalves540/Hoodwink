# Caos do Hoodwink

Harness sem dependências para testar o gameServer e a interface em situação de caos.
Lê os `.env` locais do backend e do gameServer para emitir JWTs de teste — não use em produção.

## Pré-requisitos

- backend em `http://localhost:8088` e gameServer em `http://localhost:4444`
- `D:/GitHub/Hoodwink-server/backend/.env` e `D:/GitHub/Hoodwink-server/gameServer/.env`
  (ou aponte `HOODWINK_BACKEND_ENV` / `HOODWINK_GAME_ENV`)

## Rodar

```sh
node tests/caos/suite.mjs      # 21 rodadas HTTP/WS; imprime JSON e o resumo
node tests/caos/smoke.mjs      # cria sala, conecta 3 jogadores e imprime o estado
node tests/caos/match.mjs      # bot de partida completa (pode não terminar com mãos sem ataque)
node tests/caos/start-test.mjs # reproduz o start por quem não é anfitrião
```

As salas são descartáveis e os jogadores usam ids com sufixo de execução.
`*.log`, `*.err` e `*.json` são artefatos locais e não entram no git.

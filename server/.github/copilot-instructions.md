# Copilot Instructions for Hoodwink

## Visão Geral da Arquitetura
- O projeto é dividido em dois principais serviços: `backend` (API REST, autenticação, persistência) e `gameServer` (WebSocket, lógica de jogo em tempo real).
- Comunicação entre serviços ocorre via Redis (Pub/Sub, registro de jogadores em sala, eventos de jogo).
- O backend utiliza PostgreSQL para persistência de dados e Redis para cache/eventos.
- O gameServer gerencia salas, jogadores, e eventos do jogo, com autenticação JWT.

## Fluxos e Convenções
- O backend e o gameServer possuem seus próprios arquivos `main.go` e inicialização de rotas, serviços e handlers.
- O gameServer utiliza JWT para autenticação de jogadores nas salas. Tokens podem ser passados via query string.
- O upgrade para WebSocket é feito em `gameServer/routes/rHandlers/websocket.go` usando um `websocket.Upgrader` customizado. O CORS pode ser desabilitado via env `CORS=false`.
- Após o upgrade para WebSocket, nunca escreva na resposta HTTP (`WriteHeader`, `Write`). Trate erros antes do upgrade.
- O middleware de recovery (`gameServer/routes/middlewares/recovery.go`) implementa `http.ResponseWriter` e `http.Hijacker` para compatibilidade com WebSocket.

## Build, Testes e Execução
- Use os comandos do `makefile` no diretório raiz para subir dependências:
  - `make db-up` para PostgreSQL
  - `make redis-up` para Redis
  - `make commander-up` para Redis Commander
  - `make start` para iniciar todos os containers
  - `make stop` para parar todos os containers
  - `make clean` para remover todos os containers
- O backend e o gameServer podem ser executados separadamente, cada um ouvindo na porta definida por `PORT` no `.env`.

## Padrões Específicos
- Handlers de WebSocket sempre validam JWT e registro do jogador antes do upgrade.
- Eventos do jogo são processados e publicados via Redis Pub/Sub para sincronização entre instâncias.
- O código utiliza convenções de erro explícitas e logging detalhado para rastreabilidade.
- Estruturas de dados do jogo (cartas, jogadores, eventos) estão em `gameServer/game/room/roomStructs`.
- Utilitários e helpers estão em `gameServer/utils` e `backend/utils`.

## Exemplos de Integração
- Para adicionar uma nova ação de jogo, crie um handler em `gameServer/game/room/handlers` e registre no processador de eventos.
- Para autenticação customizada, edite `backend/routes/auth` ou `gameServer/routes/rHandlers/wsHandlers/connectionValidator.go`.
- Para novas integrações com Redis, utilize helpers em `gameServer/game/room/redisHandlers`.

## Observações
- Sempre valide dados críticos antes de upgrades de conexão ou operações sensíveis.
- Siga os exemplos de logging e tratamento de erro para manter rastreabilidade.
- Consulte os arquivos `README.md` para regras do jogo e comandos de desenvolvimento.
- SEMPRE APLIQUE AS MODIFICAÇÕES RECOMENDADAS NOS ARQUIVOS

---

Seções incompletas ou dúvidas? Solicite exemplos ou detalhes de fluxos específicos para melhorar esta instrução.

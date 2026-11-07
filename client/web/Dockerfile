# Build stage
FROM node:22-slim AS builder
WORKDIR /app

# Pública Ed25519 de produção (hex) e kid, passadas pelo deploy da VM.
ARG VITE_REPLAY_PUBLIC_KEY=""
ARG VITE_REPLAY_KID="replay-v1"
ENV VITE_REPLAY_PUBLIC_KEY=$VITE_REPLAY_PUBLIC_KEY
ENV VITE_REPLAY_KID=$VITE_REPLAY_KID

# Versão do jogo (VERSION na raiz). O contexto do Docker é client/web, então a
# raiz não entra; sem este ARG o Vite usa "dev".
ARG VITE_GAME_VERSION=""
ENV VITE_GAME_VERSION=$VITE_GAME_VERSION

# Copia package lock e package para instalar dependências
COPY package.json package-lock.json ./
RUN npm ci

# Assets do replay WASM, gerados por `make replay-wasm` em server/: um
# replay-v<N>-<sha>.wasm.br por versão do motor + wasm_exec.js. Copiados antes
# do restante para aproveitar a camada de cache; o build falha cedo (e não no
# navegador) se os assets não foram gerados.
COPY public/replay ./public/replay
RUN test -s public/replay/wasm_exec.js && ls public/replay/replay-v*.wasm.br >/dev/null 2>&1 || { \
      echo 'ERRO: assets do replay ausentes em public/replay. Rode `make replay-wasm` em server/ antes do docker build.'; \
      exit 1; \
    }

# Copia o restante da aplicação e gera a build de produção
# (client/web/.env.production define as URLs públicas do domínio)
COPY . .
RUN test -s src/replay/manifest.gen.ts || { \
      echo 'ERRO: src/replay/manifest.gen.ts ausente. Rode `make replay-wasm` em server/ antes do docker build.'; \
      exit 1; \
    }
RUN npm run build

# Runtime stage
FROM nginx:stable-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html
# Falha no build (e não em CrashLoopBackOff por 10 min no helm) se o config for inválido.
RUN nginx -t

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

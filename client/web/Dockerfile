# Build stage
FROM node:22-slim AS builder
WORKDIR /app

# Copia package lock e package para instalar dependências
COPY package.json package-lock.json ./
RUN npm ci

# Copia o restante da aplicação e gera a build de produção
# (client/web/.env.production define as URLs públicas do domínio)
COPY . .
RUN npm run build

# Runtime stage
FROM nginx:stable-alpine
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]

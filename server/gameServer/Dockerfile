# Build a partir da raiz do monorepo server/ (o gameServer depende do módulo local replay):
#   sudo docker build -f server/gameServer/Dockerfile -t hoodwink-gameserver server
#   podman run -d --name hoodwink-gameserver --restart=always -p 4444:4444 --env-file .env hoodwink-gameserver

#############

# Multi-stage build for lightweight and secure image
# Build stage
FROM docker.io/library/golang:1.26.5-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files dos módulos locais (replace ../replay)
COPY replay/go.mod replay/go.sum ./replay/
COPY gameServer/go.mod gameServer/go.sum ./gameServer/

# Download dependencies
RUN cd gameServer && go mod download

# Copy source code
COPY replay ./replay
COPY gameServer ./gameServer

# Build the application with optimizations. A versão do jogo vem do deploy
# (VERSION na raiz); sem ela o binário reporta "dev".
ARG GAME_VERSION="dev"
ARG GAME_COMMIT="unknown"
RUN cd gameServer && CGO_ENABLED=0 GOOS=linux go build -trimpath \
      -ldflags="-w -s -X github.com/MatheusGoncalves540/Hoodwink-gameServer/version.Version=${GAME_VERSION} -X github.com/MatheusGoncalves540/Hoodwink-gameServer/version.Commit=${GAME_COMMIT}" \
      -o /app/main .

# Runtime stage
FROM docker.io/library/alpine:latest

# Install ca-certificates for HTTPS and curl for healthcheck
RUN apk --no-cache add ca-certificates curl

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /home/appuser/

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy game assets (rules, etc.) from builder stage
COPY --from=builder /app/gameServer/game /home/appuser/game

# Change ownership to non-root user
RUN chown -R appuser:appgroup main game

# Switch to non-root user
USER appuser

# Expose port (configurable via PORT env var)
EXPOSE 4444

# Run the application
CMD ["./main"]

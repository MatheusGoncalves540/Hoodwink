# syntax=docker/dockerfile:1.7
# Build a partir da raiz do monorepo server/ (o backend depende do módulo local replay):
#   sudo docker build -f server/backend/Dockerfile -t hoodwink-backend server
#   podman run -d --name hoodwink-backend --restart=always -p 8088:8088 --env-file .env hoodwink-backend

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
COPY backend/go.mod backend/go.sum ./backend/

# Download dependencies (cache de módulos entre builds)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd backend && go mod download

# Copy source code
COPY replay ./replay
COPY backend ./backend

# Build the application with optimizations (cache de compilação entre builds)
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd backend && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o /app/main .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /home/appuser/

# Copy binary from builder stage
COPY --from=builder /app/main .

# Change ownership to non-root user
RUN chown appuser:appgroup main

# Switch to non-root user
USER appuser

# Expose port (configurable via PORT env var)
EXPOSE 8088

# Run the application
CMD ["./main"]

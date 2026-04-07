# Multi-stage build for lightweight and secure image
# Build stage
FROM docker.io/library/golang:1.24.4-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main .

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
COPY --from=builder /app/game /home/appuser/game

# Change ownership to non-root user
RUN chown -R appuser:appgroup main game

# Switch to non-root user
USER appuser

# Expose port (configurable via PORT env var)
EXPOSE 4444

# Run the application
CMD ["./main"]
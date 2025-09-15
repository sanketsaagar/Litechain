# Multi-stage build for LightChain L1
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build

# Final stage - lightweight runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 lightchain && \
    adduser -D -s /bin/sh -u 1001 -G lightchain lightchain

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/build/lightchain /app/lightchain

# Copy configuration files
COPY --from=builder /app/configs /app/configs

# Create data directory
RUN mkdir -p /app/data && \
    chown -R lightchain:lightchain /app

# Switch to non-root user
USER lightchain

# Expose ports
EXPOSE 8545 8546 30303 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8545 || exit 1

# Set entrypoint
ENTRYPOINT ["/app/lightchain"]

# Default command
CMD ["--config", "/app/configs/validator.yaml"]

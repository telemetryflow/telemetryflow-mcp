# ==============================================================================
# TelemetryFlow GO MCP Server Dockerfile
# Multi-stage build for minimal production image
# ==============================================================================

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Set GOPRIVATE to bypass checksum database for telemetryflow SDK
ENV GOPRIVATE=github.com/telemetryflow/*

# Copy go module files first for better caching
COPY go.mod go.sum ./

# Download dependencies (leverages Docker layer caching)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments
ARG VERSION=1.1.2
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" \
    -o tfo-mcp \
    ./cmd/mcp

# ==============================================================================
# Production stage
# ==============================================================================
FROM alpine:3.21

# =============================================================================
# TelemetryFlow Metadata Labels (OCI Image Spec)
# =============================================================================
LABEL org.opencontainers.image.title="TelemetryFlow GO MCP Server" \
    org.opencontainers.image.description="Enterprise MCP (Model Context Protocol) server for AI-powered observability - Community Enterprise Observability Platform (CEOP)" \
    org.opencontainers.image.version="1.1.2" \
    org.opencontainers.image.vendor="TelemetryFlow" \
    org.opencontainers.image.authors="DevOpsCorner Indonesia <support@devopscorner.id>" \
    org.opencontainers.image.url="https://telemetryflow.id" \
    org.opencontainers.image.documentation="https://docs.telemetryflow.id" \
    org.opencontainers.image.source="https://github.com/telemetryflow/telemetryflow-go-mcp" \
    org.opencontainers.image.licenses="Apache-2.0" \
    org.opencontainers.image.base.name="alpine:3.21" \
    # TelemetryFlow specific labels
    io.telemetryflow.product="TelemetryFlow GO MCP Server" \
    io.telemetryflow.component="tfo-mcp" \
    io.telemetryflow.platform="CEOP" \
    io.telemetryflow.maintainer="DevOpsCorner Indonesia"

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -S telemetryflow && adduser -S telemetryflow -G telemetryflow

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/tfo-mcp /app/tfo-mcp

# Copy default config
COPY configs/tfo-mcp.yaml /app/configs/tfo-mcp.yaml

# Set ownership
RUN chown -R telemetryflow:telemetryflow /app

# Switch to non-root user
USER telemetryflow

# =============================================================================
# Environment Variables (synchronized with .env.example)
# =============================================================================

# Claude API
ENV ANTHROPIC_API_KEY=""

# TelemetryFlow Observability (TelemetryFlow SDK)
ENV TELEMETRYFLOW_API_KEY=""
ENV TELEMETRYFLOW_ENDPOINT="https://api.telemetryflow.io"
ENV TELEMETRYFLOW_SERVICE_NAME="telemetryflow-go-mcp"
ENV TELEMETRYFLOW_SERVICE_VERSION="1.1.2"
ENV TELEMETRYFLOW_ENVIRONMENT="production"

# Server Configuration
ENV TELEMETRYFLOW_MCP_SERVER_HOST="0.0.0.0"
ENV TELEMETRYFLOW_MCP_SERVER_PORT="8080"
ENV TELEMETRYFLOW_MCP_SERVER_TRANSPORT="stdio"
ENV TELEMETRYFLOW_MCP_DEBUG="false"

# Logging
ENV TELEMETRYFLOW_MCP_LOG_LEVEL="info"
ENV TELEMETRYFLOW_MCP_LOG_FORMAT="json"

# Redis (Caching & Queue)
ENV TELEMETRYFLOW_MCP_REDIS_URL="redis://localhost:6379"
ENV TELEMETRYFLOW_MCP_CACHE_ENABLED="true"
ENV TELEMETRYFLOW_MCP_CACHE_TTL="300"
ENV TELEMETRYFLOW_MCP_QUEUE_ENABLED="true"
ENV TELEMETRYFLOW_MCP_QUEUE_CONCURRENCY="5"

# Database (PostgreSQL)
ENV TELEMETRYFLOW_MCP_POSTGRES_URL=""
ENV TELEMETRYFLOW_MCP_POSTGRES_MAX_CONNS="25"
ENV TELEMETRYFLOW_MCP_POSTGRES_MIN_CONNS="5"

# Analytics Database (ClickHouse)
ENV TELEMETRYFLOW_MCP_CLICKHOUSE_URL=""

# OpenTelemetry (Fallback)
ENV TELEMETRYFLOW_MCP_TELEMETRY_ENABLED="true"
ENV TELEMETRYFLOW_MCP_OTLP_ENDPOINT="localhost:4317"
ENV TELEMETRYFLOW_MCP_SERVICE_NAME="telemetryflow-go-mcp"

# Health check (for SSE/WebSocket modes)
# HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
#     CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Expose port (for SSE/WebSocket modes)
EXPOSE 8080

# Entry point
ENTRYPOINT ["/app/tfo-mcp"]

# Default command
CMD ["--config", "/app/configs/tfo-mcp.yaml"]

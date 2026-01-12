<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-mcp-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-mcp-light.svg">
    <img src="https://github.com/telemetryflow/.github/raw/main/docs/assets/tfo-logo-mcp-light.svg" alt="TelemetryFlow Logo" width="80%">
  </picture>

  <h3>TelemetryFlow GO MCP Server (TFO-GO-MCP)</h3>

[![Version](https://img.shields.io/badge/Version-1.1.2-orange.svg)](CHANGELOG.md)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go)](https://golang.org/)
[![MCP Protocol](https://img.shields.io/badge/MCP-2024--11--05-purple?logo=data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgZmlsbD0ibm9uZSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCBkPSJNMTIgMkM2LjQ4IDIgMiA2LjQ4IDIgMTJzNC40OCAxMCAxMCAxMCAxMC00LjQ4IDEwLTEwUzE3LjUyIDIgMTIgMnoiIGZpbGw9IiNmZmYiLz48L3N2Zz4=)](https://modelcontextprotocol.io/)
[![Claude API](https://img.shields.io/badge/Claude-Opus%204%20%7C%20Sonnet%204-E1BEE7?logo=anthropic)](https://anthropic.com)
[![OTEL SDK](https://img.shields.io/badge/OpenTelemetry_SDK-1.39.0-blueviolet)](https://opentelemetry.io/)
[![Architecture](https://img.shields.io/badge/Architecture-DDD%2FCQRS-success)](docs/ARCHITECTURE.md)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?logo=postgresql)](https://www.postgresql.org/)
[![ClickHouse](https://img.shields.io/badge/ClickHouse-23+-FFCC00?logo=clickhouse)](https://clickhouse.com/)

</div>

---

# Changelog

All notable changes to TelemetryFlow GO MCP Server (TFO-GO-MCP) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.2] - 2025-01-09

### Added

#### Database Infrastructure

- PostgreSQL database support with GORM ORM
  - Session repository for persistent session storage
  - Conversation repository for message history
  - Database models with JSONB support for flexible data storage
  - Connection pooling and health checks
  - Custom GORM types: `JSONB`, `JSONBArray`, `StringArray`
- ClickHouse analytics database support
  - Tool call analytics table with MergeTree engine
  - API request analytics table for Claude API usage
  - Session analytics table for lifecycle events
  - Error analytics table for debugging
  - Time series aggregation queries with percentiles (p50/p95/p99)
  - Hourly aggregation tables (SummingMergeTree, ReplacingMergeTree)
  - Materialized views for real-time aggregations
  - Batch insert support for high throughput

#### Analytics & Dashboard

- Analytics repository with comprehensive dashboard queries
  - Token usage statistics by model
  - Tool usage statistics with p50/p95/p99 percentiles
  - Request/token/latency time series
  - Error rate tracking over time
  - Dashboard summary with key metrics

#### Migration & Seeding Infrastructure

- Database migration infrastructure
  - Versioned migrations for PostgreSQL and ClickHouse
  - Migration runner with up/down/reset/fresh operations
  - Migration status tracking via `schema_migrations` table
  - GORM AutoMigrate integration
- Database seeding infrastructure
  - Idempotent seeders using FirstOrCreate pattern
  - Default seeders: tools, resources, prompts, api_keys, demo_session
  - Production-safe seeder subset (excludes demo data)
  - SeederResult tracking for executed/skipped/failed seeders

#### Caching & Queue Infrastructure

- Redis caching infrastructure
  - Cache service with TTL support
  - Session and conversation caching
  - Cache invalidation strategies
- NATS JetStream queue infrastructure
  - Durable message queuing
  - Publisher/subscriber pattern
  - Task acknowledgment with retry support

#### Testing & CI/CD

- Comprehensive test suite
  - Session aggregate tests
  - Conversation aggregate tests
  - Tool entity tests
  - MCP protocol tests
  - Migration and seeder tests
  - GORM model tests
  - Benchmarks for performance testing
- CI-specific Makefile targets for GitHub Actions
  - `test-unit-ci`: Unit tests with coverage output
  - `test-integration-ci`: Integration tests with coverage
  - `test-e2e-ci`: End-to-end tests
  - `ci-build`: Cross-platform CI builds with GOOS/GOARCH
  - `ci-test`: Combined format, vet, lint, test pipeline
  - `test-all`: Run all test types
  - `deps-verify`: Dependency verification
  - `staticcheck`: Static analysis
  - `govulncheck`: Vulnerability scanning
  - `coverage-report`: Merged coverage report generation
- GitHub Actions CI/CD workflows
  - Lint, test, build pipeline
  - Multi-platform build support (Linux, macOS, Windows)
  - Security scanning with Gosec and govulncheck
  - Coverage reporting with merged reports

### Changed

- **BREAKING**: Refactored all environment variable keys from `TFO_*` to `TELEMETRYFLOW_*`
  - `TFO_CLAUDE_API_KEY` → `TELEMETRYFLOW_MCP_CLAUDE_API_KEY`
  - `TFO_LOG_LEVEL` → `TELEMETRYFLOW_MCP_LOG_LEVEL`
  - `TFO_SERVER_HOST` → `TELEMETRYFLOW_MCP_SERVER_HOST`
  - `TFO_SERVER_PORT` → `TELEMETRYFLOW_MCP_SERVER_PORT`
  - All other `TFO_*` variables follow the same pattern
- Updated go.mod with GORM, ClickHouse, Redis, and NATS dependencies
- Enhanced configuration with database, cache, and queue settings
- Updated Viper SetEnvPrefix from `TFO_MCP` to `TELEMETRYFLOW_MCP`
- Improved error handling in analytics repository with proper resource cleanup
- Enhanced event publishing with explicit error ignoring for best-effort delivery

### Fixed

- Fixed all golangci-lint errors (0 issues)
  - Added proper error handling for `rows.Close()` in analytics repository
  - Added explicit error ignoring for event publisher calls (best-effort delivery)
  - Fixed empty branches in test assertions
  - Removed unused variables in server tests
- Fixed anthropic SDK API compatibility in client.go
- Fixed observability.go StartSpan return type

---

## [1.1.1] - 2025-01-05

### Added

- Initial project structure
- Basic MCP protocol support
- Claude API integration prototype

### Changed

- Refined DDD architecture

---

## [1.1.0] - 2025-01-01

### Added

- Project inception
- Architecture design
- Technology stack selection

---

## Version History Summary

| Version | Date       | Highlights                                                     |
| ------- | ---------- | -------------------------------------------------------------- |
| 1.1.2   | 2025-01-09 | Database infrastructure, CI/CD, analytics, comprehensive tests |
| 1.1.1   | 2025-01-05 | Initial structure, basic protocol support                      |
| 1.1.0   | 2025-01-01 | Project inception                                              |

---

## Migration Guide

### Upgrading to 1.1.2

No breaking changes. Update your binary and restart the server.

```bash
# Download new version
curl -LO https://github.com/telemetryflow/telemetryflow-go-mcp/releases/download/v1.1.2/tfo-mcp_$(uname -s)_$(uname -m).tar.gz

# Extract and install
tar -xzf tfo-mcp_*.tar.gz
sudo mv tfo-mcp /usr/local/bin/

# Verify
tfo-mcp version
```

---

## Links

- [GitHub Repository](https://github.com/telemetryflow/telemetryflow-go-mcp)
- [Documentation](docs/README.md)
- [Issue Tracker](https://github.com/telemetryflow/telemetryflow-go-mcp/issues)

[Unreleased]: https://github.com/telemetryflow/telemetryflow-go-mcp/compare/v1.1.2...HEAD
[1.1.2]: https://github.com/telemetryflow/telemetryflow-go-mcp/compare/v1.1.1...v1.1.2
[1.1.1]: https://github.com/telemetryflow/telemetryflow-go-mcp/compare/v1.1.0...v1.1.1
[1.1.0]: https://github.com/telemetryflow/telemetryflow-go-mcp/releases/tag/v1.1.0

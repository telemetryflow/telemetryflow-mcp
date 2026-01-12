# TFO-MCP Technology Stack

## Overview

TelemetryFlow GO MCP Server is built with Go following Domain-Driven Design patterns and the TelemetryFlow technology standards.

## Core Stack

| Technology        | Version       | Purpose            |
| ----------------- | ------------- | ------------------ |
| Go                | 1.24+         | Primary language   |
| anthropic-sdk-go  | v0.2.0-beta.3 | Claude API client  |
| OpenTelemetry SDK | v1.39.0       | Observability      |
| Zerolog           | v1.33.0       | Structured logging |
| Cobra             | v1.8.1        | CLI framework      |
| Viper             | v1.19.0       | Configuration      |

## Architecture Patterns

### Domain-Driven Design (DDD)

- **Aggregates**: Session, Conversation
- **Entities**: Message, Tool, Resource, Prompt
- **Value Objects**: IDs, Content types, MCP types
- **Domain Events**: Session lifecycle, Message events
- **Repository Interfaces**: Abstract persistence

### CQRS (Command Query Responsibility Segregation)

- **Commands**: Write operations (InitializeSession, SendMessage, ExecuteTool)
- **Queries**: Read operations (ListTools, GetResource, ListPrompts)
- **Handlers**: Business logic orchestration

### Clean Architecture

```
Presentation → Application → Domain ← Infrastructure
```

## MCP Protocol

- **Version**: 2024-11-05
- **Transport**: JSON-RPC 2.0 over stdio
- **Capabilities**: Tools, Resources, Prompts, Logging

## Claude API Integration

- **Models**: Claude 4 Opus, Claude 4 Sonnet, Claude 3.5 Sonnet/Haiku
- **Features**: Streaming, Tool use, Multi-turn conversations
- **Authentication**: API key via environment variable

## Observability

- **Tracing**: OpenTelemetry with OTLP export
- **Logging**: Zerolog (JSON/text format)
- **Metrics**: Prometheus-compatible (planned)

## Configuration

- **Format**: YAML with environment variable overrides
- **Manager**: Viper with automatic env binding
- **Validation**: Struct-based validation

## Security

- **API Keys**: Environment variable storage
- **Rate Limiting**: Configurable per-minute limits
- **CORS**: Configurable allowed origins

## Testing

- **Unit Tests**: Table-driven tests
- **Integration Tests**: Mock clients
- **Coverage Target**: ≥80%

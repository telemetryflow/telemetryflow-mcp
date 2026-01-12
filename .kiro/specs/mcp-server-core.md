---
title: TelemetryFlow GO MCP Server Core
version: 1.0.0
status: draft
created: 2025-01-09
updated: 2025-01-09
author: TelemetryFlow Team
inclusion: manual
---

# TelemetryFlow GO MCP Server Core Specification

## Overview

This specification defines the core MCP (Model Context Protocol) server implementation for TelemetryFlow, providing AI-powered capabilities through Claude API integration with comprehensive observability and domain-driven architecture.

## Goals

- Implement MCP Protocol 2024-11-05 compliant server
- Provide seamless Claude API integration with tool execution
- Enable comprehensive observability with OpenTelemetry
- Follow Domain-Driven Design and CQRS patterns
- Achieve ≥80% test coverage with robust error handling

## Architecture

### System Context

```mermaid
graph TB
    subgraph "External Systems"
        CLIENT[MCP Client]
        CLAUDE[Claude API]
        OTEL[OpenTelemetry Collector]
        DB[(Database)]
    end

    subgraph "TelemetryFlow GO MCP Server"
        TRANSPORT[Transport Layer]
        PROTOCOL[MCP Protocol Handler]
        APPLICATION[Application Layer]
        DOMAIN[Domain Layer]
        INFRA[Infrastructure Layer]
    end

    CLIENT -->|JSON-RPC 2.0| TRANSPORT
    TRANSPORT --> PROTOCOL
    PROTOCOL --> APPLICATION
    APPLICATION --> DOMAIN
    APPLICATION --> INFRA
    INFRA --> CLAUDE
    INFRA --> DB
    INFRA --> OTEL
```

### Domain Model

#### Core Aggregates

```go
// Session Aggregate - Root aggregate for MCP session lifecycle
type Session struct {
    id           vo.SessionID
    clientInfo   ClientInfo
    state        SessionState
    capabilities ServerCapabilities
    tools        map[string]*entities.Tool
    resources    map[string]*entities.Resource
    prompts      map[string]*entities.Prompt
    createdAt    time.Time
    lastActivity time.Time
}

// Conversation Aggregate - Manages Claude conversation state
type Conversation struct {
    id        vo.ConversationID
    sessionID vo.SessionID
    messages  []*entities.Message
    context   ConversationContext
    state     ConversationState
    createdAt time.Time
    updatedAt time.Time
}
```

#### Value Objects

```go
// SessionID - Unique session identifier
type SessionID struct {
    value string
}

func NewSessionID(value string) (SessionID, error) {
    if _, err := uuid.Parse(value); err != nil {
        return SessionID{}, ErrInvalidSessionID
    }
    return SessionID{value: value}, nil
}

// ConversationID - Unique conversation identifier
type ConversationID struct {
    value string
}

// ToolName - Tool identifier with validation
type ToolName struct {
    value string
}

func NewToolName(value string) (ToolName, error) {
    if !isValidToolName(value) {
        return ToolName{}, ErrInvalidToolName
    }
    return ToolName{value: value}, nil
}
```

#### Entities

```go
// Tool - MCP tool definition
type Tool struct {
    name        vo.ToolName
    description string
    inputSchema JSONSchema
    enabled     bool
    metadata    map[string]interface{}
}

// Resource - MCP resource definition
type Resource struct {
    uri         string
    name        string
    description string
    mimeType    string
    metadata    map[string]interface{}
}

// Message - Conversation message
type Message struct {
    id        vo.MessageID
    role      MessageRole
    content   []ContentBlock
    timestamp time.Time
    metadata  map[string]interface{}
}
```

## Implementation Requirements

### 1. MCP Protocol Compliance

#### Transport Layer

- **MUST** implement JSON-RPC 2.0 over stdio
- **MUST** support bidirectional communication
- **MUST** handle connection lifecycle properly
- **SHOULD** implement connection pooling for multiple clients

```go
type MCPTransport interface {
    Send(ctx context.Context, message *jsonrpc.Message) error
    Receive(ctx context.Context) (*jsonrpc.Message, error)
    Close() error
}

type StdioTransport struct {
    reader *bufio.Reader
    writer *bufio.Writer
    mutex  sync.RWMutex
}
```

#### Protocol Handler

- **MUST** implement all required MCP methods
- **MUST** validate protocol version compatibility
- **MUST** handle initialization sequence correctly
- **SHOULD** implement optional capabilities

```go
type MCPHandler interface {
    Initialize(ctx context.Context, req *InitializeRequest) (*InitializeResponse, error)
    ListTools(ctx context.Context, req *ListToolsRequest) (*ListToolsResponse, error)
    CallTool(ctx context.Context, req *CallToolRequest) (*CallToolResponse, error)
    ListResources(ctx context.Context, req *ListResourcesRequest) (*ListResourcesResponse, error)
    ReadResource(ctx context.Context, req *ReadResourceRequest) (*ReadResourceResponse, error)
    ListPrompts(ctx context.Context, req *ListPromptsRequest) (*ListPromptsResponse, error)
    GetPrompt(ctx context.Context, req *GetPromptRequest) (*GetPromptResponse, error)
}
```

### 2. Claude API Integration

#### Client Configuration

- **MUST** support multiple Claude models (Claude 4, Claude 3.5)
- **MUST** implement proper authentication with API keys
- **MUST** handle rate limiting and retries
- **SHOULD** support streaming responses

```go
type ClaudeClient struct {
    client      *anthropic.Client
    config      *ClaudeConfig
    rateLimiter *RateLimiter
    metrics     *ClaudeMetrics
}

type ClaudeConfig struct {
    APIKey      string        `yaml:"api_key" env:"ANTHROPIC_API_KEY"`
    BaseURL     string        `yaml:"base_url"`
    Timeout     time.Duration `yaml:"timeout"`
    MaxRetries  int           `yaml:"max_retries"`
    RateLimit   RateLimit     `yaml:"rate_limit"`
}
```

#### Tool Execution

- **MUST** convert MCP tools to Claude tool format
- **MUST** validate tool inputs against schemas
- **MUST** handle tool execution errors gracefully
- **SHOULD** support parallel tool execution

```go
func (h *ToolHandler) ExecuteTool(ctx context.Context, req *CallToolRequest) (*CallToolResponse, error) {
    // 1. Validate tool exists and is enabled
    tool, err := h.toolRepo.FindByName(ctx, req.Name)
    if err != nil {
        return nil, fmt.Errorf("tool not found: %w", err)
    }

    // 2. Validate input against schema
    if err := h.validator.ValidateInput(tool.InputSchema, req.Arguments); err != nil {
        return nil, fmt.Errorf("invalid input: %w", err)
    }

    // 3. Execute tool logic
    result, err := h.executeToolLogic(ctx, tool, req.Arguments)
    if err != nil {
        return nil, fmt.Errorf("tool execution failed: %w", err)
    }

    return &CallToolResponse{
        Content: result.Content,
        IsError: result.IsError,
    }, nil
}
```

### 3. Domain Layer Implementation

#### Session Management

- **MUST** implement session lifecycle (create, activate, close)
- **MUST** track session state and capabilities
- **MUST** validate client compatibility
- **SHOULD** implement session timeout handling

```go
func (s *Session) AddTool(tool *entities.Tool) error {
    if tool == nil {
        return ErrInvalidTool
    }

    if _, exists := s.tools[tool.Name().String()]; exists {
        return ErrToolAlreadyExists
    }

    s.tools[tool.Name().String()] = tool
    s.lastActivity = time.Now()

    // Publish domain event
    s.publishEvent(events.NewToolAddedEvent(s.id, tool.Name()))

    return nil
}

func (s *Session) IsActive() bool {
    return s.state == SessionStateActive &&
           time.Since(s.lastActivity) < SessionTimeout
}
```

#### Conversation Management

- **MUST** maintain conversation history
- **MUST** handle multi-turn conversations
- **MUST** implement conversation context management
- **SHOULD** support conversation branching

```go
func (c *Conversation) AddMessage(message *entities.Message) error {
    if message == nil {
        return ErrInvalidMessage
    }

    c.messages = append(c.messages, message)
    c.updatedAt = time.Now()

    // Publish domain event
    c.publishEvent(events.NewMessageAddedEvent(c.id, message.ID()))

    return nil
}

func (c *Conversation) GetRecentMessages(limit int) []*entities.Message {
    if len(c.messages) <= limit {
        return c.messages
    }
    return c.messages[len(c.messages)-limit:]
}
```

### 4. Application Layer (CQRS)

#### Commands

- **MUST** implement command handlers for write operations
- **MUST** validate commands before execution
- **MUST** publish domain events after successful execution
- **SHOULD** implement command validation pipeline

```go
type InitializeSessionCommand struct {
    ClientName      string
    ClientVersion   string
    ProtocolVersion string
    Capabilities    ClientCapabilities
}

func (h *SessionHandler) HandleInitializeSession(ctx context.Context, cmd *InitializeSessionCommand) (*aggregates.Session, error) {
    // 1. Validate protocol version
    if !h.isValidProtocolVersion(cmd.ProtocolVersion) {
        return nil, ErrInvalidProtocolVersion
    }

    // 2. Create session aggregate
    sessionID := vo.NewSessionID(uuid.New().String())
    session := aggregates.NewSession(sessionID, cmd.ClientName, cmd.ClientVersion)

    // 3. Configure capabilities
    if err := session.ConfigureCapabilities(cmd.Capabilities); err != nil {
        return nil, fmt.Errorf("failed to configure capabilities: %w", err)
    }

    // 4. Persist session
    if err := h.sessionRepo.Save(ctx, session); err != nil {
        return nil, fmt.Errorf("failed to save session: %w", err)
    }

    // 5. Publish event
    event := events.NewSessionCreatedEvent(sessionID, cmd.ClientName)
    if err := h.eventPublisher.Publish(ctx, event); err != nil {
        h.logger.Error().Err(err).Msg("Failed to publish session created event")
    }

    return session, nil
}
```

#### Queries

- **MUST** implement query handlers for read operations
- **MUST** optimize queries for performance
- **SHOULD** implement caching where appropriate
- **SHOULD** support pagination for large result sets

```go
type ListToolsQuery struct {
    SessionID   vo.SessionID
    EnabledOnly bool
    Limit       int
    Offset      int
}

func (h *ToolHandler) HandleListTools(ctx context.Context, query *ListToolsQuery) (*ToolListResult, error) {
    // 1. Validate session exists
    session, err := h.sessionRepo.FindByID(ctx, query.SessionID)
    if err != nil {
        return nil, fmt.Errorf("session not found: %w", err)
    }

    // 2. Get tools from session
    tools := session.GetTools(query.EnabledOnly)

    // 3. Apply pagination
    total := len(tools)
    start := query.Offset
    end := start + query.Limit

    if start > total {
        tools = []*entities.Tool{}
    } else if end > total {
        tools = tools[start:]
    } else {
        tools = tools[start:end]
    }

    return &ToolListResult{
        Tools:  tools,
        Total:  total,
        Limit:  query.Limit,
        Offset: query.Offset,
    }, nil
}
```

### 5. Infrastructure Layer

#### Repository Implementation

- **MUST** implement repository interfaces defined in domain
- **MUST** handle database transactions properly
- **MUST** implement proper error handling
- **SHOULD** support multiple storage backends

```go
type PostgresSessionRepository struct {
    db     *sql.DB
    logger *Logger
}

func (r *PostgresSessionRepository) Save(ctx context.Context, session *aggregates.Session) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()

    // Save session
    query := `
        INSERT INTO sessions (id, client_name, client_version, state, capabilities, created_at, last_activity)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (id) DO UPDATE SET
            state = EXCLUDED.state,
            capabilities = EXCLUDED.capabilities,
            last_activity = EXCLUDED.last_activity
    `

    _, err = tx.ExecContext(ctx, query,
        session.ID().String(),
        session.ClientName(),
        session.ClientVersion(),
        session.State(),
        session.Capabilities(),
        session.CreatedAt(),
        session.LastActivity(),
    )
    if err != nil {
        return fmt.Errorf("failed to save session: %w", err)
    }

    // Save tools
    if err := r.saveSessionTools(ctx, tx, session); err != nil {
        return fmt.Errorf("failed to save session tools: %w", err)
    }

    return tx.Commit()
}
```

#### Event Publishing

- **MUST** implement reliable event publishing
- **MUST** handle event publishing failures gracefully
- **SHOULD** support event ordering guarantees
- **SHOULD** implement event replay capabilities

```go
type EventPublisher struct {
    broker  MessageBroker
    logger  *Logger
    metrics *EventMetrics
}

func (p *EventPublisher) Publish(ctx context.Context, event DomainEvent) error {
    // 1. Serialize event
    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to serialize event: %w", err)
    }

    // 2. Create message
    message := &Message{
        ID:        uuid.New().String(),
        Type:      event.Type(),
        Data:      data,
        Timestamp: time.Now(),
        Metadata:  event.Metadata(),
    }

    // 3. Publish with retry
    return p.publishWithRetry(ctx, message, 3)
}
```

### 6. Observability Implementation

#### Distributed Tracing

- **MUST** implement OpenTelemetry tracing
- **MUST** propagate trace context across service boundaries
- **MUST** add relevant span attributes for MCP operations
- **SHOULD** implement intelligent sampling

```go
func (h *MCPHandler) HandleRequest(ctx context.Context, req *MCPRequest) (*MCPResponse, error) {
    ctx, span := StartMCPSpan(ctx, fmt.Sprintf("mcp.%s", req.Method),
        attribute.String("mcp.method", req.Method),
        attribute.String("mcp.request.id", req.ID),
        attribute.String("mcp.protocol.version", MCPProtocolVersion),
    )
    defer span.End()

    // Add request-specific attributes
    if sessionID := extractSessionID(req); sessionID != "" {
        span.SetAttributes(attribute.String("mcp.session.id", sessionID))
    }

    // Execute handler
    resp, err := h.executeHandler(ctx, req)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    span.SetStatus(codes.Ok, "Request completed successfully")
    return resp, nil
}
```

#### Metrics Collection

- **MUST** collect business and technical metrics
- **MUST** implement Prometheus-compatible metrics
- **SHOULD** implement custom dashboards
- **SHOULD** set up alerting rules

```go
type MCPMetrics struct {
    // Session metrics
    SessionsActive    prometheus.Gauge
    SessionsTotal     prometheus.Counter
    SessionDuration   prometheus.Histogram

    // Request metrics
    RequestsTotal     prometheus.Counter
    RequestDuration   prometheus.Histogram
    RequestErrors     prometheus.Counter

    // Tool metrics
    ToolCallsTotal    prometheus.Counter
    ToolCallDuration  prometheus.Histogram
    ToolCallErrors    prometheus.Counter

    // Claude metrics
    ClaudeRequestsTotal    prometheus.Counter
    ClaudeRequestDuration  prometheus.Histogram
    ClaudeTokensUsed       prometheus.Counter
}
```

#### Structured Logging

- **MUST** implement structured logging with Zerolog
- **MUST** include trace correlation IDs
- **MUST** log at appropriate levels
- **SHOULD** implement log aggregation

```go
func (l *Logger) LogMCPRequest(ctx context.Context, method string, params interface{}) {
    l.WithMCPContext(ctx).Info().
        Str("mcp_method", method).
        Interface("mcp_params", params).
        Msg("MCP request received")
}

func (l *Logger) LogMCPResponse(ctx context.Context, method string, duration time.Duration, err error) {
    event := l.WithMCPContext(ctx).Info()
    if err != nil {
        event = l.WithMCPContext(ctx).Error().Err(err)
    }

    event.
        Str("mcp_method", method).
        Dur("duration", duration).
        Msg("MCP request completed")
}
```

## Testing Requirements

### Unit Testing

- **MUST** achieve ≥85% code coverage for domain layer
- **MUST** implement table-driven tests
- **MUST** use mocks for external dependencies
- **SHOULD** test error conditions thoroughly

### Integration Testing

- **MUST** test MCP protocol compliance
- **MUST** test Claude API integration
- **MUST** test database operations
- **SHOULD** test performance under load

### End-to-End Testing

- **MUST** test complete user scenarios
- **MUST** test error recovery scenarios
- **SHOULD** test concurrent operations
- **SHOULD** implement performance benchmarks

## Configuration

### Environment Variables

```yaml
# Server Configuration
MCP_SERVER_HOST: "localhost"
MCP_SERVER_PORT: "8080"
MCP_PROTOCOL_VERSION: "2024-11-05"

# Claude API Configuration
ANTHROPIC_API_KEY: "sk-ant-..."
ANTHROPIC_BASE_URL: "https://api.anthropic.com"
ANTHROPIC_TIMEOUT: "30s"
ANTHROPIC_MAX_RETRIES: "3"

# Database Configuration
DATABASE_URL: "postgres://user:pass@localhost/tfo_mcp"
DATABASE_MAX_CONNECTIONS: "10"
DATABASE_TIMEOUT: "5s"

# Observability Configuration
TELEMETRYFLOW_MCP_SERVICE_NAME: "telemetryflow-mcp"
TELEMETRYFLOW_MCP_OTEL_SERVICE_VERSION: "1.0.0"
TELEMETRYFLOW_MCP_OTLP_ENDPOINT: "http://localhost:4317"
LOG_LEVEL: "info"
LOG_FORMAT: "json"
```

### Configuration File

```yaml
server:
  host: "localhost"
  port: 8080
  protocol_version: "2024-11-05"
  timeout: "30s"

claude:
  api_key: "${ANTHROPIC_API_KEY}"
  base_url: "https://api.anthropic.com"
  timeout: "30s"
  max_retries: 3
  rate_limit:
    requests_per_minute: 60
    tokens_per_minute: 100000

database:
  url: "${DATABASE_URL}"
  max_connections: 10
  timeout: "5s"
  migrations_path: "./migrations"

observability:
  tracing:
    enabled: true
    sample_rate: 0.1
  metrics:
    enabled: true
    interval: "10s"
  logging:
    level: "info"
    format: "json"
    structured: true
```

## Deployment

### Docker Configuration

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o tfo-mcp ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/tfo-mcp .
COPY --from=builder /app/configs ./configs

CMD ["./tfo-mcp", "server", "--config", "configs/production.yaml"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: telemetryflow-mcp
spec:
  replicas: 3
  selector:
    matchLabels:
      app: telemetryflow-mcp
  template:
    metadata:
      labels:
        app: telemetryflow-mcp
    spec:
      containers:
        - name: mcp-server
          image: telemetryflow/mcp-server:latest
          ports:
            - containerPort: 8080
          env:
            - name: ANTHROPIC_API_KEY
              valueFrom:
                secretKeyRef:
                  name: claude-api-secret
                  key: api-key
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: database-secret
                  key: url
          resources:
            requests:
              memory: "256Mi"
              cpu: "250m"
            limits:
              memory: "512Mi"
              cpu: "500m"
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 30
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /ready
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 5
```

## Success Criteria

### Functional Requirements

- [ ] MCP Protocol 2024-11-05 compliance verified
- [ ] Claude API integration working with all supported models
- [ ] All core MCP capabilities implemented (tools, resources, prompts)
- [ ] Session management working correctly
- [ ] Conversation management supporting multi-turn interactions

### Non-Functional Requirements

- [ ] ≥80% test coverage achieved
- [ ] Response time <100ms for P95 of requests
- [ ] Support for ≥100 concurrent sessions
- [ ] Zero data loss during normal operations
- [ ] Comprehensive observability implemented

### Quality Gates

- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] All E2E tests passing
- [ ] Security scan passing
- [ ] Performance benchmarks meeting requirements
- [ ] Documentation complete and up-to-date

## References

- [MCP Protocol Specification 2024-11-05](https://spec.modelcontextprotocol.io/specification/)
- [Anthropic Claude API Documentation](https://docs.anthropic.com/claude/reference/)
- [OpenTelemetry Go SDK Documentation](https://opentelemetry.io/docs/instrumentation/go/)
- [Domain-Driven Design Patterns](https://martinfowler.com/tags/domain%20driven%20design.html)
- [CQRS Pattern Documentation](https://martinfowler.com/bliki/CQRS.html)

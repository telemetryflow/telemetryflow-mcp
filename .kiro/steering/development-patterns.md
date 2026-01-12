# TFO-GO-MCP Development Patterns

## LEGO Builder Methodology

The TFO-GO-MCP follows the LEGO Builder methodology for modular, composable architecture.

### Module Structure

```
internal/
├── domain/           # Business logic (LEGO block core)
├── application/      # Use cases (LEGO block interface)
├── infrastructure/   # Technical implementation (LEGO block adapter)
└── presentation/     # API layer (LEGO block facade)
```

## Domain-Driven Design Patterns

### Aggregates

Aggregates are the consistency boundaries:

```go
// Session aggregate - manages MCP session lifecycle
type Session struct {
    id           vo.SessionID
    state        SessionState
    tools        map[string]*entities.Tool
    resources    map[string]*entities.Resource
    prompts      map[string]*entities.Prompt
    // ...
}

// Conversation aggregate - manages conversation lifecycle
type Conversation struct {
    id        vo.ConversationID
    sessionID vo.SessionID
    messages  []*entities.Message
    // ...
}
```

### Value Objects

Immutable, self-validating types:

```go
// SessionID with validation
type SessionID struct {
    value string
}

func NewSessionID(value string) (SessionID, error) {
    if _, err := uuid.Parse(value); err != nil {
        return SessionID{}, ErrInvalidSessionID
    }
    return SessionID{value: value}, nil
}
```

### Domain Events

Events for cross-aggregate communication:

```go
type SessionCreatedEvent struct {
    BaseEvent
}

func NewSessionCreatedEvent(sessionID vo.SessionID) *SessionCreatedEvent {
    return &SessionCreatedEvent{
        BaseEvent: newBaseEvent("session.created", sessionID.String(), "Session", ...),
    }
}
```

## CQRS Patterns

### Commands

Write operations with handlers:

```go
type InitializeSessionCommand struct {
    ClientName      string
    ClientVersion   string
    ProtocolVersion string
}

func (h *SessionHandler) HandleInitializeSession(ctx context.Context, cmd *InitializeSessionCommand) (*aggregates.Session, error) {
    // Create session, validate, persist, publish events
}
```

### Queries

Read operations:

```go
type ListToolsQuery struct {
    SessionID   vo.SessionID
    EnabledOnly bool
}

func (h *ToolHandler) HandleListTools(ctx context.Context, query *ListToolsQuery) (*ToolListResult, error) {
    // Query repository, return result
}
```

## Error Handling

Domain-specific errors:

```go
var (
    ErrSessionNotFound    = errors.New("session not found")
    ErrConversationClosed = errors.New("conversation is closed")
    ErrToolNotFound       = errors.New("tool not found")
)
```

MCP-specific errors:

```go
type MCPError struct {
    Code    vo.MCPErrorCode
    Message string
}
```

## Repository Pattern

Interface-based repositories:

```go
type ISessionRepository interface {
    Save(ctx context.Context, session *aggregates.Session) error
    FindByID(ctx context.Context, id vo.SessionID) (*aggregates.Session, error)
    FindActive(ctx context.Context) ([]*aggregates.Session, error)
    Delete(ctx context.Context, id vo.SessionID) error
}
```

## Dependency Injection

Constructor-based injection:

```go
func NewToolHandler(
    sessionRepo repositories.ISessionRepository,
    toolRepo repositories.IToolRepository,
    eventPublisher EventPublisher,
) *ToolHandler {
    return &ToolHandler{
        sessionRepo:    sessionRepo,
        toolRepo:       toolRepo,
        eventPublisher: eventPublisher,
    }
}
```

## Testing Patterns

### Table-driven tests

```go
func TestNewSessionID(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid uuid", "123e4567-e89b-12d3-a456-426614174000", false},
        {"invalid uuid", "invalid", true},
        {"empty", "", true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := NewSessionID(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewSessionID() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Mock implementations

```go
type MockSessionRepository struct {
    sessions map[string]*aggregates.Session
}

func (m *MockSessionRepository) Save(ctx context.Context, session *aggregates.Session) error {
    m.sessions[session.ID().String()] = session
    return nil
}
```

## Naming Conventions

| Type         | Pattern                   | Example                    |
| ------------ | ------------------------- | -------------------------- |
| Value Object | PascalCase                | `SessionID`, `ToolName`    |
| Aggregate    | PascalCase                | `Session`, `Conversation`  |
| Entity       | PascalCase                | `Message`, `Tool`          |
| Command      | `{Action}{Entity}Command` | `InitializeSessionCommand` |
| Query        | `{Action}{Entity}Query`   | `ListToolsQuery`           |
| Handler      | `{Entity}Handler`         | `SessionHandler`           |
| Repository   | `I{Entity}Repository`     | `ISessionRepository`       |
| Event        | `{Entity}{Action}Event`   | `SessionCreatedEvent`      |

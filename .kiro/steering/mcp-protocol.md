# TFO-GO-MCP Protocol Standards

## MCP Protocol Implementation

The TelemetryFlow GO MCP Server strictly follows the Model Context Protocol (MCP) specification version 2024-11-05.

### Protocol Fundamentals

```go
// MCP Protocol Version
const MCPProtocolVersion = "2024-11-05"

// Transport Layer - JSON-RPC 2.0 over stdio
type MCPTransport interface {
    Send(ctx context.Context, message *jsonrpc.Message) error
    Receive(ctx context.Context) (*jsonrpc.Message, error)
    Close() error
}
```

### Core MCP Capabilities

#### 1. Tools Capability

```go
type ToolsCapability struct {
    ListChanged bool `json:"listChanged,omitempty"`
}

// Tool definition following MCP spec
type Tool struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    InputSchema JSONSchema  `json:"inputSchema"`
}
```

#### 2. Resources Capability

```go
type ResourcesCapability struct {
    Subscribe   bool `json:"subscribe,omitempty"`
    ListChanged bool `json:"listChanged,omitempty"`
}

type Resource struct {
    URI         string `json:"uri"`
    Name        string `json:"name"`
    Description string `json:"description,omitempty"`
    MimeType    string `json:"mimeType,omitempty"`
}
```

#### 3. Prompts Capability

```go
type PromptsCapability struct {
    ListChanged bool `json:"listChanged,omitempty"`
}

type Prompt struct {
    Name        string          `json:"name"`
    Description string          `json:"description,omitempty"`
    Arguments   []PromptArgument `json:"arguments,omitempty"`
}
```

### Message Patterns

#### Request/Response Pattern

```go
// All MCP requests follow this pattern
type MCPRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   *MCPError   `json:"error,omitempty"`
}
```

#### Notification Pattern

```go
// Notifications have no ID and expect no response
type MCPNotification struct {
    JSONRPC string      `json:"jsonrpc"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}
```

### Error Handling

#### MCP Error Codes

```go
const (
    // Standard JSON-RPC errors
    ParseError     = -32700
    InvalidRequest = -32600
    MethodNotFound = -32601
    InvalidParams  = -32602
    InternalError  = -32603

    // MCP-specific errors
    InvalidTool     = -32000
    InvalidResource = -32001
    InvalidPrompt   = -32002
)

type MCPError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

### Session Management

#### Initialization Sequence

```go
// 1. Client sends initialize request
type InitializeRequest struct {
    ProtocolVersion string                 `json:"protocolVersion"`
    Capabilities    ClientCapabilities     `json:"capabilities"`
    ClientInfo      ClientInfo             `json:"clientInfo"`
}

// 2. Server responds with capabilities
type InitializeResponse struct {
    ProtocolVersion string                 `json:"protocolVersion"`
    Capabilities    ServerCapabilities     `json:"capabilities"`
    ServerInfo      ServerInfo             `json:"serverInfo"`
}

// 3. Client sends initialized notification
type InitializedNotification struct {
    // Empty params
}
```

### Content Types

#### Text Content

```go
type TextContent struct {
    Type string `json:"type"` // "text"
    Text string `json:"text"`
}
```

#### Image Content

```go
type ImageContent struct {
    Type     string `json:"type"` // "image"
    Data     string `json:"data"`     // base64 encoded
    MimeType string `json:"mimeType"` // image/png, image/jpeg, etc.
}
```

#### Resource Content

```go
type ResourceContent struct {
    Type     string `json:"type"` // "resource"
    Resource struct {
        URI      string `json:"uri"`
        Text     string `json:"text,omitempty"`
        Blob     string `json:"blob,omitempty"`
        MimeType string `json:"mimeType,omitempty"`
    } `json:"resource"`
}
```

### Validation Rules

#### Request Validation

- All requests MUST include `jsonrpc: "2.0"`
- Request ID MUST be string, number, or null
- Method names MUST follow MCP specification
- Params MUST match method schema

#### Response Validation

- Response ID MUST match request ID
- Either `result` or `error` MUST be present, not both
- Error codes MUST follow JSON-RPC and MCP specifications

#### Content Validation

- Text content MUST be valid UTF-8
- Image data MUST be valid base64
- Resource URIs MUST be valid according to RFC 3986

### Implementation Guidelines

#### Handler Registration

```go
type MCPHandler interface {
    Handle(ctx context.Context, req *MCPRequest) (*MCPResponse, error)
}

// Register handlers for each MCP method
func (s *MCPServer) RegisterHandler(method string, handler MCPHandler) {
    s.handlers[method] = handler
}
```

#### Middleware Pattern

```go
type MCPMiddleware func(MCPHandler) MCPHandler

// Common middleware: logging, validation, rate limiting
func LoggingMiddleware(next MCPHandler) MCPHandler {
    return MCPHandlerFunc(func(ctx context.Context, req *MCPRequest) (*MCPResponse, error) {
        // Log request
        resp, err := next.Handle(ctx, req)
        // Log response
        return resp, err
    })
}
```

#### Streaming Support

```go
// For tools that support streaming responses
type StreamingResponse struct {
    Type string      `json:"type"` // "progress" or "result"
    Data interface{} `json:"data"`
}
```

### Testing Standards

#### Protocol Compliance Tests

```go
func TestMCPProtocolCompliance(t *testing.T) {
    tests := []struct {
        name     string
        request  *MCPRequest
        wantCode int
    }{
        {
            name: "valid initialize request",
            request: &MCPRequest{
                JSONRPC: "2.0",
                ID:      "1",
                Method:  "initialize",
                Params:  validInitializeParams,
            },
            wantCode: 0, // success
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, err := server.Handle(ctx, tt.request)
            // Assert protocol compliance
        })
    }
}
```

### Security Considerations

#### Input Sanitization

- All string inputs MUST be sanitized
- File paths MUST be validated and sandboxed
- Resource URIs MUST be validated

#### Rate Limiting

- Implement per-client rate limiting
- Different limits for different operation types
- Graceful degradation under load

#### Authentication

- Support for API key authentication
- Optional client certificate validation
- Secure credential storage

# TFO-MCP Claude Integration Standards

## Claude API Integration

The TelemetryFlow GO MCP Server integrates with Anthropic's Claude API to provide AI-powered capabilities through the MCP protocol.

### Claude Client Configuration

```go
// Claude client setup following anthropic-sdk-go patterns
type ClaudeConfig struct {
    APIKey      string        `yaml:"api_key" env:"ANTHROPIC_API_KEY"`
    BaseURL     string        `yaml:"base_url" env:"ANTHROPIC_BASE_URL"`
    Timeout     time.Duration `yaml:"timeout" env:"ANTHROPIC_TIMEOUT"`
    MaxRetries  int           `yaml:"max_retries" env:"ANTHROPIC_MAX_RETRIES"`
    RateLimit   RateLimit     `yaml:"rate_limit"`
}

type RateLimit struct {
    RequestsPerMinute int `yaml:"requests_per_minute"`
    TokensPerMinute   int `yaml:"tokens_per_minute"`
}
```

### Supported Claude Models

```go
const (
    // Claude 4 Models (latest)
    Claude4Opus   = "claude-4-opus-20241022"
    Claude4Sonnet = "claude-4-sonnet-20241022"

    // Claude 3.5 Models
    Claude35Sonnet = "claude-3-5-sonnet-20241022"
    Claude35Haiku  = "claude-3-5-haiku-20241022"

    // Default model for MCP operations
    DefaultModel = Claude35Sonnet
)

type ModelCapabilities struct {
    MaxTokens       int  `json:"max_tokens"`
    SupportsVision  bool `json:"supports_vision"`
    SupportsTools   bool `json:"supports_tools"`
    SupportsStreaming bool `json:"supports_streaming"`
}
```

### Message Handling Patterns

#### Single Message Processing

```go
type ClaudeMessageRequest struct {
    Model       string          `json:"model"`
    Messages    []ClaudeMessage `json:"messages"`
    MaxTokens   int             `json:"max_tokens"`
    Temperature float64         `json:"temperature,omitempty"`
    Tools       []ClaudeTool    `json:"tools,omitempty"`
    ToolChoice  interface{}     `json:"tool_choice,omitempty"`
}

type ClaudeMessage struct {
    Role    string        `json:"role"` // "user", "assistant"
    Content []ContentBlock `json:"content"`
}

type ContentBlock struct {
    Type string `json:"type"` // "text", "image", "tool_use", "tool_result"
    // Type-specific fields populated based on Type
    Text     string                 `json:"text,omitempty"`
    Source   *ImageSource           `json:"source,omitempty"`
    ToolUseID string                `json:"tool_use_id,omitempty"`
    Name     string                 `json:"name,omitempty"`
    Input    map[string]interface{} `json:"input,omitempty"`
    Content  string                 `json:"content,omitempty"`
    IsError  bool                   `json:"is_error,omitempty"`
}
```

#### Streaming Response Handling

```go
type StreamingHandler struct {
    onStart    func()
    onContent  func(content string)
    onToolUse  func(toolUse ToolUse)
    onComplete func(response ClaudeResponse)
    onError    func(error)
}

func (c *ClaudeClient) StreamMessage(ctx context.Context, req *ClaudeMessageRequest, handler *StreamingHandler) error {
    // Implementation follows anthropic-sdk-go streaming patterns
    stream, err := c.client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
        Model:     anthropic.F(req.Model),
        Messages:  convertMessages(req.Messages),
        MaxTokens: anthropic.F(req.MaxTokens),
        Stream:    anthropic.F(true),
    })

    for stream.Next() {
        event := stream.Current()
        switch event.Type {
        case "content_block_delta":
            if handler.onContent != nil {
                handler.onContent(event.Delta.Text)
            }
        case "message_stop":
            if handler.onComplete != nil {
                handler.onComplete(convertResponse(event))
            }
        }
    }

    return stream.Err()
}
```

### Tool Integration

#### Claude Tool Definition

```go
type ClaudeTool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"input_schema"`
}

// Convert MCP Tool to Claude Tool format
func ConvertMCPToolToClaudeTool(mcpTool *Tool) *ClaudeTool {
    return &ClaudeTool{
        Name:        mcpTool.Name,
        Description: mcpTool.Description,
        InputSchema: mcpTool.InputSchema,
    }
}
```

#### Tool Execution Flow

```go
type ToolExecutionContext struct {
    SessionID     vo.SessionID
    ConversationID vo.ConversationID
    ToolName      string
    Input         map[string]interface{}
    RequestID     string
}

func (h *ClaudeHandler) ExecuteToolUse(ctx context.Context, toolUse ToolUse) (*ToolResult, error) {
    // 1. Validate tool exists and is enabled
    tool, err := h.toolRepo.FindByName(ctx, toolUse.Name)
    if err != nil {
        return nil, fmt.Errorf("tool not found: %w", err)
    }

    // 2. Validate input against schema
    if err := h.validator.ValidateToolInput(tool.InputSchema, toolUse.Input); err != nil {
        return nil, fmt.Errorf("invalid tool input: %w", err)
    }

    // 3. Execute tool through MCP protocol
    result, err := h.mcpClient.CallTool(ctx, &MCPToolCallRequest{
        Name:   toolUse.Name,
        Arguments: toolUse.Input,
    })

    // 4. Convert result to Claude format
    return &ToolResult{
        ToolUseID: toolUse.ID,
        Content:   result.Content,
        IsError:   err != nil,
    }, nil
}
```

### Conversation Management

#### Multi-turn Conversations

```go
type ConversationManager struct {
    repo        repositories.IConversationRepository
    claudeClient *ClaudeClient
    maxHistory   int
}

func (cm *ConversationManager) ContinueConversation(ctx context.Context, req *ContinueConversationRequest) (*ClaudeResponse, error) {
    // 1. Load conversation history
    conversation, err := cm.repo.FindByID(ctx, req.ConversationID)
    if err != nil {
        return nil, err
    }

    // 2. Build message history for Claude
    messages := cm.buildMessageHistory(conversation.Messages, req.NewMessage)

    // 3. Truncate if exceeding max history
    if len(messages) > cm.maxHistory {
        messages = messages[len(messages)-cm.maxHistory:]
    }

    // 4. Send to Claude
    response, err := cm.claudeClient.SendMessage(ctx, &ClaudeMessageRequest{
        Model:     req.Model,
        Messages:  messages,
        MaxTokens: req.MaxTokens,
        Tools:     req.AvailableTools,
    })

    // 5. Store response in conversation
    conversation.AddMessage(entities.NewMessage(
        vo.MessageID(uuid.New().String()),
        "assistant",
        response.Content,
        time.Now(),
    ))

    return response, cm.repo.Save(ctx, conversation)
}
```

### Error Handling

#### Claude API Errors

```go
type ClaudeError struct {
    Type    string `json:"type"`
    Message string `json:"message"`
    Code    int    `json:"code,omitempty"`
}

func (e *ClaudeError) Error() string {
    return fmt.Sprintf("Claude API error [%s]: %s", e.Type, e.Message)
}

// Convert Claude errors to MCP errors
func ConvertClaudeErrorToMCP(claudeErr *ClaudeError) *MCPError {
    switch claudeErr.Type {
    case "invalid_request_error":
        return &MCPError{
            Code:    InvalidParams,
            Message: claudeErr.Message,
        }
    case "rate_limit_error":
        return &MCPError{
            Code:    InternalError,
            Message: "Rate limit exceeded",
            Data:    map[string]interface{}{"retry_after": 60},
        }
    default:
        return &MCPError{
            Code:    InternalError,
            Message: claudeErr.Message,
        }
    }
}
```

### Rate Limiting and Retry Logic

#### Exponential Backoff

```go
type RetryConfig struct {
    MaxRetries      int           `yaml:"max_retries"`
    InitialDelay    time.Duration `yaml:"initial_delay"`
    MaxDelay        time.Duration `yaml:"max_delay"`
    BackoffFactor   float64       `yaml:"backoff_factor"`
    RetryableErrors []string      `yaml:"retryable_errors"`
}

func (c *ClaudeClient) executeWithRetry(ctx context.Context, operation func() error) error {
    var lastErr error

    for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
        if attempt > 0 {
            delay := time.Duration(float64(c.retryConfig.InitialDelay) *
                math.Pow(c.retryConfig.BackoffFactor, float64(attempt-1)))
            if delay > c.retryConfig.MaxDelay {
                delay = c.retryConfig.MaxDelay
            }

            select {
            case <-ctx.Done():
                return ctx.Err()
            case <-time.After(delay):
            }
        }

        if err := operation(); err != nil {
            lastErr = err
            if !c.isRetryableError(err) {
                return err
            }
            continue
        }

        return nil
    }

    return fmt.Errorf("operation failed after %d attempts: %w",
        c.retryConfig.MaxRetries, lastErr)
}
```

### Security and Authentication

#### API Key Management

```go
type APIKeyManager struct {
    keyRotation time.Duration
    validator   APIKeyValidator
}

func (akm *APIKeyManager) ValidateAPIKey(ctx context.Context, key string) error {
    // 1. Check key format
    if !akm.validator.IsValidFormat(key) {
        return errors.New("invalid API key format")
    }

    // 2. Test key with Claude API
    client := anthropic.NewClient(
        option.WithAPIKey(key),
    )

    // Make a minimal test request
    _, err := client.Messages.New(ctx, anthropic.MessageNewParams{
        Model:     anthropic.F(Claude35Haiku),
        Messages:  anthropic.F([]anthropic.MessageParam{{
            Role:    anthropic.F(anthropic.MessageParamRoleUser),
            Content: anthropic.F([]anthropic.ContentBlockParamUnion{
                anthropic.TextBlockParam{
                    Type: anthropic.F(anthropic.TextBlockParamTypeText),
                    Text: anthropic.F("test"),
                },
            }),
        }}),
        MaxTokens: anthropic.F(1),
    })

    return err
}
```

### Monitoring and Observability

#### Claude API Metrics

```go
type ClaudeMetrics struct {
    RequestsTotal     prometheus.Counter
    RequestDuration   prometheus.Histogram
    TokensUsed        prometheus.Counter
    ErrorsTotal       prometheus.Counter
    RateLimitHits     prometheus.Counter
}

func (cm *ClaudeMetrics) RecordRequest(duration time.Duration, tokens int, err error) {
    cm.RequestsTotal.Inc()
    cm.RequestDuration.Observe(duration.Seconds())
    cm.TokensUsed.Add(float64(tokens))

    if err != nil {
        cm.ErrorsTotal.Inc()
        if isRateLimitError(err) {
            cm.RateLimitHits.Inc()
        }
    }
}
```

### Testing Patterns

#### Claude Integration Tests

```go
func TestClaudeIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Claude integration test in short mode")
    }

    client := NewClaudeClient(&ClaudeConfig{
        APIKey: os.Getenv("ANTHROPIC_API_KEY"),
        Model:  Claude35Sonnet,
    })

    tests := []struct {
        name     string
        messages []ClaudeMessage
        wantErr  bool
    }{
        {
            name: "simple text message",
            messages: []ClaudeMessage{{
                Role: "user",
                Content: []ContentBlock{{
                    Type: "text",
                    Text: "Hello, Claude!",
                }},
            }},
            wantErr: false,
        },
        // ... more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, err := client.SendMessage(ctx, &ClaudeMessageRequest{
                Model:     Claude35Sonnet,
                Messages:  tt.messages,
                MaxTokens: 100,
            })

            if (err != nil) != tt.wantErr {
                t.Errorf("SendMessage() error = %v, wantErr %v", err, tt.wantErr)
            }

            if !tt.wantErr && resp == nil {
                t.Error("Expected response but got nil")
            }
        })
    }
}
```

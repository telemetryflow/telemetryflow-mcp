# TFO-MCP Testing Standards

## Testing Philosophy

The TelemetryFlow GO MCP Server follows a comprehensive testing strategy with ≥80% coverage target, emphasizing table-driven tests, mock implementations, and integration testing.

### Testing Pyramid

```go
// Testing levels in order of execution speed and isolation
const (
    UnitTestLevel        = "unit"        // Fast, isolated, mocked dependencies
    IntegrationTestLevel = "integration" // Medium, real dependencies, controlled environment
    E2ETestLevel        = "e2e"         // Slow, full system, real environment
)
```

### Test Organization

```
tests/
├── unit/           # Unit tests (fast, isolated)
│   ├── domain/     # Domain logic tests
│   ├── application/ # Use case tests
│   └── infrastructure/ # Infrastructure tests with mocks
├── integration/    # Integration tests (medium speed)
│   ├── mcp/        # MCP protocol integration
│   ├── claude/     # Claude API integration
│   └── persistence/ # Database integration
└── e2e/           # End-to-end tests (slow)
    ├── scenarios/  # Complete user scenarios
    └── performance/ # Performance tests
```

## Unit Testing Standards

### Table-Driven Test Pattern

```go
func TestSessionID_Validation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        errType error
    }{
        {
            name:    "valid UUID",
            input:   "123e4567-e89b-12d3-a456-426614174000",
            wantErr: false,
        },
        {
            name:    "invalid UUID format",
            input:   "invalid-uuid",
            wantErr: true,
            errType: vo.ErrInvalidSessionID,
        },
        {
            name:    "empty string",
            input:   "",
            wantErr: true,
            errType: vo.ErrInvalidSessionID,
        },
        {
            name:    "nil UUID",
            input:   "00000000-0000-0000-0000-000000000000",
            wantErr: true,
            errType: vo.ErrInvalidSessionID,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            sessionID, err := vo.NewSessionID(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errType != nil {
                    assert.ErrorIs(t, err, tt.errType)
                }
                assert.Empty(t, sessionID.String())
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.input, sessionID.String())
            }
        })
    }
}
```

### Domain Logic Testing

```go
func TestSession_AddTool(t *testing.T) {
    // Arrange
    sessionID, _ := vo.NewSessionID("123e4567-e89b-12d3-a456-426614174000")
    session := aggregates.NewSession(sessionID, "test-client", "1.0.0")

    tool := entities.NewTool(
        vo.ToolName("test_tool"),
        "Test tool description",
        map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "input": map[string]interface{}{
                    "type": "string",
                },
            },
        },
    )

    tests := []struct {
        name    string
        tool    *entities.Tool
        wantErr bool
        errType error
    }{
        {
            name:    "add valid tool",
            tool:    tool,
            wantErr: false,
        },
        {
            name:    "add duplicate tool",
            tool:    tool, // Same tool again
            wantErr: true,
            errType: aggregates.ErrToolAlreadyExists,
        },
        {
            name:    "add nil tool",
            tool:    nil,
            wantErr: true,
            errType: aggregates.ErrInvalidTool,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := session.AddTool(tt.tool)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errType != nil {
                    assert.ErrorIs(t, err, tt.errType)
                }
            } else {
                assert.NoError(t, err)
                assert.True(t, session.HasTool(tt.tool.Name()))
            }
        })
    }
}
```

### CQRS Handler Testing

```go
func TestSessionHandler_HandleInitializeSession(t *testing.T) {
    // Test setup
    mockRepo := &mocks.MockSessionRepository{}
    mockEventPublisher := &mocks.MockEventPublisher{}
    handler := application.NewSessionHandler(mockRepo, mockEventPublisher)

    tests := []struct {
        name    string
        command *application.InitializeSessionCommand
        setup   func(*mocks.MockSessionRepository, *mocks.MockEventPublisher)
        wantErr bool
        errType error
    }{
        {
            name: "successful initialization",
            command: &application.InitializeSessionCommand{
                ClientName:      "test-client",
                ClientVersion:   "1.0.0",
                ProtocolVersion: "2024-11-05",
            },
            setup: func(repo *mocks.MockSessionRepository, pub *mocks.MockEventPublisher) {
                repo.On("Save", mock.Anything, mock.AnythingOfType("*aggregates.Session")).
                    Return(nil)
                pub.On("Publish", mock.Anything, mock.AnythingOfType("*events.SessionCreatedEvent")).
                    Return(nil)
            },
            wantErr: false,
        },
        {
            name: "invalid protocol version",
            command: &application.InitializeSessionCommand{
                ClientName:      "test-client",
                ClientVersion:   "1.0.0",
                ProtocolVersion: "invalid-version",
            },
            setup:   func(repo *mocks.MockSessionRepository, pub *mocks.MockEventPublisher) {},
            wantErr: true,
            errType: application.ErrInvalidProtocolVersion,
        },
        {
            name: "repository save failure",
            command: &application.InitializeSessionCommand{
                ClientName:      "test-client",
                ClientVersion:   "1.0.0",
                ProtocolVersion: "2024-11-05",
            },
            setup: func(repo *mocks.MockSessionRepository, pub *mocks.MockEventPublisher) {
                repo.On("Save", mock.Anything, mock.AnythingOfType("*aggregates.Session")).
                    Return(errors.New("database error"))
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            tt.setup(mockRepo, mockEventPublisher)

            // Execute
            ctx := context.Background()
            session, err := handler.HandleInitializeSession(ctx, tt.command)

            // Assert
            if tt.wantErr {
                assert.Error(t, err)
                assert.Nil(t, session)
                if tt.errType != nil {
                    assert.ErrorIs(t, err, tt.errType)
                }
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, session)
                assert.Equal(t, tt.command.ClientName, session.ClientName())
            }

            // Verify mock expectations
            mockRepo.AssertExpectations(t)
            mockEventPublisher.AssertExpectations(t)
        })
    }
}
```

## Mock Implementations

### Repository Mocks

```go
type MockSessionRepository struct {
    mock.Mock
    sessions map[string]*aggregates.Session
}

func NewMockSessionRepository() *MockSessionRepository {
    return &MockSessionRepository{
        sessions: make(map[string]*aggregates.Session),
    }
}

func (m *MockSessionRepository) Save(ctx context.Context, session *aggregates.Session) error {
    args := m.Called(ctx, session)
    if args.Error(0) == nil {
        m.sessions[session.ID().String()] = session
    }
    return args.Error(0)
}

func (m *MockSessionRepository) FindByID(ctx context.Context, id vo.SessionID) (*aggregates.Session, error) {
    args := m.Called(ctx, id)
    if session, exists := m.sessions[id.String()]; exists {
        return session, args.Error(1)
    }
    return nil, repositories.ErrSessionNotFound
}

func (m *MockSessionRepository) FindActive(ctx context.Context) ([]*aggregates.Session, error) {
    args := m.Called(ctx)
    var active []*aggregates.Session
    for _, session := range m.sessions {
        if session.IsActive() {
            active = append(active, session)
        }
    }
    return active, args.Error(1)
}

func (m *MockSessionRepository) Delete(ctx context.Context, id vo.SessionID) error {
    args := m.Called(ctx, id)
    if args.Error(0) == nil {
        delete(m.sessions, id.String())
    }
    return args.Error(0)
}
```

### Claude Client Mock

```go
type MockClaudeClient struct {
    mock.Mock
}

func (m *MockClaudeClient) SendMessage(ctx context.Context, req *claude.MessageRequest) (*claude.Response, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*claude.Response), args.Error(1)
}

func (m *MockClaudeClient) StreamMessage(ctx context.Context, req *claude.MessageRequest, handler *claude.StreamingHandler) error {
    args := m.Called(ctx, req, handler)

    // Simulate streaming response
    if args.Error(0) == nil && handler.OnContent != nil {
        handler.OnContent("Mocked streaming response")
    }

    return args.Error(0)
}

// Helper method to setup common mock responses
func (m *MockClaudeClient) SetupSuccessResponse(content string, tokens int) {
    response := &claude.Response{
        Content: []claude.ContentBlock{{
            Type: "text",
            Text: content,
        }},
        Usage: &claude.Usage{
            InputTokens:  10,
            OutputTokens: tokens,
        },
    }

    m.On("SendMessage", mock.Anything, mock.Anything).Return(response, nil)
}
```

## Integration Testing

### MCP Protocol Integration Tests

```go
func TestMCPProtocolIntegration(t *testing.T) {
    // Setup test server
    server := setupTestMCPServer(t)
    defer server.Close()

    client := setupTestMCPClient(t, server.Address())
    defer client.Close()

    tests := []struct {
        name     string
        request  *mcp.Request
        wantCode int
        validate func(*testing.T, *mcp.Response)
    }{
        {
            name: "initialize session",
            request: &mcp.Request{
                JSONRPC: "2.0",
                ID:      "1",
                Method:  "initialize",
                Params: map[string]interface{}{
                    "protocolVersion": "2024-11-05",
                    "clientInfo": map[string]interface{}{
                        "name":    "test-client",
                        "version": "1.0.0",
                    },
                    "capabilities": map[string]interface{}{
                        "tools": map[string]interface{}{},
                    },
                },
            },
            wantCode: 0,
            validate: func(t *testing.T, resp *mcp.Response) {
                assert.NotNil(t, resp.Result)
                result := resp.Result.(map[string]interface{})
                assert.Equal(t, "2024-11-05", result["protocolVersion"])
                assert.Contains(t, result, "capabilities")
                assert.Contains(t, result, "serverInfo")
            },
        },
        {
            name: "list tools",
            request: &mcp.Request{
                JSONRPC: "2.0",
                ID:      "2",
                Method:  "tools/list",
                Params:  map[string]interface{}{},
            },
            wantCode: 0,
            validate: func(t *testing.T, resp *mcp.Response) {
                assert.NotNil(t, resp.Result)
                result := resp.Result.(map[string]interface{})
                assert.Contains(t, result, "tools")
                tools := result["tools"].([]interface{})
                assert.GreaterOrEqual(t, len(tools), 0)
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            resp, err := client.SendRequest(context.Background(), tt.request)
            assert.NoError(t, err)
            assert.NotNil(t, resp)

            if tt.wantCode == 0 {
                assert.Nil(t, resp.Error)
                assert.NotNil(t, resp.Result)
            } else {
                assert.NotNil(t, resp.Error)
                assert.Equal(t, tt.wantCode, resp.Error.Code)
            }

            if tt.validate != nil {
                tt.validate(t, resp)
            }
        })
    }
}
```

### Claude API Integration Tests

```go
func TestClaudeAPIIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Claude API integration test in short mode")
    }

    apiKey := os.Getenv("ANTHROPIC_API_KEY")
    if apiKey == "" {
        t.Skip("ANTHROPIC_API_KEY not set, skipping Claude integration tests")
    }

    client := claude.NewClient(&claude.Config{
        APIKey:  apiKey,
        Timeout: 30 * time.Second,
    })

    tests := []struct {
        name     string
        request  *claude.MessageRequest
        wantErr  bool
        validate func(*testing.T, *claude.Response)
    }{
        {
            name: "simple text message",
            request: &claude.MessageRequest{
                Model: claude.Claude35Sonnet,
                Messages: []claude.Message{{
                    Role: "user",
                    Content: []claude.ContentBlock{{
                        Type: "text",
                        Text: "Hello, Claude! Please respond with exactly 'Hello, World!'",
                    }},
                }},
                MaxTokens: 50,
            },
            wantErr: false,
            validate: func(t *testing.T, resp *claude.Response) {
                assert.NotEmpty(t, resp.Content)
                assert.Contains(t, resp.Content[0].Text, "Hello")
                assert.NotNil(t, resp.Usage)
                assert.Greater(t, resp.Usage.OutputTokens, 0)
            },
        },
        {
            name: "tool use request",
            request: &claude.MessageRequest{
                Model: claude.Claude35Sonnet,
                Messages: []claude.Message{{
                    Role: "user",
                    Content: []claude.ContentBlock{{
                        Type: "text",
                        Text: "What's the current time?",
                    }},
                }},
                MaxTokens: 100,
                Tools: []claude.Tool{{
                    Name:        "get_current_time",
                    Description: "Get the current time",
                    InputSchema: map[string]interface{}{
                        "type":       "object",
                        "properties": map[string]interface{}{},
                    },
                }},
            },
            wantErr: false,
            validate: func(t *testing.T, resp *claude.Response) {
                assert.NotEmpty(t, resp.Content)
                // Should contain tool use
                hasToolUse := false
                for _, content := range resp.Content {
                    if content.Type == "tool_use" {
                        hasToolUse = true
                        assert.Equal(t, "get_current_time", content.Name)
                        break
                    }
                }
                assert.True(t, hasToolUse, "Expected tool use in response")
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
            defer cancel()

            resp, err := client.SendMessage(ctx, tt.request)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, resp)

                if tt.validate != nil {
                    tt.validate(t, resp)
                }
            }
        })
    }
}
```

## End-to-End Testing

### Complete Scenario Tests

```go
func TestE2E_CompleteConversationFlow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping E2E test in short mode")
    }

    // Setup complete test environment
    testEnv := setupE2EEnvironment(t)
    defer testEnv.Cleanup()

    t.Run("complete conversation with tool use", func(t *testing.T) {
        // 1. Initialize MCP session
        initResp, err := testEnv.MCPClient.Initialize(context.Background(), &mcp.InitializeRequest{
            ProtocolVersion: "2024-11-05",
            ClientInfo: mcp.ClientInfo{
                Name:    "e2e-test-client",
                Version: "1.0.0",
            },
            Capabilities: mcp.ClientCapabilities{
                Tools: &mcp.ToolsCapability{},
            },
        })
        assert.NoError(t, err)
        assert.NotNil(t, initResp)

        // 2. List available tools
        toolsResp, err := testEnv.MCPClient.ListTools(context.Background())
        assert.NoError(t, err)
        assert.NotEmpty(t, toolsResp.Tools)

        // 3. Start conversation with Claude
        conversationResp, err := testEnv.MCPClient.StartConversation(context.Background(), &mcp.StartConversationRequest{
            Message: "Please use the search_files tool to find all Go files in the current directory.",
        })
        assert.NoError(t, err)
        assert.NotNil(t, conversationResp)

        // 4. Verify tool was called
        assert.True(t, conversationResp.ToolsUsed)
        assert.Contains(t, conversationResp.ToolCalls, "search_files")

        // 5. Verify response contains file information
        assert.Contains(t, conversationResp.Response, ".go")
        assert.NotEmpty(t, conversationResp.ConversationID)

        // 6. Continue conversation
        continueResp, err := testEnv.MCPClient.ContinueConversation(context.Background(), &mcp.ContinueConversationRequest{
            ConversationID: conversationResp.ConversationID,
            Message:        "How many Go files did you find?",
        })
        assert.NoError(t, err)
        assert.NotNil(t, continueResp)
        assert.Contains(t, continueResp.Response, "found")
    })
}
```

### Performance Tests

```go
func TestPerformance_ConcurrentSessions(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping performance test in short mode")
    }

    server := setupTestMCPServer(t)
    defer server.Close()

    const numConcurrentSessions = 10
    const requestsPerSession = 5

    var wg sync.WaitGroup
    results := make(chan time.Duration, numConcurrentSessions*requestsPerSession)
    errors := make(chan error, numConcurrentSessions*requestsPerSession)

    for i := 0; i < numConcurrentSessions; i++ {
        wg.Add(1)
        go func(sessionID int) {
            defer wg.Done()

            client := setupTestMCPClient(t, server.Address())
            defer client.Close()

            // Initialize session
            _, err := client.Initialize(context.Background(), &mcp.InitializeRequest{
                ProtocolVersion: "2024-11-05",
                ClientInfo: mcp.ClientInfo{
                    Name:    fmt.Sprintf("perf-test-client-%d", sessionID),
                    Version: "1.0.0",
                },
            })
            if err != nil {
                errors <- err
                return
            }

            // Make multiple requests
            for j := 0; j < requestsPerSession; j++ {
                start := time.Now()

                _, err := client.ListTools(context.Background())
                duration := time.Since(start)

                if err != nil {
                    errors <- err
                } else {
                    results <- duration
                }
            }
        }(i)
    }

    wg.Wait()
    close(results)
    close(errors)

    // Collect results
    var durations []time.Duration
    for duration := range results {
        durations = append(durations, duration)
    }

    // Check for errors
    var testErrors []error
    for err := range errors {
        testErrors = append(testErrors, err)
    }

    // Assert performance requirements
    assert.Empty(t, testErrors, "No errors should occur during concurrent testing")
    assert.Len(t, durations, numConcurrentSessions*requestsPerSession, "All requests should complete")

    // Calculate performance metrics
    if len(durations) > 0 {
        sort.Slice(durations, func(i, j int) bool {
            return durations[i] < durations[j]
        })

        p50 := durations[len(durations)/2]
        p95 := durations[int(float64(len(durations))*0.95)]
        p99 := durations[int(float64(len(durations))*0.99)]

        t.Logf("Performance metrics:")
        t.Logf("  P50: %v", p50)
        t.Logf("  P95: %v", p95)
        t.Logf("  P99: %v", p99)

        // Assert performance requirements
        assert.Less(t, p50, 100*time.Millisecond, "P50 should be under 100ms")
        assert.Less(t, p95, 500*time.Millisecond, "P95 should be under 500ms")
        assert.Less(t, p99, 1*time.Second, "P99 should be under 1s")
    }
}
```

## Test Utilities and Helpers

### Test Environment Setup

```go
type TestEnvironment struct {
    MCPServer   *mcp.Server
    MCPClient   *mcp.Client
    ClaudeClient *claude.Client
    Database    *sql.DB
    Redis       *redis.Client
    TempDir     string
}

func setupE2EEnvironment(t *testing.T) *TestEnvironment {
    // Create temporary directory
    tempDir, err := os.MkdirTemp("", "tfo-mcp-e2e-*")
    require.NoError(t, err)

    // Setup test database
    db := setupTestDatabase(t)

    // Setup test Redis
    redisClient := setupTestRedis(t)

    // Setup mock Claude client for E2E tests
    claudeClient := setupMockClaudeClient(t)

    // Setup MCP server
    server := setupTestMCPServer(t, &TestServerConfig{
        Database:     db,
        Redis:        redisClient,
        ClaudeClient: claudeClient,
        TempDir:      tempDir,
    })

    // Setup MCP client
    client := setupTestMCPClient(t, server.Address())

    return &TestEnvironment{
        MCPServer:    server,
        MCPClient:    client,
        ClaudeClient: claudeClient,
        Database:     db,
        Redis:        redisClient,
        TempDir:      tempDir,
    }
}

func (te *TestEnvironment) Cleanup() {
    if te.MCPClient != nil {
        te.MCPClient.Close()
    }
    if te.MCPServer != nil {
        te.MCPServer.Close()
    }
    if te.Database != nil {
        te.Database.Close()
    }
    if te.Redis != nil {
        te.Redis.Close()
    }
    if te.TempDir != "" {
        os.RemoveAll(te.TempDir)
    }
}
```

### Test Data Builders

```go
type SessionBuilder struct {
    id           vo.SessionID
    clientName   string
    clientVersion string
    tools        []*entities.Tool
    resources    []*entities.Resource
}

func NewSessionBuilder() *SessionBuilder {
    id, _ := vo.NewSessionID(uuid.New().String())
    return &SessionBuilder{
        id:            id,
        clientName:    "test-client",
        clientVersion: "1.0.0",
        tools:         []*entities.Tool{},
        resources:     []*entities.Resource{},
    }
}

func (sb *SessionBuilder) WithID(id string) *SessionBuilder {
    sessionID, _ := vo.NewSessionID(id)
    sb.id = sessionID
    return sb
}

func (sb *SessionBuilder) WithClient(name, version string) *SessionBuilder {
    sb.clientName = name
    sb.clientVersion = version
    return sb
}

func (sb *SessionBuilder) WithTool(name, description string, schema map[string]interface{}) *SessionBuilder {
    tool := entities.NewTool(vo.ToolName(name), description, schema)
    sb.tools = append(sb.tools, tool)
    return sb
}

func (sb *SessionBuilder) Build() *aggregates.Session {
    session := aggregates.NewSession(sb.id, sb.clientName, sb.clientVersion)

    for _, tool := range sb.tools {
        session.AddTool(tool)
    }

    for _, resource := range sb.resources {
        session.AddResource(resource)
    }

    return session
}
```

## Coverage and Quality Gates

### Coverage Requirements

```go
// Coverage thresholds for different test types
const (
    UnitTestCoverageThreshold        = 85.0 // 85% minimum for unit tests
    IntegrationTestCoverageThreshold = 70.0 // 70% minimum for integration tests
    OverallCoverageThreshold         = 80.0 // 80% minimum overall
)

// Critical paths that must have 100% coverage
var CriticalPaths = []string{
    "internal/domain/aggregates",
    "internal/domain/vo",
    "internal/application/commands",
    "internal/application/queries",
}
```

### Test Quality Metrics

```go
type TestQualityMetrics struct {
    TotalTests        int
    PassingTests      int
    FailingTests      int
    SkippedTests      int
    CoveragePercent   float64
    TestDuration      time.Duration
    SlowTests         []SlowTest
}

type SlowTest struct {
    Name     string
    Duration time.Duration
    Threshold time.Duration
}
```

# TFO-MCP Observability Standards

## OpenTelemetry Integration

The TelemetryFlow GO MCP Server implements comprehensive observability using OpenTelemetry SDK v1.39.0 with structured logging via Zerolog v1.33.0.

### Telemetry Configuration

```go
type TelemetryConfig struct {
    ServiceName    string            `yaml:"service_name" env:"TELEMETRYFLOW_MCP_SERVICE_NAME"`
    ServiceVersion string            `yaml:"service_version" env:"TELEMETRYFLOW_MCP_OTEL_SERVICE_VERSION"`
    Environment    string            `yaml:"environment" env:"OTEL_ENVIRONMENT"`
    Tracing        TracingConfig     `yaml:"tracing"`
    Metrics        MetricsConfig     `yaml:"metrics"`
    Logging        LoggingConfig     `yaml:"logging"`
    Exporters      ExportersConfig   `yaml:"exporters"`
}

type TracingConfig struct {
    Enabled     bool    `yaml:"enabled" env:"OTEL_TRACES_ENABLED"`
    SampleRate  float64 `yaml:"sample_rate" env:"OTEL_TRACES_SAMPLE_RATE"`
    MaxSpans    int     `yaml:"max_spans" env:"OTEL_TRACES_MAX_SPANS"`
}

type MetricsConfig struct {
    Enabled  bool          `yaml:"enabled" env:"OTEL_METRICS_ENABLED"`
    Interval time.Duration `yaml:"interval" env:"OTEL_METRICS_INTERVAL"`
}

type LoggingConfig struct {
    Level      string `yaml:"level" env:"LOG_LEVEL"`
    Format     string `yaml:"format" env:"LOG_FORMAT"` // "json" or "text"
    Output     string `yaml:"output" env:"LOG_OUTPUT"` // "stdout", "stderr", or file path
    Structured bool   `yaml:"structured" env:"LOG_STRUCTURED"`
}
```

### Distributed Tracing

#### Trace Context Propagation

```go
// Trace context keys for MCP operations
const (
    TraceKeySessionID      = "mcp.session.id"
    TraceKeyConversationID = "mcp.conversation.id"
    TraceKeyToolName       = "mcp.tool.name"
    TraceKeyResourceURI    = "mcp.resource.uri"
    TraceKeyPromptName     = "mcp.prompt.name"
    TraceKeyClaudeModel    = "claude.model"
    TraceKeyRequestID      = "mcp.request.id"
)

// Span creation patterns
func StartMCPSpan(ctx context.Context, operationName string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
    tracer := otel.Tracer("telemetryflow-mcp")

    // Add standard MCP attributes
    standardAttrs := []attribute.KeyValue{
        attribute.String("mcp.protocol.version", MCPProtocolVersion),
        attribute.String("service.name", "telemetryflow-mcp"),
    }

    allAttrs := append(standardAttrs, attrs...)
    return tracer.Start(ctx, operationName, trace.WithAttributes(allAttrs...))
}
```

#### MCP Operation Tracing

```go
// Session lifecycle tracing
func (h *SessionHandler) HandleInitializeSession(ctx context.Context, cmd *InitializeSessionCommand) (*aggregates.Session, error) {
    ctx, span := StartMCPSpan(ctx, "mcp.session.initialize",
        attribute.String("mcp.client.name", cmd.ClientName),
        attribute.String("mcp.client.version", cmd.ClientVersion),
        attribute.String("mcp.protocol.version", cmd.ProtocolVersion),
    )
    defer span.End()

    // Implementation with span events
    span.AddEvent("session.validation.start")

    session, err := h.createSession(ctx, cmd)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    span.AddEvent("session.created", trace.WithAttributes(
        attribute.String(TraceKeySessionID, session.ID().String()),
    ))

    span.SetStatus(codes.Ok, "Session initialized successfully")
    return session, nil
}

// Tool execution tracing
func (h *ToolHandler) HandleCallTool(ctx context.Context, cmd *CallToolCommand) (*ToolResult, error) {
    ctx, span := StartMCPSpan(ctx, "mcp.tool.call",
        attribute.String(TraceKeyToolName, cmd.Name),
        attribute.String(TraceKeySessionID, cmd.SessionID.String()),
    )
    defer span.End()

    // Add tool-specific attributes
    span.SetAttributes(
        attribute.Int("mcp.tool.input.size", len(cmd.Arguments)),
        attribute.String("mcp.tool.execution.mode", "sync"),
    )

    result, err := h.executeTool(ctx, cmd)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    span.SetAttributes(
        attribute.Int("mcp.tool.result.size", len(result.Content)),
        attribute.Bool("mcp.tool.result.success", !result.IsError),
    )

    return result, nil
}
```

#### Claude API Tracing

```go
func (c *ClaudeClient) SendMessage(ctx context.Context, req *ClaudeMessageRequest) (*ClaudeResponse, error) {
    ctx, span := StartMCPSpan(ctx, "claude.message.send",
        attribute.String(TraceKeyClaudeModel, req.Model),
        attribute.Int("claude.message.count", len(req.Messages)),
        attribute.Int("claude.max_tokens", req.MaxTokens),
    )
    defer span.End()

    // Record request details
    span.AddEvent("claude.request.start", trace.WithAttributes(
        attribute.Float64("claude.temperature", req.Temperature),
        attribute.Int("claude.tools.count", len(req.Tools)),
    ))

    start := time.Now()
    response, err := c.client.Messages.New(ctx, convertToAnthropicParams(req))
    duration := time.Since(start)

    // Record response metrics
    span.SetAttributes(
        attribute.Int64("claude.request.duration_ms", duration.Milliseconds()),
    )

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        return nil, err
    }

    // Record token usage
    if response.Usage != nil {
        span.SetAttributes(
            attribute.Int("claude.tokens.input", response.Usage.InputTokens),
            attribute.Int("claude.tokens.output", response.Usage.OutputTokens),
            attribute.Int("claude.tokens.total", response.Usage.InputTokens+response.Usage.OutputTokens),
        )
    }

    span.AddEvent("claude.response.received")
    span.SetStatus(codes.Ok, "Message sent successfully")

    return convertFromAnthropicResponse(response), nil
}
```

### Metrics Collection

#### Custom Metrics

```go
type MCPMetrics struct {
    // Session metrics
    SessionsActive    prometheus.Gauge
    SessionsTotal     prometheus.Counter
    SessionDuration   prometheus.Histogram

    // Tool metrics
    ToolCallsTotal    prometheus.Counter
    ToolCallDuration  prometheus.Histogram
    ToolCallErrors    prometheus.Counter

    // Resource metrics
    ResourceReads     prometheus.Counter
    ResourceReadSize  prometheus.Histogram

    // Claude API metrics
    ClaudeRequestsTotal    prometheus.Counter
    ClaudeRequestDuration  prometheus.Histogram
    ClaudeTokensUsed       prometheus.Counter
    ClaudeErrors           prometheus.Counter

    // MCP Protocol metrics
    MCPMessagesTotal       prometheus.Counter
    MCPMessageSize         prometheus.Histogram
    MCPProtocolErrors      prometheus.Counter
}

func NewMCPMetrics() *MCPMetrics {
    return &MCPMetrics{
        SessionsActive: prometheus.NewGauge(prometheus.GaugeOpts{
            Name: "mcp_sessions_active",
            Help: "Number of active MCP sessions",
        }),
        SessionsTotal: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "mcp_sessions_total",
            Help: "Total number of MCP sessions created",
        }),
        ToolCallsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "mcp_tool_calls_total",
            Help: "Total number of tool calls",
        }, []string{"tool_name", "status"}),
        ClaudeRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
            Name: "claude_requests_total",
            Help: "Total number of Claude API requests",
        }, []string{"model", "status"}),
        // ... other metrics
    }
}
```

#### Metrics Recording

```go
func (m *MCPMetrics) RecordSessionCreated() {
    m.SessionsTotal.Inc()
    m.SessionsActive.Inc()
}

func (m *MCPMetrics) RecordSessionClosed(duration time.Duration) {
    m.SessionsActive.Dec()
    m.SessionDuration.Observe(duration.Seconds())
}

func (m *MCPMetrics) RecordToolCall(toolName string, duration time.Duration, err error) {
    status := "success"
    if err != nil {
        status = "error"
        m.ToolCallErrors.Inc()
    }

    m.ToolCallsTotal.WithLabelValues(toolName, status).Inc()
    m.ToolCallDuration.WithLabelValues(toolName).Observe(duration.Seconds())
}
```

### Structured Logging

#### Logger Configuration

```go
type Logger struct {
    *zerolog.Logger
    level  zerolog.Level
    fields map[string]interface{}
}

func NewLogger(config LoggingConfig) *Logger {
    var output io.Writer = os.Stdout

    if config.Output != "stdout" && config.Output != "stderr" {
        if file, err := os.OpenFile(config.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err == nil {
            output = file
        }
    }

    var logger zerolog.Logger
    if config.Format == "json" || config.Structured {
        logger = zerolog.New(output).With().Timestamp().Logger()
    } else {
        logger = zerolog.New(zerolog.ConsoleWriter{Out: output}).With().Timestamp().Logger()
    }

    level, _ := zerolog.ParseLevel(config.Level)
    logger = logger.Level(level)

    return &Logger{
        Logger: &logger,
        level:  level,
        fields: make(map[string]interface{}),
    }
}
```

#### Contextual Logging

```go
// Add MCP context to logs
func (l *Logger) WithMCPContext(ctx context.Context) *Logger {
    newLogger := l.Logger.With()

    // Extract trace information
    if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
        newLogger = newLogger.
            Str("trace_id", span.SpanContext().TraceID().String()).
            Str("span_id", span.SpanContext().SpanID().String())
    }

    // Extract MCP-specific context
    if sessionID := ctx.Value(TraceKeySessionID); sessionID != nil {
        newLogger = newLogger.Str("session_id", sessionID.(string))
    }

    if conversationID := ctx.Value(TraceKeyConversationID); conversationID != nil {
        newLogger = newLogger.Str("conversation_id", conversationID.(string))
    }

    logger := newLogger.Logger()
    return &Logger{Logger: &logger, level: l.level, fields: l.fields}
}

// Structured logging for MCP operations
func (l *Logger) LogMCPRequest(method string, params interface{}) {
    l.Info().
        Str("mcp_method", method).
        Interface("mcp_params", params).
        Msg("MCP request received")
}

func (l *Logger) LogMCPResponse(method string, duration time.Duration, err error) {
    event := l.Info()
    if err != nil {
        event = l.Error().Err(err)
    }

    event.
        Str("mcp_method", method).
        Dur("duration", duration).
        Msg("MCP request completed")
}

func (l *Logger) LogClaudeRequest(model string, tokenCount int) {
    l.Info().
        Str("claude_model", model).
        Int("token_count", tokenCount).
        Msg("Claude API request")
}
```

### Health Checks and Monitoring

#### Health Check Endpoints

```go
type HealthChecker struct {
    claudeClient *ClaudeClient
    db          *sql.DB
    redis       *redis.Client
}

func (hc *HealthChecker) CheckHealth(ctx context.Context) *HealthStatus {
    status := &HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Checks:    make(map[string]CheckResult),
    }

    // Check Claude API connectivity
    if err := hc.checkClaude(ctx); err != nil {
        status.Checks["claude"] = CheckResult{
            Status: "unhealthy",
            Error:  err.Error(),
        }
        status.Status = "unhealthy"
    } else {
        status.Checks["claude"] = CheckResult{Status: "healthy"}
    }

    // Check database connectivity
    if err := hc.db.PingContext(ctx); err != nil {
        status.Checks["database"] = CheckResult{
            Status: "unhealthy",
            Error:  err.Error(),
        }
        status.Status = "unhealthy"
    } else {
        status.Checks["database"] = CheckResult{Status: "healthy"}
    }

    return status
}

type HealthStatus struct {
    Status    string                   `json:"status"`
    Timestamp time.Time               `json:"timestamp"`
    Checks    map[string]CheckResult  `json:"checks"`
}

type CheckResult struct {
    Status string `json:"status"`
    Error  string `json:"error,omitempty"`
}
```

### Error Tracking and Alerting

#### Error Aggregation

```go
type ErrorTracker struct {
    logger  *Logger
    metrics *MCPMetrics
}

func (et *ErrorTracker) TrackError(ctx context.Context, err error, operation string, metadata map[string]interface{}) {
    // Log structured error
    event := et.logger.WithMCPContext(ctx).Error().
        Err(err).
        Str("operation", operation)

    for key, value := range metadata {
        event = event.Interface(key, value)
    }

    event.Msg("Operation failed")

    // Record metrics
    et.metrics.MCPProtocolErrors.Inc()

    // Check for critical errors that need immediate attention
    if et.isCriticalError(err) {
        et.sendAlert(ctx, err, operation, metadata)
    }
}

func (et *ErrorTracker) isCriticalError(err error) bool {
    // Define critical error patterns
    criticalPatterns := []string{
        "database connection lost",
        "claude api key invalid",
        "out of memory",
        "disk space full",
    }

    errMsg := strings.ToLower(err.Error())
    for _, pattern := range criticalPatterns {
        if strings.Contains(errMsg, pattern) {
            return true
        }
    }

    return false
}
```

### Performance Monitoring

#### Request Tracing

```go
type PerformanceMonitor struct {
    slowRequestThreshold time.Duration
    logger              *Logger
    metrics             *MCPMetrics
}

func (pm *PerformanceMonitor) MonitorRequest(ctx context.Context, operation string, fn func() error) error {
    start := time.Now()

    err := fn()
    duration := time.Since(start)

    // Log slow requests
    if duration > pm.slowRequestThreshold {
        pm.logger.WithMCPContext(ctx).Warn().
            Str("operation", operation).
            Dur("duration", duration).
            Dur("threshold", pm.slowRequestThreshold).
            Msg("Slow request detected")
    }

    // Record performance metrics
    pm.recordPerformanceMetrics(operation, duration, err)

    return err
}

func (pm *PerformanceMonitor) recordPerformanceMetrics(operation string, duration time.Duration, err error) {
    // Record operation-specific metrics based on operation type
    switch {
    case strings.HasPrefix(operation, "mcp.tool"):
        pm.metrics.ToolCallDuration.Observe(duration.Seconds())
    case strings.HasPrefix(operation, "claude"):
        pm.metrics.ClaudeRequestDuration.Observe(duration.Seconds())
    }
}
```

### Observability Best Practices

#### Context Propagation

```go
// Always propagate context through the call chain
func (h *Handler) ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
    // Add request ID to context
    ctx = context.WithValue(ctx, TraceKeyRequestID, req.ID)

    // Start span for this operation
    ctx, span := StartMCPSpan(ctx, "request.process")
    defer span.End()

    // Pass context to downstream operations
    return h.service.Handle(ctx, req)
}
```

#### Sampling Strategy

```go
// Implement intelligent sampling for high-volume operations
type SamplingStrategy struct {
    defaultRate    float64
    operationRates map[string]float64
}

func (ss *SamplingStrategy) ShouldSample(operation string) bool {
    rate := ss.defaultRate
    if opRate, exists := ss.operationRates[operation]; exists {
        rate = opRate
    }

    return rand.Float64() < rate
}
```

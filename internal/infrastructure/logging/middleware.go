// Package logging provides request/response logging middleware.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package logging

import (
	"context"
	"encoding/json"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// RequestLogger provides request/response logging for MCP operations.
type RequestLogger struct {
	logger    *Logger
	mcpLogger *MCPLogger
	config    *RequestLoggerConfig
}

// RequestLoggerConfig configures the request logger.
type RequestLoggerConfig struct {
	// LogRequestBody enables logging of request bodies
	LogRequestBody bool
	// LogResponseBody enables logging of response bodies
	LogResponseBody bool
	// MaxBodySize limits the size of logged bodies
	MaxBodySize int
	// SlowRequestThreshold marks requests slower than this as slow
	SlowRequestThreshold time.Duration
	// SensitiveFields are fields to redact from logs
	SensitiveFields []string
	// IncludeTraceInfo adds trace/span IDs to logs
	IncludeTraceInfo bool
}

// DefaultRequestLoggerConfig returns default configuration.
func DefaultRequestLoggerConfig() *RequestLoggerConfig {
	return &RequestLoggerConfig{
		LogRequestBody:       true,
		LogResponseBody:      false,
		MaxBodySize:          4096,
		SlowRequestThreshold: 5 * time.Second,
		SensitiveFields:      []string{"api_key", "apiKey", "password", "secret", "token"},
		IncludeTraceInfo:     true,
	}
}

// NewRequestLogger creates a new request logger.
func NewRequestLogger(logger *Logger, mcpLogger *MCPLogger, config *RequestLoggerConfig) *RequestLogger {
	if config == nil {
		config = DefaultRequestLoggerConfig()
	}
	return &RequestLogger{
		logger:    logger,
		mcpLogger: mcpLogger,
		config:    config,
	}
}

// RequestInfo contains information about an MCP request.
type RequestInfo struct {
	ID        string
	Method    string
	SessionID string
	Params    interface{}
	StartTime time.Time
}

// ResponseInfo contains information about an MCP response.
type ResponseInfo struct {
	ID        string
	Method    string
	SessionID string
	Result    interface{}
	Error     error
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time
}

// LogRequest logs an incoming MCP request.
func (l *RequestLogger) LogRequest(ctx context.Context, info *RequestInfo) {
	event := l.logger.Info().
		Str("request_id", info.ID).
		Str("method", info.Method).
		Str("session_id", info.SessionID).
		Time("start_time", info.StartTime)

	if l.config.IncludeTraceInfo {
		if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
			event = event.
				Str("trace_id", spanCtx.TraceID().String()).
				Str("span_id", spanCtx.SpanID().String())
		}
	}

	if l.config.LogRequestBody && info.Params != nil {
		sanitized := l.sanitizeBody(info.Params)
		event = event.Interface("params", sanitized)
	}

	event.Msg("MCP request received")

	// Also log to MCP logger for clients
	if l.mcpLogger != nil {
		l.mcpLogger.Debug(ctx, map[string]interface{}{
			"type":       "request",
			"request_id": info.ID,
			"method":     info.Method,
			"session_id": info.SessionID,
		})
	}
}

// LogResponse logs an MCP response.
func (l *RequestLogger) LogResponse(ctx context.Context, info *ResponseInfo) {
	event := l.logger.Info().
		Str("request_id", info.ID).
		Str("method", info.Method).
		Str("session_id", info.SessionID).
		Dur("duration", info.Duration).
		Time("start_time", info.StartTime).
		Time("end_time", info.EndTime)

	if l.config.IncludeTraceInfo {
		if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
			event = event.
				Str("trace_id", spanCtx.TraceID().String()).
				Str("span_id", spanCtx.SpanID().String())
		}
	}

	// Check for slow request
	isSlow := info.Duration > l.config.SlowRequestThreshold
	event = event.Bool("slow_request", isSlow)

	// Handle errors
	if info.Error != nil {
		event = event.Err(info.Error)
		event.Msg("MCP request failed")
	} else {
		if l.config.LogResponseBody && info.Result != nil {
			sanitized := l.sanitizeBody(info.Result)
			event = event.Interface("result", sanitized)
		}
		event.Msg("MCP request completed")
	}

	// Log slow requests as warnings
	if isSlow {
		l.logger.Warn().
			Str("request_id", info.ID).
			Str("method", info.Method).
			Dur("duration", info.Duration).
			Msg("Slow MCP request detected")
	}

	// Also log to MCP logger
	if l.mcpLogger != nil {
		level := MCPLogLevelInfo
		if info.Error != nil {
			level = MCPLogLevelError
		} else if isSlow {
			level = MCPLogLevelWarning
		}

		l.mcpLogger.Log(ctx, level, map[string]interface{}{
			"type":        "response",
			"request_id":  info.ID,
			"method":      info.Method,
			"session_id":  info.SessionID,
			"duration_ms": info.Duration.Milliseconds(),
			"success":     info.Error == nil,
		})
	}
}

// LogToolCall logs a tool execution.
func (l *RequestLogger) LogToolCall(ctx context.Context, toolName string, input map[string]interface{}, result interface{}, err error, duration time.Duration) {
	event := l.logger.Info().
		Str("tool_name", toolName).
		Dur("duration", duration)

	if l.config.IncludeTraceInfo {
		if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
			event = event.
				Str("trace_id", spanCtx.TraceID().String()).
				Str("span_id", spanCtx.SpanID().String())
		}
	}

	if input != nil {
		sanitized := l.sanitizeBody(input)
		event = event.Interface("input", sanitized)
	}

	if err != nil {
		event = event.Err(err)
		event.Msg("Tool execution failed")
	} else {
		event.Msg("Tool executed successfully")
	}
}

// LogClaudeRequest logs a Claude API request.
func (l *RequestLogger) LogClaudeRequest(ctx context.Context, model string, inputTokens, outputTokens int, duration time.Duration, err error) {
	event := l.logger.Info().
		Str("model", model).
		Int("input_tokens", inputTokens).
		Int("output_tokens", outputTokens).
		Int("total_tokens", inputTokens+outputTokens).
		Dur("duration", duration)

	if l.config.IncludeTraceInfo {
		if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
			event = event.
				Str("trace_id", spanCtx.TraceID().String()).
				Str("span_id", spanCtx.SpanID().String())
		}
	}

	if err != nil {
		event = event.Err(err)
		event.Msg("Claude API request failed")
	} else {
		event.Msg("Claude API request completed")
	}
}

// LogSessionEvent logs session lifecycle events.
func (l *RequestLogger) LogSessionEvent(ctx context.Context, sessionID string, event string, details map[string]interface{}) {
	logEvent := l.logger.Info().
		Str("session_id", sessionID).
		Str("event", event)

	if l.config.IncludeTraceInfo {
		if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
			logEvent = logEvent.
				Str("trace_id", spanCtx.TraceID().String()).
				Str("span_id", spanCtx.SpanID().String())
		}
	}

	if details != nil {
		sanitized := l.sanitizeBody(details)
		logEvent = logEvent.Interface("details", sanitized)
	}

	logEvent.Msg("Session event")

	// Also log to MCP logger
	if l.mcpLogger != nil {
		l.mcpLogger.Info(ctx, map[string]interface{}{
			"type":       "session_event",
			"session_id": sessionID,
			"event":      event,
			"details":    details,
		})
	}
}

// sanitizeBody redacts sensitive fields from the body.
func (l *RequestLogger) sanitizeBody(body interface{}) interface{} {
	// Convert to JSON and back to map for sanitization
	data, err := json.Marshal(body)
	if err != nil {
		return "[unable to serialize]"
	}

	// Truncate if too large
	if len(data) > l.config.MaxBodySize {
		return "[body truncated]"
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		// Not a map, return as string
		return string(data)
	}

	l.redactSensitiveFields(result)
	return result
}

// redactSensitiveFields recursively redacts sensitive fields.
func (l *RequestLogger) redactSensitiveFields(data map[string]interface{}) {
	for key, value := range data {
		// Check if this is a sensitive field
		for _, sensitiveField := range l.config.SensitiveFields {
			if key == sensitiveField {
				data[key] = "[REDACTED]"
				break
			}
		}

		// Recursively check nested maps
		if nested, ok := value.(map[string]interface{}); ok {
			l.redactSensitiveFields(nested)
		}

		// Check arrays
		if arr, ok := value.([]interface{}); ok {
			for _, item := range arr {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					l.redactSensitiveFields(nestedMap)
				}
			}
		}
	}
}

// OperationLogger provides a convenient way to log operation timing.
type OperationLogger struct {
	logger    *RequestLogger
	ctx       context.Context
	operation string
	startTime time.Time
	fields    map[string]interface{}
}

// StartOperation starts timing an operation.
func (l *RequestLogger) StartOperation(ctx context.Context, operation string) *OperationLogger {
	return &OperationLogger{
		logger:    l,
		ctx:       ctx,
		operation: operation,
		startTime: time.Now(),
		fields:    make(map[string]interface{}),
	}
}

// WithField adds a field to the operation log.
func (o *OperationLogger) WithField(key string, value interface{}) *OperationLogger {
	o.fields[key] = value
	return o
}

// End logs the operation completion.
func (o *OperationLogger) End(err error) {
	duration := time.Since(o.startTime)

	event := o.logger.logger.Info().
		Str("operation", o.operation).
		Dur("duration", duration)

	for k, v := range o.fields {
		event = event.Interface(k, v)
	}

	if err != nil {
		event = event.Err(err)
		event.Msg("Operation failed")
	} else {
		event.Msg("Operation completed")
	}
}

// EndWithResult logs the operation completion with a result.
func (o *OperationLogger) EndWithResult(result interface{}, err error) {
	duration := time.Since(o.startTime)

	event := o.logger.logger.Info().
		Str("operation", o.operation).
		Dur("duration", duration)

	for k, v := range o.fields {
		event = event.Interface(k, v)
	}

	if err != nil {
		event = event.Err(err)
		event.Msg("Operation failed")
	} else {
		if result != nil {
			event = event.Interface("result", result)
		}
		event.Msg("Operation completed")
	}
}

// TFORequestLogger provides request/response logging using TFO SDK.
// It wraps the TFOAdapter to provide MCP-specific logging methods.
type TFORequestLogger struct {
	adapter *TFOAdapter
	config  *RequestLoggerConfig
}

// NewTFORequestLogger creates a new TFO-based request logger.
func NewTFORequestLogger(adapter *TFOAdapter, config *RequestLoggerConfig) *TFORequestLogger {
	if config == nil {
		config = DefaultRequestLoggerConfig()
	}
	return &TFORequestLogger{
		adapter: adapter,
		config:  config,
	}
}

// LogRequest logs an incoming MCP request using TFO SDK.
func (l *TFORequestLogger) LogRequest(ctx context.Context, info *RequestInfo) {
	l.adapter.LogMCPRequest(ctx, info.ID, info.Method, info.SessionID)
}

// LogResponse logs an MCP response using TFO SDK.
func (l *TFORequestLogger) LogResponse(ctx context.Context, info *ResponseInfo) {
	l.adapter.LogMCPResponse(ctx, info.ID, info.Method, info.SessionID, info.Duration, info.Error)
}

// LogToolCall logs a tool execution using TFO SDK.
func (l *TFORequestLogger) LogToolCall(ctx context.Context, toolName string, input map[string]interface{}, result interface{}, err error, duration time.Duration) {
	// Get session ID from context if available
	sessionID := ""
	if sid := ctx.Value("session_id"); sid != nil {
		if s, ok := sid.(string); ok {
			sessionID = s
		}
	}
	l.adapter.LogToolCall(ctx, toolName, sessionID, duration, err)
}

// LogClaudeRequest logs a Claude API request using TFO SDK.
func (l *TFORequestLogger) LogClaudeRequest(ctx context.Context, model string, inputTokens, outputTokens int, duration time.Duration, err error) {
	l.adapter.LogClaudeRequest(ctx, model, inputTokens, outputTokens, duration, err)
}

// LogSessionEvent logs session lifecycle events using TFO SDK.
func (l *TFORequestLogger) LogSessionEvent(ctx context.Context, sessionID string, event string, details map[string]interface{}) {
	l.adapter.LogSessionEvent(ctx, sessionID, event, details)
}

// StartOperation starts timing an operation using TFO SDK.
func (l *TFORequestLogger) StartOperation(ctx context.Context, operation string) *TFOOperationLogger {
	return &TFOOperationLogger{
		adapter:   l.adapter,
		ctx:       ctx,
		operation: operation,
		startTime: time.Now(),
		fields:    make(map[string]interface{}),
	}
}

// TFOOperationLogger provides operation timing using TFO SDK.
type TFOOperationLogger struct {
	adapter   *TFOAdapter
	ctx       context.Context
	operation string
	startTime time.Time
	fields    map[string]interface{}
	spanID    string
}

// WithField adds a field to the operation log.
func (o *TFOOperationLogger) WithField(key string, value interface{}) *TFOOperationLogger {
	o.fields[key] = value
	return o
}

// WithSpan starts a trace span for this operation.
func (o *TFOOperationLogger) WithSpan(kind string) *TFOOperationLogger {
	spanID, _ := o.adapter.StartSpan(o.ctx, o.operation, kind, o.fields)
	o.spanID = spanID
	return o
}

// End logs the operation completion.
func (o *TFOOperationLogger) End(err error) {
	duration := time.Since(o.startTime)

	// End span if started
	if o.spanID != "" {
		_ = o.adapter.EndSpan(o.ctx, o.spanID, err)
	}

	// Log the operation
	attrs := map[string]interface{}{
		"operation":   o.operation,
		"duration_ms": duration.Milliseconds(),
	}
	for k, v := range o.fields {
		attrs[k] = v
	}

	if err != nil {
		attrs["error"] = err.Error()
		o.adapter.Error(o.ctx, "Operation failed", attrs)
	} else {
		o.adapter.Info(o.ctx, "Operation completed", attrs)
	}

	// Record histogram
	_ = o.adapter.RecordHistogram(o.ctx, "operation.duration", float64(duration.Milliseconds()), "ms", map[string]interface{}{
		"operation": o.operation,
	})
}

// EndWithResult logs the operation completion with a result.
func (o *TFOOperationLogger) EndWithResult(result interface{}, err error) {
	duration := time.Since(o.startTime)

	// End span if started
	if o.spanID != "" {
		_ = o.adapter.EndSpan(o.ctx, o.spanID, err)
	}

	// Log the operation
	attrs := map[string]interface{}{
		"operation":   o.operation,
		"duration_ms": duration.Milliseconds(),
	}
	for k, v := range o.fields {
		attrs[k] = v
	}

	if err != nil {
		attrs["error"] = err.Error()
		o.adapter.Error(o.ctx, "Operation failed", attrs)
	} else {
		if result != nil {
			attrs["result"] = result
		}
		o.adapter.Info(o.ctx, "Operation completed", attrs)
	}

	// Record histogram
	_ = o.adapter.RecordHistogram(o.ctx, "operation.duration", float64(duration.Milliseconds()), "ms", map[string]interface{}{
		"operation": o.operation,
	})
}

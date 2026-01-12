// Package logging provides MCP protocol logging capability support.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package logging

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// MCPLogLevel represents MCP protocol log levels as per spec.
type MCPLogLevel string

const (
	// MCPLogLevelDebug for debugging information.
	MCPLogLevelDebug MCPLogLevel = "debug"
	// MCPLogLevelInfo for general information.
	MCPLogLevelInfo MCPLogLevel = "info"
	// MCPLogLevelNotice for normal but significant events.
	MCPLogLevelNotice MCPLogLevel = "notice"
	// MCPLogLevelWarning for warning conditions.
	MCPLogLevelWarning MCPLogLevel = "warning"
	// MCPLogLevelError for error conditions.
	MCPLogLevelError MCPLogLevel = "error"
	// MCPLogLevelCritical for critical conditions.
	MCPLogLevelCritical MCPLogLevel = "critical"
	// MCPLogLevelAlert for action must be taken immediately.
	MCPLogLevelAlert MCPLogLevel = "alert"
	// MCPLogLevelEmergency for system is unusable.
	MCPLogLevelEmergency MCPLogLevel = "emergency"
)

// MCPLogMessage represents a log message as per MCP protocol.
type MCPLogMessage struct {
	Level     MCPLogLevel            `json:"level"`
	Logger    string                 `json:"logger,omitempty"`
	Data      interface{}            `json:"data"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// MCPLogHandler handles MCP log messages.
type MCPLogHandler func(ctx context.Context, msg *MCPLogMessage)

// MCPLogger provides MCP protocol logging capability.
type MCPLogger struct {
	mu             sync.RWMutex
	name           string
	minLevel       MCPLogLevel
	handlers       []MCPLogHandler
	bufferSize     int
	buffer         []*MCPLogMessage
	bufferEnabled  bool
	flushInterval  time.Duration
	lastFlush      time.Time
	internalLogger *Logger
}

// MCPLoggerOption configures the MCPLogger.
type MCPLoggerOption func(*MCPLogger)

// WithMCPLoggerName sets the logger name.
func WithMCPLoggerName(name string) MCPLoggerOption {
	return func(l *MCPLogger) {
		l.name = name
	}
}

// WithMCPMinLevel sets the minimum log level.
func WithMCPMinLevel(level MCPLogLevel) MCPLoggerOption {
	return func(l *MCPLogger) {
		l.minLevel = level
	}
}

// WithMCPHandler adds a log handler.
func WithMCPHandler(handler MCPLogHandler) MCPLoggerOption {
	return func(l *MCPLogger) {
		l.handlers = append(l.handlers, handler)
	}
}

// WithMCPBuffer enables log buffering.
func WithMCPBuffer(size int, flushInterval time.Duration) MCPLoggerOption {
	return func(l *MCPLogger) {
		l.bufferEnabled = true
		l.bufferSize = size
		l.flushInterval = flushInterval
		l.buffer = make([]*MCPLogMessage, 0, size)
	}
}

// WithInternalLogger sets the internal logger for error reporting.
func WithInternalLogger(logger *Logger) MCPLoggerOption {
	return func(l *MCPLogger) {
		l.internalLogger = logger
	}
}

// NewMCPLogger creates a new MCP protocol logger.
func NewMCPLogger(opts ...MCPLoggerOption) *MCPLogger {
	l := &MCPLogger{
		name:      "mcp-server",
		minLevel:  MCPLogLevelInfo,
		handlers:  make([]MCPLogHandler, 0),
		lastFlush: time.Now(),
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

// Log logs a message at the specified level.
func (l *MCPLogger) Log(ctx context.Context, level MCPLogLevel, data interface{}, extra ...map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	msg := &MCPLogMessage{
		Level:     level,
		Logger:    l.name,
		Data:      data,
		Timestamp: time.Now(),
	}

	if len(extra) > 0 {
		msg.Extra = extra[0]
	}

	l.dispatch(ctx, msg)
}

// Debug logs a debug message.
func (l *MCPLogger) Debug(ctx context.Context, data interface{}, extra ...map[string]interface{}) {
	l.Log(ctx, MCPLogLevelDebug, data, extra...)
}

// Info logs an info message.
func (l *MCPLogger) Info(ctx context.Context, data interface{}, extra ...map[string]interface{}) {
	l.Log(ctx, MCPLogLevelInfo, data, extra...)
}

// Notice logs a notice message.
func (l *MCPLogger) Notice(ctx context.Context, data interface{}, extra ...map[string]interface{}) {
	l.Log(ctx, MCPLogLevelNotice, data, extra...)
}

// Warning logs a warning message.
func (l *MCPLogger) Warning(ctx context.Context, data interface{}, extra ...map[string]interface{}) {
	l.Log(ctx, MCPLogLevelWarning, data, extra...)
}

// Error logs an error message.
func (l *MCPLogger) Error(ctx context.Context, data interface{}, extra ...map[string]interface{}) {
	l.Log(ctx, MCPLogLevelError, data, extra...)
}

// Critical logs a critical message.
func (l *MCPLogger) Critical(ctx context.Context, data interface{}, extra ...map[string]interface{}) {
	l.Log(ctx, MCPLogLevelCritical, data, extra...)
}

// Alert logs an alert message.
func (l *MCPLogger) Alert(ctx context.Context, data interface{}, extra ...map[string]interface{}) {
	l.Log(ctx, MCPLogLevelAlert, data, extra...)
}

// Emergency logs an emergency message.
func (l *MCPLogger) Emergency(ctx context.Context, data interface{}, extra ...map[string]interface{}) {
	l.Log(ctx, MCPLogLevelEmergency, data, extra...)
}

// SetLevel sets the minimum log level.
func (l *MCPLogger) SetLevel(level MCPLogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// GetLevel returns the current minimum log level.
func (l *MCPLogger) GetLevel() MCPLogLevel {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.minLevel
}

// AddHandler adds a log handler.
func (l *MCPLogger) AddHandler(handler MCPLogHandler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handlers = append(l.handlers, handler)
}

// Flush flushes buffered log messages.
func (l *MCPLogger) Flush(ctx context.Context) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.bufferEnabled || len(l.buffer) == 0 {
		return
	}

	for _, msg := range l.buffer {
		l.dispatchToHandlers(ctx, msg)
	}

	l.buffer = l.buffer[:0]
	l.lastFlush = time.Now()
}

// dispatch handles message dispatch (with optional buffering).
func (l *MCPLogger) dispatch(ctx context.Context, msg *MCPLogMessage) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.bufferEnabled {
		l.buffer = append(l.buffer, msg)
		if len(l.buffer) >= l.bufferSize || time.Since(l.lastFlush) >= l.flushInterval {
			for _, m := range l.buffer {
				l.dispatchToHandlers(ctx, m)
			}
			l.buffer = l.buffer[:0]
			l.lastFlush = time.Now()
		}
		return
	}

	l.dispatchToHandlers(ctx, msg)
}

// dispatchToHandlers sends message to all handlers.
func (l *MCPLogger) dispatchToHandlers(ctx context.Context, msg *MCPLogMessage) {
	for _, handler := range l.handlers {
		handler(ctx, msg)
	}
}

// shouldLog determines if a message should be logged based on level.
func (l *MCPLogger) shouldLog(level MCPLogLevel) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return levelPriority(level) >= levelPriority(l.minLevel)
}

// levelPriority returns numeric priority for level comparison.
func levelPriority(level MCPLogLevel) int {
	switch level {
	case MCPLogLevelDebug:
		return 0
	case MCPLogLevelInfo:
		return 1
	case MCPLogLevelNotice:
		return 2
	case MCPLogLevelWarning:
		return 3
	case MCPLogLevelError:
		return 4
	case MCPLogLevelCritical:
		return 5
	case MCPLogLevelAlert:
		return 6
	case MCPLogLevelEmergency:
		return 7
	default:
		return 1
	}
}

// ToJSON converts the log message to JSON.
func (m *MCPLogMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// MCPLoggingCapability represents the MCP logging capability.
type MCPLoggingCapability struct {
	logger *MCPLogger
}

// NewMCPLoggingCapability creates a new MCP logging capability.
func NewMCPLoggingCapability(logger *MCPLogger) *MCPLoggingCapability {
	return &MCPLoggingCapability{
		logger: logger,
	}
}

// SetLogLevel handles the logging/setLevel request.
func (c *MCPLoggingCapability) SetLogLevel(ctx context.Context, level MCPLogLevel) error {
	c.logger.SetLevel(level)
	return nil
}

// GetLogLevel returns the current log level.
func (c *MCPLoggingCapability) GetLogLevel() MCPLogLevel {
	return c.logger.GetLevel()
}

// Logger returns the underlying MCPLogger.
func (c *MCPLoggingCapability) Logger() *MCPLogger {
	return c.logger
}

// MCPNotificationHandler creates a handler that sends log messages as MCP notifications.
func MCPNotificationHandler(sender func(ctx context.Context, method string, params interface{}) error) MCPLogHandler {
	return func(ctx context.Context, msg *MCPLogMessage) {
		params := map[string]interface{}{
			"level":  msg.Level,
			"logger": msg.Logger,
			"data":   msg.Data,
		}
		if msg.Extra != nil {
			for k, v := range msg.Extra {
				params[k] = v
			}
		}

		// Silently ignore send errors to avoid infinite loops
		_ = sender(ctx, "notifications/message", params)
	}
}

// InternalLoggerHandler creates a handler that logs to the internal logger.
func InternalLoggerHandler(logger *Logger) MCPLogHandler {
	return func(ctx context.Context, msg *MCPLogMessage) {
		event := logger.Info()

		switch msg.Level {
		case MCPLogLevelDebug:
			event = logger.Debug()
		case MCPLogLevelInfo, MCPLogLevelNotice:
			event = logger.Info()
		case MCPLogLevelWarning:
			event = logger.Warn()
		case MCPLogLevelError, MCPLogLevelCritical, MCPLogLevelAlert, MCPLogLevelEmergency:
			event = logger.Error()
		}

		event.
			Str("mcp_level", string(msg.Level)).
			Str("mcp_logger", msg.Logger).
			Interface("data", msg.Data).
			Msg("MCP log message")
	}
}

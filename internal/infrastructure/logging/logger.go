// Package logging provides structured logging infrastructure for the MCP server.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package logging

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// LogLevel represents the severity level of log messages.
type LogLevel int8

const (
	// DebugLevel defines debug log level.
	DebugLevel LogLevel = iota
	// InfoLevel defines info log level.
	InfoLevel
	// WarnLevel defines warn log level.
	WarnLevel
	// ErrorLevel defines error log level.
	ErrorLevel
	// FatalLevel defines fatal log level.
	FatalLevel
	// PanicLevel defines panic log level.
	PanicLevel
	// NoLevel defines an absent log level.
	NoLevel
	// Disabled disables the logger.
	Disabled
	// TraceLevel defines trace log level.
	TraceLevel LogLevel = -1
)

// Logger wraps zerolog.Logger with additional functionality.
type Logger struct {
	logger zerolog.Logger
	level  LogLevel
}

// LoggerOption configures the Logger.
type LoggerOption func(*loggerConfig)

type loggerConfig struct {
	level       LogLevel
	output      io.Writer
	timeFormat  string
	serviceName string
	version     string
	prettyPrint bool
	caller      bool
	hooks       []zerolog.Hook
}

// WithLevel sets the log level.
func WithLevel(level LogLevel) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.level = level
	}
}

// WithOutput sets the output writer.
func WithOutput(w io.Writer) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.output = w
	}
}

// WithTimeFormat sets the time format.
func WithTimeFormat(format string) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.timeFormat = format
	}
}

// WithServiceName sets the service name in logs.
func WithServiceName(name string) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.serviceName = name
	}
}

// WithVersion sets the service version in logs.
func WithVersion(version string) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.version = version
	}
}

// WithPrettyPrint enables console-friendly output.
func WithPrettyPrint(enabled bool) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.prettyPrint = enabled
	}
}

// WithCaller includes caller information in logs.
func WithCaller(enabled bool) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.caller = enabled
	}
}

// WithHook adds a zerolog hook.
func WithHook(hook zerolog.Hook) LoggerOption {
	return func(cfg *loggerConfig) {
		cfg.hooks = append(cfg.hooks, hook)
	}
}

// NewLogger creates a new Logger with the given options.
func NewLogger(opts ...LoggerOption) *Logger {
	cfg := &loggerConfig{
		level:      InfoLevel,
		output:     os.Stdout,
		timeFormat: time.RFC3339,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// Set global time format
	zerolog.TimeFieldFormat = cfg.timeFormat

	// Configure output
	var output = cfg.output
	if cfg.prettyPrint {
		output = zerolog.ConsoleWriter{
			Out:        cfg.output,
			TimeFormat: cfg.timeFormat,
		}
	}

	// Create logger context
	logCtx := zerolog.New(output).With().Timestamp()

	// Add service info
	if cfg.serviceName != "" {
		logCtx = logCtx.Str("service", cfg.serviceName)
	}
	if cfg.version != "" {
		logCtx = logCtx.Str("version", cfg.version)
	}

	// Add caller if enabled
	if cfg.caller {
		logCtx = logCtx.Caller()
	}

	logger := logCtx.Logger().Level(toZerologLevel(cfg.level))

	// Add hooks
	for _, hook := range cfg.hooks {
		logger = logger.Hook(hook)
	}

	return &Logger{
		logger: logger,
		level:  cfg.level,
	}
}

// NewNopLogger creates a no-operation logger that discards all logs.
func NewNopLogger() *Logger {
	return &Logger{
		logger: zerolog.Nop(),
		level:  Disabled,
	}
}

// With creates a child logger with additional fields.
func (l *Logger) With() *LogContext {
	return &LogContext{
		ctx: l.logger.With(),
	}
}

// Level returns the current log level.
func (l *Logger) Level() LogLevel {
	return l.level
}

// SetLevel dynamically changes the log level.
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
	l.logger = l.logger.Level(toZerologLevel(level))
}

// Zerolog returns the underlying zerolog.Logger.
func (l *Logger) Zerolog() zerolog.Logger {
	return l.logger
}

// Debug logs a debug message.
func (l *Logger) Debug() *zerolog.Event {
	return l.logger.Debug()
}

// Info logs an info message.
func (l *Logger) Info() *zerolog.Event {
	return l.logger.Info()
}

// Warn logs a warning message.
func (l *Logger) Warn() *zerolog.Event {
	return l.logger.Warn()
}

// Error logs an error message.
func (l *Logger) Error() *zerolog.Event {
	return l.logger.Error()
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal() *zerolog.Event {
	return l.logger.Fatal()
}

// Panic logs a panic message and panics.
func (l *Logger) Panic() *zerolog.Event {
	return l.logger.Panic()
}

// Trace logs a trace message.
func (l *Logger) Trace() *zerolog.Event {
	return l.logger.Trace()
}

// WithContext returns a logger with context.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		logger: l.logger.With().Logger(),
		level:  l.level,
	}
}

// WithError returns a logger with error field.
func (l *Logger) WithError(err error) *zerolog.Event {
	return l.logger.Err(err)
}

// WithField returns a logger with a single field.
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return &Logger{
		logger: l.logger.With().Interface(key, value).Logger(),
		level:  l.level,
	}
}

// WithFields returns a logger with multiple fields.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	ctx := l.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	return &Logger{
		logger: ctx.Logger(),
		level:  l.level,
	}
}

// LogContext provides a fluent API for building loggers with context.
type LogContext struct {
	ctx zerolog.Context
}

// Str adds a string field.
func (c *LogContext) Str(key, val string) *LogContext {
	c.ctx = c.ctx.Str(key, val)
	return c
}

// Int adds an int field.
func (c *LogContext) Int(key string, val int) *LogContext {
	c.ctx = c.ctx.Int(key, val)
	return c
}

// Int64 adds an int64 field.
func (c *LogContext) Int64(key string, val int64) *LogContext {
	c.ctx = c.ctx.Int64(key, val)
	return c
}

// Float64 adds a float64 field.
func (c *LogContext) Float64(key string, val float64) *LogContext {
	c.ctx = c.ctx.Float64(key, val)
	return c
}

// Bool adds a bool field.
func (c *LogContext) Bool(key string, val bool) *LogContext {
	c.ctx = c.ctx.Bool(key, val)
	return c
}

// Time adds a time field.
func (c *LogContext) Time(key string, val time.Time) *LogContext {
	c.ctx = c.ctx.Time(key, val)
	return c
}

// Dur adds a duration field.
func (c *LogContext) Dur(key string, val time.Duration) *LogContext {
	c.ctx = c.ctx.Dur(key, val)
	return c
}

// Err adds an error field.
func (c *LogContext) Err(err error) *LogContext {
	c.ctx = c.ctx.Err(err)
	return c
}

// Interface adds an interface field.
func (c *LogContext) Interface(key string, val interface{}) *LogContext {
	c.ctx = c.ctx.Interface(key, val)
	return c
}

// Logger returns a new Logger with the context fields.
func (c *LogContext) Logger() *Logger {
	return &Logger{
		logger: c.ctx.Logger(),
	}
}

// toZerologLevel converts LogLevel to zerolog.Level.
func toZerologLevel(level LogLevel) zerolog.Level {
	switch level {
	case TraceLevel:
		return zerolog.TraceLevel
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	case FatalLevel:
		return zerolog.FatalLevel
	case PanicLevel:
		return zerolog.PanicLevel
	case NoLevel:
		return zerolog.NoLevel
	case Disabled:
		return zerolog.Disabled
	default:
		return zerolog.InfoLevel
	}
}

// ParseLevel parses a level string and returns a LogLevel.
func ParseLevel(levelStr string) LogLevel {
	switch levelStr {
	case "trace":
		return TraceLevel
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn", "warning":
		return WarnLevel
	case "error":
		return ErrorLevel
	case "fatal":
		return FatalLevel
	case "panic":
		return PanicLevel
	case "disabled":
		return Disabled
	default:
		return InfoLevel
	}
}

// String returns the string representation of the log level.
func (l LogLevel) String() string {
	switch l {
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	case PanicLevel:
		return "panic"
	case Disabled:
		return "disabled"
	default:
		return "info"
	}
}

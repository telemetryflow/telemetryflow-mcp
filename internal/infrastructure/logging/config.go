// Package logging provides configuration for the logging infrastructure.
//
// TelemetryFlow MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Config holds the complete logging configuration.
type Config struct {
	// Level is the minimum log level
	Level string `mapstructure:"level" yaml:"level" json:"level"`
	// Format is the output format (json or console)
	Format string `mapstructure:"format" yaml:"format" json:"format"`
	// Output is the output destination (stdout, stderr, or file path)
	Output string `mapstructure:"output" yaml:"output" json:"output"`
	// TimeFormat is the timestamp format
	TimeFormat string `mapstructure:"time_format" yaml:"time_format" json:"time_format"`
	// IncludeCaller adds caller information to logs
	IncludeCaller bool `mapstructure:"include_caller" yaml:"include_caller" json:"include_caller"`
	// ServiceName is included in all log entries
	ServiceName string `mapstructure:"service_name" yaml:"service_name" json:"service_name"`
	// ServiceVersion is included in all log entries
	ServiceVersion string `mapstructure:"service_version" yaml:"service_version" json:"service_version"`
	// File contains file-specific configuration
	File *FileConfig `mapstructure:"file" yaml:"file" json:"file"`
	// MCP contains MCP-specific logging configuration
	MCP *MCPConfig `mapstructure:"mcp" yaml:"mcp" json:"mcp"`
	// Request contains request logging configuration
	Request *RequestConfig `mapstructure:"request" yaml:"request" json:"request"`
}

// FileConfig holds file logging configuration.
type FileConfig struct {
	// Path is the log file path
	Path string `mapstructure:"path" yaml:"path" json:"path"`
	// MaxSize is the maximum size in megabytes before rotation
	MaxSize int `mapstructure:"max_size" yaml:"max_size" json:"max_size"`
	// MaxBackups is the maximum number of old log files to retain
	MaxBackups int `mapstructure:"max_backups" yaml:"max_backups" json:"max_backups"`
	// MaxAge is the maximum number of days to retain old log files
	MaxAge int `mapstructure:"max_age" yaml:"max_age" json:"max_age"`
	// Compress determines if rotated files should be compressed
	Compress bool `mapstructure:"compress" yaml:"compress" json:"compress"`
}

// MCPConfig holds MCP logging capability configuration.
type MCPConfig struct {
	// Enabled enables MCP logging capability
	Enabled bool `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	// Level is the minimum MCP log level
	Level string `mapstructure:"level" yaml:"level" json:"level"`
	// BufferSize is the size of the log buffer (0 disables buffering)
	BufferSize int `mapstructure:"buffer_size" yaml:"buffer_size" json:"buffer_size"`
	// FlushInterval is how often to flush the buffer
	FlushInterval time.Duration `mapstructure:"flush_interval" yaml:"flush_interval" json:"flush_interval"`
	// LoggerName is the name reported in MCP log messages
	LoggerName string `mapstructure:"logger_name" yaml:"logger_name" json:"logger_name"`
}

// RequestConfig holds request logging configuration.
type RequestConfig struct {
	// LogRequestBody enables logging of request bodies
	LogRequestBody bool `mapstructure:"log_request_body" yaml:"log_request_body" json:"log_request_body"`
	// LogResponseBody enables logging of response bodies
	LogResponseBody bool `mapstructure:"log_response_body" yaml:"log_response_body" json:"log_response_body"`
	// MaxBodySize limits the size of logged bodies
	MaxBodySize int `mapstructure:"max_body_size" yaml:"max_body_size" json:"max_body_size"`
	// SlowRequestThreshold marks requests slower than this as slow
	SlowRequestThreshold time.Duration `mapstructure:"slow_request_threshold" yaml:"slow_request_threshold" json:"slow_request_threshold"`
	// SensitiveFields are fields to redact from logs
	SensitiveFields []string `mapstructure:"sensitive_fields" yaml:"sensitive_fields" json:"sensitive_fields"`
	// IncludeTraceInfo adds trace/span IDs to logs
	IncludeTraceInfo bool `mapstructure:"include_trace_info" yaml:"include_trace_info" json:"include_trace_info"`
}

// DefaultConfig returns the default logging configuration.
func DefaultConfig() *Config {
	return &Config{
		Level:          "info",
		Format:         "json",
		Output:         "stdout",
		TimeFormat:     time.RFC3339,
		IncludeCaller:  false,
		ServiceName:    "telemetryflow-mcp",
		ServiceVersion: "0.1.0",
		File: &FileConfig{
			Path:       "/var/log/telemetryflow-mcp/server.log",
			MaxSize:    100, // megabytes
			MaxBackups: 3,
			MaxAge:     28, // days
			Compress:   true,
		},
		MCP: &MCPConfig{
			Enabled:       true,
			Level:         "info",
			BufferSize:    100,
			FlushInterval: 5 * time.Second,
			LoggerName:    "telemetryflow-mcp",
		},
		Request: &RequestConfig{
			LogRequestBody:       true,
			LogResponseBody:      false,
			MaxBodySize:          4096,
			SlowRequestThreshold: 5 * time.Second,
			SensitiveFields:      []string{"api_key", "apiKey", "password", "secret", "token", "authorization"},
			IncludeTraceInfo:     true,
		},
	}
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Validate log level
	validLevels := map[string]bool{
		"trace": true, "debug": true, "info": true, "warn": true,
		"warning": true, "error": true, "fatal": true, "panic": true,
		"disabled": true,
	}
	if !validLevels[c.Level] {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	// Validate format
	if c.Format != "json" && c.Format != "console" {
		return fmt.Errorf("invalid log format: %s (must be 'json' or 'console')", c.Format)
	}

	// Validate output
	if c.Output != "stdout" && c.Output != "stderr" && c.Output != "file" {
		return fmt.Errorf("invalid output: %s (must be 'stdout', 'stderr', or 'file')", c.Output)
	}

	// Validate file config if output is file
	if c.Output == "file" {
		if c.File == nil || c.File.Path == "" {
			return fmt.Errorf("file path is required when output is 'file'")
		}
	}

	// Validate MCP config
	if c.MCP != nil && c.MCP.Enabled {
		validMCPLevels := map[string]bool{
			"debug": true, "info": true, "notice": true, "warning": true,
			"error": true, "critical": true, "alert": true, "emergency": true,
		}
		if !validMCPLevels[c.MCP.Level] {
			return fmt.Errorf("invalid MCP log level: %s", c.MCP.Level)
		}
	}

	return nil
}

// BuildLogger creates a Logger from the configuration.
func (c *Config) BuildLogger() (*Logger, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logging config: %w", err)
	}

	opts := []LoggerOption{
		WithLevel(ParseLevel(c.Level)),
		WithTimeFormat(c.TimeFormat),
		WithPrettyPrint(c.Format == "console"),
		WithCaller(c.IncludeCaller),
	}

	if c.ServiceName != "" {
		opts = append(opts, WithServiceName(c.ServiceName))
	}
	if c.ServiceVersion != "" {
		opts = append(opts, WithVersion(c.ServiceVersion))
	}

	// Configure output
	output, err := c.buildOutput()
	if err != nil {
		return nil, err
	}
	opts = append(opts, WithOutput(output))

	return NewLogger(opts...), nil
}

// BuildMCPLogger creates an MCPLogger from the configuration.
func (c *Config) BuildMCPLogger(internalLogger *Logger) *MCPLogger {
	if c.MCP == nil || !c.MCP.Enabled {
		return nil
	}

	opts := []MCPLoggerOption{
		WithMCPLoggerName(c.MCP.LoggerName),
		WithMCPMinLevel(MCPLogLevel(c.MCP.Level)),
		WithInternalLogger(internalLogger),
	}

	if c.MCP.BufferSize > 0 {
		opts = append(opts, WithMCPBuffer(c.MCP.BufferSize, c.MCP.FlushInterval))
	}

	// Add internal logger handler
	if internalLogger != nil {
		opts = append(opts, WithMCPHandler(InternalLoggerHandler(internalLogger)))
	}

	return NewMCPLogger(opts...)
}

// BuildRequestLogger creates a RequestLogger from the configuration.
func (c *Config) BuildRequestLogger(logger *Logger, mcpLogger *MCPLogger) *RequestLogger {
	if c.Request == nil {
		return NewRequestLogger(logger, mcpLogger, DefaultRequestLoggerConfig())
	}

	config := &RequestLoggerConfig{
		LogRequestBody:       c.Request.LogRequestBody,
		LogResponseBody:      c.Request.LogResponseBody,
		MaxBodySize:          c.Request.MaxBodySize,
		SlowRequestThreshold: c.Request.SlowRequestThreshold,
		SensitiveFields:      c.Request.SensitiveFields,
		IncludeTraceInfo:     c.Request.IncludeTraceInfo,
	}

	return NewRequestLogger(logger, mcpLogger, config)
}

// buildOutput creates the output writer based on configuration.
func (c *Config) buildOutput() (io.Writer, error) {
	switch c.Output {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	case "file":
		return c.buildFileOutput()
	default:
		return os.Stdout, nil
	}
}

// buildFileOutput creates a file output with rotation.
func (c *Config) buildFileOutput() (io.Writer, error) {
	if c.File == nil {
		return nil, fmt.Errorf("file configuration is required for file output")
	}

	// Ensure directory exists
	dir := filepath.Dir(c.File.Path)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create log directory %s: %w", dir, err)
	}

	return &lumberjack.Logger{
		Filename:   c.File.Path,
		MaxSize:    c.File.MaxSize,
		MaxBackups: c.File.MaxBackups,
		MaxAge:     c.File.MaxAge,
		Compress:   c.File.Compress,
	}, nil
}

// LoggingSetup contains all configured loggers.
type LoggingSetup struct {
	Logger        *Logger
	MCPLogger     *MCPLogger
	RequestLogger *RequestLogger
}

// Setup creates all loggers from the configuration.
func (c *Config) Setup() (*LoggingSetup, error) {
	logger, err := c.BuildLogger()
	if err != nil {
		return nil, err
	}

	mcpLogger := c.BuildMCPLogger(logger)
	requestLogger := c.BuildRequestLogger(logger, mcpLogger)

	return &LoggingSetup{
		Logger:        logger,
		MCPLogger:     mcpLogger,
		RequestLogger: requestLogger,
	}, nil
}

// Close closes all loggers and flushes any buffered data.
func (s *LoggingSetup) Close() {
	// Nothing to close for basic loggers
	// File-based outputs handle their own cleanup
}

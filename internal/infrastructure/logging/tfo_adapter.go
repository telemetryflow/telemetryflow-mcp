// Package logging provides TFO Go SDK integration for observability.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package logging

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/telemetryflow/telemetryflow-go-sdk/pkg/telemetryflow"
)

// TFOAdapter wraps the TelemetryFlow Go SDK client for MCP server observability.
// It provides a unified interface for logging, metrics, and tracing that sends
// all telemetry through the TFO platform pipeline.
type TFOAdapter struct {
	client         *telemetryflow.Client
	fallbackLogger *Logger
	serviceName    string
	serviceVersion string
	environment    string
	initialized    bool
	mu             sync.RWMutex
}

// TFOAdapterConfig configures the TFO adapter.
type TFOAdapterConfig struct {
	// APIKeyID is the TelemetryFlow API key ID (tfk_xxx)
	APIKeyID string `mapstructure:"api_key_id" yaml:"api_key_id" json:"api_key_id"`
	// APIKeySecret is the TelemetryFlow API key secret (tfs_xxx)
	APIKeySecret string `mapstructure:"api_key_secret" yaml:"api_key_secret" json:"api_key_secret"`
	// Endpoint is the TelemetryFlow collector endpoint
	Endpoint string `mapstructure:"endpoint" yaml:"endpoint" json:"endpoint"`
	// ServiceName identifies this service
	ServiceName string `mapstructure:"service_name" yaml:"service_name" json:"service_name"`
	// ServiceVersion is the version of this service
	ServiceVersion string `mapstructure:"service_version" yaml:"service_version" json:"service_version"`
	// ServiceNamespace groups related services
	ServiceNamespace string `mapstructure:"service_namespace" yaml:"service_namespace" json:"service_namespace"`
	// Environment is the deployment environment (production, staging, etc.)
	Environment string `mapstructure:"environment" yaml:"environment" json:"environment"`
	// CollectorID identifies the collector instance
	CollectorID string `mapstructure:"collector_id" yaml:"collector_id" json:"collector_id"`
	// CollectorName is the human-readable collector name
	CollectorName string `mapstructure:"collector_name" yaml:"collector_name" json:"collector_name"`
	// EnableMetrics enables metrics collection
	EnableMetrics bool `mapstructure:"enable_metrics" yaml:"enable_metrics" json:"enable_metrics"`
	// EnableLogs enables log collection
	EnableLogs bool `mapstructure:"enable_logs" yaml:"enable_logs" json:"enable_logs"`
	// EnableTraces enables trace collection
	EnableTraces bool `mapstructure:"enable_traces" yaml:"enable_traces" json:"enable_traces"`
	// UseGRPC uses gRPC protocol (default), otherwise HTTP
	UseGRPC bool `mapstructure:"use_grpc" yaml:"use_grpc" json:"use_grpc"`
	// Insecure disables TLS
	Insecure bool `mapstructure:"insecure" yaml:"insecure" json:"insecure"`
	// Timeout is the connection timeout
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout" json:"timeout"`
	// FallbackToLocal enables local zerolog fallback when SDK is unavailable
	FallbackToLocal bool `mapstructure:"fallback_to_local" yaml:"fallback_to_local" json:"fallback_to_local"`
}

// DefaultTFOAdapterConfig returns default configuration.
func DefaultTFOAdapterConfig() *TFOAdapterConfig {
	return &TFOAdapterConfig{
		Endpoint:         "api.telemetryflow.id:4317",
		ServiceName:      "telemetryflow-go-mcp",
		ServiceVersion:   "0.1.0",
		ServiceNamespace: "telemetryflow",
		Environment:      "production",
		EnableMetrics:    true,
		EnableLogs:       true,
		EnableTraces:     true,
		UseGRPC:          true,
		Insecure:         false,
		Timeout:          30 * time.Second,
		FallbackToLocal:  true,
	}
}

// NewTFOAdapter creates a new TFO SDK adapter.
func NewTFOAdapter(cfg *TFOAdapterConfig) (*TFOAdapter, error) {
	if cfg == nil {
		cfg = DefaultTFOAdapterConfig()
	}

	adapter := &TFOAdapter{
		serviceName:    cfg.ServiceName,
		serviceVersion: cfg.ServiceVersion,
		environment:    cfg.Environment,
	}

	// Create fallback logger for local logging
	if cfg.FallbackToLocal {
		adapter.fallbackLogger = NewLogger(
			WithServiceName(cfg.ServiceName),
			WithVersion(cfg.ServiceVersion),
			WithLevel(InfoLevel),
		)
	}

	// Build TFO SDK client
	builder := telemetryflow.NewBuilder().
		WithAPIKey(cfg.APIKeyID, cfg.APIKeySecret).
		WithEndpoint(cfg.Endpoint).
		WithService(cfg.ServiceName, cfg.ServiceVersion).
		WithServiceNamespace(cfg.ServiceNamespace).
		WithEnvironment(cfg.Environment).
		WithSignals(cfg.EnableMetrics, cfg.EnableLogs, cfg.EnableTraces).
		WithInsecure(cfg.Insecure).
		WithTimeout(cfg.Timeout)

	if cfg.UseGRPC {
		builder = builder.WithGRPC()
	} else {
		builder = builder.WithHTTP()
	}

	if cfg.CollectorID != "" {
		builder = builder.WithCollectorID(cfg.CollectorID)
	}
	if cfg.CollectorName != "" {
		builder = builder.WithCollectorName(cfg.CollectorName)
	}

	// Add MCP-specific attributes
	builder = builder.
		WithCustomAttribute("mcp.protocol.version", "2024-11-05").
		WithCustomAttribute("mcp.server.type", "telemetryflow")

	client, err := builder.Build()
	if err != nil {
		if cfg.FallbackToLocal && adapter.fallbackLogger != nil {
			adapter.fallbackLogger.Warn().
				Err(err).
				Msg("Failed to create TFO SDK client, using local fallback")
			return adapter, nil
		}
		return nil, fmt.Errorf("failed to create TFO SDK client: %w", err)
	}

	adapter.client = client
	return adapter, nil
}

// NewTFOAdapterFromEnv creates a TFO adapter using environment variables.
func NewTFOAdapterFromEnv() (*TFOAdapter, error) {
	cfg := DefaultTFOAdapterConfig()
	cfg.FallbackToLocal = true

	adapter := &TFOAdapter{
		serviceName:    cfg.ServiceName,
		serviceVersion: cfg.ServiceVersion,
		environment:    cfg.Environment,
		fallbackLogger: NewLogger(
			WithServiceName(cfg.ServiceName),
			WithVersion(cfg.ServiceVersion),
			WithLevel(InfoLevel),
		),
	}

	// Try to create client from env
	client, err := telemetryflow.NewFromEnv()
	if err != nil {
		adapter.fallbackLogger.Warn().
			Err(err).
			Msg("Failed to create TFO SDK client from env, using local fallback")
		return adapter, nil
	}

	adapter.client = client
	return adapter, nil
}

// Initialize initializes the TFO SDK.
func (a *TFOAdapter) Initialize(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.initialized {
		return nil
	}

	if a.client == nil {
		a.log(ctx, "info", "TFO SDK not available, using local fallback", nil)
		a.initialized = true
		return nil
	}

	if err := a.client.Initialize(ctx); err != nil {
		a.logLocal(zerolog.WarnLevel, "Failed to initialize TFO SDK, using local fallback", map[string]interface{}{
			"error": err.Error(),
		})
		return nil // Don't fail, use fallback
	}

	a.initialized = true
	a.log(ctx, "info", "TFO SDK initialized successfully", map[string]interface{}{
		"service":     a.serviceName,
		"version":     a.serviceVersion,
		"environment": a.environment,
	})

	return nil
}

// Shutdown gracefully shuts down the TFO SDK.
func (a *TFOAdapter) Shutdown(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.initialized {
		return nil
	}

	if a.client != nil {
		if err := a.client.Shutdown(ctx); err != nil {
			a.logLocal(zerolog.ErrorLevel, "Failed to shutdown TFO SDK", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	a.initialized = false
	return nil
}

// Flush forces a flush of pending telemetry.
func (a *TFOAdapter) Flush(ctx context.Context) error {
	if a.client != nil && a.isInitialized() {
		return a.client.Flush(ctx)
	}
	return nil
}

// ===== LOGGING API =====

// Log emits a structured log entry.
func (a *TFOAdapter) Log(ctx context.Context, severity string, message string, attributes map[string]interface{}) {
	a.log(ctx, severity, message, attributes)
}

// Debug logs a debug message.
func (a *TFOAdapter) Debug(ctx context.Context, message string, attributes map[string]interface{}) {
	a.log(ctx, "debug", message, attributes)
}

// Info logs an info message.
func (a *TFOAdapter) Info(ctx context.Context, message string, attributes map[string]interface{}) {
	a.log(ctx, "info", message, attributes)
}

// Warn logs a warning message.
func (a *TFOAdapter) Warn(ctx context.Context, message string, attributes map[string]interface{}) {
	a.log(ctx, "warn", message, attributes)
}

// Error logs an error message.
func (a *TFOAdapter) Error(ctx context.Context, message string, attributes map[string]interface{}) {
	a.log(ctx, "error", message, attributes)
}

// log is the internal logging method that routes to SDK or fallback.
func (a *TFOAdapter) log(ctx context.Context, severity string, message string, attributes map[string]interface{}) {
	if a.client != nil && a.isInitialized() {
		if err := a.client.Log(ctx, severity, message, attributes); err != nil {
			// Fallback to local on error
			a.logLocal(severityToZerolog(severity), message, attributes)
		}
		return
	}

	// Use local fallback
	a.logLocal(severityToZerolog(severity), message, attributes)
}

// logLocal logs to the local zerolog logger.
func (a *TFOAdapter) logLocal(level zerolog.Level, message string, attributes map[string]interface{}) {
	if a.fallbackLogger == nil {
		return
	}

	var event *zerolog.Event
	switch level {
	case zerolog.DebugLevel:
		event = a.fallbackLogger.Debug()
	case zerolog.InfoLevel:
		event = a.fallbackLogger.Info()
	case zerolog.WarnLevel:
		event = a.fallbackLogger.Warn()
	case zerolog.ErrorLevel:
		event = a.fallbackLogger.Error()
	default:
		event = a.fallbackLogger.Info()
	}

	for k, v := range attributes {
		event = event.Interface(k, v)
	}
	event.Msg(message)
}

// ===== METRICS API =====

// RecordMetric records a generic metric.
func (a *TFOAdapter) RecordMetric(ctx context.Context, name string, value float64, unit string, attributes map[string]interface{}) error {
	if a.client != nil && a.isInitialized() {
		return a.client.RecordMetric(ctx, name, value, unit, attributes)
	}
	// Log locally as fallback
	a.logLocal(zerolog.DebugLevel, "metric", map[string]interface{}{
		"name":  name,
		"value": value,
		"unit":  unit,
	})
	return nil
}

// IncrementCounter increments a counter metric.
func (a *TFOAdapter) IncrementCounter(ctx context.Context, name string, value int64, attributes map[string]interface{}) error {
	if a.client != nil && a.isInitialized() {
		return a.client.IncrementCounter(ctx, name, value, attributes)
	}
	return nil
}

// RecordGauge records a gauge metric.
func (a *TFOAdapter) RecordGauge(ctx context.Context, name string, value float64, attributes map[string]interface{}) error {
	if a.client != nil && a.isInitialized() {
		return a.client.RecordGauge(ctx, name, value, attributes)
	}
	return nil
}

// RecordHistogram records a histogram measurement.
func (a *TFOAdapter) RecordHistogram(ctx context.Context, name string, value float64, unit string, attributes map[string]interface{}) error {
	if a.client != nil && a.isInitialized() {
		return a.client.RecordHistogram(ctx, name, value, unit, attributes)
	}
	return nil
}

// ===== TRACING API =====

// StartSpan starts a new trace span.
func (a *TFOAdapter) StartSpan(ctx context.Context, name string, kind string, attributes map[string]interface{}) (string, error) {
	if a.client != nil && a.isInitialized() {
		return a.client.StartSpan(ctx, name, kind, attributes)
	}
	// Return empty span ID for fallback
	return "", nil
}

// EndSpan ends an active span.
func (a *TFOAdapter) EndSpan(ctx context.Context, spanID string, err error) error {
	if a.client != nil && a.isInitialized() && spanID != "" {
		return a.client.EndSpan(ctx, spanID, err)
	}
	return nil
}

// AddSpanEvent adds an event to an active span.
func (a *TFOAdapter) AddSpanEvent(ctx context.Context, spanID string, name string, attributes map[string]interface{}) error {
	if a.client != nil && a.isInitialized() && spanID != "" {
		return a.client.AddSpanEvent(ctx, spanID, name, attributes)
	}
	return nil
}

// ===== MCP-SPECIFIC METHODS =====

// LogMCPRequest logs an MCP request.
func (a *TFOAdapter) LogMCPRequest(ctx context.Context, requestID, method, sessionID string) {
	a.Info(ctx, "MCP request received", map[string]interface{}{
		"request_id": requestID,
		"method":     method,
		"session_id": sessionID,
		"type":       "mcp.request",
	})

	_ = a.IncrementCounter(ctx, "mcp.requests.total", 1, map[string]interface{}{
		"method": method,
	})
}

// LogMCPResponse logs an MCP response.
func (a *TFOAdapter) LogMCPResponse(ctx context.Context, requestID, method, sessionID string, duration time.Duration, err error) {
	attrs := map[string]interface{}{
		"request_id":  requestID,
		"method":      method,
		"session_id":  sessionID,
		"duration_ms": duration.Milliseconds(),
		"type":        "mcp.response",
	}

	if err != nil {
		attrs["error"] = err.Error()
		attrs["success"] = false
		a.Error(ctx, "MCP request failed", attrs)

		_ = a.IncrementCounter(ctx, "mcp.requests.errors", 1, map[string]interface{}{
			"method": method,
		})
	} else {
		attrs["success"] = true
		a.Info(ctx, "MCP request completed", attrs)
	}

	_ = a.RecordHistogram(ctx, "mcp.request.duration", float64(duration.Milliseconds()), "ms", map[string]interface{}{
		"method": method,
	})
}

// LogToolCall logs a tool execution.
func (a *TFOAdapter) LogToolCall(ctx context.Context, toolName string, sessionID string, duration time.Duration, err error) {
	attrs := map[string]interface{}{
		"tool_name":   toolName,
		"session_id":  sessionID,
		"duration_ms": duration.Milliseconds(),
		"type":        "mcp.tool_call",
	}

	if err != nil {
		attrs["error"] = err.Error()
		attrs["success"] = false
		a.Error(ctx, "Tool execution failed", attrs)
	} else {
		attrs["success"] = true
		a.Info(ctx, "Tool executed successfully", attrs)
	}

	_ = a.RecordHistogram(ctx, "mcp.tool.duration", float64(duration.Milliseconds()), "ms", map[string]interface{}{
		"tool_name": toolName,
	})
}

// LogClaudeRequest logs a Claude API request.
func (a *TFOAdapter) LogClaudeRequest(ctx context.Context, model string, inputTokens, outputTokens int, duration time.Duration, err error) {
	attrs := map[string]interface{}{
		"model":         model,
		"input_tokens":  inputTokens,
		"output_tokens": outputTokens,
		"total_tokens":  inputTokens + outputTokens,
		"duration_ms":   duration.Milliseconds(),
		"type":          "claude.request",
	}

	if err != nil {
		attrs["error"] = err.Error()
		attrs["success"] = false
		a.Error(ctx, "Claude API request failed", attrs)
	} else {
		attrs["success"] = true
		a.Info(ctx, "Claude API request completed", attrs)
	}

	_ = a.RecordHistogram(ctx, "claude.request.duration", float64(duration.Milliseconds()), "ms", map[string]interface{}{
		"model": model,
	})
	_ = a.RecordGauge(ctx, "claude.tokens.input", float64(inputTokens), map[string]interface{}{
		"model": model,
	})
	_ = a.RecordGauge(ctx, "claude.tokens.output", float64(outputTokens), map[string]interface{}{
		"model": model,
	})
}

// LogSessionEvent logs a session lifecycle event.
func (a *TFOAdapter) LogSessionEvent(ctx context.Context, sessionID, event string, details map[string]interface{}) {
	attrs := map[string]interface{}{
		"session_id": sessionID,
		"event":      event,
		"type":       "mcp.session",
	}
	for k, v := range details {
		attrs[k] = v
	}
	a.Info(ctx, "Session event", attrs)

	_ = a.IncrementCounter(ctx, "mcp.session.events", 1, map[string]interface{}{
		"event": event,
	})
}

// ===== TRACED OPERATIONS =====

// TracedOperation executes a function with automatic tracing.
func (a *TFOAdapter) TracedOperation(ctx context.Context, spanName string, kind string, fn func(context.Context) error) error {
	spanID, err := a.StartSpan(ctx, spanName, kind, nil)
	if err != nil {
		return fn(ctx)
	}

	err = fn(ctx)
	_ = a.EndSpan(ctx, spanID, err)
	return err
}

// TracedOperationWithResult executes a function with automatic tracing and returns a result.
func TracedOperationWithResult[T any](a *TFOAdapter, ctx context.Context, spanName string, kind string, fn func(context.Context) (T, error)) (T, error) {
	spanID, spanErr := a.StartSpan(ctx, spanName, kind, nil)
	if spanErr != nil {
		return fn(ctx)
	}

	result, err := fn(ctx)
	_ = a.EndSpan(ctx, spanID, err)
	return result, err
}

// ===== HELPER METHODS =====

func (a *TFOAdapter) isInitialized() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.initialized
}

// Client returns the underlying TFO SDK client (for advanced usage).
func (a *TFOAdapter) Client() *telemetryflow.Client {
	return a.client
}

// FallbackLogger returns the local fallback logger.
func (a *TFOAdapter) FallbackLogger() *Logger {
	return a.fallbackLogger
}

// IsSDKAvailable returns true if TFO SDK is available and initialized.
func (a *TFOAdapter) IsSDKAvailable() bool {
	return a.client != nil && a.isInitialized()
}

// severityToZerolog converts a severity string to zerolog.Level.
func severityToZerolog(severity string) zerolog.Level {
	switch severity {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

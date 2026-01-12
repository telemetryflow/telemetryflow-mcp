// Package telemetry provides a unified observability facade for the MCP server.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package telemetry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/telemetryflow/telemetryflow-go-sdk/pkg/telemetryflow"
)

// Observability provides a unified interface for all telemetry operations.
// It wraps the TFO Go SDK to provide logging, metrics, and tracing in a
// single, easy-to-use facade.
type Observability struct {
	client         *telemetryflow.Client
	tracer         *TFOTracer
	serviceName    string
	serviceVersion string
	environment    string
	initialized    bool
	mu             sync.RWMutex
}

// ObservabilityConfig configures the observability facade.
type ObservabilityConfig struct {
	// APIKeyID is the TelemetryFlow API key ID
	APIKeyID string
	// APIKeySecret is the TelemetryFlow API key secret
	APIKeySecret string
	// Endpoint is the TelemetryFlow collector endpoint
	Endpoint string
	// ServiceName identifies this service
	ServiceName string
	// ServiceVersion is the version of this service
	ServiceVersion string
	// ServiceNamespace groups related services
	ServiceNamespace string
	// Environment is the deployment environment
	Environment string
	// CollectorID identifies the collector instance
	CollectorID string
	// EnableMetrics enables metrics collection
	EnableMetrics bool
	// EnableLogs enables log collection
	EnableLogs bool
	// EnableTraces enables trace collection
	EnableTraces bool
	// UseGRPC uses gRPC protocol (default), otherwise HTTP
	UseGRPC bool
	// Insecure disables TLS
	Insecure bool
	// Timeout is the connection timeout
	Timeout time.Duration
}

// DefaultObservabilityConfig returns default configuration.
func DefaultObservabilityConfig() *ObservabilityConfig {
	return &ObservabilityConfig{
		Endpoint:         "api.telemetryflow.id:4317",
		ServiceName:      "telemetryflow-go-mcp",
		ServiceVersion:   "1.1.2",
		ServiceNamespace: "telemetryflow",
		Environment:      "production",
		EnableMetrics:    true,
		EnableLogs:       true,
		EnableTraces:     true,
		UseGRPC:          true,
		Insecure:         false,
		Timeout:          30 * time.Second,
	}
}

// NewObservability creates a new observability facade.
func NewObservability(cfg *ObservabilityConfig) (*Observability, error) {
	if cfg == nil {
		cfg = DefaultObservabilityConfig()
	}

	obs := &Observability{
		serviceName:    cfg.ServiceName,
		serviceVersion: cfg.ServiceVersion,
		environment:    cfg.Environment,
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
		WithTimeout(cfg.Timeout).
		WithCustomAttribute("mcp.protocol.version", "2024-11-05").
		WithCustomAttribute("mcp.server.type", "telemetryflow")

	if cfg.UseGRPC {
		builder = builder.WithGRPC()
	} else {
		builder = builder.WithHTTP()
	}

	if cfg.CollectorID != "" {
		builder = builder.WithCollectorID(cfg.CollectorID)
	}

	client, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create TFO SDK client: %w", err)
	}

	obs.client = client
	obs.tracer = NewTFOTracer(client, cfg.ServiceName, cfg.ServiceVersion)

	return obs, nil
}

// NewObservabilityFromEnv creates an observability facade using environment variables.
func NewObservabilityFromEnv() (*Observability, error) {
	client, err := telemetryflow.NewFromEnv()
	if err != nil {
		return nil, err
	}

	return &Observability{
		client:         client,
		tracer:         NewTFOTracer(client, "telemetryflow-go-mcp", "1.1.2"),
		serviceName:    "telemetryflow-go-mcp",
		serviceVersion: "1.1.2",
		environment:    "production",
	}, nil
}

// Initialize initializes the observability system.
func (o *Observability) Initialize(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.initialized {
		return nil
	}

	if o.client == nil {
		return fmt.Errorf("TFO SDK client not configured")
	}

	if err := o.client.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize TFO SDK: %w", err)
	}

	o.initialized = true

	// Log initialization
	_ = o.client.LogInfo(ctx, "Observability initialized", map[string]interface{}{
		"service":     o.serviceName,
		"version":     o.serviceVersion,
		"environment": o.environment,
	})

	return nil
}

// Shutdown gracefully shuts down the observability system.
func (o *Observability) Shutdown(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if !o.initialized {
		return nil
	}

	if o.client != nil {
		if err := o.client.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown TFO SDK: %w", err)
		}
	}

	o.initialized = false
	return nil
}

// Flush forces a flush of pending telemetry.
func (o *Observability) Flush(ctx context.Context) error {
	if o.client != nil && o.isInitialized() {
		return o.client.Flush(ctx)
	}
	return nil
}

// ===== LOGGING API =====

// Log emits a structured log entry.
func (o *Observability) Log(ctx context.Context, severity string, message string, attrs map[string]interface{}) error {
	if !o.isInitialized() {
		return nil
	}
	return o.client.Log(ctx, severity, message, attrs)
}

// Debug logs a debug message.
func (o *Observability) Debug(ctx context.Context, message string, attrs map[string]interface{}) {
	_ = o.Log(ctx, "debug", message, attrs)
}

// Info logs an info message.
func (o *Observability) Info(ctx context.Context, message string, attrs map[string]interface{}) {
	_ = o.Log(ctx, "info", message, attrs)
}

// Warn logs a warning message.
func (o *Observability) Warn(ctx context.Context, message string, attrs map[string]interface{}) {
	_ = o.Log(ctx, "warn", message, attrs)
}

// Error logs an error message.
func (o *Observability) Error(ctx context.Context, message string, attrs map[string]interface{}) {
	_ = o.Log(ctx, "error", message, attrs)
}

// ===== METRICS API =====

// RecordMetric records a generic metric.
func (o *Observability) RecordMetric(ctx context.Context, name string, value float64, unit string, attrs map[string]interface{}) error {
	if !o.isInitialized() {
		return nil
	}
	return o.client.RecordMetric(ctx, name, value, unit, attrs)
}

// IncrementCounter increments a counter metric.
func (o *Observability) IncrementCounter(ctx context.Context, name string, value int64, attrs map[string]interface{}) error {
	if !o.isInitialized() {
		return nil
	}
	return o.client.IncrementCounter(ctx, name, value, attrs)
}

// RecordGauge records a gauge metric.
func (o *Observability) RecordGauge(ctx context.Context, name string, value float64, attrs map[string]interface{}) error {
	if !o.isInitialized() {
		return nil
	}
	return o.client.RecordGauge(ctx, name, value, attrs)
}

// RecordHistogram records a histogram measurement.
func (o *Observability) RecordHistogram(ctx context.Context, name string, value float64, unit string, attrs map[string]interface{}) error {
	if !o.isInitialized() {
		return nil
	}
	return o.client.RecordHistogram(ctx, name, value, unit, attrs)
}

// ===== TRACING API =====

// Tracer returns the TFO tracer.
func (o *Observability) Tracer() *TFOTracer {
	return o.tracer
}

// StartSpan starts a new trace span.
func (o *Observability) StartSpan(ctx context.Context, name string, kind string, attrs map[string]interface{}) (context.Context, string, error) {
	if !o.isInitialized() {
		return ctx, "", nil
	}
	spanID, err := o.client.StartSpan(ctx, name, kind, attrs)
	return ctx, spanID, err
}

// EndSpan ends an active span.
func (o *Observability) EndSpan(ctx context.Context, spanID string, err error) error {
	if !o.isInitialized() || spanID == "" {
		return nil
	}
	return o.client.EndSpan(ctx, spanID, err)
}

// AddSpanEvent adds an event to an active span.
func (o *Observability) AddSpanEvent(ctx context.Context, spanID string, name string, attrs map[string]interface{}) error {
	if !o.isInitialized() || spanID == "" {
		return nil
	}
	return o.client.AddSpanEvent(ctx, spanID, name, attrs)
}

// ===== MCP-SPECIFIC OPERATIONS =====

// LogMCPRequest logs an MCP request with metrics.
func (o *Observability) LogMCPRequest(ctx context.Context, requestID, method, sessionID string) {
	o.Info(ctx, "MCP request received", map[string]interface{}{
		"request_id": requestID,
		"method":     method,
		"session_id": sessionID,
		"type":       "mcp.request",
	})

	_ = o.IncrementCounter(ctx, "mcp.requests.total", 1, map[string]interface{}{
		"method": method,
	})
}

// LogMCPResponse logs an MCP response with metrics.
func (o *Observability) LogMCPResponse(ctx context.Context, requestID, method, sessionID string, duration time.Duration, err error) {
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
		o.Error(ctx, "MCP request failed", attrs)

		_ = o.IncrementCounter(ctx, "mcp.requests.errors", 1, map[string]interface{}{
			"method": method,
		})
	} else {
		attrs["success"] = true
		o.Info(ctx, "MCP request completed", attrs)
	}

	_ = o.RecordHistogram(ctx, "mcp.request.duration", float64(duration.Milliseconds()), "ms", map[string]interface{}{
		"method": method,
	})
}

// LogToolExecution logs a tool execution with metrics.
func (o *Observability) LogToolExecution(ctx context.Context, toolName, sessionID string, duration time.Duration, err error) {
	attrs := map[string]interface{}{
		"tool_name":   toolName,
		"session_id":  sessionID,
		"duration_ms": duration.Milliseconds(),
		"type":        "mcp.tool",
	}

	if err != nil {
		attrs["error"] = err.Error()
		attrs["success"] = false
		o.Error(ctx, "Tool execution failed", attrs)

		_ = o.IncrementCounter(ctx, "mcp.tools.errors", 1, map[string]interface{}{
			"tool_name": toolName,
		})
	} else {
		attrs["success"] = true
		o.Info(ctx, "Tool executed successfully", attrs)
	}

	_ = o.IncrementCounter(ctx, "mcp.tools.calls", 1, map[string]interface{}{
		"tool_name": toolName,
	})

	_ = o.RecordHistogram(ctx, "mcp.tool.duration", float64(duration.Milliseconds()), "ms", map[string]interface{}{
		"tool_name": toolName,
	})
}

// LogClaudeRequest logs a Claude API request with metrics.
func (o *Observability) LogClaudeRequest(ctx context.Context, model string, inputTokens, outputTokens int, duration time.Duration, err error) {
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
		o.Error(ctx, "Claude API request failed", attrs)

		_ = o.IncrementCounter(ctx, "claude.requests.errors", 1, map[string]interface{}{
			"model": model,
		})
	} else {
		attrs["success"] = true
		o.Info(ctx, "Claude API request completed", attrs)
	}

	_ = o.IncrementCounter(ctx, "claude.requests.total", 1, map[string]interface{}{
		"model": model,
	})

	_ = o.RecordHistogram(ctx, "claude.request.duration", float64(duration.Milliseconds()), "ms", map[string]interface{}{
		"model": model,
	})

	_ = o.RecordGauge(ctx, "claude.tokens.input", float64(inputTokens), map[string]interface{}{
		"model": model,
	})

	_ = o.RecordGauge(ctx, "claude.tokens.output", float64(outputTokens), map[string]interface{}{
		"model": model,
	})
}

// LogSessionEvent logs a session lifecycle event.
func (o *Observability) LogSessionEvent(ctx context.Context, sessionID, event string, details map[string]interface{}) {
	attrs := map[string]interface{}{
		"session_id": sessionID,
		"event":      event,
		"type":       "mcp.session",
	}
	for k, v := range details {
		attrs[k] = v
	}
	o.Info(ctx, "Session event", attrs)

	_ = o.IncrementCounter(ctx, "mcp.session.events", 1, map[string]interface{}{
		"event": event,
	})
}

// ===== TRACED OPERATIONS =====

// TracedOperation executes a function with automatic tracing, logging, and metrics.
func (o *Observability) TracedOperation(ctx context.Context, spanName string, kind string, fn func(context.Context) error) error {
	start := time.Now()

	ctx, spanID, _ := o.StartSpan(ctx, spanName, kind, nil)

	err := fn(ctx)
	duration := time.Since(start)

	// End span
	_ = o.EndSpan(ctx, spanID, err)

	// Record metrics
	_ = o.RecordHistogram(ctx, "operation.duration", float64(duration.Milliseconds()), "ms", map[string]interface{}{
		"operation": spanName,
	})

	if err != nil {
		_ = o.IncrementCounter(ctx, "operation.errors", 1, map[string]interface{}{
			"operation": spanName,
		})
	}

	return err
}

// TracedOperationWithResult executes a function with automatic tracing and returns a result.
func TracedOperationWithResult[T any](o *Observability, ctx context.Context, spanName string, kind string, fn func(context.Context) (T, error)) (T, error) {
	start := time.Now()

	ctx, spanID, _ := o.StartSpan(ctx, spanName, kind, nil)

	result, err := fn(ctx)
	duration := time.Since(start)

	// End span
	_ = o.EndSpan(ctx, spanID, err)

	// Record metrics
	_ = o.RecordHistogram(ctx, "operation.duration", float64(duration.Milliseconds()), "ms", map[string]interface{}{
		"operation": spanName,
	})

	if err != nil {
		_ = o.IncrementCounter(ctx, "operation.errors", 1, map[string]interface{}{
			"operation": spanName,
		})
	}

	return result, err
}

// ===== HELPER METHODS =====

func (o *Observability) isInitialized() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.initialized && o.client != nil
}

// Client returns the underlying TFO SDK client.
func (o *Observability) Client() *telemetryflow.Client {
	return o.client
}

// IsAvailable returns true if the observability system is initialized.
func (o *Observability) IsAvailable() bool {
	return o.isInitialized()
}

// ServiceInfo returns the service information.
func (o *Observability) ServiceInfo() (name, version, env string) {
	return o.serviceName, o.serviceVersion, o.environment
}

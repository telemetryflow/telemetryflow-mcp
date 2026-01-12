// Package logging provides zerolog hooks for integration with telemetry.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package logging

import (
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// TraceHook adds trace context to log events.
type TraceHook struct{}

// Run implements zerolog.Hook.
func (h TraceHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}

	spanCtx := trace.SpanContextFromContext(ctx)
	if !spanCtx.IsValid() {
		return
	}

	e.Str("trace_id", spanCtx.TraceID().String())
	e.Str("span_id", spanCtx.SpanID().String())
	if spanCtx.IsSampled() {
		e.Bool("trace_sampled", true)
	}
}

// SpanEventHook sends log events to the current span as span events.
type SpanEventHook struct {
	// MinLevel is the minimum level to send to spans
	MinLevel zerolog.Level
}

// NewSpanEventHook creates a new SpanEventHook.
func NewSpanEventHook(minLevel zerolog.Level) *SpanEventHook {
	return &SpanEventHook{
		MinLevel: minLevel,
	}
}

// Run implements zerolog.Hook.
func (h *SpanEventHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level < h.MinLevel {
		return
	}

	ctx := e.GetCtx()
	if ctx == nil {
		return
	}

	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	// Add log as span event
	attrs := []attribute.KeyValue{
		attribute.String("log.level", level.String()),
		attribute.String("log.message", msg),
	}

	span.AddEvent("log", trace.WithAttributes(attrs...))

	// Set span status to error if level is error or above
	if level >= zerolog.ErrorLevel {
		span.SetStatus(codes.Error, msg)
	}
}

// MetricsHook records log metrics.
type MetricsHook struct {
	counter LogCounter
}

// LogCounter interface for counting log events.
type LogCounter interface {
	Inc(level string)
}

// NewMetricsHook creates a new MetricsHook.
func NewMetricsHook(counter LogCounter) *MetricsHook {
	return &MetricsHook{
		counter: counter,
	}
}

// Run implements zerolog.Hook.
func (h *MetricsHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if h.counter != nil {
		h.counter.Inc(level.String())
	}
}

// SamplingHook samples log events based on rate.
type SamplingHook struct {
	// Rate is the sampling rate (1.0 = 100%, 0.1 = 10%)
	Rate float64
	// MinLevel is the minimum level to always log (bypasses sampling)
	MinLevel zerolog.Level
	counter  uint64
}

// NewSamplingHook creates a new SamplingHook.
func NewSamplingHook(rate float64, minLevel zerolog.Level) *SamplingHook {
	return &SamplingHook{
		Rate:     rate,
		MinLevel: minLevel,
	}
}

// Run implements zerolog.Hook.
func (h *SamplingHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	// Always log errors and above
	if level >= h.MinLevel {
		return
	}

	// Simple counter-based sampling
	h.counter++
	sampleInterval := uint64(1.0 / h.Rate)
	if sampleInterval == 0 {
		sampleInterval = 1
	}

	if h.counter%sampleInterval != 0 {
		e.Discard()
	}
}

// ContextFieldsHook adds fields from context to log events.
type ContextFieldsHook struct {
	// Keys are the context keys to extract
	Keys []ContextKey
}

// ContextKey represents a context key and its log field name.
type ContextKey struct {
	Key      interface{}
	LogField string
}

// NewContextFieldsHook creates a new ContextFieldsHook.
func NewContextFieldsHook(keys []ContextKey) *ContextFieldsHook {
	return &ContextFieldsHook{
		Keys: keys,
	}
}

// Run implements zerolog.Hook.
func (h *ContextFieldsHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}

	for _, key := range h.Keys {
		if value := ctx.Value(key.Key); value != nil {
			switch v := value.(type) {
			case string:
				e.Str(key.LogField, v)
			case int:
				e.Int(key.LogField, v)
			case int64:
				e.Int64(key.LogField, v)
			case float64:
				e.Float64(key.LogField, v)
			case bool:
				e.Bool(key.LogField, v)
			default:
				e.Interface(key.LogField, v)
			}
		}
	}
}

// ErrorStackHook adds stack traces for error-level logs.
type ErrorStackHook struct{}

// Run implements zerolog.Hook.
func (h ErrorStackHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level >= zerolog.ErrorLevel {
		e.Stack()
	}
}

// SessionHook adds session information to log events.
type SessionHook struct {
	sessionIDKey interface{}
}

// NewSessionHook creates a new SessionHook.
func NewSessionHook(sessionIDKey interface{}) *SessionHook {
	return &SessionHook{
		sessionIDKey: sessionIDKey,
	}
}

// Run implements zerolog.Hook.
func (h *SessionHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	ctx := e.GetCtx()
	if ctx == nil {
		return
	}

	if sessionID := ctx.Value(h.sessionIDKey); sessionID != nil {
		if id, ok := sessionID.(string); ok && id != "" {
			e.Str("session_id", id)
		}
	}
}

// CompositeHook combines multiple hooks.
type CompositeHook struct {
	hooks []zerolog.Hook
}

// NewCompositeHook creates a new CompositeHook.
func NewCompositeHook(hooks ...zerolog.Hook) *CompositeHook {
	return &CompositeHook{
		hooks: hooks,
	}
}

// Run implements zerolog.Hook.
func (h *CompositeHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	for _, hook := range h.hooks {
		hook.Run(e, level, msg)
	}
}

// Add adds a hook to the composite.
func (h *CompositeHook) Add(hook zerolog.Hook) {
	h.hooks = append(h.hooks, hook)
}

// Package telemetry provides TFO SDK-based tracing for MCP operations.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package telemetry

import (
	"context"
	"sync"

	"github.com/telemetryflow/telemetryflow-go-sdk/pkg/telemetryflow"
)

// TFOTracer wraps TFO Go SDK tracing functionality for MCP operations.
// It provides a unified interface for distributed tracing that sends
// traces through the TelemetryFlow platform.
type TFOTracer struct {
	client         *telemetryflow.Client
	serviceName    string
	serviceVersion string
	spans          map[string]*spanInfo
	mu             sync.RWMutex
}

// spanInfo tracks active span information.
type spanInfo struct {
	name       string
	attributes map[string]interface{}
}

// NewTFOTracer creates a new TFO SDK-based tracer.
func NewTFOTracer(client *telemetryflow.Client, serviceName, serviceVersion string) *TFOTracer {
	return &TFOTracer{
		client:         client,
		serviceName:    serviceName,
		serviceVersion: serviceVersion,
		spans:          make(map[string]*spanInfo),
	}
}

// StartSpan starts a new span with the given name and options.
func (t *TFOTracer) StartSpan(ctx context.Context, spanName string, kind string, opts ...TraceOption) (context.Context, string, error) {
	options := &traceOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Build attributes from options
	attrs := make(map[string]interface{})

	if options.sessionID != "" {
		attrs[AttrSessionID] = options.sessionID
	}
	if options.conversationID != "" {
		attrs[AttrConversationID] = options.conversationID
	}
	if options.toolName != "" {
		attrs[AttrToolName] = options.toolName
	}
	if options.resourceURI != "" {
		attrs[AttrResourceURI] = options.resourceURI
	}
	if options.promptName != "" {
		attrs[AttrPromptName] = options.promptName
	}
	if options.mcpMethod != "" {
		attrs[AttrMCPMethod] = options.mcpMethod
	}
	if options.model != "" {
		attrs[AttrClaudeModel] = options.model
	}

	// Add custom attributes
	for _, attr := range options.attributes {
		attrs[string(attr.Key)] = attr.Value.AsInterface()
	}

	// Start span using TFO SDK
	if t.client == nil || !t.client.IsInitialized() {
		return ctx, "", nil
	}

	spanID, err := t.client.StartSpan(ctx, spanName, kind, attrs)
	if err != nil {
		return ctx, "", err
	}

	// Track span info
	t.mu.Lock()
	t.spans[spanID] = &spanInfo{
		name:       spanName,
		attributes: attrs,
	}
	t.mu.Unlock()

	return ctx, spanID, nil
}

// EndSpan ends an active span.
func (t *TFOTracer) EndSpan(ctx context.Context, spanID string, err error) error {
	if t.client == nil || !t.client.IsInitialized() || spanID == "" {
		return nil
	}

	// Remove span tracking
	t.mu.Lock()
	delete(t.spans, spanID)
	t.mu.Unlock()

	return t.client.EndSpan(ctx, spanID, err)
}

// AddSpanEvent adds an event to an active span.
func (t *TFOTracer) AddSpanEvent(ctx context.Context, spanID string, name string, attrs map[string]interface{}) error {
	if t.client == nil || !t.client.IsInitialized() || spanID == "" {
		return nil
	}

	return t.client.AddSpanEvent(ctx, spanID, name, attrs)
}

// MCP Operation Spans

// StartMCPRequestSpan starts a span for an MCP request.
func (t *TFOTracer) StartMCPRequestSpan(ctx context.Context, method string, sessionID string) (context.Context, string, error) {
	return t.StartSpan(ctx, "mcp.request", "server",
		WithMCPMethod(method),
		WithSessionID(sessionID),
	)
}

// StartInitializeSpan starts a span for session initialization.
func (t *TFOTracer) StartInitializeSpan(ctx context.Context) (context.Context, string, error) {
	return t.StartSpan(ctx, "mcp.initialize", "server",
		WithMCPMethod("initialize"),
	)
}

// StartToolCallSpan starts a span for tool execution.
func (t *TFOTracer) StartToolCallSpan(ctx context.Context, toolName, sessionID string) (context.Context, string, error) {
	return t.StartSpan(ctx, "mcp.tool.call", "internal",
		WithMCPMethod("tools/call"),
		WithToolName(toolName),
		WithSessionID(sessionID),
	)
}

// StartToolListSpan starts a span for tool listing.
func (t *TFOTracer) StartToolListSpan(ctx context.Context, sessionID string) (context.Context, string, error) {
	return t.StartSpan(ctx, "mcp.tool.list", "server",
		WithMCPMethod("tools/list"),
		WithSessionID(sessionID),
	)
}

// StartResourceReadSpan starts a span for resource reading.
func (t *TFOTracer) StartResourceReadSpan(ctx context.Context, uri, sessionID string) (context.Context, string, error) {
	return t.StartSpan(ctx, "mcp.resource.read", "internal",
		WithMCPMethod("resources/read"),
		WithResourceURI(uri),
		WithSessionID(sessionID),
	)
}

// StartResourceListSpan starts a span for resource listing.
func (t *TFOTracer) StartResourceListSpan(ctx context.Context, sessionID string) (context.Context, string, error) {
	return t.StartSpan(ctx, "mcp.resource.list", "server",
		WithMCPMethod("resources/list"),
		WithSessionID(sessionID),
	)
}

// StartPromptGetSpan starts a span for prompt retrieval.
func (t *TFOTracer) StartPromptGetSpan(ctx context.Context, promptName, sessionID string) (context.Context, string, error) {
	return t.StartSpan(ctx, "mcp.prompt.get", "internal",
		WithMCPMethod("prompts/get"),
		WithPromptName(promptName),
		WithSessionID(sessionID),
	)
}

// StartPromptListSpan starts a span for prompt listing.
func (t *TFOTracer) StartPromptListSpan(ctx context.Context, sessionID string) (context.Context, string, error) {
	return t.StartSpan(ctx, "mcp.prompt.list", "server",
		WithMCPMethod("prompts/list"),
		WithSessionID(sessionID),
	)
}

// Claude API Spans

// StartClaudeRequestSpan starts a span for Claude API requests.
func (t *TFOTracer) StartClaudeRequestSpan(ctx context.Context, model, sessionID, conversationID string) (context.Context, string, error) {
	return t.StartSpan(ctx, "claude.request", "client",
		WithModel(model),
		WithSessionID(sessionID),
		WithConversationID(conversationID),
	)
}

// StartClaudeStreamSpan starts a span for Claude streaming requests.
func (t *TFOTracer) StartClaudeStreamSpan(ctx context.Context, model, sessionID, conversationID string) (context.Context, string, error) {
	return t.StartSpan(ctx, "claude.stream", "client",
		WithModel(model),
		WithSessionID(sessionID),
		WithConversationID(conversationID),
	)
}

// StartTokenCountSpan starts a span for token counting.
func (t *TFOTracer) StartTokenCountSpan(ctx context.Context, model string) (context.Context, string, error) {
	return t.StartSpan(ctx, "claude.token_count", "internal",
		WithModel(model),
	)
}

// Session Lifecycle Spans

// StartSessionCreateSpan starts a span for session creation.
func (t *TFOTracer) StartSessionCreateSpan(ctx context.Context) (context.Context, string, error) {
	return t.StartSpan(ctx, "session.create", "internal")
}

// StartSessionCloseSpan starts a span for session closure.
func (t *TFOTracer) StartSessionCloseSpan(ctx context.Context, sessionID string) (context.Context, string, error) {
	return t.StartSpan(ctx, "session.close", "internal",
		WithSessionID(sessionID),
	)
}

// Conversation Spans

// StartConversationCreateSpan starts a span for conversation creation.
func (t *TFOTracer) StartConversationCreateSpan(ctx context.Context, sessionID, model string) (context.Context, string, error) {
	return t.StartSpan(ctx, "conversation.create", "internal",
		WithSessionID(sessionID),
		WithModel(model),
	)
}

// StartConversationCloseSpan starts a span for conversation closure.
func (t *TFOTracer) StartConversationCloseSpan(ctx context.Context, sessionID, conversationID string) (context.Context, string, error) {
	return t.StartSpan(ctx, "conversation.close", "internal",
		WithSessionID(sessionID),
		WithConversationID(conversationID),
	)
}

// StartMessageAddSpan starts a span for adding a message.
func (t *TFOTracer) StartMessageAddSpan(ctx context.Context, sessionID, conversationID, role string) (context.Context, string, error) {
	return t.StartSpan(ctx, "conversation.message.add", "internal",
		WithSessionID(sessionID),
		WithConversationID(conversationID),
		WithAttribute("message.role", role),
	)
}

// Traced Operations

// TFOTracedOperation wraps an operation with TFO SDK tracing.
func TFOTracedOperation[T any](ctx context.Context, tracer *TFOTracer, spanName string, kind string, opts []TraceOption, fn func(context.Context) (T, error)) (T, error) {
	ctx, spanID, err := tracer.StartSpan(ctx, spanName, kind, opts...)
	if err != nil {
		return fn(ctx)
	}

	result, fnErr := fn(ctx)
	_ = tracer.EndSpan(ctx, spanID, fnErr)
	return result, fnErr
}

// TFOTracedVoidOperation wraps a void operation with TFO SDK tracing.
func TFOTracedVoidOperation(ctx context.Context, tracer *TFOTracer, spanName string, kind string, opts []TraceOption, fn func(context.Context) error) error {
	ctx, spanID, err := tracer.StartSpan(ctx, spanName, kind, opts...)
	if err != nil {
		return fn(ctx)
	}

	fnErr := fn(ctx)
	_ = tracer.EndSpan(ctx, spanID, fnErr)
	return fnErr
}

// AddTokenInfo adds token information to a span.
func (t *TFOTracer) AddTokenInfo(ctx context.Context, spanID string, inputTokens, outputTokens int) error {
	return t.AddSpanEvent(ctx, spanID, "token_count", map[string]interface{}{
		AttrTokensInput:  inputTokens,
		AttrTokensOutput: outputTokens,
		"tokens.total":   inputTokens + outputTokens,
	})
}

// AddToolEvent adds a tool event to a span.
func (t *TFOTracer) AddToolEvent(ctx context.Context, spanID string, eventType string, toolName string, attrs map[string]interface{}) error {
	eventAttrs := map[string]interface{}{
		"tool.name": toolName,
	}
	for k, v := range attrs {
		eventAttrs[k] = v
	}
	return t.AddSpanEvent(ctx, spanID, eventType, eventAttrs)
}

// IsAvailable returns true if the TFO SDK client is available and initialized.
func (t *TFOTracer) IsAvailable() bool {
	return t.client != nil && t.client.IsInitialized()
}

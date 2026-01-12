// Package queue provides predefined task types for the MCP server.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Predefined task types
const (
	// TaskTypeClaudeRequest handles Claude API requests asynchronously.
	TaskTypeClaudeRequest = "claude.request"

	// TaskTypeToolExecution handles tool execution asynchronously.
	TaskTypeToolExecution = "tool.execute"

	// TaskTypeTelemetryExport handles telemetry data export.
	TaskTypeTelemetryExport = "telemetry.export"

	// TaskTypeSessionCleanup handles session cleanup tasks.
	TaskTypeSessionCleanup = "session.cleanup"

	// TaskTypeConversationArchive handles conversation archival.
	TaskTypeConversationArchive = "conversation.archive"

	// TaskTypeCacheInvalidation handles cache invalidation.
	TaskTypeCacheInvalidation = "cache.invalidate"

	// TaskTypeWebhookDelivery handles webhook delivery.
	TaskTypeWebhookDelivery = "webhook.deliver"

	// TaskTypeEmailNotification handles email notifications.
	TaskTypeEmailNotification = "email.notify"
)

// Event types
const (
	EventTypeSessionCreated      = "session.created"
	EventTypeSessionClosed       = "session.closed"
	EventTypeConversationCreated = "conversation.created"
	EventTypeConversationClosed  = "conversation.closed"
	EventTypeMessageSent         = "message.sent"
	EventTypeToolExecuted        = "tool.executed"
	EventTypeResourceRead        = "resource.read"
	EventTypePromptGenerated     = "prompt.generated"
	EventTypeAPIRequestCompleted = "api.request.completed"
	EventTypeAPIRequestFailed    = "api.request.failed"
)

// Telemetry types
const (
	TelemetryTypeSpan   = "span"
	TelemetryTypeMetric = "metric"
	TelemetryTypeLog    = "log"
	TelemetryTypeTrace  = "trace"
)

// ClaudeRequestPayload represents a Claude API request task payload.
type ClaudeRequestPayload struct {
	SessionID      string                 `json:"session_id"`
	ConversationID string                 `json:"conversation_id"`
	Model          string                 `json:"model"`
	Messages       []MessagePayload       `json:"messages"`
	SystemPrompt   string                 `json:"system_prompt,omitempty"`
	MaxTokens      int                    `json:"max_tokens,omitempty"`
	Temperature    float64                `json:"temperature,omitempty"`
	Tools          []ToolPayload          `json:"tools,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// MessagePayload represents a message in the Claude request.
type MessagePayload struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a content block in a message.
type ContentBlock struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

// ToolPayload represents a tool in the Claude request.
type ToolPayload struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"`
}

// ToolExecutionPayload represents a tool execution task payload.
type ToolExecutionPayload struct {
	SessionID  string                 `json:"session_id"`
	ToolName   string                 `json:"tool_name"`
	Input      map[string]interface{} `json:"input"`
	CallbackID string                 `json:"callback_id,omitempty"`
	Timeout    time.Duration          `json:"timeout,omitempty"`
}

// TelemetryExportPayload represents a telemetry export task payload.
type TelemetryExportPayload struct {
	ServiceName string                 `json:"service_name"`
	SpanID      string                 `json:"span_id,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	Metrics     map[string]interface{} `json:"metrics,omitempty"`
	Logs        []LogEntry             `json:"logs,omitempty"`
	Destination string                 `json:"destination"`
}

// LogEntry represents a log entry for export.
type LogEntry struct {
	Timestamp  time.Time              `json:"timestamp"`
	Level      string                 `json:"level"`
	Message    string                 `json:"message"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// SessionCleanupPayload represents a session cleanup task payload.
type SessionCleanupPayload struct {
	SessionID  string    `json:"session_id"`
	ClosedAt   time.Time `json:"closed_at"`
	CleanupAll bool      `json:"cleanup_all"`
}

// ConversationArchivePayload represents a conversation archive task payload.
type ConversationArchivePayload struct {
	ConversationID string    `json:"conversation_id"`
	SessionID      string    `json:"session_id"`
	ClosedAt       time.Time `json:"closed_at"`
	Destination    string    `json:"destination"`
}

// CacheInvalidationPayload represents a cache invalidation task payload.
type CacheInvalidationPayload struct {
	Pattern    string   `json:"pattern,omitempty"`
	Keys       []string `json:"keys,omitempty"`
	InvalidAll bool     `json:"invalidate_all"`
}

// WebhookDeliveryPayload represents a webhook delivery task payload.
type WebhookDeliveryPayload struct {
	URL         string                 `json:"url"`
	Method      string                 `json:"method"`
	Headers     map[string]string      `json:"headers,omitempty"`
	Body        map[string]interface{} `json:"body"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	CallbackURL string                 `json:"callback_url,omitempty"`
}

// EmailNotificationPayload represents an email notification task payload.
type EmailNotificationPayload struct {
	To          []string               `json:"to"`
	CC          []string               `json:"cc,omitempty"`
	BCC         []string               `json:"bcc,omitempty"`
	Subject     string                 `json:"subject"`
	Body        string                 `json:"body"`
	HTML        bool                   `json:"html"`
	Attachments []AttachmentPayload    `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AttachmentPayload represents an email attachment.
type AttachmentPayload struct {
	Name     string `json:"name"`
	MimeType string `json:"mime_type"`
	Data     string `json:"data"` // Base64 encoded
}

// TaskBuilder helps build tasks with proper payloads.
type TaskBuilder struct {
	taskType string
	payload  map[string]interface{}
	subject  string
	priority TaskPriority
	maxRetry int
	timeout  time.Duration
	deadline time.Time
	metadata map[string]string
}

// NewTaskBuilder creates a new task builder.
func NewTaskBuilder(taskType string) *TaskBuilder {
	return &TaskBuilder{
		taskType: taskType,
		payload:  make(map[string]interface{}),
		subject:  fmt.Sprintf("%s.%s", SubjectTaskPrefix, taskType),
		priority: PriorityDefault,
		maxRetry: 3,
		timeout:  30 * time.Second,
		metadata: make(map[string]string),
	}
}

// WithPayload sets the task payload.
func (b *TaskBuilder) WithPayload(payload interface{}) *TaskBuilder {
	data, err := json.Marshal(payload)
	if err != nil {
		return b
	}
	_ = json.Unmarshal(data, &b.payload)
	return b
}

// WithSubject sets the NATS subject.
func (b *TaskBuilder) WithSubject(subject string) *TaskBuilder {
	b.subject = subject
	return b
}

// WithPriority sets the task priority.
func (b *TaskBuilder) WithPriority(priority TaskPriority) *TaskBuilder {
	b.priority = priority
	return b
}

// WithMaxRetry sets the maximum retry count.
func (b *TaskBuilder) WithMaxRetry(maxRetry int) *TaskBuilder {
	b.maxRetry = maxRetry
	return b
}

// WithTimeout sets the task timeout.
func (b *TaskBuilder) WithTimeout(timeout time.Duration) *TaskBuilder {
	b.timeout = timeout
	return b
}

// WithDeadline sets the task deadline.
func (b *TaskBuilder) WithDeadline(deadline time.Time) *TaskBuilder {
	b.deadline = deadline
	return b
}

// WithMetadata adds metadata to the task.
func (b *TaskBuilder) WithMetadata(key, value string) *TaskBuilder {
	b.metadata[key] = value
	return b
}

// Build builds the task.
func (b *TaskBuilder) Build() *Task {
	return &Task{
		ID:        generateTaskID(),
		Type:      b.taskType,
		Payload:   b.payload,
		Subject:   b.subject,
		Priority:  b.priority,
		MaxRetry:  b.maxRetry,
		Timeout:   b.timeout,
		Deadline:  b.deadline,
		CreatedAt: time.Now(),
		Metadata:  b.metadata,
	}
}

// Publish builds and publishes the task.
func (b *TaskBuilder) Publish(ctx context.Context, q *NATSQueue) (string, error) {
	task := b.Build()
	return q.Publish(ctx, task)
}

// RegisterDefaultHandlers registers default task handlers.
func RegisterDefaultHandlers(q *NATSQueue) {
	// Claude request handler (stub - implement with actual Claude service)
	q.RegisterHandler(TaskTypeClaudeRequest, func(ctx context.Context, task *Task) error {
		var payload ClaudeRequestPayload
		data, _ := json.Marshal(task.Payload)
		if err := json.Unmarshal(data, &payload); err != nil {
			return err
		}
		// TODO: Implement Claude API call
		fmt.Printf("Processing Claude request for session %s\n", payload.SessionID)
		return nil
	})

	// Tool execution handler (stub)
	q.RegisterHandler(TaskTypeToolExecution, func(ctx context.Context, task *Task) error {
		var payload ToolExecutionPayload
		data, _ := json.Marshal(task.Payload)
		if err := json.Unmarshal(data, &payload); err != nil {
			return err
		}
		// TODO: Implement tool execution
		fmt.Printf("Executing tool %s for session %s\n", payload.ToolName, payload.SessionID)
		return nil
	})

	// Telemetry export handler (stub)
	q.RegisterHandler(TaskTypeTelemetryExport, func(ctx context.Context, task *Task) error {
		var payload TelemetryExportPayload
		data, _ := json.Marshal(task.Payload)
		if err := json.Unmarshal(data, &payload); err != nil {
			return err
		}
		// TODO: Implement telemetry export
		fmt.Printf("Exporting telemetry for service %s to %s\n", payload.ServiceName, payload.Destination)
		return nil
	})

	// Session cleanup handler (stub)
	q.RegisterHandler(TaskTypeSessionCleanup, func(ctx context.Context, task *Task) error {
		var payload SessionCleanupPayload
		data, _ := json.Marshal(task.Payload)
		if err := json.Unmarshal(data, &payload); err != nil {
			return err
		}
		// TODO: Implement session cleanup
		fmt.Printf("Cleaning up session %s\n", payload.SessionID)
		return nil
	})

	// Cache invalidation handler (stub)
	q.RegisterHandler(TaskTypeCacheInvalidation, func(ctx context.Context, task *Task) error {
		var payload CacheInvalidationPayload
		data, _ := json.Marshal(task.Payload)
		if err := json.Unmarshal(data, &payload); err != nil {
			return err
		}
		// TODO: Implement cache invalidation
		fmt.Printf("Invalidating cache pattern: %s\n", payload.Pattern)
		return nil
	})

	// Webhook delivery handler (stub)
	q.RegisterHandler(TaskTypeWebhookDelivery, func(ctx context.Context, task *Task) error {
		var payload WebhookDeliveryPayload
		data, _ := json.Marshal(task.Payload)
		if err := json.Unmarshal(data, &payload); err != nil {
			return err
		}
		// TODO: Implement webhook delivery
		fmt.Printf("Delivering webhook to %s\n", payload.URL)
		return nil
	})
}

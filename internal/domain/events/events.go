// Package events contains domain events for the TelemetryFlow GO MCP service
package events

import (
	"time"

	"github.com/google/uuid"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// DomainEvent is the interface for all domain events
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	AggregateType() string
	OccurredAt() time.Time
	Payload() map[string]interface{}
}

// BaseEvent contains common event fields
type BaseEvent struct {
	eventID       string
	eventType     string
	aggregateID   string
	aggregateType string
	occurredAt    time.Time
	payload       map[string]interface{}
}

// EventID returns the event ID
func (e *BaseEvent) EventID() string {
	return e.eventID
}

// EventType returns the event type
func (e *BaseEvent) EventType() string {
	return e.eventType
}

// AggregateID returns the aggregate ID
func (e *BaseEvent) AggregateID() string {
	return e.aggregateID
}

// AggregateType returns the aggregate type
func (e *BaseEvent) AggregateType() string {
	return e.aggregateType
}

// OccurredAt returns when the event occurred
func (e *BaseEvent) OccurredAt() time.Time {
	return e.occurredAt
}

// Payload returns the event payload
func (e *BaseEvent) Payload() map[string]interface{} {
	return e.payload
}

// newBaseEvent creates a new base event
func newBaseEvent(eventType, aggregateID, aggregateType string, payload map[string]interface{}) BaseEvent {
	return BaseEvent{
		eventID:       uuid.New().String(),
		eventType:     eventType,
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
		occurredAt:    time.Now().UTC(),
		payload:       payload,
	}
}

// Session Events

// SessionCreatedEvent is emitted when a session is created
type SessionCreatedEvent struct {
	BaseEvent
}

// NewSessionCreatedEvent creates a new SessionCreatedEvent
func NewSessionCreatedEvent(sessionID vo.SessionID) *SessionCreatedEvent {
	return &SessionCreatedEvent{
		BaseEvent: newBaseEvent(
			"session.created",
			sessionID.String(),
			"Session",
			map[string]interface{}{
				"sessionId": sessionID.String(),
			},
		),
	}
}

// SessionInitializedEvent is emitted when a session is initialized
type SessionInitializedEvent struct {
	BaseEvent
}

// NewSessionInitializedEvent creates a new SessionInitializedEvent
func NewSessionInitializedEvent(sessionID vo.SessionID, clientName, clientVersion string) *SessionInitializedEvent {
	return &SessionInitializedEvent{
		BaseEvent: newBaseEvent(
			"session.initialized",
			sessionID.String(),
			"Session",
			map[string]interface{}{
				"sessionId":     sessionID.String(),
				"clientName":    clientName,
				"clientVersion": clientVersion,
			},
		),
	}
}

// SessionClosedEvent is emitted when a session is closed
type SessionClosedEvent struct {
	BaseEvent
}

// NewSessionClosedEvent creates a new SessionClosedEvent
func NewSessionClosedEvent(sessionID vo.SessionID) *SessionClosedEvent {
	return &SessionClosedEvent{
		BaseEvent: newBaseEvent(
			"session.closed",
			sessionID.String(),
			"Session",
			map[string]interface{}{
				"sessionId": sessionID.String(),
			},
		),
	}
}

// Conversation Events

// ConversationCreatedEvent is emitted when a conversation is created
type ConversationCreatedEvent struct {
	BaseEvent
}

// NewConversationCreatedEvent creates a new ConversationCreatedEvent
func NewConversationCreatedEvent(conversationID vo.ConversationID, sessionID vo.SessionID, model vo.Model) *ConversationCreatedEvent {
	return &ConversationCreatedEvent{
		BaseEvent: newBaseEvent(
			"conversation.created",
			conversationID.String(),
			"Conversation",
			map[string]interface{}{
				"conversationId": conversationID.String(),
				"sessionId":      sessionID.String(),
				"model":          model.String(),
			},
		),
	}
}

// ConversationClosedEvent is emitted when a conversation is closed
type ConversationClosedEvent struct {
	BaseEvent
}

// NewConversationClosedEvent creates a new ConversationClosedEvent
func NewConversationClosedEvent(conversationID vo.ConversationID) *ConversationClosedEvent {
	return &ConversationClosedEvent{
		BaseEvent: newBaseEvent(
			"conversation.closed",
			conversationID.String(),
			"Conversation",
			map[string]interface{}{
				"conversationId": conversationID.String(),
			},
		),
	}
}

// Message Events

// MessageAddedEvent is emitted when a message is added to a conversation
type MessageAddedEvent struct {
	BaseEvent
}

// NewMessageAddedEvent creates a new MessageAddedEvent
func NewMessageAddedEvent(conversationID vo.ConversationID, messageID vo.MessageID, role vo.Role) *MessageAddedEvent {
	return &MessageAddedEvent{
		BaseEvent: newBaseEvent(
			"message.added",
			conversationID.String(),
			"Conversation",
			map[string]interface{}{
				"conversationId": conversationID.String(),
				"messageId":      messageID.String(),
				"role":           role.String(),
			},
		),
	}
}

// Tool Events

// ToolRegisteredEvent is emitted when a tool is registered
type ToolRegisteredEvent struct {
	BaseEvent
}

// NewToolRegisteredEvent creates a new ToolRegisteredEvent
func NewToolRegisteredEvent(sessionID vo.SessionID, toolName string) *ToolRegisteredEvent {
	return &ToolRegisteredEvent{
		BaseEvent: newBaseEvent(
			"tool.registered",
			sessionID.String(),
			"Session",
			map[string]interface{}{
				"sessionId": sessionID.String(),
				"toolName":  toolName,
			},
		),
	}
}

// ToolExecutedEvent is emitted when a tool is executed
type ToolExecutedEvent struct {
	BaseEvent
}

// NewToolExecutedEvent creates a new ToolExecutedEvent
func NewToolExecutedEvent(sessionID vo.SessionID, toolName string, success bool, duration time.Duration) *ToolExecutedEvent {
	return &ToolExecutedEvent{
		BaseEvent: newBaseEvent(
			"tool.executed",
			sessionID.String(),
			"Session",
			map[string]interface{}{
				"sessionId":  sessionID.String(),
				"toolName":   toolName,
				"success":    success,
				"durationMs": duration.Milliseconds(),
			},
		),
	}
}

// Resource Events

// ResourceRegisteredEvent is emitted when a resource is registered
type ResourceRegisteredEvent struct {
	BaseEvent
}

// NewResourceRegisteredEvent creates a new ResourceRegisteredEvent
func NewResourceRegisteredEvent(sessionID vo.SessionID, resourceURI string) *ResourceRegisteredEvent {
	return &ResourceRegisteredEvent{
		BaseEvent: newBaseEvent(
			"resource.registered",
			sessionID.String(),
			"Session",
			map[string]interface{}{
				"sessionId":   sessionID.String(),
				"resourceUri": resourceURI,
			},
		),
	}
}

// ResourceReadEvent is emitted when a resource is read
type ResourceReadEvent struct {
	BaseEvent
}

// NewResourceReadEvent creates a new ResourceReadEvent
func NewResourceReadEvent(sessionID vo.SessionID, resourceURI string, success bool) *ResourceReadEvent {
	return &ResourceReadEvent{
		BaseEvent: newBaseEvent(
			"resource.read",
			sessionID.String(),
			"Session",
			map[string]interface{}{
				"sessionId":   sessionID.String(),
				"resourceUri": resourceURI,
				"success":     success,
			},
		),
	}
}

// Prompt Events

// PromptRegisteredEvent is emitted when a prompt is registered
type PromptRegisteredEvent struct {
	BaseEvent
}

// NewPromptRegisteredEvent creates a new PromptRegisteredEvent
func NewPromptRegisteredEvent(sessionID vo.SessionID, promptName string) *PromptRegisteredEvent {
	return &PromptRegisteredEvent{
		BaseEvent: newBaseEvent(
			"prompt.registered",
			sessionID.String(),
			"Session",
			map[string]interface{}{
				"sessionId":  sessionID.String(),
				"promptName": promptName,
			},
		),
	}
}

// PromptExecutedEvent is emitted when a prompt is executed
type PromptExecutedEvent struct {
	BaseEvent
}

// NewPromptExecutedEvent creates a new PromptExecutedEvent
func NewPromptExecutedEvent(sessionID vo.SessionID, promptName string, success bool) *PromptExecutedEvent {
	return &PromptExecutedEvent{
		BaseEvent: newBaseEvent(
			"prompt.executed",
			sessionID.String(),
			"Session",
			map[string]interface{}{
				"sessionId":  sessionID.String(),
				"promptName": promptName,
				"success":    success,
			},
		),
	}
}

// API Events

// APIRequestEvent is emitted when an API request is made
type APIRequestEvent struct {
	BaseEvent
}

// NewAPIRequestEvent creates a new APIRequestEvent
func NewAPIRequestEvent(sessionID vo.SessionID, model string, inputTokens, outputTokens int) *APIRequestEvent {
	return &APIRequestEvent{
		BaseEvent: newBaseEvent(
			"api.request",
			sessionID.String(),
			"Session",
			map[string]interface{}{
				"sessionId":    sessionID.String(),
				"model":        model,
				"inputTokens":  inputTokens,
				"outputTokens": outputTokens,
			},
		),
	}
}

// APIErrorEvent is emitted when an API error occurs
type APIErrorEvent struct {
	BaseEvent
}

// NewAPIErrorEvent creates a new APIErrorEvent
func NewAPIErrorEvent(sessionID vo.SessionID, errorType, errorMessage string) *APIErrorEvent {
	return &APIErrorEvent{
		BaseEvent: newBaseEvent(
			"api.error",
			sessionID.String(),
			"Session",
			map[string]interface{}{
				"sessionId":    sessionID.String(),
				"errorType":    errorType,
				"errorMessage": errorMessage,
			},
		),
	}
}

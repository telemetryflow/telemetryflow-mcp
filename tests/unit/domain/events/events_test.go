// Package events_test contains unit tests for domain events
package events_test

import (
	"testing"
	"time"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/events"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestNewSessionCreatedEvent(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	event := events.NewSessionCreatedEvent(sessionID)

	if event.EventType() != "session.created" {
		t.Errorf("EventType() = %v, want session.created", event.EventType())
	}
	if event.AggregateType() != "Session" {
		t.Errorf("AggregateType() = %v, want Session", event.AggregateType())
	}
	if event.AggregateID() != sessionID.String() {
		t.Errorf("AggregateID() = %v, want %v", event.AggregateID(), sessionID.String())
	}
	if event.EventID() == "" {
		t.Error("EventID() should not be empty")
	}
	if event.OccurredAt().IsZero() {
		t.Error("OccurredAt() should not be zero")
	}

	payload := event.Payload()
	if payload["sessionId"] != sessionID.String() {
		t.Errorf("Payload sessionId = %v, want %v", payload["sessionId"], sessionID.String())
	}
}

func TestNewSessionInitializedEvent(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	clientName := "test-client"
	clientVersion := "1.0.0"

	event := events.NewSessionInitializedEvent(sessionID, clientName, clientVersion)

	if event.EventType() != "session.initialized" {
		t.Errorf("EventType() = %v, want session.initialized", event.EventType())
	}

	payload := event.Payload()
	if payload["clientName"] != clientName {
		t.Errorf("Payload clientName = %v, want %v", payload["clientName"], clientName)
	}
	if payload["clientVersion"] != clientVersion {
		t.Errorf("Payload clientVersion = %v, want %v", payload["clientVersion"], clientVersion)
	}
}

func TestNewSessionClosedEvent(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	event := events.NewSessionClosedEvent(sessionID)

	if event.EventType() != "session.closed" {
		t.Errorf("EventType() = %v, want session.closed", event.EventType())
	}
	if event.AggregateID() != sessionID.String() {
		t.Errorf("AggregateID() = %v, want %v", event.AggregateID(), sessionID.String())
	}
}

func TestNewConversationCreatedEvent(t *testing.T) {
	conversationID := vo.GenerateConversationID()
	sessionID := vo.GenerateSessionID()
	model := vo.ModelClaude4Sonnet

	event := events.NewConversationCreatedEvent(conversationID, sessionID, model)

	if event.EventType() != "conversation.created" {
		t.Errorf("EventType() = %v, want conversation.created", event.EventType())
	}
	if event.AggregateType() != "Conversation" {
		t.Errorf("AggregateType() = %v, want Conversation", event.AggregateType())
	}
	if event.AggregateID() != conversationID.String() {
		t.Errorf("AggregateID() = %v, want %v", event.AggregateID(), conversationID.String())
	}

	payload := event.Payload()
	if payload["sessionId"] != sessionID.String() {
		t.Errorf("Payload sessionId = %v, want %v", payload["sessionId"], sessionID.String())
	}
	if payload["model"] != model.String() {
		t.Errorf("Payload model = %v, want %v", payload["model"], model.String())
	}
}

func TestNewConversationClosedEvent(t *testing.T) {
	conversationID := vo.GenerateConversationID()
	event := events.NewConversationClosedEvent(conversationID)

	if event.EventType() != "conversation.closed" {
		t.Errorf("EventType() = %v, want conversation.closed", event.EventType())
	}
}

func TestNewMessageAddedEvent(t *testing.T) {
	conversationID := vo.GenerateConversationID()
	messageID := vo.GenerateMessageID()
	role := vo.RoleUser

	event := events.NewMessageAddedEvent(conversationID, messageID, role)

	if event.EventType() != "message.added" {
		t.Errorf("EventType() = %v, want message.added", event.EventType())
	}

	payload := event.Payload()
	if payload["messageId"] != messageID.String() {
		t.Errorf("Payload messageId = %v, want %v", payload["messageId"], messageID.String())
	}
	if payload["role"] != role.String() {
		t.Errorf("Payload role = %v, want %v", payload["role"], role.String())
	}
}

func TestNewToolRegisteredEvent(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	toolName := "test_tool"

	event := events.NewToolRegisteredEvent(sessionID, toolName)

	if event.EventType() != "tool.registered" {
		t.Errorf("EventType() = %v, want tool.registered", event.EventType())
	}

	payload := event.Payload()
	if payload["toolName"] != toolName {
		t.Errorf("Payload toolName = %v, want %v", payload["toolName"], toolName)
	}
}

func TestNewToolExecutedEvent(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	toolName := "test_tool"
	success := true
	duration := 100 * time.Millisecond

	event := events.NewToolExecutedEvent(sessionID, toolName, success, duration)

	if event.EventType() != "tool.executed" {
		t.Errorf("EventType() = %v, want tool.executed", event.EventType())
	}

	payload := event.Payload()
	if payload["success"] != success {
		t.Errorf("Payload success = %v, want %v", payload["success"], success)
	}
	if payload["durationMs"] != duration.Milliseconds() {
		t.Errorf("Payload durationMs = %v, want %v", payload["durationMs"], duration.Milliseconds())
	}
}

func TestNewResourceRegisteredEvent(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	resourceURI := "file:///path/to/resource"

	event := events.NewResourceRegisteredEvent(sessionID, resourceURI)

	if event.EventType() != "resource.registered" {
		t.Errorf("EventType() = %v, want resource.registered", event.EventType())
	}

	payload := event.Payload()
	if payload["resourceUri"] != resourceURI {
		t.Errorf("Payload resourceUri = %v, want %v", payload["resourceUri"], resourceURI)
	}
}

func TestNewResourceReadEvent(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	resourceURI := "file:///path/to/resource"
	success := true

	event := events.NewResourceReadEvent(sessionID, resourceURI, success)

	if event.EventType() != "resource.read" {
		t.Errorf("EventType() = %v, want resource.read", event.EventType())
	}

	payload := event.Payload()
	if payload["success"] != success {
		t.Errorf("Payload success = %v, want %v", payload["success"], success)
	}
}

func TestNewPromptRegisteredEvent(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	promptName := "test_prompt"

	event := events.NewPromptRegisteredEvent(sessionID, promptName)

	if event.EventType() != "prompt.registered" {
		t.Errorf("EventType() = %v, want prompt.registered", event.EventType())
	}

	payload := event.Payload()
	if payload["promptName"] != promptName {
		t.Errorf("Payload promptName = %v, want %v", payload["promptName"], promptName)
	}
}

func TestNewPromptExecutedEvent(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	promptName := "test_prompt"
	success := false

	event := events.NewPromptExecutedEvent(sessionID, promptName, success)

	if event.EventType() != "prompt.executed" {
		t.Errorf("EventType() = %v, want prompt.executed", event.EventType())
	}

	payload := event.Payload()
	if payload["success"] != success {
		t.Errorf("Payload success = %v, want %v", payload["success"], success)
	}
}

func TestNewAPIRequestEvent(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	model := "claude-3-opus-20240229"
	inputTokens := 100
	outputTokens := 200

	event := events.NewAPIRequestEvent(sessionID, model, inputTokens, outputTokens)

	if event.EventType() != "api.request" {
		t.Errorf("EventType() = %v, want api.request", event.EventType())
	}

	payload := event.Payload()
	if payload["model"] != model {
		t.Errorf("Payload model = %v, want %v", payload["model"], model)
	}
	if payload["inputTokens"] != inputTokens {
		t.Errorf("Payload inputTokens = %v, want %v", payload["inputTokens"], inputTokens)
	}
	if payload["outputTokens"] != outputTokens {
		t.Errorf("Payload outputTokens = %v, want %v", payload["outputTokens"], outputTokens)
	}
}

func TestNewAPIErrorEvent(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	errorType := "rate_limit_error"
	errorMessage := "Too many requests"

	event := events.NewAPIErrorEvent(sessionID, errorType, errorMessage)

	if event.EventType() != "api.error" {
		t.Errorf("EventType() = %v, want api.error", event.EventType())
	}

	payload := event.Payload()
	if payload["errorType"] != errorType {
		t.Errorf("Payload errorType = %v, want %v", payload["errorType"], errorType)
	}
	if payload["errorMessage"] != errorMessage {
		t.Errorf("Payload errorMessage = %v, want %v", payload["errorMessage"], errorMessage)
	}
}

func TestDomainEventInterface(t *testing.T) {
	// Verify all events implement DomainEvent interface
	var _ events.DomainEvent = (*events.SessionCreatedEvent)(nil)
	var _ events.DomainEvent = (*events.SessionInitializedEvent)(nil)
	var _ events.DomainEvent = (*events.SessionClosedEvent)(nil)
	var _ events.DomainEvent = (*events.ConversationCreatedEvent)(nil)
	var _ events.DomainEvent = (*events.ConversationClosedEvent)(nil)
	var _ events.DomainEvent = (*events.MessageAddedEvent)(nil)
	var _ events.DomainEvent = (*events.ToolRegisteredEvent)(nil)
	var _ events.DomainEvent = (*events.ToolExecutedEvent)(nil)
	var _ events.DomainEvent = (*events.ResourceRegisteredEvent)(nil)
	var _ events.DomainEvent = (*events.ResourceReadEvent)(nil)
	var _ events.DomainEvent = (*events.PromptRegisteredEvent)(nil)
	var _ events.DomainEvent = (*events.PromptExecutedEvent)(nil)
	var _ events.DomainEvent = (*events.APIRequestEvent)(nil)
	var _ events.DomainEvent = (*events.APIErrorEvent)(nil)
}

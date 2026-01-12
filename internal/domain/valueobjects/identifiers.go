// Package valueobjects contains immutable, self-validating value objects
// following DDD patterns for the TelemetryFlow GO MCP service.
package valueobjects

import (
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// Common errors for value object validation
var (
	ErrInvalidID             = errors.New("invalid identifier format")
	ErrInvalidConversationID = errors.New("invalid conversation ID format")
	ErrInvalidMessageID      = errors.New("invalid message ID format")
	ErrInvalidToolID         = errors.New("invalid tool ID format")
	ErrInvalidResourceID     = errors.New("invalid resource ID format")
	ErrInvalidPromptID       = errors.New("invalid prompt ID format")
	ErrInvalidSessionID      = errors.New("invalid session ID format")
	ErrEmptyID               = errors.New("identifier cannot be empty")
)

// ConversationID represents a unique identifier for a conversation
type ConversationID struct {
	value string
}

// NewConversationID creates a new ConversationID with validation
func NewConversationID(value string) (ConversationID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return ConversationID{}, ErrEmptyID
	}
	if _, err := uuid.Parse(value); err != nil {
		return ConversationID{}, ErrInvalidConversationID
	}
	return ConversationID{value: value}, nil
}

// GenerateConversationID creates a new random ConversationID
func GenerateConversationID() ConversationID {
	return ConversationID{value: uuid.New().String()}
}

// String returns the string representation of the ID
func (c ConversationID) String() string {
	return c.value
}

// IsEmpty checks if the ID is empty
func (c ConversationID) IsEmpty() bool {
	return c.value == ""
}

// Equals compares two ConversationIDs
func (c ConversationID) Equals(other ConversationID) bool {
	return c.value == other.value
}

// MessageID represents a unique identifier for a message
type MessageID struct {
	value string
}

// NewMessageID creates a new MessageID with validation
func NewMessageID(value string) (MessageID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return MessageID{}, ErrEmptyID
	}
	if _, err := uuid.Parse(value); err != nil {
		return MessageID{}, ErrInvalidMessageID
	}
	return MessageID{value: value}, nil
}

// GenerateMessageID creates a new random MessageID
func GenerateMessageID() MessageID {
	return MessageID{value: uuid.New().String()}
}

// String returns the string representation of the ID
func (m MessageID) String() string {
	return m.value
}

// IsEmpty checks if the ID is empty
func (m MessageID) IsEmpty() bool {
	return m.value == ""
}

// Equals compares two MessageIDs
func (m MessageID) Equals(other MessageID) bool {
	return m.value == other.value
}

// ToolID represents a unique identifier for an MCP tool
type ToolID struct {
	value string
}

// Tool ID pattern: alphanumeric with underscores and hyphens, max 64 chars
var toolIDPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{0,63}$`)

// NewToolID creates a new ToolID with validation
func NewToolID(value string) (ToolID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return ToolID{}, ErrEmptyID
	}
	if !toolIDPattern.MatchString(value) {
		return ToolID{}, ErrInvalidToolID
	}
	return ToolID{value: value}, nil
}

// String returns the string representation of the ID
func (t ToolID) String() string {
	return t.value
}

// IsEmpty checks if the ID is empty
func (t ToolID) IsEmpty() bool {
	return t.value == ""
}

// Equals compares two ToolIDs
func (t ToolID) Equals(other ToolID) bool {
	return t.value == other.value
}

// ResourceID represents a unique identifier for an MCP resource
type ResourceID struct {
	value string
}

// Resource ID pattern: URI-like format
var resourceIDPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*://[^\s]+$`)

// NewResourceID creates a new ResourceID with validation
func NewResourceID(value string) (ResourceID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return ResourceID{}, ErrEmptyID
	}
	if !resourceIDPattern.MatchString(value) {
		return ResourceID{}, ErrInvalidResourceID
	}
	return ResourceID{value: value}, nil
}

// String returns the string representation of the ID
func (r ResourceID) String() string {
	return r.value
}

// IsEmpty checks if the ID is empty
func (r ResourceID) IsEmpty() bool {
	return r.value == ""
}

// Equals compares two ResourceIDs
func (r ResourceID) Equals(other ResourceID) bool {
	return r.value == other.value
}

// PromptID represents a unique identifier for an MCP prompt
type PromptID struct {
	value string
}

// Prompt ID pattern: alphanumeric with underscores and hyphens
var promptIDPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]{0,63}$`)

// NewPromptID creates a new PromptID with validation
func NewPromptID(value string) (PromptID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return PromptID{}, ErrEmptyID
	}
	if !promptIDPattern.MatchString(value) {
		return PromptID{}, ErrInvalidPromptID
	}
	return PromptID{value: value}, nil
}

// String returns the string representation of the ID
func (p PromptID) String() string {
	return p.value
}

// IsEmpty checks if the ID is empty
func (p PromptID) IsEmpty() bool {
	return p.value == ""
}

// Equals compares two PromptIDs
func (p PromptID) Equals(other PromptID) bool {
	return p.value == other.value
}

// SessionID represents a unique identifier for an MCP session
type SessionID struct {
	value string
}

// NewSessionID creates a new SessionID with validation
func NewSessionID(value string) (SessionID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return SessionID{}, ErrEmptyID
	}
	if _, err := uuid.Parse(value); err != nil {
		return SessionID{}, ErrInvalidSessionID
	}
	return SessionID{value: value}, nil
}

// GenerateSessionID creates a new random SessionID
func GenerateSessionID() SessionID {
	return SessionID{value: uuid.New().String()}
}

// String returns the string representation of the ID
func (s SessionID) String() string {
	return s.value
}

// IsEmpty checks if the ID is empty
func (s SessionID) IsEmpty() bool {
	return s.value == ""
}

// Equals compares two SessionIDs
func (s SessionID) Equals(other SessionID) bool {
	return s.value == other.value
}

// RequestID represents a unique identifier for an MCP request
type RequestID struct {
	value string
}

// NewRequestID creates a new RequestID with validation
func NewRequestID(value string) (RequestID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return RequestID{}, ErrEmptyID
	}
	return RequestID{value: value}, nil
}

// GenerateRequestID creates a new random RequestID
func GenerateRequestID() RequestID {
	return RequestID{value: uuid.New().String()}
}

// String returns the string representation of the ID
func (r RequestID) String() string {
	return r.value
}

// IsEmpty checks if the ID is empty
func (r RequestID) IsEmpty() bool {
	return r.value == ""
}

// Equals compares two RequestIDs
func (r RequestID) Equals(other RequestID) bool {
	return r.value == other.value
}

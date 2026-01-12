// Package entities contains domain entities for the TelemetryFlow GO MCP service
package entities

import (
	"time"

	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// Message represents a message entity in a conversation
type Message struct {
	id        vo.MessageID
	role      vo.Role
	content   []ContentBlock
	createdAt time.Time
	metadata  map[string]interface{}
}

// ContentBlock represents a block of content within a message
type ContentBlock struct {
	Type      vo.ContentType         `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`          // For tool_use
	Name      string                 `json:"name,omitempty"`        // For tool_use
	Input     map[string]interface{} `json:"input,omitempty"`       // For tool_use
	ToolUseID string                 `json:"tool_use_id,omitempty"` // For tool_result
	Content   string                 `json:"content,omitempty"`     // For tool_result
	IsError   bool                   `json:"is_error,omitempty"`    // For tool_result
	Source    *ImageSource           `json:"source,omitempty"`      // For image
}

// ImageSource represents an image source
type ImageSource struct {
	Type      string `json:"type"` // "base64" or "url"
	MediaType string `json:"media_type,omitempty"`
	Data      string `json:"data,omitempty"`
	URL       string `json:"url,omitempty"`
}

// NewMessage creates a new Message entity
func NewMessage(role vo.Role, content []ContentBlock) (*Message, error) {
	if !role.IsValid() {
		return nil, vo.ErrInvalidRole
	}

	return &Message{
		id:        vo.GenerateMessageID(),
		role:      role,
		content:   content,
		createdAt: time.Now().UTC(),
		metadata:  make(map[string]interface{}),
	}, nil
}

// NewTextMessage creates a new text message
func NewTextMessage(role vo.Role, text string) (*Message, error) {
	content := []ContentBlock{
		{
			Type: vo.ContentTypeText,
			Text: text,
		},
	}
	return NewMessage(role, content)
}

// ID returns the message ID
func (m *Message) ID() vo.MessageID {
	return m.id
}

// Role returns the message role
func (m *Message) Role() vo.Role {
	return m.role
}

// Content returns the message content blocks
func (m *Message) Content() []ContentBlock {
	return m.content
}

// CreatedAt returns the creation timestamp
func (m *Message) CreatedAt() time.Time {
	return m.createdAt
}

// Metadata returns the message metadata
func (m *Message) Metadata() map[string]interface{} {
	return m.metadata
}

// SetMetadata sets a metadata value
func (m *Message) SetMetadata(key string, value interface{}) {
	m.metadata[key] = value
}

// GetMetadata gets a metadata value
func (m *Message) GetMetadata(key string) (interface{}, bool) {
	v, ok := m.metadata[key]
	return v, ok
}

// GetTextContent returns all text content from the message
func (m *Message) GetTextContent() string {
	var text string
	for _, block := range m.content {
		if block.Type == vo.ContentTypeText {
			text += block.Text
		}
	}
	return text
}

// HasToolUse checks if the message contains tool use blocks
func (m *Message) HasToolUse() bool {
	for _, block := range m.content {
		if block.Type == vo.ContentTypeToolUse {
			return true
		}
	}
	return false
}

// GetToolUseBlocks returns all tool use blocks from the message
func (m *Message) GetToolUseBlocks() []ContentBlock {
	var toolUses []ContentBlock
	for _, block := range m.content {
		if block.Type == vo.ContentTypeToolUse {
			toolUses = append(toolUses, block)
		}
	}
	return toolUses
}

// AddContent adds a content block to the message
func (m *Message) AddContent(block ContentBlock) {
	m.content = append(m.content, block)
}

// IsUserMessage checks if the message is from a user
func (m *Message) IsUserMessage() bool {
	return m.role == vo.RoleUser
}

// IsAssistantMessage checks if the message is from the assistant
func (m *Message) IsAssistantMessage() bool {
	return m.role == vo.RoleAssistant
}

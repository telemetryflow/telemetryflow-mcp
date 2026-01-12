// Package aggregates contains domain aggregates for the TelemetryFlow GO MCP service
package aggregates

import (
	"errors"
	"sync"
	"time"

	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/events"
	vo "github.com/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
)

// Common conversation errors
var (
	ErrConversationNotFound  = errors.New("conversation not found")
	ErrConversationClosed    = errors.New("conversation is closed")
	ErrEmptyMessage          = errors.New("message cannot be empty")
	ErrInvalidMessageOrder   = errors.New("invalid message order")
	ErrMaxMessagesExceeded   = errors.New("maximum messages exceeded")
	ErrSystemPromptImmutable = errors.New("system prompt cannot be changed after conversation started")
)

// ConversationStatus represents the status of a conversation
type ConversationStatus string

const (
	ConversationStatusActive   ConversationStatus = "active"
	ConversationStatusPaused   ConversationStatus = "paused"
	ConversationStatusClosed   ConversationStatus = "closed"
	ConversationStatusArchived ConversationStatus = "archived"
)

// MaxMessages is the maximum number of messages allowed in a conversation
const MaxMessages = 10000

// Conversation represents a conversation aggregate
type Conversation struct {
	mu sync.RWMutex

	id            vo.ConversationID
	sessionID     vo.SessionID
	model         vo.Model
	systemPrompt  vo.SystemPrompt
	messages      []*entities.Message
	status        ConversationStatus
	maxTokens     int
	temperature   float64
	topP          float64
	topK          int
	stopSequences []string
	tools         []*entities.Tool
	createdAt     time.Time
	updatedAt     time.Time
	closedAt      *time.Time
	metadata      map[string]interface{}
	events        []events.DomainEvent
}

// NewConversation creates a new Conversation aggregate
func NewConversation(sessionID vo.SessionID, model vo.Model) *Conversation {
	now := time.Now().UTC()
	conv := &Conversation{
		id:          vo.GenerateConversationID(),
		sessionID:   sessionID,
		model:       model,
		messages:    make([]*entities.Message, 0),
		status:      ConversationStatusActive,
		maxTokens:   4096, // Default max tokens
		temperature: 1.0,  // Default temperature
		topP:        1.0,  // Default top_p
		topK:        0,    // Default top_k (0 = disabled)
		tools:       make([]*entities.Tool, 0),
		createdAt:   now,
		updatedAt:   now,
		metadata:    make(map[string]interface{}),
		events:      make([]events.DomainEvent, 0),
	}

	conv.addEvent(events.NewConversationCreatedEvent(conv.id, sessionID, model))
	return conv
}

// ID returns the conversation ID
func (c *Conversation) ID() vo.ConversationID {
	return c.id
}

// SessionID returns the session ID
func (c *Conversation) SessionID() vo.SessionID {
	return c.sessionID
}

// Model returns the model
func (c *Conversation) Model() vo.Model {
	return c.model
}

// SetModel sets the model
func (c *Conversation) SetModel(model vo.Model) error {
	if !model.IsValid() {
		return vo.ErrInvalidModel
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.model = model
	c.updatedAt = time.Now().UTC()
	return nil
}

// SystemPrompt returns the system prompt
func (c *Conversation) SystemPrompt() vo.SystemPrompt {
	return c.systemPrompt
}

// SetSystemPrompt sets the system prompt
func (c *Conversation) SetSystemPrompt(prompt vo.SystemPrompt) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Only allow setting system prompt if no user messages yet
	if c.hasUserMessages() {
		return ErrSystemPromptImmutable
	}

	c.systemPrompt = prompt
	c.updatedAt = time.Now().UTC()
	return nil
}

// Messages returns all messages
func (c *Conversation) Messages() []*entities.Message {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.messages
}

// MessageCount returns the number of messages
func (c *Conversation) MessageCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.messages)
}

// LastMessage returns the last message
func (c *Conversation) LastMessage() *entities.Message {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.messages) == 0 {
		return nil
	}
	return c.messages[len(c.messages)-1]
}

// AddMessage adds a message to the conversation
func (c *Conversation) AddMessage(message *entities.Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.status == ConversationStatusClosed {
		return ErrConversationClosed
	}

	if len(c.messages) >= MaxMessages {
		return ErrMaxMessagesExceeded
	}

	// Validate message order (user -> assistant -> user -> ...)
	if len(c.messages) > 0 {
		lastMsg := c.messages[len(c.messages)-1]
		if lastMsg.Role() == message.Role() && message.Role() != vo.RoleAssistant {
			// Allow consecutive assistant messages (tool use responses)
			return ErrInvalidMessageOrder
		}
	}

	c.messages = append(c.messages, message)
	c.updatedAt = time.Now().UTC()

	c.addEvent(events.NewMessageAddedEvent(c.id, message.ID(), message.Role()))
	return nil
}

// AddUserMessage adds a user message
func (c *Conversation) AddUserMessage(text string) (*entities.Message, error) {
	msg, err := entities.NewTextMessage(vo.RoleUser, text)
	if err != nil {
		return nil, err
	}
	if err := c.AddMessage(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// AddAssistantMessage adds an assistant message
func (c *Conversation) AddAssistantMessage(content []entities.ContentBlock) (*entities.Message, error) {
	msg, err := entities.NewMessage(vo.RoleAssistant, content)
	if err != nil {
		return nil, err
	}
	if err := c.AddMessage(msg); err != nil {
		return nil, err
	}
	return msg, nil
}

// Status returns the conversation status
func (c *Conversation) Status() ConversationStatus {
	return c.status
}

// IsActive returns whether the conversation is active
func (c *Conversation) IsActive() bool {
	return c.status == ConversationStatusActive
}

// Pause pauses the conversation
func (c *Conversation) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.status == ConversationStatusActive {
		c.status = ConversationStatusPaused
		c.updatedAt = time.Now().UTC()
	}
}

// Resume resumes the conversation
func (c *Conversation) Resume() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.status == ConversationStatusPaused {
		c.status = ConversationStatusActive
		c.updatedAt = time.Now().UTC()
	}
}

// Close closes the conversation
func (c *Conversation) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.status != ConversationStatusClosed {
		c.status = ConversationStatusClosed
		now := time.Now().UTC()
		c.closedAt = &now
		c.updatedAt = now
		c.addEvent(events.NewConversationClosedEvent(c.id))
	}
}

// Archive archives the conversation
func (c *Conversation) Archive() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.status == ConversationStatusClosed {
		c.status = ConversationStatusArchived
		c.updatedAt = time.Now().UTC()
	}
}

// MaxTokens returns the max tokens setting
func (c *Conversation) MaxTokens() int {
	return c.maxTokens
}

// SetMaxTokens sets the max tokens
func (c *Conversation) SetMaxTokens(maxTokens int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxTokens = maxTokens
	c.updatedAt = time.Now().UTC()
}

// Temperature returns the temperature setting
func (c *Conversation) Temperature() float64 {
	return c.temperature
}

// SetTemperature sets the temperature
func (c *Conversation) SetTemperature(temperature float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if temperature < 0 {
		temperature = 0
	}
	if temperature > 2 {
		temperature = 2
	}
	c.temperature = temperature
	c.updatedAt = time.Now().UTC()
}

// TopP returns the top_p setting
func (c *Conversation) TopP() float64 {
	return c.topP
}

// SetTopP sets the top_p
func (c *Conversation) SetTopP(topP float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if topP < 0 {
		topP = 0
	}
	if topP > 1 {
		topP = 1
	}
	c.topP = topP
	c.updatedAt = time.Now().UTC()
}

// TopK returns the top_k setting
func (c *Conversation) TopK() int {
	return c.topK
}

// SetTopK sets the top_k
func (c *Conversation) SetTopK(topK int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if topK < 0 {
		topK = 0
	}
	c.topK = topK
	c.updatedAt = time.Now().UTC()
}

// StopSequences returns the stop sequences
func (c *Conversation) StopSequences() []string {
	return c.stopSequences
}

// SetStopSequences sets the stop sequences
func (c *Conversation) SetStopSequences(sequences []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stopSequences = sequences
	c.updatedAt = time.Now().UTC()
}

// Tools returns the available tools
func (c *Conversation) Tools() []*entities.Tool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tools
}

// AddTool adds a tool to the conversation
func (c *Conversation) AddTool(tool *entities.Tool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tools = append(c.tools, tool)
	c.updatedAt = time.Now().UTC()
}

// RemoveTool removes a tool by name
func (c *Conversation) RemoveTool(name vo.ToolName) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i, tool := range c.tools {
		if tool.Name().Equals(name) {
			c.tools = append(c.tools[:i], c.tools[i+1:]...)
			c.updatedAt = time.Now().UTC()
			return
		}
	}
}

// GetTool gets a tool by name
func (c *Conversation) GetTool(name vo.ToolName) *entities.Tool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, tool := range c.tools {
		if tool.Name().Equals(name) {
			return tool
		}
	}
	return nil
}

// CreatedAt returns the creation timestamp
func (c *Conversation) CreatedAt() time.Time {
	return c.createdAt
}

// UpdatedAt returns the last update timestamp
func (c *Conversation) UpdatedAt() time.Time {
	return c.updatedAt
}

// ClosedAt returns the closed timestamp
func (c *Conversation) ClosedAt() *time.Time {
	return c.closedAt
}

// Metadata returns the conversation metadata
func (c *Conversation) Metadata() map[string]interface{} {
	return c.metadata
}

// SetMetadata sets a metadata value
func (c *Conversation) SetMetadata(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metadata[key] = value
	c.updatedAt = time.Now().UTC()
}

// GetMetadata gets a metadata value
func (c *Conversation) GetMetadata(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.metadata[key]
	return v, ok
}

// Events returns and clears domain events
func (c *Conversation) Events() []events.DomainEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	evts := c.events
	c.events = make([]events.DomainEvent, 0)
	return evts
}

// ClearEvents clears domain events
func (c *Conversation) ClearEvents() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = make([]events.DomainEvent, 0)
}

// addEvent adds a domain event
func (c *Conversation) addEvent(event events.DomainEvent) {
	c.events = append(c.events, event)
}

// hasUserMessages checks if there are any user messages
func (c *Conversation) hasUserMessages() bool {
	for _, msg := range c.messages {
		if msg.IsUserMessage() {
			return true
		}
	}
	return false
}

// GetMessagesForAPI returns messages formatted for the Claude API
func (c *Conversation) GetMessagesForAPI() []map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]map[string]interface{}, len(c.messages))
	for i, msg := range c.messages {
		content := make([]map[string]interface{}, len(msg.Content()))
		for j, block := range msg.Content() {
			contentBlock := map[string]interface{}{
				"type": block.Type.String(),
			}
			switch block.Type {
			case vo.ContentTypeText:
				contentBlock["text"] = block.Text
			case vo.ContentTypeToolUse:
				contentBlock["id"] = block.ID
				contentBlock["name"] = block.Name
				contentBlock["input"] = block.Input
			case vo.ContentTypeToolResult:
				contentBlock["tool_use_id"] = block.ToolUseID
				contentBlock["content"] = block.Content
				if block.IsError {
					contentBlock["is_error"] = block.IsError
				}
			}
			content[j] = contentBlock
		}

		result[i] = map[string]interface{}{
			"role":    msg.Role().String(),
			"content": content,
		}
	}

	return result
}

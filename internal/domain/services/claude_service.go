// Package services contains domain services for the TelemetryFlow GO MCP service
package services

import (
	"context"

	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
)

// ClaudeRequest represents a request to the Claude API
type ClaudeRequest struct {
	Model         vo.Model
	SystemPrompt  vo.SystemPrompt
	Messages      []ClaudeMessage
	MaxTokens     int
	Temperature   float64
	TopP          float64
	TopK          int
	StopSequences []string
	Tools         []ClaudeTool
	Stream        bool
	Metadata      map[string]interface{}
}

// ClaudeMessage represents a message in the Claude API format
type ClaudeMessage struct {
	Role    vo.Role
	Content []entities.ContentBlock
}

// ClaudeTool represents a tool in the Claude API format
type ClaudeTool struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	InputSchema *entities.JSONSchema `json:"input_schema"`
}

// ClaudeResponse represents a response from the Claude API
type ClaudeResponse struct {
	ID           string
	Type         string
	Role         vo.Role
	Content      []entities.ContentBlock
	Model        string
	StopReason   string
	StopSequence string
	Usage        *ClaudeUsage
}

// ClaudeUsage represents token usage information
type ClaudeUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ClaudeStreamEvent represents a streaming event from the Claude API
type ClaudeStreamEvent struct {
	Type         string
	Index        int
	ContentBlock *entities.ContentBlock
	Delta        *ClaudeDelta
	Message      *ClaudeResponse
	Usage        *ClaudeUsage
	Error        error
}

// ClaudeDelta represents a delta update in streaming
type ClaudeDelta struct {
	Type         string `json:"type"`
	Text         string `json:"text,omitempty"`
	PartialJSON  string `json:"partial_json,omitempty"`
	StopReason   string `json:"stop_reason,omitempty"`
	StopSequence string `json:"stop_sequence,omitempty"`
}

// IClaudeService defines the interface for Claude API interactions
type IClaudeService interface {
	// CreateMessage creates a message (non-streaming)
	CreateMessage(ctx context.Context, request *ClaudeRequest) (*ClaudeResponse, error)

	// CreateMessageStream creates a message with streaming
	CreateMessageStream(ctx context.Context, request *ClaudeRequest) (<-chan *ClaudeStreamEvent, error)

	// CountTokens counts tokens for a message
	CountTokens(ctx context.Context, request *ClaudeRequest) (int, error)

	// ValidateRequest validates a Claude request
	ValidateRequest(request *ClaudeRequest) error
}

// IConversationService defines the interface for conversation management
type IConversationService interface {
	// SendMessage sends a message and gets a response
	SendMessage(ctx context.Context, conversation *aggregates.Conversation, text string) (*ClaudeResponse, error)

	// SendMessageWithTools sends a message with tool use support
	SendMessageWithTools(ctx context.Context, conversation *aggregates.Conversation, text string, tools []*entities.Tool) (*ClaudeResponse, error)

	// SendMessageStream sends a message with streaming response
	SendMessageStream(ctx context.Context, conversation *aggregates.Conversation, text string) (<-chan *ClaudeStreamEvent, error)

	// ProcessToolUse processes tool use requests from Claude
	ProcessToolUse(ctx context.Context, conversation *aggregates.Conversation, toolUseBlocks []entities.ContentBlock) (*ClaudeResponse, error)

	// BuildRequest builds a Claude request from a conversation
	BuildRequest(conversation *aggregates.Conversation) (*ClaudeRequest, error)
}

// IToolExecutionService defines the interface for tool execution
type IToolExecutionService interface {
	// Execute executes a tool
	Execute(ctx context.Context, tool *entities.Tool, input map[string]interface{}) (*entities.ToolResult, error)

	// ExecuteWithTimeout executes a tool with a timeout
	ExecuteWithTimeout(ctx context.Context, tool *entities.Tool, input map[string]interface{}) (*entities.ToolResult, error)

	// ValidateInput validates tool input against the schema
	ValidateInput(tool *entities.Tool, input map[string]interface{}) error
}

// IResourceService defines the interface for resource management
type IResourceService interface {
	// Read reads a resource
	Read(ctx context.Context, resource *entities.Resource) (*entities.ResourceContent, error)

	// ReadByURI reads a resource by URI
	ReadByURI(ctx context.Context, uri string) (*entities.ResourceContent, error)

	// Subscribe subscribes to resource updates
	Subscribe(ctx context.Context, uri string, callback func(*entities.ResourceContent)) error

	// Unsubscribe unsubscribes from resource updates
	Unsubscribe(ctx context.Context, uri string) error

	// NotifyUpdate notifies subscribers of a resource update
	NotifyUpdate(ctx context.Context, uri string, content *entities.ResourceContent) error
}

// IPromptService defines the interface for prompt management
type IPromptService interface {
	// Generate generates prompt messages
	Generate(ctx context.Context, prompt *entities.Prompt, args map[string]string) (*entities.PromptMessages, error)

	// Validate validates prompt arguments
	Validate(prompt *entities.Prompt, args map[string]string) error
}

// IEventPublisher defines the interface for publishing domain events
type IEventPublisher interface {
	// Publish publishes a domain event
	Publish(ctx context.Context, event interface{}) error

	// PublishAll publishes multiple domain events
	PublishAll(ctx context.Context, events []interface{}) error

	// Subscribe subscribes to domain events
	Subscribe(eventType string, handler func(interface{})) error

	// Unsubscribe unsubscribes from domain events
	Unsubscribe(eventType string) error
}

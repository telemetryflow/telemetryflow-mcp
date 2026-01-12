// Package repositories contains repository interfaces for the TelemetryFlow GO MCP service
package repositories

import (
	"context"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// ISessionRepository defines the interface for session persistence
type ISessionRepository interface {
	// Save persists a session
	Save(ctx context.Context, session *aggregates.Session) error

	// FindByID retrieves a session by ID
	FindByID(ctx context.Context, id vo.SessionID) (*aggregates.Session, error)

	// FindAll retrieves all sessions
	FindAll(ctx context.Context) ([]*aggregates.Session, error)

	// FindActive retrieves all active sessions
	FindActive(ctx context.Context) ([]*aggregates.Session, error)

	// Delete removes a session
	Delete(ctx context.Context, id vo.SessionID) error

	// Exists checks if a session exists
	Exists(ctx context.Context, id vo.SessionID) (bool, error)

	// Count returns the total number of sessions
	Count(ctx context.Context) (int, error)
}

// IConversationRepository defines the interface for conversation persistence
type IConversationRepository interface {
	// Save persists a conversation
	Save(ctx context.Context, conversation *aggregates.Conversation) error

	// FindByID retrieves a conversation by ID
	FindByID(ctx context.Context, id vo.ConversationID) (*aggregates.Conversation, error)

	// FindBySessionID retrieves conversations by session ID
	FindBySessionID(ctx context.Context, sessionID vo.SessionID) ([]*aggregates.Conversation, error)

	// FindActive retrieves all active conversations
	FindActive(ctx context.Context) ([]*aggregates.Conversation, error)

	// Delete removes a conversation
	Delete(ctx context.Context, id vo.ConversationID) error

	// Exists checks if a conversation exists
	Exists(ctx context.Context, id vo.ConversationID) (bool, error)

	// Count returns the total number of conversations
	Count(ctx context.Context) (int, error)

	// CountBySessionID returns the number of conversations for a session
	CountBySessionID(ctx context.Context, sessionID vo.SessionID) (int, error)
}

// IToolRepository defines the interface for tool registry
type IToolRepository interface {
	// Register registers a tool
	Register(ctx context.Context, tool *entities.Tool) error

	// Unregister removes a tool
	Unregister(ctx context.Context, name vo.ToolName) error

	// FindByName retrieves a tool by name
	FindByName(ctx context.Context, name vo.ToolName) (*entities.Tool, error)

	// FindAll retrieves all tools
	FindAll(ctx context.Context) ([]*entities.Tool, error)

	// FindByCategory retrieves tools by category
	FindByCategory(ctx context.Context, category string) ([]*entities.Tool, error)

	// FindByTag retrieves tools by tag
	FindByTag(ctx context.Context, tag string) ([]*entities.Tool, error)

	// FindEnabled retrieves all enabled tools
	FindEnabled(ctx context.Context) ([]*entities.Tool, error)

	// Exists checks if a tool exists
	Exists(ctx context.Context, name vo.ToolName) (bool, error)

	// Count returns the total number of tools
	Count(ctx context.Context) (int, error)
}

// IResourceRepository defines the interface for resource registry
type IResourceRepository interface {
	// Register registers a resource
	Register(ctx context.Context, resource *entities.Resource) error

	// Unregister removes a resource
	Unregister(ctx context.Context, uri vo.ResourceURI) error

	// FindByURI retrieves a resource by URI
	FindByURI(ctx context.Context, uri vo.ResourceURI) (*entities.Resource, error)

	// FindAll retrieves all resources
	FindAll(ctx context.Context) ([]*entities.Resource, error)

	// FindTemplates retrieves all resource templates
	FindTemplates(ctx context.Context) ([]*entities.Resource, error)

	// Exists checks if a resource exists
	Exists(ctx context.Context, uri vo.ResourceURI) (bool, error)

	// Count returns the total number of resources
	Count(ctx context.Context) (int, error)
}

// IPromptRepository defines the interface for prompt registry
type IPromptRepository interface {
	// Register registers a prompt
	Register(ctx context.Context, prompt *entities.Prompt) error

	// Unregister removes a prompt
	Unregister(ctx context.Context, name vo.ToolName) error

	// FindByName retrieves a prompt by name
	FindByName(ctx context.Context, name vo.ToolName) (*entities.Prompt, error)

	// FindAll retrieves all prompts
	FindAll(ctx context.Context) ([]*entities.Prompt, error)

	// Exists checks if a prompt exists
	Exists(ctx context.Context, name vo.ToolName) (bool, error)

	// Count returns the total number of prompts
	Count(ctx context.Context) (int, error)
}

// IEventRepository defines the interface for domain event persistence
type IEventRepository interface {
	// Store stores a domain event
	Store(ctx context.Context, event interface{}) error

	// StoreAll stores multiple domain events
	StoreAll(ctx context.Context, events []interface{}) error

	// FindByAggregateID retrieves events by aggregate ID
	FindByAggregateID(ctx context.Context, aggregateID string) ([]interface{}, error)

	// FindByEventType retrieves events by type
	FindByEventType(ctx context.Context, eventType string) ([]interface{}, error)

	// FindAll retrieves all events with pagination
	FindAll(ctx context.Context, offset, limit int) ([]interface{}, error)

	// Count returns the total number of events
	Count(ctx context.Context) (int, error)
}

// ICacheRepository defines the interface for caching
type ICacheRepository interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (interface{}, error)

	// Set stores a value in cache
	Set(ctx context.Context, key string, value interface{}, ttlSeconds int) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// Clear clears all cache entries
	Clear(ctx context.Context) error
}

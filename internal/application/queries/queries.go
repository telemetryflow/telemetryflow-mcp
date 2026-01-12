// Package queries contains CQRS queries for the TelemetryFlow GO MCP service
package queries

import (
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// Query is the base interface for all queries
type Query interface {
	QueryName() string
}

// Session Queries

// GetSessionQuery retrieves a session by ID
type GetSessionQuery struct {
	SessionID vo.SessionID
}

func (q *GetSessionQuery) QueryName() string {
	return "GetSession"
}

// ListSessionsQuery lists all sessions
type ListSessionsQuery struct {
	ActiveOnly bool
	Cursor     string
	Limit      int
}

func (q *ListSessionsQuery) QueryName() string {
	return "ListSessions"
}

// GetSessionStatsQuery gets session statistics
type GetSessionStatsQuery struct {
	SessionID vo.SessionID
}

func (q *GetSessionStatsQuery) QueryName() string {
	return "GetSessionStats"
}

// Conversation Queries

// GetConversationQuery retrieves a conversation by ID
type GetConversationQuery struct {
	ConversationID vo.ConversationID
}

func (q *GetConversationQuery) QueryName() string {
	return "GetConversation"
}

// ListConversationsQuery lists conversations
type ListConversationsQuery struct {
	SessionID  vo.SessionID
	ActiveOnly bool
	Cursor     string
	Limit      int
}

func (q *ListConversationsQuery) QueryName() string {
	return "ListConversations"
}

// GetConversationMessagesQuery gets messages from a conversation
type GetConversationMessagesQuery struct {
	ConversationID vo.ConversationID
	Offset         int
	Limit          int
}

func (q *GetConversationMessagesQuery) QueryName() string {
	return "GetConversationMessages"
}

// Tool Queries

// GetToolQuery retrieves a tool by name
type GetToolQuery struct {
	SessionID vo.SessionID
	Name      string
}

func (q *GetToolQuery) QueryName() string {
	return "GetTool"
}

// ListToolsQuery lists all tools
type ListToolsQuery struct {
	SessionID   vo.SessionID
	Category    string
	Tag         string
	EnabledOnly bool
	Cursor      string
	Limit       int
}

func (q *ListToolsQuery) QueryName() string {
	return "ListTools"
}

// Resource Queries

// GetResourceQuery retrieves a resource by URI
type GetResourceQuery struct {
	SessionID vo.SessionID
	URI       string
}

func (q *GetResourceQuery) QueryName() string {
	return "GetResource"
}

// ReadResourceQuery reads the content of a resource
type ReadResourceQuery struct {
	SessionID vo.SessionID
	URI       string
}

func (q *ReadResourceQuery) QueryName() string {
	return "ReadResource"
}

// ListResourcesQuery lists all resources
type ListResourcesQuery struct {
	SessionID     vo.SessionID
	TemplatesOnly bool
	Cursor        string
	Limit         int
}

func (q *ListResourcesQuery) QueryName() string {
	return "ListResources"
}

// Prompt Queries

// GetPromptQuery retrieves a prompt by name
type GetPromptQuery struct {
	SessionID vo.SessionID
	Name      string
	Arguments map[string]string
}

func (q *GetPromptQuery) QueryName() string {
	return "GetPrompt"
}

// ListPromptsQuery lists all prompts
type ListPromptsQuery struct {
	SessionID vo.SessionID
	Cursor    string
	Limit     int
}

func (q *ListPromptsQuery) QueryName() string {
	return "ListPrompts"
}

// Completion Queries

// CompleteQuery handles completion requests
type CompleteQuery struct {
	SessionID vo.SessionID
	Ref       CompletionRef
	Argument  CompletionArgument
}

func (q *CompleteQuery) QueryName() string {
	return "Complete"
}

// CompletionRef identifies the prompt or resource for completion
type CompletionRef struct {
	Type string // "ref/prompt" or "ref/resource"
	Name string // Prompt name or resource URI
}

// CompletionArgument represents the argument being completed
type CompletionArgument struct {
	Name  string
	Value string
}

// Health Queries

// HealthCheckQuery checks the health of the service
type HealthCheckQuery struct{}

func (q *HealthCheckQuery) QueryName() string {
	return "HealthCheck"
}

// GetMetricsQuery gets service metrics
type GetMetricsQuery struct {
	SessionID vo.SessionID
}

func (q *GetMetricsQuery) QueryName() string {
	return "GetMetrics"
}

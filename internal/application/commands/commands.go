// Package commands contains CQRS commands for the TelemetryFlow GO MCP service
package commands

import (
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// Command is the base interface for all commands
type Command interface {
	CommandName() string
}

// Session Commands

// InitializeSessionCommand initializes a new MCP session
type InitializeSessionCommand struct {
	ClientName      string
	ClientVersion   string
	ProtocolVersion string
	Capabilities    map[string]interface{}
}

func (c *InitializeSessionCommand) CommandName() string {
	return "InitializeSession"
}

// CloseSessionCommand closes an MCP session
type CloseSessionCommand struct {
	SessionID vo.SessionID
}

func (c *CloseSessionCommand) CommandName() string {
	return "CloseSession"
}

// SetLogLevelCommand sets the log level for a session
type SetLogLevelCommand struct {
	SessionID vo.SessionID
	Level     vo.MCPLogLevel
}

func (c *SetLogLevelCommand) CommandName() string {
	return "SetLogLevel"
}

// Conversation Commands

// CreateConversationCommand creates a new conversation
type CreateConversationCommand struct {
	SessionID    vo.SessionID
	Model        vo.Model
	SystemPrompt string
	MaxTokens    int
	Temperature  float64
}

func (c *CreateConversationCommand) CommandName() string {
	return "CreateConversation"
}

// SendMessageCommand sends a message in a conversation
type SendMessageCommand struct {
	ConversationID vo.ConversationID
	Content        string
	Stream         bool
}

func (c *SendMessageCommand) CommandName() string {
	return "SendMessage"
}

// AddToolResultCommand adds a tool result to a conversation
type AddToolResultCommand struct {
	ConversationID vo.ConversationID
	ToolUseID      string
	Content        string
	IsError        bool
}

func (c *AddToolResultCommand) CommandName() string {
	return "AddToolResult"
}

// CloseConversationCommand closes a conversation
type CloseConversationCommand struct {
	ConversationID vo.ConversationID
}

func (c *CloseConversationCommand) CommandName() string {
	return "CloseConversation"
}

// Tool Commands

// RegisterToolCommand registers a new tool
type RegisterToolCommand struct {
	SessionID   vo.SessionID
	Name        string
	Description string
	InputSchema *entities.JSONSchema
	Category    string
	Tags        []string
}

func (c *RegisterToolCommand) CommandName() string {
	return "RegisterTool"
}

// UnregisterToolCommand unregisters a tool
type UnregisterToolCommand struct {
	SessionID vo.SessionID
	Name      string
}

func (c *UnregisterToolCommand) CommandName() string {
	return "UnregisterTool"
}

// ExecuteToolCommand executes a tool
type ExecuteToolCommand struct {
	SessionID vo.SessionID
	Name      string
	Arguments map[string]interface{}
}

func (c *ExecuteToolCommand) CommandName() string {
	return "ExecuteTool"
}

// Resource Commands

// RegisterResourceCommand registers a new resource
type RegisterResourceCommand struct {
	SessionID   vo.SessionID
	URI         string
	Name        string
	Description string
	MimeType    string
}

func (c *RegisterResourceCommand) CommandName() string {
	return "RegisterResource"
}

// UnregisterResourceCommand unregisters a resource
type UnregisterResourceCommand struct {
	SessionID vo.SessionID
	URI       string
}

func (c *UnregisterResourceCommand) CommandName() string {
	return "UnregisterResource"
}

// SubscribeResourceCommand subscribes to a resource
type SubscribeResourceCommand struct {
	SessionID vo.SessionID
	URI       string
}

func (c *SubscribeResourceCommand) CommandName() string {
	return "SubscribeResource"
}

// UnsubscribeResourceCommand unsubscribes from a resource
type UnsubscribeResourceCommand struct {
	SessionID vo.SessionID
	URI       string
}

func (c *UnsubscribeResourceCommand) CommandName() string {
	return "UnsubscribeResource"
}

// Prompt Commands

// RegisterPromptCommand registers a new prompt
type RegisterPromptCommand struct {
	SessionID   vo.SessionID
	Name        string
	Description string
	Arguments   []*entities.PromptArgument
}

func (c *RegisterPromptCommand) CommandName() string {
	return "RegisterPrompt"
}

// UnregisterPromptCommand unregisters a prompt
type UnregisterPromptCommand struct {
	SessionID vo.SessionID
	Name      string
}

func (c *UnregisterPromptCommand) CommandName() string {
	return "UnregisterPrompt"
}

// ExecutePromptCommand executes a prompt
type ExecutePromptCommand struct {
	SessionID vo.SessionID
	Name      string
	Arguments map[string]string
}

func (c *ExecutePromptCommand) CommandName() string {
	return "ExecutePrompt"
}

// MCP Protocol Commands

// PingCommand handles ping requests
type PingCommand struct {
	SessionID vo.SessionID
}

func (c *PingCommand) CommandName() string {
	return "Ping"
}

// CancelRequestCommand cancels a pending request
type CancelRequestCommand struct {
	SessionID vo.SessionID
	RequestID string
	Reason    string
}

func (c *CancelRequestCommand) CommandName() string {
	return "CancelRequest"
}

// SendNotificationCommand sends a notification to the client
type SendNotificationCommand struct {
	SessionID vo.SessionID
	Method    vo.MCPMethod
	Params    map[string]interface{}
}

func (c *SendNotificationCommand) CommandName() string {
	return "SendNotification"
}

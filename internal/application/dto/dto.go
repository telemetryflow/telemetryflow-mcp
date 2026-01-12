// Package dto provides Data Transfer Objects for the application layer
package dto

import (
	"time"

	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// SessionDTO represents a session data transfer object
type SessionDTO struct {
	ID            string    `json:"id"`
	State         string    `json:"state"`
	ClientName    string    `json:"clientName,omitempty"`
	ClientVersion string    `json:"clientVersion,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// ConversationDTO represents a conversation data transfer object
type ConversationDTO struct {
	ID           string       `json:"id"`
	SessionID    string       `json:"sessionId"`
	Model        string       `json:"model"`
	SystemPrompt string       `json:"systemPrompt,omitempty"`
	Messages     []MessageDTO `json:"messages"`
	IsActive     bool         `json:"isActive"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    time.Time    `json:"updatedAt"`
}

// MessageDTO represents a message data transfer object
type MessageDTO struct {
	ID        string            `json:"id"`
	Role      string            `json:"role"`
	Content   []ContentBlockDTO `json:"content"`
	CreatedAt time.Time         `json:"createdAt"`
}

// ContentBlockDTO represents a content block data transfer object
type ContentBlockDTO struct {
	Type      string                 `json:"type"`
	Text      string                 `json:"text,omitempty"`
	ID        string                 `json:"id,omitempty"`
	Name      string                 `json:"name,omitempty"`
	Input     map[string]interface{} `json:"input,omitempty"`
	ToolUseID string                 `json:"tool_use_id,omitempty"`
	Content   string                 `json:"content,omitempty"`
	IsError   bool                   `json:"is_error,omitempty"`
}

// ToolDTO represents a tool data transfer object
type ToolDTO struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ResourceDTO represents a resource data transfer object
type ResourceDTO struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// PromptDTO represents a prompt data transfer object
type PromptDTO struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Arguments   []ArgumentDTO `json:"arguments,omitempty"`
}

// ArgumentDTO represents a prompt argument data transfer object
type ArgumentDTO struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// InitializeRequest represents an MCP initialize request
type InitializeRequest struct {
	ProtocolVersion string                `json:"protocolVersion"`
	Capabilities    ClientCapabilitiesDTO `json:"capabilities"`
	ClientInfo      ClientInfoDTO         `json:"clientInfo"`
}

// ClientCapabilitiesDTO represents client capabilities
type ClientCapabilitiesDTO struct {
	Experimental map[string]interface{} `json:"experimental,omitempty"`
	Sampling     *SamplingCapabilityDTO `json:"sampling,omitempty"`
	Roots        *RootsCapabilityDTO    `json:"roots,omitempty"`
}

// SamplingCapabilityDTO represents sampling capability
type SamplingCapabilityDTO struct{}

// RootsCapabilityDTO represents roots capability
type RootsCapabilityDTO struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ClientInfoDTO represents client information
type ClientInfoDTO struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult represents an MCP initialize result
type InitializeResult struct {
	ProtocolVersion string                `json:"protocolVersion"`
	Capabilities    ServerCapabilitiesDTO `json:"capabilities"`
	ServerInfo      ServerInfoDTO         `json:"serverInfo"`
	Instructions    string                `json:"instructions,omitempty"`
}

// ServerCapabilitiesDTO represents server capabilities
type ServerCapabilitiesDTO struct {
	Experimental map[string]interface{}  `json:"experimental,omitempty"`
	Logging      *LoggingCapabilityDTO   `json:"logging,omitempty"`
	Prompts      *PromptsCapabilityDTO   `json:"prompts,omitempty"`
	Resources    *ResourcesCapabilityDTO `json:"resources,omitempty"`
	Tools        *ToolsCapabilityDTO     `json:"tools,omitempty"`
}

// LoggingCapabilityDTO represents logging capability
type LoggingCapabilityDTO struct{}

// PromptsCapabilityDTO represents prompts capability
type PromptsCapabilityDTO struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapabilityDTO represents resources capability
type ResourcesCapabilityDTO struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ToolsCapabilityDTO represents tools capability
type ToolsCapabilityDTO struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ServerInfoDTO represents server information
type ServerInfoDTO struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// CallToolRequest represents a tool call request
type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallToolResult represents a tool call result
type CallToolResult struct {
	Content []ContentBlockDTO `json:"content"`
	IsError bool              `json:"isError,omitempty"`
}

// ToModel converts a string to a Model value object
func ToModel(s string) vo.Model {
	return vo.Model(s)
}

// ToRole converts a string to a Role value object
func ToRole(s string) vo.Role {
	return vo.Role(s)
}

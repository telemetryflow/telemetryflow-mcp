// Package entities contains domain entities for the TelemetryFlow GO MCP service
package entities

import (
	"encoding/json"
	"time"

	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// Tool represents an MCP tool entity
type Tool struct {
	name        vo.ToolName
	description vo.ToolDescription
	inputSchema *JSONSchema
	handler     ToolHandler
	category    string
	tags        []string
	isEnabled   bool
	rateLimit   *RateLimit
	timeout     time.Duration
	createdAt   time.Time
	updatedAt   time.Time
	metadata    map[string]interface{}
}

// ToolHandler is the function signature for tool execution
type ToolHandler func(input map[string]interface{}) (*ToolResult, error)

// JSONSchema represents a JSON Schema for tool input validation
type JSONSchema struct {
	Type                 string                 `json:"type"`
	Properties           map[string]*JSONSchema `json:"properties,omitempty"`
	Required             []string               `json:"required,omitempty"`
	Items                *JSONSchema            `json:"items,omitempty"`
	Description          string                 `json:"description,omitempty"`
	Default              interface{}            `json:"default,omitempty"`
	Enum                 []interface{}          `json:"enum,omitempty"`
	Minimum              *float64               `json:"minimum,omitempty"`
	Maximum              *float64               `json:"maximum,omitempty"`
	MinLength            *int                   `json:"minLength,omitempty"`
	MaxLength            *int                   `json:"maxLength,omitempty"`
	Pattern              string                 `json:"pattern,omitempty"`
	Format               string                 `json:"format,omitempty"`
	AdditionalProperties *bool                  `json:"additionalProperties,omitempty"`
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Content []ToolResultContent `json:"content"`
	IsError bool                `json:"isError,omitempty"`
}

// ToolResultContent represents content in a tool result
type ToolResultContent struct {
	Type     string `json:"type"` // "text", "image", "resource"
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`     // For image (base64)
	MimeType string `json:"mimeType,omitempty"` // For image
	URI      string `json:"uri,omitempty"`      // For resource
}

// RateLimit represents rate limiting configuration for a tool
type RateLimit struct {
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
}

// NewTool creates a new Tool entity
func NewTool(name vo.ToolName, description vo.ToolDescription, inputSchema *JSONSchema) (*Tool, error) {
	now := time.Now().UTC()
	return &Tool{
		name:        name,
		description: description,
		inputSchema: inputSchema,
		isEnabled:   true,
		timeout:     30 * time.Second, // Default timeout
		createdAt:   now,
		updatedAt:   now,
		metadata:    make(map[string]interface{}),
	}, nil
}

// Name returns the tool name
func (t *Tool) Name() vo.ToolName {
	return t.name
}

// Description returns the tool description
func (t *Tool) Description() vo.ToolDescription {
	return t.description
}

// InputSchema returns the tool input schema
func (t *Tool) InputSchema() *JSONSchema {
	return t.inputSchema
}

// Handler returns the tool handler
func (t *Tool) Handler() ToolHandler {
	return t.handler
}

// SetHandler sets the tool handler
func (t *Tool) SetHandler(handler ToolHandler) {
	t.handler = handler
	t.updatedAt = time.Now().UTC()
}

// Category returns the tool category
func (t *Tool) Category() string {
	return t.category
}

// SetCategory sets the tool category
func (t *Tool) SetCategory(category string) {
	t.category = category
	t.updatedAt = time.Now().UTC()
}

// Tags returns the tool tags
func (t *Tool) Tags() []string {
	return t.tags
}

// SetTags sets the tool tags
func (t *Tool) SetTags(tags []string) {
	t.tags = tags
	t.updatedAt = time.Now().UTC()
}

// AddTag adds a tag to the tool
func (t *Tool) AddTag(tag string) {
	t.tags = append(t.tags, tag)
	t.updatedAt = time.Now().UTC()
}

// IsEnabled returns whether the tool is enabled
func (t *Tool) IsEnabled() bool {
	return t.isEnabled
}

// Enable enables the tool
func (t *Tool) Enable() {
	t.isEnabled = true
	t.updatedAt = time.Now().UTC()
}

// Disable disables the tool
func (t *Tool) Disable() {
	t.isEnabled = false
	t.updatedAt = time.Now().UTC()
}

// RateLimitConfig returns the rate limit configuration
func (t *Tool) RateLimitConfig() *RateLimit {
	return t.rateLimit
}

// SetRateLimit sets the rate limit configuration
func (t *Tool) SetRateLimit(limit *RateLimit) {
	t.rateLimit = limit
	t.updatedAt = time.Now().UTC()
}

// Timeout returns the tool timeout
func (t *Tool) Timeout() time.Duration {
	return t.timeout
}

// SetTimeout sets the tool timeout
func (t *Tool) SetTimeout(timeout time.Duration) {
	t.timeout = timeout
	t.updatedAt = time.Now().UTC()
}

// CreatedAt returns the creation timestamp
func (t *Tool) CreatedAt() time.Time {
	return t.createdAt
}

// UpdatedAt returns the last update timestamp
func (t *Tool) UpdatedAt() time.Time {
	return t.updatedAt
}

// Metadata returns the tool metadata
func (t *Tool) Metadata() map[string]interface{} {
	return t.metadata
}

// SetMetadata sets a metadata value
func (t *Tool) SetMetadata(key string, value interface{}) {
	t.metadata[key] = value
	t.updatedAt = time.Now().UTC()
}

// Execute executes the tool with the given input
func (t *Tool) Execute(input map[string]interface{}) (*ToolResult, error) {
	if t.handler == nil {
		return &ToolResult{
			Content: []ToolResultContent{{Type: "text", Text: "Tool handler not configured"}},
			IsError: true,
		}, nil
	}
	return t.handler(input)
}

// ToMCPTool converts the tool to MCP format
func (t *Tool) ToMCPTool() map[string]interface{} {
	result := map[string]interface{}{
		"name":        t.name.String(),
		"description": t.description.String(),
	}
	if t.inputSchema != nil {
		result["inputSchema"] = t.inputSchema
	}
	return result
}

// ToJSON returns the tool as JSON bytes
func (t *Tool) ToJSON() ([]byte, error) {
	return json.Marshal(t.ToMCPTool())
}

// NewTextToolResult creates a text tool result
func NewTextToolResult(text string) *ToolResult {
	return &ToolResult{
		Content: []ToolResultContent{
			{Type: "text", Text: text},
		},
	}
}

// NewErrorToolResult creates an error tool result
func NewErrorToolResult(err error) *ToolResult {
	return &ToolResult{
		Content: []ToolResultContent{
			{Type: "text", Text: err.Error()},
		},
		IsError: true,
	}
}

// NewImageToolResult creates an image tool result
func NewImageToolResult(data, mimeType string) *ToolResult {
	return &ToolResult{
		Content: []ToolResultContent{
			{Type: "image", Data: data, MimeType: mimeType},
		},
	}
}

// NewResourceToolResult creates a resource tool result
func NewResourceToolResult(uri, text, mimeType string) *ToolResult {
	return &ToolResult{
		Content: []ToolResultContent{
			{Type: "resource", URI: uri, Text: text, MimeType: mimeType},
		},
	}
}

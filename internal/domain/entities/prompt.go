// Package entities contains domain entities for the TelemetryFlow GO MCP service
package entities

import (
	"encoding/json"
	"time"

	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// Prompt represents an MCP prompt entity
type Prompt struct {
	name        vo.ToolName // Reusing ToolName validation as prompt names follow same pattern
	description string
	arguments   []*PromptArgument
	generator   PromptGenerator
	createdAt   time.Time
	updatedAt   time.Time
	metadata    map[string]interface{}
}

// PromptArgument represents an argument for a prompt
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// PromptGenerator is the function signature for generating prompt messages
type PromptGenerator func(args map[string]string) (*PromptMessages, error)

// PromptMessages represents the generated messages from a prompt
type PromptMessages struct {
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

// PromptMessage represents a message in prompt output
type PromptMessage struct {
	Role    string        `json:"role"`
	Content PromptContent `json:"content"`
}

// PromptContent represents content in a prompt message
type PromptContent struct {
	Type     string `json:"type"` // "text", "image", "resource"
	Text     string `json:"text,omitempty"`
	Data     string `json:"data,omitempty"`     // For image
	MimeType string `json:"mimeType,omitempty"` // For image
	URI      string `json:"uri,omitempty"`      // For resource
}

// NewPrompt creates a new Prompt entity
func NewPrompt(name vo.ToolName, description string) (*Prompt, error) {
	now := time.Now().UTC()
	return &Prompt{
		name:        name,
		description: description,
		arguments:   make([]*PromptArgument, 0),
		createdAt:   now,
		updatedAt:   now,
		metadata:    make(map[string]interface{}),
	}, nil
}

// Name returns the prompt name
func (p *Prompt) Name() vo.ToolName {
	return p.name
}

// Description returns the prompt description
func (p *Prompt) Description() string {
	return p.description
}

// SetDescription sets the prompt description
func (p *Prompt) SetDescription(description string) {
	p.description = description
	p.updatedAt = time.Now().UTC()
}

// Arguments returns the prompt arguments
func (p *Prompt) Arguments() []*PromptArgument {
	return p.arguments
}

// AddArgument adds an argument to the prompt
func (p *Prompt) AddArgument(arg *PromptArgument) {
	p.arguments = append(p.arguments, arg)
	p.updatedAt = time.Now().UTC()
}

// RemoveArgument removes an argument by name
func (p *Prompt) RemoveArgument(name string) {
	for i, arg := range p.arguments {
		if arg.Name == name {
			p.arguments = append(p.arguments[:i], p.arguments[i+1:]...)
			p.updatedAt = time.Now().UTC()
			return
		}
	}
}

// GetArgument gets an argument by name
func (p *Prompt) GetArgument(name string) *PromptArgument {
	for _, arg := range p.arguments {
		if arg.Name == name {
			return arg
		}
	}
	return nil
}

// RequiredArguments returns the required arguments
func (p *Prompt) RequiredArguments() []*PromptArgument {
	var required []*PromptArgument
	for _, arg := range p.arguments {
		if arg.Required {
			required = append(required, arg)
		}
	}
	return required
}

// Generator returns the prompt generator
func (p *Prompt) Generator() PromptGenerator {
	return p.generator
}

// SetGenerator sets the prompt generator
func (p *Prompt) SetGenerator(generator PromptGenerator) {
	p.generator = generator
	p.updatedAt = time.Now().UTC()
}

// CreatedAt returns the creation timestamp
func (p *Prompt) CreatedAt() time.Time {
	return p.createdAt
}

// UpdatedAt returns the last update timestamp
func (p *Prompt) UpdatedAt() time.Time {
	return p.updatedAt
}

// Metadata returns the prompt metadata
func (p *Prompt) Metadata() map[string]interface{} {
	return p.metadata
}

// SetMetadata sets a metadata value
func (p *Prompt) SetMetadata(key string, value interface{}) {
	p.metadata[key] = value
	p.updatedAt = time.Now().UTC()
}

// Generate generates the prompt messages
func (p *Prompt) Generate(args map[string]string) (*PromptMessages, error) {
	if p.generator == nil {
		// Return default message if no generator
		return &PromptMessages{
			Description: p.description,
			Messages: []PromptMessage{
				{
					Role: "user",
					Content: PromptContent{
						Type: "text",
						Text: p.description,
					},
				},
			},
		}, nil
	}
	return p.generator(args)
}

// ValidateArguments validates the provided arguments
func (p *Prompt) ValidateArguments(args map[string]string) error {
	for _, required := range p.RequiredArguments() {
		if _, ok := args[required.Name]; !ok {
			return &MissingArgumentError{ArgumentName: required.Name}
		}
	}
	return nil
}

// MissingArgumentError represents a missing required argument error
type MissingArgumentError struct {
	ArgumentName string
}

func (e *MissingArgumentError) Error() string {
	return "missing required argument: " + e.ArgumentName
}

// ToMCPPrompt converts the prompt to MCP format
func (p *Prompt) ToMCPPrompt() map[string]interface{} {
	result := map[string]interface{}{
		"name": p.name.String(),
	}

	if p.description != "" {
		result["description"] = p.description
	}

	if len(p.arguments) > 0 {
		args := make([]map[string]interface{}, len(p.arguments))
		for i, arg := range p.arguments {
			argMap := map[string]interface{}{
				"name": arg.Name,
			}
			if arg.Description != "" {
				argMap["description"] = arg.Description
			}
			if arg.Required {
				argMap["required"] = arg.Required
			}
			args[i] = argMap
		}
		result["arguments"] = args
	}

	return result
}

// ToJSON returns the prompt as JSON bytes
func (p *Prompt) ToJSON() ([]byte, error) {
	return json.Marshal(p.ToMCPPrompt())
}

// PromptList represents a list of prompts
type PromptList struct {
	Prompts    []*Prompt
	NextCursor string
}

// NewPromptList creates a new PromptList
func NewPromptList() *PromptList {
	return &PromptList{
		Prompts: make([]*Prompt, 0),
	}
}

// Add adds a prompt to the list
func (pl *PromptList) Add(prompt *Prompt) {
	pl.Prompts = append(pl.Prompts, prompt)
}

// Count returns the number of prompts
func (pl *PromptList) Count() int {
	return len(pl.Prompts)
}

// IsEmpty returns whether the list is empty
func (pl *PromptList) IsEmpty() bool {
	return len(pl.Prompts) == 0
}

// ToMCPPromptList converts the list to MCP format
func (pl *PromptList) ToMCPPromptList() map[string]interface{} {
	prompts := make([]map[string]interface{}, len(pl.Prompts))
	for i, p := range pl.Prompts {
		prompts[i] = p.ToMCPPrompt()
	}

	result := map[string]interface{}{
		"prompts": prompts,
	}
	if pl.NextCursor != "" {
		result["nextCursor"] = pl.NextCursor
	}
	return result
}

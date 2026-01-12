// Package claude contains the Claude API client implementation
package claude

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/rs/zerolog"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/services"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/config"
)

// Client errors
var (
	ErrAPIKeyRequired     = errors.New("API key is required")
	ErrInvalidRequest     = errors.New("invalid request")
	ErrAPIError           = errors.New("API error")
	ErrRateLimited        = errors.New("rate limited")
	ErrContextCancelled   = errors.New("context cancelled")
	ErrMaxRetriesExceeded = errors.New("max retries exceeded")
)

// Client implements the Claude API client
type Client struct {
	client anthropic.Client
	config *config.ClaudeConfig
	logger zerolog.Logger
}

// NewClient creates a new Claude API client
func NewClient(cfg *config.ClaudeConfig, logger zerolog.Logger) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, ErrAPIKeyRequired
	}

	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
	}

	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	client := anthropic.NewClient(opts...)

	return &Client{
		client: client,
		config: cfg,
		logger: logger.With().Str("component", "claude-client").Logger(),
	}, nil
}

// CreateMessage creates a message (non-streaming)
func (c *Client) CreateMessage(ctx context.Context, request *services.ClaudeRequest) (*services.ClaudeResponse, error) {
	if err := c.ValidateRequest(request); err != nil {
		return nil, err
	}

	c.logger.Debug().
		Str("model", request.Model.String()).
		Int("max_tokens", request.MaxTokens).
		Int("message_count", len(request.Messages)).
		Msg("Creating message")

	// Build API request
	params := c.buildMessageParams(request)

	// Execute with retry
	var response *anthropic.Message
	var err error

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			c.logger.Debug().Int("attempt", attempt).Msg("Retrying API request")
			time.Sleep(c.config.RetryDelay * time.Duration(attempt))
		}

		response, err = c.client.Messages.New(ctx, params)
		if err == nil {
			break
		}

		// Check if error is retryable
		if !c.isRetryableError(err) {
			return nil, fmt.Errorf("%w: %v", ErrAPIError, err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrMaxRetriesExceeded, err)
	}

	return c.convertResponse(response), nil
}

// CreateMessageStream creates a message with streaming
func (c *Client) CreateMessageStream(ctx context.Context, request *services.ClaudeRequest) (<-chan *services.ClaudeStreamEvent, error) {
	if err := c.ValidateRequest(request); err != nil {
		return nil, err
	}

	c.logger.Debug().
		Str("model", request.Model.String()).
		Int("max_tokens", request.MaxTokens).
		Msg("Creating streaming message")

	params := c.buildMessageParams(request)

	eventChan := make(chan *services.ClaudeStreamEvent, 100)

	go func() {
		defer close(eventChan)

		stream := c.client.Messages.NewStreaming(ctx, params)

		for stream.Next() {
			event := stream.Current()
			streamEvent := c.convertStreamEvent(event)
			if streamEvent != nil {
				select {
				case eventChan <- streamEvent:
				case <-ctx.Done():
					eventChan <- &services.ClaudeStreamEvent{Error: ctx.Err()}
					return
				}
			}
		}

		if err := stream.Err(); err != nil {
			eventChan <- &services.ClaudeStreamEvent{Error: err}
		}
	}()

	return eventChan, nil
}

// CountTokens counts tokens for a message
func (c *Client) CountTokens(ctx context.Context, request *services.ClaudeRequest) (int, error) {
	if err := c.ValidateRequest(request); err != nil {
		return 0, err
	}

	// Build messages for token counting
	messages := c.buildMessages(request.Messages)

	params := anthropic.MessageCountTokensParams{
		Model:    anthropic.Model(request.Model.String()),
		Messages: messages,
	}

	if !request.SystemPrompt.IsEmpty() {
		params.System = anthropic.MessageCountTokensParamsSystemUnion{
			OfMessageCountTokenssSystemArray: []anthropic.TextBlockParam{
				{Text: request.SystemPrompt.String()},
			},
		}
	}

	result, err := c.client.Messages.CountTokens(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrAPIError, err)
	}

	return int(result.InputTokens), nil
}

// ValidateRequest validates a Claude request
func (c *Client) ValidateRequest(request *services.ClaudeRequest) error {
	if request == nil {
		return ErrInvalidRequest
	}

	if !request.Model.IsValid() {
		return fmt.Errorf("%w: invalid model", ErrInvalidRequest)
	}

	if len(request.Messages) == 0 {
		return fmt.Errorf("%w: messages required", ErrInvalidRequest)
	}

	if request.MaxTokens <= 0 {
		request.MaxTokens = c.config.MaxTokens
	}

	return nil
}

// buildMessageParams builds the API request parameters
func (c *Client) buildMessageParams(request *services.ClaudeRequest) anthropic.MessageNewParams {
	messages := c.buildMessages(request.Messages)

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(request.Model.String()),
		MaxTokens: int64(request.MaxTokens),
		Messages:  messages,
	}

	// System prompt
	if !request.SystemPrompt.IsEmpty() {
		params.System = []anthropic.TextBlockParam{
			{Text: request.SystemPrompt.String()},
		}
	}

	// Temperature (only set if not default)
	if request.Temperature > 0 && request.Temperature != 1.0 {
		params.Temperature = anthropic.Float(request.Temperature)
	}

	// Top P
	if request.TopP > 0 && request.TopP < 1.0 {
		params.TopP = anthropic.Float(request.TopP)
	}

	// Top K
	if request.TopK > 0 {
		params.TopK = anthropic.Int(int64(request.TopK))
	}

	// Stop sequences
	if len(request.StopSequences) > 0 {
		params.StopSequences = request.StopSequences
	}

	// Tools
	if len(request.Tools) > 0 {
		params.Tools = c.buildTools(request.Tools)
	}

	return params
}

// buildMessages builds API messages from domain messages
func (c *Client) buildMessages(messages []services.ClaudeMessage) []anthropic.MessageParam {
	result := make([]anthropic.MessageParam, len(messages))

	for i, msg := range messages {
		var content []anthropic.ContentBlockParamUnion

		for _, block := range msg.Content {
			switch block.Type {
			case vo.ContentTypeText:
				content = append(content, anthropic.NewTextBlock(block.Text))

			case vo.ContentTypeToolUse:
				// Tool use blocks are only in assistant responses
				content = append(content, anthropic.ContentBlockParamOfRequestToolUseBlock(
					block.ID,
					block.Input,
					block.Name,
				))

			case vo.ContentTypeToolResult:
				content = append(content, anthropic.NewToolResultBlock(
					block.ToolUseID,
					block.Content,
					block.IsError,
				))
			}
		}

		result[i] = anthropic.MessageParam{
			Role:    anthropic.MessageParamRole(msg.Role.String()),
			Content: content,
		}
	}

	return result
}

// buildTools builds API tools from domain tools
func (c *Client) buildTools(tools []services.ClaudeTool) []anthropic.ToolUnionParam {
	result := make([]anthropic.ToolUnionParam, len(tools))

	for i, tool := range tools {
		inputSchema := c.convertJSONSchema(tool.InputSchema)

		toolParam := anthropic.ToolUnionParamOfTool(inputSchema, tool.Name)
		if toolParam.OfTool != nil {
			toolParam.OfTool.Description = anthropic.String(tool.Description)
		}
		result[i] = toolParam
	}

	return result
}

// convertJSONSchema converts domain JSON schema to API format
func (c *Client) convertJSONSchema(schema *entities.JSONSchema) anthropic.ToolInputSchemaParam {
	if schema == nil {
		return anthropic.ToolInputSchemaParam{}
	}

	properties := make(map[string]interface{})
	for name, prop := range schema.Properties {
		properties[name] = c.convertSchemaProperty(prop)
	}

	return anthropic.ToolInputSchemaParam{
		Properties: properties,
	}
}

// convertSchemaProperty converts a schema property
func (c *Client) convertSchemaProperty(prop *entities.JSONSchema) map[string]interface{} {
	if prop == nil {
		return nil
	}

	result := map[string]interface{}{
		"type": prop.Type,
	}

	if prop.Description != "" {
		result["description"] = prop.Description
	}

	if len(prop.Enum) > 0 {
		result["enum"] = prop.Enum
	}

	return result
}

// convertResponse converts API response to domain response
func (c *Client) convertResponse(msg *anthropic.Message) *services.ClaudeResponse {
	content := make([]entities.ContentBlock, 0, len(msg.Content))

	for _, block := range msg.Content {
		switch block.Type {
		case "text":
			content = append(content, entities.ContentBlock{
				Type: vo.ContentTypeText,
				Text: block.Text,
			})

		case "tool_use":
			input := make(map[string]interface{})
			if len(block.Input) > 0 {
				// Parse JSON from RawMessage
				_ = json.Unmarshal(block.Input, &input)
			}
			content = append(content, entities.ContentBlock{
				Type:  vo.ContentTypeToolUse,
				ID:    block.ID,
				Name:  block.Name,
				Input: input,
			})
		}
	}

	return &services.ClaudeResponse{
		ID:         msg.ID,
		Type:       string(msg.Type),
		Role:       vo.RoleAssistant,
		Content:    content,
		Model:      string(msg.Model),
		StopReason: string(msg.StopReason),
		Usage: &services.ClaudeUsage{
			InputTokens:  int(msg.Usage.InputTokens),
			OutputTokens: int(msg.Usage.OutputTokens),
		},
	}
}

// convertStreamEvent converts a streaming event
func (c *Client) convertStreamEvent(event anthropic.MessageStreamEventUnion) *services.ClaudeStreamEvent {
	switch event.Type {
	case "message_start":
		if event.Message.ID != "" {
			return &services.ClaudeStreamEvent{
				Type: event.Type,
				Message: &services.ClaudeResponse{
					ID:    event.Message.ID,
					Model: string(event.Message.Model),
					Role:  vo.RoleAssistant,
				},
			}
		}

	case "content_block_start":
		block := event.ContentBlock
		switch block.Type {
		case "text":
			return &services.ClaudeStreamEvent{
				Type:  event.Type,
				Index: int(event.Index),
				ContentBlock: &entities.ContentBlock{
					Type: vo.ContentTypeText,
					Text: block.Text,
				},
			}
		case "tool_use":
			return &services.ClaudeStreamEvent{
				Type:  event.Type,
				Index: int(event.Index),
				ContentBlock: &entities.ContentBlock{
					Type: vo.ContentTypeToolUse,
					ID:   block.ID,
					Name: block.Name,
				},
			}
		}

	case "content_block_delta":
		delta := event.Delta
		return &services.ClaudeStreamEvent{
			Type:  event.Type,
			Index: int(event.Index),
			Delta: &services.ClaudeDelta{
				Type: delta.Type,
				Text: delta.Text,
			},
		}

	case "message_delta":
		return &services.ClaudeStreamEvent{
			Type: event.Type,
			Delta: &services.ClaudeDelta{
				StopReason: string(event.Delta.StopReason),
			},
			Usage: &services.ClaudeUsage{
				OutputTokens: int(event.Usage.OutputTokens),
			},
		}

	case "message_stop":
		return &services.ClaudeStreamEvent{
			Type: event.Type,
		}
	}

	return nil
}

// isRetryableError checks if an error is retryable
func (c *Client) isRetryableError(err error) bool {
	// Check for rate limiting or temporary errors
	// This would need to inspect the actual error type from the SDK
	return false
}

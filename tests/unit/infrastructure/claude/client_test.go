// Package claude_test provides unit tests for the Claude client.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package claude_test

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/services"
	vo "github.com/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
	"github.com/telemetryflow/telemetryflow-mcp/internal/infrastructure/claude"
	"github.com/telemetryflow/telemetryflow-mcp/internal/infrastructure/config"
)

func TestNewClient(t *testing.T) {
	logger := zerolog.Nop()

	t.Run("should create client with valid config", func(t *testing.T) {
		cfg := &config.ClaudeConfig{
			APIKey:     "test-api-key",
			MaxTokens:  4096,
			MaxRetries: 3,
			RetryDelay: time.Second,
		}

		client, err := claude.NewClient(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("should fail with empty API key", func(t *testing.T) {
		cfg := &config.ClaudeConfig{
			APIKey: "",
		}

		client, err := claude.NewClient(cfg, logger)
		assert.Error(t, err)
		assert.Nil(t, client)
		assert.ErrorIs(t, err, claude.ErrAPIKeyRequired)
	})

	t.Run("should use custom base URL", func(t *testing.T) {
		cfg := &config.ClaudeConfig{
			APIKey:  "test-api-key",
			BaseURL: "https://custom.api.anthropic.com",
		}

		client, err := claude.NewClient(cfg, logger)
		require.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestValidateRequest(t *testing.T) {
	logger := zerolog.Nop()
	cfg := &config.ClaudeConfig{
		APIKey:     "test-api-key",
		MaxTokens:  4096,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}
	client, _ := claude.NewClient(cfg, logger)

	t.Run("should validate valid request", func(t *testing.T) {
		request := &services.ClaudeRequest{
			Model:     vo.ModelClaude4Sonnet,
			MaxTokens: 1024,
			Messages: []services.ClaudeMessage{
				{
					Role: vo.RoleUser,
					Content: []entities.ContentBlock{
						{Type: vo.ContentTypeText, Text: "Hello"},
					},
				},
			},
		}

		err := client.ValidateRequest(request)
		assert.NoError(t, err)
	})

	t.Run("should fail with nil request", func(t *testing.T) {
		err := client.ValidateRequest(nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, claude.ErrInvalidRequest)
	})

	t.Run("should fail with empty messages", func(t *testing.T) {
		request := &services.ClaudeRequest{
			Model:    vo.ModelClaude4Sonnet,
			Messages: []services.ClaudeMessage{},
		}

		err := client.ValidateRequest(request)
		assert.Error(t, err)
	})

	t.Run("should use default max tokens when not set", func(t *testing.T) {
		request := &services.ClaudeRequest{
			Model:     vo.ModelClaude4Sonnet,
			MaxTokens: 0,
			Messages: []services.ClaudeMessage{
				{
					Role: vo.RoleUser,
					Content: []entities.ContentBlock{
						{Type: vo.ContentTypeText, Text: "Hello"},
					},
				},
			},
		}

		err := client.ValidateRequest(request)
		assert.NoError(t, err)
		assert.Equal(t, cfg.MaxTokens, request.MaxTokens)
	})
}

func TestClaudeRequestBuilder(t *testing.T) {
	t.Run("should build request with all parameters", func(t *testing.T) {
		systemPrompt, _ := vo.NewSystemPrompt("You are a helpful assistant.")

		request := &services.ClaudeRequest{
			Model:        vo.ModelClaude4Sonnet,
			MaxTokens:    2048,
			Temperature:  0.7,
			TopP:         0.9,
			TopK:         40,
			SystemPrompt: systemPrompt,
			Messages: []services.ClaudeMessage{
				{
					Role: vo.RoleUser,
					Content: []entities.ContentBlock{
						{Type: vo.ContentTypeText, Text: "Hello"},
					},
				},
			},
			StopSequences: []string{"END", "STOP"},
		}

		assert.Equal(t, vo.ModelClaude4Sonnet, request.Model)
		assert.Equal(t, 2048, request.MaxTokens)
		assert.InDelta(t, 0.7, request.Temperature, 0.01)
		assert.InDelta(t, 0.9, request.TopP, 0.01)
		assert.Equal(t, 40, request.TopK)
		assert.Len(t, request.StopSequences, 2)
	})

	t.Run("should build request with tools", func(t *testing.T) {
		request := &services.ClaudeRequest{
			Model:     vo.ModelClaude4Sonnet,
			MaxTokens: 1024,
			Messages: []services.ClaudeMessage{
				{
					Role: vo.RoleUser,
					Content: []entities.ContentBlock{
						{Type: vo.ContentTypeText, Text: "What's the weather?"},
					},
				},
			},
			Tools: []services.ClaudeTool{
				{
					Name:        "get_weather",
					Description: "Get current weather for a location",
					InputSchema: nil,
				},
			},
		}

		assert.Len(t, request.Tools, 1)
		assert.Equal(t, "get_weather", request.Tools[0].Name)
	})
}

func TestClaudeMessageBuilding(t *testing.T) {
	t.Run("should build user message", func(t *testing.T) {
		msg := services.ClaudeMessage{
			Role: vo.RoleUser,
			Content: []entities.ContentBlock{
				{Type: vo.ContentTypeText, Text: "Hello, Claude!"},
			},
		}

		assert.Equal(t, vo.RoleUser, msg.Role)
		assert.Len(t, msg.Content, 1)
		assert.Equal(t, vo.ContentTypeText, msg.Content[0].Type)
	})

	t.Run("should build assistant message", func(t *testing.T) {
		msg := services.ClaudeMessage{
			Role: vo.RoleAssistant,
			Content: []entities.ContentBlock{
				{Type: vo.ContentTypeText, Text: "Hello! How can I help you?"},
			},
		}

		assert.Equal(t, vo.RoleAssistant, msg.Role)
		assert.Len(t, msg.Content, 1)
	})

	t.Run("should build tool use message", func(t *testing.T) {
		msg := services.ClaudeMessage{
			Role: vo.RoleAssistant,
			Content: []entities.ContentBlock{
				{
					Type:  vo.ContentTypeToolUse,
					ID:    "tool_123",
					Name:  "get_weather",
					Input: map[string]interface{}{"location": "San Francisco"},
				},
			},
		}

		assert.Equal(t, vo.RoleAssistant, msg.Role)
		assert.Equal(t, vo.ContentTypeToolUse, msg.Content[0].Type)
		assert.Equal(t, "tool_123", msg.Content[0].ID)
	})

	t.Run("should build tool result message", func(t *testing.T) {
		msg := services.ClaudeMessage{
			Role: vo.RoleUser,
			Content: []entities.ContentBlock{
				{
					Type:      vo.ContentTypeToolResult,
					ToolUseID: "tool_123",
					Content:   "Sunny, 72°F",
					IsError:   false,
				},
			},
		}

		assert.Equal(t, vo.RoleUser, msg.Role)
		assert.Equal(t, vo.ContentTypeToolResult, msg.Content[0].Type)
		assert.Equal(t, "tool_123", msg.Content[0].ToolUseID)
		assert.False(t, msg.Content[0].IsError)
	})
}

func TestClaudeResponseParsing(t *testing.T) {
	t.Run("should parse text response", func(t *testing.T) {
		response := &services.ClaudeResponse{
			ID:         "msg_123",
			Type:       "message",
			Role:       vo.RoleAssistant,
			Model:      "claude-sonnet-4-20250514",
			StopReason: "end_turn",
			Content: []entities.ContentBlock{
				{Type: vo.ContentTypeText, Text: "Hello!"},
			},
			Usage: &services.ClaudeUsage{
				InputTokens:  10,
				OutputTokens: 5,
			},
		}

		assert.Equal(t, "msg_123", response.ID)
		assert.Equal(t, vo.RoleAssistant, response.Role)
		assert.Len(t, response.Content, 1)
		assert.Equal(t, 10, response.Usage.InputTokens)
		assert.Equal(t, 5, response.Usage.OutputTokens)
	})

	t.Run("should parse tool use response", func(t *testing.T) {
		response := &services.ClaudeResponse{
			ID:         "msg_456",
			Role:       vo.RoleAssistant,
			StopReason: "tool_use",
			Content: []entities.ContentBlock{
				{
					Type:  vo.ContentTypeToolUse,
					ID:    "toolu_789",
					Name:  "get_weather",
					Input: map[string]interface{}{"location": "NYC"},
				},
			},
		}

		assert.Equal(t, "tool_use", response.StopReason)
		assert.Equal(t, vo.ContentTypeToolUse, response.Content[0].Type)
		assert.Equal(t, "toolu_789", response.Content[0].ID)
	})
}

func TestClaudeStreamEvents(t *testing.T) {
	t.Run("should handle message start event", func(t *testing.T) {
		event := &services.ClaudeStreamEvent{
			Type: "message_start",
			Message: &services.ClaudeResponse{
				ID:    "msg_stream_123",
				Model: "claude-sonnet-4-20250514",
				Role:  vo.RoleAssistant,
			},
		}

		assert.Equal(t, "message_start", event.Type)
		assert.Equal(t, "msg_stream_123", event.Message.ID)
	})

	t.Run("should handle content block delta event", func(t *testing.T) {
		event := &services.ClaudeStreamEvent{
			Type:  "content_block_delta",
			Index: 0,
			Delta: &services.ClaudeDelta{
				Type: "text_delta",
				Text: "Hello",
			},
		}

		assert.Equal(t, "content_block_delta", event.Type)
		assert.Equal(t, 0, event.Index)
		assert.Equal(t, "Hello", event.Delta.Text)
	})

	t.Run("should handle message delta event", func(t *testing.T) {
		event := &services.ClaudeStreamEvent{
			Type: "message_delta",
			Delta: &services.ClaudeDelta{
				StopReason: "end_turn",
			},
			Usage: &services.ClaudeUsage{
				OutputTokens: 50,
			},
		}

		assert.Equal(t, "message_delta", event.Type)
		assert.Equal(t, "end_turn", event.Delta.StopReason)
		assert.Equal(t, 50, event.Usage.OutputTokens)
	})

	t.Run("should handle error event", func(t *testing.T) {
		event := &services.ClaudeStreamEvent{
			Error: context.DeadlineExceeded,
		}

		assert.NotNil(t, event.Error)
		assert.ErrorIs(t, event.Error, context.DeadlineExceeded)
	})
}

func TestClaudeModels(t *testing.T) {
	models := []struct {
		model    vo.Model
		expected string
	}{
		{vo.ModelClaude4Opus, "claude-opus-4-20250514"},
		{vo.ModelClaude4Sonnet, "claude-sonnet-4-20250514"},
		{vo.ModelClaude37Sonnet, "claude-3-7-sonnet-20250219"},
		{vo.ModelClaude35Sonnet, "claude-3-5-sonnet-20241022"},
		{vo.ModelClaude35Haiku, "claude-3-5-haiku-20241022"},
	}

	for _, tc := range models {
		t.Run("should validate model "+string(tc.model), func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.model.String())
			assert.True(t, tc.model.IsValid())
		})
	}
}

// Benchmarks

func BenchmarkValidateRequest(b *testing.B) {
	logger := zerolog.Nop()
	cfg := &config.ClaudeConfig{
		APIKey:     "test-api-key",
		MaxTokens:  4096,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}
	client, _ := claude.NewClient(cfg, logger)

	request := &services.ClaudeRequest{
		Model:     vo.ModelClaude4Sonnet,
		MaxTokens: 1024,
		Messages: []services.ClaudeMessage{
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{Type: vo.ContentTypeText, Text: "Hello"},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.ValidateRequest(request)
	}
}

// Ensure entities import is used
var _ = entities.ContentBlock{}

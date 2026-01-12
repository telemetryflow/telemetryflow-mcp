// Package mocks provides mock implementations for testing.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/services"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// MockClaudeService is a mock implementation of the Claude service
type MockClaudeService struct {
	mock.Mock
}

// NewMockClaudeService creates a new mock Claude service
func NewMockClaudeService() *MockClaudeService {
	return &MockClaudeService{}
}

// CreateMessage mocks the CreateMessage method
func (m *MockClaudeService) CreateMessage(ctx context.Context, request *services.ClaudeRequest) (*services.ClaudeResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.ClaudeResponse), args.Error(1)
}

// CreateMessageStream mocks the CreateMessageStream method
func (m *MockClaudeService) CreateMessageStream(ctx context.Context, request *services.ClaudeRequest) (<-chan *services.ClaudeStreamEvent, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(<-chan *services.ClaudeStreamEvent), args.Error(1)
}

// CountTokens mocks the CountTokens method
func (m *MockClaudeService) CountTokens(ctx context.Context, request *services.ClaudeRequest) (int, error) {
	args := m.Called(ctx, request)
	return args.Int(0), args.Error(1)
}

// ValidateRequest mocks the ValidateRequest method
func (m *MockClaudeService) ValidateRequest(request *services.ClaudeRequest) error {
	args := m.Called(request)
	return args.Error(0)
}

// MockClaudeResponse creates a mock Claude response
func MockClaudeResponse(text string) *services.ClaudeResponse {
	return &services.ClaudeResponse{
		ID:         "msg_mock_123",
		Type:       "message",
		Role:       vo.RoleAssistant,
		Model:      "claude-sonnet-4-20250514",
		StopReason: "end_turn",
		Content: []entities.ContentBlock{
			{Type: vo.ContentTypeText, Text: text},
		},
		Usage: &services.ClaudeUsage{
			InputTokens:  100,
			OutputTokens: 50,
		},
	}
}

// MockClaudeToolUseResponse creates a mock Claude response with tool use
func MockClaudeToolUseResponse(toolName, toolID string, input map[string]interface{}) *services.ClaudeResponse {
	return &services.ClaudeResponse{
		ID:         "msg_mock_tool_456",
		Type:       "message",
		Role:       vo.RoleAssistant,
		Model:      "claude-sonnet-4-20250514",
		StopReason: "tool_use",
		Content: []entities.ContentBlock{
			{
				Type:  vo.ContentTypeToolUse,
				ID:    toolID,
				Name:  toolName,
				Input: input,
			},
		},
		Usage: &services.ClaudeUsage{
			InputTokens:  150,
			OutputTokens: 30,
		},
	}
}

// MockClaudeStreamEvents returns a channel of mock streaming events
func MockClaudeStreamEvents(text string) <-chan *services.ClaudeStreamEvent {
	ch := make(chan *services.ClaudeStreamEvent, 10)

	go func() {
		defer close(ch)

		// Message start
		ch <- &services.ClaudeStreamEvent{
			Type: "message_start",
			Message: &services.ClaudeResponse{
				ID:    "msg_stream_mock",
				Model: "claude-sonnet-4-20250514",
				Role:  vo.RoleAssistant,
			},
		}

		// Content block start
		ch <- &services.ClaudeStreamEvent{
			Type:  "content_block_start",
			Index: 0,
			ContentBlock: &entities.ContentBlock{
				Type: vo.ContentTypeText,
			},
		}

		// Stream text in chunks
		for i := 0; i < len(text); i += 10 {
			end := i + 10
			if end > len(text) {
				end = len(text)
			}
			ch <- &services.ClaudeStreamEvent{
				Type:  "content_block_delta",
				Index: 0,
				Delta: &services.ClaudeDelta{
					Type: "text_delta",
					Text: text[i:end],
				},
			}
		}

		// Content block stop
		ch <- &services.ClaudeStreamEvent{
			Type:  "content_block_stop",
			Index: 0,
		}

		// Message delta with stop reason
		ch <- &services.ClaudeStreamEvent{
			Type: "message_delta",
			Delta: &services.ClaudeDelta{
				StopReason: "end_turn",
			},
			Usage: &services.ClaudeUsage{
				OutputTokens: len(text) / 4, // Rough token estimate
			},
		}

		// Message stop
		ch <- &services.ClaudeStreamEvent{
			Type: "message_stop",
		}
	}()

	return ch
}

// MockClaudeRequest creates a mock Claude request
func MockClaudeRequest(message string) *services.ClaudeRequest {
	return &services.ClaudeRequest{
		Model:     vo.ModelClaude4Sonnet,
		MaxTokens: 4096,
		Messages: []services.ClaudeMessage{
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{Type: vo.ContentTypeText, Text: message},
				},
			},
		},
	}
}

// MockClaudeRequestWithSystem creates a mock Claude request with system prompt
func MockClaudeRequestWithSystem(systemPrompt, message string) *services.ClaudeRequest {
	sp, _ := vo.NewSystemPrompt(systemPrompt)
	return &services.ClaudeRequest{
		Model:        vo.ModelClaude4Sonnet,
		MaxTokens:    4096,
		SystemPrompt: sp,
		Messages: []services.ClaudeMessage{
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{Type: vo.ContentTypeText, Text: message},
				},
			},
		},
	}
}

// MockClaudeRequestWithTools creates a mock Claude request with tools
func MockClaudeRequestWithTools(message string, tools []services.ClaudeTool) *services.ClaudeRequest {
	return &services.ClaudeRequest{
		Model:     vo.ModelClaude4Sonnet,
		MaxTokens: 4096,
		Messages: []services.ClaudeMessage{
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{Type: vo.ContentTypeText, Text: message},
				},
			},
		},
		Tools: tools,
	}
}

// MockClaudeTool creates a mock Claude tool definition
func MockClaudeTool(name, description string) services.ClaudeTool {
	return services.ClaudeTool{
		Name:        name,
		Description: description,
		InputSchema: &entities.JSONSchema{
			Type: "object",
			Properties: map[string]*entities.JSONSchema{
				"input": {
					Type:        "string",
					Description: "Input parameter",
				},
			},
		},
	}
}

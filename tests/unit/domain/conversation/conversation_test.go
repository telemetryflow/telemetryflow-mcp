// Package conversation_test provides unit tests for the conversation aggregate.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package conversation_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestNewConversation(t *testing.T) {
	t.Run("should create conversation with unique ID", func(t *testing.T) {
		sessionID := vo.GenerateSessionID()
		conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
		require.NotNil(t, conv)
		assert.NotEmpty(t, conv.ID().String())
	})

	t.Run("should create conversation with specified session ID", func(t *testing.T) {
		sessionID := vo.GenerateSessionID()
		conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
		assert.Equal(t, sessionID, conv.SessionID())
	})

	t.Run("should create conversation with specified model", func(t *testing.T) {
		sessionID := vo.GenerateSessionID()
		conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Opus)
		assert.Equal(t, vo.ModelClaude4Opus, conv.Model())
	})

	t.Run("should create conversation in active status", func(t *testing.T) {
		sessionID := vo.GenerateSessionID()
		conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
		assert.Equal(t, aggregates.ConversationStatusActive, conv.Status())
	})

	t.Run("should have empty message list initially", func(t *testing.T) {
		sessionID := vo.GenerateSessionID()
		conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
		assert.Empty(t, conv.Messages())
	})

	t.Run("should generate unique IDs for different conversations", func(t *testing.T) {
		sessionID := vo.GenerateSessionID()
		conv1 := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
		conv2 := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
		assert.NotEqual(t, conv1.ID().String(), conv2.ID().String())
	})
}

func TestConversationMessages(t *testing.T) {
	t.Run("should add user message", func(t *testing.T) {
		conv := createTestConversation(t)
		msg := createTextMessage(t, vo.RoleUser, "Hello, Claude!")

		err := conv.AddMessage(msg)
		require.NoError(t, err)
		assert.Len(t, conv.Messages(), 1)
	})

	t.Run("should add assistant message", func(t *testing.T) {
		conv := createTestConversation(t)
		// First add user message
		userMsg := createTextMessage(t, vo.RoleUser, "Hello!")
		err := conv.AddMessage(userMsg)
		require.NoError(t, err)

		// Then add assistant message
		msg := createTextMessage(t, vo.RoleAssistant, "Hello! How can I help you?")
		err = conv.AddMessage(msg)
		require.NoError(t, err)
		assert.Len(t, conv.Messages(), 2)
	})

	t.Run("should add multiple messages in order", func(t *testing.T) {
		conv := createTestConversation(t)

		err := conv.AddMessage(createTextMessage(t, vo.RoleUser, "First"))
		require.NoError(t, err)
		err = conv.AddMessage(createTextMessage(t, vo.RoleAssistant, "Second"))
		require.NoError(t, err)
		err = conv.AddMessage(createTextMessage(t, vo.RoleUser, "Third"))
		require.NoError(t, err)

		messages := conv.Messages()
		assert.Len(t, messages, 3)
		assert.Equal(t, vo.RoleUser, messages[0].Role())
		assert.Equal(t, vo.RoleAssistant, messages[1].Role())
		assert.Equal(t, vo.RoleUser, messages[2].Role())
	})

	t.Run("should not add message to closed conversation", func(t *testing.T) {
		conv := createTestConversation(t)
		conv.Close()

		err := conv.AddMessage(createTextMessage(t, vo.RoleUser, "Should fail"))
		assert.Error(t, err)
	})

	t.Run("should track message count", func(t *testing.T) {
		conv := createTestConversation(t)

		for i := 0; i < 10; i++ {
			if i%2 == 0 {
				_ = conv.AddMessage(createTextMessage(t, vo.RoleUser, "Message"))
			} else {
				_ = conv.AddMessage(createTextMessage(t, vo.RoleAssistant, "Response"))
			}
		}

		assert.Equal(t, 10, conv.MessageCount())
	})

	t.Run("should get last message", func(t *testing.T) {
		conv := createTestConversation(t)

		_ = conv.AddMessage(createTextMessage(t, vo.RoleUser, "First"))
		_ = conv.AddMessage(createTextMessage(t, vo.RoleAssistant, "Last"))

		lastMsg := conv.LastMessage()
		require.NotNil(t, lastMsg)
		assert.Equal(t, vo.RoleAssistant, lastMsg.Role())
	})

	t.Run("should return nil for last message on empty conversation", func(t *testing.T) {
		conv := createTestConversation(t)
		lastMsg := conv.LastMessage()
		assert.Nil(t, lastMsg)
	})
}

func TestConversationSystemPrompt(t *testing.T) {
	t.Run("should set system prompt", func(t *testing.T) {
		conv := createTestConversation(t)
		systemPrompt, err := vo.NewSystemPrompt("You are a helpful assistant.")
		require.NoError(t, err)

		err = conv.SetSystemPrompt(systemPrompt)
		require.NoError(t, err)
		assert.Equal(t, systemPrompt, conv.SystemPrompt())
	})

	t.Run("should update system prompt before user messages", func(t *testing.T) {
		conv := createTestConversation(t)

		firstPrompt, _ := vo.NewSystemPrompt("First prompt")
		_ = conv.SetSystemPrompt(firstPrompt)

		updatedPrompt, _ := vo.NewSystemPrompt("Updated prompt")
		err := conv.SetSystemPrompt(updatedPrompt)
		require.NoError(t, err)
		assert.Equal(t, "Updated prompt", conv.SystemPrompt().String())
	})

	t.Run("should not set system prompt after user message", func(t *testing.T) {
		conv := createTestConversation(t)
		_ = conv.AddMessage(createTextMessage(t, vo.RoleUser, "Hello"))

		failPrompt, _ := vo.NewSystemPrompt("Should fail")
		err := conv.SetSystemPrompt(failPrompt)
		assert.Error(t, err)
	})
}

func TestConversationSettings(t *testing.T) {
	t.Run("should have default max tokens", func(t *testing.T) {
		conv := createTestConversation(t)
		assert.Greater(t, conv.MaxTokens(), 0)
	})

	t.Run("should set max tokens", func(t *testing.T) {
		conv := createTestConversation(t)
		conv.SetMaxTokens(2048)
		assert.Equal(t, 2048, conv.MaxTokens())
	})

	t.Run("should have default temperature", func(t *testing.T) {
		conv := createTestConversation(t)
		assert.InDelta(t, 1.0, conv.Temperature(), 0.01)
	})

	t.Run("should set temperature", func(t *testing.T) {
		conv := createTestConversation(t)
		conv.SetTemperature(0.7)
		assert.InDelta(t, 0.7, conv.Temperature(), 0.01)
	})

	t.Run("should clamp temperature to valid range", func(t *testing.T) {
		conv := createTestConversation(t)
		conv.SetTemperature(2.5)
		assert.InDelta(t, 2.0, conv.Temperature(), 0.01)
	})

	t.Run("should set top P", func(t *testing.T) {
		conv := createTestConversation(t)
		conv.SetTopP(0.9)
		assert.InDelta(t, 0.9, conv.TopP(), 0.01)
	})

	t.Run("should set top K", func(t *testing.T) {
		conv := createTestConversation(t)
		conv.SetTopK(40)
		assert.Equal(t, 40, conv.TopK())
	})
}

func TestConversationClose(t *testing.T) {
	t.Run("should close active conversation", func(t *testing.T) {
		conv := createTestConversation(t)

		conv.Close()
		assert.Equal(t, aggregates.ConversationStatusClosed, conv.Status())
	})

	t.Run("should set closed time", func(t *testing.T) {
		conv := createTestConversation(t)
		beforeClose := time.Now()

		conv.Close()

		closedAt := conv.ClosedAt()
		require.NotNil(t, closedAt)
		assert.True(t, closedAt.After(beforeClose) || closedAt.Equal(beforeClose))
	})

	t.Run("should not change status on double close", func(t *testing.T) {
		conv := createTestConversation(t)

		conv.Close()
		firstClosedAt := conv.ClosedAt()

		conv.Close()
		// Should remain closed with same timestamp
		assert.Equal(t, aggregates.ConversationStatusClosed, conv.Status())
		assert.Equal(t, firstClosedAt, conv.ClosedAt())
	})

	t.Run("should preserve messages after close", func(t *testing.T) {
		conv := createTestConversation(t)
		_ = conv.AddMessage(createTextMessage(t, vo.RoleUser, "Test message"))

		conv.Close()

		assert.Len(t, conv.Messages(), 1)
	})
}

func TestConversationTools(t *testing.T) {
	t.Run("should add tool", func(t *testing.T) {
		conv := createTestConversation(t)
		tool := createTestTool(t, "test_tool", "Test description")

		conv.AddTool(tool)
		assert.Len(t, conv.Tools(), 1)
	})

	t.Run("should get tool by name", func(t *testing.T) {
		conv := createTestConversation(t)
		tool := createTestTool(t, "my_tool", "My description")

		conv.AddTool(tool)

		toolName, _ := vo.NewToolName("my_tool")
		retrievedTool := conv.GetTool(toolName)
		assert.NotNil(t, retrievedTool)
	})

	t.Run("should return nil for non-existent tool", func(t *testing.T) {
		conv := createTestConversation(t)
		toolName, _ := vo.NewToolName("non_existent")
		tool := conv.GetTool(toolName)
		assert.Nil(t, tool)
	})

	t.Run("should remove tool", func(t *testing.T) {
		conv := createTestConversation(t)
		tool := createTestTool(t, "removable_tool", "To remove")

		conv.AddTool(tool)
		assert.Len(t, conv.Tools(), 1)

		toolName, _ := vo.NewToolName("removable_tool")
		conv.RemoveTool(toolName)
		assert.Empty(t, conv.Tools())
	})
}

func TestConversationMetadata(t *testing.T) {
	t.Run("should set metadata", func(t *testing.T) {
		conv := createTestConversation(t)
		conv.SetMetadata("key1", "value1")
		conv.SetMetadata("key2", 123)

		metadata := conv.Metadata()
		assert.Equal(t, "value1", metadata["key1"])
		assert.Equal(t, 123, metadata["key2"])
	})

	t.Run("should get metadata value", func(t *testing.T) {
		conv := createTestConversation(t)
		conv.SetMetadata("test_key", "test_value")

		value, ok := conv.GetMetadata("test_key")
		assert.True(t, ok)
		assert.Equal(t, "test_value", value)
	})

	t.Run("should return false for non-existent metadata key", func(t *testing.T) {
		conv := createTestConversation(t)
		_, ok := conv.GetMetadata("non_existent")
		assert.False(t, ok)
	})
}

func TestConversationCreatedAt(t *testing.T) {
	t.Run("should set created time", func(t *testing.T) {
		beforeCreate := time.Now()
		conv := createTestConversation(t)
		afterCreate := time.Now()

		createdAt := conv.CreatedAt()
		assert.True(t, createdAt.After(beforeCreate) || createdAt.Equal(beforeCreate))
		assert.True(t, createdAt.Before(afterCreate) || createdAt.Equal(afterCreate))
	})
}

func TestConversationModels(t *testing.T) {
	models := []vo.Model{
		vo.ModelClaude4Opus,
		vo.ModelClaude4Sonnet,
		vo.ModelClaude37Sonnet,
		vo.ModelClaude35Sonnet,
		vo.ModelClaude35Haiku,
	}

	for _, model := range models {
		t.Run("should create conversation with model "+string(model), func(t *testing.T) {
			sessionID := vo.GenerateSessionID()
			conv := aggregates.NewConversation(sessionID, model)
			assert.Equal(t, model, conv.Model())
		})
	}
}

func TestConversationMessageContent(t *testing.T) {
	t.Run("should get text content from message", func(t *testing.T) {
		conv := createTestConversation(t)

		msg := createTextMessage(t, vo.RoleUser, "Hello, Claude!")
		err := conv.AddMessage(msg)
		require.NoError(t, err)

		messages := conv.Messages()
		require.Len(t, messages, 1)
		content := messages[0].Content()
		require.Len(t, content, 1)
		assert.Equal(t, "Hello, Claude!", content[0].Text)
	})
}

// Helper functions

func createTestConversation(t *testing.T) *aggregates.Conversation {
	t.Helper()
	sessionID := vo.GenerateSessionID()
	return aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
}

func createTextMessage(t *testing.T, role vo.Role, text string) *entities.Message {
	t.Helper()
	msg, err := entities.NewTextMessage(role, text)
	require.NoError(t, err)
	return msg
}

func createTestTool(t *testing.T, name, description string) *entities.Tool {
	t.Helper()
	toolName, err := vo.NewToolName(name)
	require.NoError(t, err)
	toolDesc, err := vo.NewToolDescription(description)
	require.NoError(t, err)
	tool, err := entities.NewTool(toolName, toolDesc, nil)
	require.NoError(t, err)
	return tool
}

// Benchmarks

func BenchmarkNewConversation(b *testing.B) {
	sessionID := vo.GenerateSessionID()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
	}
}

func BenchmarkConversationAddMessage(b *testing.B) {
	sessionID := vo.GenerateSessionID()
	conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)
	msg, _ := entities.NewTextMessage(vo.RoleUser, "Benchmark message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conv.AddMessage(msg)
	}
}

func BenchmarkConversationGetMessages(b *testing.B) {
	sessionID := vo.GenerateSessionID()
	conv := aggregates.NewConversation(sessionID, vo.ModelClaude4Sonnet)

	// Add 100 messages alternating roles
	for i := 0; i < 100; i++ {
		var msg *entities.Message
		if i%2 == 0 {
			msg, _ = entities.NewTextMessage(vo.RoleUser, "Message")
		} else {
			msg, _ = entities.NewTextMessage(vo.RoleAssistant, "Response")
		}
		_ = conv.AddMessage(msg)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conv.Messages()
	}
}

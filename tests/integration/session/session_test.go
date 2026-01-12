// Package session_test provides integration tests for session management.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
)

func TestSessionLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("complete session lifecycle", func(t *testing.T) {
		// Create session
		session := aggregates.NewSession()
		require.NotNil(t, session)
		assert.Equal(t, aggregates.SessionStateCreated, session.State())

		// Initialize session
		clientInfo := &aggregates.ClientInfo{
			Name:    "IntegrationTestClient",
			Version: "1.0.0",
		}
		err := session.Initialize(clientInfo, "2024-11-05")
		require.NoError(t, err)
		assert.Equal(t, aggregates.SessionStateInitializing, session.State())

		// Mark ready
		session.MarkReady()
		assert.Equal(t, aggregates.SessionStateReady, session.State())

		// Register tools
		tool := createIntegrationTool(t, "test_tool", "Integration test tool")
		session.RegisterTool(tool)
		assert.Len(t, session.ListTools(), 1)

		// Register resources
		uri, _ := vo.NewResourceURI("file:///integration/test")
		resource, _ := entities.NewResource(uri, "Integration Resource")
		session.RegisterResource(resource)
		assert.Len(t, session.ListResources(), 1)

		// Create conversation
		conv, err := session.CreateConversation(vo.ModelClaude4Sonnet)
		require.NoError(t, err)
		assert.Len(t, session.ListConversations(), 1)

		// Close conversation
		conv.Close()
		assert.Equal(t, aggregates.ConversationStatusClosed, conv.Status())

		// Close session
		session.Close()
		assert.Equal(t, aggregates.SessionStateClosed, session.State())
	})

	t.Run("session with multiple conversations", func(t *testing.T) {
		session := createReadySession(t)

		// Create multiple conversations
		models := []vo.Model{
			vo.ModelClaude4Opus,
			vo.ModelClaude4Sonnet,
			vo.ModelClaude35Sonnet,
		}

		for _, model := range models {
			conv, err := session.CreateConversation(model)
			require.NoError(t, err)
			require.NotNil(t, conv)
		}

		assert.Len(t, session.ListConversations(), 3)

		// Close all conversations
		for _, conv := range session.ListConversations() {
			conv.Close()
		}

		// Close session
		session.Close()
	})

	t.Run("session with tool registration and unregistration", func(t *testing.T) {
		session := createReadySession(t)

		// Register multiple tools
		tools := []string{"read_file", "write_file", "execute_command", "echo"}
		for _, name := range tools {
			tool := createIntegrationTool(t, name, "Tool: "+name)
			session.RegisterTool(tool)
		}
		assert.Len(t, session.ListTools(), 4)

		// Unregister a tool
		session.UnregisterTool("echo")
		assert.Len(t, session.ListTools(), 3)

		// Verify tool was removed
		_, ok := session.GetTool("echo")
		assert.False(t, ok)
		_, ok = session.GetTool("read_file")
		assert.True(t, ok)
	})
}

func TestConversationLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("complete conversation lifecycle", func(t *testing.T) {
		session := createReadySession(t)
		conv, err := session.CreateConversation(vo.ModelClaude4Sonnet)
		require.NoError(t, err)

		// Set system prompt
		systemPrompt, err := vo.NewSystemPrompt("You are a helpful assistant for integration testing.")
		require.NoError(t, err)
		err = conv.SetSystemPrompt(systemPrompt)
		require.NoError(t, err)

		// Add user message
		userMsg, err := entities.NewTextMessage(vo.RoleUser, "Hello, this is an integration test.")
		require.NoError(t, err)
		err = conv.AddMessage(userMsg)
		require.NoError(t, err)

		// Add assistant response
		assistantMsg, err := entities.NewTextMessage(vo.RoleAssistant, "Hello! I'm ready to help with testing.")
		require.NoError(t, err)
		err = conv.AddMessage(assistantMsg)
		require.NoError(t, err)

		// Verify messages
		assert.Equal(t, 2, conv.MessageCount())
		messages := conv.Messages()
		assert.Equal(t, vo.RoleUser, messages[0].Role())
		assert.Equal(t, vo.RoleAssistant, messages[1].Role())

		// Close conversation
		conv.Close()
	})

	t.Run("conversation with tool use", func(t *testing.T) {
		session := createReadySession(t)
		conv, err := session.CreateConversation(vo.ModelClaude4Sonnet)
		require.NoError(t, err)

		// Register tool in conversation
		tool := createIntegrationTool(t, "get_weather", "Get weather for a location")
		conv.AddTool(tool)

		// Add user message requesting tool use
		userMsg, err := entities.NewTextMessage(vo.RoleUser, "What's the weather in Jakarta?")
		require.NoError(t, err)
		err = conv.AddMessage(userMsg)
		require.NoError(t, err)

		// Simulate assistant response with tool use
		// In real integration, this would come from Claude API

		// Close conversation
		conv.Close()
	})

	t.Run("conversation settings persistence", func(t *testing.T) {
		session := createReadySession(t)
		conv, err := session.CreateConversation(vo.ModelClaude4Sonnet)
		require.NoError(t, err)

		// Set various settings
		conv.SetMaxTokens(2048)
		conv.SetTemperature(0.7)
		conv.SetTopP(0.95)
		conv.SetTopK(50)

		// Verify settings
		assert.Equal(t, 2048, conv.MaxTokens())
		assert.InDelta(t, 0.7, conv.Temperature(), 0.01)
		assert.InDelta(t, 0.95, conv.TopP(), 0.01)
		assert.Equal(t, 50, conv.TopK())
	})
}

func TestConcurrentSessionOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("concurrent tool registration", func(t *testing.T) {
		session := createReadySession(t)

		// Register tools concurrently
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(idx int) {
				name := "concurrent_tool_" + string(rune('0'+idx))
				tool := createIntegrationTool(t, name, "Concurrent tool")
				session.RegisterTool(tool)
				done <- true
			}(i)
		}

		// Wait for all registrations
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify some tools were registered (may have duplicates blocked)
		assert.NotEmpty(t, session.ListTools())
	})

	t.Run("concurrent message addition", func(t *testing.T) {
		session := createReadySession(t)
		conv, err := session.CreateConversation(vo.ModelClaude4Sonnet)
		require.NoError(t, err)

		// Add messages concurrently
		done := make(chan bool, 20)
		for i := 0; i < 20; i++ {
			go func(idx int) {
				role := vo.RoleUser
				if idx%2 == 1 {
					role = vo.RoleAssistant
				}
				msg, _ := entities.NewTextMessage(role, "Message "+string(rune('0'+idx%10)))
				_ = conv.AddMessage(msg)
				done <- true
			}(i)
		}

		// Wait for all additions
		for i := 0; i < 20; i++ {
			<-done
		}

		// Verify messages were added
		assert.NotEmpty(t, conv.Messages())
	})
}

func TestSessionTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("session operations with context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		session := createReadySession(t)

		// Operations should complete before timeout
		tool := createIntegrationTool(t, "timeout_tool", "Timeout test tool")
		session.RegisterTool(tool)

		// Simulate some work
		select {
		case <-ctx.Done():
			t.Log("Context timed out as expected for long operations")
		case <-time.After(50 * time.Millisecond):
			t.Log("Operations completed within timeout")
		}
	})
}

func TestSessionCapabilitiesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("verify all capabilities are set", func(t *testing.T) {
		session := aggregates.NewSession()
		caps := session.Capabilities()

		// Tools capability
		assert.True(t, caps.Tools.ListChanged)

		// Resources capability
		assert.True(t, caps.Resources.Subscribe)
		assert.True(t, caps.Resources.ListChanged)

		// Prompts capability
		assert.True(t, caps.Prompts.ListChanged)

		// Logging capability should exist
		assert.NotNil(t, caps.Logging)
	})
}

// Helper functions

func createReadySession(t *testing.T) *aggregates.Session {
	t.Helper()
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{
		Name:    "IntegrationTest",
		Version: "1.0.0",
	}
	err := session.Initialize(clientInfo, "2024-11-05")
	require.NoError(t, err)
	session.MarkReady()
	return session
}

func createIntegrationTool(t *testing.T, name, description string) *entities.Tool {
	t.Helper()
	toolName, err := vo.NewToolName(name)
	require.NoError(t, err)
	toolDesc, err := vo.NewToolDescription(description)
	require.NoError(t, err)
	tool, err := entities.NewTool(toolName, toolDesc, nil)
	require.NoError(t, err)

	// Set a simple handler
	tool.SetHandler(func(input map[string]interface{}) (*entities.ToolResult, error) {
		return entities.NewTextToolResult("Integration test result"), nil
	})

	return tool
}

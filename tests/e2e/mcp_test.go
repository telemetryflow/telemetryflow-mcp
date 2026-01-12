// Package e2e provides end-to-end tests for the MCP server.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
)

// TestMCPServerStartup tests the server startup sequence
func TestMCPServerStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	t.Run("should create session and initialize", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Simulate MCP initialize request
		session := aggregates.NewSession()
		require.NotNil(t, session)

		clientInfo := &aggregates.ClientInfo{
			Name:    "E2E-Test-Client",
			Version: "1.0.0",
		}

		// Initialize
		err := session.Initialize(clientInfo, "2024-11-05")
		require.NoError(t, err)

		// Mark ready (simulates notifications/initialized)
		session.MarkReady()

		// Verify session is ready
		assert.Equal(t, aggregates.SessionStateReady, session.State())

		// Verify server info
		serverInfo := session.ServerInfo()
		assert.NotEmpty(t, serverInfo.Name)
		assert.NotEmpty(t, serverInfo.Version)

		// Verify capabilities
		caps := session.Capabilities()
		assert.True(t, caps.Tools.ListChanged)
		assert.True(t, caps.Resources.Subscribe)
		assert.True(t, caps.Prompts.ListChanged)

		// Simulate graceful shutdown
		select {
		case <-ctx.Done():
			t.Fatal("Test timed out")
		default:
			session.Close()
		}
	})
}

// TestMCPToolsFlow tests the complete tools workflow
func TestMCPToolsFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	t.Run("should register, list, and execute tools", func(t *testing.T) {
		session := createE2ESession(t)

		// Register built-in tools
		builtInTools := []struct {
			name        string
			description string
		}{
			{"echo", "Echo the input message"},
			{"read_file", "Read file contents"},
			{"write_file", "Write content to a file"},
			{"list_directory", "List directory contents"},
			{"execute_command", "Execute a shell command"},
			{"system_info", "Get system information"},
		}

		for _, bt := range builtInTools {
			tool := createE2ETool(t, bt.name, bt.description)
			session.RegisterTool(tool)
		}

		// List tools (simulates tools/list)
		tools := session.ListTools()
		assert.Len(t, tools, len(builtInTools))

		// Get specific tool
		echoTool, ok := session.GetTool("echo")
		require.True(t, ok)
		require.NotNil(t, echoTool)
		assert.Equal(t, "echo", echoTool.Name().String())

		// Execute tool (simulates tools/call)
		input := map[string]interface{}{"message": "Hello E2E Test"}
		result, err := echoTool.Execute(input)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsError)
	})

	t.Run("should handle tool with schema validation", func(t *testing.T) {
		session := createE2ESession(t)

		// Create tool with schema
		toolName, _ := vo.NewToolName("validated_tool")
		toolDesc, _ := vo.NewToolDescription("Tool with input validation")
		schema := &entities.JSONSchema{
			Type: "object",
			Properties: map[string]*entities.JSONSchema{
				"required_field": {
					Type:        "string",
					Description: "A required string field",
				},
				"optional_field": {
					Type:        "integer",
					Description: "An optional integer field",
				},
			},
			Required: []string{"required_field"},
		}

		tool, err := entities.NewTool(toolName, toolDesc, schema)
		require.NoError(t, err)
		tool.SetHandler(func(input map[string]interface{}) (*entities.ToolResult, error) {
			return entities.NewTextToolResult("Validated successfully"), nil
		})

		session.RegisterTool(tool)

		// Execute with valid input
		validInput := map[string]interface{}{
			"required_field": "test value",
			"optional_field": 42,
		}
		result, err := tool.Execute(validInput)
		require.NoError(t, err)
		assert.False(t, result.IsError)
	})
}

// TestMCPResourcesFlow tests the complete resources workflow
func TestMCPResourcesFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	t.Run("should register, list, and read resources", func(t *testing.T) {
		session := createE2ESession(t)

		// Register resources
		resources := []struct {
			uri      string
			name     string
			mimeType string
		}{
			{"file:///etc/config.yaml", "Configuration File", "text/yaml"},
			{"file:///var/log/app.log", "Application Log", "text/plain"},
			{"memory://cache/data", "Cached Data", "application/json"},
		}

		for _, res := range resources {
			uri, err := vo.NewResourceURI(res.uri)
			require.NoError(t, err)
			resource, err := entities.NewResource(uri, res.name)
			require.NoError(t, err)
			mimeType, err := vo.NewMimeType(res.mimeType)
			require.NoError(t, err)
			resource.SetMimeType(mimeType)
			session.RegisterResource(resource)
		}

		// List resources (simulates resources/list)
		registeredResources := session.ListResources()
		assert.Len(t, registeredResources, len(resources))

		// Get specific resource
		configResource, ok := session.GetResource("file:///etc/config.yaml")
		require.True(t, ok)
		require.NotNil(t, configResource)
	})
}

// TestMCPPromptsFlow tests the complete prompts workflow
func TestMCPPromptsFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	t.Run("should register, list, and get prompts", func(t *testing.T) {
		session := createE2ESession(t)

		// Register prompts
		prompts := []struct {
			name        string
			description string
		}{
			{"code_review", "Review code for best practices"},
			{"explain_code", "Explain what the code does"},
			{"refactor", "Suggest refactoring improvements"},
		}

		for _, p := range prompts {
			promptName, err := vo.NewToolName(p.name)
			require.NoError(t, err)
			prompt, err := entities.NewPrompt(promptName, p.description)
			require.NoError(t, err)
			session.RegisterPrompt(prompt)
		}

		// List prompts (simulates prompts/list)
		registeredPrompts := session.ListPrompts()
		assert.Len(t, registeredPrompts, len(prompts))

		// Get specific prompt
		reviewPrompt, ok := session.GetPrompt("code_review")
		require.True(t, ok)
		require.NotNil(t, reviewPrompt)
	})
}

// TestMCPConversationFlow tests the complete conversation workflow
func TestMCPConversationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	t.Run("should handle multi-turn conversation", func(t *testing.T) {
		session := createE2ESession(t)
		conv := aggregates.NewConversation(session.ID(), vo.ModelClaude4Sonnet)

		// Set system prompt
		systemPrompt, err := vo.NewSystemPrompt("You are a helpful coding assistant.")
		require.NoError(t, err)
		err = conv.SetSystemPrompt(systemPrompt)
		require.NoError(t, err)

		// Add user message
		_, err = conv.AddUserMessage("Hello, can you help me with Go programming?")
		require.NoError(t, err)

		// Add assistant response
		assistantContent := []entities.ContentBlock{{Type: vo.ContentTypeText, Text: "Of course! I'd be happy to help you with Go programming."}}
		_, err = conv.AddAssistantMessage(assistantContent)
		require.NoError(t, err)

		// Add another user message
		_, err = conv.AddUserMessage("How do I create a simple HTTP server?")
		require.NoError(t, err)

		// Verify conversation state
		assert.Equal(t, 3, conv.MessageCount())
		assert.Equal(t, aggregates.ConversationStatusActive, conv.Status())

		// Close conversation
		conv.Close()
		assert.Equal(t, aggregates.ConversationStatusClosed, conv.Status())
	})

	t.Run("should handle conversation with tool use", func(t *testing.T) {
		session := createE2ESession(t)
		conv := aggregates.NewConversation(session.ID(), vo.ModelClaude4Sonnet)

		// Add tool
		tool := createE2ETool(t, "get_file_content", "Get contents of a file")
		conv.AddTool(tool)

		// User asks for file content
		_, err := conv.AddUserMessage("Can you show me the contents of main.go?")
		require.NoError(t, err)

		// Verify tool is available
		assert.Len(t, conv.Tools(), 1)
		toolName, err := vo.NewToolName("get_file_content")
		require.NoError(t, err)
		assert.NotNil(t, conv.GetTool(toolName))
	})
}

// TestMCPJSONRPCFormat tests JSON-RPC 2.0 message formatting
func TestMCPJSONRPCFormat(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	t.Run("should handle valid JSON-RPC request format", func(t *testing.T) {
		// Simulate initialize request
		request := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  "initialize",
			"params": map[string]interface{}{
				"protocolVersion": "2024-11-05",
				"capabilities":    map[string]interface{}{},
				"clientInfo": map[string]interface{}{
					"name":    "test-client",
					"version": "1.0.0",
				},
			},
		}

		// Verify serialization
		data, err := json.Marshal(request)
		require.NoError(t, err)
		assert.Contains(t, string(data), "jsonrpc")
		assert.Contains(t, string(data), "initialize")

		// Verify deserialization
		var parsed map[string]interface{}
		err = json.Unmarshal(data, &parsed)
		require.NoError(t, err)
		assert.Equal(t, "2.0", parsed["jsonrpc"])
		assert.Equal(t, "initialize", parsed["method"])
	})

	t.Run("should format tools/list response correctly", func(t *testing.T) {
		// Simulate tools/list response
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      2,
			"result": map[string]interface{}{
				"tools": []map[string]interface{}{
					{
						"name":        "echo",
						"description": "Echo the input",
						"inputSchema": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"message": map[string]interface{}{
									"type":        "string",
									"description": "Message to echo",
								},
							},
							"required": []string{"message"},
						},
					},
				},
			},
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)
		assert.Contains(t, string(data), "tools")
		assert.Contains(t, string(data), "echo")
	})

	t.Run("should format error response correctly", func(t *testing.T) {
		// Simulate error response
		response := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      3,
			"error": map[string]interface{}{
				"code":    -32601,
				"message": "Method not found",
			},
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)
		assert.Contains(t, string(data), "error")
		assert.Contains(t, string(data), "-32601")
	})
}

// TestMCPErrorHandling tests error handling scenarios
func TestMCPErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	t.Run("should handle session already initialized error", func(t *testing.T) {
		session := createE2ESession(t)

		// Try to initialize again
		clientInfo := &aggregates.ClientInfo{
			Name:    "Another-Client",
			Version: "2.0.0",
		}
		err := session.Initialize(clientInfo, "2024-11-05")
		assert.Error(t, err)
	})

	t.Run("should handle tool not found error", func(t *testing.T) {
		session := createE2ESession(t)

		tool, ok := session.GetTool("nonexistent_tool")
		assert.False(t, ok)
		assert.Nil(t, tool)
	})

	t.Run("should handle resource not found error", func(t *testing.T) {
		session := createE2ESession(t)

		resource, ok := session.GetResource("file:///nonexistent/path")
		assert.False(t, ok)
		assert.Nil(t, resource)
	})

	t.Run("should handle closed session operations", func(t *testing.T) {
		session := createE2ESession(t)
		session.Close()

		// Verify session is closed
		assert.True(t, session.IsClosed())
	})

	t.Run("should handle closed conversation operations", func(t *testing.T) {
		session := createE2ESession(t)
		conv := aggregates.NewConversation(session.ID(), vo.ModelClaude4Sonnet)

		conv.Close()

		// Try to add message to closed conversation
		_, err := conv.AddUserMessage("Should fail")
		assert.Error(t, err)
	})
}

// Helper functions

func createE2ESession(t *testing.T) *aggregates.Session {
	t.Helper()
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{
		Name:    "E2E-Test",
		Version: "1.0.0",
	}
	err := session.Initialize(clientInfo, "2024-11-05")
	require.NoError(t, err)
	session.MarkReady()
	return session
}

func createE2ETool(t *testing.T, name, description string) *entities.Tool {
	t.Helper()
	toolName, err := vo.NewToolName(name)
	require.NoError(t, err)
	toolDesc, err := vo.NewToolDescription(description)
	require.NoError(t, err)
	tool, err := entities.NewTool(toolName, toolDesc, nil)
	require.NoError(t, err)

	tool.SetHandler(func(input map[string]interface{}) (*entities.ToolResult, error) {
		if msg, ok := input["message"].(string); ok {
			return entities.NewTextToolResult("Echo: " + msg), nil
		}
		return entities.NewTextToolResult("E2E test result"), nil
	})

	return tool
}

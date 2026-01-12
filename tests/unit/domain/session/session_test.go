// Package session_test provides unit tests for the session aggregate.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package session_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestNewSession(t *testing.T) {
	t.Run("should create session with unique ID", func(t *testing.T) {
		session := aggregates.NewSession()
		require.NotNil(t, session)
		assert.NotEmpty(t, session.ID().String())
	})

	t.Run("should create session in created state", func(t *testing.T) {
		session := aggregates.NewSession()
		assert.Equal(t, aggregates.SessionStateCreated, session.State())
	})

	t.Run("should set default protocol version", func(t *testing.T) {
		session := aggregates.NewSession()
		assert.NotEmpty(t, session.ProtocolVersion().String())
	})

	t.Run("should generate unique IDs for different sessions", func(t *testing.T) {
		session1 := aggregates.NewSession()
		session2 := aggregates.NewSession()
		assert.NotEqual(t, session1.ID().String(), session2.ID().String())
	})

	t.Run("should have empty tool list initially", func(t *testing.T) {
		session := aggregates.NewSession()
		assert.Empty(t, session.ListTools())
	})

	t.Run("should have empty resource list initially", func(t *testing.T) {
		session := aggregates.NewSession()
		assert.Empty(t, session.ListResources())
	})

	t.Run("should have empty prompt list initially", func(t *testing.T) {
		session := aggregates.NewSession()
		assert.Empty(t, session.ListPrompts())
	})
}

func TestSessionInitialize(t *testing.T) {
	t.Run("should initialize session with client info", func(t *testing.T) {
		session := aggregates.NewSession()
		clientInfo := &aggregates.ClientInfo{
			Name:    "TestClient",
			Version: "1.0.0",
		}

		err := session.Initialize(clientInfo, "2024-11-05")
		require.NoError(t, err)
		assert.Equal(t, aggregates.SessionStateInitializing, session.State())
	})

	t.Run("should store client info", func(t *testing.T) {
		session := aggregates.NewSession()
		clientInfo := &aggregates.ClientInfo{
			Name:    "TestClient",
			Version: "2.0.0",
		}

		err := session.Initialize(clientInfo, "2024-11-05")
		require.NoError(t, err)

		storedClientInfo := session.ClientInfo()
		assert.Equal(t, "TestClient", storedClientInfo.Name)
		assert.Equal(t, "2.0.0", storedClientInfo.Version)
	})

	t.Run("should not initialize already initialized session", func(t *testing.T) {
		session := aggregates.NewSession()
		clientInfo := &aggregates.ClientInfo{
			Name:    "TestClient",
			Version: "1.0.0",
		}

		err := session.Initialize(clientInfo, "2024-11-05")
		require.NoError(t, err)

		err = session.Initialize(clientInfo, "2024-11-05")
		assert.Error(t, err)
	})
}

func TestSessionMarkReady(t *testing.T) {
	t.Run("should mark session as ready after initialization", func(t *testing.T) {
		session := aggregates.NewSession()
		clientInfo := &aggregates.ClientInfo{
			Name:    "TestClient",
			Version: "1.0.0",
		}

		err := session.Initialize(clientInfo, "2024-11-05")
		require.NoError(t, err)

		session.MarkReady()
		assert.Equal(t, aggregates.SessionStateReady, session.State())
	})

	t.Run("should not mark ready from created state", func(t *testing.T) {
		session := aggregates.NewSession()
		session.MarkReady()
		// Should remain in created state
		assert.Equal(t, aggregates.SessionStateCreated, session.State())
	})
}

func TestSessionClose(t *testing.T) {
	t.Run("should close session from ready state", func(t *testing.T) {
		session := createReadySession(t)

		session.Close()
		assert.Equal(t, aggregates.SessionStateClosed, session.State())
	})

	t.Run("should set closed time", func(t *testing.T) {
		session := createReadySession(t)
		beforeClose := time.Now()

		session.Close()

		closedAt := session.ClosedAt()
		require.NotNil(t, closedAt)
		assert.True(t, closedAt.After(beforeClose) || closedAt.Equal(beforeClose))
	})

	t.Run("should not change on double close", func(t *testing.T) {
		session := createReadySession(t)

		session.Close()
		firstClosedAt := session.ClosedAt()

		session.Close()
		// Should remain closed with same timestamp
		assert.Equal(t, aggregates.SessionStateClosed, session.State())
		assert.Equal(t, firstClosedAt, session.ClosedAt())
	})
}

func TestSessionToolManagement(t *testing.T) {
	t.Run("should register tool", func(t *testing.T) {
		session := createReadySession(t)
		tool := createTestTool(t, "test_tool", "Test tool description")

		session.RegisterTool(tool)
		assert.Len(t, session.ListTools(), 1)
	})

	t.Run("should get registered tool by name", func(t *testing.T) {
		session := createReadySession(t)
		tool := createTestTool(t, "my_tool", "My tool description")

		session.RegisterTool(tool)

		retrievedTool, ok := session.GetTool("my_tool")
		assert.True(t, ok)
		assert.NotNil(t, retrievedTool)
		assert.Equal(t, "my_tool", retrievedTool.Name().String())
	})

	t.Run("should return false for non-existent tool", func(t *testing.T) {
		session := createReadySession(t)
		_, ok := session.GetTool("non_existent")
		assert.False(t, ok)
	})

	t.Run("should register multiple tools", func(t *testing.T) {
		session := createReadySession(t)

		for i := 0; i < 5; i++ {
			tool := createTestTool(t, "tool_"+string(rune('a'+i)), "Description "+string(rune('a'+i)))
			session.RegisterTool(tool)
		}

		assert.Len(t, session.ListTools(), 5)
	})

	t.Run("should unregister tool", func(t *testing.T) {
		session := createReadySession(t)
		tool := createTestTool(t, "removable_tool", "To be removed")

		session.RegisterTool(tool)
		assert.Len(t, session.ListTools(), 1)

		session.UnregisterTool("removable_tool")
		assert.Empty(t, session.ListTools())
	})
}

func TestSessionResourceManagement(t *testing.T) {
	t.Run("should register resource", func(t *testing.T) {
		session := createReadySession(t)
		uri, _ := vo.NewResourceURI("file:///test/path")
		resource, _ := entities.NewResource(uri, "Test Resource")

		session.RegisterResource(resource)
		assert.Len(t, session.ListResources(), 1)
	})

	t.Run("should get registered resource by URI", func(t *testing.T) {
		session := createReadySession(t)
		uri, _ := vo.NewResourceURI("file:///my/resource")
		resource, _ := entities.NewResource(uri, "My Resource")

		session.RegisterResource(resource)

		retrievedResource, ok := session.GetResource("file:///my/resource")
		assert.True(t, ok)
		assert.NotNil(t, retrievedResource)
	})

	t.Run("should return false for non-existent resource", func(t *testing.T) {
		session := createReadySession(t)
		_, ok := session.GetResource("file:///non/existent")
		assert.False(t, ok)
	})
}

func TestSessionPromptManagement(t *testing.T) {
	t.Run("should register prompt", func(t *testing.T) {
		session := createReadySession(t)
		promptName, _ := vo.NewToolName("test_prompt")
		prompt, err := entities.NewPrompt(promptName, "Test prompt description")
		require.NoError(t, err)

		session.RegisterPrompt(prompt)
		assert.Len(t, session.ListPrompts(), 1)
	})

	t.Run("should get registered prompt by name", func(t *testing.T) {
		session := createReadySession(t)
		promptName, _ := vo.NewToolName("my_prompt")
		prompt, err := entities.NewPrompt(promptName, "My prompt description")
		require.NoError(t, err)

		session.RegisterPrompt(prompt)

		retrievedPrompt, ok := session.GetPrompt("my_prompt")
		assert.True(t, ok)
		assert.NotNil(t, retrievedPrompt)
	})

	t.Run("should return false for non-existent prompt", func(t *testing.T) {
		session := createReadySession(t)
		_, ok := session.GetPrompt("non_existent")
		assert.False(t, ok)
	})
}

func TestSessionConversations(t *testing.T) {
	t.Run("should create conversation", func(t *testing.T) {
		session := createReadySession(t)

		conv, err := session.CreateConversation(vo.ModelClaude4Sonnet)
		require.NoError(t, err)
		assert.NotNil(t, conv)
		assert.Len(t, session.ListConversations(), 1)
	})

	t.Run("should get conversation by ID", func(t *testing.T) {
		session := createReadySession(t)

		conv, err := session.CreateConversation(vo.ModelClaude4Sonnet)
		require.NoError(t, err)

		retrievedConv, ok := session.GetConversation(conv.ID())
		assert.True(t, ok)
		assert.NotNil(t, retrievedConv)
		assert.Equal(t, conv.ID(), retrievedConv.ID())
	})

	t.Run("should return false for non-existent conversation", func(t *testing.T) {
		session := createReadySession(t)
		nonExistentID := vo.GenerateConversationID()
		_, ok := session.GetConversation(nonExistentID)
		assert.False(t, ok)
	})
}

func TestSessionCapabilities(t *testing.T) {
	t.Run("should have default capabilities", func(t *testing.T) {
		session := aggregates.NewSession()
		caps := session.Capabilities()
		assert.NotNil(t, caps)
	})

	t.Run("should enable tools capability", func(t *testing.T) {
		session := aggregates.NewSession()
		caps := session.Capabilities()
		assert.True(t, caps.Tools.ListChanged)
	})

	t.Run("should enable resources capability", func(t *testing.T) {
		session := aggregates.NewSession()
		caps := session.Capabilities()
		assert.True(t, caps.Resources.Subscribe)
		assert.True(t, caps.Resources.ListChanged)
	})

	t.Run("should enable prompts capability", func(t *testing.T) {
		session := aggregates.NewSession()
		caps := session.Capabilities()
		assert.True(t, caps.Prompts.ListChanged)
	})
}

func TestSessionServerInfo(t *testing.T) {
	t.Run("should return server info", func(t *testing.T) {
		session := aggregates.NewSession()
		info := session.ServerInfo()
		assert.NotEmpty(t, info.Name)
		assert.NotEmpty(t, info.Version)
	})
}

func TestSessionCreatedAt(t *testing.T) {
	t.Run("should set created time", func(t *testing.T) {
		beforeCreate := time.Now()
		session := aggregates.NewSession()
		afterCreate := time.Now()

		createdAt := session.CreatedAt()
		assert.True(t, createdAt.After(beforeCreate) || createdAt.Equal(beforeCreate))
		assert.True(t, createdAt.Before(afterCreate) || createdAt.Equal(afterCreate))
	})
}

// Helper functions

func createReadySession(t *testing.T) *aggregates.Session {
	t.Helper()
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{
		Name:    "TestClient",
		Version: "1.0.0",
	}
	err := session.Initialize(clientInfo, "2024-11-05")
	require.NoError(t, err)
	session.MarkReady()
	return session
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

func BenchmarkNewSession(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = aggregates.NewSession()
	}
}

func BenchmarkSessionInitialize(b *testing.B) {
	clientInfo := &aggregates.ClientInfo{
		Name:    "BenchClient",
		Version: "1.0.0",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session := aggregates.NewSession()
		_ = session.Initialize(clientInfo, "2024-11-05")
	}
}

func BenchmarkSessionRegisterTool(b *testing.B) {
	toolName, _ := vo.NewToolName("bench_tool")
	toolDesc, _ := vo.NewToolDescription("Benchmark tool")
	tool, _ := entities.NewTool(toolName, toolDesc, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session := aggregates.NewSession()
		clientInfo := &aggregates.ClientInfo{Name: "Bench", Version: "1.0.0"}
		_ = session.Initialize(clientInfo, "2024-11-05")
		session.MarkReady()
		session.RegisterTool(tool)
	}
}

func BenchmarkSessionGetTool(b *testing.B) {
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{Name: "Bench", Version: "1.0.0"}
	_ = session.Initialize(clientInfo, "2024-11-05")
	session.MarkReady()

	// Register 100 tools
	for i := 0; i < 100; i++ {
		name := "bench_tool_" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10))
		toolName, _ := vo.NewToolName(name)
		toolDesc, _ := vo.NewToolDescription("Benchmark tool")
		tool, _ := entities.NewTool(toolName, toolDesc, nil)
		session.RegisterTool(tool)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = session.GetTool("bench_tool_50")
	}
}

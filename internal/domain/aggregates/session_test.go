// Package aggregates contains tests for domain aggregates
package aggregates

import (
	"fmt"
	"sync"
	"testing"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestNewSession(t *testing.T) {
	session := NewSession()

	if session == nil {
		t.Fatal("NewSession() returned nil")
	}

	if session.ID().IsEmpty() {
		t.Error("Session ID should not be empty")
	}

	if session.State() != SessionStateCreated {
		t.Errorf("Expected state %s, got %s", SessionStateCreated, session.State())
	}

	if session.ServerInfo() == nil {
		t.Error("ServerInfo should not be nil")
	}

	if session.ServerInfo().Name != "TelemetryFlow-MCP" {
		t.Errorf("Expected server name TelemetryFlow-MCP, got %s", session.ServerInfo().Name)
	}

	if session.Capabilities() == nil {
		t.Error("Capabilities should not be nil")
	}

	if session.Capabilities().Tools == nil {
		t.Error("Tools capability should not be nil")
	}

	if !session.Capabilities().Tools.ListChanged {
		t.Error("Tools.ListChanged should be true")
	}

	if session.CreatedAt().IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestSession_Initialize(t *testing.T) {
	session := NewSession()

	clientInfo := &ClientInfo{
		Name:    "TestClient",
		Version: "1.0.0",
	}

	err := session.Initialize(clientInfo, "2024-11-05")
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	if session.State() != SessionStateInitializing {
		t.Errorf("Expected state %s, got %s", SessionStateInitializing, session.State())
	}

	if session.ClientInfo() == nil {
		t.Error("ClientInfo should not be nil after initialization")
	}

	if session.ClientInfo().Name != "TestClient" {
		t.Errorf("Expected client name TestClient, got %s", session.ClientInfo().Name)
	}

	// Try to initialize again - should fail
	err = session.Initialize(clientInfo, "2024-11-05")
	if err == nil {
		t.Error("Expected error when initializing already initialized session")
	}
}

func TestSession_MarkReady(t *testing.T) {
	session := NewSession()
	clientInfo := &ClientInfo{Name: "Test", Version: "1.0"}

	// Should not change state if not initializing
	session.MarkReady()
	if session.State() != SessionStateCreated {
		t.Error("State should not change from Created to Ready directly")
	}

	// Initialize first
	_ = session.Initialize(clientInfo, "2024-11-05")

	// Now mark ready
	session.MarkReady()
	if session.State() != SessionStateReady {
		t.Errorf("Expected state %s, got %s", SessionStateReady, session.State())
	}

	if !session.IsReady() {
		t.Error("IsReady() should return true")
	}
}

func TestSession_Close(t *testing.T) {
	session := NewSession()

	session.Close()

	if session.State() != SessionStateClosed {
		t.Errorf("Expected state %s, got %s", SessionStateClosed, session.State())
	}

	if !session.IsClosed() {
		t.Error("IsClosed() should return true")
	}

	if session.ClosedAt() == nil {
		t.Error("ClosedAt should not be nil after closing")
	}

	// Closing again should be idempotent
	session.Close()
	if session.State() != SessionStateClosed {
		t.Error("Session should remain closed")
	}
}

func TestSession_Tools(t *testing.T) {
	session := NewSession()

	toolName, _ := vo.NewToolName("test_tool")
	toolDesc, _ := vo.NewToolDescription("A test tool")
	tool, _ := entities.NewTool(toolName, toolDesc, nil)

	// Register tool
	session.RegisterTool(tool)

	// Get tool
	retrieved, ok := session.GetTool("test_tool")
	if !ok {
		t.Error("Should find registered tool")
	}
	if retrieved.Name().String() != "test_tool" {
		t.Errorf("Expected tool name test_tool, got %s", retrieved.Name().String())
	}

	// List tools
	tools := session.ListTools()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}

	// Unregister tool
	session.UnregisterTool("test_tool")
	_, ok = session.GetTool("test_tool")
	if ok {
		t.Error("Should not find unregistered tool")
	}
}

func TestSession_Resources(t *testing.T) {
	session := NewSession()

	resourceURI, _ := vo.NewResourceURI("file:///test/resource")
	resource, _ := entities.NewResource(resourceURI, "Test Resource")

	// Register resource
	session.RegisterResource(resource)

	// Get resource
	retrieved, ok := session.GetResource("file:///test/resource")
	if !ok {
		t.Error("Should find registered resource")
	}
	if retrieved.URI().String() != "file:///test/resource" {
		t.Errorf("Expected URI file:///test/resource, got %s", retrieved.URI().String())
	}

	// List resources
	resources := session.ListResources()
	if len(resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(resources))
	}

	// Unregister resource
	session.UnregisterResource("file:///test/resource")
	_, ok = session.GetResource("file:///test/resource")
	if ok {
		t.Error("Should not find unregistered resource")
	}
}

func TestSession_Subscriptions(t *testing.T) {
	session := NewSession()

	uri := "file:///test/resource"

	// Subscribe to resource
	err := session.SubscribeResource(uri)
	if err != nil {
		t.Fatalf("SubscribeResource() failed: %v", err)
	}

	if !session.IsSubscribed(uri) {
		t.Error("Should be subscribed to resource")
	}

	// Unsubscribe
	session.UnsubscribeResource(uri)
	if session.IsSubscribed(uri) {
		t.Error("Should not be subscribed after unsubscribe")
	}
}

func TestSession_Prompts(t *testing.T) {
	session := NewSession()

	promptName, _ := vo.NewToolName("test_prompt")
	prompt, _ := entities.NewPrompt(promptName, "Test prompt description")

	// Register prompt
	session.RegisterPrompt(prompt)

	// Get prompt
	retrieved, ok := session.GetPrompt("test_prompt")
	if !ok {
		t.Error("Should find registered prompt")
	}
	if retrieved.Name().String() != "test_prompt" {
		t.Errorf("Expected prompt name test_prompt, got %s", retrieved.Name().String())
	}

	// List prompts
	prompts := session.ListPrompts()
	if len(prompts) != 1 {
		t.Errorf("Expected 1 prompt, got %d", len(prompts))
	}

	// Unregister prompt
	session.UnregisterPrompt("test_prompt")
	_, ok = session.GetPrompt("test_prompt")
	if ok {
		t.Error("Should not find unregistered prompt")
	}
}

func TestSession_Conversations(t *testing.T) {
	session := NewSession()
	clientInfo := &ClientInfo{Name: "Test", Version: "1.0"}
	_ = session.Initialize(clientInfo, "2024-11-05")
	session.MarkReady()

	model := vo.ModelClaude4Sonnet

	// Create conversation
	conv, err := session.CreateConversation(model)
	if err != nil {
		t.Fatalf("CreateConversation() failed: %v", err)
	}

	if conv == nil {
		t.Fatal("Conversation should not be nil")
	}

	// Get conversation
	retrieved, ok := session.GetConversation(conv.ID())
	if !ok {
		t.Error("Should find created conversation")
	}
	if !retrieved.ID().Equals(conv.ID()) {
		t.Error("Retrieved conversation ID should match")
	}

	// List conversations
	convs := session.ListConversations()
	if len(convs) != 1 {
		t.Errorf("Expected 1 conversation, got %d", len(convs))
	}

	// Close conversation
	err = session.CloseConversation(conv.ID())
	if err != nil {
		t.Fatalf("CloseConversation() failed: %v", err)
	}
}

func TestSession_CreateConversation_WhenClosed(t *testing.T) {
	session := NewSession()
	session.Close()

	model := vo.ModelClaude4Sonnet
	_, err := session.CreateConversation(model)
	if err != ErrSessionClosed {
		t.Errorf("Expected ErrSessionClosed, got %v", err)
	}
}

func TestSession_LogLevel(t *testing.T) {
	session := NewSession()

	// Default log level
	if session.LogLevel() != vo.LogLevelInfo {
		t.Errorf("Expected default log level %s, got %s", vo.LogLevelInfo, session.LogLevel())
	}

	// Set log level
	err := session.SetLogLevel(vo.LogLevelDebug)
	if err != nil {
		t.Fatalf("SetLogLevel() failed: %v", err)
	}

	if session.LogLevel() != vo.LogLevelDebug {
		t.Errorf("Expected log level %s, got %s", vo.LogLevelDebug, session.LogLevel())
	}
}

func TestSession_Metadata(t *testing.T) {
	session := NewSession()

	// Set metadata
	session.SetMetadata("key1", "value1")
	session.SetMetadata("key2", 42)

	// Get metadata
	val, ok := session.GetMetadata("key1")
	if !ok {
		t.Error("Should find metadata key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	val, ok = session.GetMetadata("key2")
	if !ok {
		t.Error("Should find metadata key2")
	}
	if val != 42 {
		t.Errorf("Expected 42, got %v", val)
	}

	// Get all metadata
	metadata := session.Metadata()
	if len(metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(metadata))
	}
}

func TestSession_Events(t *testing.T) {
	session := NewSession()

	// Should have created event
	events := session.Events()
	if len(events) == 0 {
		t.Error("Should have at least one event after creation")
	}

	// Events should be cleared after retrieval
	events = session.Events()
	if len(events) != 0 {
		t.Error("Events should be cleared after retrieval")
	}
}

func TestSession_ToInitializeResult(t *testing.T) {
	session := NewSession()

	result := session.ToInitializeResult()

	if result["serverInfo"] == nil {
		t.Error("Result should contain serverInfo")
	}

	if result["capabilities"] == nil {
		t.Error("Result should contain capabilities")
	}

	if result["protocolVersion"] == nil {
		t.Error("Result should contain protocolVersion")
	}
}

func TestSession_ConcurrentAccess(t *testing.T) {
	session := NewSession()
	clientInfo := &ClientInfo{Name: "Test", Version: "1.0"}
	_ = session.Initialize(clientInfo, "2024-11-05")
	session.MarkReady()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Concurrent tool registration
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			toolName, _ := vo.NewToolName(fmt.Sprintf("tool_%d", idx))
			toolDesc, _ := vo.NewToolDescription("Description")
			tool, _ := entities.NewTool(toolName, toolDesc, nil)
			session.RegisterTool(tool)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = session.ListTools()
			_ = session.State()
			_ = session.IsReady()
		}()
	}

	wg.Wait()

	// Verify all tools were registered
	tools := session.ListTools()
	if len(tools) != numGoroutines {
		t.Errorf("Expected %d tools, got %d", numGoroutines, len(tools))
	}
}

func BenchmarkSession_RegisterTool(b *testing.B) {
	session := NewSession()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		toolName, _ := vo.NewToolName(fmt.Sprintf("tool_%d", i))
		toolDesc, _ := vo.NewToolDescription("Description")
		tool, _ := entities.NewTool(toolName, toolDesc, nil)
		session.RegisterTool(tool)
	}
}

func BenchmarkSession_GetTool(b *testing.B) {
	session := NewSession()

	// Pre-register tools
	for i := 0; i < 100; i++ {
		toolName, _ := vo.NewToolName(fmt.Sprintf("tool_%d", i))
		toolDesc, _ := vo.NewToolDescription("Description")
		tool, _ := entities.NewTool(toolName, toolDesc, nil)
		session.RegisterTool(tool)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		session.GetTool("tool_50")
	}
}

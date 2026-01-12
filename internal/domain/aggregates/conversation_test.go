// Package aggregates contains tests for domain aggregates
package aggregates

import (
	"sync"
	"testing"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestNewConversation(t *testing.T) {
	sessionID := vo.GenerateSessionID()
	model := vo.ModelClaude4Sonnet

	conv := NewConversation(sessionID, model)

	if conv == nil {
		t.Fatal("NewConversation() returned nil")
	}

	if conv.ID().IsEmpty() {
		t.Error("Conversation ID should not be empty")
	}

	if !conv.SessionID().Equals(sessionID) {
		t.Error("SessionID should match")
	}

	if conv.Model() != model {
		t.Errorf("Expected model %s, got %s", model, conv.Model())
	}

	if conv.Status() != ConversationStatusActive {
		t.Errorf("Expected status %s, got %s", ConversationStatusActive, conv.Status())
	}

	if !conv.IsActive() {
		t.Error("IsActive() should return true")
	}

	if conv.MessageCount() != 0 {
		t.Errorf("Expected 0 messages, got %d", conv.MessageCount())
	}

	if conv.CreatedAt().IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestConversation_SetModel(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	err := conv.SetModel(vo.ModelClaude3Opus)
	if err != nil {
		t.Fatalf("SetModel() failed: %v", err)
	}

	if conv.Model() != vo.ModelClaude3Opus {
		t.Errorf("Expected model %s, got %s", vo.ModelClaude3Opus, conv.Model())
	}

	// Invalid model should fail
	err = conv.SetModel(vo.Model("invalid-model"))
	if err == nil {
		t.Error("Expected error for invalid model")
	}
}

func TestConversation_SystemPrompt(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	prompt, _ := vo.NewSystemPrompt("You are a helpful assistant.")
	err := conv.SetSystemPrompt(prompt)
	if err != nil {
		t.Fatalf("SetSystemPrompt() failed: %v", err)
	}

	if conv.SystemPrompt().String() != "You are a helpful assistant." {
		t.Errorf("Expected system prompt 'You are a helpful assistant.', got '%s'", conv.SystemPrompt().String())
	}
}

func TestConversation_AddUserMessage(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	msg, err := conv.AddUserMessage("Hello, Claude!")
	if err != nil {
		t.Fatalf("AddUserMessage() failed: %v", err)
	}

	if msg == nil {
		t.Fatal("Message should not be nil")
	}

	if conv.MessageCount() != 1 {
		t.Errorf("Expected 1 message, got %d", conv.MessageCount())
	}

	if conv.LastMessage() != msg {
		t.Error("LastMessage should be the added message")
	}

	messages := conv.Messages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message in Messages(), got %d", len(messages))
	}
}

func TestConversation_SystemPromptImmutableAfterUserMessage(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	// Add user message first
	_, _ = conv.AddUserMessage("Hello!")

	// Now try to set system prompt - should fail
	prompt, _ := vo.NewSystemPrompt("You are a helpful assistant.")
	err := conv.SetSystemPrompt(prompt)
	if err != ErrSystemPromptImmutable {
		t.Errorf("Expected ErrSystemPromptImmutable, got %v", err)
	}
}

func TestConversation_AddMessageToClosedConversation(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)
	conv.Close()

	_, err := conv.AddUserMessage("Hello!")
	if err != ErrConversationClosed {
		t.Errorf("Expected ErrConversationClosed, got %v", err)
	}
}

func TestConversation_PauseAndResume(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	if !conv.IsActive() {
		t.Error("Conversation should be active initially")
	}

	conv.Pause()
	if conv.Status() != ConversationStatusPaused {
		t.Errorf("Expected status %s, got %s", ConversationStatusPaused, conv.Status())
	}
	if conv.IsActive() {
		t.Error("IsActive() should return false when paused")
	}

	conv.Resume()
	if conv.Status() != ConversationStatusActive {
		t.Errorf("Expected status %s, got %s", ConversationStatusActive, conv.Status())
	}
	if !conv.IsActive() {
		t.Error("IsActive() should return true after resume")
	}
}

func TestConversation_Close(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	conv.Close()

	if conv.Status() != ConversationStatusClosed {
		t.Errorf("Expected status %s, got %s", ConversationStatusClosed, conv.Status())
	}

	if conv.ClosedAt() == nil {
		t.Error("ClosedAt should not be nil after closing")
	}

	// Closing again should be idempotent
	conv.Close()
	if conv.Status() != ConversationStatusClosed {
		t.Error("Conversation should remain closed")
	}
}

func TestConversation_Archive(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	// Can't archive an active conversation
	conv.Archive()
	if conv.Status() == ConversationStatusArchived {
		t.Error("Should not be able to archive active conversation")
	}

	// Close first, then archive
	conv.Close()
	conv.Archive()
	if conv.Status() != ConversationStatusArchived {
		t.Errorf("Expected status %s, got %s", ConversationStatusArchived, conv.Status())
	}
}

func TestConversation_MaxTokens(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	// Default max tokens
	if conv.MaxTokens() != 4096 {
		t.Errorf("Expected default max tokens 4096, got %d", conv.MaxTokens())
	}

	conv.SetMaxTokens(8192)
	if conv.MaxTokens() != 8192 {
		t.Errorf("Expected max tokens 8192, got %d", conv.MaxTokens())
	}
}

func TestConversation_Temperature(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	// Default temperature
	if conv.Temperature() != 1.0 {
		t.Errorf("Expected default temperature 1.0, got %f", conv.Temperature())
	}

	conv.SetTemperature(0.5)
	if conv.Temperature() != 0.5 {
		t.Errorf("Expected temperature 0.5, got %f", conv.Temperature())
	}

	// Should clamp values
	conv.SetTemperature(-1.0)
	if conv.Temperature() != 0 {
		t.Errorf("Expected temperature 0 (clamped), got %f", conv.Temperature())
	}

	conv.SetTemperature(3.0)
	if conv.Temperature() != 2.0 {
		t.Errorf("Expected temperature 2.0 (clamped), got %f", conv.Temperature())
	}
}

func TestConversation_TopP(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	// Default top_p
	if conv.TopP() != 1.0 {
		t.Errorf("Expected default top_p 1.0, got %f", conv.TopP())
	}

	conv.SetTopP(0.9)
	if conv.TopP() != 0.9 {
		t.Errorf("Expected top_p 0.9, got %f", conv.TopP())
	}

	// Should clamp values
	conv.SetTopP(-0.5)
	if conv.TopP() != 0 {
		t.Errorf("Expected top_p 0 (clamped), got %f", conv.TopP())
	}

	conv.SetTopP(1.5)
	if conv.TopP() != 1.0 {
		t.Errorf("Expected top_p 1.0 (clamped), got %f", conv.TopP())
	}
}

func TestConversation_TopK(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	// Default top_k
	if conv.TopK() != 0 {
		t.Errorf("Expected default top_k 0, got %d", conv.TopK())
	}

	conv.SetTopK(50)
	if conv.TopK() != 50 {
		t.Errorf("Expected top_k 50, got %d", conv.TopK())
	}

	// Should clamp negative values
	conv.SetTopK(-10)
	if conv.TopK() != 0 {
		t.Errorf("Expected top_k 0 (clamped), got %d", conv.TopK())
	}
}

func TestConversation_StopSequences(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	if conv.StopSequences() != nil && len(conv.StopSequences()) != 0 {
		t.Error("Expected empty stop sequences initially")
	}

	sequences := []string{"STOP", "END"}
	conv.SetStopSequences(sequences)

	result := conv.StopSequences()
	if len(result) != 2 {
		t.Errorf("Expected 2 stop sequences, got %d", len(result))
	}
}

func TestConversation_Tools(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	toolName, _ := vo.NewToolName("test_tool")
	toolDesc, _ := vo.NewToolDescription("A test tool")
	tool, _ := entities.NewTool(toolName, toolDesc, nil)

	// Add tool
	conv.AddTool(tool)

	tools := conv.Tools()
	if len(tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(tools))
	}

	// Get tool
	retrieved := conv.GetTool(toolName)
	if retrieved == nil {
		t.Error("Should find added tool")
	}

	// Remove tool
	conv.RemoveTool(toolName)
	tools = conv.Tools()
	if len(tools) != 0 {
		t.Errorf("Expected 0 tools after removal, got %d", len(tools))
	}
}

func TestConversation_Metadata(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	// Set metadata
	conv.SetMetadata("key1", "value1")
	conv.SetMetadata("key2", 42)

	// Get metadata
	val, ok := conv.GetMetadata("key1")
	if !ok {
		t.Error("Should find metadata key1")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	val, ok = conv.GetMetadata("key2")
	if !ok {
		t.Error("Should find metadata key2")
	}
	if val != 42 {
		t.Errorf("Expected 42, got %v", val)
	}

	// Get all metadata
	metadata := conv.Metadata()
	if len(metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(metadata))
	}
}

func TestConversation_Events(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	// Should have created event
	events := conv.Events()
	if len(events) == 0 {
		t.Error("Should have at least one event after creation")
	}

	// Events should be cleared after retrieval
	events = conv.Events()
	if len(events) != 0 {
		t.Error("Events should be cleared after retrieval")
	}

	// Clear events explicitly
	_, _ = conv.AddUserMessage("Test")
	conv.ClearEvents()
	events = conv.Events()
	if len(events) != 0 {
		t.Error("Events should be empty after ClearEvents()")
	}
}

func TestConversation_GetMessagesForAPI(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	_, _ = conv.AddUserMessage("Hello, Claude!")

	messages := conv.GetMessagesForAPI()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg["role"] != "user" {
		t.Errorf("Expected role 'user', got '%v'", msg["role"])
	}
}

func TestConversation_ConcurrentAccess(t *testing.T) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	var wg sync.WaitGroup
	numGoroutines := 50

	// Concurrent metadata writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			conv.SetMetadata("key", idx)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = conv.Status()
			_ = conv.MessageCount()
			_ = conv.IsActive()
			_ = conv.Messages()
		}()
	}

	wg.Wait()
}

func BenchmarkConversation_AddUserMessage(b *testing.B) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Reset conversation for each iteration to avoid MaxMessages limit
		if conv.MessageCount() >= 1000 {
			conv = NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)
		}
		_, _ = conv.AddUserMessage("Test message")
	}
}

func BenchmarkConversation_GetMessagesForAPI(b *testing.B) {
	conv := NewConversation(vo.GenerateSessionID(), vo.ModelClaude4Sonnet)

	// Add some messages
	for i := 0; i < 100; i++ {
		_, _ = conv.AddUserMessage("Test message")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = conv.GetMessagesForAPI()
	}
}

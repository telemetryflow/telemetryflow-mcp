// Package entities contains tests for domain entities
package entities

import (
	"errors"
	"testing"
	"time"

	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestNewTool(t *testing.T) {
	name, _ := vo.NewToolName("test_tool")
	desc, _ := vo.NewToolDescription("A test tool for testing")

	tool, err := NewTool(name, desc, nil)
	if err != nil {
		t.Fatalf("NewTool() failed: %v", err)
	}

	if tool == nil {
		t.Fatal("Tool should not be nil")
	}

	if tool.Name().String() != "test_tool" {
		t.Errorf("Expected name 'test_tool', got '%s'", tool.Name().String())
	}

	if tool.Description().String() != "A test tool for testing" {
		t.Errorf("Expected description 'A test tool for testing', got '%s'", tool.Description().String())
	}

	if !tool.IsEnabled() {
		t.Error("Tool should be enabled by default")
	}

	if tool.Timeout() != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", tool.Timeout())
	}
}

func TestTool_WithInputSchema(t *testing.T) {
	name, _ := vo.NewToolName("search_tool")
	desc, _ := vo.NewToolDescription("Search for items")

	schema := &JSONSchema{
		Type: "object",
		Properties: map[string]*JSONSchema{
			"query": {
				Type:        "string",
				Description: "Search query",
			},
			"limit": {
				Type:        "integer",
				Description: "Max results",
			},
		},
		Required: []string{"query"},
	}

	tool, err := NewTool(name, desc, schema)
	if err != nil {
		t.Fatalf("NewTool() failed: %v", err)
	}

	if tool.InputSchema() == nil {
		t.Error("InputSchema should not be nil")
	}

	if tool.InputSchema().Type != "object" {
		t.Errorf("Expected schema type 'object', got '%s'", tool.InputSchema().Type)
	}

	if len(tool.InputSchema().Properties) != 2 {
		t.Errorf("Expected 2 properties, got %d", len(tool.InputSchema().Properties))
	}
}

func TestTool_SetHandler(t *testing.T) {
	name, _ := vo.NewToolName("echo_tool")
	desc, _ := vo.NewToolDescription("Echo tool")

	tool, _ := NewTool(name, desc, nil)

	// Initially no handler
	if tool.Handler() != nil {
		t.Error("Handler should be nil initially")
	}

	// Set handler
	handler := func(input map[string]interface{}) (*ToolResult, error) {
		return NewTextToolResult("echoed: " + input["text"].(string)), nil
	}

	tool.SetHandler(handler)

	if tool.Handler() == nil {
		t.Error("Handler should be set")
	}
}

func TestTool_Execute_WithHandler(t *testing.T) {
	name, _ := vo.NewToolName("add_tool")
	desc, _ := vo.NewToolDescription("Add numbers")

	tool, _ := NewTool(name, desc, nil)

	handler := func(input map[string]interface{}) (*ToolResult, error) {
		a := input["a"].(float64)
		b := input["b"].(float64)
		result := a + b
		return NewTextToolResult("Result: " + string(rune(int(result)))), nil
	}

	tool.SetHandler(handler)

	result, err := tool.Execute(map[string]interface{}{
		"a": float64(5),
		"b": float64(3),
	})

	if err != nil {
		t.Fatalf("Execute() failed: %v", err)
	}

	if result.IsError {
		t.Error("Result should not be an error")
	}
}

func TestTool_Execute_WithoutHandler(t *testing.T) {
	name, _ := vo.NewToolName("no_handler_tool")
	desc, _ := vo.NewToolDescription("No handler tool")

	tool, _ := NewTool(name, desc, nil)

	result, err := tool.Execute(map[string]interface{}{})
	if err != nil {
		t.Fatalf("Execute() should not return error: %v", err)
	}

	if !result.IsError {
		t.Error("Result should be an error when no handler")
	}

	if result.Content[0].Text != "Tool handler not configured" {
		t.Errorf("Unexpected error message: %s", result.Content[0].Text)
	}
}

func TestTool_EnableDisable(t *testing.T) {
	name, _ := vo.NewToolName("toggle_tool")
	desc, _ := vo.NewToolDescription("Toggle tool")

	tool, _ := NewTool(name, desc, nil)

	if !tool.IsEnabled() {
		t.Error("Tool should be enabled by default")
	}

	tool.Disable()
	if tool.IsEnabled() {
		t.Error("Tool should be disabled after Disable()")
	}

	tool.Enable()
	if !tool.IsEnabled() {
		t.Error("Tool should be enabled after Enable()")
	}
}

func TestTool_Category(t *testing.T) {
	name, _ := vo.NewToolName("categorized_tool")
	desc, _ := vo.NewToolDescription("Categorized tool")

	tool, _ := NewTool(name, desc, nil)

	if tool.Category() != "" {
		t.Error("Category should be empty initially")
	}

	tool.SetCategory("utilities")
	if tool.Category() != "utilities" {
		t.Errorf("Expected category 'utilities', got '%s'", tool.Category())
	}
}

func TestTool_Tags(t *testing.T) {
	name, _ := vo.NewToolName("tagged_tool")
	desc, _ := vo.NewToolDescription("Tagged tool")

	tool, _ := NewTool(name, desc, nil)

	if tool.Tags() != nil && len(tool.Tags()) != 0 {
		t.Error("Tags should be empty initially")
	}

	tool.SetTags([]string{"tag1", "tag2"})
	if len(tool.Tags()) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tool.Tags()))
	}

	tool.AddTag("tag3")
	if len(tool.Tags()) != 3 {
		t.Errorf("Expected 3 tags after AddTag, got %d", len(tool.Tags()))
	}
}

func TestTool_RateLimit(t *testing.T) {
	name, _ := vo.NewToolName("limited_tool")
	desc, _ := vo.NewToolDescription("Limited tool")

	tool, _ := NewTool(name, desc, nil)

	if tool.RateLimitConfig() != nil {
		t.Error("RateLimit should be nil initially")
	}

	limit := &RateLimit{
		RequestsPerMinute: 10,
		RequestsPerHour:   100,
		RequestsPerDay:    1000,
	}

	tool.SetRateLimit(limit)
	if tool.RateLimitConfig() == nil {
		t.Error("RateLimit should be set")
	}

	if tool.RateLimitConfig().RequestsPerMinute != 10 {
		t.Errorf("Expected RequestsPerMinute 10, got %d", tool.RateLimitConfig().RequestsPerMinute)
	}
}

func TestTool_Timeout(t *testing.T) {
	name, _ := vo.NewToolName("timeout_tool")
	desc, _ := vo.NewToolDescription("Timeout tool")

	tool, _ := NewTool(name, desc, nil)

	// Default timeout
	if tool.Timeout() != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", tool.Timeout())
	}

	tool.SetTimeout(60 * time.Second)
	if tool.Timeout() != 60*time.Second {
		t.Errorf("Expected timeout 60s, got %v", tool.Timeout())
	}
}

func TestTool_Metadata(t *testing.T) {
	name, _ := vo.NewToolName("meta_tool")
	desc, _ := vo.NewToolDescription("Meta tool")

	tool, _ := NewTool(name, desc, nil)

	tool.SetMetadata("key1", "value1")
	tool.SetMetadata("key2", 42)

	metadata := tool.Metadata()
	if len(metadata) != 2 {
		t.Errorf("Expected 2 metadata entries, got %d", len(metadata))
	}

	if metadata["key1"] != "value1" {
		t.Errorf("Expected key1='value1', got '%v'", metadata["key1"])
	}
}

func TestTool_Timestamps(t *testing.T) {
	name, _ := vo.NewToolName("timestamp_tool")
	desc, _ := vo.NewToolDescription("Timestamp tool")

	beforeCreate := time.Now().UTC()
	tool, _ := NewTool(name, desc, nil)
	afterCreate := time.Now().UTC()

	if tool.CreatedAt().Before(beforeCreate) || tool.CreatedAt().After(afterCreate) {
		t.Error("CreatedAt should be within test bounds")
	}

	if tool.UpdatedAt().Before(beforeCreate) || tool.UpdatedAt().After(afterCreate) {
		t.Error("UpdatedAt should be within test bounds")
	}

	// Update should change UpdatedAt
	time.Sleep(time.Millisecond)
	tool.SetCategory("test")

	if !tool.UpdatedAt().After(tool.CreatedAt()) {
		t.Error("UpdatedAt should be after CreatedAt after update")
	}
}

func TestTool_ToMCPTool(t *testing.T) {
	name, _ := vo.NewToolName("mcp_tool")
	desc, _ := vo.NewToolDescription("MCP tool")

	schema := &JSONSchema{
		Type: "object",
		Properties: map[string]*JSONSchema{
			"input": {Type: "string"},
		},
	}

	tool, _ := NewTool(name, desc, schema)

	mcpTool := tool.ToMCPTool()

	if mcpTool["name"] != "mcp_tool" {
		t.Errorf("Expected name 'mcp_tool', got '%v'", mcpTool["name"])
	}

	if mcpTool["description"] != "MCP tool" {
		t.Errorf("Expected description 'MCP tool', got '%v'", mcpTool["description"])
	}

	if mcpTool["inputSchema"] == nil {
		t.Error("inputSchema should not be nil")
	}
}

func TestTool_ToJSON(t *testing.T) {
	name, _ := vo.NewToolName("json_tool")
	desc, _ := vo.NewToolDescription("JSON tool")

	tool, _ := NewTool(name, desc, nil)

	jsonBytes, err := tool.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() failed: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Error("JSON output should not be empty")
	}
}

func TestNewTextToolResult(t *testing.T) {
	result := NewTextToolResult("Hello, World!")

	if result.IsError {
		t.Error("Text result should not be an error")
	}

	if len(result.Content) != 1 {
		t.Errorf("Expected 1 content item, got %d", len(result.Content))
	}

	if result.Content[0].Type != "text" {
		t.Errorf("Expected type 'text', got '%s'", result.Content[0].Type)
	}

	if result.Content[0].Text != "Hello, World!" {
		t.Errorf("Expected text 'Hello, World!', got '%s'", result.Content[0].Text)
	}
}

func TestNewErrorToolResult(t *testing.T) {
	testErr := errors.New("something went wrong")
	result := NewErrorToolResult(testErr)

	if !result.IsError {
		t.Error("Error result should have IsError=true")
	}

	if result.Content[0].Text != "something went wrong" {
		t.Errorf("Expected error message 'something went wrong', got '%s'", result.Content[0].Text)
	}
}

func TestNewImageToolResult(t *testing.T) {
	result := NewImageToolResult("base64data", "image/png")

	if result.IsError {
		t.Error("Image result should not be an error")
	}

	if result.Content[0].Type != "image" {
		t.Errorf("Expected type 'image', got '%s'", result.Content[0].Type)
	}

	if result.Content[0].Data != "base64data" {
		t.Errorf("Expected data 'base64data', got '%s'", result.Content[0].Data)
	}

	if result.Content[0].MimeType != "image/png" {
		t.Errorf("Expected mimeType 'image/png', got '%s'", result.Content[0].MimeType)
	}
}

func TestNewResourceToolResult(t *testing.T) {
	result := NewResourceToolResult("file:///test", "content", "text/plain")

	if result.IsError {
		t.Error("Resource result should not be an error")
	}

	if result.Content[0].Type != "resource" {
		t.Errorf("Expected type 'resource', got '%s'", result.Content[0].Type)
	}

	if result.Content[0].URI != "file:///test" {
		t.Errorf("Expected URI 'file:///test', got '%s'", result.Content[0].URI)
	}
}

func BenchmarkNewTool(b *testing.B) {
	name, _ := vo.NewToolName("bench_tool")
	desc, _ := vo.NewToolDescription("Benchmark tool")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewTool(name, desc, nil)
	}
}

func BenchmarkTool_Execute(b *testing.B) {
	name, _ := vo.NewToolName("exec_tool")
	desc, _ := vo.NewToolDescription("Execute tool")

	tool, _ := NewTool(name, desc, nil)
	tool.SetHandler(func(input map[string]interface{}) (*ToolResult, error) {
		return NewTextToolResult("result"), nil
	})

	input := map[string]interface{}{"test": "value"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tool.Execute(input)
	}
}

func BenchmarkTool_ToMCPTool(b *testing.B) {
	name, _ := vo.NewToolName("mcp_bench_tool")
	desc, _ := vo.NewToolDescription("MCP benchmark tool")

	tool, _ := NewTool(name, desc, &JSONSchema{Type: "object"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tool.ToMCPTool()
	}
}

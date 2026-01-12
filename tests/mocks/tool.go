// Package mocks provides mock implementations for testing.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
)

// MockToolHandler is a mock tool handler for testing
type MockToolHandler struct {
	mock.Mock
}

// Execute mocks tool execution
func (m *MockToolHandler) Execute(input map[string]interface{}) (*entities.ToolResult, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ToolResult), args.Error(1)
}

// MockTool creates a mock tool for testing
func MockTool(name, description string) *entities.Tool {
	toolName, _ := vo.NewToolName(name)
	toolDesc, _ := vo.NewToolDescription(description)

	tool, _ := entities.NewTool(toolName, toolDesc, nil)
	tool.SetHandler(func(input map[string]interface{}) (*entities.ToolResult, error) {
		return entities.NewTextToolResult("Mock result"), nil
	})

	return tool
}

// MockToolWithSchema creates a mock tool with input schema
func MockToolWithSchema(name, description string) *entities.Tool {
	toolName, _ := vo.NewToolName(name)
	toolDesc, _ := vo.NewToolDescription(description)

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"input": {
				Type:        "string",
				Description: "Input parameter",
			},
			"count": {
				Type:        "integer",
				Description: "Count parameter",
			},
		},
		Required: []string{"input"},
	}

	tool, _ := entities.NewTool(toolName, toolDesc, schema)
	tool.SetHandler(func(input map[string]interface{}) (*entities.ToolResult, error) {
		return entities.NewTextToolResult("Mock result with schema"), nil
	})

	return tool
}

// MockToolWithError creates a mock tool that returns an error
func MockToolWithError(name, description string) *entities.Tool {
	toolName, _ := vo.NewToolName(name)
	toolDesc, _ := vo.NewToolDescription(description)

	tool, _ := entities.NewTool(toolName, toolDesc, nil)
	tool.SetHandler(func(input map[string]interface{}) (*entities.ToolResult, error) {
		return entities.NewErrorToolResult(errMockToolError), nil
	})

	return tool
}

var errMockToolError = &mockError{message: "mock tool error"}

type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

// MockToolRepository is a mock implementation of the tool repository
type MockToolRepository struct {
	mock.Mock
	tools map[string]*entities.Tool
}

// NewMockToolRepository creates a new mock tool repository
func NewMockToolRepository() *MockToolRepository {
	return &MockToolRepository{
		tools: make(map[string]*entities.Tool),
	}
}

// Save saves a tool
func (m *MockToolRepository) Save(ctx context.Context, tool *entities.Tool) error {
	args := m.Called(ctx, tool)
	if args.Error(0) == nil {
		m.tools[tool.Name().String()] = tool
	}
	return args.Error(0)
}

// GetByName retrieves a tool by name
func (m *MockToolRepository) GetByName(ctx context.Context, name vo.ToolName) (*entities.Tool, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Tool), args.Error(1)
}

// Delete deletes a tool
func (m *MockToolRepository) Delete(ctx context.Context, name vo.ToolName) error {
	args := m.Called(ctx, name)
	if args.Error(0) == nil {
		delete(m.tools, name.String())
	}
	return args.Error(0)
}

// List lists all tools
func (m *MockToolRepository) List(ctx context.Context) ([]*entities.Tool, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Tool), args.Error(1)
}

// GetTools returns all stored tools (for testing)
func (m *MockToolRepository) GetTools() map[string]*entities.Tool {
	return m.tools
}

// MockToolCall represents a mock tool call for testing
type MockToolCall struct {
	ToolName   string
	Input      map[string]interface{}
	Output     string
	IsError    bool
	Duration   time.Duration
	StartedAt  time.Time
	FinishedAt time.Time
}

// MockToolCalls returns a set of mock tool calls for testing
func MockToolCalls() []MockToolCall {
	now := time.Now()
	return []MockToolCall{
		{
			ToolName:   "read_file",
			Input:      map[string]interface{}{"path": "/test/file.txt"},
			Output:     "File content here",
			IsError:    false,
			Duration:   50 * time.Millisecond,
			StartedAt:  now.Add(-100 * time.Millisecond),
			FinishedAt: now.Add(-50 * time.Millisecond),
		},
		{
			ToolName:   "write_file",
			Input:      map[string]interface{}{"path": "/test/output.txt", "content": "Hello"},
			Output:     "File written successfully",
			IsError:    false,
			Duration:   30 * time.Millisecond,
			StartedAt:  now.Add(-50 * time.Millisecond),
			FinishedAt: now.Add(-20 * time.Millisecond),
		},
		{
			ToolName:   "execute_command",
			Input:      map[string]interface{}{"command": "ls -la"},
			Output:     "Error: permission denied",
			IsError:    true,
			Duration:   10 * time.Millisecond,
			StartedAt:  now.Add(-20 * time.Millisecond),
			FinishedAt: now.Add(-10 * time.Millisecond),
		},
	}
}

// BuiltInTools returns the names of built-in MCP tools
func BuiltInTools() []string {
	return []string{
		"claude_conversation",
		"read_file",
		"write_file",
		"list_directory",
		"execute_command",
		"search_files",
		"system_info",
		"echo",
	}
}

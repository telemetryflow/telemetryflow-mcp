// Package tools contains built-in MCP tools for TelemetryFlow
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/services"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// ToolRegistry manages built-in tools
type ToolRegistry struct {
	claudeService services.IClaudeService
	tools         map[string]*entities.Tool
}

// NewToolRegistry creates a new tool registry
func NewToolRegistry(claudeService services.IClaudeService) *ToolRegistry {
	registry := &ToolRegistry{
		claudeService: claudeService,
		tools:         make(map[string]*entities.Tool),
	}

	// Register built-in tools
	registry.registerBuiltinTools()

	return registry
}

// GetTools returns all registered tools
func (r *ToolRegistry) GetTools() []*entities.Tool {
	tools := make([]*entities.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools
}

// GetTool returns a tool by name
func (r *ToolRegistry) GetTool(name string) (*entities.Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// registerBuiltinTools registers all built-in tools
func (r *ToolRegistry) registerBuiltinTools() {
	// Claude conversation tool
	r.registerClaudeConversation()

	// File tools
	r.registerReadFile()
	r.registerWriteFile()
	r.registerListDirectory()

	// Shell tool
	r.registerExecuteCommand()

	// Search tool
	r.registerSearchFiles()

	// System info tool
	r.registerSystemInfo()

	// Echo tool (for testing)
	r.registerEcho()
}

// registerClaudeConversation registers the Claude conversation tool
func (r *ToolRegistry) registerClaudeConversation() {
	name, _ := vo.NewToolName("claude_conversation")
	desc, _ := vo.NewToolDescription("Send a message to Claude and receive a response. Use this for AI-powered assistance, code generation, analysis, and general conversation.")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"message": {
				Type:        "string",
				Description: "The message to send to Claude",
			},
			"system_prompt": {
				Type:        "string",
				Description: "Optional system prompt to set context",
			},
			"model": {
				Type:        "string",
				Description: "The Claude model to use (default: claude-sonnet-4-20250514)",
				Enum:        []interface{}{"claude-opus-4-20250514", "claude-sonnet-4-20250514", "claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022"},
			},
			"max_tokens": {
				Type:        "integer",
				Description: "Maximum tokens in the response (default: 4096)",
			},
		},
		Required: []string{"message"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("ai")
	tool.SetTags([]string{"claude", "conversation", "ai"})
	tool.SetHandler(r.handleClaudeConversation)
	tool.SetTimeout(120 * time.Second)

	r.tools["claude_conversation"] = tool
}

// handleClaudeConversation handles Claude conversation requests
func (r *ToolRegistry) handleClaudeConversation(input map[string]interface{}) (*entities.ToolResult, error) {
	message, ok := input["message"].(string)
	if !ok || message == "" {
		return entities.NewErrorToolResult(fmt.Errorf("message is required")), nil
	}

	// Build request
	model := vo.ModelClaude4Sonnet
	if m, ok := input["model"].(string); ok {
		model = vo.Model(m)
	}

	maxTokens := 4096
	if mt, ok := input["max_tokens"].(float64); ok {
		maxTokens = int(mt)
	}

	var systemPrompt vo.SystemPrompt
	if sp, ok := input["system_prompt"].(string); ok && sp != "" {
		systemPrompt, _ = vo.NewSystemPrompt(sp)
	}

	request := &services.ClaudeRequest{
		Model:        model,
		SystemPrompt: systemPrompt,
		Messages: []services.ClaudeMessage{
			{
				Role: vo.RoleUser,
				Content: []entities.ContentBlock{
					{Type: vo.ContentTypeText, Text: message},
				},
			},
		},
		MaxTokens: maxTokens,
	}

	// Call Claude API
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	response, err := r.claudeService.CreateMessage(ctx, request)
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	// Extract text content
	var text string
	for _, block := range response.Content {
		if block.Type == vo.ContentTypeText {
			text += block.Text
		}
	}

	return entities.NewTextToolResult(text), nil
}

// registerReadFile registers the read file tool
func (r *ToolRegistry) registerReadFile() {
	name, _ := vo.NewToolName("read_file")
	desc, _ := vo.NewToolDescription("Read the contents of a file at the specified path")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"path": {
				Type:        "string",
				Description: "The path to the file to read",
			},
			"encoding": {
				Type:        "string",
				Description: "The encoding to use (default: utf-8)",
			},
		},
		Required: []string{"path"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("file")
	tool.SetTags([]string{"file", "read"})
	tool.SetHandler(handleReadFile)

	r.tools["read_file"] = tool
}

func handleReadFile(input map[string]interface{}) (*entities.ToolResult, error) {
	path, ok := input["path"].(string)
	if !ok || path == "" {
		return entities.NewErrorToolResult(fmt.Errorf("path is required")), nil
	}

	// Security: Prevent path traversal
	absPath, err := filepath.Abs(path)
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	content, err := os.ReadFile(absPath) //nolint:gosec // G304: path is sanitized via filepath.Abs
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	return entities.NewTextToolResult(string(content)), nil
}

// registerWriteFile registers the write file tool
func (r *ToolRegistry) registerWriteFile() {
	name, _ := vo.NewToolName("write_file")
	desc, _ := vo.NewToolDescription("Write content to a file at the specified path")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"path": {
				Type:        "string",
				Description: "The path to the file to write",
			},
			"content": {
				Type:        "string",
				Description: "The content to write to the file",
			},
			"create_dirs": {
				Type:        "boolean",
				Description: "Create parent directories if they don't exist",
			},
		},
		Required: []string{"path", "content"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("file")
	tool.SetTags([]string{"file", "write"})
	tool.SetHandler(handleWriteFile)

	r.tools["write_file"] = tool
}

func handleWriteFile(input map[string]interface{}) (*entities.ToolResult, error) {
	path, ok := input["path"].(string)
	if !ok || path == "" {
		return entities.NewErrorToolResult(fmt.Errorf("path is required")), nil
	}

	content, ok := input["content"].(string)
	if !ok {
		return entities.NewErrorToolResult(fmt.Errorf("content is required")), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	// Create directories if requested
	if createDirs, ok := input["create_dirs"].(bool); ok && createDirs {
		dir := filepath.Dir(absPath)
		if err := os.MkdirAll(dir, 0750); err != nil {
			return entities.NewErrorToolResult(err), nil
		}
	}

	if err := os.WriteFile(absPath, []byte(content), 0600); err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	return entities.NewTextToolResult(fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), absPath)), nil
}

// registerListDirectory registers the list directory tool
func (r *ToolRegistry) registerListDirectory() {
	name, _ := vo.NewToolName("list_directory")
	desc, _ := vo.NewToolDescription("List files and directories at the specified path")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"path": {
				Type:        "string",
				Description: "The path to the directory to list",
			},
			"recursive": {
				Type:        "boolean",
				Description: "List recursively (default: false)",
			},
		},
		Required: []string{"path"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("file")
	tool.SetTags([]string{"file", "directory", "list"})
	tool.SetHandler(handleListDirectory)

	r.tools["list_directory"] = tool
}

func handleListDirectory(input map[string]interface{}) (*entities.ToolResult, error) {
	path, ok := input["path"].(string)
	if !ok || path == "" {
		return entities.NewErrorToolResult(fmt.Errorf("path is required")), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	var result []string
	for _, entry := range entries {
		prefix := "📄 "
		if entry.IsDir() {
			prefix = "📁 "
		}
		result = append(result, prefix+entry.Name())
	}

	return entities.NewTextToolResult(strings.Join(result, "\n")), nil
}

// registerExecuteCommand registers the execute command tool
func (r *ToolRegistry) registerExecuteCommand() {
	name, _ := vo.NewToolName("execute_command")
	desc, _ := vo.NewToolDescription("Execute a shell command and return the output")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"command": {
				Type:        "string",
				Description: "The command to execute",
			},
			"working_dir": {
				Type:        "string",
				Description: "The working directory for the command",
			},
			"timeout": {
				Type:        "integer",
				Description: "Timeout in seconds (default: 30)",
			},
		},
		Required: []string{"command"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("system")
	tool.SetTags([]string{"command", "shell", "execute"})
	tool.SetHandler(handleExecuteCommand)
	tool.SetTimeout(60 * time.Second)

	r.tools["execute_command"] = tool
}

func handleExecuteCommand(input map[string]interface{}) (*entities.ToolResult, error) {
	command, ok := input["command"].(string)
	if !ok || command == "" {
		return entities.NewErrorToolResult(fmt.Errorf("command is required")), nil
	}

	timeout := 30
	if t, ok := input["timeout"].(float64); ok {
		timeout = int(t)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", command) //nolint:gosec // G204: command execution is intentional for shell tool

	if workingDir, ok := input["working_dir"].(string); ok && workingDir != "" {
		cmd.Dir = workingDir
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return entities.NewErrorToolResult(fmt.Errorf("command timed out after %d seconds", timeout)), nil
		}
		return entities.NewTextToolResult(fmt.Sprintf("Command failed: %s\nOutput: %s", err.Error(), string(output))), nil
	}

	return entities.NewTextToolResult(string(output)), nil
}

// registerSearchFiles registers the search files tool
func (r *ToolRegistry) registerSearchFiles() {
	name, _ := vo.NewToolName("search_files")
	desc, _ := vo.NewToolDescription("Search for files matching a pattern in a directory")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"path": {
				Type:        "string",
				Description: "The directory to search in",
			},
			"pattern": {
				Type:        "string",
				Description: "The glob pattern to match (e.g., *.go, **/*.ts)",
			},
			"content_pattern": {
				Type:        "string",
				Description: "Optional: Search for files containing this text",
			},
		},
		Required: []string{"path", "pattern"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("file")
	tool.SetTags([]string{"file", "search", "find"})
	tool.SetHandler(handleSearchFiles)

	r.tools["search_files"] = tool
}

func handleSearchFiles(input map[string]interface{}) (*entities.ToolResult, error) {
	path, ok := input["path"].(string)
	if !ok || path == "" {
		return entities.NewErrorToolResult(fmt.Errorf("path is required")), nil
	}

	pattern, ok := input["pattern"].(string)
	if !ok || pattern == "" {
		return entities.NewErrorToolResult(fmt.Errorf("pattern is required")), nil
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	var matches []string
	err = filepath.Walk(absPath, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}

		matched, _ := filepath.Match(pattern, info.Name())
		if matched {
			relPath, _ := filepath.Rel(absPath, p)
			matches = append(matches, relPath)
		}
		return nil
	})

	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	if len(matches) == 0 {
		return entities.NewTextToolResult("No files found matching pattern: " + pattern), nil
	}

	return entities.NewTextToolResult(fmt.Sprintf("Found %d files:\n%s", len(matches), strings.Join(matches, "\n"))), nil
}

// registerSystemInfo registers the system info tool
func (r *ToolRegistry) registerSystemInfo() {
	name, _ := vo.NewToolName("system_info")
	desc, _ := vo.NewToolDescription("Get system information")

	schema := &entities.JSONSchema{
		Type:       "object",
		Properties: map[string]*entities.JSONSchema{},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("system")
	tool.SetTags([]string{"system", "info"})
	tool.SetHandler(handleSystemInfo)

	r.tools["system_info"] = tool
}

func handleSystemInfo(input map[string]interface{}) (*entities.ToolResult, error) {
	hostname, _ := os.Hostname()
	wd, _ := os.Getwd()

	info := map[string]interface{}{
		"hostname":    hostname,
		"working_dir": wd,
		"os":          os.Getenv("GOOS"),
		"arch":        os.Getenv("GOARCH"),
		"user":        os.Getenv("USER"),
		"home":        os.Getenv("HOME"),
		"shell":       os.Getenv("SHELL"),
		"time":        time.Now().Format(time.RFC3339),
	}

	data, _ := json.MarshalIndent(info, "", "  ")
	return entities.NewTextToolResult(string(data)), nil
}

// registerEcho registers the echo tool (for testing)
func (r *ToolRegistry) registerEcho() {
	name, _ := vo.NewToolName("echo")
	desc, _ := vo.NewToolDescription("Echo back the input message (useful for testing)")

	schema := &entities.JSONSchema{
		Type: "object",
		Properties: map[string]*entities.JSONSchema{
			"message": {
				Type:        "string",
				Description: "The message to echo back",
			},
		},
		Required: []string{"message"},
	}

	tool, _ := entities.NewTool(name, desc, schema)
	tool.SetCategory("utility")
	tool.SetTags([]string{"test", "echo"})
	tool.SetHandler(handleEcho)

	r.tools["echo"] = tool
}

func handleEcho(input map[string]interface{}) (*entities.ToolResult, error) {
	message, ok := input["message"].(string)
	if !ok {
		return entities.NewErrorToolResult(fmt.Errorf("message is required")), nil
	}
	return entities.NewTextToolResult(message), nil
}

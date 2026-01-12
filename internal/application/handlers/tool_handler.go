// Package handlers contains CQRS handlers for the TelemetryFlow GO MCP service
package handlers

import (
	"context"
	"errors"
	"time"

	"github.com/telemetryflow/telemetryflow-mcp/internal/application/commands"
	"github.com/telemetryflow/telemetryflow-mcp/internal/application/queries"
	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/events"
	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/repositories"
	vo "github.com/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
)

// Tool handler errors
var (
	ErrToolNotFound      = errors.New("tool not found")
	ErrToolAlreadyExists = errors.New("tool already exists")
	ErrToolDisabled      = errors.New("tool is disabled")
	ErrInvalidToolInput  = errors.New("invalid tool input")
	ErrToolExecution     = errors.New("tool execution failed")
)

// ToolHandler handles tool-related commands and queries
type ToolHandler struct {
	sessionRepo    repositories.ISessionRepository
	toolRepo       repositories.IToolRepository
	eventPublisher EventPublisher
	toolRegistry   map[string]entities.ToolHandler
}

// NewToolHandler creates a new ToolHandler
func NewToolHandler(
	sessionRepo repositories.ISessionRepository,
	toolRepo repositories.IToolRepository,
	eventPublisher EventPublisher,
) *ToolHandler {
	return &ToolHandler{
		sessionRepo:    sessionRepo,
		toolRepo:       toolRepo,
		eventPublisher: eventPublisher,
		toolRegistry:   make(map[string]entities.ToolHandler),
	}
}

// RegisterToolHandler registers a tool handler function
func (h *ToolHandler) RegisterToolHandler(name string, handler entities.ToolHandler) {
	h.toolRegistry[name] = handler
}

// HandleRegisterTool handles RegisterToolCommand
func (h *ToolHandler) HandleRegisterTool(ctx context.Context, cmd *commands.RegisterToolCommand) (*entities.Tool, error) {
	// Verify session exists
	session, err := h.sessionRepo.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Create tool name value object
	name, err := vo.NewToolName(cmd.Name)
	if err != nil {
		return nil, err
	}

	// Check if tool already exists
	existing, _ := h.toolRepo.FindByName(ctx, name)
	if existing != nil {
		return nil, ErrToolAlreadyExists
	}

	// Create description value object
	description, err := vo.NewToolDescription(cmd.Description)
	if err != nil {
		return nil, err
	}

	// Create tool entity
	tool, err := entities.NewTool(name, description, cmd.InputSchema)
	if err != nil {
		return nil, err
	}

	// Set optional properties
	if cmd.Category != "" {
		tool.SetCategory(cmd.Category)
	}
	if len(cmd.Tags) > 0 {
		tool.SetTags(cmd.Tags)
	}

	// Set handler if registered
	if handler, ok := h.toolRegistry[cmd.Name]; ok {
		tool.SetHandler(handler)
	}

	// Register tool in repository
	if err := h.toolRepo.Register(ctx, tool); err != nil {
		return nil, err
	}

	// Register in session
	session.RegisterTool(tool)
	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}

	// Publish event (best-effort, don't fail on publish errors)
	event := events.NewToolRegisteredEvent(cmd.SessionID, cmd.Name)
	_ = h.eventPublisher.Publish(ctx, event)

	return tool, nil
}

// HandleUnregisterTool handles UnregisterToolCommand
func (h *ToolHandler) HandleUnregisterTool(ctx context.Context, cmd *commands.UnregisterToolCommand) error {
	// Verify session exists
	session, err := h.sessionRepo.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}

	// Create tool name value object
	name, err := vo.NewToolName(cmd.Name)
	if err != nil {
		return err
	}

	// Verify tool exists
	exists, err := h.toolRepo.Exists(ctx, name)
	if err != nil {
		return err
	}
	if !exists {
		return ErrToolNotFound
	}

	// Unregister from repository
	if err := h.toolRepo.Unregister(ctx, name); err != nil {
		return err
	}

	// Unregister from session
	session.UnregisterTool(cmd.Name)
	return h.sessionRepo.Save(ctx, session)
}

// HandleExecuteTool handles ExecuteToolCommand
func (h *ToolHandler) HandleExecuteTool(ctx context.Context, cmd *commands.ExecuteToolCommand) (*entities.ToolResult, error) {
	startTime := time.Now()

	// Verify session exists
	session, err := h.sessionRepo.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Create tool name value object
	name, err := vo.NewToolName(cmd.Name)
	if err != nil {
		return nil, err
	}

	// Get tool
	tool, err := h.toolRepo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if tool == nil {
		return nil, ErrToolNotFound
	}

	// Check if tool is enabled
	if !tool.IsEnabled() {
		return nil, ErrToolDisabled
	}

	// Execute tool with timeout
	execCtx, cancel := context.WithTimeout(ctx, tool.Timeout())
	defer cancel()

	result, err := h.executeToolWithContext(execCtx, tool, cmd.Arguments)
	duration := time.Since(startTime)

	// Publish execution event (best-effort, don't fail on publish errors)
	success := err == nil && (result == nil || !result.IsError)
	event := events.NewToolExecutedEvent(cmd.SessionID, cmd.Name, success, duration)
	_ = h.eventPublisher.Publish(ctx, event)

	if err != nil {
		return entities.NewErrorToolResult(err), nil
	}

	return result, nil
}

// executeToolWithContext executes a tool with context
func (h *ToolHandler) executeToolWithContext(ctx context.Context, tool *entities.Tool, input map[string]interface{}) (*entities.ToolResult, error) {
	resultChan := make(chan *entities.ToolResult, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := tool.Execute(input)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		return nil, err
	case result := <-resultChan:
		return result, nil
	}
}

// HandleGetTool handles GetToolQuery
func (h *ToolHandler) HandleGetTool(ctx context.Context, query *queries.GetToolQuery) (*entities.Tool, error) {
	// Create tool name value object
	name, err := vo.NewToolName(query.Name)
	if err != nil {
		return nil, err
	}

	tool, err := h.toolRepo.FindByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if tool == nil {
		return nil, ErrToolNotFound
	}

	return tool, nil
}

// ToolListResult represents the result of listing tools
type ToolListResult struct {
	Tools      []*entities.Tool
	NextCursor string
}

// HandleListTools handles ListToolsQuery
func (h *ToolHandler) HandleListTools(ctx context.Context, query *queries.ListToolsQuery) (*ToolListResult, error) {
	var tools []*entities.Tool
	var err error

	if query.Category != "" {
		tools, err = h.toolRepo.FindByCategory(ctx, query.Category)
	} else if query.Tag != "" {
		tools, err = h.toolRepo.FindByTag(ctx, query.Tag)
	} else if query.EnabledOnly {
		tools, err = h.toolRepo.FindEnabled(ctx)
	} else {
		tools, err = h.toolRepo.FindAll(ctx)
	}

	if err != nil {
		return nil, err
	}

	// Apply pagination if needed
	if query.Limit > 0 && len(tools) > query.Limit {
		tools = tools[:query.Limit]
	}

	return &ToolListResult{
		Tools:      tools,
		NextCursor: "", // Pagination cursor implementation
	}, nil
}

// ToMCPToolList converts tools to MCP format
func (r *ToolListResult) ToMCPToolList() map[string]interface{} {
	tools := make([]map[string]interface{}, len(r.Tools))
	for i, tool := range r.Tools {
		tools[i] = tool.ToMCPTool()
	}

	result := map[string]interface{}{
		"tools": tools,
	}
	if r.NextCursor != "" {
		result["nextCursor"] = r.NextCursor
	}
	return result
}

// Package handlers contains CQRS handlers for the TelemetryFlow GO MCP service
package handlers

import (
	"context"
	"errors"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/commands"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/queries"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/repositories"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/services"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// Conversation handler errors
var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrMessageEmpty         = errors.New("message cannot be empty")
)

// ConversationHandler handles conversation-related commands and queries
type ConversationHandler struct {
	sessionRepo      repositories.ISessionRepository
	conversationRepo repositories.IConversationRepository
	claudeService    services.IClaudeService
	eventPublisher   EventPublisher
}

// NewConversationHandler creates a new ConversationHandler
func NewConversationHandler(
	sessionRepo repositories.ISessionRepository,
	conversationRepo repositories.IConversationRepository,
	claudeService services.IClaudeService,
	eventPublisher EventPublisher,
) *ConversationHandler {
	return &ConversationHandler{
		sessionRepo:      sessionRepo,
		conversationRepo: conversationRepo,
		claudeService:    claudeService,
		eventPublisher:   eventPublisher,
	}
}

// HandleCreateConversation handles CreateConversationCommand
func (h *ConversationHandler) HandleCreateConversation(ctx context.Context, cmd *commands.CreateConversationCommand) (*aggregates.Conversation, error) {
	// Verify session exists
	session, err := h.sessionRepo.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Validate model
	model := cmd.Model
	if !model.IsValid() {
		model = vo.DefaultModel
	}

	// Create conversation
	conversation, err := session.CreateConversation(model)
	if err != nil {
		return nil, err
	}

	// Set system prompt if provided
	if cmd.SystemPrompt != "" {
		systemPrompt, err := vo.NewSystemPrompt(cmd.SystemPrompt)
		if err != nil {
			return nil, err
		}
		if err := conversation.SetSystemPrompt(systemPrompt); err != nil {
			return nil, err
		}
	}

	// Set optional parameters
	if cmd.MaxTokens > 0 {
		conversation.SetMaxTokens(cmd.MaxTokens)
	}
	if cmd.Temperature >= 0 {
		conversation.SetTemperature(cmd.Temperature)
	}

	// Save session and conversation
	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}
	if err := h.conversationRepo.Save(ctx, conversation); err != nil {
		return nil, err
	}

	// Publish events (best-effort, don't fail on publish errors)
	for _, event := range conversation.Events() {
		_ = h.eventPublisher.Publish(ctx, event)
	}

	return conversation, nil
}

// SendMessageResult represents the result of sending a message
type SendMessageResult struct {
	Response   *services.ClaudeResponse
	ToolUses   []entities.ContentBlock
	HasToolUse bool
}

// HandleSendMessage handles SendMessageCommand
func (h *ConversationHandler) HandleSendMessage(ctx context.Context, cmd *commands.SendMessageCommand) (*SendMessageResult, error) {
	// Validate message
	if cmd.Content == "" {
		return nil, ErrMessageEmpty
	}

	// Get conversation
	conversation, err := h.conversationRepo.FindByID(ctx, cmd.ConversationID)
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, ErrConversationNotFound
	}

	// Check if conversation is active
	if !conversation.IsActive() {
		return nil, aggregates.ErrConversationClosed
	}

	// Add user message
	_, err = conversation.AddUserMessage(cmd.Content)
	if err != nil {
		return nil, err
	}

	// Build Claude request
	request := h.buildClaudeRequest(conversation)

	// Call Claude API
	var response *services.ClaudeResponse
	if cmd.Stream {
		// For streaming, we collect events and build response
		response, err = h.handleStreamingRequest(ctx, request)
	} else {
		response, err = h.claudeService.CreateMessage(ctx, request)
	}

	if err != nil {
		return nil, err
	}

	// Add assistant message
	_, err = conversation.AddAssistantMessage(response.Content)
	if err != nil {
		return nil, err
	}

	// Save conversation
	if err := h.conversationRepo.Save(ctx, conversation); err != nil {
		return nil, err
	}

	// Publish events (best-effort, don't fail on publish errors)
	for _, event := range conversation.Events() {
		_ = h.eventPublisher.Publish(ctx, event)
	}

	// Check for tool use
	var toolUses []entities.ContentBlock
	hasToolUse := false
	for _, block := range response.Content {
		if block.Type == vo.ContentTypeToolUse {
			toolUses = append(toolUses, block)
			hasToolUse = true
		}
	}

	return &SendMessageResult{
		Response:   response,
		ToolUses:   toolUses,
		HasToolUse: hasToolUse,
	}, nil
}

// HandleAddToolResult handles AddToolResultCommand
func (h *ConversationHandler) HandleAddToolResult(ctx context.Context, cmd *commands.AddToolResultCommand) error {
	// Get conversation
	conversation, err := h.conversationRepo.FindByID(ctx, cmd.ConversationID)
	if err != nil {
		return err
	}
	if conversation == nil {
		return ErrConversationNotFound
	}

	// Create tool result message
	toolResultBlock := entities.ContentBlock{
		Type:      vo.ContentTypeToolResult,
		ToolUseID: cmd.ToolUseID,
		Content:   cmd.Content,
		IsError:   cmd.IsError,
	}

	msg, err := entities.NewMessage(vo.RoleUser, []entities.ContentBlock{toolResultBlock})
	if err != nil {
		return err
	}

	if err := conversation.AddMessage(msg); err != nil {
		return err
	}

	// Save conversation
	if err := h.conversationRepo.Save(ctx, conversation); err != nil {
		return err
	}

	// Publish events (best-effort, don't fail on publish errors)
	for _, event := range conversation.Events() {
		_ = h.eventPublisher.Publish(ctx, event)
	}

	return nil
}

// HandleCloseConversation handles CloseConversationCommand
func (h *ConversationHandler) HandleCloseConversation(ctx context.Context, cmd *commands.CloseConversationCommand) error {
	conversation, err := h.conversationRepo.FindByID(ctx, cmd.ConversationID)
	if err != nil {
		return err
	}
	if conversation == nil {
		return ErrConversationNotFound
	}

	conversation.Close()

	if err := h.conversationRepo.Save(ctx, conversation); err != nil {
		return err
	}

	// Publish events (best-effort, don't fail on publish errors)
	for _, event := range conversation.Events() {
		_ = h.eventPublisher.Publish(ctx, event)
	}

	return nil
}

// HandleGetConversation handles GetConversationQuery
func (h *ConversationHandler) HandleGetConversation(ctx context.Context, query *queries.GetConversationQuery) (*aggregates.Conversation, error) {
	conversation, err := h.conversationRepo.FindByID(ctx, query.ConversationID)
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, ErrConversationNotFound
	}
	return conversation, nil
}

// ConversationListResult represents the result of listing conversations
type ConversationListResult struct {
	Conversations []*aggregates.Conversation
	NextCursor    string
}

// HandleListConversations handles ListConversationsQuery
func (h *ConversationHandler) HandleListConversations(ctx context.Context, query *queries.ListConversationsQuery) (*ConversationListResult, error) {
	var conversations []*aggregates.Conversation
	var err error

	if !query.SessionID.IsEmpty() {
		conversations, err = h.conversationRepo.FindBySessionID(ctx, query.SessionID)
	} else if query.ActiveOnly {
		conversations, err = h.conversationRepo.FindActive(ctx)
	} else {
		// Get all (would need pagination in real implementation)
		conversations, err = h.conversationRepo.FindActive(ctx)
	}

	if err != nil {
		return nil, err
	}

	// Apply pagination
	if query.Limit > 0 && len(conversations) > query.Limit {
		conversations = conversations[:query.Limit]
	}

	return &ConversationListResult{
		Conversations: conversations,
		NextCursor:    "",
	}, nil
}

// HandleGetConversationMessages handles GetConversationMessagesQuery
func (h *ConversationHandler) HandleGetConversationMessages(ctx context.Context, query *queries.GetConversationMessagesQuery) ([]*entities.Message, error) {
	conversation, err := h.conversationRepo.FindByID(ctx, query.ConversationID)
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, ErrConversationNotFound
	}

	messages := conversation.Messages()

	// Apply pagination
	if query.Offset > 0 && query.Offset < len(messages) {
		messages = messages[query.Offset:]
	}
	if query.Limit > 0 && len(messages) > query.Limit {
		messages = messages[:query.Limit]
	}

	return messages, nil
}

// buildClaudeRequest builds a Claude API request from a conversation
func (h *ConversationHandler) buildClaudeRequest(conversation *aggregates.Conversation) *services.ClaudeRequest {
	messages := make([]services.ClaudeMessage, len(conversation.Messages()))
	for i, msg := range conversation.Messages() {
		messages[i] = services.ClaudeMessage{
			Role:    msg.Role(),
			Content: msg.Content(),
		}
	}

	// Convert tools
	var tools []services.ClaudeTool
	for _, tool := range conversation.Tools() {
		tools = append(tools, services.ClaudeTool{
			Name:        tool.Name().String(),
			Description: tool.Description().String(),
			InputSchema: tool.InputSchema(),
		})
	}

	return &services.ClaudeRequest{
		Model:         conversation.Model(),
		SystemPrompt:  conversation.SystemPrompt(),
		Messages:      messages,
		MaxTokens:     conversation.MaxTokens(),
		Temperature:   conversation.Temperature(),
		TopP:          conversation.TopP(),
		TopK:          conversation.TopK(),
		StopSequences: conversation.StopSequences(),
		Tools:         tools,
	}
}

// handleStreamingRequest handles streaming response
func (h *ConversationHandler) handleStreamingRequest(ctx context.Context, request *services.ClaudeRequest) (*services.ClaudeResponse, error) {
	request.Stream = true
	eventChan, err := h.claudeService.CreateMessageStream(ctx, request)
	if err != nil {
		return nil, err
	}

	var response *services.ClaudeResponse
	var contentBlocks []entities.ContentBlock

	for event := range eventChan {
		if event.Error != nil {
			return nil, event.Error
		}
		if event.Message != nil {
			response = event.Message
		}
		if event.ContentBlock != nil {
			contentBlocks = append(contentBlocks, *event.ContentBlock)
		}
	}

	if response != nil {
		response.Content = contentBlocks
	}

	return response, nil
}

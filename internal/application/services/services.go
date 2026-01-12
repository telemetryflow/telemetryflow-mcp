// Package services provides application-level services for the TelemetryFlow GO MCP
package services

import (
	"context"
	"errors"
	"time"

	"github.com/telemetryflow/telemetryflow-mcp/internal/application/dto"
	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/repositories"
	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/services"
	vo "github.com/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
)

// Application service errors
var (
	ErrServiceUnavailable = errors.New("service temporarily unavailable")
	ErrInvalidInput       = errors.New("invalid input provided")
	ErrOperationFailed    = errors.New("operation failed")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrRateLimited        = errors.New("rate limit exceeded")
)

// IConversationService defines the conversation application service interface
type IConversationService interface {
	// StartConversation initiates a new conversation
	StartConversation(ctx context.Context, req *StartConversationRequest) (*dto.ConversationDTO, error)

	// SendMessage sends a message and gets a response
	SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error)

	// GetConversation retrieves a conversation by ID
	GetConversation(ctx context.Context, conversationID string) (*dto.ConversationDTO, error)

	// ListConversations lists conversations for a session
	ListConversations(ctx context.Context, sessionID string) ([]dto.ConversationDTO, error)

	// CloseConversation closes an active conversation
	CloseConversation(ctx context.Context, conversationID string) error
}

// ISessionService defines the session application service interface
type ISessionService interface {
	// InitializeSession creates a new MCP session
	InitializeSession(ctx context.Context, req *dto.InitializeRequest) (*dto.InitializeResult, error)

	// GetSession retrieves a session by ID
	GetSession(ctx context.Context, sessionID string) (*dto.SessionDTO, error)

	// CloseSession terminates a session
	CloseSession(ctx context.Context, sessionID string) error

	// ListActiveSessions lists all active sessions
	ListActiveSessions(ctx context.Context) ([]dto.SessionDTO, error)
}

// IToolService defines the tool application service interface
type IToolService interface {
	// ExecuteTool executes a tool with the given arguments
	ExecuteTool(ctx context.Context, req *ExecuteToolRequest) (*dto.CallToolResult, error)

	// ListTools returns all available tools
	ListTools(ctx context.Context) ([]dto.ToolDTO, error)

	// GetTool retrieves a tool by name
	GetTool(ctx context.Context, name string) (*dto.ToolDTO, error)

	// RegisterTool registers a new tool
	RegisterTool(ctx context.Context, tool dto.ToolDTO) error
}

// IAnalyticsService defines the analytics application service interface
type IAnalyticsService interface {
	// GetDashboardSummary returns dashboard summary data
	GetDashboardSummary(ctx context.Context, since, until time.Time) (*DashboardSummaryResponse, error)

	// GetTokenUsage returns token usage statistics
	GetTokenUsage(ctx context.Context, since, until time.Time) ([]TokenUsageResponse, error)

	// GetToolUsage returns tool usage statistics
	GetToolUsage(ctx context.Context, since, until time.Time) ([]ToolUsageResponse, error)
}

// StartConversationRequest represents a request to start a conversation
type StartConversationRequest struct {
	SessionID    string  `json:"sessionId"`
	Model        string  `json:"model,omitempty"`
	SystemPrompt string  `json:"systemPrompt,omitempty"`
	MaxTokens    int     `json:"maxTokens,omitempty"`
	Temperature  float64 `json:"temperature,omitempty"`
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	ConversationID string `json:"conversationId"`
	Content        string `json:"content"`
	Stream         bool   `json:"stream,omitempty"`
}

// SendMessageResponse represents the response from sending a message
type SendMessageResponse struct {
	Message    dto.MessageDTO      `json:"message"`
	ToolCalls  []dto.ToolDTO       `json:"toolCalls,omitempty"`
	HasToolUse bool                `json:"hasToolUse"`
	Usage      *TokenUsageResponse `json:"usage,omitempty"`
}

// ExecuteToolRequest represents a request to execute a tool
type ExecuteToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Timeout   time.Duration          `json:"timeout,omitempty"`
}

// DashboardSummaryResponse represents dashboard summary data
type DashboardSummaryResponse struct {
	TotalRequests     uint64  `json:"totalRequests"`
	TotalTokens       uint64  `json:"totalTokens"`
	TotalToolCalls    uint64  `json:"totalToolCalls"`
	TotalSessions     uint64  `json:"totalSessions"`
	AvgLatencyMs      float64 `json:"avgLatencyMs"`
	ErrorRate         float64 `json:"errorRate"`
	RequestsPerMinute float64 `json:"requestsPerMinute"`
}

// TokenUsageResponse represents token usage data
type TokenUsageResponse struct {
	Model        string `json:"model"`
	InputTokens  uint64 `json:"inputTokens"`
	OutputTokens uint64 `json:"outputTokens"`
	TotalTokens  uint64 `json:"totalTokens"`
	RequestCount uint64 `json:"requestCount"`
}

// ToolUsageResponse represents tool usage data
type ToolUsageResponse struct {
	ToolName    string  `json:"toolName"`
	CallCount   uint64  `json:"callCount"`
	ErrorCount  uint64  `json:"errorCount"`
	SuccessRate float64 `json:"successRate"`
	AvgDuration float64 `json:"avgDurationMs"`
}

// ConversationService implements IConversationService
type ConversationService struct {
	sessionRepo      repositories.ISessionRepository
	conversationRepo repositories.IConversationRepository
	claudeService    services.IClaudeService
}

// NewConversationService creates a new ConversationService
func NewConversationService(
	sessionRepo repositories.ISessionRepository,
	conversationRepo repositories.IConversationRepository,
	claudeService services.IClaudeService,
) *ConversationService {
	return &ConversationService{
		sessionRepo:      sessionRepo,
		conversationRepo: conversationRepo,
		claudeService:    claudeService,
	}
}

// StartConversation initiates a new conversation
func (s *ConversationService) StartConversation(ctx context.Context, req *StartConversationRequest) (*dto.ConversationDTO, error) {
	if req.SessionID == "" {
		return nil, ErrInvalidInput
	}

	sessionID, err := vo.NewSessionID(req.SessionID)
	if err != nil {
		return nil, err
	}

	session, err := s.sessionRepo.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrInvalidInput
	}

	model := vo.DefaultModel
	if req.Model != "" {
		model = vo.Model(req.Model)
	}

	conversation, err := session.CreateConversation(model)
	if err != nil {
		return nil, err
	}

	if err := s.conversationRepo.Save(ctx, conversation); err != nil {
		return nil, err
	}

	return &dto.ConversationDTO{
		ID:        conversation.ID().String(),
		SessionID: conversation.SessionID().String(),
		Model:     string(conversation.Model()),
		IsActive:  conversation.IsActive(),
		CreatedAt: conversation.CreatedAt(),
		UpdatedAt: conversation.UpdatedAt(),
	}, nil
}

// SendMessage sends a message and gets a response
func (s *ConversationService) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
	if req.ConversationID == "" || req.Content == "" {
		return nil, ErrInvalidInput
	}

	conversationID, err := vo.NewConversationID(req.ConversationID)
	if err != nil {
		return nil, err
	}

	conversation, err := s.conversationRepo.FindByID(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, ErrInvalidInput
	}

	// Add user message
	msg, err := conversation.AddUserMessage(req.Content)
	if err != nil {
		return nil, err
	}

	// Save conversation
	if err := s.conversationRepo.Save(ctx, conversation); err != nil {
		return nil, err
	}

	// Convert message to DTO
	var contentBlocks []dto.ContentBlockDTO
	for _, block := range msg.Content() {
		contentBlocks = append(contentBlocks, dto.ContentBlockDTO{
			Type: string(block.Type),
			Text: block.Text,
		})
	}

	return &SendMessageResponse{
		Message: dto.MessageDTO{
			ID:        msg.ID().String(),
			Role:      string(msg.Role()),
			Content:   contentBlocks,
			CreatedAt: msg.CreatedAt(),
		},
		HasToolUse: false,
	}, nil
}

// GetConversation retrieves a conversation by ID
func (s *ConversationService) GetConversation(ctx context.Context, conversationID string) (*dto.ConversationDTO, error) {
	id, err := vo.NewConversationID(conversationID)
	if err != nil {
		return nil, err
	}

	conversation, err := s.conversationRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if conversation == nil {
		return nil, ErrInvalidInput
	}

	return &dto.ConversationDTO{
		ID:        conversation.ID().String(),
		SessionID: conversation.SessionID().String(),
		Model:     string(conversation.Model()),
		IsActive:  conversation.IsActive(),
		CreatedAt: conversation.CreatedAt(),
		UpdatedAt: conversation.UpdatedAt(),
	}, nil
}

// ListConversations lists conversations for a session
func (s *ConversationService) ListConversations(ctx context.Context, sessionID string) ([]dto.ConversationDTO, error) {
	id, err := vo.NewSessionID(sessionID)
	if err != nil {
		return nil, err
	}

	conversations, err := s.conversationRepo.FindBySessionID(ctx, id)
	if err != nil {
		return nil, err
	}

	var result []dto.ConversationDTO
	for _, conv := range conversations {
		result = append(result, dto.ConversationDTO{
			ID:        conv.ID().String(),
			SessionID: conv.SessionID().String(),
			Model:     string(conv.Model()),
			IsActive:  conv.IsActive(),
			CreatedAt: conv.CreatedAt(),
			UpdatedAt: conv.UpdatedAt(),
		})
	}

	return result, nil
}

// CloseConversation closes an active conversation
func (s *ConversationService) CloseConversation(ctx context.Context, conversationID string) error {
	id, err := vo.NewConversationID(conversationID)
	if err != nil {
		return err
	}

	conversation, err := s.conversationRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if conversation == nil {
		return ErrInvalidInput
	}

	conversation.Close()

	return s.conversationRepo.Save(ctx, conversation)
}

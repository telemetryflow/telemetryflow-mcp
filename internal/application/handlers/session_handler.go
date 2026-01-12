// Package handlers contains CQRS handlers for the TelemetryFlow GO MCP service
package handlers

import (
	"context"
	"errors"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/commands"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/queries"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/repositories"
)

// Common handler errors
var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionAlreadyExists = errors.New("session already exists")
	ErrInvalidCommand       = errors.New("invalid command")
	ErrInvalidQuery         = errors.New("invalid query")
)

// SessionHandler handles session-related commands and queries
type SessionHandler struct {
	sessionRepo    repositories.ISessionRepository
	eventPublisher EventPublisher
}

// EventPublisher is the interface for publishing events
type EventPublisher interface {
	Publish(ctx context.Context, event interface{}) error
}

// NewSessionHandler creates a new SessionHandler
func NewSessionHandler(sessionRepo repositories.ISessionRepository, eventPublisher EventPublisher) *SessionHandler {
	return &SessionHandler{
		sessionRepo:    sessionRepo,
		eventPublisher: eventPublisher,
	}
}

// HandleInitializeSession handles InitializeSessionCommand
func (h *SessionHandler) HandleInitializeSession(ctx context.Context, cmd *commands.InitializeSessionCommand) (*aggregates.Session, error) {
	// Create new session
	session := aggregates.NewSession()

	// Initialize with client info
	clientInfo := &aggregates.ClientInfo{
		Name:    cmd.ClientName,
		Version: cmd.ClientVersion,
	}

	if err := session.Initialize(clientInfo, cmd.ProtocolVersion); err != nil {
		return nil, err
	}

	// Mark as ready
	session.MarkReady()

	// Save session
	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return nil, err
	}

	// Publish events (best-effort, don't fail on publish errors)
	for _, event := range session.Events() {
		_ = h.eventPublisher.Publish(ctx, event)
	}

	return session, nil
}

// HandleCloseSession handles CloseSessionCommand
func (h *SessionHandler) HandleCloseSession(ctx context.Context, cmd *commands.CloseSessionCommand) error {
	session, err := h.sessionRepo.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}

	session.Close()

	if err := h.sessionRepo.Save(ctx, session); err != nil {
		return err
	}

	// Publish events (best-effort, don't fail on publish errors)
	for _, event := range session.Events() {
		_ = h.eventPublisher.Publish(ctx, event)
	}

	return nil
}

// HandleSetLogLevel handles SetLogLevelCommand
func (h *SessionHandler) HandleSetLogLevel(ctx context.Context, cmd *commands.SetLogLevelCommand) error {
	session, err := h.sessionRepo.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}

	if err := session.SetLogLevel(cmd.Level); err != nil {
		return err
	}

	return h.sessionRepo.Save(ctx, session)
}

// HandlePing handles PingCommand
func (h *SessionHandler) HandlePing(ctx context.Context, cmd *commands.PingCommand) error {
	// Verify session exists and is active
	session, err := h.sessionRepo.FindByID(ctx, cmd.SessionID)
	if err != nil {
		return err
	}
	if session == nil {
		return ErrSessionNotFound
	}
	if session.IsClosed() {
		return aggregates.ErrSessionClosed
	}

	return nil
}

// HandleGetSession handles GetSessionQuery
func (h *SessionHandler) HandleGetSession(ctx context.Context, query *queries.GetSessionQuery) (*aggregates.Session, error) {
	session, err := h.sessionRepo.FindByID(ctx, query.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

// HandleListSessions handles ListSessionsQuery
func (h *SessionHandler) HandleListSessions(ctx context.Context, query *queries.ListSessionsQuery) ([]*aggregates.Session, error) {
	if query.ActiveOnly {
		return h.sessionRepo.FindActive(ctx)
	}
	return h.sessionRepo.FindAll(ctx)
}

// SessionStats represents session statistics
type SessionStats struct {
	TotalConversations  int
	ActiveConversations int
	TotalMessages       int
	TotalToolCalls      int
	TotalTokensUsed     int
}

// HandleGetSessionStats handles GetSessionStatsQuery
func (h *SessionHandler) HandleGetSessionStats(ctx context.Context, query *queries.GetSessionStatsQuery) (*SessionStats, error) {
	session, err := h.sessionRepo.FindByID(ctx, query.SessionID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrSessionNotFound
	}

	// Calculate statistics
	stats := &SessionStats{
		TotalConversations:  len(session.ListConversations()),
		ActiveConversations: 0,
		TotalMessages:       0,
	}

	for _, conv := range session.ListConversations() {
		if conv.IsActive() {
			stats.ActiveConversations++
		}
		stats.TotalMessages += conv.MessageCount()
	}

	return stats, nil
}

// Package mocks provides mock implementations for testing.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package mocks

import (
	"context"
	"sync"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/aggregates"
	vo "github.com/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
)

// MockConversationRepository is a mock implementation of the conversation repository
type MockConversationRepository struct {
	mock.Mock
	mu            sync.RWMutex
	conversations map[string]*aggregates.Conversation
}

// NewMockConversationRepository creates a new mock conversation repository
func NewMockConversationRepository() *MockConversationRepository {
	return &MockConversationRepository{
		conversations: make(map[string]*aggregates.Conversation),
	}
}

// Save saves a conversation
func (m *MockConversationRepository) Save(ctx context.Context, conversation *aggregates.Conversation) error {
	args := m.Called(ctx, conversation)
	if args.Error(0) == nil {
		m.mu.Lock()
		m.conversations[conversation.ID().String()] = conversation
		m.mu.Unlock()
	}
	return args.Error(0)
}

// GetByID retrieves a conversation by ID
func (m *MockConversationRepository) GetByID(ctx context.Context, id vo.ConversationID) (*aggregates.Conversation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregates.Conversation), args.Error(1)
}

// Delete deletes a conversation
func (m *MockConversationRepository) Delete(ctx context.Context, id vo.ConversationID) error {
	args := m.Called(ctx, id)
	if args.Error(0) == nil {
		m.mu.Lock()
		delete(m.conversations, id.String())
		m.mu.Unlock()
	}
	return args.Error(0)
}

// ListBySession lists conversations for a session
func (m *MockConversationRepository) ListBySession(ctx context.Context, sessionID vo.SessionID) ([]*aggregates.Conversation, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*aggregates.Conversation), args.Error(1)
}

// GetConversations returns all stored conversations (for testing)
func (m *MockConversationRepository) GetConversations() map[string]*aggregates.Conversation {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.conversations
}

// InMemorySessionRepository is an in-memory implementation for testing
type InMemorySessionRepository struct {
	mu       sync.RWMutex
	sessions map[string]*aggregates.Session
}

// NewInMemorySessionRepository creates a new in-memory session repository
func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		sessions: make(map[string]*aggregates.Session),
	}
}

// Save saves a session
func (r *InMemorySessionRepository) Save(ctx context.Context, session *aggregates.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.ID().String()] = session
	return nil
}

// GetByID retrieves a session by ID
func (r *InMemorySessionRepository) GetByID(ctx context.Context, id vo.SessionID) (*aggregates.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[id.String()]
	if !ok {
		return nil, ErrNotFound
	}
	return session, nil
}

// Delete deletes a session
func (r *InMemorySessionRepository) Delete(ctx context.Context, id vo.SessionID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, id.String())
	return nil
}

// ListActive lists all active sessions
func (r *InMemorySessionRepository) ListActive(ctx context.Context) ([]*aggregates.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var active []*aggregates.Session
	for _, session := range r.sessions {
		if session.State() == aggregates.SessionStateReady {
			active = append(active, session)
		}
	}
	return active, nil
}

// Count returns the number of sessions
func (r *InMemorySessionRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.sessions), nil
}

// InMemoryConversationRepository is an in-memory implementation for testing
type InMemoryConversationRepository struct {
	mu            sync.RWMutex
	conversations map[string]*aggregates.Conversation
}

// NewInMemoryConversationRepository creates a new in-memory conversation repository
func NewInMemoryConversationRepository() *InMemoryConversationRepository {
	return &InMemoryConversationRepository{
		conversations: make(map[string]*aggregates.Conversation),
	}
}

// Save saves a conversation
func (r *InMemoryConversationRepository) Save(ctx context.Context, conversation *aggregates.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conversations[conversation.ID().String()] = conversation
	return nil
}

// GetByID retrieves a conversation by ID
func (r *InMemoryConversationRepository) GetByID(ctx context.Context, id vo.ConversationID) (*aggregates.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conversation, ok := r.conversations[id.String()]
	if !ok {
		return nil, ErrNotFound
	}
	return conversation, nil
}

// Delete deletes a conversation
func (r *InMemoryConversationRepository) Delete(ctx context.Context, id vo.ConversationID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conversations, id.String())
	return nil
}

// ListBySession lists conversations for a session
func (r *InMemoryConversationRepository) ListBySession(ctx context.Context, sessionID vo.SessionID) ([]*aggregates.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*aggregates.Conversation
	for _, conv := range r.conversations {
		if conv.SessionID().String() == sessionID.String() {
			result = append(result, conv)
		}
	}
	return result, nil
}

// ListActive lists all active conversations
func (r *InMemoryConversationRepository) ListActive(ctx context.Context) ([]*aggregates.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var active []*aggregates.Conversation
	for _, conv := range r.conversations {
		if conv.Status() == aggregates.ConversationStatusActive {
			active = append(active, conv)
		}
	}
	return active, nil
}

// MockAuditLogger is a mock implementation of an audit logger
type MockAuditLogger struct {
	mock.Mock
	mu     sync.Mutex
	events []AuditEvent
}

// AuditEvent represents an audit log event
type AuditEvent struct {
	Timestamp time.Time
	SessionID string
	Action    string
	Resource  string
	Details   map[string]interface{}
}

// NewMockAuditLogger creates a new mock audit logger
func NewMockAuditLogger() *MockAuditLogger {
	return &MockAuditLogger{
		events: make([]AuditEvent, 0),
	}
}

// Log logs an audit event
func (m *MockAuditLogger) Log(ctx context.Context, sessionID, action, resource string, details map[string]interface{}) error {
	args := m.Called(ctx, sessionID, action, resource, details)

	m.mu.Lock()
	m.events = append(m.events, AuditEvent{
		Timestamp: time.Now(),
		SessionID: sessionID,
		Action:    action,
		Resource:  resource,
		Details:   details,
	})
	m.mu.Unlock()

	return args.Error(0)
}

// GetEvents returns all logged events
func (m *MockAuditLogger) GetEvents() []AuditEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.events
}

// ClearEvents clears all logged events
func (m *MockAuditLogger) ClearEvents() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = make([]AuditEvent, 0)
}

// Common errors
var (
	ErrNotFound = &notFoundError{message: "not found"}
)

type notFoundError struct {
	message string
}

func (e *notFoundError) Error() string {
	return e.message
}

// IsNotFound returns true if the error is a not found error
func IsNotFound(err error) bool {
	_, ok := err.(*notFoundError)
	return ok
}

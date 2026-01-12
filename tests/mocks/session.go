// Package mocks provides mock implementations for testing.
//
// TelemetryFlow GO MCP Server - Model Context Protocol Server
// Copyright (c) 2024-2026 TelemetryFlow. All rights reserved.
package mocks

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/telemetryflow/telemetryflow-mcp/internal/domain/aggregates"
	vo "github.com/telemetryflow/telemetryflow-mcp/internal/domain/valueobjects"
)

// MockSessionRepository is a mock implementation of the session repository
type MockSessionRepository struct {
	mock.Mock
	sessions map[string]*aggregates.Session
}

// NewMockSessionRepository creates a new mock session repository
func NewMockSessionRepository() *MockSessionRepository {
	return &MockSessionRepository{
		sessions: make(map[string]*aggregates.Session),
	}
}

// Save saves a session
func (m *MockSessionRepository) Save(ctx context.Context, session *aggregates.Session) error {
	args := m.Called(ctx, session)
	if args.Error(0) == nil {
		m.sessions[session.ID().String()] = session
	}
	return args.Error(0)
}

// GetByID retrieves a session by ID
func (m *MockSessionRepository) GetByID(ctx context.Context, id vo.SessionID) (*aggregates.Session, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*aggregates.Session), args.Error(1)
}

// Delete deletes a session
func (m *MockSessionRepository) Delete(ctx context.Context, id vo.SessionID) error {
	args := m.Called(ctx, id)
	if args.Error(0) == nil {
		delete(m.sessions, id.String())
	}
	return args.Error(0)
}

// ListActive lists all active sessions
func (m *MockSessionRepository) ListActive(ctx context.Context) ([]*aggregates.Session, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*aggregates.Session), args.Error(1)
}

// Count returns the number of sessions
func (m *MockSessionRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

// GetSessions returns all stored sessions (for testing)
func (m *MockSessionRepository) GetSessions() map[string]*aggregates.Session {
	return m.sessions
}

// MockSession creates a mock session for testing
func MockSession() *aggregates.Session {
	session := aggregates.NewSession()
	clientInfo := &aggregates.ClientInfo{
		Name:    "MockClient",
		Version: "1.0.0",
	}
	_ = session.Initialize(clientInfo, "2024-11-05")
	session.MarkReady()
	return session
}

// MockSessionWithTools creates a mock session with tools
func MockSessionWithTools(toolCount int) *aggregates.Session {
	session := MockSession()
	for i := 0; i < toolCount; i++ {
		toolName, _ := vo.NewToolName(sprintf("mock_tool_%d", i))
		toolDesc, _ := vo.NewToolDescription(sprintf("Mock tool %d description", i))
		tool := MockTool(toolName.String(), toolDesc.String())
		session.RegisterTool(tool)
	}
	return session
}

// sprintf is a helper to avoid importing fmt in mocks
func sprintf(format string, args ...interface{}) string {
	if len(args) == 0 {
		return format
	}
	// Simple implementation for mock purposes
	result := format
	for i, arg := range args {
		placeholder := "%d"
		if i == 0 {
			switch v := arg.(type) {
			case int:
				result = replaceFirst(result, placeholder, intToString(v))
			case string:
				result = replaceFirst(result, "%s", v)
			}
		}
	}
	return result
}

func replaceFirst(s, old, new string) string {
	for i := 0; i <= len(s)-len(old); i++ {
		if s[i:i+len(old)] == old {
			return s[:i] + new + s[i+len(old):]
		}
	}
	return s
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}

// MockClientInfo returns mock client info
func MockClientInfo() *aggregates.ClientInfo {
	return &aggregates.ClientInfo{
		Name:    "TestClient",
		Version: "1.0.0",
	}
}

// MockSessionStats returns mock session stats
type MockSessionStats struct {
	ActiveSessions int
	TotalSessions  int
	AverageUptime  time.Duration
}

// GetMockSessionStats returns mock session statistics
func GetMockSessionStats() *MockSessionStats {
	return &MockSessionStats{
		ActiveSessions: 5,
		TotalSessions:  100,
		AverageUptime:  30 * time.Minute,
	}
}

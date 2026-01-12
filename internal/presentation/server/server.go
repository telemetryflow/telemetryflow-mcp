// Package server contains the MCP server implementation
package server

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/commands"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/handlers"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/application/queries"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/config"
)

// Server errors
var (
	ErrServerClosed     = errors.New("server closed")
	ErrInvalidTransport = errors.New("invalid transport")
	ErrSessionRequired  = errors.New("session required")
)

// Server represents the MCP server
type Server struct {
	config *config.Config
	logger zerolog.Logger

	// Handlers
	sessionHandler      *handlers.SessionHandler
	toolHandler         *handlers.ToolHandler
	conversationHandler *handlers.ConversationHandler

	// State
	mu             sync.RWMutex
	currentSession *aggregates.Session
	running        bool
	done           chan struct{}

	// I/O
	reader io.Reader
	writer io.Writer
}

// NewServer creates a new MCP server
func NewServer(
	cfg *config.Config,
	logger zerolog.Logger,
	sessionHandler *handlers.SessionHandler,
	toolHandler *handlers.ToolHandler,
	conversationHandler *handlers.ConversationHandler,
) *Server {
	return &Server{
		config:              cfg,
		logger:              logger.With().Str("component", "mcp-server").Logger(),
		sessionHandler:      sessionHandler,
		toolHandler:         toolHandler,
		conversationHandler: conversationHandler,
		done:                make(chan struct{}),
		reader:              os.Stdin,
		writer:              os.Stdout,
	}
}

// SetIO sets custom I/O for the server (useful for testing)
func (s *Server) SetIO(reader io.Reader, writer io.Writer) {
	s.reader = reader
	s.writer = writer
}

// Run starts the MCP server
func (s *Server) Run(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return errors.New("server already running")
	}
	s.running = true
	s.mu.Unlock()

	s.logger.Info().
		Str("transport", s.config.Server.Transport).
		Str("version", s.config.Server.Version).
		Msg("Starting MCP server")

	switch s.config.Server.Transport {
	case "stdio":
		return s.runStdio(ctx)
	default:
		return ErrInvalidTransport
	}
}

// Stop stops the server
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		s.running = false
		close(s.done)
	}
}

// runStdio runs the server using stdio transport
func (s *Server) runStdio(ctx context.Context) error {
	scanner := bufio.NewScanner(s.reader)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // 10MB max message size

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-s.done:
			return ErrServerClosed
		default:
			if !scanner.Scan() {
				if err := scanner.Err(); err != nil {
					s.logger.Error().Err(err).Msg("Scanner error")
					return err
				}
				return io.EOF
			}

			line := scanner.Text()
			if line == "" {
				continue
			}

			s.logger.Debug().Str("request", line).Msg("Received request")

			response, err := s.handleRequest(ctx, []byte(line))
			if err != nil {
				s.logger.Error().Err(err).Msg("Error handling request")
				response = s.createErrorResponse(nil, vo.ErrorCodeInternalError, err.Error())
			}

			if response != nil {
				if err := s.sendResponse(response); err != nil {
					s.logger.Error().Err(err).Msg("Error sending response")
				}
			}
		}
	}
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id,omitempty"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// handleRequest handles a JSON-RPC request
func (s *Server) handleRequest(ctx context.Context, data []byte) (*JSONRPCResponse, error) {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return s.createErrorResponse(nil, vo.ErrorCodeParseError, "Invalid JSON"), nil
	}

	if req.JSONRPC != "2.0" {
		return s.createErrorResponse(req.ID, vo.ErrorCodeInvalidRequest, "Invalid JSON-RPC version"), nil
	}

	s.logger.Debug().
		Str("method", req.Method).
		Interface("id", req.ID).
		Msg("Processing request")

	// Route to appropriate handler
	method := vo.MCPMethod(req.Method)

	// Handle notifications (no response expected)
	if method.IsNotification() {
		s.handleNotification(ctx, method, req.Params)
		return nil, nil
	}

	// Handle regular methods
	result, err := s.dispatchMethod(ctx, method, req.Params)
	if err != nil {
		if mcpErr, ok := err.(*MCPError); ok {
			return s.createErrorResponse(req.ID, mcpErr.Code, mcpErr.Message), nil
		}
		return s.createErrorResponse(req.ID, vo.ErrorCodeInternalError, err.Error()), nil
	}

	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}, nil
}

// MCPError represents an MCP-specific error
type MCPError struct {
	Code    vo.MCPErrorCode
	Message string
}

func (e *MCPError) Error() string {
	return e.Message
}

// dispatchMethod dispatches a method to the appropriate handler
func (s *Server) dispatchMethod(ctx context.Context, method vo.MCPMethod, params json.RawMessage) (interface{}, error) {
	switch method {
	case vo.MethodInitialize:
		return s.handleInitialize(ctx, params)
	case vo.MethodPing:
		return s.handlePing(ctx)
	case vo.MethodToolsList:
		return s.handleToolsList(ctx, params)
	case vo.MethodToolsCall:
		return s.handleToolsCall(ctx, params)
	case vo.MethodResourcesList:
		return s.handleResourcesList(ctx, params)
	case vo.MethodResourcesRead:
		return s.handleResourcesRead(ctx, params)
	case vo.MethodPromptsList:
		return s.handlePromptsList(ctx, params)
	case vo.MethodPromptsGet:
		return s.handlePromptsGet(ctx, params)
	case vo.MethodLoggingSetLevel:
		return s.handleLoggingSetLevel(ctx, params)
	case vo.MethodCompletionComplete:
		return s.handleCompletionComplete(ctx, params)
	default:
		return nil, &MCPError{Code: vo.ErrorCodeMethodNotFound, Message: "Method not found"}
	}
}

// handleNotification handles notifications
func (s *Server) handleNotification(ctx context.Context, method vo.MCPMethod, params json.RawMessage) {
	switch method {
	case vo.MethodInitialized:
		s.logger.Info().Msg("Client initialized")
	case vo.MethodNotificationsCancelled:
		s.logger.Debug().Msg("Request cancelled")
	default:
		s.logger.Debug().Str("method", method.String()).Msg("Unknown notification")
	}
}

// InitializeParams represents initialize request parameters
type InitializeParams struct {
	ProtocolVersion string                 `json:"protocolVersion"`
	Capabilities    map[string]interface{} `json:"capabilities"`
	ClientInfo      ClientInfo             `json:"clientInfo"`
}

// ClientInfo represents client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// handleInitialize handles the initialize request
func (s *Server) handleInitialize(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p InitializeParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &MCPError{Code: vo.ErrorCodeInvalidParams, Message: "Invalid params"}
	}

	cmd := &commands.InitializeSessionCommand{
		ClientName:      p.ClientInfo.Name,
		ClientVersion:   p.ClientInfo.Version,
		ProtocolVersion: p.ProtocolVersion,
		Capabilities:    p.Capabilities,
	}

	session, err := s.sessionHandler.HandleInitializeSession(ctx, cmd)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.currentSession = session
	s.mu.Unlock()

	s.logger.Info().
		Str("session_id", session.ID().String()).
		Str("client", p.ClientInfo.Name).
		Msg("Session initialized")

	return session.ToInitializeResult(), nil
}

// handlePing handles the ping request
func (s *Server) handlePing(ctx context.Context) (interface{}, error) {
	return map[string]interface{}{}, nil
}

// handleToolsList handles tools/list request
func (s *Server) handleToolsList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.mu.RLock()
	session := s.currentSession
	s.mu.RUnlock()

	if session == nil {
		return nil, &MCPError{Code: vo.ErrorCodeInternalError, Message: "Session not initialized"}
	}

	query := &queries.ListToolsQuery{
		SessionID:   session.ID(),
		EnabledOnly: true,
	}

	result, err := s.toolHandler.HandleListTools(ctx, query)
	if err != nil {
		return nil, err
	}

	return result.ToMCPToolList(), nil
}

// ToolCallParams represents tools/call request parameters
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// handleToolsCall handles tools/call request
func (s *Server) handleToolsCall(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p ToolCallParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &MCPError{Code: vo.ErrorCodeInvalidParams, Message: "Invalid params"}
	}

	s.mu.RLock()
	session := s.currentSession
	s.mu.RUnlock()

	if session == nil {
		return nil, &MCPError{Code: vo.ErrorCodeInternalError, Message: "Session not initialized"}
	}

	cmd := &commands.ExecuteToolCommand{
		SessionID: session.ID(),
		Name:      p.Name,
		Arguments: p.Arguments,
	}

	result, err := s.toolHandler.HandleExecuteTool(ctx, cmd)
	if err != nil {
		return nil, &MCPError{Code: vo.ErrorCodeToolExecutionError, Message: err.Error()}
	}

	return result, nil
}

// handleResourcesList handles resources/list request
func (s *Server) handleResourcesList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.mu.RLock()
	session := s.currentSession
	s.mu.RUnlock()

	if session == nil {
		return nil, &MCPError{Code: vo.ErrorCodeInternalError, Message: "Session not initialized"}
	}

	resources := session.ListResources()
	result := make([]map[string]interface{}, len(resources))
	for i, r := range resources {
		result[i] = r.ToMCPResource()
	}

	return map[string]interface{}{
		"resources": result,
	}, nil
}

// ResourceReadParams represents resources/read request parameters
type ResourceReadParams struct {
	URI string `json:"uri"`
}

// handleResourcesRead handles resources/read request
func (s *Server) handleResourcesRead(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p ResourceReadParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &MCPError{Code: vo.ErrorCodeInvalidParams, Message: "Invalid params"}
	}

	s.mu.RLock()
	session := s.currentSession
	s.mu.RUnlock()

	if session == nil {
		return nil, &MCPError{Code: vo.ErrorCodeInternalError, Message: "Session not initialized"}
	}

	resource, ok := session.GetResource(p.URI)
	if !ok {
		return nil, &MCPError{Code: vo.ErrorCodeResourceNotFound, Message: "Resource not found"}
	}

	content, err := resource.Read()
	if err != nil {
		return nil, &MCPError{Code: vo.ErrorCodeResourceReadError, Message: err.Error()}
	}

	return map[string]interface{}{
		"contents": []interface{}{content},
	}, nil
}

// handlePromptsList handles prompts/list request
func (s *Server) handlePromptsList(ctx context.Context, params json.RawMessage) (interface{}, error) {
	s.mu.RLock()
	session := s.currentSession
	s.mu.RUnlock()

	if session == nil {
		return nil, &MCPError{Code: vo.ErrorCodeInternalError, Message: "Session not initialized"}
	}

	prompts := session.ListPrompts()
	result := make([]map[string]interface{}, len(prompts))
	for i, p := range prompts {
		result[i] = p.ToMCPPrompt()
	}

	return map[string]interface{}{
		"prompts": result,
	}, nil
}

// PromptGetParams represents prompts/get request parameters
type PromptGetParams struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments,omitempty"`
}

// handlePromptsGet handles prompts/get request
func (s *Server) handlePromptsGet(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p PromptGetParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &MCPError{Code: vo.ErrorCodeInvalidParams, Message: "Invalid params"}
	}

	s.mu.RLock()
	session := s.currentSession
	s.mu.RUnlock()

	if session == nil {
		return nil, &MCPError{Code: vo.ErrorCodeInternalError, Message: "Session not initialized"}
	}

	prompt, ok := session.GetPrompt(p.Name)
	if !ok {
		return nil, &MCPError{Code: vo.ErrorCodePromptNotFound, Message: "Prompt not found"}
	}

	messages, err := prompt.Generate(p.Arguments)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// LoggingSetLevelParams represents logging/setLevel request parameters
type LoggingSetLevelParams struct {
	Level string `json:"level"`
}

// handleLoggingSetLevel handles logging/setLevel request
func (s *Server) handleLoggingSetLevel(ctx context.Context, params json.RawMessage) (interface{}, error) {
	var p LoggingSetLevelParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, &MCPError{Code: vo.ErrorCodeInvalidParams, Message: "Invalid params"}
	}

	s.mu.RLock()
	session := s.currentSession
	s.mu.RUnlock()

	if session == nil {
		return nil, &MCPError{Code: vo.ErrorCodeInternalError, Message: "Session not initialized"}
	}

	level := vo.MCPLogLevel(p.Level)
	if !level.IsValid() {
		return nil, &MCPError{Code: vo.ErrorCodeInvalidParams, Message: "Invalid log level"}
	}

	if err := session.SetLogLevel(level); err != nil {
		return nil, err
	}

	return map[string]interface{}{}, nil
}

// handleCompletionComplete handles completion/complete request
func (s *Server) handleCompletionComplete(ctx context.Context, params json.RawMessage) (interface{}, error) {
	// Completion is optional, return empty completions
	return map[string]interface{}{
		"completion": map[string]interface{}{
			"values":  []string{},
			"hasMore": false,
		},
	}, nil
}

// createErrorResponse creates an error response
func (s *Server) createErrorResponse(id interface{}, code vo.MCPErrorCode, message string) *JSONRPCResponse {
	return &JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    int(code),
			Message: message,
		},
	}
}

// sendResponse sends a response
func (s *Server) sendResponse(response *JSONRPCResponse) error {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}

	s.logger.Debug().Str("response", string(data)).Msg("Sending response")

	_, err = fmt.Fprintf(s.writer, "%s\n", data)
	return err
}

// SendNotification sends a notification to the client
func (s *Server) SendNotification(method vo.MCPMethod, params interface{}) error {
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method.String(),
	}
	if params != nil {
		notification["params"] = params
	}

	data, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(s.writer, "%s\n", data)
	return err
}

// Session returns the current session
func (s *Server) Session() *aggregates.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentSession
}

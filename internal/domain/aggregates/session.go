// Package aggregates contains domain aggregates for the TelemetryFlow GO MCP service
package aggregates

import (
	"errors"
	"sync"
	"time"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/events"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// Session errors
var (
	ErrSessionNotFound        = errors.New("session not found")
	ErrSessionClosed          = errors.New("session is closed")
	ErrSessionNotInitialized  = errors.New("session not initialized")
	ErrCapabilityNotSupported = errors.New("capability not supported")
)

// SessionState represents the state of an MCP session
type SessionState string

const (
	SessionStateCreated      SessionState = "created"
	SessionStateInitializing SessionState = "initializing"
	SessionStateReady        SessionState = "ready"
	SessionStateClosed       SessionState = "closed"
)

// Session represents an MCP session aggregate
type Session struct {
	mu sync.RWMutex

	id              vo.SessionID
	protocolVersion vo.MCPProtocolVersion
	state           SessionState
	clientInfo      *ClientInfo
	serverInfo      *ServerInfo
	capabilities    *SessionCapabilities
	tools           map[string]*entities.Tool
	resources       map[string]*entities.Resource
	prompts         map[string]*entities.Prompt
	subscriptions   map[string]bool // Resource URI -> subscribed
	conversations   map[string]*Conversation
	logLevel        vo.MCPLogLevel
	createdAt       time.Time
	updatedAt       time.Time
	closedAt        *time.Time
	metadata        map[string]interface{}
	events          []events.DomainEvent
}

// ClientInfo represents information about the MCP client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo represents information about the MCP server
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// SessionCapabilities represents the capabilities negotiated in a session
type SessionCapabilities struct {
	Tools        *ToolsCapability       `json:"tools,omitempty"`
	Resources    *ResourcesCapability   `json:"resources,omitempty"`
	Prompts      *PromptsCapability     `json:"prompts,omitempty"`
	Logging      *LoggingCapability     `json:"logging,omitempty"`
	Experimental map[string]interface{} `json:"experimental,omitempty"`
}

// ToolsCapability represents tools capability
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resources capability
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// PromptsCapability represents prompts capability
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// LoggingCapability represents logging capability
type LoggingCapability struct{}

// NewSession creates a new Session aggregate
func NewSession() *Session {
	now := time.Now().UTC()
	session := &Session{
		id:              vo.GenerateSessionID(),
		protocolVersion: vo.NewMCPProtocolVersion(""),
		state:           SessionStateCreated,
		serverInfo: &ServerInfo{
			Name:    "TelemetryFlow-MCP",
			Version: "1.1.2",
		},
		capabilities: &SessionCapabilities{
			Tools:     &ToolsCapability{ListChanged: true},
			Resources: &ResourcesCapability{Subscribe: true, ListChanged: true},
			Prompts:   &PromptsCapability{ListChanged: true},
			Logging:   &LoggingCapability{},
		},
		tools:         make(map[string]*entities.Tool),
		resources:     make(map[string]*entities.Resource),
		prompts:       make(map[string]*entities.Prompt),
		subscriptions: make(map[string]bool),
		conversations: make(map[string]*Conversation),
		logLevel:      vo.LogLevelInfo,
		createdAt:     now,
		updatedAt:     now,
		metadata:      make(map[string]interface{}),
		events:        make([]events.DomainEvent, 0),
	}

	session.addEvent(events.NewSessionCreatedEvent(session.id))
	return session
}

// ID returns the session ID
func (s *Session) ID() vo.SessionID {
	return s.id
}

// ProtocolVersion returns the protocol version
func (s *Session) ProtocolVersion() vo.MCPProtocolVersion {
	return s.protocolVersion
}

// State returns the session state
func (s *Session) State() SessionState {
	return s.state
}

// ClientInfo returns the client info
func (s *Session) ClientInfo() *ClientInfo {
	return s.clientInfo
}

// ServerInfo returns the server info
func (s *Session) ServerInfo() *ServerInfo {
	return s.serverInfo
}

// Capabilities returns the session capabilities
func (s *Session) Capabilities() *SessionCapabilities {
	return s.capabilities
}

// Initialize initializes the session with client info
func (s *Session) Initialize(clientInfo *ClientInfo, protocolVersion string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != SessionStateCreated {
		return errors.New("session already initialized")
	}

	s.clientInfo = clientInfo
	s.protocolVersion = vo.NewMCPProtocolVersion(protocolVersion)
	s.state = SessionStateInitializing
	s.updatedAt = time.Now().UTC()

	s.addEvent(events.NewSessionInitializedEvent(s.id, clientInfo.Name, clientInfo.Version))
	return nil
}

// MarkReady marks the session as ready
func (s *Session) MarkReady() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == SessionStateInitializing {
		s.state = SessionStateReady
		s.updatedAt = time.Now().UTC()
	}
}

// IsReady returns whether the session is ready
func (s *Session) IsReady() bool {
	return s.state == SessionStateReady
}

// IsClosed returns whether the session is closed
func (s *Session) IsClosed() bool {
	return s.state == SessionStateClosed
}

// Close closes the session
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state != SessionStateClosed {
		s.state = SessionStateClosed
		now := time.Now().UTC()
		s.closedAt = &now
		s.updatedAt = now

		// Close all conversations
		for _, conv := range s.conversations {
			conv.Close()
		}

		s.addEvent(events.NewSessionClosedEvent(s.id))
	}
}

// Tools

// RegisterTool registers a tool
func (s *Session) RegisterTool(tool *entities.Tool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tools[tool.Name().String()] = tool
	s.updatedAt = time.Now().UTC()
	s.addEvent(events.NewToolRegisteredEvent(s.id, tool.Name().String()))
}

// UnregisterTool unregisters a tool
func (s *Session) UnregisterTool(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tools, name)
	s.updatedAt = time.Now().UTC()
}

// GetTool gets a tool by name
func (s *Session) GetTool(name string) (*entities.Tool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tool, ok := s.tools[name]
	return tool, ok
}

// ListTools lists all tools
func (s *Session) ListTools() []*entities.Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tools := make([]*entities.Tool, 0, len(s.tools))
	for _, tool := range s.tools {
		if tool.IsEnabled() {
			tools = append(tools, tool)
		}
	}
	return tools
}

// Resources

// RegisterResource registers a resource
func (s *Session) RegisterResource(resource *entities.Resource) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := resource.URI().String()
	if resource.IsTemplate() {
		key = resource.URITemplate()
	}
	s.resources[key] = resource
	s.updatedAt = time.Now().UTC()
	s.addEvent(events.NewResourceRegisteredEvent(s.id, key))
}

// UnregisterResource unregisters a resource
func (s *Session) UnregisterResource(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.resources, uri)
	delete(s.subscriptions, uri)
	s.updatedAt = time.Now().UTC()
}

// GetResource gets a resource by URI
func (s *Session) GetResource(uri string) (*entities.Resource, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resource, ok := s.resources[uri]
	return resource, ok
}

// ListResources lists all resources
func (s *Session) ListResources() []*entities.Resource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resources := make([]*entities.Resource, 0, len(s.resources))
	for _, resource := range s.resources {
		resources = append(resources, resource)
	}
	return resources
}

// SubscribeResource subscribes to a resource
func (s *Session) SubscribeResource(uri string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.capabilities.Resources == nil || !s.capabilities.Resources.Subscribe {
		return ErrCapabilityNotSupported
	}

	s.subscriptions[uri] = true
	s.updatedAt = time.Now().UTC()
	return nil
}

// UnsubscribeResource unsubscribes from a resource
func (s *Session) UnsubscribeResource(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.subscriptions, uri)
	s.updatedAt = time.Now().UTC()
}

// IsSubscribed checks if subscribed to a resource
func (s *Session) IsSubscribed(uri string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.subscriptions[uri]
}

// Prompts

// RegisterPrompt registers a prompt
func (s *Session) RegisterPrompt(prompt *entities.Prompt) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.prompts[prompt.Name().String()] = prompt
	s.updatedAt = time.Now().UTC()
	s.addEvent(events.NewPromptRegisteredEvent(s.id, prompt.Name().String()))
}

// UnregisterPrompt unregisters a prompt
func (s *Session) UnregisterPrompt(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.prompts, name)
	s.updatedAt = time.Now().UTC()
}

// GetPrompt gets a prompt by name
func (s *Session) GetPrompt(name string) (*entities.Prompt, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prompt, ok := s.prompts[name]
	return prompt, ok
}

// ListPrompts lists all prompts
func (s *Session) ListPrompts() []*entities.Prompt {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prompts := make([]*entities.Prompt, 0, len(s.prompts))
	for _, prompt := range s.prompts {
		prompts = append(prompts, prompt)
	}
	return prompts
}

// Conversations

// CreateConversation creates a new conversation
func (s *Session) CreateConversation(model vo.Model) (*Conversation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.state == SessionStateClosed {
		return nil, ErrSessionClosed
	}

	conv := NewConversation(s.id, model)
	s.conversations[conv.ID().String()] = conv
	s.updatedAt = time.Now().UTC()

	return conv, nil
}

// GetConversation gets a conversation by ID
func (s *Session) GetConversation(id vo.ConversationID) (*Conversation, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conv, ok := s.conversations[id.String()]
	return conv, ok
}

// ListConversations lists all conversations
func (s *Session) ListConversations() []*Conversation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	convs := make([]*Conversation, 0, len(s.conversations))
	for _, conv := range s.conversations {
		convs = append(convs, conv)
	}
	return convs
}

// CloseConversation closes a conversation
func (s *Session) CloseConversation(id vo.ConversationID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv, ok := s.conversations[id.String()]
	if !ok {
		return ErrConversationNotFound
	}

	conv.Close()
	s.updatedAt = time.Now().UTC()
	return nil
}

// Logging

// LogLevel returns the log level
func (s *Session) LogLevel() vo.MCPLogLevel {
	return s.logLevel
}

// SetLogLevel sets the log level
func (s *Session) SetLogLevel(level vo.MCPLogLevel) error {
	if !level.IsValid() {
		return errors.New("invalid log level")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logLevel = level
	s.updatedAt = time.Now().UTC()
	return nil
}

// Timestamps

// CreatedAt returns the creation timestamp
func (s *Session) CreatedAt() time.Time {
	return s.createdAt
}

// UpdatedAt returns the last update timestamp
func (s *Session) UpdatedAt() time.Time {
	return s.updatedAt
}

// ClosedAt returns the closed timestamp
func (s *Session) ClosedAt() *time.Time {
	return s.closedAt
}

// Metadata

// Metadata returns the session metadata
func (s *Session) Metadata() map[string]interface{} {
	return s.metadata
}

// SetMetadata sets a metadata value
func (s *Session) SetMetadata(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metadata[key] = value
	s.updatedAt = time.Now().UTC()
}

// GetMetadata gets a metadata value
func (s *Session) GetMetadata(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.metadata[key]
	return v, ok
}

// Events

// Events returns and clears domain events
func (s *Session) Events() []events.DomainEvent {
	s.mu.Lock()
	defer s.mu.Unlock()
	evts := s.events
	s.events = make([]events.DomainEvent, 0)
	return evts
}

// ClearEvents clears domain events
func (s *Session) ClearEvents() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = make([]events.DomainEvent, 0)
}

// addEvent adds a domain event
func (s *Session) addEvent(event events.DomainEvent) {
	s.events = append(s.events, event)
}

// ToInitializeResult returns the initialize result for MCP
func (s *Session) ToInitializeResult() map[string]interface{} {
	result := map[string]interface{}{
		"protocolVersion": s.protocolVersion.String(),
		"serverInfo":      s.serverInfo,
		"capabilities":    s.capabilities,
	}
	return result
}

// Package persistence contains repository implementations
package persistence

import (
	"context"
	"sync"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/aggregates"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/repositories"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

// InMemorySessionRepository implements ISessionRepository using in-memory storage
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

func (r *InMemorySessionRepository) Save(ctx context.Context, session *aggregates.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.ID().String()] = session
	return nil
}

func (r *InMemorySessionRepository) FindByID(ctx context.Context, id vo.SessionID) (*aggregates.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	session, ok := r.sessions[id.String()]
	if !ok {
		return nil, nil
	}
	return session, nil
}

func (r *InMemorySessionRepository) FindAll(ctx context.Context) ([]*aggregates.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sessions := make([]*aggregates.Session, 0, len(r.sessions))
	for _, session := range r.sessions {
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (r *InMemorySessionRepository) FindActive(ctx context.Context) ([]*aggregates.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	sessions := make([]*aggregates.Session, 0)
	for _, session := range r.sessions {
		if session.IsReady() && !session.IsClosed() {
			sessions = append(sessions, session)
		}
	}
	return sessions, nil
}

func (r *InMemorySessionRepository) Delete(ctx context.Context, id vo.SessionID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, id.String())
	return nil
}

func (r *InMemorySessionRepository) Exists(ctx context.Context, id vo.SessionID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.sessions[id.String()]
	return ok, nil
}

func (r *InMemorySessionRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.sessions), nil
}

// Ensure interface compliance
var _ repositories.ISessionRepository = (*InMemorySessionRepository)(nil)

// InMemoryConversationRepository implements IConversationRepository using in-memory storage
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

func (r *InMemoryConversationRepository) Save(ctx context.Context, conversation *aggregates.Conversation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.conversations[conversation.ID().String()] = conversation
	return nil
}

func (r *InMemoryConversationRepository) FindByID(ctx context.Context, id vo.ConversationID) (*aggregates.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conv, ok := r.conversations[id.String()]
	if !ok {
		return nil, nil
	}
	return conv, nil
}

func (r *InMemoryConversationRepository) FindBySessionID(ctx context.Context, sessionID vo.SessionID) ([]*aggregates.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conversations := make([]*aggregates.Conversation, 0)
	for _, conv := range r.conversations {
		if conv.SessionID().Equals(sessionID) {
			conversations = append(conversations, conv)
		}
	}
	return conversations, nil
}

func (r *InMemoryConversationRepository) FindActive(ctx context.Context) ([]*aggregates.Conversation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	conversations := make([]*aggregates.Conversation, 0)
	for _, conv := range r.conversations {
		if conv.IsActive() {
			conversations = append(conversations, conv)
		}
	}
	return conversations, nil
}

func (r *InMemoryConversationRepository) Delete(ctx context.Context, id vo.ConversationID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.conversations, id.String())
	return nil
}

func (r *InMemoryConversationRepository) Exists(ctx context.Context, id vo.ConversationID) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.conversations[id.String()]
	return ok, nil
}

func (r *InMemoryConversationRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.conversations), nil
}

func (r *InMemoryConversationRepository) CountBySessionID(ctx context.Context, sessionID vo.SessionID) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, conv := range r.conversations {
		if conv.SessionID().Equals(sessionID) {
			count++
		}
	}
	return count, nil
}

var _ repositories.IConversationRepository = (*InMemoryConversationRepository)(nil)

// InMemoryToolRepository implements IToolRepository using in-memory storage
type InMemoryToolRepository struct {
	mu    sync.RWMutex
	tools map[string]*entities.Tool
}

// NewInMemoryToolRepository creates a new in-memory tool repository
func NewInMemoryToolRepository() *InMemoryToolRepository {
	return &InMemoryToolRepository{
		tools: make(map[string]*entities.Tool),
	}
}

func (r *InMemoryToolRepository) Register(ctx context.Context, tool *entities.Tool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name().String()] = tool
	return nil
}

func (r *InMemoryToolRepository) Unregister(ctx context.Context, name vo.ToolName) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.tools, name.String())
	return nil
}

func (r *InMemoryToolRepository) FindByName(ctx context.Context, name vo.ToolName) (*entities.Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name.String()]
	if !ok {
		return nil, nil
	}
	return tool, nil
}

func (r *InMemoryToolRepository) FindAll(ctx context.Context) ([]*entities.Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tools := make([]*entities.Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		tools = append(tools, tool)
	}
	return tools, nil
}

func (r *InMemoryToolRepository) FindByCategory(ctx context.Context, category string) ([]*entities.Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tools := make([]*entities.Tool, 0)
	for _, tool := range r.tools {
		if tool.Category() == category {
			tools = append(tools, tool)
		}
	}
	return tools, nil
}

func (r *InMemoryToolRepository) FindByTag(ctx context.Context, tag string) ([]*entities.Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tools := make([]*entities.Tool, 0)
	for _, tool := range r.tools {
		for _, t := range tool.Tags() {
			if t == tag {
				tools = append(tools, tool)
				break
			}
		}
	}
	return tools, nil
}

func (r *InMemoryToolRepository) FindEnabled(ctx context.Context) ([]*entities.Tool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tools := make([]*entities.Tool, 0)
	for _, tool := range r.tools {
		if tool.IsEnabled() {
			tools = append(tools, tool)
		}
	}
	return tools, nil
}

func (r *InMemoryToolRepository) Exists(ctx context.Context, name vo.ToolName) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.tools[name.String()]
	return ok, nil
}

func (r *InMemoryToolRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools), nil
}

var _ repositories.IToolRepository = (*InMemoryToolRepository)(nil)

// InMemoryResourceRepository implements IResourceRepository using in-memory storage
type InMemoryResourceRepository struct {
	mu        sync.RWMutex
	resources map[string]*entities.Resource
}

// NewInMemoryResourceRepository creates a new in-memory resource repository
func NewInMemoryResourceRepository() *InMemoryResourceRepository {
	return &InMemoryResourceRepository{
		resources: make(map[string]*entities.Resource),
	}
}

func (r *InMemoryResourceRepository) Register(ctx context.Context, resource *entities.Resource) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := resource.URI().String()
	if resource.IsTemplate() {
		key = resource.URITemplate()
	}
	r.resources[key] = resource
	return nil
}

func (r *InMemoryResourceRepository) Unregister(ctx context.Context, uri vo.ResourceURI) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.resources, uri.String())
	return nil
}

func (r *InMemoryResourceRepository) FindByURI(ctx context.Context, uri vo.ResourceURI) (*entities.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resource, ok := r.resources[uri.String()]
	if !ok {
		return nil, nil
	}
	return resource, nil
}

func (r *InMemoryResourceRepository) FindAll(ctx context.Context) ([]*entities.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resources := make([]*entities.Resource, 0, len(r.resources))
	for _, resource := range r.resources {
		resources = append(resources, resource)
	}
	return resources, nil
}

func (r *InMemoryResourceRepository) FindTemplates(ctx context.Context) ([]*entities.Resource, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	resources := make([]*entities.Resource, 0)
	for _, resource := range r.resources {
		if resource.IsTemplate() {
			resources = append(resources, resource)
		}
	}
	return resources, nil
}

func (r *InMemoryResourceRepository) Exists(ctx context.Context, uri vo.ResourceURI) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.resources[uri.String()]
	return ok, nil
}

func (r *InMemoryResourceRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.resources), nil
}

var _ repositories.IResourceRepository = (*InMemoryResourceRepository)(nil)

// InMemoryPromptRepository implements IPromptRepository using in-memory storage
type InMemoryPromptRepository struct {
	mu      sync.RWMutex
	prompts map[string]*entities.Prompt
}

// NewInMemoryPromptRepository creates a new in-memory prompt repository
func NewInMemoryPromptRepository() *InMemoryPromptRepository {
	return &InMemoryPromptRepository{
		prompts: make(map[string]*entities.Prompt),
	}
}

func (r *InMemoryPromptRepository) Register(ctx context.Context, prompt *entities.Prompt) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.prompts[prompt.Name().String()] = prompt
	return nil
}

func (r *InMemoryPromptRepository) Unregister(ctx context.Context, name vo.ToolName) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.prompts, name.String())
	return nil
}

func (r *InMemoryPromptRepository) FindByName(ctx context.Context, name vo.ToolName) (*entities.Prompt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	prompt, ok := r.prompts[name.String()]
	if !ok {
		return nil, nil
	}
	return prompt, nil
}

func (r *InMemoryPromptRepository) FindAll(ctx context.Context) ([]*entities.Prompt, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	prompts := make([]*entities.Prompt, 0, len(r.prompts))
	for _, prompt := range r.prompts {
		prompts = append(prompts, prompt)
	}
	return prompts, nil
}

func (r *InMemoryPromptRepository) Exists(ctx context.Context, name vo.ToolName) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.prompts[name.String()]
	return ok, nil
}

func (r *InMemoryPromptRepository) Count(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.prompts), nil
}

var _ repositories.IPromptRepository = (*InMemoryPromptRepository)(nil)

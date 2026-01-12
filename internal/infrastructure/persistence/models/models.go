// Package models provides GORM database models for TelemetryFlow GO MCP
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================================================
// Custom Types
// ============================================================================

// JSONB represents a JSONB field in PostgreSQL
type JSONB map[string]interface{}

// Value returns the JSON encoding for the database driver
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan decodes a JSON value from the database driver
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// JSONBArray represents a JSONB array field in PostgreSQL
type JSONBArray []interface{}

// Value returns the JSON encoding for the database driver
func (j JSONBArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan decodes a JSON value from the database driver
func (j *JSONBArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, j)
}

// StringArray represents a string array stored as JSONB
type StringArray []string

// Value returns the JSON encoding for the database driver
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan decodes a JSON value from the database driver
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, s)
}

// ============================================================================
// Base Model
// ============================================================================

// BaseModel contains common columns for all models
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

// BeforeCreate generates a UUID if not set
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// ============================================================================
// Session Model
// ============================================================================

// Session represents an MCP session in the database
type Session struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ProtocolVersion string     `gorm:"type:varchar(20);not null;default:'2024-11-05'" json:"protocolVersion"`
	State           string     `gorm:"type:varchar(20);not null;default:'created'" json:"state"`
	ClientName      string     `gorm:"type:varchar(255)" json:"clientName,omitempty"`
	ClientVersion   string     `gorm:"type:varchar(50)" json:"clientVersion,omitempty"`
	ServerName      string     `gorm:"type:varchar(255);not null;default:'TelemetryFlow-MCP'" json:"serverName"`
	ServerVersion   string     `gorm:"type:varchar(50);not null;default:'1.1.2'" json:"serverVersion"`
	Capabilities    JSONB      `gorm:"type:jsonb;not null;default:'{}'" json:"capabilities"`
	LogLevel        string     `gorm:"type:varchar(20);not null;default:'info'" json:"logLevel"`
	Metadata        JSONB      `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt       time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	ClosedAt        *time.Time `json:"closedAt,omitempty"`

	// Relationships
	Conversations []Conversation `gorm:"foreignKey:SessionID;constraint:OnDelete:CASCADE" json:"conversations,omitempty"`
}

// TableName returns the table name for Session
func (Session) TableName() string {
	return "sessions"
}

// BeforeCreate generates a UUID if not set
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// ============================================================================
// Conversation Model
// ============================================================================

// Conversation represents a conversation in the database
type Conversation struct {
	ID            uuid.UUID   `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	SessionID     uuid.UUID   `gorm:"type:uuid;not null;index" json:"sessionId"`
	Model         string      `gorm:"type:varchar(100);not null;default:'claude-sonnet-4-20250514'" json:"model"`
	SystemPrompt  string      `gorm:"type:text" json:"systemPrompt,omitempty"`
	Status        string      `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	MaxTokens     int         `gorm:"not null;default:4096" json:"maxTokens"`
	Temperature   float64     `gorm:"type:decimal(3,2);not null;default:1.0" json:"temperature"`
	TopP          float64     `gorm:"type:decimal(3,2);not null;default:1.0" json:"topP"`
	TopK          int         `gorm:"not null;default:0" json:"topK"`
	StopSequences StringArray `gorm:"type:jsonb;not null;default:'[]'" json:"stopSequences"`
	Metadata      JSONB       `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt     time.Time   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
	ClosedAt      *time.Time  `json:"closedAt,omitempty"`

	// Relationships
	Session  Session   `gorm:"foreignKey:SessionID" json:"session,omitempty"`
	Messages []Message `gorm:"foreignKey:ConversationID;constraint:OnDelete:CASCADE" json:"messages,omitempty"`
}

// TableName returns the table name for Conversation
func (Conversation) TableName() string {
	return "conversations"
}

// BeforeCreate generates a UUID if not set
func (c *Conversation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// ============================================================================
// Message Model
// ============================================================================

// Message represents a message in the database
type Message struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ConversationID uuid.UUID  `gorm:"type:uuid;not null;index" json:"conversationId"`
	Role           string     `gorm:"type:varchar(20);not null" json:"role"`
	Content        JSONBArray `gorm:"type:jsonb;not null;default:'[]'" json:"content"`
	TokenCount     int        `gorm:"default:0" json:"tokenCount"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"createdAt"`

	// Relationships
	Conversation Conversation `gorm:"foreignKey:ConversationID" json:"conversation,omitempty"`
}

// TableName returns the table name for Message
func (Message) TableName() string {
	return "messages"
}

// BeforeCreate generates a UUID if not set
func (m *Message) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// ============================================================================
// Tool Model
// ============================================================================

// Tool represents a tool in the database
type Tool struct {
	ID             uuid.UUID   `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name           string      `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Description    string      `gorm:"type:text;not null" json:"description"`
	InputSchema    JSONB       `gorm:"type:jsonb;not null;default:'{}'" json:"inputSchema"`
	Category       string      `gorm:"type:varchar(100)" json:"category,omitempty"`
	Tags           StringArray `gorm:"type:jsonb;not null;default:'[]'" json:"tags"`
	IsEnabled      bool        `gorm:"not null;default:true" json:"isEnabled"`
	RateLimit      JSONB       `gorm:"type:jsonb" json:"rateLimit,omitempty"`
	TimeoutSeconds int         `gorm:"not null;default:30" json:"timeoutSeconds"`
	Metadata       JSONB       `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt      time.Time   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName returns the table name for Tool
func (Tool) TableName() string {
	return "tools"
}

// BeforeCreate generates a UUID if not set
func (t *Tool) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// ============================================================================
// Resource Model
// ============================================================================

// Resource represents a resource in the database
type Resource struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	URI         string    `gorm:"type:varchar(2048);not null;uniqueIndex" json:"uri"`
	URITemplate string    `gorm:"type:varchar(2048)" json:"uriTemplate,omitempty"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	MimeType    string    `gorm:"type:varchar(255)" json:"mimeType,omitempty"`
	IsTemplate  bool      `gorm:"not null;default:false" json:"isTemplate"`
	Metadata    JSONB     `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName returns the table name for Resource
func (Resource) TableName() string {
	return "resources"
}

// BeforeCreate generates a UUID if not set
func (r *Resource) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}

// ============================================================================
// Prompt Model
// ============================================================================

// Prompt represents a prompt in the database
type Prompt struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string     `gorm:"type:varchar(255);not null;uniqueIndex" json:"name"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	Arguments   JSONBArray `gorm:"type:jsonb;not null;default:'[]'" json:"arguments"`
	Template    string     `gorm:"type:text" json:"template,omitempty"`
	Metadata    JSONB      `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName returns the table name for Prompt
func (Prompt) TableName() string {
	return "prompts"
}

// BeforeCreate generates a UUID if not set
func (p *Prompt) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// ============================================================================
// ResourceSubscription Model
// ============================================================================

// ResourceSubscription represents a resource subscription in the database
type ResourceSubscription struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	SessionID    uuid.UUID `gorm:"type:uuid;not null;index" json:"sessionId"`
	ResourceURI  string    `gorm:"type:varchar(2048);not null;index" json:"resourceUri"`
	SubscribedAt time.Time `gorm:"autoCreateTime" json:"subscribedAt"`

	// Relationships
	Session Session `gorm:"foreignKey:SessionID" json:"session,omitempty"`
}

// TableName returns the table name for ResourceSubscription
func (ResourceSubscription) TableName() string {
	return "resource_subscriptions"
}

// BeforeCreate generates a UUID if not set
func (rs *ResourceSubscription) BeforeCreate(tx *gorm.DB) error {
	if rs.ID == uuid.Nil {
		rs.ID = uuid.New()
	}
	return nil
}

// ============================================================================
// ToolExecution Model (for auditing)
// ============================================================================

// ToolExecution represents a tool execution record in the database
type ToolExecution struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	SessionID      *uuid.UUID `gorm:"type:uuid;index" json:"sessionId,omitempty"`
	ConversationID *uuid.UUID `gorm:"type:uuid" json:"conversationId,omitempty"`
	ToolName       string     `gorm:"type:varchar(255);not null;index" json:"toolName"`
	Input          JSONB      `gorm:"type:jsonb;not null;default:'{}'" json:"input"`
	Output         JSONB      `gorm:"type:jsonb" json:"output,omitempty"`
	IsError        bool       `gorm:"not null;default:false;index" json:"isError"`
	ErrorMessage   string     `gorm:"type:text" json:"errorMessage,omitempty"`
	DurationMs     int        `gorm:"" json:"durationMs,omitempty"`
	ExecutedAt     time.Time  `gorm:"autoCreateTime;index" json:"executedAt"`
}

// TableName returns the table name for ToolExecution
func (ToolExecution) TableName() string {
	return "tool_executions"
}

// BeforeCreate generates a UUID if not set
func (te *ToolExecution) BeforeCreate(tx *gorm.DB) error {
	if te.ID == uuid.Nil {
		te.ID = uuid.New()
	}
	return nil
}

// ============================================================================
// APIKey Model
// ============================================================================

// APIKey represents an API key in the database
type APIKey struct {
	ID                 uuid.UUID   `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	KeyHash            string      `gorm:"type:varchar(255);not null;uniqueIndex" json:"-"`
	Name               string      `gorm:"type:varchar(255);not null" json:"name"`
	Description        string      `gorm:"type:text" json:"description,omitempty"`
	Scopes             StringArray `gorm:"type:jsonb;not null;default:'[\"read\", \"write\"]'" json:"scopes"`
	RateLimitPerMinute int         `gorm:"default:60" json:"rateLimitPerMinute"`
	RateLimitPerHour   int         `gorm:"default:1000" json:"rateLimitPerHour"`
	IsActive           bool        `gorm:"not null;default:true;index" json:"isActive"`
	ExpiresAt          *time.Time  `json:"expiresAt,omitempty"`
	LastUsedAt         *time.Time  `json:"lastUsedAt,omitempty"`
	CreatedAt          time.Time   `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt          time.Time   `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName returns the table name for APIKey
func (APIKey) TableName() string {
	return "api_keys"
}

// BeforeCreate generates a UUID if not set
func (ak *APIKey) BeforeCreate(tx *gorm.DB) error {
	if ak.ID == uuid.Nil {
		ak.ID = uuid.New()
	}
	return nil
}

// ============================================================================
// SchemaMigration Model
// ============================================================================

// SchemaMigration tracks applied migrations
type SchemaMigration struct {
	Version   string    `gorm:"type:varchar(255);primary_key" json:"version"`
	AppliedAt time.Time `gorm:"autoCreateTime" json:"appliedAt"`
}

// TableName returns the table name for SchemaMigration
func (SchemaMigration) TableName() string {
	return "schema_migrations"
}

// ============================================================================
// AllModels returns all GORM models for auto-migration
// ============================================================================

// AllModels returns a slice of all model pointers for GORM auto-migration
func AllModels() []interface{} {
	return []interface{}{
		&Session{},
		&Conversation{},
		&Message{},
		&Tool{},
		&Resource{},
		&Prompt{},
		&ResourceSubscription{},
		&ToolExecution{},
		&APIKey{},
		&SchemaMigration{},
	}
}

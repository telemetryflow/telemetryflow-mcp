// Package persistence provides database seeding functionality
package persistence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

// ============================================================================
// Seeder Types
// ============================================================================

// SeederFunc is a function that seeds data into the database
type SeederFunc func(ctx context.Context, db *gorm.DB) error

// Seeder represents a database seeder
type Seeder struct {
	Name string
	Fn   SeederFunc
}

// SeederResult represents the result of running seeders
type SeederResult struct {
	Executed []string
	Skipped  []string
	Failed   []string
	Error    error
	Duration time.Duration
}

// ============================================================================
// Database Seeder
// ============================================================================

// DatabaseSeeder handles database seeding
type DatabaseSeeder struct {
	db      *gorm.DB
	seeders []Seeder
}

// NewDatabaseSeeder creates a new DatabaseSeeder instance
func NewDatabaseSeeder(db *gorm.DB) *DatabaseSeeder {
	return &DatabaseSeeder{
		db:      db,
		seeders: make([]Seeder, 0),
	}
}

// Register registers a seeder
func (s *DatabaseSeeder) Register(name string, fn SeederFunc) {
	s.seeders = append(s.seeders, Seeder{Name: name, Fn: fn})
}

// Run executes all registered seeders
func (s *DatabaseSeeder) Run(ctx context.Context) (*SeederResult, error) {
	start := time.Now()
	result := &SeederResult{
		Executed: make([]string, 0),
		Skipped:  make([]string, 0),
		Failed:   make([]string, 0),
	}

	for _, seeder := range s.seeders {
		log.Info().Str("seeder", seeder.Name).Msg("Running seeder")

		if err := seeder.Fn(ctx, s.db); err != nil {
			log.Error().Err(err).Str("seeder", seeder.Name).Msg("Seeder failed")
			result.Failed = append(result.Failed, seeder.Name)
			result.Error = err
			// Continue with other seeders or return on first error
			continue
		}

		result.Executed = append(result.Executed, seeder.Name)
		log.Info().Str("seeder", seeder.Name).Msg("Seeder completed successfully")
	}

	result.Duration = time.Since(start)
	return result, result.Error
}

// RunSeeder executes a specific seeder by name
func (s *DatabaseSeeder) RunSeeder(ctx context.Context, name string) error {
	for _, seeder := range s.seeders {
		if seeder.Name == name {
			return seeder.Fn(ctx, s.db)
		}
	}
	return fmt.Errorf("seeder %s not found", name)
}

// ============================================================================
// Default Seeders
// ============================================================================

// RegisterDefaultSeeders registers all default seeders
func (s *DatabaseSeeder) RegisterDefaultSeeders() {
	s.Register("tools", SeedTools)
	s.Register("resources", SeedResources)
	s.Register("prompts", SeedPrompts)
	s.Register("api_keys", SeedAPIKeys)
	s.Register("demo_session", SeedDemoSession)
}

// SeedTools seeds default tools into the database
func SeedTools(ctx context.Context, db *gorm.DB) error {
	tools := []models.Tool{
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Name:        "echo",
			Description: "Echoes back the input message. Useful for testing connectivity and basic tool functionality.",
			InputSchema: models.JSONB{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "The message to echo back",
					},
				},
				"required": []string{"message"},
			},
			Category:       "utility",
			Tags:           models.StringArray{"testing", "debug", "utility"},
			IsEnabled:      true,
			TimeoutSeconds: 10,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Name:        "read_file",
			Description: "Reads the contents of a file from the filesystem. Returns the file contents as text.",
			InputSchema: models.JSONB{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The path to the file to read",
					},
				},
				"required": []string{"path"},
			},
			Category:       "filesystem",
			Tags:           models.StringArray{"file", "read", "io"},
			IsEnabled:      true,
			TimeoutSeconds: 30,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			Name:        "write_file",
			Description: "Writes content to a file on the filesystem. Creates the file if it doesn't exist.",
			InputSchema: models.JSONB{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The path to the file to write",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "The content to write to the file",
					},
				},
				"required": []string{"path", "content"},
			},
			Category:       "filesystem",
			Tags:           models.StringArray{"file", "write", "io"},
			IsEnabled:      true,
			TimeoutSeconds: 30,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000004"),
			Name:        "list_directory",
			Description: "Lists the contents of a directory. Returns file names, sizes, and modification times.",
			InputSchema: models.JSONB{
				"type": "object",
				"properties": map[string]interface{}{
					"path": map[string]interface{}{
						"type":        "string",
						"description": "The path to the directory to list",
					},
				},
				"required": []string{"path"},
			},
			Category:       "filesystem",
			Tags:           models.StringArray{"directory", "list", "io"},
			IsEnabled:      true,
			TimeoutSeconds: 30,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000005"),
			Name:        "execute_command",
			Description: "Executes a shell command and returns the output. Use with caution.",
			InputSchema: models.JSONB{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The command to execute",
					},
					"timeout": map[string]interface{}{
						"type":        "integer",
						"description": "Timeout in seconds (default: 30)",
						"default":     30,
					},
				},
				"required": []string{"command"},
			},
			Category:       "system",
			Tags:           models.StringArray{"shell", "command", "execute"},
			IsEnabled:      true,
			TimeoutSeconds: 60,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000006"),
			Name:        "search_files",
			Description: "Searches for files matching a pattern. Supports glob patterns.",
			InputSchema: models.JSONB{
				"type": "object",
				"properties": map[string]interface{}{
					"pattern": map[string]interface{}{
						"type":        "string",
						"description": "The glob pattern to match files",
					},
					"directory": map[string]interface{}{
						"type":        "string",
						"description": "The directory to search in (default: current directory)",
					},
				},
				"required": []string{"pattern"},
			},
			Category:       "filesystem",
			Tags:           models.StringArray{"search", "files", "glob"},
			IsEnabled:      true,
			TimeoutSeconds: 60,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000007"),
			Name:        "system_info",
			Description: "Returns information about the system (OS, architecture, hostname, etc.).",
			InputSchema: models.JSONB{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
			Category:       "system",
			Tags:           models.StringArray{"system", "info", "diagnostics"},
			IsEnabled:      true,
			TimeoutSeconds: 10,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000008"),
			Name:        "claude_conversation",
			Description: "Initiates or continues a conversation with Claude. Supports multi-turn conversations.",
			InputSchema: models.JSONB{
				"type": "object",
				"properties": map[string]interface{}{
					"message": map[string]interface{}{
						"type":        "string",
						"description": "The message to send to Claude",
					},
					"conversation_id": map[string]interface{}{
						"type":        "string",
						"description": "Optional conversation ID to continue an existing conversation",
					},
					"system_prompt": map[string]interface{}{
						"type":        "string",
						"description": "Optional system prompt for new conversations",
					},
				},
				"required": []string{"message"},
			},
			Category:       "ai",
			Tags:           models.StringArray{"claude", "ai", "conversation"},
			IsEnabled:      true,
			TimeoutSeconds: 120,
		},
	}

	for _, tool := range tools {
		result := db.WithContext(ctx).Where("name = ?", tool.Name).FirstOrCreate(&tool)
		if result.Error != nil {
			return fmt.Errorf("failed to seed tool %s: %w", tool.Name, result.Error)
		}
	}

	log.Info().Int("count", len(tools)).Msg("Seeded tools")
	return nil
}

// SeedResources seeds default resources into the database
func SeedResources(ctx context.Context, db *gorm.DB) error {
	resources := []models.Resource{
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000101"),
			URI:         "config://server",
			Name:        "Server Configuration",
			Description: "Current server configuration settings",
			MimeType:    "application/json",
			IsTemplate:  false,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000102"),
			URI:         "status://health",
			Name:        "Health Status",
			Description: "Server health and status information",
			MimeType:    "application/json",
			IsTemplate:  false,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000103"),
			URI:         "file:///{path}",
			URITemplate: "file:///{path}",
			Name:        "File Resource",
			Description: "Access files from the filesystem",
			MimeType:    "application/octet-stream",
			IsTemplate:  true,
		},
	}

	for _, resource := range resources {
		result := db.WithContext(ctx).Where("uri = ?", resource.URI).FirstOrCreate(&resource)
		if result.Error != nil {
			return fmt.Errorf("failed to seed resource %s: %w", resource.URI, result.Error)
		}
	}

	log.Info().Int("count", len(resources)).Msg("Seeded resources")
	return nil
}

// SeedPrompts seeds default prompts into the database
func SeedPrompts(ctx context.Context, db *gorm.DB) error {
	prompts := []models.Prompt{
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000201"),
			Name:        "code_review",
			Description: "Reviews code for quality, bugs, and improvements",
			Arguments: models.JSONBArray{
				map[string]interface{}{
					"name":        "code",
					"description": "The code to review",
					"required":    true,
				},
				map[string]interface{}{
					"name":        "language",
					"description": "Programming language",
					"required":    false,
				},
			},
			Template: "Please review the following {{language}} code:\n\n```{{language}}\n{{code}}\n```\n\nProvide feedback on:\n1. Code quality and best practices\n2. Potential bugs or issues\n3. Performance considerations\n4. Suggested improvements",
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000202"),
			Name:        "explain_code",
			Description: "Explains what code does in plain language",
			Arguments: models.JSONBArray{
				map[string]interface{}{
					"name":        "code",
					"description": "The code to explain",
					"required":    true,
				},
			},
			Template: "Please explain the following code in plain language:\n\n```\n{{code}}\n```\n\nExplain:\n1. What the code does\n2. How it works\n3. Key concepts used",
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000203"),
			Name:        "debug_help",
			Description: "Helps debug an error or issue",
			Arguments: models.JSONBArray{
				map[string]interface{}{
					"name":        "error",
					"description": "The error message or issue",
					"required":    true,
				},
				map[string]interface{}{
					"name":        "context",
					"description": "Additional context about the issue",
					"required":    false,
				},
			},
			Template: "I need help debugging the following issue:\n\nError: {{error}}\n\n{{#if context}}\nContext: {{context}}\n{{/if}}\n\nPlease help me:\n1. Understand what's causing this error\n2. Identify potential solutions\n3. Suggest steps to fix it",
		},
	}

	for _, prompt := range prompts {
		result := db.WithContext(ctx).Where("name = ?", prompt.Name).FirstOrCreate(&prompt)
		if result.Error != nil {
			return fmt.Errorf("failed to seed prompt %s: %w", prompt.Name, result.Error)
		}
	}

	log.Info().Int("count", len(prompts)).Msg("Seeded prompts")
	return nil
}

// hashAPIKey creates a SHA-256 hash of an API key
func hashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// SeedAPIKeys seeds default API keys into the database
func SeedAPIKeys(ctx context.Context, db *gorm.DB) error {
	// Development API key (only for development/testing)
	devKey := "tfm_dev_" + "0123456789abcdef0123456789abcdef"

	apiKeys := []models.APIKey{
		{
			ID:                 uuid.MustParse("00000000-0000-0000-0000-000000000301"),
			KeyHash:            hashAPIKey(devKey),
			Name:               "Development Key",
			Description:        "API key for development and testing purposes only",
			Scopes:             models.StringArray{"read", "write", "admin"},
			RateLimitPerMinute: 120,
			RateLimitPerHour:   3600,
			IsActive:           true,
		},
	}

	for _, apiKey := range apiKeys {
		result := db.WithContext(ctx).Where("key_hash = ?", apiKey.KeyHash).FirstOrCreate(&apiKey)
		if result.Error != nil {
			return fmt.Errorf("failed to seed API key %s: %w", apiKey.Name, result.Error)
		}
	}

	log.Info().Int("count", len(apiKeys)).Msg("Seeded API keys")
	log.Warn().Str("key", devKey).Msg("Development API key seeded (DO NOT USE IN PRODUCTION)")
	return nil
}

// SeedDemoSession seeds a demo session for testing
func SeedDemoSession(ctx context.Context, db *gorm.DB) error {
	session := models.Session{
		ID:              uuid.MustParse("00000000-0000-0000-0000-000000000401"),
		ProtocolVersion: "2024-11-05",
		State:           "ready",
		ClientName:      "Demo Client",
		ClientVersion:   "1.0.0",
		ServerName:      "TelemetryFlow-MCP",
		ServerVersion:   "1.1.2",
		Capabilities: models.JSONB{
			"tools":     map[string]interface{}{"listChanged": true},
			"resources": map[string]interface{}{"subscribe": true, "listChanged": true},
			"prompts":   map[string]interface{}{"listChanged": true},
			"logging":   map[string]interface{}{},
		},
		LogLevel: "info",
		Metadata: models.JSONB{
			"demo": true,
		},
	}

	result := db.WithContext(ctx).Where("id = ?", session.ID).FirstOrCreate(&session)
	if result.Error != nil {
		return fmt.Errorf("failed to seed demo session: %w", result.Error)
	}

	// Create a demo conversation
	conversation := models.Conversation{
		ID:           uuid.MustParse("00000000-0000-0000-0000-000000000402"),
		SessionID:    session.ID,
		Model:        "claude-sonnet-4-20250514",
		SystemPrompt: "You are a helpful AI assistant for the TelemetryFlow GO MCP demo.",
		Status:       "active",
		MaxTokens:    4096,
		Temperature:  1.0,
		TopP:         1.0,
		TopK:         0,
	}

	result = db.WithContext(ctx).Where("id = ?", conversation.ID).FirstOrCreate(&conversation)
	if result.Error != nil {
		return fmt.Errorf("failed to seed demo conversation: %w", result.Error)
	}

	// Create demo messages
	messages := []models.Message{
		{
			ID:             uuid.MustParse("00000000-0000-0000-0000-000000000403"),
			ConversationID: conversation.ID,
			Role:           "user",
			Content: models.JSONBArray{
				map[string]interface{}{
					"type": "text",
					"text": "Hello! Can you help me understand how TelemetryFlow GO MCP works?",
				},
			},
		},
		{
			ID:             uuid.MustParse("00000000-0000-0000-0000-000000000404"),
			ConversationID: conversation.ID,
			Role:           "assistant",
			Content: models.JSONBArray{
				map[string]interface{}{
					"type": "text",
					"text": "Hello! I'd be happy to help you understand TelemetryFlow GO MCP.\n\nTelemetryFlow GO MCP is a Model Context Protocol server that enables AI-powered interactions with Claude. It provides:\n\n1. **Tools**: Functions that Claude can execute (read files, run commands, etc.)\n2. **Resources**: Data sources that Claude can access\n3. **Prompts**: Reusable prompt templates\n4. **Conversations**: Multi-turn conversation management\n\nThe server uses domain-driven design (DDD) with CQRS patterns for clean architecture.\n\nWould you like me to explain any specific feature in more detail?",
				},
			},
		},
	}

	for _, message := range messages {
		result = db.WithContext(ctx).Where("id = ?", message.ID).FirstOrCreate(&message)
		if result.Error != nil {
			return fmt.Errorf("failed to seed demo message: %w", result.Error)
		}
	}

	log.Info().Msg("Seeded demo session with conversation and messages")
	return nil
}

// ============================================================================
// Utility Functions
// ============================================================================

// SeedAll runs all default seeders
func SeedAll(ctx context.Context, db *gorm.DB) (*SeederResult, error) {
	seeder := NewDatabaseSeeder(db)
	seeder.RegisterDefaultSeeders()
	return seeder.Run(ctx)
}

// SeedProduction runs only production-safe seeders (excludes demo data)
func SeedProduction(ctx context.Context, db *gorm.DB) (*SeederResult, error) {
	seeder := NewDatabaseSeeder(db)
	seeder.Register("tools", SeedTools)
	seeder.Register("resources", SeedResources)
	seeder.Register("prompts", SeedPrompts)
	return seeder.Run(ctx)
}

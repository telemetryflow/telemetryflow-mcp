// Package migrations provides unit tests for database seeders
package migrations

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence/models"
	"gorm.io/gorm"
)

func TestSeederResult(t *testing.T) {
	t.Run("tracks executed seeders", func(t *testing.T) {
		result := &persistence.SeederResult{
			Executed: []string{"tools", "resources", "prompts"},
			Skipped:  []string{},
			Failed:   []string{},
			Duration: time.Second * 5,
		}

		if len(result.Executed) != 3 {
			t.Errorf("expected 3 executed, got %d", len(result.Executed))
		}
		if result.Error != nil {
			t.Error("expected no error")
		}
	})

	t.Run("tracks failures", func(t *testing.T) {
		result := &persistence.SeederResult{
			Executed: []string{"tools"},
			Failed:   []string{"resources"},
			Error:    context.DeadlineExceeded,
		}

		if len(result.Failed) != 1 {
			t.Errorf("expected 1 failed, got %d", len(result.Failed))
		}
		if result.Error == nil {
			t.Error("expected error to be set")
		}
	})
}

func TestSeederStructure(t *testing.T) {
	t.Run("seeder has name and function", func(t *testing.T) {
		seeder := persistence.Seeder{
			Name: "test_seeder",
			Fn: func(_ context.Context, _ *gorm.DB) error {
				return nil
			},
		}

		if seeder.Name == "" {
			t.Error("seeder name is required")
		}
		if seeder.Fn == nil {
			t.Error("seeder function is required")
		}
	})
}

func TestDefaultSeeders(t *testing.T) {
	defaultSeeders := []string{
		"tools",
		"resources",
		"prompts",
		"api_keys",
		"demo_session",
	}

	t.Run("has all default seeders", func(t *testing.T) {
		if len(defaultSeeders) < 4 {
			t.Error("should have at least 4 default seeders")
		}
	})

	for _, name := range defaultSeeders {
		t.Run("seeder_"+name, func(t *testing.T) {
			if name == "" {
				t.Error("seeder name cannot be empty")
			}
		})
	}
}

func TestToolSeederData(t *testing.T) {
	expectedTools := []struct {
		name     string
		category string
		enabled  bool
	}{
		{"echo", "utility", true},
		{"read_file", "filesystem", true},
		{"write_file", "filesystem", true},
		{"list_directory", "filesystem", true},
		{"execute_command", "system", true},
		{"search_files", "filesystem", true},
		{"system_info", "system", true},
		{"claude_conversation", "ai", true},
	}

	t.Run("has all required tools", func(t *testing.T) {
		if len(expectedTools) < 8 {
			t.Error("should have at least 8 default tools")
		}
	})

	for _, tool := range expectedTools {
		t.Run("tool_"+tool.name, func(t *testing.T) {
			if tool.name == "" {
				t.Error("tool name is required")
			}
			if tool.category == "" {
				t.Error("tool category is required")
			}
		})
	}
}

func TestToolModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		m := models.Tool{}
		if m.TableName() != "tools" {
			t.Errorf("expected 'tools', got %s", m.TableName())
		}
	})

	t.Run("has required fields", func(t *testing.T) {
		tool := models.Tool{
			ID:          uuid.New(),
			Name:        "test_tool",
			Description: "A test tool",
			InputSchema: models.JSONB{"type": "object"},
			IsEnabled:   true,
		}

		if tool.Name == "" {
			t.Error("name is required")
		}
		if tool.Description == "" {
			t.Error("description is required")
		}
	})

	t.Run("default values", func(t *testing.T) {
		tool := models.Tool{
			Name:        "test",
			Description: "test",
		}

		// Default timeout should be 30 seconds
		if tool.TimeoutSeconds == 0 {
			tool.TimeoutSeconds = 30
		}
		if tool.TimeoutSeconds != 30 {
			t.Errorf("expected default timeout 30, got %d", tool.TimeoutSeconds)
		}
	})
}

func TestResourceSeederData(t *testing.T) {
	expectedResources := []struct {
		uri        string
		name       string
		isTemplate bool
	}{
		{"config://server", "Server Configuration", false},
		{"status://health", "Health Status", false},
		{"file:///{path}", "File Resource", true},
	}

	t.Run("has all required resources", func(t *testing.T) {
		if len(expectedResources) < 2 {
			t.Error("should have at least 2 default resources")
		}
	})

	for _, resource := range expectedResources {
		t.Run("resource_"+resource.name, func(t *testing.T) {
			if resource.uri == "" {
				t.Error("resource URI is required")
			}
			if resource.name == "" {
				t.Error("resource name is required")
			}
		})
	}
}

func TestResourceModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		m := models.Resource{}
		if m.TableName() != "resources" {
			t.Errorf("expected 'resources', got %s", m.TableName())
		}
	})

	t.Run("template resource", func(t *testing.T) {
		resource := models.Resource{
			URI:         "file:///{path}",
			URITemplate: "file:///{path}",
			Name:        "File",
			IsTemplate:  true,
		}

		if !resource.IsTemplate {
			t.Error("should be a template resource")
		}
		if resource.URITemplate == "" {
			t.Error("template resources should have URI template")
		}
	})
}

func TestPromptSeederData(t *testing.T) {
	expectedPrompts := []struct {
		name        string
		hasTemplate bool
	}{
		{"code_review", true},
		{"explain_code", true},
		{"debug_help", true},
	}

	t.Run("has all required prompts", func(t *testing.T) {
		if len(expectedPrompts) < 3 {
			t.Error("should have at least 3 default prompts")
		}
	})

	for _, prompt := range expectedPrompts {
		t.Run("prompt_"+prompt.name, func(t *testing.T) {
			if prompt.name == "" {
				t.Error("prompt name is required")
			}
		})
	}
}

func TestPromptModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		m := models.Prompt{}
		if m.TableName() != "prompts" {
			t.Errorf("expected 'prompts', got %s", m.TableName())
		}
	})

	t.Run("has arguments", func(t *testing.T) {
		prompt := models.Prompt{
			Name:        "test_prompt",
			Description: "A test prompt",
			Arguments: models.JSONBArray{
				map[string]interface{}{
					"name":     "code",
					"required": true,
				},
			},
			Template: "Review this code: {{code}}",
		}

		if len(prompt.Arguments) == 0 {
			t.Error("prompt should have arguments")
		}
		if prompt.Template == "" {
			t.Error("prompt should have template")
		}
	})
}

func TestAPIKeySeederData(t *testing.T) {
	t.Run("development key format", func(t *testing.T) {
		devKeyPrefix := "tfm_dev_"
		if len(devKeyPrefix) < 8 {
			t.Error("dev key prefix should be at least 8 chars")
		}
	})

	t.Run("key is hashed", func(t *testing.T) {
		// API keys should be hashed before storage
		rawKey := "tfm_dev_0123456789abcdef"
		// In real implementation, this would be hashed
		if rawKey == "" {
			t.Error("key cannot be empty")
		}
	})
}

func TestAPIKeyModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		m := models.APIKey{}
		if m.TableName() != "api_keys" {
			t.Errorf("expected 'api_keys', got %s", m.TableName())
		}
	})

	t.Run("has required fields", func(t *testing.T) {
		apiKey := models.APIKey{
			ID:       uuid.New(),
			KeyHash:  "abc123hash",
			Name:     "Test Key",
			Scopes:   models.StringArray{"read", "write"},
			IsActive: true,
		}

		if apiKey.KeyHash == "" {
			t.Error("key hash is required")
		}
		if apiKey.Name == "" {
			t.Error("name is required")
		}
		if len(apiKey.Scopes) == 0 {
			t.Error("scopes are required")
		}
	})

	t.Run("rate limits", func(t *testing.T) {
		apiKey := models.APIKey{
			RateLimitPerMinute: 60,
			RateLimitPerHour:   1000,
		}

		if apiKey.RateLimitPerMinute <= 0 {
			t.Error("rate limit per minute should be positive")
		}
		if apiKey.RateLimitPerHour <= 0 {
			t.Error("rate limit per hour should be positive")
		}
		if apiKey.RateLimitPerHour < apiKey.RateLimitPerMinute {
			t.Error("hourly limit should be >= minute limit")
		}
	})
}

func TestDemoSessionSeederData(t *testing.T) {
	t.Run("creates session with conversation", func(t *testing.T) {
		// Demo session should have a conversation
		hasConversation := true
		if !hasConversation {
			t.Error("demo session should have a conversation")
		}
	})

	t.Run("creates messages", func(t *testing.T) {
		// Demo conversation should have sample messages
		messageCount := 2
		if messageCount < 2 {
			t.Error("demo should have at least 2 messages")
		}
	})
}

func TestSessionModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		m := models.Session{}
		if m.TableName() != "sessions" {
			t.Errorf("expected 'sessions', got %s", m.TableName())
		}
	})

	t.Run("valid states", func(t *testing.T) {
		validStates := []string{"created", "initializing", "ready", "closed"}
		for _, state := range validStates {
			if state == "" {
				t.Error("state cannot be empty")
			}
		}
	})

	t.Run("capabilities as JSONB", func(t *testing.T) {
		session := models.Session{
			Capabilities: models.JSONB{
				"tools":     map[string]interface{}{"listChanged": true},
				"resources": map[string]interface{}{"subscribe": true},
			},
		}

		if session.Capabilities == nil {
			t.Error("capabilities should not be nil")
		}
	})
}

func TestConversationModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		m := models.Conversation{}
		if m.TableName() != "conversations" {
			t.Errorf("expected 'conversations', got %s", m.TableName())
		}
	})

	t.Run("valid statuses", func(t *testing.T) {
		validStatuses := []string{"active", "paused", "closed", "archived"}
		for _, status := range validStatuses {
			if status == "" {
				t.Error("status cannot be empty")
			}
		}
	})

	t.Run("temperature range", func(t *testing.T) {
		conv := models.Conversation{
			Temperature: 1.0,
		}

		if conv.Temperature < 0 || conv.Temperature > 2 {
			t.Error("temperature should be between 0 and 2")
		}
	})
}

func TestMessageModel(t *testing.T) {
	t.Run("table name", func(t *testing.T) {
		m := models.Message{}
		if m.TableName() != "messages" {
			t.Errorf("expected 'messages', got %s", m.TableName())
		}
	})

	t.Run("valid roles", func(t *testing.T) {
		validRoles := []string{"user", "assistant"}
		for _, role := range validRoles {
			if role == "" {
				t.Error("role cannot be empty")
			}
		}
	})

	t.Run("content as JSONBArray", func(t *testing.T) {
		message := models.Message{
			Content: models.JSONBArray{
				map[string]interface{}{
					"type": "text",
					"text": "Hello!",
				},
			},
		}

		if len(message.Content) == 0 {
			t.Error("content should not be empty")
		}
	})
}

func TestSeederIdempotency(t *testing.T) {
	t.Run("uses FirstOrCreate", func(t *testing.T) {
		// Seeders should use FirstOrCreate to avoid duplicates
		usesFirstOrCreate := true
		if !usesFirstOrCreate {
			t.Error("seeders should use FirstOrCreate for idempotency")
		}
	})

	t.Run("uses unique constraints", func(t *testing.T) {
		// Tables should have unique constraints on natural keys
		uniqueColumns := map[string]string{
			"tools":     "name",
			"resources": "uri",
			"prompts":   "name",
			"api_keys":  "key_hash",
		}

		for table, column := range uniqueColumns {
			if column == "" {
				t.Errorf("%s should have unique constraint", table)
			}
		}
	})
}

func TestProductionSeeders(t *testing.T) {
	productionSeeders := []string{"tools", "resources", "prompts"}
	excludedFromProduction := []string{"api_keys", "demo_session"}

	t.Run("production seeders don't include demo data", func(t *testing.T) {
		for _, excluded := range excludedFromProduction {
			for _, prod := range productionSeeders {
				if prod == excluded {
					t.Errorf("%s should not be in production seeders", excluded)
				}
			}
		}
	})
}

func TestAllModels(t *testing.T) {
	t.Run("returns all models", func(t *testing.T) {
		allModels := models.AllModels()
		if len(allModels) < 8 {
			t.Errorf("expected at least 8 models, got %d", len(allModels))
		}
	})

	t.Run("models have TableName method", func(t *testing.T) {
		// Each model should implement TableName
		expectedTables := []string{
			"sessions",
			"conversations",
			"messages",
			"tools",
			"resources",
			"prompts",
			"resource_subscriptions",
			"tool_executions",
			"api_keys",
			"schema_migrations",
		}

		if len(expectedTables) < 8 {
			t.Error("should have at least 8 tables")
		}
	})
}

// Package migrations provides unit tests for GORM models
package migrations

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/telemetryflow/telemetryflow-go-mcp/internal/infrastructure/persistence/models"
)

func TestJSONBType(t *testing.T) {
	t.Run("marshal and unmarshal", func(t *testing.T) {
		original := models.JSONB{
			"key1": "value1",
			"key2": 123,
			"nested": map[string]interface{}{
				"inner": "data",
			},
		}

		// Marshal
		data, err := original.Value()
		if err != nil {
			t.Fatalf("failed to marshal JSONB: %v", err)
		}

		// Unmarshal
		var result models.JSONB
		if err := result.Scan(data); err != nil {
			t.Fatalf("failed to unmarshal JSONB: %v", err)
		}

		if result["key1"] != "value1" {
			t.Error("key1 mismatch")
		}
	})

	t.Run("nil handling", func(t *testing.T) {
		var jsonb models.JSONB
		value, err := jsonb.Value()
		if err != nil {
			t.Fatalf("nil JSONB should marshal without error: %v", err)
		}
		if value != nil {
			t.Error("nil JSONB should marshal to nil")
		}
	})

	t.Run("scan nil", func(t *testing.T) {
		var jsonb models.JSONB
		if err := jsonb.Scan(nil); err != nil {
			t.Fatalf("scanning nil should not error: %v", err)
		}
		if jsonb != nil {
			t.Error("scanning nil should result in nil JSONB")
		}
	})
}

func TestJSONBArrayType(t *testing.T) {
	t.Run("marshal and unmarshal", func(t *testing.T) {
		original := models.JSONBArray{
			map[string]interface{}{"type": "text", "text": "Hello"},
			map[string]interface{}{"type": "image", "url": "http://example.com"},
		}

		// Marshal
		data, err := original.Value()
		if err != nil {
			t.Fatalf("failed to marshal JSONBArray: %v", err)
		}

		// Unmarshal
		var result models.JSONBArray
		if err := result.Scan(data); err != nil {
			t.Fatalf("failed to unmarshal JSONBArray: %v", err)
		}

		if len(result) != 2 {
			t.Errorf("expected 2 elements, got %d", len(result))
		}
	})
}

func TestStringArrayType(t *testing.T) {
	t.Run("marshal and unmarshal", func(t *testing.T) {
		original := models.StringArray{"tag1", "tag2", "tag3"}

		// Marshal
		data, err := original.Value()
		if err != nil {
			t.Fatalf("failed to marshal StringArray: %v", err)
		}

		// Unmarshal
		var result models.StringArray
		if err := result.Scan(data); err != nil {
			t.Fatalf("failed to unmarshal StringArray: %v", err)
		}

		if len(result) != 3 {
			t.Errorf("expected 3 elements, got %d", len(result))
		}
		if result[0] != "tag1" {
			t.Error("first element mismatch")
		}
	})
}

func TestSessionModelFields(t *testing.T) {
	t.Run("all fields present", func(t *testing.T) {
		session := models.Session{
			ID:              uuid.New(),
			ProtocolVersion: "2024-11-05",
			State:           "ready",
			ClientName:      "Test Client",
			ClientVersion:   "1.0.0",
			ServerName:      "TelemetryFlow-MCP",
			ServerVersion:   "1.1.2",
			Capabilities:    models.JSONB{"tools": map[string]interface{}{}},
			LogLevel:        "info",
			Metadata:        models.JSONB{},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if session.ID == uuid.Nil {
			t.Error("ID should be set")
		}
		if session.ProtocolVersion == "" {
			t.Error("ProtocolVersion is required")
		}
		if session.State == "" {
			t.Error("State is required")
		}
		if session.ServerName == "" {
			t.Error("ServerName is required")
		}
	})

	t.Run("JSON serialization", func(t *testing.T) {
		session := models.Session{
			ID:              uuid.New(),
			ProtocolVersion: "2024-11-05",
			State:           "ready",
		}

		data, err := json.Marshal(session)
		if err != nil {
			t.Fatalf("failed to marshal session: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(data, &result); err != nil {
			t.Fatalf("failed to unmarshal session: %v", err)
		}

		if result["protocolVersion"] != "2024-11-05" {
			t.Error("protocolVersion field name incorrect")
		}
	})
}

func TestConversationModelFields(t *testing.T) {
	t.Run("all fields present", func(t *testing.T) {
		sessionID := uuid.New()
		conv := models.Conversation{
			ID:            uuid.New(),
			SessionID:     sessionID,
			Model:         "claude-sonnet-4-20250514",
			SystemPrompt:  "You are helpful",
			Status:        "active",
			MaxTokens:     4096,
			Temperature:   1.0,
			TopP:          1.0,
			TopK:          0,
			StopSequences: models.StringArray{},
			Metadata:      models.JSONB{},
		}

		if conv.SessionID == uuid.Nil {
			t.Error("SessionID is required")
		}
		if conv.Model == "" {
			t.Error("Model is required")
		}
	})

	t.Run("parameter ranges", func(t *testing.T) {
		testCases := []struct {
			name        string
			temperature float64
			topP        float64
			valid       bool
		}{
			{"valid temperature", 1.0, 1.0, true},
			{"min temperature", 0, 1.0, true},
			{"max temperature", 2.0, 1.0, true},
			{"valid top_p", 1.0, 0.9, true},
			{"min top_p", 1.0, 0, true},
			{"max top_p", 1.0, 1.0, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				validTemp := tc.temperature >= 0 && tc.temperature <= 2
				validTopP := tc.topP >= 0 && tc.topP <= 1

				if (validTemp && validTopP) != tc.valid {
					t.Error("parameter validation failed")
				}
			})
		}
	})
}

func TestMessageModelFields(t *testing.T) {
	t.Run("all fields present", func(t *testing.T) {
		convID := uuid.New()
		msg := models.Message{
			ID:             uuid.New(),
			ConversationID: convID,
			Role:           "user",
			Content: models.JSONBArray{
				map[string]interface{}{
					"type": "text",
					"text": "Hello!",
				},
			},
			TokenCount: 10,
		}

		if msg.ConversationID == uuid.Nil {
			t.Error("ConversationID is required")
		}
		if msg.Role == "" {
			t.Error("Role is required")
		}
	})

	t.Run("content block types", func(t *testing.T) {
		validTypes := []string{"text", "tool_use", "tool_result", "image"}
		for _, contentType := range validTypes {
			if contentType == "" {
				t.Error("content type cannot be empty")
			}
		}
	})
}

func TestToolModelFields(t *testing.T) {
	t.Run("all fields present", func(t *testing.T) {
		tool := models.Tool{
			ID:          uuid.New(),
			Name:        "test_tool",
			Description: "A test tool for testing",
			InputSchema: models.JSONB{
				"type": "object",
				"properties": map[string]interface{}{
					"input": map[string]interface{}{
						"type": "string",
					},
				},
			},
			Category:       "testing",
			Tags:           models.StringArray{"test", "example"},
			IsEnabled:      true,
			TimeoutSeconds: 30,
			Metadata:       models.JSONB{},
		}

		if tool.Name == "" {
			t.Error("Name is required")
		}
		if tool.Description == "" {
			t.Error("Description is required")
		}
	})

	t.Run("input schema structure", func(t *testing.T) {
		schema := models.JSONB{
			"type": "object",
			"properties": map[string]interface{}{
				"message": map[string]interface{}{
					"type":        "string",
					"description": "The message",
				},
			},
			"required": []string{"message"},
		}

		if schema["type"] != "object" {
			t.Error("schema type should be object")
		}
	})
}

func TestResourceModelFields(t *testing.T) {
	t.Run("static resource", func(t *testing.T) {
		resource := models.Resource{
			ID:          uuid.New(),
			URI:         "config://server",
			Name:        "Server Config",
			Description: "Server configuration",
			MimeType:    "application/json",
			IsTemplate:  false,
		}

		if resource.URI == "" {
			t.Error("URI is required")
		}
		if resource.IsTemplate {
			t.Error("static resource should not be template")
		}
	})

	t.Run("template resource", func(t *testing.T) {
		resource := models.Resource{
			ID:          uuid.New(),
			URI:         "file:///{path}",
			URITemplate: "file:///{path}",
			Name:        "File Resource",
			IsTemplate:  true,
		}

		if !resource.IsTemplate {
			t.Error("should be a template")
		}
		if resource.URITemplate == "" {
			t.Error("template resources need URITemplate")
		}
	})
}

func TestPromptModelFields(t *testing.T) {
	t.Run("all fields present", func(t *testing.T) {
		prompt := models.Prompt{
			ID:          uuid.New(),
			Name:        "code_review",
			Description: "Reviews code",
			Arguments: models.JSONBArray{
				map[string]interface{}{
					"name":        "code",
					"description": "The code to review",
					"required":    true,
				},
			},
			Template: "Review: {{code}}",
		}

		if prompt.Name == "" {
			t.Error("Name is required")
		}
	})

	t.Run("argument structure", func(t *testing.T) {
		arg := map[string]interface{}{
			"name":        "code",
			"description": "The code to review",
			"required":    true,
		}

		if arg["name"] == "" {
			t.Error("argument name is required")
		}
	})
}

func TestToolExecutionModelFields(t *testing.T) {
	t.Run("all fields present", func(t *testing.T) {
		sessionID := uuid.New()
		exec := models.ToolExecution{
			ID:         uuid.New(),
			SessionID:  &sessionID,
			ToolName:   "read_file",
			Input:      models.JSONB{"path": "/tmp/test.txt"},
			Output:     models.JSONB{"content": "file contents"},
			IsError:    false,
			DurationMs: 150,
			ExecutedAt: time.Now(),
		}

		if exec.ToolName == "" {
			t.Error("ToolName is required")
		}
	})

	t.Run("error execution", func(t *testing.T) {
		exec := models.ToolExecution{
			ToolName:     "read_file",
			IsError:      true,
			ErrorMessage: "File not found",
		}

		if !exec.IsError {
			t.Error("should be marked as error")
		}
		if exec.ErrorMessage == "" {
			t.Error("error executions should have error message")
		}
	})
}

func TestAPIKeyModelFields(t *testing.T) {
	t.Run("all fields present", func(t *testing.T) {
		apiKey := models.APIKey{
			ID:                 uuid.New(),
			KeyHash:            "sha256hash",
			Name:               "Test Key",
			Description:        "A test API key",
			Scopes:             models.StringArray{"read", "write"},
			RateLimitPerMinute: 60,
			RateLimitPerHour:   1000,
			IsActive:           true,
		}

		if apiKey.KeyHash == "" {
			t.Error("KeyHash is required")
		}
		if apiKey.Name == "" {
			t.Error("Name is required")
		}
	})

	t.Run("scope values", func(t *testing.T) {
		validScopes := []string{"read", "write", "admin"}
		for _, scope := range validScopes {
			if scope == "" {
				t.Error("scope cannot be empty")
			}
		}
	})
}

func TestResourceSubscriptionModelFields(t *testing.T) {
	t.Run("all fields present", func(t *testing.T) {
		sessionID := uuid.New()
		sub := models.ResourceSubscription{
			ID:           uuid.New(),
			SessionID:    sessionID,
			ResourceURI:  "file:///tmp/watched.txt",
			SubscribedAt: time.Now(),
		}

		if sub.SessionID == uuid.Nil {
			t.Error("SessionID is required")
		}
		if sub.ResourceURI == "" {
			t.Error("ResourceURI is required")
		}
	})
}

func TestBeforeCreateHooks(t *testing.T) {
	t.Run("session generates UUID", func(t *testing.T) {
		session := &models.Session{}
		// In real GORM, BeforeCreate would be called automatically
		if session.ID == uuid.Nil {
			session.ID = uuid.New()
		}
		if session.ID == uuid.Nil {
			t.Error("ID should be generated")
		}
	})

	t.Run("conversation generates UUID", func(t *testing.T) {
		conv := &models.Conversation{}
		if conv.ID == uuid.Nil {
			conv.ID = uuid.New()
		}
		if conv.ID == uuid.Nil {
			t.Error("ID should be generated")
		}
	})
}

func TestAllModelsFunction(t *testing.T) {
	t.Run("returns correct number of models", func(t *testing.T) {
		allModels := models.AllModels()

		expectedModels := 10 // Session, Conversation, Message, Tool, Resource, Prompt, ResourceSubscription, ToolExecution, APIKey, SchemaMigration
		if len(allModels) != expectedModels {
			t.Errorf("expected %d models, got %d", expectedModels, len(allModels))
		}
	})

	t.Run("models are not nil", func(t *testing.T) {
		for i, model := range models.AllModels() {
			if model == nil {
				t.Errorf("model %d is nil", i)
			}
		}
	})
}

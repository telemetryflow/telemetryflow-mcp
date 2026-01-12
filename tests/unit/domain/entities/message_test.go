// Package entities_test contains unit tests for domain entities
package entities_test

import (
	"testing"

	"github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/entities"
	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestNewMessage(t *testing.T) {
	tests := []struct {
		name        string
		role        vo.Role
		content     []entities.ContentBlock
		wantErr     bool
		expectedErr error
	}{
		{
			name: "valid user message",
			role: vo.RoleUser,
			content: []entities.ContentBlock{
				{Type: vo.ContentTypeText, Text: "Hello"},
			},
			wantErr: false,
		},
		{
			name: "valid assistant message",
			role: vo.RoleAssistant,
			content: []entities.ContentBlock{
				{Type: vo.ContentTypeText, Text: "Hi there!"},
			},
			wantErr: false,
		},
		{
			name:        "invalid role",
			role:        vo.Role("invalid"),
			content:     []entities.ContentBlock{},
			wantErr:     true,
			expectedErr: vo.ErrInvalidRole,
		},
		{
			name:    "empty content is valid",
			role:    vo.RoleUser,
			content: []entities.ContentBlock{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := entities.NewMessage(tt.role, tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewMessage() expected error, got nil")
				}
				if tt.expectedErr != nil && err != tt.expectedErr {
					t.Errorf("NewMessage() error = %v, expectedErr %v", err, tt.expectedErr)
				}
				return
			}
			if err != nil {
				t.Errorf("NewMessage() unexpected error = %v", err)
				return
			}
			if msg.Role() != tt.role {
				t.Errorf("NewMessage() role = %v, want %v", msg.Role(), tt.role)
			}
			if len(msg.Content()) != len(tt.content) {
				t.Errorf("NewMessage() content length = %v, want %v", len(msg.Content()), len(tt.content))
			}
			if msg.ID().String() == "" {
				t.Error("NewMessage() ID should not be empty")
			}
			if msg.CreatedAt().IsZero() {
				t.Error("NewMessage() CreatedAt should not be zero")
			}
		})
	}
}

func TestNewTextMessage(t *testing.T) {
	tests := []struct {
		name    string
		role    vo.Role
		text    string
		wantErr bool
	}{
		{
			name:    "valid user text message",
			role:    vo.RoleUser,
			text:    "Hello, Claude!",
			wantErr: false,
		},
		{
			name:    "valid assistant text message",
			role:    vo.RoleAssistant,
			text:    "Hello! How can I help you?",
			wantErr: false,
		},
		{
			name:    "invalid role",
			role:    vo.Role("invalid"),
			text:    "Hello",
			wantErr: true,
		},
		{
			name:    "empty text is valid",
			role:    vo.RoleUser,
			text:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := entities.NewTextMessage(tt.role, tt.text)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTextMessage() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("NewTextMessage() unexpected error = %v", err)
				return
			}
			if msg.GetTextContent() != tt.text {
				t.Errorf("NewTextMessage() text = %v, want %v", msg.GetTextContent(), tt.text)
			}
		})
	}
}

func TestMessage_GetTextContent(t *testing.T) {
	content := []entities.ContentBlock{
		{Type: vo.ContentTypeText, Text: "Hello "},
		{Type: vo.ContentTypeText, Text: "World"},
		{Type: vo.ContentTypeToolUse, Name: "test_tool"},
	}

	msg, err := entities.NewMessage(vo.RoleAssistant, content)
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}

	text := msg.GetTextContent()
	expected := "Hello World"
	if text != expected {
		t.Errorf("GetTextContent() = %v, want %v", text, expected)
	}
}

func TestMessage_HasToolUse(t *testing.T) {
	tests := []struct {
		name    string
		content []entities.ContentBlock
		want    bool
	}{
		{
			name: "has tool use",
			content: []entities.ContentBlock{
				{Type: vo.ContentTypeText, Text: "Using tool"},
				{Type: vo.ContentTypeToolUse, Name: "test_tool", ID: "123"},
			},
			want: true,
		},
		{
			name: "no tool use",
			content: []entities.ContentBlock{
				{Type: vo.ContentTypeText, Text: "Just text"},
			},
			want: false,
		},
		{
			name:    "empty content",
			content: []entities.ContentBlock{},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := entities.NewMessage(vo.RoleAssistant, tt.content)
			if err != nil {
				t.Fatalf("NewMessage() error = %v", err)
			}
			if got := msg.HasToolUse(); got != tt.want {
				t.Errorf("HasToolUse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_GetToolUseBlocks(t *testing.T) {
	content := []entities.ContentBlock{
		{Type: vo.ContentTypeText, Text: "Using tools"},
		{Type: vo.ContentTypeToolUse, Name: "tool1", ID: "1"},
		{Type: vo.ContentTypeToolUse, Name: "tool2", ID: "2"},
		{Type: vo.ContentTypeText, Text: "More text"},
	}

	msg, err := entities.NewMessage(vo.RoleAssistant, content)
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}

	toolUses := msg.GetToolUseBlocks()
	if len(toolUses) != 2 {
		t.Errorf("GetToolUseBlocks() count = %v, want 2", len(toolUses))
	}
}

func TestMessage_Metadata(t *testing.T) {
	msg, err := entities.NewTextMessage(vo.RoleUser, "Hello")
	if err != nil {
		t.Fatalf("NewTextMessage() error = %v", err)
	}

	// Test SetMetadata and GetMetadata
	msg.SetMetadata("key1", "value1")
	msg.SetMetadata("key2", 42)

	val1, ok1 := msg.GetMetadata("key1")
	if !ok1 || val1 != "value1" {
		t.Errorf("GetMetadata(key1) = %v, %v; want value1, true", val1, ok1)
	}

	val2, ok2 := msg.GetMetadata("key2")
	if !ok2 || val2 != 42 {
		t.Errorf("GetMetadata(key2) = %v, %v; want 42, true", val2, ok2)
	}

	_, ok3 := msg.GetMetadata("nonexistent")
	if ok3 {
		t.Error("GetMetadata(nonexistent) should return false")
	}
}

func TestMessage_IsUserMessage(t *testing.T) {
	userMsg, _ := entities.NewTextMessage(vo.RoleUser, "Hello")
	assistantMsg, _ := entities.NewTextMessage(vo.RoleAssistant, "Hi")

	if !userMsg.IsUserMessage() {
		t.Error("IsUserMessage() should return true for user message")
	}
	if userMsg.IsAssistantMessage() {
		t.Error("IsAssistantMessage() should return false for user message")
	}

	if assistantMsg.IsUserMessage() {
		t.Error("IsUserMessage() should return false for assistant message")
	}
	if !assistantMsg.IsAssistantMessage() {
		t.Error("IsAssistantMessage() should return true for assistant message")
	}
}

func TestMessage_AddContent(t *testing.T) {
	msg, _ := entities.NewTextMessage(vo.RoleUser, "Hello")

	initialLen := len(msg.Content())

	msg.AddContent(entities.ContentBlock{
		Type: vo.ContentTypeText,
		Text: " World",
	})

	if len(msg.Content()) != initialLen+1 {
		t.Errorf("AddContent() content length = %v, want %v", len(msg.Content()), initialLen+1)
	}
}

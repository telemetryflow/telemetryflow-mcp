// Package valueobjects_test contains unit tests for domain value objects
package valueobjects_test

import (
	"strings"
	"testing"

	vo "github.com/telemetryflow/telemetryflow-go-mcp/internal/domain/valueobjects"
)

func TestContentType_IsValid(t *testing.T) {
	tests := []struct {
		name        string
		contentType vo.ContentType
		want        bool
	}{
		{"text is valid", vo.ContentTypeText, true},
		{"image is valid", vo.ContentTypeImage, true},
		{"tool_use is valid", vo.ContentTypeToolUse, true},
		{"tool_result is valid", vo.ContentTypeToolResult, true},
		{"invalid type", vo.ContentType("invalid"), false},
		{"empty type", vo.ContentType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.contentType.IsValid(); got != tt.want {
				t.Errorf("ContentType.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContentType_String(t *testing.T) {
	tests := []struct {
		contentType vo.ContentType
		want        string
	}{
		{vo.ContentTypeText, "text"},
		{vo.ContentTypeImage, "image"},
		{vo.ContentTypeToolUse, "tool_use"},
		{vo.ContentTypeToolResult, "tool_result"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.contentType.String(); got != tt.want {
				t.Errorf("ContentType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRole_IsValid(t *testing.T) {
	tests := []struct {
		name string
		role vo.Role
		want bool
	}{
		{"user is valid", vo.RoleUser, true},
		{"assistant is valid", vo.RoleAssistant, true},
		{"system is valid", vo.RoleSystem, true},
		{"invalid role", vo.Role("invalid"), false},
		{"empty role", vo.Role(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.role.IsValid(); got != tt.want {
				t.Errorf("Role.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRole_String(t *testing.T) {
	tests := []struct {
		role vo.Role
		want string
	}{
		{vo.RoleUser, "user"},
		{vo.RoleAssistant, "assistant"},
		{vo.RoleSystem, "system"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.role.String(); got != tt.want {
				t.Errorf("Role.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		model vo.Model
		want  bool
	}{
		{"claude-opus-4 is valid", vo.ModelClaude4Opus, true},
		{"claude-sonnet-4 is valid", vo.ModelClaude4Sonnet, true},
		{"claude-3-7-sonnet is valid", vo.ModelClaude37Sonnet, true},
		{"claude-3-5-sonnet is valid", vo.ModelClaude35Sonnet, true},
		{"claude-3-5-sonnet-v2 is valid", vo.ModelClaude35SonnetV2, true},
		{"claude-3-5-haiku is valid", vo.ModelClaude35Haiku, true},
		{"claude-3-opus is valid", vo.ModelClaude3Opus, true},
		{"claude-3-sonnet is valid", vo.ModelClaude3Sonnet, true},
		{"claude-3-haiku is valid", vo.ModelClaude3Haiku, true},
		{"invalid model", vo.Model("invalid-model"), false},
		{"empty model", vo.Model(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.IsValid(); got != tt.want {
				t.Errorf("Model.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestModel_String(t *testing.T) {
	model := vo.ModelClaude4Sonnet
	if got := model.String(); got != "claude-sonnet-4-20250514" {
		t.Errorf("Model.String() = %v, want claude-sonnet-4-20250514", got)
	}
}

func TestNewTextContent(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr error
	}{
		{"valid text", "Hello, World!", nil},
		{"valid unicode", "こんにちは世界", nil},
		{"empty string", "", vo.ErrEmptyContent},
		{"whitespace only", "   ", vo.ErrEmptyContent},
		{"text with whitespace", "  Hello  ", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := vo.NewTextContent(tt.value)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewTextContent() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("NewTextContent() unexpected error = %v", err)
				return
			}
			if content.String() != tt.value {
				t.Errorf("TextContent.String() = %v, want %v", content.String(), tt.value)
			}
		})
	}
}

func TestTextContent_TooLong(t *testing.T) {
	longText := strings.Repeat("a", vo.MaxTextContentLength+1)
	_, err := vo.NewTextContent(longText)
	if err != vo.ErrContentTooLong {
		t.Errorf("NewTextContent() error = %v, want ErrContentTooLong", err)
	}
}

func TestTextContent_Methods(t *testing.T) {
	content, err := vo.NewTextContent("Hello, World!")
	if err != nil {
		t.Fatalf("NewTextContent() error = %v", err)
	}

	// Test IsEmpty
	if content.IsEmpty() {
		t.Error("TextContent.IsEmpty() should return false for non-empty content")
	}

	// Test Length
	if content.Length() != 13 {
		t.Errorf("TextContent.Length() = %v, want 13", content.Length())
	}

	// Test Truncate
	truncated := content.Truncate(5)
	if truncated != "Hello..." {
		t.Errorf("TextContent.Truncate(5) = %v, want Hello...", truncated)
	}

	// Test Truncate with larger limit
	notTruncated := content.Truncate(100)
	if notTruncated != "Hello, World!" {
		t.Errorf("TextContent.Truncate(100) = %v, want Hello, World!", notTruncated)
	}
}

func TestNewSystemPrompt(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr error
	}{
		{"valid prompt", "You are a helpful assistant", nil},
		{"empty prompt", "", nil}, // Empty is valid for system prompt
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt, err := vo.NewSystemPrompt(tt.value)
			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("NewSystemPrompt() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("NewSystemPrompt() unexpected error = %v", err)
				return
			}
			if prompt.String() != tt.value {
				t.Errorf("SystemPrompt.String() = %v, want %v", prompt.String(), tt.value)
			}
		})
	}
}

func TestSystemPrompt_TooLong(t *testing.T) {
	longPrompt := strings.Repeat("a", vo.MaxSystemPromptLength+1)
	_, err := vo.NewSystemPrompt(longPrompt)
	if err != vo.ErrContentTooLong {
		t.Errorf("NewSystemPrompt() error = %v, want ErrContentTooLong", err)
	}
}

func TestSystemPrompt_IsEmpty(t *testing.T) {
	emptyPrompt, _ := vo.NewSystemPrompt("")
	if !emptyPrompt.IsEmpty() {
		t.Error("SystemPrompt.IsEmpty() should return true for empty prompt")
	}

	nonEmptyPrompt, _ := vo.NewSystemPrompt("Hello")
	if nonEmptyPrompt.IsEmpty() {
		t.Error("SystemPrompt.IsEmpty() should return false for non-empty prompt")
	}
}

func TestNewMimeType(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"json", "application/json", "application/json"},
		{"plain text", "text/plain", "text/plain"},
		{"empty defaults to plain text", "", vo.MimeTypePlainText},
		{"whitespace defaults to plain text", "   ", vo.MimeTypePlainText},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mimeType, err := vo.NewMimeType(tt.value)
			if err != nil {
				t.Errorf("NewMimeType() error = %v", err)
				return
			}
			if mimeType.String() != tt.expected {
				t.Errorf("MimeType.String() = %v, want %v", mimeType.String(), tt.expected)
			}
		})
	}
}

func TestMimeType_IsText(t *testing.T) {
	tests := []struct {
		mimeType string
		want     bool
	}{
		{vo.MimeTypePlainText, true},
		{vo.MimeTypeMarkdown, true},
		{vo.MimeTypeHTML, true},
		{vo.MimeTypeJSON, true},
		{vo.MimeTypeXML, true},
		{vo.MimeTypePNG, false},
		{vo.MimeTypeJPEG, false},
		{vo.MimeTypePDF, false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			mimeType, _ := vo.NewMimeType(tt.mimeType)
			if got := mimeType.IsText(); got != tt.want {
				t.Errorf("MimeType.IsText() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMimeType_IsImage(t *testing.T) {
	tests := []struct {
		mimeType string
		want     bool
	}{
		{vo.MimeTypePNG, true},
		{vo.MimeTypeJPEG, true},
		{vo.MimeTypeGIF, true},
		{vo.MimeTypeWebP, true},
		{vo.MimeTypePlainText, false},
		{vo.MimeTypeJSON, false},
		{vo.MimeTypePDF, false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			mimeType, _ := vo.NewMimeType(tt.mimeType)
			if got := mimeType.IsImage(); got != tt.want {
				t.Errorf("MimeType.IsImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultModel(t *testing.T) {
	if vo.DefaultModel != vo.ModelClaude4Sonnet {
		t.Errorf("DefaultModel = %v, want %v", vo.DefaultModel, vo.ModelClaude4Sonnet)
	}
}

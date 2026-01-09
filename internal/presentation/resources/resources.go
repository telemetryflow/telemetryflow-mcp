// Package resources provides MCP resource handling for TelemetryFlow MCP Server
package resources

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	vo "github.com/telemetryflow/telemetryflow-mcp/telemetryflow-mcp/internal/domain/valueobjects"
)

// ResourceHandler handles MCP resource operations
type ResourceHandler struct {
	allowedPaths []string
	maxFileSize  int64
}

// NewResourceHandler creates a new ResourceHandler
func NewResourceHandler(allowedPaths []string, maxFileSize int64) *ResourceHandler {
	return &ResourceHandler{
		allowedPaths: allowedPaths,
		maxFileSize:  maxFileSize,
	}
}

// ResourceContent represents the content of a resource
type ResourceContent struct {
	URI      string
	MimeType string
	Text     string
	Blob     []byte
}

// ReadResource reads a resource by URI
func (h *ResourceHandler) ReadResource(ctx context.Context, uri string) (*ResourceContent, error) {
	// Parse URI
	if !strings.HasPrefix(uri, "file://") {
		return nil, fmt.Errorf("unsupported URI scheme: %s", uri)
	}

	path := strings.TrimPrefix(uri, "file://")

	// Validate path is allowed
	if !h.isPathAllowed(path) {
		return nil, ErrPathNotAllowed
	}

	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrResourceNotFound
		}
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Check file size
	if info.Size() > h.maxFileSize {
		return nil, ErrFileTooLarge
	}

	// Read file
	file, err := os.Open(path) //nolint:gosec // G304: path is validated and cleaned before use
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Determine MIME type
	mimeType := h.detectMimeType(path)

	return &ResourceContent{
		URI:      uri,
		MimeType: mimeType,
		Text:     string(data),
	}, nil
}

// ListResources lists available resources
func (h *ResourceHandler) ListResources(ctx context.Context) ([]ResourceInfo, error) {
	var resources []ResourceInfo

	for _, basePath := range h.allowedPaths {
		err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip files we can't access
			}
			if info.IsDir() {
				return nil
			}

			resources = append(resources, ResourceInfo{
				URI:         "file://" + path,
				Name:        info.Name(),
				Description: fmt.Sprintf("File: %s", path),
				MimeType:    h.detectMimeType(path),
			})
			return nil
		})
		if err != nil {
			continue // Skip paths we can't walk
		}
	}

	return resources, nil
}

// ResourceInfo represents metadata about a resource
type ResourceInfo struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MimeType    string `json:"mimeType,omitempty"`
}

// isPathAllowed checks if a path is within allowed directories
func (h *ResourceHandler) isPathAllowed(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	for _, allowed := range h.allowedPaths {
		allowedAbs, err := filepath.Abs(allowed)
		if err != nil {
			continue
		}
		if strings.HasPrefix(absPath, allowedAbs) {
			return true
		}
	}
	return false
}

// detectMimeType detects the MIME type based on file extension
func (h *ResourceHandler) detectMimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	mimeTypes := map[string]string{
		".txt":  vo.MimeTypePlainText,
		".md":   vo.MimeTypeMarkdown,
		".json": vo.MimeTypeJSON,
		".html": vo.MimeTypeHTML,
		".xml":  vo.MimeTypeXML,
		".go":   "text/x-go",
		".py":   "text/x-python",
		".js":   "text/javascript",
		".ts":   "text/typescript",
		".yaml": "text/yaml",
		".yml":  "text/yaml",
		".png":  vo.MimeTypePNG,
		".jpg":  vo.MimeTypeJPEG,
		".jpeg": vo.MimeTypeJPEG,
		".gif":  vo.MimeTypeGIF,
		".webp": vo.MimeTypeWebP,
		".pdf":  vo.MimeTypePDF,
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}
	return vo.MimeTypePlainText
}

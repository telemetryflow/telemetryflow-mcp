// Package resources provides MCP resource handling for TelemetryFlow GO MCP Server
package resources

import "errors"

// Resource errors
var (
	ErrResourceNotFound = errors.New("resource not found")
	ErrPathNotAllowed   = errors.New("path not allowed")
	ErrFileTooLarge     = errors.New("file too large")
	ErrInvalidURI       = errors.New("invalid resource URI")
)

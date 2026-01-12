// Package middleware provides HTTP/MCP middleware components for TelemetryFlow GO MCP Server
package middleware

import "errors"

// Middleware errors
var (
	ErrInternalError     = errors.New("internal server error")
	ErrRequestTimeout    = errors.New("request timeout")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrForbidden         = errors.New("forbidden")
)

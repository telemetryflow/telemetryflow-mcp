// Package middleware provides HTTP/MCP middleware components for TelemetryFlow GO MCP Server
package middleware

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// Middleware represents a middleware function type
type Middleware func(Handler) Handler

// Handler represents a request handler
type Handler func(ctx context.Context, request interface{}) (interface{}, error)

// Chain chains multiple middleware together
func Chain(middlewares ...Middleware) Middleware {
	return func(final Handler) Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// LoggingMiddleware creates a middleware that logs requests and responses
func LoggingMiddleware(logger zerolog.Logger) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			start := time.Now()

			// Log request
			logger.Debug().
				Interface("request", request).
				Msg("Handling request")

			// Call next handler
			response, err := next(ctx, request)

			// Log response
			duration := time.Since(start)
			logEvent := logger.Info().
				Dur("duration", duration)

			if err != nil {
				logEvent.Err(err).Msg("Request failed")
			} else {
				logEvent.Msg("Request completed")
			}

			return response, err
		}
	}
}

// RecoveryMiddleware creates a middleware that recovers from panics
func RecoveryMiddleware(logger zerolog.Logger) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func() {
				if r := recover(); r != nil {
					logger.Error().
						Interface("panic", r).
						Msg("Recovered from panic")
					err = ErrInternalError
				}
			}()
			return next(ctx, request)
		}
	}
}

// TimeoutMiddleware creates a middleware that enforces request timeouts
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next Handler) Handler {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			done := make(chan struct {
				response interface{}
				err      error
			}, 1)

			go func() {
				response, err := next(ctx, request)
				done <- struct {
					response interface{}
					err      error
				}{response, err}
			}()

			select {
			case result := <-done:
				return result.response, result.err
			case <-ctx.Done():
				return nil, ErrRequestTimeout
			}
		}
	}
}

// RateLimitMiddleware creates a middleware that enforces rate limiting
func RateLimitMiddleware(requestsPerMinute int) Middleware {
	limiter := newRateLimiter(requestsPerMinute)

	return func(next Handler) Handler {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			if !limiter.Allow() {
				return nil, ErrRateLimitExceeded
			}
			return next(ctx, request)
		}
	}
}

// rateLimiter is a simple token bucket rate limiter
type rateLimiter struct {
	tokens         int
	maxTokens      int
	refillInterval time.Duration
	lastRefill     time.Time
}

func newRateLimiter(requestsPerMinute int) *rateLimiter {
	return &rateLimiter{
		tokens:         requestsPerMinute,
		maxTokens:      requestsPerMinute,
		refillInterval: time.Minute,
		lastRefill:     time.Now(),
	}
}

func (r *rateLimiter) Allow() bool {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill)

	if elapsed >= r.refillInterval {
		r.tokens = r.maxTokens
		r.lastRefill = now
	}

	if r.tokens > 0 {
		r.tokens--
		return true
	}
	return false
}

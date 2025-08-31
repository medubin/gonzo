package middleware

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// LoggingMiddleware logs request and response information
type LoggingMiddleware struct {
	BaseMiddleware
}

func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

func (m *LoggingMiddleware) BeforeHandler(ctx context.Context, req *MiddlewareRequest, info *RouteInfo) (context.Context, *MiddlewareRequest, error) {
	// Add start time to context for duration calculation
	ctx = context.WithValue(ctx, "middleware_start_time", time.Now())
	log.Printf("→ %s %s (%s.%s)", req.Method, req.Path, info.Server, info.Endpoint)
	return ctx, req, nil
}

func (m *LoggingMiddleware) AfterHandler(ctx context.Context, req *MiddlewareRequest, resp *MiddlewareResponse, info *RouteInfo) (*MiddlewareResponse, error) {
	if startTime, ok := ctx.Value("middleware_start_time").(time.Time); ok {
		duration := time.Since(startTime)
		log.Printf("← %s %s -> %d (%v)", req.Method, req.Path, resp.Status, duration)
	}
	return resp, nil
}

func (m *LoggingMiddleware) OnError(ctx context.Context, req *MiddlewareRequest, err error, info *RouteInfo) (*MiddlewareResponse, error) {
	if startTime, ok := ctx.Value("middleware_start_time").(time.Time); ok {
		duration := time.Since(startTime)
		log.Printf("✗ %s %s -> ERROR (%v): %v", req.Method, req.Path, duration, err)
	}
	return nil, err // Let error propagate
}

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	BaseMiddleware
	RequiredPaths map[string]bool // Paths that require authentication
}

func NewAuthMiddleware(requiredPaths ...string) *AuthMiddleware {
	pathMap := make(map[string]bool)
	for _, path := range requiredPaths {
		pathMap[path] = true
	}
	return &AuthMiddleware{RequiredPaths: pathMap}
}

func (m *AuthMiddleware) BeforeHandler(ctx context.Context, req *MiddlewareRequest, info *RouteInfo) (context.Context, *MiddlewareRequest, error) {
	// Check if this path requires authentication
	if !m.RequiredPaths[req.Path] && len(m.RequiredPaths) > 0 {
		return ctx, req, nil // Path doesn't require auth
	}

	// Extract token from Authorization header
	authHeader, exists := req.Headers["Authorization"]
	if !exists {
		authHeader, exists = req.Headers["authorization"] // Try lowercase
	}

	if !exists || authHeader == "" {
		return ctx, req, fmt.Errorf("missing authorization header")
	}

	// Simple bearer token validation (replace with your auth logic)
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return ctx, req, fmt.Errorf("invalid authorization format, expected 'Bearer <token>'")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return ctx, req, fmt.Errorf("empty authorization token")
	}

	// Mock token validation - replace with your implementation
	user, err := m.validateToken(token)
	if err != nil {
		return ctx, req, fmt.Errorf("invalid token: %v", err)
	}

	// Add user to context
	ctx = context.WithValue(ctx, "user", user)
	ctx = context.WithValue(ctx, "authenticated", true)

	return ctx, req, nil
}

// Mock token validation - replace with your actual implementation
func (m *AuthMiddleware) validateToken(token string) (interface{}, error) {
	// This is a simple mock - replace with real validation
	if token == "valid-token" {
		return map[string]interface{}{
			"id":   "user-123",
			"name": "Test User",
		}, nil
	}
	return nil, fmt.Errorf("token not recognized")
}

// CORSMiddleware handles Cross-Origin Resource Sharing
type CORSMiddleware struct {
	BaseMiddleware
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

func NewCORSMiddleware(origins, methods, headers []string) *CORSMiddleware {
	return &CORSMiddleware{
		AllowedOrigins: origins,
		AllowedMethods: methods,
		AllowedHeaders: headers,
	}
}

func (m *CORSMiddleware) BeforeRouting(req *MiddlewareRequest) (*MiddlewareRequest, error) {
	// CORS preflight requests should be handled early
	if req.Method == "OPTIONS" {
		// This will be handled in AfterHandler to add proper headers
		return req, nil
	}
	return req, nil
}

func (m *CORSMiddleware) AfterHandler(ctx context.Context, req *MiddlewareRequest, resp *MiddlewareResponse, info *RouteInfo) (*MiddlewareResponse, error) {
	if resp.Headers == nil {
		resp.Headers = make(map[string]string)
	}

	// Add CORS headers
	if len(m.AllowedOrigins) > 0 {
		resp.Headers["Access-Control-Allow-Origin"] = strings.Join(m.AllowedOrigins, ", ")
	}
	if len(m.AllowedMethods) > 0 {
		resp.Headers["Access-Control-Allow-Methods"] = strings.Join(m.AllowedMethods, ", ")
	}
	if len(m.AllowedHeaders) > 0 {
		resp.Headers["Access-Control-Allow-Headers"] = strings.Join(m.AllowedHeaders, ", ")
	}

	// Handle preflight requests
	if req.Method == "OPTIONS" {
		resp.Status = 204 // No Content
		resp.Body = nil
	}

	return resp, nil
}

// ErrorHandlerMiddleware provides consistent error handling and formatting
type ErrorHandlerMiddleware struct {
	BaseMiddleware
	IncludeStackTrace bool
}

func NewErrorHandlerMiddleware(includeStackTrace bool) *ErrorHandlerMiddleware {
	return &ErrorHandlerMiddleware{IncludeStackTrace: includeStackTrace}
}

func (m *ErrorHandlerMiddleware) OnError(ctx context.Context, req *MiddlewareRequest, err error, info *RouteInfo) (*MiddlewareResponse, error) {
	// Log the error
	log.Printf("Error in %s.%s: %v", info.Server, info.Endpoint, err)

	// Create standardized error response
	errorResp := map[string]any{
		"error": true,
		"message": err.Error(),
		"endpoint": fmt.Sprintf("%s.%s", info.Server, info.Endpoint),
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Add request ID if available
	if requestID := ctx.Value("request_id"); requestID != nil {
		errorResp["request_id"] = requestID
	}

	// Determine status code based on error type
	status := 500 // Default to internal server error
	if strings.Contains(err.Error(), "missing authorization") || strings.Contains(err.Error(), "invalid token") {
		status = 401
	} else if strings.Contains(err.Error(), "required") {
		status = 400
	}

	return &MiddlewareResponse{
		Status: status,
		Body:   errorResp,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}
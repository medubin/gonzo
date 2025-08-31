package middleware

import (
	"context"
	"net/http"

	"github.com/medubin/gonzo/api/src/cookies"
)

// MiddlewareRequest represents the abstracted request available to middleware
type MiddlewareRequest struct {
	Method     string
	Path       string
	Headers    map[string]string
	Cookies    cookies.Cookies
	Body       any
	Params     any // Query parameters
	PathParams any // URL path parameters
}

// MiddlewareResponse represents the abstracted response that middleware can modify
type MiddlewareResponse struct {
	Body    any
	Status  int
	Headers map[string]string
}

// RouteInfo provides information about the matched route
type RouteInfo struct {
	Method       string
	Path         string
	Endpoint     string // "CreateUser"
	Server       string // "UserService"
	BodyType     string // "CreateUserRequest"
	ReturnType   string // "User"
	RequiresBody bool   // Whether this endpoint requires a request body
}

// Middleware defines the interface for request/response middleware
type Middleware interface {
	// BeforeRouting is called before any routing occurs
	// Return modified request or error to stop processing
	BeforeRouting(req *MiddlewareRequest) (*MiddlewareRequest, error)

	// BeforeHandler is called after routing, before the handler
	// Can modify both context and request
	BeforeHandler(ctx context.Context, req *MiddlewareRequest, info *RouteInfo) (context.Context, *MiddlewareRequest, error)

	// AfterHandler is called after successful handler execution
	// Can modify the response
	AfterHandler(ctx context.Context, req *MiddlewareRequest, resp *MiddlewareResponse, info *RouteInfo) (*MiddlewareResponse, error)

	// OnError is called only when handler returns an error
	// Can provide custom error response or let error propagate
	OnError(ctx context.Context, req *MiddlewareRequest, err error, info *RouteInfo) (*MiddlewareResponse, error)
}

// BaseMiddleware provides default implementations for all middleware methods
// Embed this in your middleware to only implement the methods you need
type BaseMiddleware struct{}

func (m *BaseMiddleware) BeforeRouting(req *MiddlewareRequest) (*MiddlewareRequest, error) {
	return req, nil
}

func (m *BaseMiddleware) BeforeHandler(ctx context.Context, req *MiddlewareRequest, info *RouteInfo) (context.Context, *MiddlewareRequest, error) {
	return ctx, req, nil
}

func (m *BaseMiddleware) AfterHandler(ctx context.Context, req *MiddlewareRequest, resp *MiddlewareResponse, info *RouteInfo) (*MiddlewareResponse, error) {
	return resp, nil
}

func (m *BaseMiddleware) OnError(ctx context.Context, req *MiddlewareRequest, err error, info *RouteInfo) (*MiddlewareResponse, error) {
	return nil, err // Let error propagate
}

// ConvertHeadersFromHTTP converts http.Header to map[string]string
func ConvertHeadersFromHTTP(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			result[key] = values[0] // Take first value
		}
	}
	return result
}

// ConvertHeadersToHTTP converts map[string]string to http.Header
func ConvertHeadersToHTTP(headers map[string]string) http.Header {
	result := make(http.Header)
	for key, value := range headers {
		result.Set(key, value)
	}
	return result
}
package middleware

import (
	"context"
	"strings"

	"github.com/medubin/gonzo/runtime/types"
)

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

func (m *CORSMiddleware) AfterHandler(ctx context.Context, req *MiddlewareRequest, resp *MiddlewareResponse, info *types.RouteInfo) (*MiddlewareResponse, error) {
	if resp.Headers == nil {
		resp.Headers = make(map[string]string)
	}

	// Match the request's Origin against the allowed list and echo back only
	// the matching value. Access-Control-Allow-Origin must be a single origin
	// or "*" — joining multiple values into one header is not valid per spec.
	requestOrigin := req.Headers["Origin"]
	allowedOrigin := ""
	for _, origin := range m.AllowedOrigins {
		if origin == "*" {
			allowedOrigin = "*"
			break
		}
		if origin == requestOrigin {
			allowedOrigin = requestOrigin
			break
		}
	}

	if allowedOrigin != "" {
		resp.Headers["Access-Control-Allow-Origin"] = allowedOrigin
		// Credentials require a specific origin — not allowed with "*"
		if allowedOrigin != "*" {
			resp.Headers["Access-Control-Allow-Credentials"] = "true"
		}
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

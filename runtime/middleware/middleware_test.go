package middleware_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/medubin/gonzo/runtime/middleware"
	"github.com/medubin/gonzo/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- BaseMiddleware ---

type baseOnlyMiddleware struct {
	middleware.BaseMiddleware
}

func TestBaseMiddleware_BeforeRouting(t *testing.T) {
	m := &baseOnlyMiddleware{}
	req := &middleware.MiddlewareRequest{Method: "GET", Path: "/"}
	result, err := m.BeforeRouting(req)
	require.NoError(t, err)
	assert.Equal(t, req, result)
}

func TestBaseMiddleware_BeforeHandler(t *testing.T) {
	m := &baseOnlyMiddleware{}
	ctx := context.Background()
	req := &middleware.MiddlewareRequest{Method: "GET", Path: "/"}
	info := &types.RouteInfo{}

	outCtx, outReq, err := m.BeforeHandler(ctx, req, info)
	require.NoError(t, err)
	assert.Equal(t, ctx, outCtx)
	assert.Equal(t, req, outReq)
}

func TestBaseMiddleware_AfterHandler(t *testing.T) {
	m := &baseOnlyMiddleware{}
	ctx := context.Background()
	req := &middleware.MiddlewareRequest{}
	resp := &middleware.MiddlewareResponse{Status: 200}
	info := &types.RouteInfo{}

	outResp, err := m.AfterHandler(ctx, req, resp, info)
	require.NoError(t, err)
	assert.Equal(t, resp, outResp)
}

func TestBaseMiddleware_OnError(t *testing.T) {
	m := &baseOnlyMiddleware{}
	ctx := context.Background()
	req := &middleware.MiddlewareRequest{}
	info := &types.RouteInfo{}
	inputErr := assert.AnError

	outResp, err := m.OnError(ctx, req, inputErr, info)
	assert.Nil(t, outResp)
	assert.Equal(t, inputErr, err)
}

// --- Header conversion ---

func TestConvertHeadersFromHTTP(t *testing.T) {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("Authorization", "Bearer token")

	result := middleware.ConvertHeadersFromHTTP(h)

	assert.Equal(t, "application/json", result["Content-Type"])
	assert.Equal(t, "Bearer token", result["Authorization"])
}

func TestConvertHeadersFromHTTP_NormalizesKeys(t *testing.T) {
	// http.Header stores canonical keys, but be explicit that the output is always canonical
	// so CORS and other middleware can reliably do map lookups like req.Headers["Origin"].
	h := http.Header{}
	h["origin"] = []string{"https://example.com"} // bypass Set() to insert a non-canonical key
	h["content-type"] = []string{"application/json"}

	result := middleware.ConvertHeadersFromHTTP(h)

	assert.Equal(t, "https://example.com", result["Origin"], "non-canonical 'origin' should be stored as 'Origin'")
	assert.Equal(t, "application/json", result["Content-Type"], "non-canonical 'content-type' should be stored as 'Content-Type'")
	assert.Empty(t, result["origin"], "non-canonical key should not appear in result")
}

func TestConvertHeadersFromHTTP_Empty(t *testing.T) {
	result := middleware.ConvertHeadersFromHTTP(http.Header{})
	assert.Empty(t, result)
}

func TestConvertHeadersToHTTP(t *testing.T) {
	m := map[string]string{
		"Content-Type": "application/json",
		"X-Custom":     "value",
	}
	result := middleware.ConvertHeadersToHTTP(m)
	assert.Equal(t, "application/json", result.Get("Content-Type"))
	assert.Equal(t, "value", result.Get("X-Custom"))
}

func TestConvertHeadersToHTTP_Empty(t *testing.T) {
	result := middleware.ConvertHeadersToHTTP(map[string]string{})
	assert.Empty(t, result)
}

func TestConvertHeadersRoundTrip(t *testing.T) {
	original := http.Header{}
	original.Set("Accept", "application/json")
	original.Set("X-Request-ID", "abc123")

	converted := middleware.ConvertHeadersFromHTTP(original)
	restored := middleware.ConvertHeadersToHTTP(converted)

	assert.Equal(t, original.Get("Accept"), restored.Get("Accept"))
	assert.Equal(t, original.Get("X-Request-ID"), restored.Get("X-Request-ID"))
}

// --- RequireBodyMiddleware ---

func TestRequireBody_NilBody_ReturnsError(t *testing.T) {
	m := middleware.NewRequireBody()
	ctx := context.Background()
	req := &middleware.MiddlewareRequest{Body: nil}
	info := &types.RouteInfo{}

	_, _, err := m.BeforeHandler(ctx, req, info)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "body")
}

func TestRequireBody_WithBody_PassesThrough(t *testing.T) {
	m := middleware.NewRequireBody()
	ctx := context.Background()
	req := &middleware.MiddlewareRequest{Body: map[string]any{"key": "val"}}
	info := &types.RouteInfo{}

	outCtx, outReq, err := m.BeforeHandler(ctx, req, info)
	require.NoError(t, err)
	assert.Equal(t, ctx, outCtx)
	assert.Equal(t, req, outReq)
}

// --- CORSMiddleware ---

func TestCORSMiddleware_AfterHandler_AddsHeaders(t *testing.T) {
	m := middleware.NewCORSMiddleware(
		[]string{"https://example.com"},
		[]string{"GET", "POST"},
		[]string{"Content-Type", "Authorization"},
	)
	ctx := context.Background()
	req := &middleware.MiddlewareRequest{
		Method:  "GET",
		Headers: map[string]string{"Origin": "https://example.com"},
	}
	resp := &middleware.MiddlewareResponse{Status: 200, Headers: map[string]string{}}
	info := &types.RouteInfo{}

	outResp, err := m.AfterHandler(ctx, req, resp, info)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com", outResp.Headers["Access-Control-Allow-Origin"])
	assert.Equal(t, "GET, POST", outResp.Headers["Access-Control-Allow-Methods"])
	assert.Equal(t, "Content-Type, Authorization", outResp.Headers["Access-Control-Allow-Headers"])
	assert.Equal(t, "true", outResp.Headers["Access-Control-Allow-Credentials"])
}

func TestCORSMiddleware_AfterHandler_InitializesNilHeaders(t *testing.T) {
	m := middleware.NewCORSMiddleware([]string{"*"}, nil, nil)
	ctx := context.Background()
	req := &middleware.MiddlewareRequest{Method: "GET", Headers: map[string]string{}}
	resp := &middleware.MiddlewareResponse{Status: 200, Headers: nil}
	info := &types.RouteInfo{}

	outResp, err := m.AfterHandler(ctx, req, resp, info)
	require.NoError(t, err)
	assert.NotNil(t, outResp.Headers)
	assert.Equal(t, "*", outResp.Headers["Access-Control-Allow-Origin"])
}

func TestCORSMiddleware_Preflight_Sets204(t *testing.T) {
	m := middleware.NewCORSMiddleware([]string{"*"}, []string{"GET"}, nil)
	ctx := context.Background()
	req := &middleware.MiddlewareRequest{Method: "OPTIONS", Headers: map[string]string{}}
	resp := &middleware.MiddlewareResponse{Status: 200, Body: "should be cleared"}
	info := &types.RouteInfo{}

	outResp, err := m.AfterHandler(ctx, req, resp, info)
	require.NoError(t, err)
	assert.Equal(t, 204, outResp.Status)
	assert.Nil(t, outResp.Body)
}

func TestCORSMiddleware_MultipleOrigins_EchosMatchingOrigin(t *testing.T) {
	m := middleware.NewCORSMiddleware(
		[]string{"https://a.com", "https://b.com"},
		nil, nil,
	)
	ctx := context.Background()
	info := &types.RouteInfo{}

	// Request from b.com — should echo b.com, not a joined string
	req := &middleware.MiddlewareRequest{
		Method:  "GET",
		Headers: map[string]string{"Origin": "https://b.com"},
	}
	resp := &middleware.MiddlewareResponse{Status: 200}

	outResp, err := m.AfterHandler(ctx, req, resp, info)
	require.NoError(t, err)
	assert.Equal(t, "https://b.com", outResp.Headers["Access-Control-Allow-Origin"])
}

func TestCORSMiddleware_DisallowedOrigin_NoOriginHeader(t *testing.T) {
	m := middleware.NewCORSMiddleware([]string{"https://allowed.com"}, nil, nil)
	ctx := context.Background()
	req := &middleware.MiddlewareRequest{
		Method:  "GET",
		Headers: map[string]string{"Origin": "https://evil.com"},
	}
	resp := &middleware.MiddlewareResponse{Status: 200}
	info := &types.RouteInfo{}

	outResp, err := m.AfterHandler(ctx, req, resp, info)
	require.NoError(t, err)
	assert.Empty(t, outResp.Headers["Access-Control-Allow-Origin"])
	assert.Empty(t, outResp.Headers["Access-Control-Allow-Credentials"])
}

func TestCORSMiddleware_WildcardOrigin_NoCredentials(t *testing.T) {
	// Access-Control-Allow-Credentials must not be set when origin is "*"
	// — browsers block credentialed requests to wildcard origins.
	m := middleware.NewCORSMiddleware([]string{"*"}, nil, nil)
	ctx := context.Background()
	req := &middleware.MiddlewareRequest{Method: "GET", Headers: map[string]string{}}
	resp := &middleware.MiddlewareResponse{Status: 200}
	info := &types.RouteInfo{}

	outResp, err := m.AfterHandler(ctx, req, resp, info)
	require.NoError(t, err)
	assert.Equal(t, "*", outResp.Headers["Access-Control-Allow-Origin"])
	assert.Empty(t, outResp.Headers["Access-Control-Allow-Credentials"])
}

func TestCORSMiddleware_SpecificOrigin_SetsCredentials(t *testing.T) {
	m := middleware.NewCORSMiddleware([]string{"https://app.com"}, nil, nil)
	ctx := context.Background()
	req := &middleware.MiddlewareRequest{
		Method:  "GET",
		Headers: map[string]string{"Origin": "https://app.com"},
	}
	resp := &middleware.MiddlewareResponse{Status: 200}
	info := &types.RouteInfo{}

	outResp, err := m.AfterHandler(ctx, req, resp, info)
	require.NoError(t, err)
	assert.Equal(t, "https://app.com", outResp.Headers["Access-Control-Allow-Origin"])
	assert.Equal(t, "true", outResp.Headers["Access-Control-Allow-Credentials"])
}

func TestCORSMiddleware_BeforeRouting_PassesThrough(t *testing.T) {
	m := middleware.NewCORSMiddleware(nil, nil, nil)
	req := &middleware.MiddlewareRequest{Method: "GET", Path: "/"}

	result, err := m.BeforeRouting(req)
	require.NoError(t, err)
	assert.Equal(t, req, result)
}

func TestCORSMiddleware_EmptySlices_NoHeaders(t *testing.T) {
	m := middleware.NewCORSMiddleware(nil, nil, nil)
	ctx := context.Background()
	req := &middleware.MiddlewareRequest{Method: "GET"}
	resp := &middleware.MiddlewareResponse{Status: 200}
	info := &types.RouteInfo{}

	outResp, err := m.AfterHandler(ctx, req, resp, info)
	require.NoError(t, err)
	assert.Empty(t, outResp.Headers["Access-Control-Allow-Origin"])
	assert.Empty(t, outResp.Headers["Access-Control-Allow-Methods"])
}

package middleware

import (
	"context"
	"database/sql"
	"testing"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/middleware"
	"github.com/medubin/gonzo/api/src/types"
)

func TestAuthMiddleware_ExtractToken_MissingCookies(t *testing.T) {
	// Setup auth middleware with mock DB (won't be used in this test)
	db := &sql.DB{} // Mock DB
	authMiddleware := NewAuthMiddleware(db)

	// Test with nil cookies (simulating the old bug)
	req := &middleware.MiddlewareRequest{
		Method:  "GET",
		Path:    "/protected",
		Headers: make(map[string]string),
		Cookies: cookies.Cookies{}, // Zero-value struct with nil internals
	}

	// This should not panic and should return empty string
	token := authMiddleware.extractToken(req)

	if token != "" {
		t.Errorf("Expected empty token for missing cookies, got %s", token)
	}
}

func TestAuthMiddleware_ExtractToken_FromAuthHeader(t *testing.T) {
	db := &sql.DB{} // Mock DB
	authMiddleware := NewAuthMiddleware(db)

	req := &middleware.MiddlewareRequest{
		Method:  "GET",
		Path:    "/protected", 
		Headers: map[string]string{
			"Authorization": "Bearer test-token-123",
		},
		Cookies: cookies.Cookies{}, // Zero-value (no cookies)
	}

	token := authMiddleware.extractToken(req)

	if token != "test-token-123" {
		t.Errorf("Expected 'test-token-123', got %s", token)
	}
}

func TestAuthMiddleware_BeforeHandler_AuthEndpoints(t *testing.T) {
	db := &sql.DB{} // Mock DB
	authMiddleware := NewAuthMiddleware(db)

	req := &middleware.MiddlewareRequest{
		Method:  "POST",
		Path:    "/auth/signup",
		Headers: make(map[string]string),
		Cookies: cookies.Cookies{}, // No cookies
	}

	routeInfo := &types.RouteInfo{
		Method:   "POST",
		Path:     "/auth/signup",
		Endpoint: "Signup",
		Server:   "GonzoServer",
	}

	ctx := context.Background()

	// Should pass through without authentication for auth endpoints
	newCtx, newReq, err := authMiddleware.BeforeHandler(ctx, req, routeInfo)

	if err != nil {
		t.Errorf("Expected no error for auth endpoint, got %v", err)
	}
	if newCtx != ctx {
		t.Error("Expected context to be unchanged")
	}
	if newReq != req {
		t.Error("Expected request to be unchanged")
	}
}

func TestAuthMiddleware_BeforeHandler_NoToken(t *testing.T) {
	db := &sql.DB{} // Mock DB
	authMiddleware := NewAuthMiddleware(db)

	req := &middleware.MiddlewareRequest{
		Method:  "GET",
		Path:    "/protected",
		Headers: make(map[string]string), // No auth header
		Cookies: cookies.Cookies{},       // No cookies
	}

	routeInfo := &types.RouteInfo{
		Method:   "GET", 
		Path:     "/protected",
		Endpoint: "GetProtectedData",
		Server:   "GonzoServer",
	}

	ctx := context.Background()

	// Should fail with clear error message
	_, _, err := authMiddleware.BeforeHandler(ctx, req, routeInfo)

	if err == nil {
		t.Error("Expected authentication error, got nil")
	}

	expectedMsg := "authentication required: no valid session token or authorization header found"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestAuthMiddleware_IsAuthEndpoint(t *testing.T) {
	db := &sql.DB{} // Mock DB
	authMiddleware := NewAuthMiddleware(db)

	tests := []struct {
		endpoint string
		expected bool
	}{
		{"Signup", true},
		{"SignIn", true}, 
		{"SignOut", true},
		{"GetUser", false},
		{"UpdateUser", false},
		{"SomeRandomEndpoint", false},
	}

	for _, test := range tests {
		result := authMiddleware.isAuthEndpoint(test.endpoint)
		if result != test.expected {
			t.Errorf("isAuthEndpoint(%s) = %v, expected %v", test.endpoint, result, test.expected)
		}
	}
}

func TestAuthMiddleware_Robustness_NilPointers(t *testing.T) {
	db := &sql.DB{} // Mock DB  
	authMiddleware := NewAuthMiddleware(db)

	// Test edge cases to ensure no panics
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Auth middleware panicked: %v", r)
		}
	}()

	// Test with request that has nil headers
	req := &middleware.MiddlewareRequest{
		Method:  "GET",
		Path:    "/test",
		Headers: nil, // nil headers
		Cookies: cookies.Cookies{},
	}

	token := authMiddleware.extractToken(req)
	if token != "" {
		t.Errorf("Expected empty token for request with nil headers, got %s", token)
	}
}
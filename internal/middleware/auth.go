package middleware

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/medubin/gonzo/api/src/middleware"
	"github.com/medubin/gonzo/api/src/types"
	"github.com/medubin/gonzo/db/queries"
)

// AuthMiddleware provides authentication for protected routes
type AuthMiddleware struct {
	db      *sql.DB
	queries *queries.Queries
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{
		db:      db,
		queries: queries.New(db),
	}
}

// AuthContextKey is used to store auth info in context
type AuthContextKey struct{}

// AuthInfo contains authenticated user information  
type AuthInfo struct {
	UserID   int32
	Username string
	Email    string
	Role     string
	Token    string
}

// BeforeRouting - no authentication needed at routing level
func (m *AuthMiddleware) BeforeRouting(req *middleware.MiddlewareRequest) (*middleware.MiddlewareRequest, error) {
	return req, nil
}

// BeforeHandler authenticates requests for protected endpoints
func (m *AuthMiddleware) BeforeHandler(ctx context.Context, req *middleware.MiddlewareRequest, info *types.RouteInfo) (context.Context, *middleware.MiddlewareRequest, error) {
	// Skip auth for authentication endpoints
	if m.isAuthEndpoint(info.Endpoint) {
		return ctx, req, nil
	}

	// Extract session token from cookie or Authorization header
	// This gracefully handles missing cookies and other cookie errors
	token := m.extractToken(req)
	if token == "" {
		// Provide a clear, user-friendly error message
		return ctx, req, errors.New("authentication required: no valid session token or authorization header found")
	}

	// Validate session and get user info
	authInfo, err := m.validateSession(ctx, token)
	if err != nil {
		// Wrap the error with context but don't expose internal details
		return ctx, req, fmt.Errorf("authentication failed: %v", err)
	}

	// Add auth info to context
	ctx = context.WithValue(ctx, AuthContextKey{}, authInfo)

	return ctx, req, nil
}

// AfterHandler - no post-processing needed for auth
func (m *AuthMiddleware) AfterHandler(ctx context.Context, req *middleware.MiddlewareRequest, resp *middleware.MiddlewareResponse, info *types.RouteInfo) (*middleware.MiddlewareResponse, error) {
	return resp, nil
}

// OnError handles authentication errors
func (m *AuthMiddleware) OnError(ctx context.Context, req *middleware.MiddlewareRequest, err error, info *types.RouteInfo) (*middleware.MiddlewareResponse, error) {
	// Return 401 for auth errors
	if strings.Contains(err.Error(), "authentication") {
		return &middleware.MiddlewareResponse{
			Status: 401,
			Body: map[string]interface{}{
				"error": "Unauthorized",
				"code":  "AUTH_REQUIRED",
			},
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}, nil
	}
	return nil, err
}

// isAuthEndpoint checks if the endpoint requires authentication
func (m *AuthMiddleware) isAuthEndpoint(endpoint string) bool {
	authEndpoints := map[string]bool{
		"Signup":  true,
		"SignIn":  true,
		"SignOut": true,
	}
	return authEndpoints[endpoint]
}

// extractToken gets the session token from cookies or Authorization header
func (m *AuthMiddleware) extractToken(req *middleware.MiddlewareRequest) string {
	// Try cookie first (preferred) - gracefully handle any cookie errors
	if sessionCookie, err := req.Cookies.Get("session_token"); err == nil && sessionCookie != nil {
		return sessionCookie.Value
	}
	// Note: cookie errors (including missing cookies) are expected and handled gracefully

	// Try Authorization header as fallback
	if req.Headers != nil {
		if authHeader, exists := req.Headers["Authorization"]; exists {
			if strings.HasPrefix(authHeader, "Bearer ") {
				return strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
	}

	return ""
}

// validateSession validates the session token and returns user info
func (m *AuthMiddleware) validateSession(ctx context.Context, token string) (*AuthInfo, error) {
	// Get session from database
	session, err := m.queries.GetSessionByToken(ctx, token)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("invalid session token")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	// Check if session is expired (if ExpiresAt is set)
	// TODO: Implement session expiration check if needed

	// Get user information
	user, err := m.queries.GetUser(ctx, session.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("database error: %v", err)
	}

	return &AuthInfo{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Token:    token,
	}, nil
}

// GetAuthInfo retrieves authentication info from context
func GetAuthInfo(ctx context.Context) (*AuthInfo, bool) {
	auth, ok := ctx.Value(AuthContextKey{}).(*AuthInfo)
	return auth, ok
}

// RequireRole checks if the authenticated user has the required role
func RequireRole(ctx context.Context, requiredRole string) error {
	auth, ok := GetAuthInfo(ctx)
	if !ok {
		return errors.New("no authentication info found")
	}

	if auth.Role != requiredRole && auth.Role != "admin" {
		return fmt.Errorf("insufficient permissions: required %s, have %s", requiredRole, auth.Role)
	}

	return nil
}

// RequireOwnershipOrAdmin checks if user owns the resource or is admin
func RequireOwnershipOrAdmin(ctx context.Context, resourceUserID int32) error {
	auth, ok := GetAuthInfo(ctx)
	if !ok {
		return errors.New("no authentication info found")
	}

	if auth.Role == "admin" || auth.UserID == resourceUserID {
		return nil
	}

	return errors.New("access denied: you can only access your own resources")
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
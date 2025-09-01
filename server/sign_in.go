package server

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
	"github.com/medubin/gonzo/db/queries"
	"github.com/medubin/gonzo/internal/services/auth"
)

// POST /auth/signin
func (s *GonzoServerImpl) SignIn(ctx context.Context, body *SignInRequest, cookie cookies.Cookies, url url.URL[struct{}, struct{}]) (*SignInResponse, error) {
	// Validate request body
	if err := body.Validate(); err != nil {
		return nil, err
	}

	// Get user by email
	user, err := s.Queries.GetUserByEmail(ctx, *body.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gerrors.UnauthenticatedError("invalid email or password")
		}
		return nil, gerrors.InternalError("database error")
	}

	// Check password
	if !auth.CheckPasswordHash(*body.Password, user.Password) {
		return nil, gerrors.UnauthenticatedError("invalid email or password")
	}

	// Generate new session token
	token, err := auth.GenerateToken()
	if err != nil {
		return nil, err
	}

	// Calculate session expiration (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	// Create session in database
	session, err := s.Queries.CreateSession(ctx, queries.CreateSessionParams{
		UserID: user.ID,
		Token:  token,
	})
	if err != nil {
		return nil, err
	}

	// Set session cookie
	cookie.Set(&http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Expires:  time.Unix(expiresAt, 0),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	// Convert types for response
	userID := UserID(user.ID)
	userRole := UserRoleFromString(user.Role)
	createdAt := user.CreatedAt.Unix()

	return &SignInResponse{
		User: &User{
			ID:        &userID,
			Username:  &user.Username,
			Email:     &user.Email,
			Role:      &userRole,
			CreatedAt: &createdAt,
		},
		Session: &Session{
			UserID:    &userID,
			Token:     &session.Token,
			ExpiresAt: &expiresAt,
		},
	}, nil
}

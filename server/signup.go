package server

import (
	"context"
	"net/http"
	"time"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
	"github.com/medubin/gonzo/db/queries"
	"github.com/medubin/gonzo/internal/services/auth"
)

// POST /auth/signup
func (s *GonzoServerImpl) Signup(ctx context.Context, body *SignupRequest, cookie cookies.Cookies, url url.URL[struct{}, struct{}]) (*SignupResponse, error) {
	// Validate request body (validation is handled by generated validation)
	if err := body.Validate(); err != nil {
		return nil, err
	}

	// Hash password
	passwordHash, err := auth.HashPassword(*body.Password)
	if err != nil {
		return nil, gerrors.InternalError("failed to hash password")
	}

	// Create user in database
	newUser, err := s.Queries.CreateUser(ctx, queries.CreateUserParams{
		Username: *body.Username,
		Email:    *body.Email,
		Password: passwordHash,
	})
	if err != nil {
		return nil, gerrors.InternalError(err.Error())
	}

	// Generate session token
	token, err := auth.GenerateToken()
	if err != nil {
		return nil, err
	}

	// Calculate session expiration (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour).Unix()

	// Create session in database
	session, err := s.Queries.CreateSession(ctx, queries.CreateSessionParams{
		UserID: newUser.ID,
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
	userID := UserID(newUser.ID)
	userRole := UserRoleFromString(newUser.Role) // Convert role from database
	createdAt := newUser.CreatedAt.Unix()

	return &SignupResponse{
		User: &User{
			ID:        &userID,
			Username:  &newUser.Username,
			Email:     &newUser.Email,
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

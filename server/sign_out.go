package server

import (
	"context"
	"net/http"
	"time"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/url"
)

// POST /auth/signout
func (s *GonzoServerImpl) SignOut(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[struct{}, struct{}]) (*SignOutResponse, error) {
	// Simple implementation - just clear the cookie and return success
	// Always clear the session cookie
	cookie.Set(&http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Unix(0, 0), // Set to past time to delete
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	success := true
	return &SignOutResponse{
		Success: &success,
	}, nil
}

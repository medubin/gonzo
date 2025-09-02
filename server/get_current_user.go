package server

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
	"github.com/medubin/gonzo/internal/middleware"
)

// GET /auth/me
func (s *GonzoServerImpl) GetCurrentUser(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[struct{}, struct{}]) (*GetUserResponse, error) {
	// Get auth info from context (set by auth middleware)
	authInfo, ok := middleware.GetAuthInfo(ctx)
	if !ok {
		return nil, gerrors.UnauthenticatedError("Authentication required")
	}

	// Get user details from database
	user, err := s.Queries.GetUser(ctx, authInfo.UserID)
	if err != nil {
		return nil, gerrors.InternalError("Failed to get user information")
	}

	// Convert database user to API user type
	userID := UserID(user.ID)
	createdAt := user.CreatedAt.Unix()
	role := UserRole(user.Role)

	return &GetUserResponse{
		User: &User{
			ID:        &userID,
			Username:  &user.Username,
			Email:     &user.Email,
			Role:      &role,
			CreatedAt: &createdAt,
		},
	}, nil
}

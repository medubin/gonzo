package server

import (
	"context"
	"database/sql"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
	"github.com/medubin/gonzo/internal/middleware"
)

// GET /user/{UserID}
func (s *GonzoServerImpl) GetUser(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[struct{}, GetUserUrl]) (*GetUserResponse, error) {
	if url.PathParams.UserID == nil {
		return nil, gerrors.MissingArgumentError("missing user ID")
	}

	requestedUserID := int32(*url.PathParams.UserID)

	// Check permissions - user can only view their own profile unless they're admin
	if err := middleware.RequireOwnershipOrAdmin(ctx, requestedUserID); err != nil {
		return nil, err
	}

	// Get user from database
	user, err := s.Queries.GetUser(ctx, requestedUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gerrors.NotFoundError("user not found")
		}
		return nil, gerrors.InternalError("database error")
	}

	// Convert types for response
	userID := UserID(user.ID)
	userRole := UserRoleFromString(user.Role)
	createdAt := user.CreatedAt.Unix()

	return &GetUserResponse{
		User: &User{
			ID:        &userID,
			Username:  &user.Username,
			Email:     &user.Email,
			Role:      &userRole,
			CreatedAt: &createdAt,
		},
	}, nil
}

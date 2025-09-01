package server

import (
	"context"
	"database/sql"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
	"github.com/medubin/gonzo/db/queries"
	"github.com/medubin/gonzo/internal/middleware"
)

// PUT /user/{UserID}
func (s *GonzoServerImpl) UpdateUser(ctx context.Context, body *UpdateUserRequest, cookie cookies.Cookies, url url.URL[struct{}, UpdateUserUrl]) (*UpdateUserResponse, error) {
	if url.PathParams.UserID == nil {
		return nil, gerrors.MissingArgumentError("missing user ID")
	}

	if err := body.Validate(); err != nil {
		return nil, err
	}

	targetUserID := int32(*url.PathParams.UserID)

	// Check permissions - user can only update their own profile unless they're admin
	if err := middleware.RequireOwnershipOrAdmin(ctx, targetUserID); err != nil {
		return nil, err
	}

	// Prepare update parameters
	updateParams := queries.UpdateUserParams{
		ID: targetUserID,
	}

	// Only update fields that are provided
	if body.Username != nil {
		updateParams.Username = *body.Username
	}
	if body.Email != nil {
		updateParams.Email = *body.Email
	}

	// TODO: Handle Role updates (requires additional permission checks)
	// For now, we don't allow role updates through this endpoint

	// Update user in database
	updatedUser, err := s.Queries.UpdateUser(ctx, updateParams)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, gerrors.NotFoundError("user not found")
		}
		return nil, gerrors.InternalError("database error")
	}

	// Convert types for response
	userID := UserID(updatedUser.ID)
	userRole := UserRoleFromString(updatedUser.Role)
	createdAt := updatedUser.CreatedAt.Unix()

	return &UpdateUserResponse{
		User: &User{
			ID:        &userID,
			Username:  &updatedUser.Username,
			Email:     &updatedUser.Email,
			Role:      &userRole,
			CreatedAt: &createdAt,
		},
	}, nil
}

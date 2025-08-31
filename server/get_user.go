package server

import (
	"context"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// GET /user/{UserID}
func (s *GonzoServerImpl) GetUser(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[GetUserParams, GetUserUrl]) (*GetUserResponse, error) {
	if url.PathParams.UserID == nil {
		return nil, gerrors.MissingArgumentError("user id")
	}

	userID := int32(*url.PathParams.UserID)

	user, err := s.Queries.GetUser(ctx, int32(userID))
	if err != nil {
		return nil, err
	}

	return &GetUserResponse{
		User: &User{
			ID:    (*UserID)(&user.ID),
			Name:  &user.Username,
			Email: &user.Email,
		},
	}, nil
}

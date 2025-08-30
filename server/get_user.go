package server

import (
	"context"
	"strconv"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// GET /user/{UserID}
func (s *GonzoServerImpl) GetUser(ctx context.Context, body *any, cookie cookies.Cookies, url url.URL[GetUserUrl]) (*GetUserResponse, error) {
	if url.Params.UserID == nil {
		return nil, gerrors.MissingArgumentError("user id")
	}

	userID, err := strconv.Atoi(*url.Params.UserID)

	if err != nil {
		return nil, gerrors.InvalidArgumentError("user id")
	}

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

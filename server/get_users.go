package server

import (
	"context"
	"errors"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/url"
)

func (s *ServerImpl) GetUsers(ctx context.Context, body *GetUsersBody, cookie cookies.Cookies, url url.URL[GetUsersUrl]) (*GetUsersResponse, error) {
	return nil, errors.New("not implemented")
}

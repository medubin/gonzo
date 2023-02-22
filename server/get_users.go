package server

import (
	"context"
	"errors"

	"github.com/medubin/gonzo/api/utils/cookies"
	"github.com/medubin/gonzo/api/utils/url"
)

func (s *ServerImpl) GetUsers(ctx context.Context, body *GetUsersBody, cookie cookies.Cookies, url url.URL[GetUsersUrl]) (*GetUsersResponse, error) {
	return nil, errors.New("not implemented")
}

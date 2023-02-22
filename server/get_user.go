package server

import (
	"context"
	"errors"

	"github.com/medubin/gonzo/api/utils/cookies"
	"github.com/medubin/gonzo/api/utils/url"
)

func (s *ServerImpl) GetUser(ctx context.Context, body *interface{}, cookie cookies.Cookies, url url.URL[GetUserUrl]) (*GetUserResponse, error) {
	return nil, errors.New("not implemented")
}

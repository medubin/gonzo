package server

import (
	"context"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

func (s *ServerImpl) GetUser(ctx context.Context, body *interface{}, cookie cookies.Cookies, url url.URL[GetUserUrl]) (*GetUserResponse, error) {
	return nil, gerrors.UnimplementedError("GetUser")
}

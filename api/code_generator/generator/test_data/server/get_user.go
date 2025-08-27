package test

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// Endpoints can contain url parameters, which can be any primitive type or enum
// GET endpoints do not contain a body
// GET /users/{id}
func (s *UserServiceImpl) GetUser(ctx context.Context, body *interface{}, cookie cookies.Cookies, url url.URL[GetUserUrl]) (*DetailedUser, error) {
  return nil, gerrors.UnimplementedError("GetUser")
}
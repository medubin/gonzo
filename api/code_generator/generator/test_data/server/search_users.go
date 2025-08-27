package test

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// GET /users/search
func (s *UserServiceImpl) SearchUsers(ctx context.Context, body *interface{}, cookie cookies.Cookies, url url.URL[interface{}]) (*UserCollection, error) {
  return nil, gerrors.UnimplementedError("SearchUsers")
}
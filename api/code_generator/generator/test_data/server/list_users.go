package server

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// Endpoints can take a struct of parameters
// GET /users
func (s *UserServiceImpl) ListUsers(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[struct{}]) (*UserCollection, error) {
  return nil, gerrors.UnimplementedError("ListUsers")
}

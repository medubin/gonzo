package server

import (
	"context"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/url"
)

// Endpoints can take a struct of parameters
// GET /users
func (s *UserServiceImpl) ListUsers(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[UserListParams, struct{}]) (*UserCollection, error) {
  return nil, gerrors.UnimplementedError("ListUsers")
}

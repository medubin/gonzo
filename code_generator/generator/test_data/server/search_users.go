package server

import (
	"context"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/handle"
	"github.com/medubin/gonzo/runtime/url"
)

// GET /users/search
func (s *UserServiceImpl) SearchUsers(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[UserSearchParams, struct{}]) (*handle.Response[UserCollection], error) {
  return nil, gerrors.UnimplementedError("SearchUsers")
}

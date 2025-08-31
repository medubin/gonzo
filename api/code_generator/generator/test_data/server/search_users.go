package server

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// GET /users/search
func (s *UserServiceImpl) SearchUsers(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[UserSearchParams, struct{}]) (*UserCollection, error) {
  return nil, gerrors.UnimplementedError("SearchUsers")
}

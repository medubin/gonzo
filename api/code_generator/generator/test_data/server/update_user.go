package server

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// PUT /users/{id}
func (s *UserServiceImpl) UpdateUser(ctx context.Context, body *UpdateUserRequest, cookie cookies.Cookies, url url.URL[struct{}, UpdateUserUrl]) (*User, error) {
  return nil, gerrors.UnimplementedError("UpdateUser")
}

package test

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// DELETE /users/{id}
func (s *UserServiceImpl) DeleteUser(ctx context.Context, body *DeleteUserRequest, cookie cookies.Cookies, url url.URL[DeleteUserUrl]) (*User, error) {
  return nil, gerrors.UnimplementedError("DeleteUser")
}
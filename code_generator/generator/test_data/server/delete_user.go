package server

import (
	"context"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/handle"
	"github.com/medubin/gonzo/runtime/url"
)

// DELETE /users/{id}
func (s *UserServiceImpl) DeleteUser(ctx context.Context, body *DeleteUserRequest, cookie cookies.Cookies, url url.URL[struct{}, DeleteUserUrl]) (*handle.Response[User], error) {
  return nil, gerrors.UnimplementedError("DeleteUser")
}

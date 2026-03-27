package user_service

import (
	"context"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/handle"
	"github.com/medubin/gonzo/runtime/url"
	server "github.com/medubin/gonzo/code_generator/generator/test_data/server"
)

// PUT /users/{id}
func (s *UserServiceImpl) UpdateUser(ctx context.Context, body *server.UpdateUserRequest, cookie cookies.Cookies, url url.URL[struct{}, server.UpdateUserUrl]) (*handle.Response[server.User], error) {
  return nil, gerrors.UnimplementedError("UpdateUser")
}

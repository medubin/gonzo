package user_service

import (
	"context"
	server "github.com/medubin/gonzo/code_generator/generator/test_data/server"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/handle"
	"github.com/medubin/gonzo/runtime/url"
)

// DELETE /users/{id}
func (s *UserServiceImpl) DeleteUser(ctx context.Context, body *server.DeleteUserRequest, cookie cookies.Cookies, url url.URL[struct{}, server.DeleteUserUrl]) (*handle.Response[server.User], error) {
	return nil, gerrors.UnimplementedError("DeleteUser")
}

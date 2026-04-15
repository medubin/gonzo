package user_service

import (
	"context"
	server "github.com/medubin/gonzo/code_generator/generator/test_data/server"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/handle"
	"github.com/medubin/gonzo/runtime/url"
)

// Endpoints can take a struct of parameters
// GET /users
func (s *UserServiceImpl) ListUsers(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[server.UserListParams, struct{}]) (*handle.Response[server.UserCollection], error) {
	return nil, gerrors.UnimplementedError("ListUsers")
}

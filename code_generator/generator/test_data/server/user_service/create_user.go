package user_service

import (
	"context"
	server "github.com/medubin/gonzo/code_generator/generator/test_data/server"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/handle"
	"github.com/medubin/gonzo/runtime/url"
)

// All other urls can contain a body
// body and return always refer to a struct type
// POST /users
func (s *UserServiceImpl) CreateUser(ctx context.Context, body *server.CreateUserRequest, cookie cookies.Cookies, url url.URL[struct{}, struct{}]) (*handle.Response[server.User], error) {
	return nil, gerrors.UnimplementedError("CreateUser")
}

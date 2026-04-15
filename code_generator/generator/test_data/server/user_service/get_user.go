package user_service

import (
	"context"
	server "github.com/medubin/gonzo/code_generator/generator/test_data/server"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/handle"
	"github.com/medubin/gonzo/runtime/url"
)

// Endpoints can contain url parameters, which can be any primitive type or enum
// GET endpoints do not contain a body
// GET /users/{id}
func (s *UserServiceImpl) GetUser(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[struct{}, server.GetUserUrl]) (*handle.Response[server.DetailedUser], error) {
	return nil, gerrors.UnimplementedError("GetUser")
}

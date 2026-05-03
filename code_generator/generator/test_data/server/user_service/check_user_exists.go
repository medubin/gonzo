package user_service

import (
	"context"
	server "github.com/medubin/gonzo/code_generator/generator/test_data/server"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/handle"
	"github.com/medubin/gonzo/runtime/url"
)

// HEAD and OPTIONS support metadata-style operations like existence checks and CORS preflight
// HEAD /users/{id}
func (s *UserServiceImpl) CheckUserExists(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[struct{}, server.CheckUserExistsUrl]) (*handle.Response[struct{}], error) {
	return nil, gerrors.UnimplementedError("CheckUserExists")
}

package user_service

import (
	"context"
	server "github.com/medubin/gonzo/code_generator/generator/test_data/server"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/handle"
	"github.com/medubin/gonzo/runtime/url"
)

// PATCH /users/{id}/profile
func (s *UserServiceImpl) PatchUserProfile(ctx context.Context, body *server.UserProfileUpdate, cookie cookies.Cookies, url url.URL[struct{}, server.PatchUserProfileUrl]) (*handle.Response[server.UserProfile], error) {
	return nil, gerrors.UnimplementedError("PatchUserProfile")
}

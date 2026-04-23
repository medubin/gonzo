package media_service

import (
	"context"
	server "github.com/medubin/gonzo/code_generator/generator/test_data/server"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/handle"
	"github.com/medubin/gonzo/runtime/url"
)

// POST /users/{id}/avatar
func (s *MediaServiceImpl) UploadAvatar(ctx context.Context, body *server.UploadAvatarRequest, cookie cookies.Cookies, url url.URL[struct{}, server.UploadAvatarUrl]) (*handle.Response[server.UploadResult], error) {
	return nil, gerrors.UnimplementedError("UploadAvatar")
}

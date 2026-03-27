package notification_service

import (
	"context"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/handle"
	"github.com/medubin/gonzo/runtime/url"
	server "github.com/medubin/gonzo/code_generator/generator/test_data/server"
)

// GET /users/{userId}/notifications
func (s *NotificationServiceImpl) GetUserNotifications(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[server.UserListParams, server.GetUserNotificationsUrl]) (*handle.Response[[]server.Notification], error) {
  return nil, gerrors.UnimplementedError("GetUserNotifications")
}

package server

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// GET /users/{userId}/notifications
func (s *NotificationServiceImpl) GetUserNotifications(ctx context.Context, body *any, cookie cookies.Cookies, url url.URL[GetUserNotificationsUrl]) (*[]Notification, error) {
  return nil, gerrors.UnimplementedError("GetUserNotifications")
}

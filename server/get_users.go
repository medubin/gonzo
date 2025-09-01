package server

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// GET /users
func (s *GonzoServerImpl) GetUsers(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[GetUsersQuery, struct{}]) (*GetUsersResponse, error) {
  return nil, gerrors.UnimplementedError("GetUsers")
}

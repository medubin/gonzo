package test

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// GET /users/role/{role}
func (s *UserServiceImpl) GetUsersByRole(ctx context.Context, body *interface{}, cookie cookies.Cookies, url url.URL[GetUsersByRoleUrl]) (*UserCollection, error) {
  return nil, gerrors.UnimplementedError("GetUsersByRole")
}
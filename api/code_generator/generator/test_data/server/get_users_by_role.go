package server

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// GET /users/role/{role}
func (s *UserServiceImpl) GetUsersByRole(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[UserListParams, GetUsersByRoleUrl]) (*UserCollection, error) {
	println(*url.Params.Page)
	println(*url.PathParams.Role)
  return nil, gerrors.UnimplementedError("GetUsersByRole")
}

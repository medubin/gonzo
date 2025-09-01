package server

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// POST /auth/signout
func (s *GonzoServerImpl) SignOut(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[struct{}, struct{}]) (*SignOutResponse, error) {
  return nil, gerrors.UnimplementedError("SignOut")
}

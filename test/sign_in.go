package server

import (
	"context"

  "github.com/medubin/gonzo/api/src/cookies"
  "github.com/medubin/gonzo/api/src/gerrors"
  "github.com/medubin/gonzo/api/src/url"
)

// POST /session/new
func (s *UserServiceImpl) SignIn(ctx context.Context, body *SignInBody, cookie cookies.Cookies, url url.URL[SignInUrl]) (*SignInResponse, error) {
	return nil, gerrors.UnimplementedError("SignIn")
}

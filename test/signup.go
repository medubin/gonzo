package server

import (
	"context"

  "github.com/medubin/gonzo/api/src/cookies"
  "github.com/medubin/gonzo/api/src/gerrors"
  "github.com/medubin/gonzo/api/src/url"
)

// POST /user/new
func (s *ServerImpl) Signup(ctx context.Context, body *SignupBody, cookie cookies.Cookies, url url.URL[SignupUrl]) (*SignupResponse, error) {
	return nil, gerrors.UnimplementedError("Signup")
}

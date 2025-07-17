package server

import (
	"context"

  "github.com/medubin/gonzo/api/src/cookies"
  "github.com/medubin/gonzo/api/src/gerrors"
  "github.com/medubin/gonzo/api/src/url"
)

// DELETE /session
func (s *ServerImpl) SignOut(ctx context.Context, body *SignOutBody, cookie cookies.Cookies, url url.URL[SignOutUrl]) (*SignOutResponse, error) {
	return nil, gerrors.UnimplementedError("SignOut")
}

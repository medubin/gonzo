package server

import (
	"context"
	"errors"

	"github.com/medubin/gonzo/utils/cookies"
	"github.com/medubin/gonzo/utils/url"
)

func (s *ServerImpl) SignIn(ctx context.Context, body *SignInBody, cookie cookies.Cookies, url url.URL[SignInUrl]) (*SignInResponse, error) {
	return nil, errors.New("Not Implemented")
}

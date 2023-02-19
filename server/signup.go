package server

import (
	"context"
	"errors"

	"github.com/medubin/gonzo/utils/cookies"
	"github.com/medubin/gonzo/utils/url"
)

func (s *ServerImpl) Signup(ctx context.Context, body *SignupBody, cookie cookies.Cookies, url url.URL[SignupUrl]) (*SignupResponse, error) {
	println("hi")
	return nil, errors.New("Not Implemented")
}

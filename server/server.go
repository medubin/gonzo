package server

import (
	"context"
	"errors"

	"github.com/medubin/gonzo/router"
)

type S struct {}

func (s S) Signup(ctx context.Context, body SignupBody, cookie router.Cookies) (*SignupResponse, error) {
	return nil, errors.New("hi")
}
func (s S)	SignIn(ctx context.Context, body SignInBody, cookie router.Cookies) (*SignInResponse, error) {
	return nil, errors.New("hi")
}
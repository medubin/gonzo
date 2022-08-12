package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/medubin/gonzo/router"
)

type S struct{}

func (s S) Signup(ctx context.Context, body SignupBody, cookie router.Cookies) (*SignupResponse, error) {
	println(body.User.ID)
	println(body.Password)

	cookie.Set(&http.Cookie{
		Name:  "UserID",
		Value: string(body.User.ID),
	})

	return &SignupResponse{User: body.User}, nil
}
func (s S) SignIn(ctx context.Context, body SignInBody, cookie router.Cookies) (*SignInResponse, error) {
	return nil, errors.New("hi")
}

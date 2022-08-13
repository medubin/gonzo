package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/medubin/gonzo/utils/cookies"
)

type S struct{}

func (s S) Signup(ctx context.Context, body SignupBody, cookie cookies.Cookies) (*SignupResponse, error) {
	println(body.User.ID)
	println(body.Password)

	v, _ := json.Marshal(ctx.Value("params"))
	println(string(v))

	cookie.Set(&http.Cookie{
		Name:  "UserID",
		Value: string(body.User.ID),
	})

	return &SignupResponse{User: body.User}, nil
}
func (s S) SignIn(ctx context.Context, body SignInBody, cookie cookies.Cookies) (*SignInResponse, error) {
	return nil, errors.New("hi")
}

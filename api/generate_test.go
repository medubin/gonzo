package api_test

import (
	"testing"

	"github.com/medubin/gonzo/api"
	"github.com/stretchr/testify/assert"
)

const expectdOutput = `package server

import (
	"context"

	"github.com/medubin/gonzo/utils/cookies"
	"github.com/medubin/gonzo/utils/handle"
	"github.com/medubin/gonzo/utils/router"
	"github.com/medubin/gonzo/utils/url"
)

type UserID string

type User struct {
	ID    UserID
	Name  string
	Email string
}

type Session struct {
	UserID UserID
	Token  string
}

type SignupBody struct {
	User     User
	Password string
}

type SignupResponse struct {
	User User
}

type SignInBody struct {
	UserID   UserID
	Password string
}

type SignInResponse struct {
	Session Session
}

type GetUserResponse struct {
	User User
}

type SignupUrl struct {
}

type SignInUrl struct {
}

type GetUserUrl struct {
	UserID string
}

type Server interface {
	Signup(ctx context.Context, body SignupBody, cookie cookies.Cookies, url url.URL[SignupUrl]) (*SignupResponse, error)
	SignIn(ctx context.Context, body SignInBody, cookie cookies.Cookies, url url.URL[SignInUrl]) (*SignInResponse, error)
	GetUser(ctx context.Context, body interface{}, cookie cookies.Cookies, url url.URL[GetUserUrl]) (*GetUserResponse, error)
}

func StartServer(s Server, r *router.Router) {
	r.Route("POST", "/user/new", handle.Handle(s.Signup))
	r.Route("POST", "/session/new", handle.Handle(s.SignIn))
	r.Route("GET", "/user/{UserID}", handle.Handle(s.GetUser))
}
`

func TestMain(t *testing.T) {
	t.Run("Success", func(t *testing.T) {

		output, err := api.GenerateAPI("test")
		println(output)
		assert.NoError(t, err)
		assert.Equal(t, expectdOutput, output)
		_ = api.WriteToFile("test", output)

	})
	t.Run("Nonexistent file", func(t *testing.T) {
		output, err := api.GenerateAPI("bleh")
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}

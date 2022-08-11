package api_test

import (
	"testing"

	"github.com/medubin/gonzo/api"
	"github.com/stretchr/testify/assert"
)

const expectdOutput = `package api

import (
	"context"
	"net/http"
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

type Server interface {
	Signup(ctx context.Context, body SignupBody, cookie http.CookieJar) (SignupResponse, error)
	SignIn(ctx context.Context, body SignInBody, cookie http.CookieJar) (SignInResponse, error)
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

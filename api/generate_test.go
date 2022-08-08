package api_test

import (
	"testing"

	"github.com/medubin/gonzo/api"
	"github.com/stretchr/testify/assert"
)

const expectdOutput = `mux.HandleFunc("user/new", utils.Handle(server.Signup))
mux.HandleFunc("session/new", utils.Handle(server.SignIn))

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
	Signup(body SignupBody) (SignupResponse, Session, error)
	SignIn(body SignInBody) (SignInResponse, Session, error)
}`

func TestMain(t *testing.T) {
	t.Run("Success", func(t *testing.T) {

		output, err := api.GenerateAPI("test.api")
		println(output)
		assert.NoError(t, err)
		assert.Equal(t, expectdOutput, output)
	})
	t.Run("Nonexistent file", func(t *testing.T) {
		output, err := api.GenerateAPI("bleh.api")
		assert.Error(t, err)
		assert.Empty(t, output)
	})
}

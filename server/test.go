package server

import (
	"context"

	"github.com/medubin/gonzo/router"
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
	Signup(ctx context.Context, body SignupBody, cookie router.Cookies) (*SignupResponse, error)
	SignIn(ctx context.Context, body SignInBody, cookie router.Cookies) (*SignInResponse, error)
}

func StartServer(s Server, r *router.Router) {
	r.Route("POST", "/user/new", router.Handle(s.Signup))
	r.Route("POST", "/session/new", router.Handle(s.SignIn))
}

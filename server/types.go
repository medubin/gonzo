package server

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

func (s *User) GetID() UserID {
	return s.ID
}

func (s *User) GetName() string {
	return s.Name
}

func (s *User) GetEmail() string {
	return s.Email
}

type Session struct {
	UserID UserID
	Token  string
}

func (s *Session) GetUserID() UserID {
	return s.UserID
}

func (s *Session) GetToken() string {
	return s.Token
}

type SignupBody struct {
	User     User
	Password string
}

func (s *SignupBody) GetUser() User {
	return s.User
}

func (s *SignupBody) GetPassword() string {
	return s.Password
}

type SignupResponse struct {
	User User
}

func (s *SignupResponse) GetUser() User {
	return s.User
}

type SignInBody struct {
	UserID   UserID
	Password string
}

func (s *SignInBody) GetUserID() UserID {
	return s.UserID
}

func (s *SignInBody) GetPassword() string {
	return s.Password
}

type SignInResponse struct {
	Session Session
}

func (s *SignInResponse) GetSession() Session {
	return s.Session
}

type GetUserResponse struct {
	User User
}

func (s *GetUserResponse) GetUser() User {
	return s.User
}

type GetUsersBody struct {
	UserIDs []UserID
}

func (s *GetUsersBody) GetUserIDs() []UserID {
	return s.UserIDs
}

type GetUsersResponse struct {
	Users map[UserID]User
}

func (s *GetUsersResponse) GetUsers() map[UserID]User {
	return s.Users
}

type SignupUrl struct {
}

type SignInUrl struct {
}

type GetUserUrl struct {
	UserID string
}

func (s *GetUserUrl) GetUserID() string {
	return s.UserID
}

type GetUsersUrl struct {
}

type Server interface {
	Signup(ctx context.Context, body *SignupBody, cookie cookies.Cookies, url url.URL[SignupUrl]) (*SignupResponse, error)
	SignIn(ctx context.Context, body *SignInBody, cookie cookies.Cookies, url url.URL[SignInUrl]) (*SignInResponse, error)
	GetUser(ctx context.Context, body *interface{}, cookie cookies.Cookies, url url.URL[GetUserUrl]) (*GetUserResponse, error)
	GetUsers(ctx context.Context, body *GetUsersBody, cookie cookies.Cookies, url url.URL[GetUsersUrl]) (*GetUsersResponse, error)
}

func StartServer(s Server, r *router.Router) {
	r.Route("POST", "/user/new", handle.Handle(s.Signup))
	r.Route("POST", "/session/new", handle.Handle(s.SignIn))
	r.Route("GET", "/user/{UserID}", handle.Handle(s.GetUser))
	r.Route("GET", "/users/", handle.Handle(s.GetUsers))
}

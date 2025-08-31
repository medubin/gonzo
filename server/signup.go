package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/url"
	"github.com/medubin/gonzo/db/queries"
	"github.com/medubin/gonzo/internal/services/auth"
)

// POST /user/new
func (s *GonzoServerImpl) Signup(ctx context.Context, body *SignupBody, cookie cookies.Cookies, url url.URL[struct{}, struct{}]) (*SignupResponse, error) {
	if body == nil {
		return nil, errors.New("missing body")
	}
	user := body.User
	password := body.Password
	if user == nil {
		return nil, errors.New("missing user")
	}

	if user.Email == nil {
		return nil, errors.New("missing email")
	}

	if user.Name == nil {
		return nil, errors.New("missing name")
	}

	if password == nil {
		return nil, errors.New("missing password")
	}

	password_hash, err := auth.HashPassword(*password)
	if err != nil {
		return nil, err
	}

	newUser, err := s.Queries.CreateUser(ctx, queries.CreateUserParams{
		Username: *user.Name,
		Email:    *user.Email,
		Password: password_hash,
	})

	if err != nil {
		return nil, err
	}

	token, err := auth.GenerateToken()
	if err != nil {
		return nil, err
	}

	session, err := s.Queries.CreateSession(ctx, queries.CreateSessionParams{
		UserID: newUser.ID,
		Token:  token,
	})
	if err != nil {
		return nil, err
	}

	id := UserID(newUser.ID)

	cookie.Set(&http.Cookie{
		Name:  "session_token",
		Value: session.Token,
	})

	return &SignupResponse{
		User: &User{
			ID:    &id,
			Name:  &newUser.Username,
			Email: &newUser.Email,
		},
	}, nil
}

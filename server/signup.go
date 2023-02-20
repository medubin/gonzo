package server

import (
	"context"
	"errors"

	"github.com/medubin/gonzo/api/utils/cookies"
	"github.com/medubin/gonzo/api/utils/url"
	"github.com/medubin/gonzo/db/queries"
)

func (s *ServerImpl) Signup(ctx context.Context, body *SignupBody, cookie cookies.Cookies, url url.URL[SignupUrl]) (*SignupResponse, error) {
	if body == nil {
		return nil, errors.New("must add body")
	}
	user := body.GetUser()
	password := body.GetPassword()
	if user.GetEmail() == nil {
		return nil, errors.New("must add email")
	}
	if user.GetName() == nil {
		return nil, errors.New("must add name")
	}
	if password == nil {
		return nil, errors.New("must add password")
	}

	res, err := s.Queries.CreateUser(ctx, queries.CreateUserParams{
		Username: *user.GetName(),
		Email:    *user.GetEmail(),
		Password: *password,
	})

	if err != nil {
		return nil, err
	}

	id := UserID(&res.ID)

	return &SignupResponse{
		User: &User{
			ID:    &id,
			Name:  &res.Username,
			Email: &res.Email,
		},
	}, nil
}

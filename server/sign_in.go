package server

import (
	"context"
	"errors"
	"net/http"

	"github.com/medubin/gonzo/api/utils/cookies"
	"github.com/medubin/gonzo/api/utils/url"
	"github.com/medubin/gonzo/db/queries"
	"github.com/medubin/gonzo/internal/services/auth"
)

func (s *ServerImpl) SignIn(ctx context.Context, body *SignInBody, cookie cookies.Cookies, url url.URL[SignInUrl]) (*SignInResponse, error) {
	if body == nil {
		return nil, errors.New("missing body")
	}
	if body.GetPassword() == nil {
		return nil, errors.New("missing password")
	}

	userID := body.GetUserID()
	if userID == nil {
		return nil, errors.New("missing user id")
	}

	user, err := s.Queries.GetUser(ctx, **userID)
	if err != nil {
		return nil, err
	}

	isSamePass := auth.CheckPasswordHash(*body.GetPassword(), user.Password)

	if !isSamePass {
		return nil, errors.New("password incorrect")
	}

	token, err := auth.GenerateToken()
	if err != nil {
		return nil, err
	}

	session, err := s.Queries.CreateSession(ctx, queries.CreateSessionParams{
		UserID: user.ID,
		Token:  token,
	})
	if err != nil {
		return nil, err
	}

	cookie.Set(&http.Cookie{
		Name:  "session_token",
		Value: session.Token,
	})

	return &SignInResponse{
		Session: &Session{
			UserID: userID,
			Token:  &token,
		},
	}, nil

}

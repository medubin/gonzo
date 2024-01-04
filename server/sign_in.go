package server

import (
	"context"
	"net/http"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
	"github.com/medubin/gonzo/db/queries"
	"github.com/medubin/gonzo/internal/services/auth"
)

// POST /session/new
func (s *ServerImpl) SignIn(ctx context.Context, body *SignInBody, cookie cookies.Cookies, url url.URL[SignInUrl]) (*SignInResponse, error) {
	if body == nil {
		return nil, gerrors.MissingArgumentError("body")
	}
	if body.Password == nil {
		return nil, gerrors.MissingArgumentError("password")
	}

	userID := body.UserID
	if userID == nil {
		return nil, gerrors.MissingArgumentError("user id")
	}

	user, err := s.Queries.GetUser(ctx, int32(*userID))
	if err != nil {
		return nil, err
	}

	isSamePass := auth.CheckPasswordHash(*body.Password, user.Password)

	if !isSamePass {
		return nil, gerrors.InvalidArgumentError("password incorrect")
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

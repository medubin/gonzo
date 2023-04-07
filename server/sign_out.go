package server

import (
	"context"
	"net/http"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
	"github.com/medubin/gonzo/db/queries"
)

// DELETE /session
func (s *ServerImpl) SignOut(ctx context.Context, body *SignOutBody, cookie cookies.Cookies, url url.URL[SignOutUrl]) (*SignOutResponse, error) {
	if body == nil {
		return nil, gerrors.MissingArgumentError("body")
	}

	session := body.GetSession()
	if session == nil {
		return nil, gerrors.MissingArgumentError("session")
	}

	if session.GetToken() == nil {
		return nil, gerrors.MissingArgumentError("token")
	}

	if session.GetUserID() == nil {
		return nil, gerrors.MissingArgumentError("userID")
	}

	err := s.Queries.DeleteSession(ctx, queries.DeleteSessionParams{
		Token:  *session.GetToken(),
		UserID: **session.GetUserID(),
	})

	if err != nil {
		return nil, err
	}

	cookie.Set(&http.Cookie{
		Name:  "session_token",
		Value: "",
	})

	return &SignOutResponse{}, nil
}

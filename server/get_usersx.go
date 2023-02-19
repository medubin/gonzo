package server

import (
	"context"
	"errors"

	"github.com/medubin/gonzo/utils/cookies"
	"github.com/medubin/gonzo/utils/url"
)

func (s *ServerImpl) GetUsersx(ctx context.Context, body *GetUsersBodyx, cookie cookies.Cookies, url url.URL[GetUsersxUrl]) (*GetUsersResponsex, error) {
	return nil, errors.New("Not Implemented")
}

package test

import (
	"context"
	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

// PATCH /users/{id}/profile
func (s *UserServiceImpl) PatchUserProfile(ctx context.Context, body *UserProfileUpdate, cookie cookies.Cookies, url url.URL[PatchUserProfileUrl]) (*UserProfile, error) {
  return nil, gerrors.UnimplementedError("PatchUserProfile")
}
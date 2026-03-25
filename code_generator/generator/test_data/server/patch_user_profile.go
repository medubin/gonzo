package server

import (
	"context"
	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/url"
)

// PATCH /users/{id}/profile
func (s *UserServiceImpl) PatchUserProfile(ctx context.Context, body *UserProfileUpdate, cookie cookies.Cookies, url url.URL[struct{}, PatchUserProfileUrl]) (*UserProfile, error) {
  return nil, gerrors.UnimplementedError("PatchUserProfile")
}

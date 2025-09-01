package middleware

import (
	"context"
	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/types"
)

// RequireBodyMiddleware ensures the request has a body
type RequireBodyMiddleware struct {
	BaseMiddleware
}

// NewRequireBody creates a new RequireBodyMiddleware
func NewRequireBody() *RequireBodyMiddleware {
	return &RequireBodyMiddleware{}
}

// BeforeHandler checks if body is required and present
func (m *RequireBodyMiddleware) BeforeHandler(ctx context.Context, req *MiddlewareRequest, info *types.RouteInfo) (context.Context, *MiddlewareRequest, error) {
	if req.Body == nil {
		return ctx, req, gerrors.MissingArgumentError("request body required")
	}
	return ctx, req, nil
}
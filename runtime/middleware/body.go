package middleware

import (
	"context"
	"strings"

	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/types"
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
	if info != nil && info.IsMultipart {
		if !strings.HasPrefix(req.Headers["Content-Type"], "multipart/form-data") {
			return ctx, req, gerrors.MissingArgumentError("multipart/form-data body required")
		}
		return ctx, req, nil
	}
	if req.Body == nil {
		return ctx, req, gerrors.MissingArgumentError("request body required")
	}
	return ctx, req, nil
}

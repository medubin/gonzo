package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/middleware"
	"github.com/medubin/gonzo/runtime/types"
	"github.com/medubin/gonzo/runtime/url"
)

type Router struct {
	routes     []RouteEntry
	middleware []middleware.Middleware
}

func (rtr *Router) Route(handlerFunc http.HandlerFunc, info *types.RouteInfo) {
	if info == nil {
		panic("RouteInfo is required")
	}

	exactPath := url.ConvertPathToRegex(info.Path)

	// Create route-specific middleware based on route info
	var routeMiddleware []middleware.Middleware
	
	// Auto-add RequireBody middleware if route requires a body
	if info.RequiresBody {
		routeMiddleware = append(routeMiddleware, middleware.NewRequireBody())
	}

	e := RouteEntry{
		Method:           info.Method,
		Path:             exactPath,
		HandlerFunc:      handlerFunc,
		Info:             info,
		RouteMiddleware:  routeMiddleware,
	}
	rtr.routes = append(rtr.routes, e)
}

func (rtr *Router) Use(m middleware.Middleware) {
	rtr.middleware = append(rtr.middleware, m)
}

// responseWriter captures response data for middleware processing
type responseWriter struct {
	statusCode int
	headers    http.Header
	body       []byte
	written    bool
}

func (rw *responseWriter) Header() http.Header {
	if rw.headers == nil {
		rw.headers = make(http.Header)
	}
	return rw.headers
}

func (rw *responseWriter) WriteHeader(code int) {
	if rw.written {
		return
	}
	rw.statusCode = code
	rw.written = true
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(200)
	}
	rw.body = append(rw.body, b...)
	return len(b), nil
}

// writeMiddlewareResponse writes a middleware response to the HTTP response
func (rtr *Router) writeMiddlewareResponse(w http.ResponseWriter, resp *middleware.MiddlewareResponse) {
	// Set headers
	for key, value := range resp.Headers {
		w.Header().Set(key, value)
	}

	// Set status code
	w.WriteHeader(resp.Status)

	// Write body if present
	if resp.Body != nil {
		if err := json.NewEncoder(w).Encode(resp.Body); err != nil {
			log.Printf("Error encoding middleware response: %v", err)
		}
	}
}

func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic: ", r) // Log the error
			gerrors.JSONError(w, fmt.Errorf("panic: %v", r))
		}
	}()

	// Parse request body for middleware if it exists
	var body any
	if r.ContentLength > 0 && r.Header.Get("Content-Type") == "application/json" {
		// Read the body into a buffer so we can use it for both middleware and handler
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			gerrors.JSONError(w, gerrors.MalformedError("failed to read request body"))
			return
		}
		
		// Replace the body with a new reader for the handler to use
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		
		// Parse JSON for middleware
		if len(bodyBytes) > 0 {
			var bodyData any
			if err := json.Unmarshal(bodyBytes, &bodyData); err != nil {
				gerrors.JSONError(w, gerrors.MalformedError("invalid JSON: "+err.Error()))
				return
			}
			body = bodyData
		}
	}

	// Always create middleware request (lightweight operation)
	middlewareReq := &middleware.MiddlewareRequest{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: middleware.ConvertHeadersFromHTTP(r.Header),
		Cookies: cookies.New(r, w),
		Params:  r.URL.Query(),
		Body:    body,
	}

	// Execute BeforeRouting middleware
	for _, m := range rtr.middleware {
		var err error
		middlewareReq, err = m.BeforeRouting(middlewareReq)
		if err != nil {
			gerrors.JSONError(w, err)
			return
		}
	}

	// Use middleware-modified method/path for routing
	method := middlewareReq.Method
	path := middlewareReq.Path

	for _, e := range rtr.routes {
		if e.Method != method {
			continue
		}
		params := e.Match(r)
		if params == nil {
			continue // No match found
		}

		ctx := context.WithValue(r.Context(), url.ParamKey{}, params)
		middlewareReq.PathParams = params

		// Get route info for middleware
		var routeInfo *types.RouteInfo
		if e.Info != nil {
			routeInfo = e.Info
		} else {
			// Fallback for manually registered routes
			routeInfo = &types.RouteInfo{
				Method:   method,
				Path:     path,
				Endpoint: "Unknown",
				Server:   "Unknown",
			}
		}

		// Execute BeforeHandler middleware: global first, then route-specific
		allMiddleware := append(rtr.middleware, e.RouteMiddleware...)
		for _, m := range allMiddleware {
			var err error
			ctx, middlewareReq, err = m.BeforeHandler(ctx, middlewareReq, routeInfo)
			if err != nil {
				rtr.handleMiddlewareError(w, ctx, middlewareReq, err, routeInfo)
				return
			}
		}

		// Execute handler with response capture
		responseCapture := &responseWriter{statusCode: 200}
		e.HandlerFunc.ServeHTTP(responseCapture, r.WithContext(ctx))
		
		// Handle response through middleware
		rtr.handleMiddlewareResponse(w, ctx, middlewareReq, responseCapture, routeInfo, allMiddleware)
		return
	}

	// Handle unmatched routes, but still allow CORS middleware to process
	// Create a fake route info for middleware
	routeInfo := &types.RouteInfo{
		Method:   method,
		Path:     path,
		Endpoint: "NotFound",
		Server:   "Router",
	}

	// Create a 404 response but let middleware process it (especially CORS)
	responseCapture := &responseWriter{statusCode: 404}
	errorMsg := map[string]string{"error": fmt.Sprintf("bad_route: %s: %s", method, path)}
	if bodyBytes, err := json.Marshal(errorMsg); err == nil {
		responseCapture.body = bodyBytes
		responseCapture.Header().Set("Content-Type", "application/json")
	}

	// Handle response through middleware (this will add CORS headers)
	rtr.handleMiddlewareResponse(w, r.Context(), middlewareReq, responseCapture, routeInfo, rtr.middleware)
}

func (rtr *Router) handleMiddlewareError(w http.ResponseWriter, ctx context.Context, req *middleware.MiddlewareRequest, err error, info *types.RouteInfo) {
	for _, errorM := range rtr.middleware {
		if errorResp, errorErr := errorM.OnError(ctx, req, err, info); errorErr == nil && errorResp != nil {
			// Run AfterHandler middleware on error responses too (for CORS, etc.)
			var middlewareErr error
			for i := len(rtr.middleware) - 1; i >= 0; i-- {
				errorResp, middlewareErr = rtr.middleware[i].AfterHandler(ctx, req, errorResp, info)
				if middlewareErr != nil {
					gerrors.JSONError(w, middlewareErr)
					return
				}
			}
			rtr.writeMiddlewareResponse(w, errorResp)
			return
		}
	}
	gerrors.JSONError(w, err)
}

func (rtr *Router) handleMiddlewareResponse(w http.ResponseWriter, ctx context.Context, req *middleware.MiddlewareRequest, responseCapture *responseWriter, info *types.RouteInfo, allMiddleware []middleware.Middleware) {
	// Parse captured body for middleware access
	var bodyForMiddleware any
	if len(responseCapture.body) > 0 {
		var jsonBody any
		if err := json.Unmarshal(responseCapture.body, &jsonBody); err == nil {
			bodyForMiddleware = jsonBody
		} else {
			bodyForMiddleware = string(responseCapture.body)
		}
	}

	middlewareResp := &middleware.MiddlewareResponse{
		Status:  responseCapture.statusCode,
		Body:    bodyForMiddleware,
		Headers: middleware.ConvertHeadersFromHTTP(responseCapture.Header()),
	}

	var err error
	
	// Execute all middleware AfterHandler in reverse order (LIFO)
	for i := len(allMiddleware) - 1; i >= 0; i-- {
		middlewareResp, err = allMiddleware[i].AfterHandler(ctx, req, middlewareResp, info)
		if err != nil {
			rtr.handleMiddlewareError(w, ctx, req, err, info)
			return
		}
	}

	rtr.writeMiddlewareResponse(w, middlewareResp)
}

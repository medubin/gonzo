package router

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/middleware"
	"github.com/medubin/gonzo/api/src/url"
)

type Router struct {
	routes     []RouteEntry
	middleware []middleware.Middleware
}

func (rtr *Router) Route(method, path string, handlerFunc http.HandlerFunc) {
	rtr.RouteWithInfo(method, path, handlerFunc, nil)
}

func (rtr *Router) RouteWithInfo(method, path string, handlerFunc http.HandlerFunc, info *middleware.RouteInfo) {
	exactPath := url.ConvertPathToRegex(path)

	// Create route-specific middleware based on route info
	var routeMiddleware []middleware.Middleware
	
	// Auto-add RequireBody middleware if route requires a body
	if info != nil && info.RequiresBody {
		routeMiddleware = append(routeMiddleware, middleware.NewRequireBody())
	}

	e := RouteEntry{
		Method:           method,
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

	hasMiddleware := len(rtr.middleware) > 0
	var middlewareReq *middleware.MiddlewareRequest
	var err error

	// Create middleware request only if needed
	if hasMiddleware {
		middlewareReq = &middleware.MiddlewareRequest{
			Method:  r.Method,
			Path:    r.URL.Path,
			Headers: middleware.ConvertHeadersFromHTTP(r.Header),
			Params:  r.URL.Query(),
		}

		// Call BeforeRouting middleware
		for _, m := range rtr.middleware {
			middlewareReq, err = m.BeforeRouting(middlewareReq)
			if err != nil {
				gerrors.JSONError(w, err)
				return
			}
		}
	}

	// Use original request method/path or modified middleware version
	method := r.Method
	path := r.URL.Path
	if hasMiddleware {
		method = middlewareReq.Method
		path = middlewareReq.Path
	}

	for _, e := range rtr.routes {
		if e.Method != method {
			continue
		}
		params := e.Match(r)
		if params == nil {
			continue // No match found
		}

		ctx := context.WithValue(r.Context(), url.ParamKey{}, params)

		// Get route info for middleware
		routeInfo := e.Info
		if routeInfo == nil {
			// Fallback for manually registered routes
			routeInfo = &middleware.RouteInfo{
				Method:   method,
				Path:     path,
				Endpoint: "Unknown",
				Server:   "Unknown",
			}
		}

		// Middleware BeforeHandler hook
		if hasMiddleware {
			middlewareReq.PathParams = params

			// Execute global middleware first
			for _, m := range rtr.middleware {
				ctx, middlewareReq, err = m.BeforeHandler(ctx, middlewareReq, routeInfo)
				if err != nil {
					rtr.handleMiddlewareError(w, ctx, middlewareReq, err, routeInfo)
					return
				}
			}
			
			// Then execute route-specific middleware
			for _, m := range e.RouteMiddleware {
				ctx, middlewareReq, err = m.BeforeHandler(ctx, middlewareReq, routeInfo)
				if err != nil {
					rtr.handleMiddlewareError(w, ctx, middlewareReq, err, routeInfo)
					return
				}
			}
		} else if len(e.RouteMiddleware) > 0 {
			// No global middleware but we have route middleware
			middlewareReq = &middleware.MiddlewareRequest{
				Method:     r.Method,
				Path:       r.URL.Path,
				Headers:    middleware.ConvertHeadersFromHTTP(r.Header),
				Params:     r.URL.Query(),
				PathParams: params,
			}
			
			for _, m := range e.RouteMiddleware {
				ctx, middlewareReq, err = m.BeforeHandler(ctx, middlewareReq, routeInfo)
				if err != nil {
					rtr.handleMiddlewareError(w, ctx, middlewareReq, err, routeInfo)
					return
				}
			}
			hasMiddleware = true // Now we have middleware to handle
		}

		// Execute handler
		if hasMiddleware {
			// With middleware: capture response
			responseCapture := &responseWriter{statusCode: 200}
			e.HandlerFunc.ServeHTTP(responseCapture, r.WithContext(ctx))
			rtr.handleMiddlewareResponse(w, ctx, middlewareReq, responseCapture, routeInfo, e.RouteMiddleware)
		} else {
			// Without middleware: direct execution
			e.HandlerFunc.ServeHTTP(w, r.WithContext(ctx))
		}
		return
	}

	gerrors.JSONError(w, gerrors.BadRouteError(fmt.Sprintf("%s: %s", method, path)))
}

func (rtr *Router) handleMiddlewareError(w http.ResponseWriter, ctx context.Context, req *middleware.MiddlewareRequest, err error, info *middleware.RouteInfo) {
	for _, errorM := range rtr.middleware {
		if errorResp, errorErr := errorM.OnError(ctx, req, err, info); errorErr == nil && errorResp != nil {
			rtr.writeMiddlewareResponse(w, errorResp)
			return
		}
	}
	gerrors.JSONError(w, err)
}

func (rtr *Router) handleMiddlewareResponse(w http.ResponseWriter, ctx context.Context, req *middleware.MiddlewareRequest, responseCapture *responseWriter, info *middleware.RouteInfo, routeMiddleware []middleware.Middleware) {
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
	
	// Execute route-specific middleware AfterHandler first (reverse order)
	for i := len(routeMiddleware) - 1; i >= 0; i-- {
		middlewareResp, err = routeMiddleware[i].AfterHandler(ctx, req, middlewareResp, info)
		if err != nil {
			rtr.handleMiddlewareError(w, ctx, req, err, info)
			return
		}
	}
	
	// Then execute global middleware AfterHandler (reverse order)
	for i := len(rtr.middleware) - 1; i >= 0; i-- {
		middlewareResp, err = rtr.middleware[i].AfterHandler(ctx, req, middlewareResp, info)
		if err != nil {
			rtr.handleMiddlewareError(w, ctx, req, err, info)
			return
		}
	}

	rtr.writeMiddlewareResponse(w, middlewareResp)
}

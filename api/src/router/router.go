package router

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/medubin/gonzo/api/src/gerrors"
	"github.com/medubin/gonzo/api/src/url"
)

type Router struct {
	routes []RouteEntry
}

func (rtr *Router) Route(method, path string, handlerFunc http.HandlerFunc) {
	exactPath := url.ConvertPathToRegex(path)

	e := RouteEntry{
		Method:      method,
		Path:        exactPath,
		HandlerFunc: handlerFunc,
	}
	rtr.routes = append(rtr.routes, e)
}

func (rtr *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("panic: ", r) // Log the error
			gerrors.JSONError(w, fmt.Errorf("panic: %v", r))
		}
	}()

	for _, e := range rtr.routes {
		if e.Method != r.Method {
			continue
		}
		params := e.Match(r)
		if params == nil {
			continue // No match found
		}

		// Create new request with params stored in context
		ctx := context.WithValue(r.Context(), url.ParamKey{}, params)
		e.HandlerFunc.ServeHTTP(w, r.WithContext(ctx))
		return
	}

	gerrors.JSONError(w, gerrors.BadRouteError(fmt.Sprintf("%s: %s", r.Method, r.URL.Path)))
}

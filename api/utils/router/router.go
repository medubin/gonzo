package router

import (
	"context"
	"log"
	"net/http"

	"github.com/medubin/gonzo/api/utils/url"
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
			log.Println("ERROR:", r) // Log the error
			http.Error(w, "Uh oh!", http.StatusInternalServerError)
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

	http.NotFound(w, r)
}

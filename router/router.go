package router

import (
	"context"
	"log"
	"net/http"
	"regexp"
)

type Router struct {
	routes []RouteEntry
}

func (rtr *Router) Route(method, path string, handlerFunc http.HandlerFunc) {
	// NOTE: ^ means start of string and $ means end. Without these,
	//   we'll still match if the path has content before or after
	//   the expression (/foo/bar/baz would match the "/bar" route).
	exactPath := regexp.MustCompile("^" + path + "$")
	
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
	println(len(rtr.routes))
	
	for _, e := range rtr.routes {
		params := e.Match(r)
		if params == nil {
			continue // No match found
		}

		// Create new request with params stored in context
		ctx := context.WithValue(r.Context(), "params", params)
		e.HandlerFunc.ServeHTTP(w, r.WithContext(ctx))
		return
	}

	http.NotFound(w, r)
}

package router

import (
	"net/http"
	"regexp"
	
	"github.com/medubin/gonzo/api/src/middleware"
	"github.com/medubin/gonzo/api/src/types"
)

type RouteEntry struct {
	Path            *regexp.Regexp
	Method          string
	HandlerFunc     http.HandlerFunc
	Info            *types.RouteInfo
	RouteMiddleware []middleware.Middleware
}

func (ent *RouteEntry) Match(r *http.Request) map[string]string {
	match := ent.Path.FindStringSubmatch(r.URL.Path)
	if match == nil {
		return nil // No match found
	}

	// Create a map to store URL parameters in
	params := make(map[string]string)
	groupNames := ent.Path.SubexpNames()
	for i, group := range match {
		if groupNames[i] == "" {
			continue
		}
		params[groupNames[i]] = group
	}

	return params
}

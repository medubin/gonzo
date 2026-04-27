package router

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/medubin/gonzo/runtime/middleware"
	"github.com/medubin/gonzo/runtime/types"
)

type RouteEntry struct {
	Path            *regexp.Regexp
	Method          string
	HandlerFunc     http.HandlerFunc
	Info            *types.RouteInfo
	RouteMiddleware []middleware.Middleware
	Segments        []Segment
}

type Segment struct {
	IsParam bool
	Literal string
}

func ParseSegments(path string) []Segment {
	path = strings.Trim(path, "/")
	if path == "" {
		return nil
	}
	parts := strings.Split(path, "/")
	out := make([]Segment, len(parts))
	for i, p := range parts {
		if strings.HasPrefix(p, "{") && strings.HasSuffix(p, "}") {
			out[i] = Segment{IsParam: true}
		} else {
			out[i] = Segment{IsParam: false, Literal: p}
		}
	}
	return out
}

// MoreSpecific reports whether a should be evaluated before b. Static
// segments beat param segments at the same position. Routes with different
// segment counts cannot shadow each other; tie-breaks fall back to a stable
// ordering that registration order preserves via sort.SliceStable.
func MoreSpecific(a, b []Segment) bool {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i].IsParam != b[i].IsParam {
			return !a[i].IsParam
		}
	}
	return len(a) > len(b)
}

// SameShape reports whether two segment lists describe the same route
// modulo param names — i.e. they would conflict in dispatch.
func SameShape(a, b []Segment) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].IsParam != b[i].IsParam {
			return false
		}
		if !a[i].IsParam && a[i].Literal != b[i].Literal {
			return false
		}
	}
	return true
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

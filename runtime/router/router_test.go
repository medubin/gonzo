package router_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/medubin/gonzo/runtime/middleware"
	"github.com/medubin/gonzo/runtime/router"
	"github.com/medubin/gonzo/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- RouteEntry.Match ---

func TestRouteEntry_Match_SimpleMatch(t *testing.T) {
	entry := router.RouteEntry{
		Path:   regexp.MustCompile(`^/hello/?$`),
		Method: "GET",
	}
	req := httptest.NewRequest("GET", "/hello", nil)
	params := entry.Match(req)
	assert.NotNil(t, params)
}

func TestRouteEntry_Match_NoMatch(t *testing.T) {
	entry := router.RouteEntry{
		Path:   regexp.MustCompile(`^/hello/?$`),
		Method: "GET",
	}
	req := httptest.NewRequest("GET", "/world", nil)
	params := entry.Match(req)
	assert.Nil(t, params)
}

func TestRouteEntry_Match_CapturesNamedGroups(t *testing.T) {
	entry := router.RouteEntry{
		Path:   regexp.MustCompile(`^/users/(?P<id>\w+)/?$`),
		Method: "GET",
	}
	req := httptest.NewRequest("GET", "/users/42", nil)
	params := entry.Match(req)
	require.NotNil(t, params)
	assert.Equal(t, "42", params["id"])
}

func TestRouteEntry_Match_MultipleParams(t *testing.T) {
	entry := router.RouteEntry{
		Path:   regexp.MustCompile(`^/users/(?P<userId>\w+)/posts/(?P<postId>\w+)/?$`),
		Method: "GET",
	}
	req := httptest.NewRequest("GET", "/users/123/posts/456", nil)
	params := entry.Match(req)
	require.NotNil(t, params)
	assert.Equal(t, "123", params["userId"])
	assert.Equal(t, "456", params["postId"])
}

// --- Router.ServeHTTP ---

func newRouteInfo(method, path string) *types.RouteInfo {
	return &types.RouteInfo{
		Method:   method,
		Path:     path,
		Endpoint: "TestEndpoint",
		Server:   "TestServer",
	}
}

func TestRouter_ServeHTTP_BasicRoute(t *testing.T) {
	rtr := &router.Router{}
	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}, newRouteInfo("GET", "/hello"))

	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRouter_ServeHTTP_NotFound(t *testing.T) {
	rtr := &router.Router{}

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouter_ServeHTTP_MethodNotMatched(t *testing.T) {
	rtr := &router.Router{}
	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, newRouteInfo("POST", "/hello"))

	req := httptest.NewRequest("GET", "/hello", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRouter_ServeHTTP_PathParams(t *testing.T) {
	rtr := &router.Router{}
	var capturedID string

	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		params := r.Context().Value(struct{}{})
		if p, ok := params.(map[string]string); ok {
			capturedID = p["id"]
		}
		w.WriteHeader(http.StatusOK)
	}, newRouteInfo("GET", "/users/{id}"))

	req := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "123", capturedID)
}

func TestRouter_ServeHTTP_WithJSONBody(t *testing.T) {
	rtr := &router.Router{}
	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}, &types.RouteInfo{
		Method:       "POST",
		Path:         "/items",
		Endpoint:     "CreateItem",
		Server:       "Test",
		RequiresBody: true,
	})

	body := `{"name":"thing"}`
	req := httptest.NewRequest("POST", "/items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRouter_ServeHTTP_RequiresBody_MissingBody(t *testing.T) {
	rtr := &router.Router{}
	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, &types.RouteInfo{
		Method:       "POST",
		Path:         "/items",
		Endpoint:     "CreateItem",
		Server:       "Test",
		RequiresBody: true,
	})

	req := httptest.NewRequest("POST", "/items", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRouter_ServeHTTP_NilRouteInfo_Panics(t *testing.T) {
	rtr := &router.Router{}
	assert.Panics(t, func() {
		rtr.Route(func(w http.ResponseWriter, r *http.Request) {}, nil)
	})
}

func TestRouter_Use_MiddlewareExecuted(t *testing.T) {
	rtr := &router.Router{}

	var beforeCalled bool
	m := &recordingMiddleware{
		onBefore: func() { beforeCalled = true },
	}
	rtr.Use(m)

	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, newRouteInfo("GET", "/test"))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.True(t, beforeCalled)
}

func TestRouter_ServeHTTP_InvalidJSON_Returns400(t *testing.T) {
	rtr := &router.Router{}
	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, newRouteInfo("POST", "/items"))

	body := `{invalid`
	req := httptest.NewRequest("POST", "/items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRouter_ServeHTTP_PanicRecovery(t *testing.T) {
	rtr := &router.Router{}
	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected error")
	}, newRouteInfo("GET", "/panic"))

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// Should not panic; router recovers
	assert.NotPanics(t, func() {
		rtr.ServeHTTP(w, req)
	})
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// --- recordingMiddleware helper ---

type recordingMiddleware struct {
	middleware.BaseMiddleware
	onBefore func()
}

func (m *recordingMiddleware) BeforeHandler(ctx context.Context, req *middleware.MiddlewareRequest, info *types.RouteInfo) (context.Context, *middleware.MiddlewareRequest, error) {
	if m.onBefore != nil {
		m.onBefore()
	}
	return ctx, req, nil
}

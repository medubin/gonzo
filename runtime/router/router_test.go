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

func TestRouter_Route_NilRouteInfo_ReturnsError(t *testing.T) {
	rtr := &router.Router{}
	err := rtr.Route(func(w http.ResponseWriter, r *http.Request) {}, nil)
	assert.Error(t, err)
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

func TestRouter_ServeHTTP_MiddlewareReceivesBody(t *testing.T) {
	rtr := &router.Router{}

	var capturedBody any
	m := &recordingMiddleware{
		onBeforeWithReq: func(req *middleware.MiddlewareRequest) {
			capturedBody = req.Body
		},
	}
	rtr.Use(m)

	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, newRouteInfo("POST", "/items"))

	body := `{"name":"thing"}`
	req := httptest.NewRequest("POST", "/items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	require.NotNil(t, capturedBody)
	raw, ok := capturedBody.(json.RawMessage)
	require.True(t, ok, "expected json.RawMessage, got %T", capturedBody)
	assert.JSONEq(t, body, string(raw))
}

func TestRouter_ServeHTTP_HandlerCanReadBufferedBody(t *testing.T) {
	rtr := &router.Router{}

	var handlerBody map[string]string
	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		err := json.NewDecoder(r.Body).Decode(&handlerBody)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}, newRouteInfo("POST", "/items"))

	body := `{"key":"value"}`
	req := httptest.NewRequest("POST", "/items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, map[string]string{"key": "value"}, handlerBody)
}

func TestRouter_ServeHTTP_NonJSONBodyPassesThrough(t *testing.T) {
	rtr := &router.Router{}

	var middlewareBody any
	m := &recordingMiddleware{
		onBeforeWithReq: func(req *middleware.MiddlewareRequest) {
			middlewareBody = req.Body
		},
	}
	rtr.Use(m)

	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, newRouteInfo("POST", "/items"))

	body := `name=thing`
	req := httptest.NewRequest("POST", "/items", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Nil(t, middlewareBody, "non-JSON body should not be parsed into middleware")
}

// --- responseWriter buffer behavior ---

// TestResponseWriter_MultipleWrites verifies that a handler calling Write
// multiple times produces a correctly concatenated response body — the key
// behavioral guarantee of the bytes.Buffer change.
func TestResponseWriter_MultipleWrites(t *testing.T) {
	rtr := &router.Router{}
	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"a":`))
		w.Write([]byte(`"hello"`))
		w.Write([]byte(`}`))
	}, newRouteInfo("GET", "/multi"))

	req := httptest.NewRequest("GET", "/multi", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"a":"hello"}`, strings.TrimSpace(w.Body.String()))
}

// TestResponseWriter_MultipleWrites_MiddlewareSeesFullBody verifies that
// AfterHandler middleware receives the fully assembled body when the handler
// writes in multiple chunks.
func TestResponseWriter_MultipleWrites_MiddlewareSeesFullBody(t *testing.T) {
	rtr := &router.Router{}

	var capturedBody any
	m := &recordingMiddleware{
		onAfterWithResp: func(resp *middleware.MiddlewareResponse) {
			capturedBody = resp.Body
		},
	}
	rtr.Use(m)

	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"part":`))
		w.Write([]byte(`"one"}`))
	}, newRouteInfo("GET", "/chunked"))

	req := httptest.NewRequest("GET", "/chunked", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	require.NotNil(t, capturedBody)
	bodyMap, ok := capturedBody.(map[string]any)
	require.True(t, ok, "expected map body, got %T", capturedBody)
	assert.Equal(t, "one", bodyMap["part"])
}

// TestResponseWriter_LargeBody verifies that a response significantly larger
// than a typical initial slice allocation is captured without truncation.
func TestResponseWriter_LargeBody(t *testing.T) {
	const payloadSize = 64 * 1024 // 64 KB

	rtr := &router.Router{}
	rtr.Route(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(strings.Repeat("x", payloadSize)))
	}, newRouteInfo("GET", "/large"))

	req := httptest.NewRequest("GET", "/large", nil)
	w := httptest.NewRecorder()
	rtr.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.GreaterOrEqual(t, w.Body.Len(), payloadSize)
}

// --- Route specificity & conflict detection ---

func makeRecordingHandler(name string, hits *map[string]int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		(*hits)[name]++
		w.WriteHeader(http.StatusOK)
	}
}

func TestRouter_Specificity_StaticBeatsParam_RegistrationOrderA(t *testing.T) {
	rtr := &router.Router{}
	hits := map[string]int{}
	require.NoError(t, rtr.Route(makeRecordingHandler("param", &hits), newRouteInfo("GET", "/ski-sessions/{id}")))
	require.NoError(t, rtr.Route(makeRecordingHandler("static", &hits), newRouteInfo("GET", "/ski-sessions/full")))

	rtr.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/ski-sessions/full", nil))
	rtr.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/ski-sessions/123", nil))

	assert.Equal(t, 1, hits["static"])
	assert.Equal(t, 1, hits["param"])
}

func TestRouter_Specificity_StaticBeatsParam_RegistrationOrderB(t *testing.T) {
	rtr := &router.Router{}
	hits := map[string]int{}
	require.NoError(t, rtr.Route(makeRecordingHandler("static", &hits), newRouteInfo("GET", "/ski-sessions/full")))
	require.NoError(t, rtr.Route(makeRecordingHandler("param", &hits), newRouteInfo("GET", "/ski-sessions/{id}")))

	rtr.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/ski-sessions/full", nil))
	rtr.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/ski-sessions/123", nil))

	assert.Equal(t, 1, hits["static"])
	assert.Equal(t, 1, hits["param"])
}

func TestRouter_Specificity_NestedPerPosition(t *testing.T) {
	for _, order := range []string{"AB", "BA"} {
		t.Run(order, func(t *testing.T) {
			rtr := &router.Router{}
			hits := map[string]int{}
			a := func() error {
				return rtr.Route(makeRecordingHandler("paramFirst", &hits), newRouteInfo("GET", "/a/{x}/b"))
			}
			b := func() error {
				return rtr.Route(makeRecordingHandler("staticFirst", &hits), newRouteInfo("GET", "/a/c/{y}"))
			}
			if order == "AB" {
				require.NoError(t, a())
				require.NoError(t, b())
			} else {
				require.NoError(t, b())
				require.NoError(t, a())
			}

			rtr.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/a/c/b", nil))
			rtr.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/a/z/b", nil))

			// /a/c/b: static "c" at position 1 wins → matches /a/c/{y}
			// /a/z/b: "z" doesn't match static "c" → matches /a/{x}/b
			assert.Equal(t, 1, hits["staticFirst"], "/a/c/b should hit /a/c/{y}")
			assert.Equal(t, 1, hits["paramFirst"], "/a/z/b should hit /a/{x}/b")
		})
	}
}

func TestRouter_Conflict_IdenticalShape_ReturnsError(t *testing.T) {
	rtr := &router.Router{}
	require.NoError(t, rtr.Route(func(w http.ResponseWriter, r *http.Request) {}, newRouteInfo("GET", "/a/{x}")))
	err := rtr.Route(func(w http.ResponseWriter, r *http.Request) {}, newRouteInfo("GET", "/a/{y}"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TestEndpoint")
	assert.Contains(t, err.Error(), "TestServer")
}

func TestRouter_SamePathDifferentMethods_Coexist(t *testing.T) {
	rtr := &router.Router{}
	hits := map[string]int{}
	require.NoError(t, rtr.Route(makeRecordingHandler("get", &hits), newRouteInfo("GET", "/a/{x}")))
	require.NoError(t, rtr.Route(makeRecordingHandler("del", &hits), newRouteInfo("DELETE", "/a/{x}")))

	rtr.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/a/1", nil))
	rtr.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("DELETE", "/a/2", nil))

	assert.Equal(t, 1, hits["get"])
	assert.Equal(t, 1, hits["del"])
}

// --- recordingMiddleware helper ---

type recordingMiddleware struct {
	middleware.BaseMiddleware
	onBefore        func()
	onBeforeWithReq func(*middleware.MiddlewareRequest)
	onAfterWithResp func(*middleware.MiddlewareResponse)
}

func (m *recordingMiddleware) BeforeHandler(ctx context.Context, req *middleware.MiddlewareRequest, info *types.RouteInfo) (context.Context, *middleware.MiddlewareRequest, error) {
	if m.onBefore != nil {
		m.onBefore()
	}
	if m.onBeforeWithReq != nil {
		m.onBeforeWithReq(req)
	}
	return ctx, req, nil
}

func (m *recordingMiddleware) AfterHandler(ctx context.Context, req *middleware.MiddlewareRequest, resp *middleware.MiddlewareResponse, info *types.RouteInfo) (*middleware.MiddlewareResponse, error) {
	if m.onAfterWithResp != nil {
		m.onAfterWithResp(resp)
	}
	return resp, nil
}

package handle_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/handle"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/url"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testBody struct {
	Name string `json:"name"`
}

type testResponse struct {
	Greeting string `json:"greeting"`
}

type testParams struct{}
type testPathParams struct{}

func TestHandle_SuccessfulRequest(t *testing.T) {
	handler := handle.Handle[testBody, testResponse, testParams, testPathParams](
		func(ctx context.Context, b *testBody, c cookies.Cookies, u url.URL[testParams, testPathParams]) (*handle.Response[testResponse], error) {
			return &handle.Response[testResponse]{Body: &testResponse{Greeting: "hello " + b.Name}}, nil
		},
	)

	body := `{"name":"world"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()

	handler(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp testResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "hello world", resp.Greeting)
}

func TestHandle_NilBody_WhenNoContent(t *testing.T) {
	var receivedBody *testBody

	handler := handle.Handle[testBody, testResponse, testParams, testPathParams](
		func(ctx context.Context, b *testBody, c cookies.Cookies, u url.URL[testParams, testPathParams]) (*handle.Response[testResponse], error) {
			receivedBody = b
			return &handle.Response[testResponse]{Body: &testResponse{Greeting: "ok"}}, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.ContentLength = 0
	w := httptest.NewRecorder()

	handler(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Nil(t, receivedBody)
}

func TestHandle_MalformedJSON(t *testing.T) {
	handler := handle.Handle[testBody, testResponse, testParams, testPathParams](
		func(ctx context.Context, b *testBody, c cookies.Cookies, u url.URL[testParams, testPathParams]) (*handle.Response[testResponse], error) {
			return &handle.Response[testResponse]{Body: &testResponse{}}, nil
		},
	)

	body := `{invalid json`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.ContentLength = int64(len(body))
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var errResp map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
	assert.Contains(t, errResp["error"], "malformed")
}

func TestHandle_HandlerError(t *testing.T) {
	handler := handle.Handle[testBody, testResponse, testParams, testPathParams](
		func(ctx context.Context, b *testBody, c cookies.Cookies, u url.URL[testParams, testPathParams]) (*handle.Response[testResponse], error) {
			return nil, gerrors.NotFoundError("thing not found")
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	var errResp map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
	assert.Contains(t, errResp["error"], "thing not found")
}

func TestHandle_UnknownContentLength_BodyParsed(t *testing.T) {
	// ContentLength == -1 (chunked / unknown) must not prevent body parsing.
	handler := handle.Handle[testBody, testResponse, testParams, testPathParams](
		func(ctx context.Context, b *testBody, c cookies.Cookies, u url.URL[testParams, testPathParams]) (*handle.Response[testResponse], error) {
			require.NotNil(t, b)
			return &handle.Response[testResponse]{Body: &testResponse{Greeting: "hello " + b.Name}}, nil
		},
	)

	body := `{"name":"chunked"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.ContentLength = -1 // unknown / chunked
	w := httptest.NewRecorder()

	handler(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	var resp testResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "hello chunked", resp.Greeting)
}

func TestHandle_SuccessResponse_HasJSONContentType(t *testing.T) {
	handler := handle.Handle[testBody, testResponse, testParams, testPathParams](
		func(ctx context.Context, b *testBody, c cookies.Cookies, u url.URL[testParams, testPathParams]) (*handle.Response[testResponse], error) {
			return &handle.Response[testResponse]{Body: &testResponse{Greeting: "ok"}}, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestHandle_QueryParamsParsed(t *testing.T) {
	type qParams struct {
		Page *string `json:"page,omitempty"`
	}

	var receivedParams url.URL[qParams, testPathParams]
	handler := handle.Handle[testBody, testResponse, qParams, testPathParams](
		func(ctx context.Context, b *testBody, c cookies.Cookies, u url.URL[qParams, testPathParams]) (*handle.Response[testResponse], error) {
			receivedParams = u
			return &handle.Response[testResponse]{Body: &testResponse{Greeting: "ok"}}, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/?page=3", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, receivedParams.Params.Page)
	assert.Equal(t, "3", *receivedParams.Params.Page)
}

func TestHandle_CustomSuccessStatus(t *testing.T) {
	handler := handle.Handle[testBody, testResponse, testParams, testPathParams](
		func(ctx context.Context, b *testBody, c cookies.Cookies, u url.URL[testParams, testPathParams]) (*handle.Response[testResponse], error) {
			return &handle.Response[testResponse]{Body: &testResponse{Greeting: "created"}, Status: http.StatusCreated}, nil
		},
	)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp testResponse
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "created", resp.Greeting)
}

func TestHandle_ErrorStatusInSuccessResponse_Returns500(t *testing.T) {
	handler := handle.Handle[testBody, testResponse, testParams, testPathParams](
		func(ctx context.Context, b *testBody, c cookies.Cookies, u url.URL[testParams, testPathParams]) (*handle.Response[testResponse], error) {
			// Programmer mistake: setting a 4xx code on a success response
			return &handle.Response[testResponse]{Body: &testResponse{}, Status: http.StatusNotFound}, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var errResp map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&errResp))
	assert.Contains(t, errResp["error"], "error status code")
}

func TestHandle_ZeroStatus_DefaultsTo200(t *testing.T) {
	handler := handle.Handle[testBody, testResponse, testParams, testPathParams](
		func(ctx context.Context, b *testBody, c cookies.Cookies, u url.URL[testParams, testPathParams]) (*handle.Response[testResponse], error) {
			return &handle.Response[testResponse]{Body: &testResponse{Greeting: "ok"}}, nil
		},
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

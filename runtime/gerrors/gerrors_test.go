package gerrors_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name           string
		fn             func(string) gerrors.GonzoError
		expectedCode   gerrors.ErrorCode
		expectedStatus int
	}{
		{"InvalidArgument", gerrors.InvalidArgumentError, gerrors.InvalidArgument, http.StatusBadRequest},
		{"MissingArgument", gerrors.MissingArgumentError, gerrors.MissingArgument, http.StatusBadRequest},
		{"NotFound", gerrors.NotFoundError, gerrors.NotFound, http.StatusNotFound},
		{"AlreadyExists", gerrors.AlreadyExistsError, gerrors.AlreadyExists, http.StatusConflict},
		{"Unauthenticated", gerrors.UnauthenticatedError, gerrors.Unauthenticated, http.StatusUnauthorized},
		{"Unimplemented", gerrors.UnimplementedError, gerrors.Unimplemented, http.StatusNotImplemented},
		{"Internal", gerrors.InternalError, gerrors.Internal, http.StatusInternalServerError},
		{"BadRoute", gerrors.BadRouteError, gerrors.BadRoute, http.StatusNotFound},
		{"Malformed", gerrors.MalformedError, gerrors.Malformed, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn("test message")
			assert.Equal(t, tt.expectedCode, err.Code())
			assert.Equal(t, tt.expectedStatus, err.StatusCode())
			assert.Contains(t, err.Error(), "test message")
			assert.Contains(t, err.Error(), string(tt.expectedCode))
		})
	}
}

func TestGonzoError_Error(t *testing.T) {
	err := gerrors.NotFoundError("user not found")
	assert.Equal(t, "not_found: user not found", err.Error())
}

func TestGonzoError_MarshalJSON(t *testing.T) {
	err := gerrors.NotFoundError("user not found")
	ge, ok := err.(json.Marshaler)
	require.True(t, ok)

	data, jsonErr := ge.MarshalJSON()
	require.NoError(t, jsonErr)

	var result map[string]string
	require.NoError(t, json.Unmarshal(data, &result))
	assert.Equal(t, "not_found: user not found", result["error"])
}

func TestToGonzoError_WithGonzoError(t *testing.T) {
	original := gerrors.NotFoundError("not found")
	converted := gerrors.ToGonzoError(original)
	assert.Equal(t, original.Code(), converted.Code())
	assert.Equal(t, original.StatusCode(), converted.StatusCode())
	assert.Equal(t, original.Error(), converted.Error())
}

func TestToGonzoError_WithGenericError(t *testing.T) {
	generic := errors.New("something went wrong")
	converted := gerrors.ToGonzoError(generic)
	assert.Equal(t, gerrors.Internal, converted.Code())
	assert.Equal(t, http.StatusInternalServerError, converted.StatusCode())
	assert.Contains(t, converted.Error(), "something went wrong")
}

func TestJSONError_WritesCorrectStatusCode(t *testing.T) {
	w := httptest.NewRecorder()
	gerrors.JSONError(w, gerrors.NotFoundError("not found"))
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestJSONError_WritesJSONContentType(t *testing.T) {
	w := httptest.NewRecorder()
	gerrors.JSONError(w, gerrors.InvalidArgumentError("bad input"))
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))
}

func TestJSONError_WritesErrorBody(t *testing.T) {
	w := httptest.NewRecorder()
	gerrors.JSONError(w, gerrors.NotFoundError("thing not found"))

	var result map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Contains(t, result["error"], "thing not found")
}

func TestJSONError_WithGenericError(t *testing.T) {
	w := httptest.NewRecorder()
	gerrors.JSONError(w, errors.New("raw error"))
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWriteJSON_HappyPath(t *testing.T) {
	w := httptest.NewRecorder()
	gerrors.WriteJSON(w, http.StatusCreated, map[string]string{"name": "alice"})

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var result map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Equal(t, "alice", result["name"])
}

func TestWriteJSON_NilBody(t *testing.T) {
	w := httptest.NewRecorder()
	gerrors.WriteJSON(w, http.StatusOK, nil)
	assert.Equal(t, http.StatusOK, w.Code)
	// Nil encodes to "null\n" — that's fine; we only assert the status got through.
}

// TestWriteJSON_EncodeError_StillSendsErrorResponse verifies that when the body
// fails to marshal, the client receives a 500 instead of a successful status
// with corrupt bytes. Channels are unmarshalable in encoding/json, so we use
// one to force the encode failure.
func TestWriteJSON_EncodeError_StillSendsErrorResponse(t *testing.T) {
	w := httptest.NewRecorder()
	bad := struct {
		Ch chan int `json:"ch"`
	}{Ch: make(chan int)}

	gerrors.WriteJSON(w, http.StatusOK, bad)

	// Status should be 500 from JSONError, not 200 from the original call.
	assert.Equal(t, http.StatusInternalServerError, w.Code,
		"encode failure must produce a 500, not the original status")

	// And the body must be a valid Gonzo error envelope, not partial garbage.
	var result map[string]string
	require.NoError(t, json.NewDecoder(w.Body).Decode(&result))
	assert.Contains(t, result["error"], "internal")
}

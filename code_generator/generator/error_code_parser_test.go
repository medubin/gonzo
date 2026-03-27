package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const minimalErrorsGoSrc = `package gerrors

import "net/http"

type ErrorCode string

const (
	NotFound   ErrorCode = "not_found"
	BadRequest ErrorCode = "bad_request"
	Internal   ErrorCode = "internal"
)

type GonzoError interface{ Error() string }

type gerr struct{}

func (g gerr) Error() string { return "" }

func newError(code ErrorCode, msg string, status int) GonzoError {
	return gerr{}
}

func NotFoundError(msg string) GonzoError {
	return newError(NotFound, msg, http.StatusNotFound)
}

func BadRequestError(msg string) GonzoError {
	return newError(BadRequest, msg, http.StatusBadRequest)
}

func InternalError(msg string) GonzoError {
	return newError(Internal, msg, http.StatusInternalServerError)
}

// unexported — should be ignored
func helperFunc(msg string) GonzoError {
	return newError(Internal, msg, http.StatusInternalServerError)
}

// returns a non-GonzoError type — should be ignored
func SomeOtherFunc(msg string) error {
	return nil
}
`

func TestParseErrorCodesFromFile_BasicParsing(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "errors.go")
	require.NoError(t, os.WriteFile(tmpFile, []byte(minimalErrorsGoSrc), 0644))

	codes, err := parseErrorCodesFromFile(tmpFile)
	require.NoError(t, err)
	require.Len(t, codes, 3, "should extract exactly the 3 exported GonzoError constructors")

	byName := make(map[string]TemplateErrorCode)
	for _, c := range codes {
		byName[c.ClassName] = c
	}

	assert.Equal(t, "not_found", byName["NotFoundError"].Code)
	assert.Equal(t, 404, byName["NotFoundError"].StatusCode)

	assert.Equal(t, "bad_request", byName["BadRequestError"].Code)
	assert.Equal(t, 400, byName["BadRequestError"].StatusCode)

	assert.Equal(t, "internal", byName["InternalError"].Code)
	assert.Equal(t, 500, byName["InternalError"].StatusCode)
}

func TestParseErrorCodesFromFile_SkipsUnexportedAndNonGonzoErrorFuncs(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "errors.go")
	require.NoError(t, os.WriteFile(tmpFile, []byte(minimalErrorsGoSrc), 0644))

	codes, err := parseErrorCodesFromFile(tmpFile)
	require.NoError(t, err)

	classNames := make(map[string]bool)
	for _, c := range codes {
		classNames[c.ClassName] = true
	}

	assert.False(t, classNames["helperFunc"], "unexported function should be excluded")
	assert.False(t, classNames["SomeOtherFunc"], "function not returning GonzoError should be excluded")
}

func TestParseErrorCodesFromFile_FileNotFound(t *testing.T) {
	_, err := parseErrorCodesFromFile("/nonexistent/path/to/errors.go")
	assert.Error(t, err)
}

func TestParseErrorCodesFromFile_InvalidGoSyntax(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "errors.go")
	require.NoError(t, os.WriteFile(tmpFile, []byte("this is not { valid go code"), 0644))

	_, err := parseErrorCodesFromFile(tmpFile)
	assert.Error(t, err)
}

func TestParseErrorCodesFromFile_EmptyFile(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "errors.go")
	require.NoError(t, os.WriteFile(tmpFile, []byte("package gerrors\n"), 0644))

	codes, err := parseErrorCodesFromFile(tmpFile)
	require.NoError(t, err)
	assert.Empty(t, codes, "file with no error constructors should return empty slice")
}

package gerrors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func JSONError(w http.ResponseWriter, err error) {
	gerr := ToGonzoError(err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(gerr.StatusCode())
	json.NewEncoder(w).Encode(gerr)
}

func ToGonzoError(err error) GonzoError {
	ge, ok := err.(GonzoError)
	if ok {
		return ge
	}
	return newError(Internal, err.Error(), http.StatusInternalServerError)
}

type GonzoError interface {
	Error() string
	Code() ErrorCode
	StatusCode() int
}

type gerr struct {
	code       ErrorCode
	msg        string
	statusCode int
}

func (g gerr) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"error": g.Error(),
	})
}

func (e gerr) Error() string {
	return fmt.Sprintf("%s: %s", e.code, e.msg)
}

func (e gerr) Code() ErrorCode {
	return e.code
}

func (e gerr) StatusCode() int {
	return e.statusCode
}

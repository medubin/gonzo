package gerrors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func JSONError(w http.ResponseWriter, err error) {
	gerr := ToGonzoError(err)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(gerr.StatusCode())
	json.NewEncoder(w).Encode(gerr)
}

// WriteJSON encodes body and writes status + body atomically: encoding into a
// buffer first so that a marshal failure can still produce a 500 response,
// and committing the status only after encoding succeeds. A failure to write
// the buffered bytes to the wire is unrecoverable (status already sent), so
// it is logged and dropped — same as a network error mid-stream.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		JSONError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if _, err := buf.WriteTo(w); err != nil {
		log.Printf("gerrors.WriteJSON: write failed after status sent: %v", err)
	}
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

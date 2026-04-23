package handle

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/form"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/url"
)

// Response wraps a handler's return value with an optional HTTP success status code.
// If Status is 0 or unset, the handler defaults to 200. Status codes >= 400 are
// rejected at runtime to prevent accidentally returning error codes as successes.
type Response[T any] struct {
	Body   *T
	Status int
}

func Handle[Body any, response any, Params any, PathParams any](handler func(ctx context.Context, b *Body, c cookies.Cookies, u url.URL[Params, PathParams]) (*Response[response], error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var body *Body

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil && err != io.EOF {
			gerr := gerrors.MalformedError(err.Error())
			gerrors.JSONError(w, gerr)
			return
		}

		cookies := cookies.New(r, w)

		Url := url.URL[Params, PathParams]{
			Params:     url.GetTypedParamsFromQuery[Params](r.URL.Query()),
			PathParams: url.GetTypedParamsFromContext[PathParams](ctx),
		}

		resp, err := handler(ctx, body, cookies, Url)
		if err != nil {
			gerrors.JSONError(w, err)
			return
		}

		statusCode := http.StatusOK
		if resp != nil && resp.Status != 0 {
			if resp.Status >= 400 {
				gerrors.JSONError(w, gerrors.InternalError("handler set an error status code on a success response"))
				return
			}
			statusCode = resp.Status
		}

		var respBody *response
		if resp != nil {
			respBody = resp.Body
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(statusCode)
		err = json.NewEncoder(w).Encode(respBody)
		if err != nil {
			gerrors.JSONError(w, err)
			return
		}
	}
}

func HandleMultipart[Body any, response any, Params any, PathParams any](handler func(ctx context.Context, b *Body, c cookies.Cookies, u url.URL[Params, PathParams]) (*Response[response], error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		body, err := form.Parse[Body](r)
		if err != nil {
			gerrors.JSONError(w, err)
			return
		}

		c := cookies.New(r, w)
		Url := url.URL[Params, PathParams]{
			Params:     url.GetTypedParamsFromQuery[Params](r.URL.Query()),
			PathParams: url.GetTypedParamsFromContext[PathParams](ctx),
		}

		resp, err := handler(ctx, body, c, Url)
		if err != nil {
			gerrors.JSONError(w, err)
			return
		}

		statusCode := http.StatusOK
		if resp != nil && resp.Status != 0 {
			if resp.Status >= 400 {
				gerrors.JSONError(w, gerrors.InternalError("handler set an error status code on a success response"))
				return
			}
			statusCode = resp.Status
		}

		var respBody *response
		if resp != nil {
			respBody = resp.Body
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(statusCode)
		if err = json.NewEncoder(w).Encode(respBody); err != nil {
			gerrors.JSONError(w, err)
			return
		}
	}
}

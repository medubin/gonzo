package handle

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/medubin/gonzo/runtime/cookies"
	"github.com/medubin/gonzo/runtime/gerrors"
	"github.com/medubin/gonzo/runtime/url"
)

func Handle[Body any, response any, Params any, PathParams any](handler func(ctx context.Context, b *Body, c cookies.Cookies, u url.URL[Params, PathParams]) (*response, error)) func(http.ResponseWriter, *http.Request) {
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

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			gerrors.JSONError(w, err)
			return
		}
	}
}

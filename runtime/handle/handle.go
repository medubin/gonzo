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
		
		// Only attempt to decode if there might be JSON content
		if r.ContentLength != 0 {
			err := json.NewDecoder(r.Body).Decode(&body)
			if err != nil {
				if err == io.EOF {
					// Empty body, body remains nil
				} else {
					gerr := gerrors.MalformedError(err.Error())
					gerrors.JSONError(w, gerr)
					return
				}
			}
		}
		// If no content or EOF, body remains nil

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

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			gerrors.JSONError(w, err)
			return
		}
	}
}

package handle

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/medubin/gonzo/api/src/cookies"
	"github.com/medubin/gonzo/api/src/url"
)

func Handle[Body any, response any, URL any](handler func(ctx context.Context, b *Body, c cookies.Cookies, u url.URL[URL]) (response, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var body *Body
		err := json.NewDecoder(r.Body).Decode(&body)

		if err != nil && err != io.EOF {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cookies := cookies.New(r, w)

		Url := url.URL[URL]{
			Values: url.Values(r.URL.Query()),
			Params: url.GetTypedParamsFromContext[URL](ctx),
		}

		resp, err := handler(ctx, body, cookies, Url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}

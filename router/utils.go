package router

import (
	"context"
	"encoding/json"
	"net/http"
)

// URLParam extracts a parameter from the URL by name
func URLParam(r *http.Request, name string) string {
	ctx := r.Context()

	// ctx.Value returns an `interface{}` type, so we
	// also have to cast it to a map, which is the
	// type we'll be using to store our parameters.
	params := ctx.Value("params").(map[string]string)
	return params[name]
}

func Handle[Body any, response any](handler func(ctx context.Context, b Body, c Cookies) (response, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var body Body

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cookies := Cookies{
			r: r,
			w: w,
		}

		resp, err := handler(r.Context(), body, cookies)
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

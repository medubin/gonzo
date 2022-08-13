package handle

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/medubin/gonzo/utils/cookies"
)

func Handle[Body any, response any](handler func(ctx context.Context, b Body, c cookies.Cookies) (response, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var body Body

		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cookies := cookies.New(r, w)

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

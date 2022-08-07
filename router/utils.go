package router

import "net/http"

// URLParam extracts a parameter from the URL by name
func URLParam(r *http.Request, name string) string {
	ctx := r.Context()

	// ctx.Value returns an `interface{}` type, so we
	// also have to cast it to a map, which is the 
	// type we'll be using to store our parameters.
	params := ctx.Value("params").(map[string]string)
	return params[name]
}
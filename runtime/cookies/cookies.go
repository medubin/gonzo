package cookies

import (
	"log"
	"net/http"
)

type Cookies struct {
	r *http.Request
	w http.ResponseWriter
}

func New(r *http.Request, w http.ResponseWriter) Cookies {
	if r == nil {
		log.Printf("Warning: cookies.New() called with nil http.Request - cookie operations will be disabled")
	}
	if w == nil {
		log.Printf("Warning: cookies.New() called with nil http.ResponseWriter - cookie setting will be disabled")
	}
	return Cookies{r, w}
}

func (c *Cookies) Get(name string) (*http.Cookie, error) {
	if c.r == nil {
		return nil, http.ErrNoCookie
	}
	return c.r.Cookie(name)
}

func (c *Cookies) All() []*http.Cookie {
	if c.r == nil {
		return nil
	}
	return c.r.Cookies()
}

func (c *Cookies) Set(cookie *http.Cookie) {
	if c.w == nil {
		log.Printf("Warning: attempting to set cookie %s but ResponseWriter is nil", cookie.Name)
		return
	}
	http.SetCookie(c.w, cookie)
}

// Opt adjusts a *http.Cookie at the call site. Used by generated SetXxx
// helpers so callers can extend a cookie (e.g., MaxAge, Path) without losing
// the security attributes that the .api declaration baked in.
type Opt func(*http.Cookie)

// MaxAge returns an Opt that sets the cookie's Max-Age in seconds. Pass a
// negative value to delete the cookie.
func MaxAge(seconds int) Opt {
	return func(c *http.Cookie) { c.MaxAge = seconds }
}

// Path scopes the cookie to a URL path prefix.
func Path(path string) Opt {
	return func(c *http.Cookie) { c.Path = path }
}

// Domain scopes the cookie to a domain (and its subdomains).
func Domain(domain string) Opt {
	return func(c *http.Cookie) { c.Domain = domain }
}

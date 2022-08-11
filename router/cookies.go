package router

import (
	"net/http"
)

type Cookies struct {
	r *http.Request
	w http.ResponseWriter
}

func (c *Cookies) Get(name string) (*http.Cookie, error) {
	return c.r.Cookie(name)
}

func (c *Cookies) All() []*http.Cookie {
	return c.r.Cookies()
}

func (c *Cookies) Set(cookie *http.Cookie) {
	http.SetCookie(c.w, cookie)
}
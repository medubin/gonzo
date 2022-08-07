package main

import (
	"net/http"

	"github.com/medubin/gonzo/internal/server"
	"github.com/medubin/gonzo/internal/utils"
	"github.com/medubin/gonzo/router"
)

func main() {

	r := &router.Router{}
	r.Route(http.MethodGet, "/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("The Best Router!"))
	})

	r.Route(http.MethodGet, `/hello/(?P<Message>\w+)`, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello " + router.URLParam(r, "Message")))
	})

	r.Route(http.MethodGet, "/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("something bad happened!")
	})
	r.Route(http.MethodGet, "/task/", utils.Handle(server.ServeIt))
	http.ListenAndServe(":8080", r)

	// s := server.NewServer()
	// mux := http.NewServeMux()

	// log.Fatal(http.ListenAndServe("localhost:8080", mux))

	// http.ListenAndServe(":8080", nil)
}

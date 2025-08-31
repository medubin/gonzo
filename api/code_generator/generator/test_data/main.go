package main

import (
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/medubin/gonzo/api/code_generator/generator/test_data/server"
	"github.com/medubin/gonzo/api/src/router"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

}

func run() error {
	println("Starting server")

	r := &router.Router{}
	s := &server.UserServiceImpl{}

	server.StartUserService(s, r)
	err := http.ListenAndServe(":8080", r)
	return err
}

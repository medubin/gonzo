package main

import (
	"log"
	"net/http"

	"github.com/medubin/gonzo/internal/server"
	"github.com/medubin/gonzo/internal/utils"
)

func main() {
	// s := server.NewServer()
	mux := http.NewServeMux()
	mux.HandleFunc("/task/", utils.Handle(server.ServeIt))

	log.Fatal(http.ListenAndServe("localhost:8080", mux))

	// http.ListenAndServe(":8080", nil)
}

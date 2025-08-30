package main

import (
	"log"
	"net/http"

	"database/sql"

	_ "github.com/lib/pq"
	"github.com/medubin/gonzo/api/src/router"
	"github.com/medubin/gonzo/db/queries"
	"github.com/medubin/gonzo/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

}

func run() error {
	println("Starting server")
	db, err := sql.Open("postgres", "user=postgres dbname=gonzo sslmode=disable")
	if err != nil {
		return err
	}

	queries := queries.New(db)
	r := &router.Router{}
	s := &server.GonzoServerImpl{
		Queries: *queries,
	}

	server.StartGonzoServer(s, r)
	err = http.ListenAndServe(":8080", r)
	return err
}

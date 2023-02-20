package main

import (
	"log"
	"net/http"

	"database/sql"

	_ "github.com/lib/pq"
	"github.com/medubin/gonzo/api/utils/router"
	"github.com/medubin/gonzo/db/queries"
	"github.com/medubin/gonzo/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

}

func run() error {
	db, err := sql.Open("postgres", "user=postgres dbname=gonzo sslmode=disable")
	if err != nil {
		return err
	}

	queries := queries.New(db)
	r := &router.Router{}
	s := &server.ServerImpl{
		Queries: *queries,
	}

	server.StartServer(s, r)
	http.ListenAndServe(":8080", r)
	return nil
}

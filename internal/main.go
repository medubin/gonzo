package main

import (
	"log"
	"net/http"

	"database/sql"

	_ "github.com/lib/pq"
	"github.com/medubin/gonzo/api/src/router"
	"github.com/medubin/gonzo/db/queries"
	"github.com/medubin/gonzo/internal/middleware"
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
	
	// Setup auth middleware
	authMiddleware := middleware.NewAuthMiddleware(db)
	r.Use(authMiddleware)
	
	s := &server.GonzoServerImpl{
		Queries: queries,
	}

	server.StartGonzoServer(s, r)
	
	log.Println("Server starting on :8080")
	err = http.ListenAndServe(":8080", r)
	return err
}

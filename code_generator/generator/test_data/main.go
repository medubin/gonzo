package main

import (
	"log"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/medubin/gonzo/code_generator/generator/test_data/server"
	"github.com/medubin/gonzo/runtime/middleware"
	"github.com/medubin/gonzo/runtime/router"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}

}

func run() error {
	println("Starting server")

	r := &router.Router{}
	
	// Add middleware
	loggingMiddleware := middleware.NewLoggingMiddleware()
	r.Use(loggingMiddleware)
	
	// corsMiddleware := middleware.NewCORSMiddleware(
	// 	[]string{"*"}, 
	// 	[]string{"GET", "POST", "PUT", "DELETE"}, 
	// 	[]string{"Content-Type", "Authorization"},
	// )
	// r.Use(corsMiddleware)
	
	// authMiddleware := middleware.NewAuthMiddleware("/users")
	// r.Use(authMiddleware)
	
	// errorMiddleware := middleware.NewErrorHandlerMiddleware(true)
	// r.Use(errorMiddleware)

	s := &server.UserServiceImpl{}

	server.StartUserService(s, r)
	err := http.ListenAndServe(":8080", r)
	return err
}

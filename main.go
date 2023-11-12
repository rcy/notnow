package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"yikes/db"
	"yikes/routes"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	ctx := context.Background()

	db.MustConnect(ctx, os.Getenv("DATABASE_URL"))
	defer db.Conn.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Route("/", routes.Router)

	// Start the server
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

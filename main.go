package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"yikes/db"
	"yikes/jobs/rescheduler"
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

	go rescheduler.Loop()

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8088"
	}
	log.Printf("Server started on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))

}

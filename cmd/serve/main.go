package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sirodoht/prophet/internal"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// debug mode
	debugMode := os.Getenv("DEBUG")

	// database connection
	databaseURL := os.Getenv("DATABASE_URL")
	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		panic(err)
	}

	// instantiate
	store := internal.NewSQLStore(db)
	handlers := internal.NewHandlers(store)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// dashboard
	r.Get("/dashboard", handlers.RenderDashboard)

	// resource posts
	r.Get("/", handlers.RenderAllPost)
	r.Get("/new/post", handlers.RenderNewPost)
	r.Post("/new/post", handlers.SaveNewPost)
	r.Get("/posts/{id}", handlers.RenderOnePost)

	// static files
	if debugMode == "1" {
		fileServer := http.FileServer(http.Dir("./static/"))
		r.Handle("/static/*", http.StripPrefix("/static", fileServer))
	}

	// serve
	fmt.Println("Listening on http://127.0.0.1:8000/")
	srv := &http.Server{
		Handler:      r,
		Addr:         ":8000",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

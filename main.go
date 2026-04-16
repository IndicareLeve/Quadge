package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Quadge - Quadlet Manager"))
	})

	port := os.Getenv("QUADGE_PORT")
	if port == "" {
		port = "4440"
	}

	fmt.Printf("Quadge starting on :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

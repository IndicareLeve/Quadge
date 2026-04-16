package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed templates/*
var templatesFS embed.FS

var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.ParseFS(templatesFS, "templates/*.html", "templates/fragments/*.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing templates: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Services": []interface{}{},
			"Selected": "",
		}
		tmpl.ExecuteTemplate(w, "index.html", data)
	})

	r.Get("/new", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "new.html", nil)
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

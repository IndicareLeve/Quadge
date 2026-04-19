package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/indicareleve/quadge/handlers"
)

//go:embed templates/*
var templatesFS embed.FS

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	tmpl, err := handlers.ParseTemplates(templatesFS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing templates: %v\n", err)
		os.Exit(1)
	}
	handlers.Tmpl = tmpl

	r.Get("/", handlers.ListServices)
	r.Get("/new", handlers.NewServiceFragment)
	r.Post("/convert", handlers.ConvertCompose)
	r.Post("/deploy", handlers.DeployService)
	r.Get("/service/{name}", handlers.GetService)
	r.Post("/start", handlers.StartService)
	r.Post("/stop", handlers.StopService)
	r.Post("/restart", handlers.RestartService)
	r.Post("/delete", handlers.DeleteService)
	r.Get("/logs", handlers.StreamLogs)
	r.Get("/edit", handlers.EditServiceForm)
	r.Post("/edit", handlers.EditService)

	port := os.Getenv("QUADGE_PORT")
	if port == "" {
		port = "4440"
	}

	fmt.Printf("Quadge starting on 0.0.0.0:%s\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, r); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

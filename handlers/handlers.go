package handlers

import (
	"embed"
	"html/template"

	"github.com/indicareleve/quadge/system"
)

var Tmpl *template.Template

type Service struct {
	Name           string
	Status         system.ServiceStatus
	QuadletContent string
}

type PageData struct {
	Services []Service
	Selected string
	Service  *Service
}

type ConvertResult struct {
	Files       []QuadletFileView
	EncodedData string
}

type QuadletFileView struct {
	Name    string
	Content string
}

func ParseTemplates(fs embed.FS) (*template.Template, error) {
	return template.ParseFS(fs, "templates/*.html", "templates/fragments/*.html")
}

func ToServices(files []system.QuadletFile) ([]Service, error) {
	services := make([]Service, 0, len(files))
	for _, f := range files {
		status, _ := system.GetServiceStatus(f.Name)
		services = append(services, Service{
			Name:           f.Name,
			Status:         status,
			QuadletContent: f.Content,
		})
	}
	return services, nil
}

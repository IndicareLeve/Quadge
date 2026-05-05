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

type GroupStatus string

const (
	GroupStatusRunning  GroupStatus = "running"
	GroupStatusDegraded GroupStatus = "degraded"
	GroupStatusStopped  GroupStatus = "stopped"
	GroupStatusFailed   GroupStatus = "failed"
)

type ServiceGroup struct {
	Name    string
	PodFile QuadletFileView
	Members []Service
	Status  GroupStatus
}

type PageData struct {
	Services   []Service
	Groups     []ServiceGroup
	Standalone []Service
	Selected   string
	Service    *Service
	Edit       *EditData
}

type EditData struct {
	Name    string
	Content string
	Ext     string
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

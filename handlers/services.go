package handlers

import (
	"net/http"

	"github.com/indicareleve/quadge/system"
)

func ListServices(w http.ResponseWriter, r *http.Request) {
	files, err := system.ListQuadletFiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	services, err := ToServices(files)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := PageData{
		Services: services,
	}

	Tmpl.ExecuteTemplate(w, "index.html", data)
}

func GetService(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	file, err := system.ReadQuadletFile(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	status, _ := system.GetServiceStatus(name)

	files, _ := system.ListQuadletFiles()
	services, _ := ToServices(files)

	data := PageData{
		Services: services,
		Selected: name,
		Service: &Service{
			Name:           file.Name,
			Status:         status,
			QuadletContent: file.Content,
		},
	}

	Tmpl.ExecuteTemplate(w, "index.html", data)
}

func StartService(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("service")
	if name == "" {
		http.Error(w, "service name required", http.StatusBadRequest)
		return
	}

	if err := system.StartService(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderServiceList(w, name)
}

func StopService(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("service")
	if name == "" {
		http.Error(w, "service name required", http.StatusBadRequest)
		return
	}

	if err := system.StopService(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderServiceList(w, name)
}

func RestartService(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("service")
	if name == "" {
		http.Error(w, "service name required", http.StatusBadRequest)
		return
	}

	if err := system.RestartService(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderServiceList(w, name)
}

func renderServiceList(w http.ResponseWriter, selected string) {
	files, _ := system.ListQuadletFiles()
	services, _ := ToServices(files)

	data := PageData{
		Services: services,
		Selected: selected,
	}

	Tmpl.ExecuteTemplate(w, "service-list", data)
}

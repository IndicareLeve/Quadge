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

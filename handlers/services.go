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

	tree := system.BuildServiceTree(files)
	groups := toServiceGroups(tree.Groups)
	standalone := toServices(tree.Standalone)

	data := PageData{
		Groups:     groups,
		Standalone: standalone,
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
	tree := system.BuildServiceTree(files)
	groups := toServiceGroups(tree.Groups)
	standalone := toServices(tree.Standalone)

	data := PageData{
		Groups:     groups,
		Standalone: standalone,
		Selected:   name,
		Service: &Service{
			Name:           file.Name,
			Status:         status,
			QuadletContent: file.Content,
		},
	}

	Tmpl.ExecuteTemplate(w, "main-content", data)
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
	tree := system.BuildServiceTree(files)
	groups := toServiceGroups(tree.Groups)
	standalone := toServices(tree.Standalone)

	data := PageData{
		Groups:     groups,
		Standalone: standalone,
		Selected:   selected,
	}

	Tmpl.ExecuteTemplate(w, "service-list", data)
}

func toServiceGroups(groups []system.ServiceGroup) []ServiceGroup {
	result := make([]ServiceGroup, 0, len(groups))
	for _, g := range groups {
		members := toServices(g.Members)
		result = append(result, ServiceGroup{
			Name:    g.Name,
			Members: members,
			Status:  calcGroupStatus(members),
		})
	}
	return result
}

func toServices(files []system.QuadletFile) []Service {
	services := make([]Service, 0, len(files))
	for _, f := range files {
		status, _ := system.GetServiceStatus(f.Name)
		services = append(services, Service{
			Name:           f.Name,
			Status:         status,
			QuadletContent: f.Content,
		})
	}
	return services
}

func calcGroupStatus(members []Service) GroupStatus {
	if len(members) == 0 {
		return GroupStatusStopped
	}
	hasRunning := false
	hasStopped := false
	for _, m := range members {
		if m.Status == system.StatusFailed {
			return GroupStatusFailed
		}
		if m.Status == system.StatusRunning {
			hasRunning = true
		} else {
			hasStopped = true
		}
	}
	if hasRunning && hasStopped {
		return GroupStatusDegraded
	}
	if hasRunning {
		return GroupStatusRunning
	}
	return GroupStatusStopped
}

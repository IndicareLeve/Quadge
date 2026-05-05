package handlers

import (
	"net/http"

	"github.com/indicareleve/quadge/system"
)

func EditServiceForm(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("service")
	if name == "" {
		http.Error(w, "service name required", http.StatusBadRequest)
		return
	}

	file, err := system.ReadQuadletFile(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	files, _ := system.ListQuadletFiles()
	tree := system.BuildServiceTree(files)
	parentGroup := findParentGroup(name, tree.Groups)

	data := PageData{
		Selected: name,
		Service: &Service{
			Name:           file.Name,
			QuadletContent: file.Content,
			ParentGroup:    parentGroup,
		},
		Edit: &EditData{
			Name:    file.Name,
			Content: file.Content,
			Ext:     "." + file.Type,
		},
	}

	Tmpl.ExecuteTemplate(w, "edit-service", data)
}

func EditService(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	content := r.FormValue("content")

	if name == "" || content == "" {
		http.Error(w, "name and content required", http.StatusBadRequest)
		return
	}

	file, err := system.ReadQuadletFile(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err := system.WriteQuadletFile(name, "."+file.Type, content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := system.DaemonReload(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	system.RestartService(name)

	status, _ := system.GetServiceStatus(name)
	files, _ := system.ListQuadletFiles()
	tree := system.BuildServiceTree(files)
	parentGroup := findParentGroup(name, tree.Groups)
	groups := toServiceGroups(tree.Groups)
	standalone := toServices(tree.Standalone)

	data := PageData{
		Groups:     groups,
		Standalone: standalone,
		Selected:   name,
		Service: &Service{
			Name:           name,
			Status:         status,
			QuadletContent: content,
			ParentGroup:    parentGroup,
		},
	}

	Tmpl.ExecuteTemplate(w, "service-detail", data)
}

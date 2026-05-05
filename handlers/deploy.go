package handlers

import (
	"net/http"

	"github.com/indicareleve/quadge/system"
)

func DeployService(w http.ResponseWriter, r *http.Request) {
	encoded := r.FormValue("quadlet-data")
	if encoded == "" {
		http.Error(w, "no quadlet data", http.StatusBadRequest)
		return
	}

	files, err := system.DecodeQuadletData(encoded)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, f := range files {
		if err := system.WriteQuadletFile(f.Name, f.Ext, f.Content); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if err := system.DaemonReload(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	podName := findPodName(files)
	if podName != "" {
		system.StartService(podName)
	} else {
		for _, f := range files {
			system.StartService(f.Name)
		}
	}

	allFiles, _ := system.ListQuadletFiles()
	tree := system.BuildServiceTree(allFiles)
	groups := toServiceGroups(tree.Groups)
	standalone := toServices(tree.Standalone)

	selected := files[0].Name
	data := PageData{
		Groups:     groups,
		Standalone: standalone,
		Selected:   selected,
	}

	Tmpl.ExecuteTemplate(w, "main-content", data)
}

func findPodName(files []system.QuadletResult) string {
	for _, f := range files {
		if f.Ext == ".pod" {
			return f.Name
		}
	}
	return ""
}

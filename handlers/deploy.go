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

	for _, f := range files {
		system.StartService(f.Name)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

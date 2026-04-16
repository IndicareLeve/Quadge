package handlers

import (
	"net/http"

	"github.com/indicareleve/quadge/system"
)

func DeleteService(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("service")
	if name == "" {
		http.Error(w, "service name required", http.StatusBadRequest)
		return
	}

	system.StopService(name)

	if err := system.DeleteQuadletFile(name); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := system.DaemonReload(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderServiceList(w, "")
}

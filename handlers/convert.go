package handlers

import (
	"net/http"

	"github.com/indicareleve/quadge/system"
)

func NewServiceFragment(w http.ResponseWriter, r *http.Request) {
	Tmpl.ExecuteTemplate(w, "new-service", nil)
}

func ConvertCompose(w http.ResponseWriter, r *http.Request) {
	compose := r.FormValue("compose")
	if compose == "" {
		http.Error(w, "compose content required", http.StatusBadRequest)
		return
	}

	results, err := system.ConvertCompose(compose)
	if err != nil {
		Tmpl.ExecuteTemplate(w, "quadlet-preview", ConvertResult{
			Files: []QuadletFileView{{
				Name:    "Error",
				Content: err.Error(),
			}},
		})
		return
	}

	encoded, _ := system.EncodeQuadletData(results)

	viewFiles := make([]QuadletFileView, len(results))
	for i, f := range results {
		viewFiles[i] = QuadletFileView{
			Name:    f.Name + f.Ext,
			Content: f.Content,
		}
	}

	Tmpl.ExecuteTemplate(w, "quadlet-preview", ConvertResult{
		Files:       viewFiles,
		EncodedData: encoded,
	})
}

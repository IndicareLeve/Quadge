package handlers

import (
	"fmt"
	"net/http"

	"github.com/indicareleve/quadge/system"
)

func StreamLogs(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("service")
	if name == "" {
		http.Error(w, "service name required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	reader, err := system.StreamLogs(name)
	if err != nil {
		fmt.Fprintf(w, "data: Error: %s\n\n", err.Error())
		flusher.Flush()
		return
	}
	defer reader.Close()

	scanner := system.NewLogScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintf(w, "data: %s\n\n", line)
		flusher.Flush()
	}

	if scanner.Err() != nil {
		fmt.Fprintf(w, "data: Error: %s\n\n", scanner.Err().Error())
		flusher.Flush()
	}
}

package handlers

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"

	"github.com/indicareleve/quadge/system"
)

type JournalEntry struct {
	PRIORITY string `json:"PRIORITY"`
	MESSAGE  string `json:"MESSAGE"`
}

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
		fmt.Fprintf(w, "data: <div class=\"log-line log-error\">Error: %s</div>\n\n", err.Error())
		flusher.Flush()
		return
	}
	defer reader.Close()

	scanner := system.NewLogScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		var entry JournalEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			fmt.Fprintf(w, "data: <div class=\"log-line\">%s</div>\n\n", html.EscapeString(line))
			flusher.Flush()
			continue
		}

		level := logLevel(entry.PRIORITY)
		fmt.Fprintf(w, "data: <div class=\"log-line %s\">%s</div>\n\n", level, html.EscapeString(entry.MESSAGE))
		flusher.Flush()
	}

	if scanner.Err() != nil {
		fmt.Fprintf(w, "data: <div class=\"log-line log-error\">Error: %s</div>\n\n", html.EscapeString(scanner.Err().Error()))
		flusher.Flush()
	}
}

func logLevel(priority string) string {
	switch priority {
	case "0", "1", "2", "3":
		return "log-error"
	case "4":
		return "log-warn"
	case "5", "6":
		return "log-info"
	case "7":
		return "log-debug"
	default:
		return ""
	}
}

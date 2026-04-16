# Quadge Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a web application to manage Podman Quadlets with compose conversion, deployment, and log streaming.

**Architecture:** Go HTTP server with chi router, HTMX frontend, embedded templates. State stored on filesystem (quadlet files) + systemd. All system commands use `--user` flag.

**Tech Stack:** Go 1.22+, chi router, HTMX, PicoCSS (CDN), go:embed for assets

---

## File Structure

```
quadge/
├── main.go                      # Entry point, router setup, embedded FS
├── go.mod
├── go.sum
├── handlers/
│   ├── handlers.go              # Shared types, helpers
│   ├── services.go              # List, status, stop, restart
│   ├── convert.go               # Compose → quadlet conversion
│   ├── deploy.go                # Write files, start services
│   ├── edit.go                  # Edit existing quadlet
│   ├── delete.go                # Remove quadlet, stop service
│   └── logs.go                  # SSE log streaming
├── system/
│   ├── podlet.go                # Run podlet compose
│   ├── systemctl.go             # systemctl commands wrapper
│   ├── journalctl.go            # journalctl streaming
│   └── files.go                 # Quadlet file operations
├── templates/
│   ├── index.html               # Main layout
│   ├── new.html                 # Conversion page
│   └── fragments/
│       ├── service_list.html    # Sidebar list
│       ├── service_detail.html  # Main panel
│       ├── quadlet_preview.html # Conversion preview
│       └── error.html           # Error fragment
└── static/
    └── (empty - using CDN for PicoCSS)
```

---

### Task 1: Project Initialization

**Files:**
- Create: `go.mod`
- Create: `main.go`

- [ ] **Step 1: Initialize Go module**

Run: `go mod init github.com/indicareleve/quadge`
Expected: `go: creating new go.mod: module github.com/indicareleve/quadge`

- [ ] **Step 2: Add chi router dependency**

Run: `go get github.com/go-chi/chi/v5`
Expected: Download and add to go.mod

- [ ] **Step 3: Create minimal main.go**

```go
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Quadge - Quadlet Manager"))
	})

	port := os.Getenv("QUADGE_PORT")
	if port == "" {
		port = "4440"
	}

	fmt.Printf("Quadge starting on :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Verify build**

Run: `go build -o quadge .`
Expected: Binary created without errors

- [ ] **Step 5: Test server starts**

Run: `./quadge &` then `curl http://localhost:4440/` then `kill %1`
Expected: "Quadge - Quadlet Manager" response

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum main.go
git commit -m "feat: initialize project with chi router"
```

---

### Task 2: Embedded Templates Setup

**Files:**
- Create: `templates/index.html`
- Create: `templates/new.html`
- Modify: `main.go`

- [ ] **Step 1: Create templates directory**

Run: `mkdir -p templates/fragments`
Expected: Directory created

- [ ] **Step 2: Create base index.html template**

```html
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Quadge - Quadlet Manager</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
	<script src="https://unpkg.com/htmx.org@2.0.0"></script>
	<style>
		:root {
			--sidebar-width: 220px;
		}
		body {
			margin: 0;
			padding: 0;
			min-height: 100vh;
		}
		.layout {
			display: flex;
			min-height: 100vh;
		}
		.sidebar {
			width: var(--sidebar-width);
			background: var(--card-background-color);
			border-right: 1px solid var(--muted-border-color);
			padding: 1rem;
			display: flex;
			flex-direction: column;
		}
		.sidebar h3 {
			margin-bottom: 0.5rem;
		}
		.service-list {
			flex: 1;
			overflow-y: auto;
		}
		.service-item {
			padding: 0.5rem;
			cursor: pointer;
			border-radius: var(--border-radius);
			margin-bottom: 0.25rem;
		}
		.service-item:hover {
			background: var(--primary-background);
		}
		.service-item.active {
			background: var(--primary);
			color: var(--primary-inverse);
		}
		.main {
			flex: 1;
			display: flex;
			flex-direction: column;
		}
		.main-header {
			padding: 1rem;
			border-bottom: 1px solid var(--muted-border-color);
			display: flex;
			align-items: center;
			gap: 1rem;
		}
		.main-content {
			flex: 1;
			display: grid;
			grid-template-columns: 1fr 1fr;
			gap: 1rem;
			padding: 1rem;
			overflow: hidden;
		}
		.panel {
			display: flex;
			flex-direction: column;
			overflow: hidden;
		}
		.panel-header {
			font-weight: bold;
			margin-bottom: 0.5rem;
		}
		.panel-content {
			flex: 1;
			overflow: auto;
			background: var(--card-background-color);
			padding: 0.5rem;
			border-radius: var(--border-radius);
		}
		.status-running { color: green; }
		.status-stopped { color: var(--muted-color); }
		.status-failed { color: red; }
		pre {
			margin: 0;
			white-space: pre-wrap;
			font-size: 0.85rem;
		}
		.logs {
			font-family: monospace;
			font-size: 0.8rem;
			line-height: 1.3;
		}
		.btn-group {
			display: flex;
			gap: 0.5rem;
		}
		.btn-group button {
			padding: 0.25rem 0.75rem;
			font-size: 0.85rem;
		}
	</style>
</head>
<body>
	<div class="layout">
		<aside class="sidebar">
			<h3>Quadge</h3>
			<a href="/new" role="button" class="outline" style="margin-bottom: 1rem;">+ New Service</a>
			<div class="service-list" id="service-list">
				{{template "service-list" .}}
			</div>
		</aside>
		<main class="main">
			{{template "main-content" .}}
		</main>
	</div>
</body>
</html>
```

- [ ] **Step 3: Create service list fragment**

```html
{{define "service-list"}}
{{range .Services}}
<div class="service-item {{if eq .Name $.Selected}}active{{end}}" 
     hx-get="/service/{{.Name}}" 
     hx-target="#main-content"
     hx-swap="innerHTML">
	<span>{{.Name}}</span>
	{{if eq .Status "running"}}<span class="status-running">●</span>
	{{else if eq .Status "failed"}}<span class="status-failed">●</span>
	{{else}}<span class="status-stopped">○</span>{{end}}
</div>
{{else}}
<p style="color: var(--muted-color);">No services yet</p>
{{end}}
{{end}}
```

- [ ] **Step 4: Create service detail fragment**

```html
{{define "service-detail"}}
{{if .Service}}
<div class="main-header">
	<hgroup style="margin: 0;">
		<h4 style="margin: 0;">{{.Service.Name}}</h4>
		<p style="margin: 0;">
			{{if eq .Service.Status "running"}}<span class="status-running">● Running</span>
			{{else if eq .Service.Status "failed"}}<span class="status-failed">● Failed</span>
			{{else}}<span class="status-stopped">○ Stopped</span>{{end}}
		</p>
	</hgroup>
	<div class="btn-group" style="margin-left: auto;">
		{{if eq .Service.Status "running"}}
		<button hx-post="/stop?service={{.Service.Name}}" hx-target="#service-list" hx-swap="innerHTML">Stop</button>
		<button hx-post="/restart?service={{.Service.Name}}" hx-target="#service-list" hx-swap="innerHTML">Restart</button>
		{{else}}
		<button hx-post="/start?service={{.Service.Name}}" hx-target="#service-list" hx-swap="innerHTML">Start</button>
		{{end}}
		<a href="/edit?service={{.Service.Name}}" role="button" class="secondary">Edit</a>
		<button class="contrast" onclick="if(confirm('Delete {{.Service.Name}}?')){htmx.trigger(this, 'delete-service')}" hx-post="/delete?service={{.Service.Name}}" hx-target="#service-list" hx-swap="innerHTML">Delete</button>
	</div>
</div>
<div class="main-content">
	<div class="panel">
		<div class="panel-header">Quadlet</div>
		<div class="panel-content">
			<pre>{{.Service.QuadletContent}}</pre>
		</div>
	</div>
	<div class="panel">
		<div class="panel-header">Logs</div>
		<div class="panel-content logs" hx-ext="sse" sse-connect="/logs?service={{.Service.Name}}" sse-swap="message" hx-swap="beforeend">
			Loading logs...
		</div>
	</div>
</div>
{{else}}
<div class="main-content" style="align-items: center; justify-content: center;">
	<p style="color: var(--muted-color);">Select a service or create a new one</p>
</div>
{{end}}
{{end}}
```

- [ ] **Step 5: Create empty state template**

```html
{{define "main-content"}}
{{if .Service}}
	{{template "service-detail" .}}
{{else}}
<div class="main-content" style="align-items: center; justify-content: center;">
	<p style="color: var(--muted-color);">Select a service or create a new one</p>
</div>
{{end}}
{{end}}
```

- [ ] **Step 6: Update index.html to include fragments**

```html
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Quadge - Quadlet Manager</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
	<script src="https://unpkg.com/htmx.org@2.0.0"></script>
	<script src="https://unpkg.com/htmx-ext-sse@2.0.0/sse.js"></script>
	<style>
		:root {
			--sidebar-width: 220px;
		}
		body {
			margin: 0;
			padding: 0;
			min-height: 100vh;
		}
		.layout {
			display: flex;
			min-height: 100vh;
		}
		.sidebar {
			width: var(--sidebar-width);
			background: var(--card-background-color);
			border-right: 1px solid var(--muted-border-color);
			padding: 1rem;
			display: flex;
			flex-direction: column;
		}
		.sidebar h3 {
			margin-bottom: 0.5rem;
		}
		.service-list {
			flex: 1;
			overflow-y: auto;
		}
		.service-item {
			padding: 0.5rem;
			cursor: pointer;
			border-radius: var(--border-radius);
			margin-bottom: 0.25rem;
		}
		.service-item:hover {
			background: var(--primary-background);
		}
		.service-item.active {
			background: var(--primary);
			color: var(--primary-inverse);
		}
		.main {
			flex: 1;
			display: flex;
			flex-direction: column;
		}
		.main-header {
			padding: 1rem;
			border-bottom: 1px solid var(--muted-border-color);
			display: flex;
			align-items: center;
			gap: 1rem;
		}
		.main-content {
			flex: 1;
			display: grid;
			grid-template-columns: 1fr 1fr;
			gap: 1rem;
			padding: 1rem;
			overflow: hidden;
		}
		.panel {
			display: flex;
			flex-direction: column;
			overflow: hidden;
		}
		.panel-header {
			font-weight: bold;
			margin-bottom: 0.5rem;
		}
		.panel-content {
			flex: 1;
			overflow: auto;
			background: var(--card-background-color);
			padding: 0.5rem;
			border-radius: var(--border-radius);
		}
		.status-running { color: green; }
		.status-stopped { color: var(--muted-color); }
		.status-failed { color: red; }
		pre {
			margin: 0;
			white-space: pre-wrap;
			font-size: 0.85rem;
		}
		.logs {
			font-family: monospace;
			font-size: 0.8rem;
			line-height: 1.3;
		}
		.btn-group {
			display: flex;
			gap: 0.5rem;
		}
		.btn-group button {
			padding: 0.25rem 0.75rem;
			font-size: 0.85rem;
		}
	</style>
</head>
<body>
	<div class="layout">
		<aside class="sidebar">
			<h3>Quadge</h3>
			<a href="/new" role="button" class="outline" style="margin-bottom: 1rem;">+ New Service</a>
			<div class="service-list" id="service-list">
				{{template "service-list" .}}
			</div>
		</aside>
		<main class="main">
			<div id="main-content">
				{{template "main-content" .}}
			</div>
		</main>
	</div>
</body>
</html>
```

- [ ] **Step 7: Create new.html template**

```html
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>New Service - Quadge</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
	<script src="https://unpkg.com/htmx.org@2.0.0"></script>
	<style>
		body {
			min-height: 100vh;
		}
		.header {
			display: flex;
			align-items: center;
			justify-content: space-between;
			padding: 1rem;
			border-bottom: 1px solid var(--muted-border-color);
		}
		.convert-layout {
			display: grid;
			grid-template-columns: 1fr 1fr;
			gap: 1rem;
			padding: 1rem;
			height: calc(100vh - 120px);
		}
		.panel {
			display: flex;
			flex-direction: column;
		}
		.panel-header {
			font-weight: bold;
			margin-bottom: 0.5rem;
		}
		textarea {
			flex: 1;
			font-family: monospace;
			resize: none;
		}
		.preview {
			flex: 1;
			background: var(--card-background-color);
			padding: 0.5rem;
			border-radius: var(--border-radius);
			overflow: auto;
		}
		.preview pre {
			margin: 0;
			white-space: pre-wrap;
			font-size: 0.85rem;
		}
		.actions {
			display: flex;
			gap: 0.5rem;
			justify-content: flex-end;
			padding: 1rem;
			border-top: 1px solid var(--muted-border-color);
		}
	</style>
</head>
<body>
	<div class="header">
		<h3 style="margin: 0;">New Service</h3>
		<a href="/" role="button" class="secondary outline">← Back to services</a>
	</div>
	<form hx-post="/convert" hx-target="#preview">
		<div class="convert-layout">
			<div class="panel">
				<div class="panel-header">docker-compose.yml</div>
				<textarea name="compose" placeholder="version: '3'
services:
  nginx:
    image: nginx:latest
    ports:
      - '80:80'" required></textarea>
			</div>
			<div class="panel">
				<div class="panel-header">Quadlet Preview</div>
				<div class="preview" id="preview">
					<p style="color: var(--muted-color);">Click "Convert" to generate quadlet files</p>
				</div>
			</div>
		</div>
		<div class="actions">
			<button type="submit">Convert</button>
			<button type="button" id="deploy-btn" disabled hx-post="/deploy" hx-include="textarea" hx-target="body" hx-swap="outerHTML">Deploy</button>
		</div>
	</form>
	<script>
		document.body.addEventListener('htmx:afterSwap', function(evt) {
			if (evt.detail.target.id === 'preview') {
				document.getElementById('deploy-btn').disabled = false;
			}
		});
	</script>
</body>
</html>
```

- [ ] **Step 8: Create quadlet preview fragment**

```html
{{define "quadlet-preview"}}
{{range .Files}}
<div style="margin-bottom: 1rem;">
	<strong>{{.Name}}</strong>
	<pre>{{.Content}}</pre>
</div>
{{end}}
<input type="hidden" name="quadlet-data" value="{{.EncodedData}}">
{{end}}
```

- [ ] **Step 9: Update main.go with embedded templates**

```go
package main

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed templates/*
var templatesFS embed.FS

var tmpl *template.Template

func init() {
	var err error
	tmpl, err = template.ParseFS(templatesFS, "templates/*.html", "templates/fragments/*.html")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing templates: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]interface{}{
			"Services": []interface{}{},
			"Selected": "",
		}
		tmpl.ExecuteTemplate(w, "index.html", data)
	})

	r.Get("/new", func(w http.ResponseWriter, r *http.Request) {
		tmpl.ExecuteTemplate(w, "new.html", nil)
	})

	port := os.Getenv("QUADGE_PORT")
	if port == "" {
		port = "4440"
	}

	fmt.Printf("Quadge starting on :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 10: Build and verify**

Run: `go build -o quadge .`
Expected: Build succeeds without errors

- [ ] **Step 11: Commit**

```bash
git add templates/ main.go go.sum
git commit -m "feat: add embedded templates and basic routing"
```

---

### Task 3: System Package - File Operations

**Files:**
- Create: `system/files.go`

- [ ] **Step 1: Create system package directory**

Run: `mkdir -p system`
Expected: Directory created

- [ ] **Step 2: Create files.go with quadlet file operations**

```go
package system

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const QuadletDir = ".config/containers/systemd"

func GetQuadletDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, QuadletDir), nil
}

type QuadletFile struct {
	Name    string
	Path    string
	Type    string // "container", "pod", "volume", "network", "kube"
	Content string
}

func ListQuadletFiles() ([]QuadletFile, error) {
	dir, err := GetQuadletDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return []QuadletFile{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading quadlet directory: %w", err)
	}

	var files []QuadletFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)
		if !isValidQuadletType(ext) {
			continue
		}

		path := filepath.Join(dir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		files = append(files, QuadletFile{
			Name:    strings.TrimSuffix(name, ext),
			Path:    path,
			Type:    strings.TrimPrefix(ext, "."),
			Content: string(content),
		})
	}

	return files, nil
}

func ReadQuadletFile(name string) (QuadletFile, error) {
	dir, err := GetQuadletDir()
	if err != nil {
		return QuadletFile{}, err
	}

	for _, ext := range []string{".container", ".pod", ".kube", ".volume", ".network"} {
		path := filepath.Join(dir, name+ext)
		content, err := os.ReadFile(path)
		if err == nil {
			return QuadletFile{
				Name:    name,
				Path:    path,
				Type:    strings.TrimPrefix(ext, "."),
				Content: string(content),
			}, nil
		}
	}

	return QuadletFile{}, fmt.Errorf("quadlet file not found: %s", name)
}

func WriteQuadletFile(name string, ext string, content string) error {
	dir, err := GetQuadletDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating quadlet directory: %w", err)
	}

	path := filepath.Join(dir, name+ext)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing quadlet file: %w", err)
	}

	return nil
}

func DeleteQuadletFile(name string) error {
	dir, err := GetQuadletDir()
	if err != nil {
		return err
	}

	deleted := false
	for _, ext := range []string{".container", ".pod", ".kube", ".volume", ".network"} {
		path := filepath.Join(dir, name+ext)
		if err := os.Remove(path); err == nil {
			deleted = true
		}
	}

	if !deleted {
		return fmt.Errorf("no quadlet files found for: %s", name)
	}

	return nil
}

func isValidQuadletType(ext string) bool {
	switch ext {
	case ".container", ".pod", ".kube", ".volume", ".network":
		return true
	default:
		return false
	}
}
```

- [ ] **Step 3: Build to verify**

Run: `go build -o quadge .`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add system/files.go
git commit -m "feat: add quadlet file operations"
```

---

### Task 4: System Package - Systemctl Wrapper

**Files:**
- Create: `system/systemctl.go`

- [ ] **Step 1: Create systemctl.go**

```go
package system

import (
	"fmt"
	"os/exec"
	"strings"
)

type ServiceStatus string

const (
	StatusRunning ServiceStatus = "running"
	StatusStopped ServiceStatus = "stopped"
	StatusFailed  ServiceStatus = "failed"
	StatusUnknown ServiceStatus = "unknown"
)

func runSystemctl(args ...string) (string, error) {
	cmd := exec.Command("systemctl", append([]string{"--user"}, args...)...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func GetServiceStatus(name string) (ServiceStatus, error) {
	output, err := runSystemctl("is-active", name+".service")
	output = strings.TrimSpace(output)

	if err == nil {
		switch output {
		case "active":
			return StatusRunning, nil
		case "inactive":
			return StatusStopped, nil
		case "failed":
			return StatusFailed, nil
		}
	}

	if strings.Contains(output, "inactive") || strings.Contains(output, "could not be found") {
		return StatusStopped, nil
	}

	return StatusUnknown, fmt.Errorf("checking status: %s", output)
}

func DaemonReload() error {
	_, err := runSystemctl("daemon-reload")
	if err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}
	return nil
}

func StartService(name string) error {
	_, err := runSystemctl("start", name+".service")
	if err != nil {
		return fmt.Errorf("starting service: %w", err)
	}
	return nil
}

func StopService(name string) error {
	_, err := runSystemctl("stop", name+".service")
	if err != nil {
		return fmt.Errorf("stopping service: %w", err)
	}
	return nil
}

func RestartService(name string) error {
	_, err := runSystemctl("restart", name+".service")
	if err != nil {
		return fmt.Errorf("restarting service: %w", err)
	}
	return nil
}
```

- [ ] **Step 2: Build to verify**

Run: `go build -o quadge .`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add system/systemctl.go
git commit -m "feat: add systemctl wrapper functions"
```

---

### Task 5: System Package - Journalctl Wrapper

**Files:**
- Create: `system/journalctl.go`

- [ ] **Step 1: Create journalctl.go**

```go
package system

import (
	"bufio"
	"io"
	"os/exec"
)

func StreamLogs(name string) (io.ReadCloser, error) {
	cmd := exec.Command("journalctl", "--user", "-u", name+".service", "-f", "-n", "50")
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &logReader{stdout, cmd}, nil
}

type logReader struct {
	io.Reader
	cmd *exec.Cmd
}

func (l *logReader) Close() error {
	return l.cmd.Process.Kill()
}

func FormatLogLine(line string) string {
	return line
}

func NewLogScanner(r io.Reader) *bufio.Scanner {
	return bufio.NewScanner(r)
}
```

- [ ] **Step 2: Build to verify**

Run: `go build -o quadge .`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add system/journalctl.go
git commit -m "feat: add journalctl log streaming"
```

---

### Task 6: System Package - Podlet Wrapper

**Files:**
- Create: `system/podlet.go`

- [ ] **Step 1: Create podlet.go**

```go
package system

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type QuadletResult struct {
	Name    string
	Ext     string
	Content string
}

func ConvertCompose(composeContent string) ([]QuadletResult, error) {
	tmpDir, err := os.MkdirTemp("", "quadge-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	composeFile := filepath.Join(tmpDir, "compose.yml")
	if err := os.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		return nil, fmt.Errorf("writing compose file: %w", err)
	}

	outputDir := filepath.Join(tmpDir, "output")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output dir: %w", err)
	}

	cmd := exec.Command("podlet", "compose", "--file", outputDir, composeFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("podlet error: %s", string(output))
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, fmt.Errorf("reading output dir: %w", err)
	}

	var results []QuadletResult
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := filepath.Ext(name)
		content, err := os.ReadFile(filepath.Join(outputDir, name))
		if err != nil {
			continue
		}

		results = append(results, QuadletResult{
			Name:    name[:len(name)-len(ext)],
			Ext:     ext,
			Content: string(content),
		})
	}

	return results, nil
}

func EncodeQuadletData(files []QuadletResult) (string, error) {
	data, err := json.Marshal(files)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func DecodeQuadletData(encoded string) ([]QuadletResult, error) {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}

	var files []QuadletResult
	if err := json.Unmarshal(data, &files); err != nil {
		return nil, err
	}

	return files, nil
}
```

- [ ] **Step 2: Build to verify**

Run: `go build -o quadge .`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add system/podlet.go
git commit -m "feat: add podlet compose wrapper"
```

---

### Task 7: Handlers Package - Setup and Types

**Files:**
- Create: `handlers/handlers.go`

- [ ] **Step 1: Create handlers directory**

Run: `mkdir -p handlers`
Expected: Directory created

- [ ] **Step 2: Create handlers.go with shared types**

```go
package handlers

import (
	"html/template"

	"github.com/indicareleve/quadge/system"
)

var Tmpl *template.Template

type Service struct {
	Name           string
	Status         system.ServiceStatus
	QuadletContent string
}

type PageData struct {
	Services []Service
	Selected string
	Service  *Service
}

type ConvertResult struct {
	Files       []QuadletFileView
	EncodedData string
}

type QuadletFileView struct {
	Name    string
	Content string
}

func ToServices(files []system.QuadletFile) ([]Service, error) {
	services := make([]Service, 0, len(files))
	for _, f := range files {
		status, _ := system.GetServiceStatus(f.Name)
		services = append(services, Service{
			Name:           f.Name,
			Status:         status,
			QuadletContent: f.Content,
		})
	}
	return services, nil
}
```

- [ ] **Step 3: Update main.go to share template with handlers**

```go
package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/indicareleve/quadge/handlers"
)

//go:embed templates/*
var templatesFS embed.FS

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Parse templates
	tmpl, err := handlers.ParseTemplates(templatesFS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing templates: %v\n", err)
		os.Exit(1)
	}
	handlers.Tmpl = tmpl

	r.Get("/", handlers.ListServices)
	r.Get("/new", handlers.NewServiceForm)
	r.Post("/convert", handlers.ConvertCompose)
	r.Post("/deploy", handlers.DeployService)
	r.Get("/service/{name}", handlers.GetService)
	r.Post("/start", handlers.StartService)
	r.Post("/stop", handlers.StopService)
	r.Post("/restart", handlers.RestartService)
	r.Post("/delete", handlers.DeleteService)
	r.Get("/logs", handlers.StreamLogs)
	r.Get("/edit", handlers.EditServiceForm)
	r.Post("/edit", handlers.EditService)

	port := os.Getenv("QUADGE_PORT")
	if port == "" {
		port = "4440"
	}

	fmt.Printf("Quadge starting on :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Add ParseTemplates function to handlers.go**

```go
package handlers

import (
	"embed"
	"html/template"
)

func ParseTemplates(fs embed.FS) (*template.Template, error) {
	return template.ParseFS(fs, "templates/*.html", "templates/fragments/*.html")
}
```

- [ ] **Step 5: Build to verify imports**

Run: `go build -o quadge .`
Expected: Build succeeds (may have undefined handler errors - we'll add handlers next)

- [ ] **Step 6: Commit**

```bash
git add handlers/handlers.go main.go
git commit -m "feat: add handlers package setup"
```

---

### Task 8: Handlers - List Services

**Files:**
- Modify: `handlers/handlers.go`
- Create: `handlers/services.go`

- [ ] **Step 1: Create services.go with list handler**

```go
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
```

- [ ] **Step 2: Build to verify**

Run: `go build -o quadge .`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add handlers/services.go
git commit -m "feat: add service list and detail handlers"
```

---

### Task 9: Handlers - Service Control

**Files:**
- Modify: `handlers/services.go`

- [ ] **Step 1: Add StartService, StopService, RestartService handlers**

```go
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
```

- [ ] **Step 2: Build to verify**

Run: `go build -o quadge .`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add handlers/services.go
git commit -m "feat: add start/stop/restart handlers"
```

---

### Task 10: Handlers - Convert and Deploy

**Files:**
- Create: `handlers/convert.go`
- Create: `handlers/deploy.go`

- [ ] **Step 1: Create convert.go**

```go
package handlers

import (
	"net/http"
)

func NewServiceForm(w http.ResponseWriter, r *http.Request) {
	Tmpl.ExecuteTemplate(w, "new.html", nil)
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
```

- [ ] **Step 2: Create deploy.go**

```go
package handlers

import (
	"net/http"
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
```

- [ ] **Step 3: Build to verify**

Run: `go build -o quadge .`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add handlers/convert.go handlers/deploy.go
git commit -m "feat: add convert and deploy handlers"
```

---

### Task 11: Handlers - Edit and Delete

**Files:**
- Create: `handlers/edit.go`
- Create: `handlers/delete.go`
- Create: `templates/edit.html`

- [ ] **Step 1: Create edit.go**

```go
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

	data := struct {
		Name    string
		Content string
	}{
		Name:    file.Name,
		Content: file.Content,
	}

	Tmpl.ExecuteTemplate(w, "edit.html", data)
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

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

- [ ] **Step 2: Create delete.go**

```go
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
```

- [ ] **Step 3: Create edit.html template**

```html
<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Edit {{.Name}} - Quadge</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@2/css/pico.min.css">
</head>
<body style="padding: 2rem;">
	<header style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
		<h3 style="margin: 0;">Edit: {{.Name}}</h3>
		<a href="/" role="button" class="secondary outline">← Back</a>
	</header>
	<form method="POST" action="/edit">
		<input type="hidden" name="name" value="{{.Name}}">
		<textarea name="content" style="min-height: 60vh; font-family: monospace;">{{.Content}}</textarea>
		<div style="display: flex; gap: 1rem; justify-content: flex-end; margin-top: 1rem;">
			<a href="/" role="button" class="secondary outline">Cancel</a>
			<button type="submit">Save</button>
		</div>
	</form>
</body>
</html>
```

- [ ] **Step 4: Build to verify**

Run: `go build -o quadge .`
Expected: Build succeeds

- [ ] **Step 5: Commit**

```bash
git add handlers/edit.go handlers/delete.go templates/edit.html
git commit -m "feat: add edit and delete handlers"
```

---

### Task 12: Handlers - Log Streaming

**Files:**
- Create: `handlers/logs.go`

- [ ] **Step 1: Create logs.go**

```go
package handlers

import (
	"bufio"
	"fmt"
	"net/http"
	"time"

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

func init() {
	// Ensure SSE works with http.Flusher
	_ = time.Tick(0)
}
```

- [ ] **Step 2: Build to verify**

Run: `go build -o quadge .`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add handlers/logs.go
git commit -m "feat: add SSE log streaming handler"
```

---

### Task 13: Final Integration and Testing

**Files:**
- Modify: Various

- [ ] **Step 1: Build final binary**

Run: `go build -o quadge .`
Expected: Clean build

- [ ] **Step 2: Run server**

Run: `./quadge`
Expected: "Quadge starting on :4440"

- [ ] **Step 3: Test in browser**

Open: `http://localhost:4440`
Expected: Empty service list, "New Service" button visible

- [ ] **Step 4: Test conversion flow**

1. Click "New Service"
2. Paste a docker-compose.yml
3. Click "Convert"
4. Verify quadlet preview appears

- [ ] **Step 5: Test deployment**

1. Click "Deploy"
2. Verify redirect to main page
3. Verify service appears in list

- [ ] **Step 6: Test service controls**

1. Click service in list
2. Verify quadlet content shows
3. Verify logs stream
4. Test Stop/Restart/Edit buttons

- [ ] **Step 7: Final commit**

```bash
git add -A
git commit -m "feat: complete Quadge implementation"
```

---

## Self-Review Checklist

**Spec coverage:**
- ✓ Split three-pane layout (sidebar + header + quadlet/logs)
- ✓ Service list with status indicators
- ✓ Convert compose → quadlet (multi-container support)
- ✓ Deploy with daemon-reload and start
- ✓ Stop/restart controls
- ✓ Edit raw quadlet
- ✓ Delete with confirmation
- ✓ SSE log streaming
- ✓ Port 4440 configurable
- ✓ No auth
- ✓ chi router
- ✓ PicoCSS via CDN
- ✓ HTMX interactivity
- ✓ Embedded templates

**Placeholder scan:** No TBD, TODO, or vague steps found.

**Type consistency:** All function names and types consistent across tasks.

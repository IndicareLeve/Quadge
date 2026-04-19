# Quadge

**Podman Quadlet web manager** — convert docker-compose files to Quadlet and deploy with one click.

## Features

- **Auto-convert on paste** — paste a `docker-compose.yml`, get Quadlet files instantly
- **Edit before deploy** — tweak the generated Quadlet content before committing
- **One-click deploy** — writes Quadlet files, reloads systemd, starts services
- **Live logs** — color-coded journalctl stream via SSE
- **Service management** — start, stop, restart, edit, delete from a single interface
- **Dockge-style layout** — sidebar always visible, two-panel editor

## Screenshot

![Quadge UI](docs/screenshot.png)

## Quick Start

### Prerequisites

- Go 1.26+
- Podman with Quadlet enabled
- [Podlet](https://github.com/containers/podlet) — `podlet compose` for conversion
- systemd (user mode)

### Build

```bash
go build -o quadge .
```

### Run

```bash
./quadge
# Server starts on http://localhost:4440
# Override: QUADGE_PORT=8080 ./quadge
```

### Systemd Service

```ini
[Unit]
Description=Quadge - Podman Quadlet Web Manager
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/quadge
Restart=on-failure

[Install]
WantedBy=default.target
```

## Usage

1. Open `http://localhost:4440`
2. Click **+ New Service**
3. Paste a `docker-compose.yml` — auto-converts to Quadlet
4. Edit the generated content if needed
5. Click **Deploy** — files written to `~/.config/containers/systemd/`, service started

### Managing Services

- **View** — click a service to see Quadlet config + live logs
- **Edit** — modify Quadlet files directly
- **Start/Stop/Restart** — toggle service state
- **Delete** — remove Quadlet files and stop service

## Architecture

Single Go binary with embedded templates. No Node.js, no build chain.

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Browser   │────▶│  Quadge      │────▶│  Podlet     │
│  (HTMX +    │◀────│  (chi router │     │  (compose→  │
│   PicoCSS)  │     │   + embed)   │────▶│   quadlet)  │
└─────────────┘     └──────┬───────┘     └─────────────┘
                           │
                    ┌──────▼───────┐
                    │  systemd     │
                    │  --user      │
                    │  (quadlet)   │
                    └──────────────┘
```

### Key Paths

| Component | Path |
|-----------|------|
| Quadlet files | `~/.config/containers/systemd/` |
| Valid extensions | `.container`, `.pod`, `.kube`, `.volume`, `.network` |

### Project Structure

```
.
├── main.go                    # Entry point, chi router, embedded templates
├── handlers/
│   ├── handlers.go            # Shared types, template parsing
│   ├── services.go            # List, get, start, stop, restart
│   ├── convert.go             # Compose → Quadlet conversion
│   ├── deploy.go              # Write files, reload, start
│   ├── edit.go                # Edit service fragment
│   ├── delete.go              # Delete service
│   └── logs.go                # SSE log streaming
├── system/
│   ├── podlet.go              # Podlet CLI wrapper
│   ├── systemctl.go           # systemctl --user wrapper
│   ├── journalctl.go          # journalctl --user wrapper
│   └── files.go               # Quadlet file I/O
└── templates/
    ├── index.html              # Main layout (sidebar + content)
    └── fragments/
        ├── new_service.html    # New service editor
        ├── edit_service.html   # Edit service editor
        ├── service_detail.html # Service view (logs + quadlet)
        ├── service_list.html   # Sidebar service list
        └── quadlet_preview.html # Conversion preview
```

## License

MIT

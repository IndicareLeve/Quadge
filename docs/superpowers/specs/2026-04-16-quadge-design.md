# Quadge - Design Specification

## Overview

Quadge is a lightweight web application for managing Podman Quadlets on Arch Linux. It converts Docker Compose files to systemd Quadlet units, deploys them, and provides real-time log streaming.

## Tech Stack

- **Backend:** Go 1.22+
- **Router:** chi (lightweight, idiomatic)
- **Frontend:** HTMX (no JS build step)
- **Styling:** PicoCSS via CDN
- **Assets:** `//go:embed` for templates compiled into binary
- **Port:** 4440 (configurable via CLI flag)

## Architecture

```
quadge/
├── main.go                 # Entry point, embedded FS, chi router setup
├── handlers/
│   ├── services.go         # GET / - list services, show detail
│   ├── convert.go          # POST /convert - podlet compose conversion
│   ├── deploy.go           # POST /deploy - write files, systemctl start
│   ├── edit.go             # POST /edit - update quadlet file
│   ├── logs.go             # GET /logs?service=<name> - SSE streaming
│   └── delete.go           # POST /delete - remove quadlet, stop service
├── system/
│   ├── podlet.go           # Run podlet compose command
│   ├── systemctl.go        # daemon-reload, start, stop, restart, is-active
│   ├── journalctl.go       # Stream logs via journalctl
│   └── files.go            # Read/write quadlet files
├── templates/
│   ├── index.html          # Main layout with sidebar
│   ├── new.html            # Conversion flow page
│   └── fragments/
│       ├── service_list.html    # Sidebar service list
│       ├── service_detail.html  # Main panel content
│       ├── quadlet_preview.html # Conversion preview
│       └── error.html           # Error display
└── go.mod
```

## User Interface

### Main Page (GET /)

**Layout:**
- **Sidebar (left):** List of services + "New Service" button
- **Main (right):** 
  - Header: service name, status badge (Running/Stopped), action buttons (Stop/Restart/Edit/Delete)
  - Content split 50/50: Quadlet file (read-only) | Live logs (SSE stream)

**Sidebar Behavior:**
- Click service → swap main content via HTMX
- Click "New Service" → navigate to /new

### New Service Page (GET /new, POST /convert, POST /deploy)

**Layout: Single page with live preview**
- Header: "New Service" + "← Back to services" link
- Left: Textarea for docker-compose.yml input
- Right: Quadlet preview (populated after Convert)
- Footer: "Convert" + "Deploy" buttons

**Flow:**
1. User pastes compose YAML
2. Click "Convert" → HTMX POST /convert → preview updates
3. Click "Deploy" → POST /deploy → creates files, starts service, redirects to main

**Multi-container handling:**
- `podlet compose` creates multiple `.container` files
- Preview shows all generated files
- Deploy writes all files, starts all services

## API Routes

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/` | services.List | Render main page with service list |
| GET | `/new` | convert.Form | Render conversion page |
| POST | `/convert` | convert.Handle | Accept compose YAML, return quadlet preview fragment |
| POST | `/deploy` | deploy.Handle | Write quadlet files, daemon-reload, start services |
| GET | `/logs?service=<name>` | logs.Stream | SSE endpoint for journalctl streaming |
| GET | `/edit?service=<name>` | edit.Form | Render edit form (textarea with quadlet content) |
| POST | `/edit` | edit.Handle | Save edited quadlet, daemon-reload, restart if running |
| POST | `/stop` | services.Stop | systemctl --user stop <name> |
| POST | `/restart` | services.Restart | systemctl --user restart <name> |
| POST | `/delete` | services.Delete | Stop, remove files, daemon-reload, return updated list |

## System Commands

All commands use `--user` flag (no root required).

### Conversion
```bash
podlet compose --file /tmp/quadge-<random>/ <compose-file>
```
Creates directory with `.container`, `.volume`, `.network` files.

### Service Management
```bash
# Get status
systemctl --user is-active <name>.service

# Reload after file changes
systemctl --user daemon-reload

# Control
systemctl --user start <name>.service
systemctl --user stop <name>.service
systemctl --user restart <name>.service
```

### Log Streaming
```bash
journalctl --user -u <name>.service -f
```
Pipe stdout to SSE events.

## File Storage

**Quadlet directory:** `$HOME/.config/containers/systemd/`

**File types:**
- `.container` - Container definitions
- `.pod` - Pod groupings
- `.volume` - Volume definitions
- `.network` - Network definitions
- `.kube` - Kubernetes YAML

**Service discovery:** Scan directory for quadlet files, map to services.

## State Management

**No database.** Stateless HTTP server.

- Service list: Scan quadlet directory on each request (or cache with TTL)
- Service status: Query systemctl each time
- Conversion preview: Stored in hidden form field, passed to deploy
- No session state

## Error Handling

**Command failures:**
- Capture stderr from podlet/systemctl
- Display error message in UI (toast/alert)
- Log to stdout (journald captures)

**File operations:**
- Check directory exists before write
- Validate temp directory creation
- Handle permission errors gracefully

**User input:**
- Invalid compose YAML: show podlet error output
- Missing service name: client-side validation
- File exists on deploy: overwrite option or error

**SSE:**
- Handle connection drop
- Client auto-reconnects (HTMX)
- Service doesn't exist: show error in stream, close connection

## Security

- **No authentication** - intended for trusted homelab network
- Runs as `--user` systemd service - limited to user permissions
- No secrets in code or config
- Behind reverse proxy in production (user responsibility)

## Confirmations

- **Delete:** JavaScript `confirm()` dialog before POST /delete
- **Stop/Restart:** No confirmation, immediate action

## Testing

**Manual testing only.** No unit tests at this stage.

**Verification checklist:**
- `go build` compiles successfully
- Templates parse without error
- UI loads at port 4440
- Convert compose YAML → see preview
- Deploy → service appears in list, shows Running
- Logs stream in real-time
- Stop/Restart/Edit/Delete work as expected

## Deployment

Quadge runs as a systemd user service:

```ini
# ~/.config/systemd/user/quadge.service
[Unit]
Description=Quadge - Quadlet Manager
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/quadge
Restart=on-failure

[Install]
WantedBy=default.target
```

Enable: `systemctl --user enable --now quadge`

## Future Considerations

- Service grouping by compose project
- Bulk operations (stop all, restart all)
- Resource usage display (cpu/mem per container)
- Quadlet creation from scratch (no compose input)
- Import existing containers as quadlets

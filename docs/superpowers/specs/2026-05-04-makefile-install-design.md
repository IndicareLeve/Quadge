# Design: Makefile + Systemd Install Script

## Overview

Add a Makefile with build/install/uninstall targets and a systemd user service file for quadge. Single command deployment: `make install`.

## Architecture

### Binary Location
- Path: `~/.local/bin/quadge`
- User-scoped, no sudo required
- Typically in PATH on Arch Linux

### Service File
- Location: `~/.config/systemd/user/quadge.service`
- Scope: `--user` (matches existing quadge patterns)
- Uses `%h` for home dir expansion (portable)

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make build` | Release build: `go build -ldflags="-s -w" -o quadge .` |
| `make install` | Build + copy binary + install service + enable + start |
| `make uninstall` | Stop + disable service + remove binary + remove service file |
| `make dev` | Dev server: `go run .` (existing workflow) |

## Install Flow

1. `go build -ldflags="-s -w" -o quadge .`
2. `mkdir -p ~/.local/bin`
3. `cp quadge ~/.local/bin/quadge`
4. `mkdir -p ~/.config/systemd/user/`
5. `cp quadge.service ~/.config/systemd/user/quadge.service`
6. `systemctl --user daemon-reload`
7. `systemctl --user enable --now quadge`

## Uninstall Flow

1. `systemctl --user disable --now quadge`
2. `rm ~/.local/bin/quadge`
3. `rm ~/.config/systemd/user/quadge.service`
4. `systemctl --user daemon-reload`

## Service File Content

```ini
[Unit]
Description=Quadge - Podman Quadlet Web Manager
After=network.target

[Service]
Type=simple
ExecStart=%h/.local/bin/quadge
Restart=on-failure
Environment=QUADGE_PORT=4440

[Install]
WantedBy=default.target
```

## Files Created

- `Makefile` — build/install targets
- `quadge.service` — systemd user service template

## Constraints

- No sudo required (user-scope only)
- Portable across users (%h expansion)
- Preserves existing `go run .` dev workflow via `make dev`

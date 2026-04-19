Terse like caveman. Technical substance exact. Only fluff die.
Drop: articles, filler (just/really/basically), pleasantries, hedging.
Fragments OK. Short synonyms. Code unchanged.
Pattern: [thing] [action] [reason]. [next step].
ACTIVE EVERY RESPONSE. No revert after many turns. No filler drift.
Code/commits/PRs: normal.
Off: "stop caveman" / "normal mode".

## Build & Run

- `go build -o quadge .` — build binary
- `go run .` — run dev server on :4440 (override with `QUADGE_PORT`)
- No Makefile, no Dockerfile, no CI, no test files yet

## Architecture

Podman Quadlet web manager. Single Go binary with embedded templates.

- `main.go` — entry point, chi router, `//go:embed templates/*`
- `handlers/` — HTTP handlers, global `Tmpl *template.Template` set at startup
- `system/` — os/exec wrappers for host commands
- `templates/` — Go HTML templates + HTMX fragments
- `templates/fragments/new_service.html` — New service editor (HTMX fragment, two-panel layout)
- Sidebar always visible; "+ New Service" loads editor into `#main-content` via HTMX
- Auto-convert on paste: textarea paste event → POST `/convert` → replace content on success
- `templates/new.html` deleted (was full-page, replaced by fragment)

## Critical Constraints

- Runs bare-metal on Arch Linux as systemd service. Heavily uses `os/exec`.
- All `systemctl` and `journalctl` calls must include `--user` flag (enforced by `system/runSystemctl`).
- Quadlet files live at `~/.config/containers/systemd/`
- Valid quadlet extensions: `.container`, `.pod`, `.kube`, `.volume`, `.network`
- External CLI dependency: `podlet` (compose-to-quadlet conversion)
- Frontend: HTMX + PicoCSS. No Node.js, no build chain. Templates rendered server-side.
- Quadlet data passed between convert→deploy via base64-encoded JSON in hidden form field (`quadlet-data`)

## Go

- Module: `github.com/indicareleve/quadge`
- Go 1.26.2, chi v5 router
- No test files. If adding tests: `go test ./...`
- Lint: `go vet ./...`

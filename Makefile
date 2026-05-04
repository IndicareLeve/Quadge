.PHONY: build install uninstall dev

BINARY_NAME=quadge
INSTALL_DIR=$(HOME)/.local/bin
SERVICE_FILE=quadge.service
SERVICE_DIR=$(HOME)/.config/systemd/user

build:
	go build -ldflags="-s -w" -o $(BINARY_NAME) .

dev:
	go run .

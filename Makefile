.PHONY: build install uninstall dev

BINARY_NAME=quadge
INSTALL_DIR=$(HOME)/.local/bin
SERVICE_FILE=quadge.service
SERVICE_DIR=$(HOME)/.config/systemd/user

build:
	go build -ldflags="-s -w" -o $(BINARY_NAME) .

dev:
	go run .

install: build
	systemctl --user stop $(BINARY_NAME)
	mkdir -p $(INSTALL_DIR)
	cp $(BINARY_NAME) $(INSTALL_DIR)/
	mkdir -p $(SERVICE_DIR)
	cp $(SERVICE_FILE) $(SERVICE_DIR)/
	systemctl --user daemon-reload
	systemctl --user enable --now $(BINARY_NAME)
	@echo "Quadge installed and started. Check status: systemctl --user status $(BINARY_NAME)"

uninstall:
	systemctl --user disable --now $(BINARY_NAME) || true
	rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	rm -f $(SERVICE_DIR)/$(SERVICE_FILE)
	systemctl --user daemon-reload
	@echo "Quadge uninstalled."

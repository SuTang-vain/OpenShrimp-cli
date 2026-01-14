.PHONY: build clean test install ui-deps ui-dev ui-build

# Go build
build:
	go build -o ai-mgr .

# Clean build artifacts
clean:
	rm -f ai-mgr
	rm -rf ui/dist

# Run tests
test:
	go test ./...

# Install to /usr/local/bin
install: build
	sudo mv ai-mgr /usr/local/bin/

# Install dependencies
deps:
	go mod tidy
	cd ui && npm install

# Run with custom config
run: build
	./ai-mgr --config ~/.ai-manager/config.yaml

# Create default config
init-config:
	go run main.go init

# Build for all platforms
release:
	GOOS=darwin GOARCH=amd64 go build -o ai-mgr-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o ai-mgr-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -o ai-mgr-linux-amd64 .

# Daemon commands
daemon-start: build
	./ai-mgr daemon

daemon-stop:
	@pkill -f "ai-mgr daemon" || true

# UI commands
ui-deps:
	cd ui && npm install

ui-dev:
	cd ui && npm run dev

ui-build:
	cd ui && npm run build

# Tauri commands (requires Rust)
tauri-install:
	curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
	cd tauri && cargo install tauri-cli

tauri-dev: ui-deps
	cd tauri && cargo tauri dev

tauri-build:
	cd ui && npm run build
	cd tauri && cargo tauri build

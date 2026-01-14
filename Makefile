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

# Shell completion
completion:
	@echo "Installing bash completion..."
	@mkdir -p ~/.bash_completion.d
	@cp completion/ai-mgr.bash ~/.bash_completion.d/
	@echo "Bash completion installed. Run: source ~/.bash_completion.d/ai-mgr.bash"

	@echo ""
	@echo "Installing zsh completion..."
	@cp completion/ai-mgr.zsh ~/.zsh/_ai-mgr
	@echo "Zsh completion installed. Restart zsh or run: autoload _ai-mgr"

# Docker
docker-build:
	docker build -t ai-manager:latest .

docker-run:
	docker run -p 19999:19999 -v ~/.ai-manager:/home/app/.ai-manager ai-manager:latest

# Lint
lint:
	golangci-lint run ./...
	cd ui && npx eslint src --ext .vue,.ts,.tsx

# Full release build
release-all: clean build ui-build
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o ai-mgr-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o ai-mgr-darwin-arm64 .
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ai-mgr-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o ai-mgr-linux-arm64 .

	@echo "Release binaries built:"
	@ls -la ai-mgr-*
	@echo ""
	@echo "UI built in ui/dist/"

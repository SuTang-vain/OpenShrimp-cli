.PHONY: build clean test install

# Build the binary
build:
	go build -o ai-mgr .

# Clean build artifacts
clean:
	rm -f ai-mgr

# Run tests
test:
	go test ./...

# Install to /usr/local/bin
install: build
	sudo mv ai-mgr /usr/local/bin/

# Install dependencies
deps:
	go mod tidy

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

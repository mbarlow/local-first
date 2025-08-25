.PHONY: all dev build clean test serve install wasm server docker-build docker-dev

# Default target
all: install build

# Install dependencies
install:
	@echo "Installing dependencies..."
	@go mod download
	@npm install

# Development mode - run both Vite and Go server
dev: wasm
	@echo "Starting Vite development server..."
	@echo "Open http://localhost:5173 in your browser"
	@npm run dev

# Build WASM binary
wasm:
	@echo "Building WASM binary..."
	@GOOS=js GOARCH=wasm go build -o web/main.wasm cmd/wasm/main.go
	@if [ ! -f web/wasm_exec.js ] || [ $$(stat -c%s web/wasm_exec.js 2>/dev/null || stat -f%z web/wasm_exec.js 2>/dev/null || echo 0) -lt 1000 ]; then \
		echo "Downloading wasm_exec.js..."; \
		curl -s "https://raw.githubusercontent.com/golang/go/release-branch.go1.21/misc/wasm/wasm_exec.js" -o web/wasm_exec.js || \
		cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" web/ 2>/dev/null; \
	fi
	@echo "WASM build complete: web/main.wasm"

# Production WASM build (optimized)
wasm-prod:
	@echo "Building optimized WASM binary..."
	@GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o web/main.wasm cmd/wasm/main.go
	@if [ ! -f web/wasm_exec.js ]; then \
		cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" web/ 2>/dev/null || \
		curl -s "https://raw.githubusercontent.com/golang/go/master/misc/wasm/wasm_exec.js" -o web/wasm_exec.js; \
	fi
	@echo "Production WASM build complete"

# Build Go server (development mode - no embedded files)
server:
	@echo "Building Go server..."
	@mkdir -p bin
	@go build -o bin/server ./cmd/server/
	@echo "Server built: bin/server"

# Build CLI
cli:
	@echo "Building CLI..."
	@mkdir -p bin
	@go build -o bin/local ./cmd/cli/
	@echo "CLI built: bin/local"

# Build server with embedded files (production)
server-embed: wasm-prod
	@echo "Preparing embedded files..."
	@mkdir -p cmd/server/web
	@cp -r web/* cmd/server/web/ 2>/dev/null || true
	@echo "Building Go server with embedded files..."
	@go build -tags embed -o bin/server ./cmd/server/
	@rm -rf cmd/server/web
	@echo "Server built with embedded files: bin/server"

# Build everything for production
build: wasm-prod server-embed cli
	@echo "Building frontend with Vite..."
	@npm run build
	@echo "Production build complete"

# Run the Go server
serve: server
	@echo "Starting Go server on http://localhost:8080"
	@./bin/server -dev

# Run tests
test:
	@echo "Running Go tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting Go code..."
	@go fmt ./...
	@gofmt -s -w .

# Lint code
lint:
	@echo "Running linters..."
	@golangci-lint run 2>/dev/null || go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/ dist/ web/main.wasm web/wasm_exec.js node_modules/
	@echo "Clean complete"

# Docker commands
docker-build:
	@echo "Building Docker image..."
	@docker build -t local-first:latest .

docker-dev:
	@echo "Running development container..."
	@docker run -it --rm \
		-p 5173:5173 \
		-p 8080:8080 \
		-v "$$(pwd)":/app \
		-w /app \
		local-first:latest

docker-prod:
	@echo "Building and running production container..."
	@docker build --target production -t local-first:prod .
	@docker run -it --rm -p 8080:8080 local-first:prod

# Help command
help:
	@echo "Available commands:"
	@echo "  make install     - Install Go and Node dependencies"
	@echo "  make dev        - Start development server with Vite"
	@echo "  make build      - Build for production"
	@echo "  make serve      - Run Go server in dev mode"
	@echo "  make test       - Run tests"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make docker-dev - Run in Docker (development)"
	@echo "  make docker-prod - Run in Docker (production)"
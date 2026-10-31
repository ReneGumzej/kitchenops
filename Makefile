# KitchenOps - Build System
# ============================================================================
# Dieses Makefile erleichtert das plattformübergreifende Bauen des Backends.
#
# Voraussetzungen:
#   - Go 1.26+
#   - GCC/C-Compiler für CGO (sqlite3 benötigt CGO)
#     * Linux:   gcc (paket build-essential oder äquivalent)
#     * MacOS:   XCode Command Line Tools (xcode-select --install)
#     * Windows: TDM-GCC (https://jmeubank.github.io/tdm-gcc/) oder MinGW-w64
#
# Beispiele:
#   make                         # Baue für die aktuelle Plattform
#   make build                   # Baue für die aktuelle Plattform
#   make build-linux             # Baue explizit für Linux (x86_64)
#   make build-linux-arm64       # Baue für Linux ARM64
#   make build-darwin            # Baue für MacOS (Intel)
#   make build-darwin-arm64     # Baue für MacOS (Apple Silicon)
#   make build-windows           # Baue für Windows (x86_64)
#   make clean                   # Lösche alle Build-Artefakte
#   make all                     # Baue für alle Plattformen

# Binärname
BINARY_NAME ?= api

# Build-Zeitstempel für Versionierung
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -ldflags "-s -w -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}"

.PHONY: default build
default: build

# Baue für die aktuelle Plattform (Go erkennt OS/Arch automatisch)
build: build-local

# Baue für Linux x86_64
build-linux:
	@echo "Building for Linux (x86_64)..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 ./cmd/api

# Baue für Linux ARM64
build-linux-arm64:
	@echo "Building for Linux (ARM64)..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-arm64 ./cmd/api

# Baue für MacOS Intel
build-darwin:
	@echo "Building for macOS (Intel)..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 ./cmd/api

# Baue für MacOS Apple Silicon
build-darwin-arm64:
	@echo "Building for macOS (Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 ./cmd/api

# Baue für Windows x86_64
build-windows:
	@echo "Building for Windows (x86_64)..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe ./cmd/api

# Baue für die aktuelle Plattform (automatische Erkennung)
build-local:
	@echo "Building for current platform ($(shell go env GOOS)/$(shell go env GOARCH))..."
	CGO_ENABLED=1 go build ${LDFLAGS} -o bin/${BINARY_NAME} ./cmd/api

# Baue für alle Plattformen
all: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows

# Baue das Docker-Image
.PHONY: docker
docker:
	@echo "Building Docker image..."
	docker build -t kitchenops:latest .

# Baue Docker-Image mit plattformspezifischen Arguments
# Beispiel: make docker GOOS=darwin GOARCH=arm64
docker-custom:
	@echo "Building Docker image for ${GOOS}/${GOARCH}..."
	docker build --build-arg GOOS=${GOOS} --build-arg GOARCH=${GOARCH} -t kitchenops:${GOOS}-${GOARCH} .

# Lösche alle Build-Artefakte
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/

# Führe das Backend aus (Standard: aktuelle Plattform)
run: build-local
	@echo "Starting ${BINARY_NAME}..."
	./bin/${BINARY_NAME}

# Führe Tests aus
test:
	@echo "Running tests..."
	go test -v ./...

# Installiere Abhängigkeiten
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Zeige Hilfe
.PHONY: help
help:
	@echo "KitchenOps Build System"
	@echo "======================="
	@echo ""
	@echo "Available targets:"
	@echo "  make              - Build for current platform"
	@echo "  make build        - Build for current platform"
	@echo "  make build-local  - Build for current platform"
	@echo "  make build-linux       - Build for Linux (x86_64)"
	@echo "  make build-linux-arm64 - Build for Linux (ARM64)"
	@echo "  make build-darwin       - Build for macOS (Intel)"
	@echo "  make build-darwin-arm64 - Build for macOS (Apple Silicon)"
	@echo "  make build-windows      - Build for Windows (x86_64)"
	@echo "  make all          - Build for all platforms"
	@echo "  make docker       - Build Docker image"
	@echo "  make clean        - Remove all build artifacts"
	@echo "  make run          - Build and run"
	@echo "  make test         - Run tests"
	@echo "  make deps         - Download dependencies"
	@echo ""
	@echo "Notes:"
	@echo "  - SQLite (mattn/go-sqlite3) requires CGO and a C compiler"
	@echo "  - Linux:   install gcc (apt-get install build-essential)"
	@echo "  - macOS:   install XCode Command Line Tools (xcode-select --install)"
	@echo "  - Windows: install TDM-GCC or MinGW-w64"

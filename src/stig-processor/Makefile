# STIG Processor Makefile - Updated for src/ directory structure

# Variables
BINARY_NAME=stig-processor
GO_VERSION=1.25
DOCKER_REGISTRY=ghcr.io
DOCKER_IMAGE=$(DOCKER_REGISTRY)/$(DOCKER_REPO)/$(BINARY_NAME)
DOCKER_REPO?=your-org
VERSION?=latest
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Default STIG input file (relative to project root)
INPUT_FILE?=../../microsoft-windows-11-security-technical-implementation-guide.json
OUTPUT_DIR?=./output
FORMAT?=yaml
SEVERITY?=

# Go build flags with version info
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.gitCommit=$(GIT_COMMIT)"
CGO_ENABLED=0
GOOS=linux
GOARCH=amd64

# Source and build directories
SRC_DIR=.
CMD_DIR=cmd
PKG_DIR=pkg
INTERNAL_DIR=internal

.PHONY: help build run test clean docker-build docker-run docker-push install deps fmt lint vet security-scan dev-setup

# Default target
help: ## Show this help message
	@echo "STIG Processor - Makefile commands (src/ structure):"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
dev-setup: ## Set up development environment
	@echo "Setting up development environment for src/ structure..."
	go version
	go mod download
	go mod tidy
	@echo "✅ Development environment ready"

deps: ## Download and verify dependencies
	go mod download
	go mod verify
	go mod tidy

fmt: ## Format Go code
	go fmt ./...
	gofmt -s -w .

lint: ## Run golangci-lint
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi

vet: ## Run go vet
	go vet ./...

# Build targets
build: deps fmt vet ## Build the binary
	@echo "Building $(BINARY_NAME) from $(CMD_DIR)..."
	cd $(CMD_DIR) && CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(LDFLAGS) -o ../$(BINARY_NAME) .
	@echo "✅ Binary built: $(BINARY_NAME)"

build-local: ## Build for local development (current OS)
	@echo "Building $(BINARY_NAME) for local development..."
	cd $(CMD_DIR) && go build $(LDFLAGS) -o ../$(BINARY_NAME) .
	@echo "✅ Local binary built: $(BINARY_NAME)"

build-debug: ## Build with debug symbols
	@echo "Building $(BINARY_NAME) with debug symbols..."
	cd $(CMD_DIR) && go build -gcflags="all=-N -l" $(LDFLAGS) -o ../$(BINARY_NAME)-debug .
	@echo "✅ Debug binary built: $(BINARY_NAME)-debug"

install: build-local ## Install binary to GOPATH/bin
	cd $(CMD_DIR) && go install $(LDFLAGS) .
	@echo "✅ Installed $(BINARY_NAME) to $(shell go env GOPATH)/bin"

# Run targets
run: build-local ## Run the processor locally
	./$(BINARY_NAME) -input $(INPUT_FILE) -output $(OUTPUT_DIR) -format $(FORMAT) $(if $(SEVERITY),-severity $(SEVERITY)) -verbose

run-dev: ## Run with go run (development)
	cd $(CMD_DIR) && go run $(LDFLAGS) . -input ../$(INPUT_FILE) -output ../$(OUTPUT_DIR) -format $(FORMAT) $(if $(SEVERITY),-severity $(SEVERITY)) -verbose

run-dry: ## Run in dry-run mode (no file output)
	./$(BINARY_NAME) -input $(INPUT_FILE) -output $(OUTPUT_DIR) -format $(FORMAT) $(if $(SEVERITY),-severity $(SEVERITY)) -verbose -dry-run

run-stats: ## Show STIG statistics only
	./$(BINARY_NAME) -input $(INPUT_FILE) -stats -verbose

run-validate: ## Validate existing policies
	./$(BINARY_NAME) -output $(OUTPUT_DIR) -validate -verbose

# Test targets
test: ## Run all tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-benchmark: ## Run benchmark tests
	go test -bench=. -benchmem ./...

test-integration: build-local ## Run integration tests
	@echo "Running integration tests..."
	./$(BINARY_NAME) -input $(INPUT_FILE) -output test-output -dry-run -verbose
	@echo "✅ Integration tests passed"

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image: $(DOCKER_IMAGE):$(VERSION)"
	docker build -t $(DOCKER_IMAGE):$(VERSION) -f Dockerfile .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(VERSION)"

docker-run: ## Run Docker container
	@echo "Running STIG processor in container..."
	@mkdir -p $(OUTPUT_DIR)
	docker run --rm \
		-v $(shell dirname $(shell pwd)/$(INPUT_FILE)):/app/input \
		-v $(shell pwd)/$(OUTPUT_DIR):/app/output \
		$(DOCKER_IMAGE):$(VERSION) \
		-input /app/input/$(shell basename $(INPUT_FILE)) \
		-output /app/output \
		-format $(FORMAT) \
		$(if $(SEVERITY),-severity $(SEVERITY)) \
		-verbose

docker-run-interactive: ## Run Docker container interactively
	docker run --rm -it \
		-v $(shell dirname $(shell pwd)/$(INPUT_FILE)):/app/input \
		-v $(shell pwd)/$(OUTPUT_DIR):/app/output \
		--entrypoint /bin/sh \
		$(DOCKER_IMAGE):$(VERSION)

docker-push: docker-build ## Push Docker image to registry
	@echo "Pushing Docker image to registry..."
	docker push $(DOCKER_IMAGE):$(VERSION)
	docker push $(DOCKER_IMAGE):latest
	@echo "✅ Docker image pushed: $(DOCKER_IMAGE):$(VERSION)"

# Utility targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME) $(BINARY_NAME)-debug
	rm -f coverage.out coverage.html
	rm -rf $(OUTPUT_DIR) test-output
	docker system prune -f --filter label=stage=builder
	@echo "✅ Cleanup complete"

security-scan: ## Run security scans
	@echo "Running security scans..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi
	@if command -v nancy >/dev/null 2>&1; then \
		go list -json -deps ./... | nancy sleuth; \
	else \
		echo "nancy not installed. Install with: go install github.com/sonatypecommunity/nancy@latest"; \
	fi

check-stig-file: ## Check if STIG input file exists
	@if [ ! -f "$(INPUT_FILE)" ]; then \
		echo "❌ STIG file not found: $(INPUT_FILE)"; \
		echo "Please ensure the STIG JSON file is available"; \
		echo "You can download it from: https://www.stigviewer.com/stigs"; \
		exit 1; \
	else \
		echo "✅ STIG file found: $(INPUT_FILE)"; \
	fi

validate-output: ## Validate generated Fleet policies
	@echo "Validating generated Fleet policies..."
	@if [ ! -d "$(OUTPUT_DIR)" ]; then \
		echo "❌ Output directory not found: $(OUTPUT_DIR)"; \
		echo "Run 'make run' first to generate policies"; \
		exit 1; \
	fi
	./$(BINARY_NAME) -output $(OUTPUT_DIR) -validate
	@policy_count=$$(find $(OUTPUT_DIR) -name "stig-*.yaml" -o -name "stig-*.json" | wc -l | tr -d ' '); \
	echo "Generated $$policy_count Fleet policies"

# Code quality targets
quality-check: fmt vet lint test ## Run all code quality checks
	@echo "✅ All quality checks passed"

# CI/CD targets
ci-build: deps quality-check build ## Full CI build pipeline
	@echo "✅ CI build pipeline complete"

ci-test: test test-coverage security-scan ## Full CI test pipeline
	@echo "✅ CI test pipeline complete"

release: clean ci-build docker-build ## Prepare release artifacts
	@echo "✅ Release artifacts ready"
	@echo "Binary: $(BINARY_NAME)"
	@echo "Docker image: $(DOCKER_IMAGE):$(VERSION)"
	@echo "Version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"

# Quick workflow commands
quick-start: check-stig-file build run validate-output ## Quick start: build, run, and validate
	@echo "🎉 STIG processor quick start complete!"
	@echo "Check the output directory: $(OUTPUT_DIR)"

demo: ## Run demo with verbose output
	@echo "Running STIG processor demo..."
	$(MAKE) run SEVERITY=medium FORMAT=yaml
	@echo ""
	@echo "Demo complete! Check $(OUTPUT_DIR) for generated Fleet policies"
	@ls -la $(OUTPUT_DIR) || echo "No output files generated"

# Development helpers
watch: ## Watch for file changes and rebuild
	@echo "Watching for changes... (requires entr: brew install entr)"
	find . -name "*.go" | entr -r make build-local

hot-reload: ## Hot reload development server
	@echo "Hot reloading... (requires air: go install github.com/cosmtrek/air@latest)"
	air

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	go doc -all . > DOCS.md
	@echo "✅ Documentation generated: DOCS.md"

# Debugging targets
debug-build: build-debug ## Build and prepare for debugging
	@echo "Debug binary ready: $(BINARY_NAME)-debug"
	@echo "Run with: dlv exec ./$(BINARY_NAME)-debug -- [args]"

debug-run: debug-build ## Run with delve debugger
	@if command -v dlv >/dev/null 2>&1; then \
		dlv exec ./$(BINARY_NAME)-debug -- -input $(INPUT_FILE) -output $(OUTPUT_DIR) -verbose; \
	else \
		echo "delve not installed. Install with: go install github.com/go-delve/delve/cmd/dlv@latest"; \
	fi

# Package structure validation
validate-structure: ## Validate package structure
	@echo "Validating package structure..."
	@test -d $(PKG_DIR) || (echo "❌ Missing $(PKG_DIR) directory" && exit 1)
	@test -d $(INTERNAL_DIR) || (echo "❌ Missing $(INTERNAL_DIR) directory" && exit 1)
	@test -d $(CMD_DIR) || (echo "❌ Missing $(CMD_DIR) directory" && exit 1)
	@test -f $(CMD_DIR)/main.go || (echo "❌ Missing $(CMD_DIR)/main.go" && exit 1)
	@echo "✅ Package structure is valid"

# Example workflows
examples: ## Show example commands
	@echo "Example commands for src/ structure:"
	@echo ""
	@echo "Basic usage:"
	@echo "  make quick-start"
	@echo ""
	@echo "Filter by severity:"
	@echo "  make run SEVERITY=high"
	@echo ""
	@echo "JSON output:"
	@echo "  make run FORMAT=json"
	@echo ""
	@echo "Development workflow:"
	@echo "  make dev-setup"
	@echo "  make run-dev"
	@echo "  make test-coverage"
	@echo ""
	@echo "Docker workflow:"
	@echo "  make docker-build"
	@echo "  make docker-run"
	@echo ""
	@echo "Statistics and validation:"
	@echo "  make run-stats"
	@echo "  make run-validate"
	@echo ""
	@echo "Debugging:"
	@echo "  make debug-build"
	@echo "  make debug-run"

# STIG Benchmark Builder - Build System
# =====================================

.PHONY: all build-schema build-web build-all dev clean help

# Default target
all: build-all

# Configuration
WIN_STIG_PATH ?= $(HOME)/Code/GitHub/win-stig
SCHEMA_OUTPUT = web/src/data/benchmark-data.json

# Help
help:
	@echo "STIG Benchmark Builder"
	@echo "======================"
	@echo ""
	@echo "Targets:"
	@echo "  build-schema  - Generate benchmark-data.json from STIG + fixes"
	@echo "  build-web     - Build the Vue.js web application"
	@echo "  build-all     - Build schema and web app"
	@echo "  dev           - Start development server"
	@echo "  clean         - Clean build artifacts"
	@echo ""
	@echo "Configuration:"
	@echo "  WIN_STIG_PATH - Path to win-stig repository (default: ~/Code/GitHub/win-stig)"
	@echo ""
	@echo "Examples:"
	@echo "  make build-all"
	@echo "  make dev"
	@echo "  WIN_STIG_PATH=/path/to/win-stig make build-schema"

# Build the Go schema-builder
build-schema-builder:
	@echo "Building schema-builder..."
	cd src/schema-builder && go build -o ../../bin/schema-builder .

# Generate benchmark data from STIG + fixes
build-schema: build-schema-builder
	@echo "Generating benchmark data..."
	@mkdir -p web/src/data
	./bin/schema-builder -win-stig "$(WIN_STIG_PATH)" -output "$(SCHEMA_OUTPUT)" -verbose

# Install web dependencies
web-deps:
	@echo "Installing web dependencies..."
	cd web && npm install

# Build the Vue.js web application
build-web: web-deps
	@echo "Building web application..."
	cd web && npm run build
	@echo "Build complete! Output in web/dist/"

# Build everything
build-all: build-schema build-web
	@echo ""
	@echo "Build complete!"
	@echo "  Schema: $(SCHEMA_OUTPUT)"
	@echo "  Web:    web/dist/"

# Start development server
dev: build-schema web-deps
	@echo "Starting development server..."
	cd web && npm run dev

# Preview production build
preview: build-all
	cd web && npm run preview

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web/dist/
	rm -rf web/node_modules/
	rm -f $(SCHEMA_OUTPUT)
	@echo "Clean complete!"

# Run the stig-processor CLI
run-processor:
	cd src/stig-processor && go run . $(ARGS)

# Build stig-processor
build-processor:
	cd src/stig-processor && go build -o ../../bin/stig-processor .

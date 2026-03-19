.PHONY: all build build-wasm clean test generate generate-example lint

# Variables
PLUGIN_NAME := sqlc-effect
WASM_DIR := dist
WASM_FILE := $(WASM_DIR)/plugin
EXAMPLE_DIR := examples/effect-v4

# Default target
all: build

# Build the wasm plugin
build: build-wasm

build-wasm: $(WASM_FILE)

$(WASM_FILE): cmd/plugin/main.go $(shell find cmd/plugin/internal -name '*.go')
	@mkdir -p $(WASM_DIR)
	@echo "Building wasm plugin..."
	GOOS=wasip1 GOARCH=wasm go build -o $(WASM_FILE) ./cmd/plugin
	@echo "✓ Built: $(WASM_FILE)"

# Generate code for the example (builds wasm first)
generate: build
	@echo "Generating code for example..."
	cd $(EXAMPLE_DIR) && sqlc generate
	@echo "✓ Code generation complete"

# Generate with debug mode enabled
generate-debug: build
	@echo "Generating code with debug mode..."
	cd $(EXAMPLE_DIR) && sqlc generate
	@echo "✓ Code generation complete (check $(EXAMPLE_DIR)/src/models/debug/)"

# Quick generate without rebuilding wasm (use when plugin hasn't changed)
generate-fast:
	@echo "Generating code (fast mode)..."
	cd $(EXAMPLE_DIR) && sqlc generate
	@echo "✓ Code generation complete"

# Clean all build artifacts and generated code
clean: clean-wasm clean-generated

# Clean wasm build
clean-wasm:
	rm -rf $(WASM_DIR)
	@echo "✓ Cleaned wasm build"

# Clean generated code
clean-generated:
	rm -rf $(EXAMPLE_DIR)/src/models/
	@echo "✓ Cleaned generated code"

# Run tests
test:
	@echo "Running tests..."
	go test ./...
	@echo "✓ Tests complete"

# Run Go linter (if golangci-lint is installed)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "⚠ golangci-lint not installed, skipping"; \
	fi

# Full clean and rebuild
clean-all: clean generate

# Development workflow - build, generate, and show output
dev: generate
	@echo ""
	@echo "Generated files:"
	@ls -la $(EXAMPLE_DIR)/src/models/ 2>/dev/null || echo "(no output directory yet)"

# Check if sqlc is installed
check-sqlc:
	@if ! command -v sqlc >/dev/null 2>&1; then \
		echo "❌ sqlc is not installed"; \
		echo "   Install with: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"; \
		exit 1; \
	fi
	@echo "✓ sqlc is installed: $$(sqlc version)"

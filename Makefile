.PHONY: all build build-effect build-native clean test generate generate-effect generate-native generate-fast-effect generate-fast-native lint

# Variables
WASM_DIR := dist
EFFECT_WASM_FILE := $(WASM_DIR)/sqlc-gen-effect.wasm
NATIVE_WASM_FILE := $(WASM_DIR)/sqlc-gen-native.wasm
EFFECT_EXAMPLE_DIR := examples/effect-v4
NATIVE_EXAMPLE_DIR := examples/native
EFFECT_SOURCES := cmd/effect/main.go $(shell find cmd/effect/internal toolbelt -type f \( -name '*.go' -o -name '*.gotmpl' \))
NATIVE_SOURCES := cmd/native/main.go $(shell find cmd/native/internal toolbelt -type f \( -name '*.go' -o -name '*.gotmpl' \))

# Default target
all: build

# Build the wasm plugins
build: build-effect build-native

build-effect: $(EFFECT_WASM_FILE)

build-native: $(NATIVE_WASM_FILE)

$(EFFECT_WASM_FILE): $(EFFECT_SOURCES)
	@mkdir -p $(WASM_DIR)
	@echo "Building Effect wasm plugin..."
	GOOS=wasip1 GOARCH=wasm go build -o $(EFFECT_WASM_FILE) ./cmd/effect
	@echo "✓ Built: $(EFFECT_WASM_FILE)"

$(NATIVE_WASM_FILE): $(NATIVE_SOURCES)
	@mkdir -p $(WASM_DIR)
	@echo "Building native wasm plugin..."
	GOOS=wasip1 GOARCH=wasm go build -o $(NATIVE_WASM_FILE) ./cmd/native
	@echo "✓ Built: $(NATIVE_WASM_FILE)"

# Generate code for examples (builds wasm first)
generate: generate-effect generate-native

generate-effect: build-effect
	@echo "Generating code for Effect example..."
	cd $(EFFECT_EXAMPLE_DIR) && sqlc generate
	@echo "✓ Effect code generation complete"

generate-native: build-native
	@echo "Generating code for native example..."
	cd $(NATIVE_EXAMPLE_DIR) && sqlc generate
	@echo "✓ Native code generation complete"

generate-fast-effect:
	@echo "Generating Effect code (fast mode)..."
	cd $(EFFECT_EXAMPLE_DIR) && sqlc generate
	@echo "✓ Effect code generation complete"

generate-fast-native:
	@echo "Generating native code (fast mode)..."
	cd $(NATIVE_EXAMPLE_DIR) && sqlc generate
	@echo "✓ Native code generation complete"

generate-fast: generate-fast-effect generate-fast-native
	@echo "✓ Code generation complete"

# Generate with debug mode enabled
generate-debug: build
	@echo "Generating code with debug mode..."
	cd $(EFFECT_EXAMPLE_DIR) && sqlc generate
	cd $(NATIVE_EXAMPLE_DIR) && sqlc generate
	@echo "✓ Code generation complete"

# Clean all build artifacts and generated code
clean: clean-wasm clean-generated

# Clean wasm build
clean-wasm:
	rm -rf $(WASM_DIR)
	@echo "✓ Cleaned wasm build"

# Clean generated code
clean-generated:
	rm -rf $(EFFECT_EXAMPLE_DIR)/src/models/ $(NATIVE_EXAMPLE_DIR)/src/*.ts
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
	@ls -la $(EFFECT_EXAMPLE_DIR)/src/models/ 2>/dev/null || echo "(no Effect output directory yet)"
	@ls -la $(NATIVE_EXAMPLE_DIR)/src/ 2>/dev/null || echo "(no native output directory yet)"

# Check if sqlc is installed
check-sqlc:
	@if ! command -v sqlc >/dev/null 2>&1; then \
		echo "❌ sqlc is not installed"; \
		echo "   Install with: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest"; \
		exit 1; \
	fi
	@echo "✓ sqlc is installed: $$(sqlc version)"

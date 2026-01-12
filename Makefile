# ==============================================================================
# TelemetryFlow GO MCP Server Makefile
# Version: 1.1.2
# ==============================================================================

# Build variables
BINARY_NAME := tfo-mcp
VERSION := 1.1.2
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION := $(shell go version | cut -d ' ' -f 3)

# Directories
BUILD_DIR := build
CMD_DIR := cmd/mcp
DIST_DIR := dist

# Go build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"
LDFLAGS_RELEASE := -ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE)"

# Go tooling
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod
GOVET := $(GOCMD) vet
GOFMT := gofmt

# Default target
.DEFAULT_GOAL := help

# ==============================================================================
# DEVELOPMENT
# ==============================================================================

.PHONY: build
build: ## Build the binary for development
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-release
build-release: ## Build optimized release binary
	@echo "Building release $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_RELEASE) -o $(BUILD_DIR)/$(BINARY_NAME) ./$(CMD_DIR)
	@echo "Release build complete: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: run
run: build ## Build and run the server
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

.PHONY: run-debug
run-debug: build ## Run the server in debug mode
	@echo "Running $(BINARY_NAME) in debug mode..."
	./$(BUILD_DIR)/$(BINARY_NAME) --debug

.PHONY: install
install: build ## Install the binary to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@$(GOCMD) clean -cache
	@echo "Clean complete"

# ==============================================================================
# DEPENDENCIES
# ==============================================================================

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	@echo "Dependencies downloaded"

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify
	@echo "Dependencies updated"

.PHONY: deps-vendor
deps-vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	$(GOMOD) vendor
	@echo "Dependencies vendored"

.PHONY: deps-refresh
deps-refresh: ## Refresh all dependencies (clean and re-download)
	@echo "Refreshing dependencies..."
	@rm -rf vendor go.sum
	@echo "Clearing module cache..."
	$(GOCMD) clean -modcache
	$(GOMOD) download
	$(GOMOD) tidy
	$(GOMOD) verify
	@echo "Dependencies refreshed"

.PHONY: deps-check
deps-check: ## Check for dependency vulnerabilities
	@echo "Checking dependencies for vulnerabilities..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Run: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

.PHONY: deps-graph
deps-graph: ## Show dependency graph
	@echo "Generating dependency graph..."
	$(GOMOD) graph

.PHONY: deps-why
deps-why: ## Explain why a dependency is needed (use DEP=module/path)
	@if [ -z "$(DEP)" ]; then \
		echo "Usage: make deps-why DEP=github.com/some/module"; \
	else \
		$(GOMOD) why -m $(DEP); \
	fi

# ==============================================================================
# CODE QUALITY
# ==============================================================================

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	@echo "Code formatted"

.PHONY: fmt-check
fmt-check: ## Check code formatting
	@echo "Checking code format..."
	@test -z "$$($(GOFMT) -l .)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)
	@echo "Code format check passed"

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...
	@echo "Vet complete"

.PHONY: lint
lint: ## Run golangci-lint
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: lint-fix
lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running linter with auto-fix..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix ./...; \
	else \
		echo "golangci-lint not installed"; \
	fi

# ==============================================================================
# TESTING
# ==============================================================================

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v -race ./...
	@echo "Tests complete"

.PHONY: test-cover
test-cover: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@mkdir -p $(BUILD_DIR)
	$(GOTEST) -v -race -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GOCMD) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report: $(BUILD_DIR)/coverage.html"

.PHONY: test-bench
test-bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

.PHONY: test-short
test-short: ## Run short tests only
	@echo "Running short tests..."
	$(GOTEST) -v -short ./...

.PHONY: test-all
test-all: ## Run all tests (unit, integration, e2e)
	@echo "Running all tests..."
	$(GOTEST) -v -race -count=1 ./...
	@echo "All tests complete"

.PHONY: ci-test
ci-test: fmt-check vet lint test ## Run CI pipeline (format, vet, lint, test)
	@echo "CI pipeline complete"

# ==============================================================================
# CI-SPECIFIC TARGETS (GitHub Actions)
# ==============================================================================

.PHONY: test-unit-ci
test-unit-ci: ## Run unit tests for CI with coverage output
	@echo "Running unit tests for CI..."
	$(GOTEST) -v -race -coverprofile=coverage-unit.out ./tests/unit/...
	@echo "Unit tests complete"

.PHONY: test-integration-ci
test-integration-ci: ## Run integration tests for CI with coverage output
	@echo "Running integration tests for CI..."
	$(GOTEST) -v -race -coverprofile=coverage-integration.out ./tests/integration/...
	@echo "Integration tests complete"

.PHONY: test-e2e-ci
test-e2e-ci: ## Run E2E tests for CI
	@echo "Running E2E tests for CI..."
	$(GOTEST) -v -race ./tests/e2e/...
	@echo "E2E tests complete"

.PHONY: ci-build
ci-build: ## Build binary for CI (uses GOOS/GOARCH env vars)
	@echo "Building for CI ($(GOOS)/$(GOARCH))..."
	@mkdir -p $(BUILD_DIR)
	@if [ "$(GOOS)" = "windows" ]; then \
		CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_RELEASE) -o $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH).exe ./$(CMD_DIR); \
	else \
		CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_RELEASE) -o $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) ./$(CMD_DIR); \
	fi
	@echo "CI build complete"

.PHONY: deps-verify
deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	$(GOMOD) download
	$(GOMOD) verify
	@echo "Dependencies verified"

.PHONY: staticcheck
staticcheck: ## Run staticcheck
	@echo "Running staticcheck..."
	@if command -v staticcheck >/dev/null 2>&1; then \
		staticcheck ./...; \
	else \
		echo "staticcheck not installed. Run: go install honnef.co/go/tools/cmd/staticcheck@latest"; \
	fi

.PHONY: govulncheck
govulncheck: ## Run govulncheck for vulnerability scanning
	@echo "Running govulncheck..."
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed. Run: go install golang.org/x/vuln/cmd/govulncheck@latest"; \
	fi

.PHONY: coverage-report
coverage-report: ## Generate merged coverage report
	@echo "Generating coverage report..."
	@if [ -f coverage-unit.out ] && [ -f coverage-integration.out ]; then \
		echo "mode: atomic" > coverage-merged.out; \
		tail -n +2 coverage-unit.out >> coverage-merged.out; \
		tail -n +2 coverage-integration.out >> coverage-merged.out; \
	elif [ -f coverage-unit.out ]; then \
		cp coverage-unit.out coverage-merged.out; \
	else \
		echo "No coverage files found"; \
		exit 1; \
	fi
	$(GOCMD) tool cover -func=coverage-merged.out > coverage-summary.txt
	$(GOCMD) tool cover -html=coverage-merged.out -o coverage.html
	@echo "Coverage report generated"
	@cat coverage-summary.txt

# ==============================================================================
# CROSS-COMPILATION
# ==============================================================================

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_RELEASE) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_RELEASE) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./$(CMD_DIR)
	# macOS
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_RELEASE) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_RELEASE) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	# Windows
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_RELEASE) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "Cross-compilation complete"
	@ls -la $(DIST_DIR)

.PHONY: build-linux
build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_RELEASE) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./$(CMD_DIR)
	@echo "Linux build complete"

.PHONY: build-darwin
build-darwin: ## Build for macOS
	@echo "Building for macOS..."
	@mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_RELEASE) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	@echo "macOS build complete"

.PHONY: build-windows
build-windows: ## Build for Windows
	@echo "Building for Windows..."
	@mkdir -p $(DIST_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS_RELEASE) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	@echo "Windows build complete"

# ==============================================================================
# DOCKER
# ==============================================================================

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t telemetryflow-mcp:$(VERSION) -t telemetryflow-mcp:latest .
	@echo "Docker image built"

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run --rm -it \
		-e ANTHROPIC_API_KEY \
		telemetryflow-mcp:latest

.PHONY: docker-push
docker-push: ## Push Docker image
	@echo "Pushing Docker image..."
	docker push telemetryflow-mcp:$(VERSION)
	docker push telemetryflow-mcp:latest

# ==============================================================================
# CI/CD
# ==============================================================================

.PHONY: ci
ci: deps fmt-check vet lint test ## Run CI pipeline
	@echo "CI pipeline complete"

.PHONY: ci-validate
ci-validate: ## Validate CI configuration
	@echo "Validating CI configuration..."
	$(GOMOD) verify
	@echo "CI configuration valid"

# ==============================================================================
# RELEASE
# ==============================================================================

.PHONY: release
release: clean build-all ## Create release artifacts
	@echo "Creating release $(VERSION)..."
	@mkdir -p $(DIST_DIR)/release
	@cd $(DIST_DIR) && \
		for f in $(BINARY_NAME)-*; do \
			if [ -f "$$f" ]; then \
				tar -czvf "release/$$f.tar.gz" "$$f"; \
			fi \
		done
	@echo "Release artifacts created in $(DIST_DIR)/release"

.PHONY: changelog
changelog: ## Generate changelog
	@echo "Generating changelog..."
	@if command -v git-chglog >/dev/null 2>&1; then \
		git-chglog -o CHANGELOG.md; \
	else \
		echo "git-chglog not installed"; \
	fi

# ==============================================================================
# UTILITIES
# ==============================================================================

.PHONY: version
version: ## Show version information
	@echo "TelemetryFlow GO MCP Server"
	@echo "Version:    $(VERSION)"
	@echo "Commit:     $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Go Version: $(GO_VERSION)"

.PHONY: validate-config
validate-config: build ## Validate configuration file
	@echo "Validating configuration..."
	./$(BUILD_DIR)/$(BINARY_NAME) validate
	@echo "Configuration valid"

.PHONY: generate
generate: ## Run go generate
	@echo "Running go generate..."
	$(GOCMD) generate ./...
	@echo "Generate complete"

.PHONY: docs
docs: ## Generate documentation
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		godoc -http=:6060 & \
		echo "Documentation server started at http://localhost:6060"; \
	else \
		echo "godoc not installed"; \
	fi

# ==============================================================================
# HELP
# ==============================================================================

.PHONY: help
help: ## Show this help message
	@echo "TelemetryFlow GO MCP Server - Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Environment Variables:"
	@echo "  ANTHROPIC_API_KEY              - Claude API key (required for running)"
	@echo "  TELEMETRYFLOW_MCP_DEBUG        - Enable debug mode"
	@echo "  TELEMETRYFLOW_MCP_LOG_LEVEL    - Log level (debug, info, warn, error)"

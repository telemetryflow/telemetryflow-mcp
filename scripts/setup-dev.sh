#!/usr/bin/env bash
# ==============================================================================
# TFO-MCP Development Setup Script
# Version: 1.1.2
# Description: Setup development environment for TelemetryFlow GO MCP Server
# ==============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ==============================================================================
# Functions
# ==============================================================================

print_header() {
    echo -e "${PURPLE}"
    echo "╔═════════════════════════════════════════════════════════════════╗"
    echo "║     TelemetryFlow GO MCP Server - Development Setup Script      ║"
    echo "╚═════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_go() {
    log_info "Checking Go installation..."

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        log_info "Please install Go 1.24 or higher from https://golang.org/dl/"
        exit 1
    fi

    GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    log_success "Go version: ${GO_VERSION}"
}

check_git() {
    log_info "Checking Git installation..."

    if ! command -v git &> /dev/null; then
        log_error "Git is not installed"
        exit 1
    fi

    GIT_VERSION=$(git --version | awk '{print $3}')
    log_success "Git version: ${GIT_VERSION}"
}

install_go_tools() {
    log_info "Installing Go development tools..."

    # golangci-lint
    log_info "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

    # goimports
    log_info "Installing goimports..."
    go install golang.org/x/tools/cmd/goimports@latest

    # staticcheck
    log_info "Installing staticcheck..."
    go install honnef.co/go/tools/cmd/staticcheck@latest

    # mockgen
    log_info "Installing mockgen..."
    go install go.uber.org/mock/mockgen@latest

    # godoc
    log_info "Installing godoc..."
    go install golang.org/x/tools/cmd/godoc@latest

    # air (live reload)
    log_info "Installing air (live reload)..."
    go install github.com/air-verse/air@latest

    # dlv (debugger)
    log_info "Installing delve debugger..."
    go install github.com/go-delve/delve/cmd/dlv@latest

    log_success "Go tools installed"
}

download_dependencies() {
    log_info "Downloading Go dependencies..."

    go mod download
    go mod verify

    log_success "Dependencies downloaded"
}

setup_git_hooks() {
    log_info "Setting up Git hooks..."

    HOOKS_DIR=".git/hooks"
    mkdir -p "${HOOKS_DIR}"

    # Pre-commit hook
    cat > "${HOOKS_DIR}/pre-commit" << 'EOF'
#!/usr/bin/env bash
# TFO-MCP Pre-commit Hook

set -e

echo "Running pre-commit checks..."

# Format code
echo "Checking code format..."
UNFORMATTED=$(gofmt -l . 2>&1 | grep -v vendor || true)
if [[ -n "${UNFORMATTED}" ]]; then
    echo "Error: Files not formatted:"
    echo "${UNFORMATTED}"
    echo "Run: gofmt -w ."
    exit 1
fi

# Run vet
echo "Running go vet..."
go vet ./...

# Run tests
echo "Running short tests..."
go test -short -race ./...

echo "Pre-commit checks passed!"
EOF

    chmod +x "${HOOKS_DIR}/pre-commit"

    # Pre-push hook
    cat > "${HOOKS_DIR}/pre-push" << 'EOF'
#!/usr/bin/env bash
# TFO-MCP Pre-push Hook

set -e

echo "Running pre-push checks..."

# Run full tests
echo "Running full test suite..."
go test -race ./...

# Run linter if available
if command -v golangci-lint &> /dev/null; then
    echo "Running linter..."
    golangci-lint run ./...
fi

echo "Pre-push checks passed!"
EOF

    chmod +x "${HOOKS_DIR}/pre-push"

    # Commit-msg hook
    cat > "${HOOKS_DIR}/commit-msg" << 'EOF'
#!/usr/bin/env bash
# TFO-MCP Commit Message Hook

COMMIT_MSG_FILE=$1
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

# Check commit message format
if ! echo "$COMMIT_MSG" | grep -qE "^(feat|fix|docs|style|refactor|test|chore|perf|ci|build|revert)(\(.+\))?: .+"; then
    echo "Error: Invalid commit message format"
    echo ""
    echo "Expected format: <type>(<scope>): <subject>"
    echo ""
    echo "Types: feat, fix, docs, style, refactor, test, chore, perf, ci, build, revert"
    echo ""
    echo "Examples:"
    echo "  feat(tools): add new search tool"
    echo "  fix(session): resolve timeout issue"
    echo "  docs: update README"
    exit 1
fi

echo "Commit message format valid"
EOF

    chmod +x "${HOOKS_DIR}/commit-msg"

    log_success "Git hooks installed"
}

create_env_file() {
    log_info "Creating .env file..."

    if [[ -f ".env" ]]; then
        log_warning ".env file already exists, skipping"
        return
    fi

    cp .env.example .env

    log_success "Created .env file from .env.example"
    log_warning "Please edit .env and add your API keys"
}

create_local_config() {
    log_info "Creating local configuration..."

    if [[ -f "config.local.yaml" ]]; then
        log_warning "config.local.yaml already exists, skipping"
        return
    fi

    cat > "config.local.yaml" << 'EOF'
# Local development configuration
# This file is gitignored - safe for local secrets

server:
  name: "tfo-mcp-dev"
  timeout: 60s

claude:
  # Set via environment variable for security
  api_key: ""
  model: "claude-3-5-haiku-20241022"  # Faster model for development
  max_tokens: 2048

logging:
  level: "debug"
  format: "text"
  caller: true

telemetry:
  enabled: false
EOF

    log_success "Created config.local.yaml"
}

setup_vscode() {
    log_info "Setting up VS Code configuration..."

    mkdir -p .vscode

    # Settings
    cat > ".vscode/settings.json" << 'EOF'
{
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "go.formatTool": "goimports",
  "go.testFlags": ["-v", "-race"],
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": "explicit"
  },
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  },
  "go.testEnvVars": {
    "TELEMETRYFLOW_MCP_CLAUDE_API_KEY": "${env:TELEMETRYFLOW_MCP_CLAUDE_API_KEY}"
  }
}
EOF

    # Launch configuration
    cat > ".vscode/launch.json" << 'EOF'
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug TFO-MCP",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/mcp",
      "args": ["run", "--log-level", "debug"],
      "envFile": "${workspaceFolder}/.env"
    },
    {
      "name": "Debug Tests",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${workspaceFolder}/internal/domain/valueobjects",
      "args": ["-test.v"]
    },
    {
      "name": "Debug Current Test",
      "type": "go",
      "request": "launch",
      "mode": "test",
      "program": "${file}",
      "args": ["-test.v", "-test.run", "${selectedText}"]
    }
  ]
}
EOF

    # Extensions recommendations
    cat > ".vscode/extensions.json" << 'EOF'
{
  "recommendations": [
    "golang.go",
    "ms-vscode.makefile-tools",
    "redhat.vscode-yaml",
    "streetsidesoftware.code-spell-checker",
    "eamodio.gitlens",
    "gruntfuggly.todo-tree"
  ]
}
EOF

    log_success "VS Code configuration created"
}

verify_setup() {
    log_info "Verifying setup..."

    # Check Go tools
    local tools=("golangci-lint" "goimports" "staticcheck" "mockgen" "dlv")

    for tool in "${tools[@]}"; do
        if command -v "${tool}" &> /dev/null; then
            log_success "${tool} available"
        else
            log_warning "${tool} not found in PATH"
        fi
    done

    # Try to build
    log_info "Attempting build..."
    if go build -o /dev/null ./cmd/mcp 2>/dev/null; then
        log_success "Build successful"
    else
        log_warning "Build failed - check dependencies"
    fi

    log_success "Setup verification complete"
}

print_next_steps() {
    echo ""
    echo -e "${GREEN}════════════════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}                    Development Setup Complete!                         ${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo "Next steps:"
    echo ""
    echo "  1. Set your Claude API key:"
    echo "     export TELEMETRYFLOW_MCP_CLAUDE_API_KEY=\"your-api-key\""
    echo ""
    echo "  2. Build the project:"
    echo "     make build"
    echo ""
    echo "  3. Run tests:"
    echo "     make test"
    echo ""
    echo "  4. Run with live reload:"
    echo "     air"
    echo ""
    echo "  5. Start debugging in VS Code:"
    echo "     Press F5 or use 'Debug TFO-MCP' configuration"
    echo ""
}

show_help() {
    echo "TFO-MCP Development Setup Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  all           Run all setup steps (default)"
    echo "  tools         Install Go tools only"
    echo "  deps          Download dependencies only"
    echo "  hooks         Setup Git hooks only"
    echo "  vscode        Setup VS Code configuration only"
    echo "  verify        Verify setup"
    echo "  help          Show this help message"
}

# ==============================================================================
# Main
# ==============================================================================

main() {
    print_header

    # Change to project root
    cd "$(dirname "$0")/.."

    local command="${1:-all}"

    case "${command}" in
        all)
            check_go
            check_git
            install_go_tools
            download_dependencies
            setup_git_hooks
            create_env_file
            create_local_config
            setup_vscode
            verify_setup
            print_next_steps
            ;;
        tools)
            check_go
            install_go_tools
            ;;
        deps)
            check_go
            download_dependencies
            ;;
        hooks)
            check_git
            setup_git_hooks
            ;;
        vscode)
            setup_vscode
            ;;
        verify)
            verify_setup
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: ${command}"
            show_help
            exit 1
            ;;
    esac
}

main "$@"

#!/usr/bin/env bash
# ==============================================================================
# TFO-GO-MCP Validation Script
# Version: 1.1.2
# Description: Validate TelemetryFlow GO MCP Server configuration and environment
# ==============================================================================

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# ==============================================================================
# Functions
# ==============================================================================

print_header() {
    echo -e "${PURPLE}"
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║       TelemetryFlow GO MCP Server - Validation Script         ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
}

validate_go() {
    log_info "Validating Go installation..."

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        return 1
    fi

    local go_version=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    local major=$(echo "${go_version}" | cut -d. -f1)
    local minor=$(echo "${go_version}" | cut -d. -f2)

    if [[ ${major} -lt 1 ]] || [[ ${major} -eq 1 && ${minor} -lt 24 ]]; then
        log_error "Go version ${go_version} is too old. Requires 1.24+"
        return 1
    fi

    log_success "Go version ${go_version}"
}

validate_dependencies() {
    log_info "Validating Go dependencies..."

    if ! go mod verify 2>/dev/null; then
        log_error "Go module verification failed"
        return 1
    fi

    log_success "Go dependencies verified"
}

validate_build() {
    log_info "Validating build..."

    if ! go build -o /dev/null ./cmd/mcp 2>/dev/null; then
        log_error "Build failed"
        return 1
    fi

    log_success "Build successful"
}

validate_tests() {
    log_info "Validating tests compile..."

    if ! go test -c -o /dev/null ./... 2>/dev/null; then
        log_error "Test compilation failed"
        return 1
    fi

    log_success "Tests compile successfully"
}

validate_config_file() {
    log_info "Validating configuration file..."

    local config_file="${1:-configs/tfo-mcp.yaml}"

    if [[ ! -f "${config_file}" ]]; then
        log_warning "Configuration file not found: ${config_file}"
        return 0
    fi

    # Check YAML syntax
    if command -v python3 &> /dev/null; then
        if ! python3 -c "import yaml; yaml.safe_load(open('${config_file}'))" 2>/dev/null; then
            log_error "Invalid YAML syntax in ${config_file}"
            return 1
        fi
    elif command -v yq &> /dev/null; then
        if ! yq eval '.' "${config_file}" >/dev/null 2>&1; then
            log_error "Invalid YAML syntax in ${config_file}"
            return 1
        fi
    else
        log_warning "No YAML validator available (python3 or yq)"
    fi

    log_success "Configuration file valid: ${config_file}"
}

validate_env_vars() {
    log_info "Validating environment variables..."

    local warnings=0

    if [[ -z "${TELEMETRYFLOW_MCP_CLAUDE_API_KEY:-}" ]]; then
        log_warning "TELEMETRYFLOW_MCP_CLAUDE_API_KEY is not set"
        warnings=$((warnings + 1))
    else
        log_success "TELEMETRYFLOW_MCP_CLAUDE_API_KEY is set"
    fi

    if [[ ${warnings} -gt 0 ]]; then
        log_warning "${warnings} environment variable(s) not set"
    fi
}

validate_api_key() {
    log_info "Validating Claude API key format..."

    if [[ -z "${TELEMETRYFLOW_MCP_CLAUDE_API_KEY:-}" ]]; then
        log_warning "TELEMETRYFLOW_MCP_CLAUDE_API_KEY not set, skipping validation"
        return 0
    fi

    if [[ ! "${TELEMETRYFLOW_MCP_CLAUDE_API_KEY}" =~ ^sk-ant-api[0-9]+-[A-Za-z0-9_-]+$ ]]; then
        log_warning "TELEMETRYFLOW_MCP_CLAUDE_API_KEY format may be invalid"
        return 0
    fi

    log_success "API key format valid"
}

validate_api_connectivity() {
    log_info "Validating Claude API connectivity..."

    if [[ -z "${TELEMETRYFLOW_MCP_CLAUDE_API_KEY:-}" ]]; then
        log_warning "TELEMETRYFLOW_MCP_CLAUDE_API_KEY not set, skipping API test"
        return 0
    fi

    if ! command -v curl &> /dev/null; then
        log_warning "curl not available, skipping API test"
        return 0
    fi

    local response=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "x-api-key: ${TELEMETRYFLOW_MCP_CLAUDE_API_KEY}" \
        -H "anthropic-version: 2023-06-01" \
        "https://api.anthropic.com/v1/messages" \
        --max-time 10 2>/dev/null || echo "000")

    case "${response}" in
        "401")
            log_error "API key is invalid (401 Unauthorized)"
            return 1
            ;;
        "403")
            log_error "API key lacks permissions (403 Forbidden)"
            return 1
            ;;
        "405"|"400")
            log_success "API connectivity verified (key valid)"
            ;;
        "000")
            log_warning "Could not connect to API (network issue?)"
            ;;
        *)
            log_warning "Unexpected API response: ${response}"
            ;;
    esac
}

validate_directory_structure() {
    log_info "Validating directory structure..."

    local required_dirs=(
        "cmd/mcp"
        "internal/domain"
        "internal/application"
        "internal/infrastructure"
        "internal/presentation"
        "pkg"
        "configs"
        "docs"
        "scripts"
    )

    local missing=0

    for dir in "${required_dirs[@]}"; do
        if [[ ! -d "${dir}" ]]; then
            log_error "Missing directory: ${dir}"
            missing=$((missing + 1))
        fi
    done

    if [[ ${missing} -gt 0 ]]; then
        return 1
    fi

    log_success "Directory structure valid"
}

validate_required_files() {
    log_info "Validating required files..."

    local required_files=(
        "go.mod"
        "go.sum"
        "Makefile"
        "Dockerfile"
        "README.md"
        ".gitignore"
    )

    local missing=0

    for file in "${required_files[@]}"; do
        if [[ ! -f "${file}" ]]; then
            log_error "Missing file: ${file}"
            missing=$((missing + 1))
        fi
    done

    if [[ ${missing} -gt 0 ]]; then
        return 1
    fi

    log_success "Required files present"
}

validate_formatting() {
    log_info "Validating code formatting..."

    local unformatted=$(gofmt -l . 2>&1 | grep -v vendor | head -10 || true)

    if [[ -n "${unformatted}" ]]; then
        log_error "Files not formatted:"
        echo "${unformatted}"
        return 1
    fi

    log_success "Code formatting valid"
}

validate_vet() {
    log_info "Running go vet..."

    if ! go vet ./... 2>&1; then
        log_error "go vet found issues"
        return 1
    fi

    log_success "go vet passed"
}

validate_lint() {
    log_info "Running linter..."

    if ! command -v golangci-lint &> /dev/null; then
        log_warning "golangci-lint not installed, skipping"
        return 0
    fi

    if ! golangci-lint run ./... 2>&1; then
        log_error "Linter found issues"
        return 1
    fi

    log_success "Linter passed"
}

print_summary() {
    local passed=$1
    local failed=$2
    local warnings=$3

    echo ""
    echo "════════════════════════════════════════════════════════════════════════"
    echo ""

    if [[ ${failed} -eq 0 ]]; then
        echo -e "${GREEN}Validation Complete: ${passed} passed, ${warnings} warnings${NC}"
    else
        echo -e "${RED}Validation Failed: ${passed} passed, ${failed} failed, ${warnings} warnings${NC}"
    fi

    echo ""
}

show_help() {
    echo "TFO-GO-MCP Validation Script"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  all           Run all validations (default)"
    echo "  quick         Quick validation (build, deps)"
    echo "  env           Validate environment only"
    echo "  config        Validate configuration only"
    echo "  code          Validate code quality only"
    echo "  api           Validate API connectivity"
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
    local passed=0
    local failed=0
    local warnings=0

    run_check() {
        if "$@"; then
            passed=$((passed + 1))
        else
            failed=$((failed + 1))
        fi
    }

    case "${command}" in
        all)
            run_check validate_go
            run_check validate_directory_structure
            run_check validate_required_files
            run_check validate_dependencies
            run_check validate_build
            run_check validate_tests
            run_check validate_config_file
            run_check validate_env_vars
            run_check validate_api_key
            run_check validate_formatting
            run_check validate_vet
            run_check validate_lint
            ;;
        quick)
            run_check validate_go
            run_check validate_dependencies
            run_check validate_build
            ;;
        env)
            run_check validate_env_vars
            run_check validate_api_key
            run_check validate_api_connectivity
            ;;
        config)
            run_check validate_config_file
            ;;
        code)
            run_check validate_formatting
            run_check validate_vet
            run_check validate_lint
            ;;
        api)
            run_check validate_api_connectivity
            ;;
        help|--help|-h)
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown command: ${command}"
            show_help
            exit 1
            ;;
    esac

    print_summary ${passed} ${failed} ${warnings}

    if [[ ${failed} -gt 0 ]]; then
        exit 1
    fi
}

main "$@"

#!/usr/bin/env bash
# ==============================================================================
# TFO-MCP Test Script
# Version: 1.1.2
# Description: Run tests for TelemetryFlow GO MCP Server
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

# Configuration
BUILD_DIR="${BUILD_DIR:-build}"
COVERAGE_DIR="${BUILD_DIR}/coverage"
MIN_COVERAGE="${MIN_COVERAGE:-80}"

# ==============================================================================
# Functions
# ==============================================================================

print_header() {
    echo -e "${PURPLE}"
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║           TelemetryFlow GO MCP Server - Test Script           ║"
    echo "╚═══════════════════════════════════════════════════════════════╝"
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

check_dependencies() {
    log_info "Checking dependencies..."

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi

    log_success "Dependencies check passed"
}

run_unit_tests() {
    log_info "Running unit tests..."

    go test -v -race -short ./...

    log_success "Unit tests passed"
}

run_integration_tests() {
    log_info "Running integration tests..."

    go test -v -race -run Integration ./...

    log_success "Integration tests passed"
}

run_all_tests() {
    log_info "Running all tests..."

    go test -v -race ./...

    log_success "All tests passed"
}

run_tests_with_coverage() {
    log_info "Running tests with coverage..."

    mkdir -p "${COVERAGE_DIR}"

    # Run tests with coverage
    go test -v -race -coverprofile="${COVERAGE_DIR}/coverage.out" -covermode=atomic ./...

    # Generate HTML report
    go tool cover -html="${COVERAGE_DIR}/coverage.out" -o "${COVERAGE_DIR}/coverage.html"

    # Generate function coverage
    go tool cover -func="${COVERAGE_DIR}/coverage.out" | tee "${COVERAGE_DIR}/coverage.txt"

    # Get total coverage
    total_coverage=$(go tool cover -func="${COVERAGE_DIR}/coverage.out" | grep total | awk '{print $3}' | sed 's/%//')

    log_success "Coverage: ${total_coverage}%"
    log_info "HTML report: ${COVERAGE_DIR}/coverage.html"

    # Check minimum coverage
    if (( $(echo "${total_coverage} < ${MIN_COVERAGE}" | bc -l) )); then
        log_error "Coverage ${total_coverage}% is below minimum ${MIN_COVERAGE}%"
        exit 1
    fi

    log_success "Coverage meets minimum requirement (${MIN_COVERAGE}%)"
}

run_benchmarks() {
    log_info "Running benchmarks..."

    mkdir -p "${BUILD_DIR}"

    go test -bench=. -benchmem -run=^$ ./... | tee "${BUILD_DIR}/benchmarks.txt"

    log_success "Benchmarks complete"
    log_info "Results: ${BUILD_DIR}/benchmarks.txt"
}

run_race_check() {
    log_info "Running race condition check..."

    go test -race -short ./...

    log_success "Race check passed"
}

run_vet() {
    log_info "Running go vet..."

    go vet ./...

    log_success "Vet passed"
}

run_staticcheck() {
    log_info "Running staticcheck..."

    if command -v staticcheck &> /dev/null; then
        staticcheck ./...
        log_success "Staticcheck passed"
    else
        log_warning "staticcheck not installed, skipping"
        echo "Install with: go install honnef.co/go/tools/cmd/staticcheck@latest"
    fi
}

run_lint() {
    log_info "Running golangci-lint..."

    if command -v golangci-lint &> /dev/null; then
        golangci-lint run ./...
        log_success "Lint passed"
    else
        log_warning "golangci-lint not installed, skipping"
        echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    fi
}

generate_report() {
    log_info "Generating test report..."

    mkdir -p "${BUILD_DIR}"

    # Run tests with JSON output
    go test -v -json ./... 2>&1 | tee "${BUILD_DIR}/test-results.json"

    log_success "Test report: ${BUILD_DIR}/test-results.json"
}

clean() {
    log_info "Cleaning test artifacts..."

    rm -rf "${BUILD_DIR}/coverage"
    rm -f "${BUILD_DIR}/test-results.json"
    rm -f "${BUILD_DIR}/benchmarks.txt"
    go clean -testcache

    log_success "Clean complete"
}

show_help() {
    echo "TFO-MCP Test Script"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  unit          Run unit tests only (default)"
    echo "  integration   Run integration tests only"
    echo "  all           Run all tests"
    echo "  coverage      Run tests with coverage report"
    echo "  bench         Run benchmarks"
    echo "  race          Run race condition check"
    echo "  vet           Run go vet"
    echo "  staticcheck   Run staticcheck"
    echo "  lint          Run golangci-lint"
    echo "  report        Generate test report"
    echo "  ci            Run full CI test suite"
    echo "  clean         Clean test artifacts"
    echo "  help          Show this help message"
    echo ""
    echo "Options:"
    echo "  MIN_COVERAGE  Minimum coverage percentage (default: ${MIN_COVERAGE})"
    echo "  BUILD_DIR     Build directory (default: ${BUILD_DIR})"
    echo ""
    echo "Examples:"
    echo "  $0 unit                    # Run unit tests"
    echo "  $0 coverage                # Run with coverage"
    echo "  MIN_COVERAGE=90 $0 coverage # Require 90% coverage"
    echo "  $0 ci                      # Run full CI suite"
}

run_ci() {
    log_info "Running CI test suite..."

    check_dependencies
    run_vet
    run_lint
    run_tests_with_coverage
    run_race_check

    log_success "CI test suite complete"
}

# ==============================================================================
# Main
# ==============================================================================

main() {
    print_header

    # Change to project root
    cd "$(dirname "$0")/.."

    local command="${1:-unit}"

    case "${command}" in
        unit)
            check_dependencies
            run_unit_tests
            ;;
        integration)
            check_dependencies
            run_integration_tests
            ;;
        all)
            check_dependencies
            run_all_tests
            ;;
        coverage)
            check_dependencies
            run_tests_with_coverage
            ;;
        bench|benchmark)
            check_dependencies
            run_benchmarks
            ;;
        race)
            check_dependencies
            run_race_check
            ;;
        vet)
            run_vet
            ;;
        staticcheck)
            run_staticcheck
            ;;
        lint)
            run_lint
            ;;
        report)
            check_dependencies
            generate_report
            ;;
        ci)
            run_ci
            ;;
        clean)
            clean
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

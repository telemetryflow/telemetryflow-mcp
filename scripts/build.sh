#!/usr/bin/env bash
# ==============================================================================
# TFO-MCP Build Script
# Version: 1.1.2
# Description: Build TelemetryFlow GO MCP Server for various platforms
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
BINARY_NAME="tfo-mcp"
VERSION="${VERSION:-1.1.2}"
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_DIR="${BUILD_DIR:-build}"
DIST_DIR="${DIST_DIR:-dist}"

# Build flags
LDFLAGS="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}"

# Supported platforms
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

# ==============================================================================
# Functions
# ==============================================================================

print_header() {
    echo -e "${PURPLE}"
    echo "╔══════════════════════════════════════════════════════════════════════╗"
    echo "║         TelemetryFlow GO MCP Server - Build Script                   ║"
    echo "║                       Version v${VERSION}                                ║"
    echo "╚══════════════════════════════════════════════════════════════════════╝"
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
        log_error "Go is not installed. Please install Go 1.24 or higher."
        exit 1
    fi

    GO_VERSION=$(go version | grep -oE 'go[0-9]+\.[0-9]+' | sed 's/go//')
    log_info "Go version: ${GO_VERSION}"

    if ! command -v git &> /dev/null; then
        log_warning "Git is not installed. Commit hash will be 'unknown'."
    fi

    log_success "Dependencies check passed"
}

download_deps() {
    log_info "Downloading Go dependencies..."
    go mod download
    go mod verify
    log_success "Dependencies downloaded"
}

build_local() {
    log_info "Building ${BINARY_NAME} for local platform..."

    mkdir -p "${BUILD_DIR}"

    CGO_ENABLED=0 go build \
        -ldflags "${LDFLAGS}" \
        -o "${BUILD_DIR}/${BINARY_NAME}" \
        ./cmd/mcp

    chmod +x "${BUILD_DIR}/${BINARY_NAME}"

    log_success "Built: ${BUILD_DIR}/${BINARY_NAME}"
    ls -lh "${BUILD_DIR}/${BINARY_NAME}"
}

build_platform() {
    local os=$1
    local arch=$2
    local output_name="${BINARY_NAME}-${os}-${arch}"

    if [[ "${os}" == "windows" ]]; then
        output_name="${output_name}.exe"
    fi

    log_info "Building for ${os}/${arch}..."

    GOOS=${os} GOARCH=${arch} CGO_ENABLED=0 go build \
        -ldflags "${LDFLAGS}" \
        -o "${DIST_DIR}/${output_name}" \
        ./cmd/mcp

    log_success "Built: ${DIST_DIR}/${output_name}"
}

build_all() {
    log_info "Building for all platforms..."

    mkdir -p "${DIST_DIR}"

    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r os arch <<< "${platform}"
        build_platform "${os}" "${arch}"
    done

    log_success "All platforms built successfully"
    echo ""
    log_info "Build artifacts:"
    ls -lh "${DIST_DIR}/"
}

build_docker() {
    log_info "Building Docker image..."

    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi

    docker build \
        --build-arg VERSION="${VERSION}" \
        --build-arg COMMIT="${COMMIT}" \
        --build-arg BUILD_DATE="${BUILD_DATE}" \
        -t "telemetryflow-mcp:${VERSION}" \
        -t "telemetryflow-mcp:latest" \
        .

    log_success "Docker image built: telemetryflow-mcp:${VERSION}"
}

create_checksums() {
    log_info "Creating checksums..."

    cd "${DIST_DIR}"

    if command -v sha256sum &> /dev/null; then
        sha256sum ${BINARY_NAME}-* > checksums.txt
    elif command -v shasum &> /dev/null; then
        shasum -a 256 ${BINARY_NAME}-* > checksums.txt
    else
        log_warning "No checksum tool found, skipping checksums"
        return
    fi

    cd - > /dev/null
    log_success "Checksums created: ${DIST_DIR}/checksums.txt"
}

create_archives() {
    log_info "Creating release archives..."

    mkdir -p "${DIST_DIR}/release"

    for file in "${DIST_DIR}"/${BINARY_NAME}-*; do
        if [[ -f "${file}" && ! "${file}" == *.txt ]]; then
            filename=$(basename "${file}")
            if [[ "${filename}" == *.exe ]]; then
                # Windows - create zip
                (cd "${DIST_DIR}" && zip -q "release/${filename%.exe}.zip" "${filename}")
            else
                # Unix - create tar.gz
                (cd "${DIST_DIR}" && tar -czf "release/${filename}.tar.gz" "${filename}")
            fi
            log_info "Created: ${DIST_DIR}/release/${filename}.*"
        fi
    done

    log_success "Release archives created in ${DIST_DIR}/release/"
}

clean() {
    log_info "Cleaning build artifacts..."
    rm -rf "${BUILD_DIR}" "${DIST_DIR}"
    go clean -cache
    log_success "Clean complete"
}

show_help() {
    echo "TFO-MCP Build Script"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  local       Build for local platform (default)"
    echo "  all         Build for all platforms"
    echo "  linux       Build for Linux (amd64, arm64)"
    echo "  darwin      Build for macOS (amd64, arm64)"
    echo "  windows     Build for Windows (amd64)"
    echo "  docker      Build Docker image"
    echo "  release     Create release with archives and checksums"
    echo "  clean       Clean build artifacts"
    echo "  help        Show this help message"
    echo ""
    echo "Options:"
    echo "  VERSION     Set version (default: ${VERSION})"
    echo "  BUILD_DIR   Set build directory (default: ${BUILD_DIR})"
    echo "  DIST_DIR    Set distribution directory (default: ${DIST_DIR})"
    echo ""
    echo "Examples:"
    echo "  $0 local                  # Build for current platform"
    echo "  $0 all                    # Build for all platforms"
    echo "  VERSION=1.2.0 $0 release  # Create release with version 1.2.0"
}

# ==============================================================================
# Main
# ==============================================================================

main() {
    print_header

    # Change to project root
    cd "$(dirname "$0")/.."

    local command="${1:-local}"

    case "${command}" in
        local)
            check_dependencies
            download_deps
            build_local
            ;;
        all)
            check_dependencies
            download_deps
            build_all
            ;;
        linux)
            check_dependencies
            download_deps
            mkdir -p "${DIST_DIR}"
            build_platform "linux" "amd64"
            build_platform "linux" "arm64"
            ;;
        darwin)
            check_dependencies
            download_deps
            mkdir -p "${DIST_DIR}"
            build_platform "darwin" "amd64"
            build_platform "darwin" "arm64"
            ;;
        windows)
            check_dependencies
            download_deps
            mkdir -p "${DIST_DIR}"
            build_platform "windows" "amd64"
            ;;
        docker)
            check_dependencies
            build_docker
            ;;
        release)
            check_dependencies
            download_deps
            clean
            build_all
            create_checksums
            create_archives
            log_success "Release ${VERSION} created successfully!"
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

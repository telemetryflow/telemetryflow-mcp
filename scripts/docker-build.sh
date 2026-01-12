#!/usr/bin/env bash
# ==============================================================================
# TFO-GO-MCP Docker Build Script
# Version: 1.1.2
# Description: Build and manage Docker images for TelemetryFlow GO MCP Server
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
IMAGE_NAME="${IMAGE_NAME:-telemetryflow-go-mcp}"
VERSION="${VERSION:-1.1.2}"
REGISTRY="${REGISTRY:-}"
PLATFORMS="${PLATFORMS:-linux/amd64,linux/arm64}"
DOCKERFILE="${DOCKERFILE:-Dockerfile}"
PUSH="${PUSH:-false}"

# ==============================================================================
# Functions
# ==============================================================================

print_header() {
    echo -e "${PURPLE}"
    echo "╔═══════════════════════════════════════════════════════════════╗"
    echo "║       TelemetryFlow GO MCP Server - Docker Build Script       ║"
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

    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed"
        exit 1
    fi

    # Check if Docker is running
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi

    log_success "Dependencies check passed"
}

get_image_tag() {
    local tag="${IMAGE_NAME}:${VERSION}"

    if [[ -n "${REGISTRY}" ]]; then
        tag="${REGISTRY}/${tag}"
    fi

    echo "${tag}"
}

build_image() {
    log_info "Building Docker image..."

    local tag=$(get_image_tag)
    local latest_tag="${IMAGE_NAME}:latest"

    if [[ -n "${REGISTRY}" ]]; then
        latest_tag="${REGISTRY}/${latest_tag}"
    fi

    local commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    local build_date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    docker build \
        --build-arg VERSION="${VERSION}" \
        --build-arg COMMIT="${commit}" \
        --build-arg BUILD_DATE="${build_date}" \
        -t "${tag}" \
        -t "${latest_tag}" \
        -f "${DOCKERFILE}" \
        .

    log_success "Built: ${tag}"
    log_success "Built: ${latest_tag}"
}

build_multiplatform() {
    log_info "Building multi-platform Docker image..."

    local tag=$(get_image_tag)
    local latest_tag="${IMAGE_NAME}:latest"

    if [[ -n "${REGISTRY}" ]]; then
        latest_tag="${REGISTRY}/${latest_tag}"
    fi

    local commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    local build_date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    # Check if buildx is available
    if ! docker buildx version &> /dev/null; then
        log_error "Docker buildx is not available"
        log_info "Install with: docker buildx install"
        exit 1
    fi

    # Create builder if it doesn't exist
    if ! docker buildx inspect tfo-mcp-builder &> /dev/null; then
        log_info "Creating buildx builder..."
        docker buildx create --name tfo-mcp-builder --use
    fi

    local push_flag=""
    if [[ "${PUSH}" == "true" ]]; then
        push_flag="--push"
    else
        push_flag="--load"
        # Can only load single platform
        PLATFORMS="linux/amd64"
        log_warning "Using --load requires single platform. Building for linux/amd64 only."
    fi

    docker buildx build \
        --platform "${PLATFORMS}" \
        --build-arg VERSION="${VERSION}" \
        --build-arg COMMIT="${commit}" \
        --build-arg BUILD_DATE="${build_date}" \
        -t "${tag}" \
        -t "${latest_tag}" \
        -f "${DOCKERFILE}" \
        ${push_flag} \
        .

    log_success "Built multi-platform image: ${tag}"
}

push_image() {
    log_info "Pushing Docker image..."

    local tag=$(get_image_tag)
    local latest_tag="${IMAGE_NAME}:latest"

    if [[ -n "${REGISTRY}" ]]; then
        latest_tag="${REGISTRY}/${latest_tag}"
    fi

    docker push "${tag}"
    docker push "${latest_tag}"

    log_success "Pushed: ${tag}"
    log_success "Pushed: ${latest_tag}"
}

run_container() {
    log_info "Running Docker container..."

    local tag=$(get_image_tag)

    docker run --rm -it \
        -e TELEMETRYFLOW_MCP_CLAUDE_API_KEY="${TELEMETRYFLOW_MCP_CLAUDE_API_KEY:-}" \
        -e TELEMETRYFLOW_MCP_LOG_LEVEL="${TELEMETRYFLOW_MCP_LOG_LEVEL:-info}" \
        "${tag}" \
        "$@"
}

scan_image() {
    log_info "Scanning Docker image for vulnerabilities..."

    local tag=$(get_image_tag)

    if command -v trivy &> /dev/null; then
        trivy image "${tag}"
    elif command -v grype &> /dev/null; then
        grype "${tag}"
    else
        log_warning "No vulnerability scanner found"
        log_info "Install trivy: https://github.com/aquasecurity/trivy"
        log_info "Install grype: https://github.com/anchore/grype"
    fi
}

show_image_info() {
    log_info "Image information:"

    local tag=$(get_image_tag)

    echo ""
    echo "Image: ${tag}"
    echo ""

    docker images "${IMAGE_NAME}"

    echo ""
    echo "Image layers:"
    docker history "${tag}" --no-trunc

    echo ""
    echo "Image size:"
    docker images "${tag}" --format "{{.Size}}"
}

clean() {
    log_info "Cleaning Docker artifacts..."

    # Remove images
    docker images "${IMAGE_NAME}" -q | xargs -r docker rmi -f 2>/dev/null || true

    # Remove dangling images
    docker image prune -f

    log_success "Clean complete"
}

show_help() {
    echo "TFO-GO-MCP Docker Build Script"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  build         Build Docker image (default)"
    echo "  multi         Build multi-platform image"
    echo "  push          Push image to registry"
    echo "  run           Run container"
    echo "  scan          Scan image for vulnerabilities"
    echo "  info          Show image information"
    echo "  clean         Clean Docker artifacts"
    echo "  help          Show this help message"
    echo ""
    echo "Options:"
    echo "  IMAGE_NAME    Image name (default: ${IMAGE_NAME})"
    echo "  VERSION       Image version (default: ${VERSION})"
    echo "  REGISTRY      Docker registry (default: none)"
    echo "  PLATFORMS     Target platforms (default: ${PLATFORMS})"
    echo "  DOCKERFILE    Dockerfile path (default: ${DOCKERFILE})"
    echo "  PUSH          Push after build (default: ${PUSH})"
    echo ""
    echo "Examples:"
    echo "  $0 build                        # Build image"
    echo "  VERSION=1.2.0 $0 build          # Build with version"
    echo "  REGISTRY=ghcr.io/user $0 build  # Build with registry"
    echo "  PUSH=true $0 multi              # Build and push multi-platform"
    echo "  $0 run                          # Run container"
    echo "  $0 run version                  # Run with command"
}

# ==============================================================================
# Main
# ==============================================================================

main() {
    print_header

    # Change to project root
    cd "$(dirname "$0")/.."

    local command="${1:-build}"
    shift || true

    case "${command}" in
        build)
            check_dependencies
            build_image
            ;;
        multi|multiplatform)
            check_dependencies
            build_multiplatform
            ;;
        push)
            check_dependencies
            push_image
            ;;
        run)
            check_dependencies
            run_container "$@"
            ;;
        scan)
            check_dependencies
            scan_image
            ;;
        info)
            check_dependencies
            show_image_info
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

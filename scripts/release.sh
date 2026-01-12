#!/usr/bin/env bash
# ==============================================================================
# TFO-GO-MCP Release Script
# Version: 1.1.2
# Description: Create releases for TelemetryFlow GO MCP Server
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
VERSION="${VERSION:-}"
DIST_DIR="${DIST_DIR:-dist}"
RELEASE_DIR="${DIST_DIR}/release"
GITHUB_REPO="telemetryflow/telemetryflow-go-mcp"
DRY_RUN="${DRY_RUN:-false}"

# ==============================================================================
# Functions
# ==============================================================================

print_header() {
    echo -e "${PURPLE}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║         TelemetryFlow GO MCP Server - Release Script         ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
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

    if ! command -v git &> /dev/null; then
        log_error "Git is not installed"
        exit 1
    fi

    if ! command -v gh &> /dev/null; then
        log_warning "GitHub CLI (gh) is not installed. GitHub release will be skipped."
    fi

    log_success "Dependencies check passed"
}

get_version() {
    if [[ -z "${VERSION}" ]]; then
        # Try to get version from git tag
        VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

        if [[ -z "${VERSION}" ]]; then
            log_error "VERSION not specified and no git tag found"
            log_info "Usage: VERSION=1.2.0 $0"
            exit 1
        fi
    fi

    # Remove 'v' prefix if present
    VERSION="${VERSION#v}"

    log_info "Release version: ${VERSION}"
}

validate_version() {
    log_info "Validating version format..."

    if ! [[ "${VERSION}" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
        log_error "Invalid version format: ${VERSION}"
        log_info "Expected format: MAJOR.MINOR.PATCH[-PRERELEASE]"
        exit 1
    fi

    log_success "Version format valid"
}

check_git_status() {
    log_info "Checking git status..."

    if [[ -n "$(git status --porcelain)" ]]; then
        log_error "Working directory is not clean"
        log_info "Please commit or stash your changes"
        exit 1
    fi

    log_success "Git status clean"
}

run_tests() {
    log_info "Running tests..."

    go test -v -race ./...

    log_success "Tests passed"
}

build_binaries() {
    log_info "Building binaries..."

    ./scripts/build.sh release

    log_success "Binaries built"
}

create_checksums() {
    log_info "Creating checksums..."

    cd "${RELEASE_DIR}"

    if command -v sha256sum &> /dev/null; then
        sha256sum *.tar.gz *.zip 2>/dev/null > checksums.txt || true
    elif command -v shasum &> /dev/null; then
        shasum -a 256 *.tar.gz *.zip 2>/dev/null > checksums.txt || true
    fi

    cd - > /dev/null

    log_success "Checksums created"
}

generate_changelog() {
    log_info "Generating changelog for ${VERSION}..."

    local changelog_file="${RELEASE_DIR}/CHANGELOG-${VERSION}.md"

    # Get commits since last tag
    local prev_tag=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")

    echo "# Release ${VERSION}" > "${changelog_file}"
    echo "" >> "${changelog_file}"
    echo "## Changes" >> "${changelog_file}"
    echo "" >> "${changelog_file}"

    if [[ -n "${prev_tag}" ]]; then
        git log "${prev_tag}..HEAD" --pretty=format:"- %s (%h)" >> "${changelog_file}"
    else
        git log --pretty=format:"- %s (%h)" -20 >> "${changelog_file}"
    fi

    echo "" >> "${changelog_file}"
    echo "" >> "${changelog_file}"
    echo "## Checksums" >> "${changelog_file}"
    echo "" >> "${changelog_file}"
    echo '```' >> "${changelog_file}"
    cat "${RELEASE_DIR}/checksums.txt" >> "${changelog_file}"
    echo '```' >> "${changelog_file}"

    log_success "Changelog generated: ${changelog_file}"
}

create_git_tag() {
    log_info "Creating git tag v${VERSION}..."

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_warning "DRY_RUN: Would create tag v${VERSION}"
        return
    fi

    # Check if tag exists
    if git rev-parse "v${VERSION}" >/dev/null 2>&1; then
        log_warning "Tag v${VERSION} already exists"
        read -p "Delete and recreate? [y/N] " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git tag -d "v${VERSION}"
            git push origin ":refs/tags/v${VERSION}" 2>/dev/null || true
        else
            return
        fi
    fi

    git tag -a "v${VERSION}" -m "Release v${VERSION}"
    log_success "Tag created: v${VERSION}"
}

push_tag() {
    log_info "Pushing tag to remote..."

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_warning "DRY_RUN: Would push tag v${VERSION}"
        return
    fi

    git push origin "v${VERSION}"
    log_success "Tag pushed"
}

create_github_release() {
    log_info "Creating GitHub release..."

    if ! command -v gh &> /dev/null; then
        log_warning "GitHub CLI not installed, skipping GitHub release"
        return
    fi

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_warning "DRY_RUN: Would create GitHub release"
        return
    fi

    local release_notes="${RELEASE_DIR}/CHANGELOG-${VERSION}.md"

    gh release create "v${VERSION}" \
        --title "v${VERSION}" \
        --notes-file "${release_notes}" \
        "${RELEASE_DIR}"/*.tar.gz \
        "${RELEASE_DIR}"/*.zip \
        "${RELEASE_DIR}"/checksums.txt

    log_success "GitHub release created"
}

build_docker() {
    log_info "Building Docker image..."

    if ! command -v docker &> /dev/null; then
        log_warning "Docker not installed, skipping Docker build"
        return
    fi

    docker build \
        --build-arg VERSION="${VERSION}" \
        -t "telemetryflow-go-mcp:${VERSION}" \
        -t "telemetryflow-go-mcp:latest" \
        .

    log_success "Docker image built: telemetryflow-go-mcp:${VERSION}"
}

push_docker() {
    log_info "Pushing Docker image..."

    if [[ "${DRY_RUN}" == "true" ]]; then
        log_warning "DRY_RUN: Would push Docker image"
        return
    fi

    if ! command -v docker &> /dev/null; then
        log_warning "Docker not installed, skipping Docker push"
        return
    fi

    docker push "telemetryflow-go-mcp:${VERSION}"
    docker push "telemetryflow-go-mcp:latest"

    log_success "Docker image pushed"
}

clean() {
    log_info "Cleaning previous release artifacts..."

    rm -rf "${DIST_DIR}"

    log_success "Clean complete"
}

show_help() {
    echo "TFO-GO-MCP Release Script"
    echo ""
    echo "Usage: VERSION=x.y.z $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  full          Full release (build, tag, GitHub, Docker)"
    echo "  build         Build release artifacts only"
    echo "  tag           Create and push git tag"
    echo "  github        Create GitHub release"
    echo "  docker        Build and push Docker image"
    echo "  clean         Clean release artifacts"
    echo "  help          Show this help message"
    echo ""
    echo "Options:"
    echo "  VERSION       Release version (required)"
    echo "  DRY_RUN       Set to 'true' for dry run"
    echo ""
    echo "Examples:"
    echo "  VERSION=1.2.0 $0 build           # Build only"
    echo "  VERSION=1.2.0 $0 full            # Full release"
    echo "  VERSION=1.2.0 DRY_RUN=true $0 full  # Dry run"
}

print_summary() {
    echo ""
    echo -e "${GREEN}════════════════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}                    Release v${VERSION} Complete!                        ${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo "Release artifacts:"
    ls -lh "${RELEASE_DIR}/" 2>/dev/null || true
    echo ""
}

# ==============================================================================
# Main
# ==============================================================================

main() {
    print_header

    # Change to project root
    cd "$(dirname "$0")/.."

    local command="${1:-full}"

    case "${command}" in
        full)
            check_dependencies
            get_version
            validate_version
            check_git_status
            clean
            run_tests
            build_binaries
            create_checksums
            generate_changelog
            create_git_tag
            push_tag
            create_github_release
            build_docker
            push_docker
            print_summary
            ;;
        build)
            check_dependencies
            get_version
            validate_version
            clean
            build_binaries
            create_checksums
            generate_changelog
            print_summary
            ;;
        tag)
            check_dependencies
            get_version
            validate_version
            check_git_status
            create_git_tag
            push_tag
            ;;
        github)
            check_dependencies
            get_version
            validate_version
            create_github_release
            ;;
        docker)
            check_dependencies
            get_version
            validate_version
            build_docker
            push_docker
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

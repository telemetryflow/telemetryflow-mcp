#!/usr/bin/env bash
# ==============================================================================
# TFO-GO-MCP Installation Script
# Version: 1.1.2
# Description: Install TelemetryFlow GO MCP Server
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
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
CONFIG_DIR="${CONFIG_DIR:-$HOME/.tfo-mcp}"
GITHUB_REPO="telemetryflow/telemetryflow-go-mcp"
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download"

# ==============================================================================
# Functions
# ==============================================================================

print_header() {
    echo -e "${PURPLE}"
    echo "╔══════════════════════════════════════════════════════════════════════╗"
    echo "║       TelemetryFlow GO MCP Server - Installation Script              ║"
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

detect_platform() {
    local os=""
    local arch=""

    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        MINGW*|MSYS*|CYGWIN*) os="windows" ;;
        *)          log_error "Unsupported OS: $(uname -s)"; exit 1 ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        *)              log_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac

    echo "${os}/${arch}"
}

check_sudo() {
    if [[ "${INSTALL_DIR}" == "/usr/local/bin" || "${INSTALL_DIR}" == "/usr/bin" ]]; then
        if [[ $EUID -ne 0 ]]; then
            log_warning "Installation to ${INSTALL_DIR} requires sudo privileges"
            SUDO="sudo"
        else
            SUDO=""
        fi
    else
        SUDO=""
    fi
}

download_binary() {
    local platform=$1
    IFS='/' read -r os arch <<< "${platform}"

    local filename="${BINARY_NAME}-${os}-${arch}"
    if [[ "${os}" == "windows" ]]; then
        filename="${filename}.exe"
    fi

    local url="${DOWNLOAD_URL}/v${VERSION}/${filename}.tar.gz"
    local temp_dir=$(mktemp -d)

    log_info "Downloading ${BINARY_NAME} v${VERSION} for ${platform}..."
    log_info "URL: ${url}"

    # Try curl first, then wget
    if command -v curl &> /dev/null; then
        curl -fsSL "${url}" -o "${temp_dir}/${filename}.tar.gz" || {
            log_error "Failed to download binary"
            rm -rf "${temp_dir}"
            exit 1
        }
    elif command -v wget &> /dev/null; then
        wget -q "${url}" -O "${temp_dir}/${filename}.tar.gz" || {
            log_error "Failed to download binary"
            rm -rf "${temp_dir}"
            exit 1
        }
    else
        log_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi

    # Extract
    log_info "Extracting archive..."
    tar -xzf "${temp_dir}/${filename}.tar.gz" -C "${temp_dir}"

    echo "${temp_dir}/${filename}"
}

install_binary() {
    local binary_path=$1

    log_info "Installing to ${INSTALL_DIR}..."

    ${SUDO} mkdir -p "${INSTALL_DIR}"
    ${SUDO} cp "${binary_path}" "${INSTALL_DIR}/${BINARY_NAME}"
    ${SUDO} chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    log_success "Installed: ${INSTALL_DIR}/${BINARY_NAME}"
}

create_config() {
    log_info "Creating configuration directory..."

    mkdir -p "${CONFIG_DIR}"

    # Create default config if it doesn't exist
    if [[ ! -f "${CONFIG_DIR}/tfo-mcp.yaml" ]]; then
        cat > "${CONFIG_DIR}/tfo-mcp.yaml" << 'EOF'
# TFO-GO-MCP Configuration File
# See documentation for full configuration options

server:
  name: "tfo-mcp"
  version: "1.1.2"
  timeout: 30s

claude:
  # Set API key via environment variable: TELEMETRYFLOW_MCP_CLAUDE_API_KEY
  api_key: ""
  model: "claude-sonnet-4-20250514"
  max_tokens: 4096
  temperature: 0.7

mcp:
  protocol_version: "2024-11-05"
  capabilities:
    tools: true
    resources: true
    prompts: true
    logging: true

logging:
  level: "info"
  format: "json"

telemetry:
  enabled: false
EOF
        log_success "Created default configuration: ${CONFIG_DIR}/tfo-mcp.yaml"
    else
        log_warning "Configuration already exists: ${CONFIG_DIR}/tfo-mcp.yaml"
    fi
}

verify_installation() {
    log_info "Verifying installation..."

    if command -v "${BINARY_NAME}" &> /dev/null; then
        local installed_version=$("${BINARY_NAME}" version 2>/dev/null | grep -oE 'Version:\s+[0-9]+\.[0-9]+\.[0-9]+' | awk '{print $2}' || echo "unknown")
        log_success "${BINARY_NAME} installed successfully!"
        log_info "Version: ${installed_version}"
        log_info "Location: $(which ${BINARY_NAME})"
    elif [[ -f "${INSTALL_DIR}/${BINARY_NAME}" ]]; then
        log_success "${BINARY_NAME} installed to ${INSTALL_DIR}/${BINARY_NAME}"
        log_warning "You may need to add ${INSTALL_DIR} to your PATH"
        echo ""
        echo "Add this to your shell profile:"
        echo "  export PATH=\"\${PATH}:${INSTALL_DIR}\""
    else
        log_error "Installation verification failed"
        exit 1
    fi
}

install_from_source() {
    log_info "Installing from source..."

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.24 or higher."
        exit 1
    fi

    # Check if we're in the project directory
    if [[ ! -f "go.mod" ]]; then
        log_error "Not in project directory. Please run from the project root."
        exit 1
    fi

    log_info "Building..."
    go build -ldflags "-s -w -X main.version=${VERSION}" -o "${BINARY_NAME}" ./cmd/mcp

    log_info "Installing..."
    ${SUDO} mv "${BINARY_NAME}" "${INSTALL_DIR}/"
    ${SUDO} chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

    log_success "Installed from source: ${INSTALL_DIR}/${BINARY_NAME}"
}

uninstall() {
    log_info "Uninstalling ${BINARY_NAME}..."

    check_sudo

    if [[ -f "${INSTALL_DIR}/${BINARY_NAME}" ]]; then
        ${SUDO} rm -f "${INSTALL_DIR}/${BINARY_NAME}"
        log_success "Removed: ${INSTALL_DIR}/${BINARY_NAME}"
    else
        log_warning "Binary not found at ${INSTALL_DIR}/${BINARY_NAME}"
    fi

    echo ""
    read -p "Remove configuration directory ${CONFIG_DIR}? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf "${CONFIG_DIR}"
        log_success "Removed: ${CONFIG_DIR}"
    else
        log_info "Configuration preserved at ${CONFIG_DIR}"
    fi

    log_success "Uninstallation complete"
}

show_help() {
    echo "TFO-GO-MCP Installation Script"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  install     Download and install (default)"
    echo "  source      Build and install from source"
    echo "  uninstall   Remove installation"
    echo "  help        Show this help message"
    echo ""
    echo "Options:"
    echo "  VERSION       Version to install (default: ${VERSION})"
    echo "  INSTALL_DIR   Installation directory (default: ${INSTALL_DIR})"
    echo "  CONFIG_DIR    Configuration directory (default: ${CONFIG_DIR})"
    echo ""
    echo "Examples:"
    echo "  $0                          # Install latest version"
    echo "  VERSION=1.2.0 $0            # Install specific version"
    echo "  INSTALL_DIR=~/bin $0        # Install to custom directory"
    echo "  $0 source                   # Install from source"
    echo "  $0 uninstall                # Uninstall"
}

print_post_install() {
    echo ""
    echo -e "${GREEN}════════════════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}                    Installation Complete!                              ${NC}"
    echo -e "${GREEN}════════════════════════════════════════════════════════════════════════${NC}"
    echo ""
    echo "Next steps:"
    echo ""
    echo "  1. Set your Claude API key:"
    echo "     export TELEMETRYFLOW_MCP_CLAUDE_API_KEY=\"your-api-key\""
    echo ""
    echo "  2. Edit configuration (optional):"
    echo "     ${CONFIG_DIR}/tfo-mcp.yaml"
    echo ""
    echo "  3. Run the server:"
    echo "     tfo-mcp run"
    echo ""
    echo "  4. Validate configuration:"
    echo "     tfo-mcp validate"
    echo ""
    echo "For more information, see the documentation:"
    echo "  https://github.com/${GITHUB_REPO}/tree/main/telemetryflow-go-mcp/docs"
    echo ""
}

# ==============================================================================
# Main
# ==============================================================================

main() {
    print_header

    local command="${1:-install}"

    case "${command}" in
        install)
            check_sudo
            platform=$(detect_platform)
            log_info "Detected platform: ${platform}"

            binary_path=$(download_binary "${platform}")
            install_binary "${binary_path}"
            create_config
            verify_installation

            # Cleanup
            rm -rf "$(dirname "${binary_path}")"

            print_post_install
            ;;
        source)
            check_sudo
            install_from_source
            create_config
            verify_installation
            print_post_install
            ;;
        uninstall)
            uninstall
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

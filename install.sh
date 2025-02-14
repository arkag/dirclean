#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print step with color
print_step() {
    echo -e "${GREEN}==>${NC} $1"
}

# Print error with color
print_error() {
    echo -e "${RED}Error:${NC} $1"
}

# Print warning with color
print_warning() {
    echo -e "${YELLOW}Warning:${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"
    
    case "${ARCH}" in
        x86_64) ARCH="amd64" ;;
        aarch64|arm64) ARCH="arm64" ;;
        *) print_error "Unsupported architecture: ${ARCH}"; exit 1 ;;
    esac
    
    case "${OS}" in
        linux|darwin) : ;;
        *) print_error "Unsupported operating system: ${OS}"; exit 1 ;;
    esac
}

# Get the latest release version
get_latest_version() {
    print_step "Fetching latest release version..."
    LATEST_VERSION=$(curl -sL https://api.github.com/repos/arkag/dirclean/releases/latest | grep '"tag_name":' | cut -d'"' -f4)
    if [ -z "${LATEST_VERSION}" ]; then
        print_error "Failed to fetch latest version"
        exit 1
    fi
}

# Download and verify the binary
install_binary() {
    local TEMP_DIR=$(mktemp -d)
    local BINARY_NAME="dirclean-${OS}-${ARCH}"
    local ARCHIVE_NAME="${BINARY_NAME}.tar.gz"
    local DOWNLOAD_URL="https://github.com/arkag/dirclean/releases/download/${LATEST_VERSION}/${ARCHIVE_NAME}"
    local CHECKSUM_URL="https://github.com/arkag/dirclean/releases/download/${LATEST_VERSION}/checksums.txt"
    
    trap 'rm -rf "${TEMP_DIR}"' EXIT
    
    print_step "Downloading ${ARCHIVE_NAME}..."
    curl -sL "${DOWNLOAD_URL}" -o "${TEMP_DIR}/${ARCHIVE_NAME}"
    
    print_step "Verifying checksum..."
    curl -sL "${CHECKSUM_URL}" -o "${TEMP_DIR}/checksums.txt"
    pushd "${TEMP_DIR}" > /dev/null
    
    # Calculate actual checksum
    local ACTUAL_CHECKSUM
    if command -v sha256sum >/dev/null 2>&1; then
        ACTUAL_CHECKSUM=$(sha256sum "${ARCHIVE_NAME}" | cut -d' ' -f1)
    elif command -v shasum >/dev/null 2>&1; then
        ACTUAL_CHECKSUM=$(shasum -a 256 "${ARCHIVE_NAME}" | cut -d' ' -f1)
    else
        print_error "Neither sha256sum nor shasum found"
        exit 1
    fi
    
    # Extract expected checksum for our file
    local EXPECTED_CHECKSUM=$(grep "${ARCHIVE_NAME}" checksums.txt | cut -d' ' -f1)
    if [ -z "${EXPECTED_CHECKSUM}" ]; then
        print_error "Checksum not found for ${ARCHIVE_NAME}"
        exit 1
    fi
    
    if [ "${EXPECTED_CHECKSUM}" != "${ACTUAL_CHECKSUM}" ]; then
        print_error "Checksum verification failed"
        print_error "Expected: ${EXPECTED_CHECKSUM}"
        print_error "Got: ${ACTUAL_CHECKSUM}"
        exit 1
    fi
    
    popd > /dev/null
    
    print_step "Extracting archive..."
    tar xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "${TEMP_DIR}"
    
    print_step "Installing binary and config..."
    if [ "$EUID" -eq 0 ]; then
        # Install binary
        mv "${TEMP_DIR}/dirclean" /usr/local/bin/
        chmod 755 /usr/local/bin/dirclean
        
        # Install config
        mkdir -p /usr/share/dirclean
        install -m 644 "${TEMP_DIR}/example.config.yaml" /usr/share/dirclean/
        
        if [ "${OS}" = "darwin" ]; then
            # Special handling for macOS
            if [ -d "/opt/homebrew" ]; then
                # Apple Silicon
                ln -sf /usr/share/dirclean/example.config.yaml /opt/homebrew/share/dirclean/example.config.yaml
            else
                # Intel Mac
                ln -sf /usr/share/dirclean/example.config.yaml /usr/local/share/dirclean/example.config.yaml
            fi
        fi
    else
        print_warning "Running without root privileges. Installing to user directories..."
        # Install binary
        mkdir -p ~/.local/bin
        mv "${TEMP_DIR}/dirclean" ~/.local/bin/
        chmod 755 ~/.local/bin/dirclean
        
        # Install config
        mkdir -p ~/.config/dirclean
        install -m 644 "${TEMP_DIR}/example.config.yaml" ~/.config/dirclean/
        
        if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
            print_warning "Please add ~/.local/bin to your PATH"
        fi
    fi
}

# Main installation process
main() {
    print_step "Installing DirClean..."
    
    # Check for required commands
    for cmd in curl tar sha256sum; do
        if ! command -v $cmd >/dev/null 2>&1; then
            print_error "Required command not found: $cmd"
            exit 1
        fi
    done
    
    detect_platform
    get_latest_version
    install_binary
    
    print_step "Installation complete!"
    echo -e "${GREEN}DirClean ${LATEST_VERSION} has been installed successfully!${NC}"
    
    if [ "$EUID" -ne 0 ]; then
        echo -e "\nTo use DirClean, either:"
        echo "1. Add ~/.local/bin to your PATH"
        echo "2. Run: export PATH=\$PATH:~/.local/bin"
    fi
}

main

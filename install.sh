#!/bin/bash


set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

REPO="Kaia-Alenia/Axon"
RELEASE_URL="https://github.com/$REPO/releases/latest/download"
VERSION=${AXON_VERSION:-latest}

detect_platform() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case "$OS" in
        linux)
            OS_NAME="linux"
            ;;
        darwin)
            OS_NAME="darwin"
            ;;
        mingw*|msys*|cygwin*)
            OS_NAME="windows"
            ;;
        *)
            echo -e "${RED}[ERROR] Unsupported OS: $OS${NC}"
            exit 1
            ;;
    esac
    
    case "$ARCH" in
        x86_64|amd64)
            ARCH_NAME="amd64"
            ;;
        aarch64|arm64)
            ARCH_NAME="arm64"
            ;;
        armv7|armv7l)
            ARCH_NAME="arm"
            ;;
        i386|i686)
            ARCH_NAME="386"
            ;;
        *)
            echo -e "${RED}[ERROR] Unsupported architecture: $ARCH${NC}"
            exit 1
            ;;
    esac
    
    echo -e "${BLUE}Detected platform: ${OS_NAME}-${ARCH_NAME}${NC}"
}

download_binary() {
    BINARY_NAME="axon-${OS_NAME}-${ARCH_NAME}"
    
    if [ "$OS_NAME" = "windows" ]; then
        BINARY_NAME="${BINARY_NAME}.exe"
    fi
    
    DOWNLOAD_URL="${RELEASE_URL}/${BINARY_NAME}"
    
    echo -e "${BLUE}[DOWNLOAD] Downloading from: $DOWNLOAD_URL${NC}"
    
    TEMP_FILE=$(mktemp)
    trap "rm -f $TEMP_FILE" EXIT
    
    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_FILE"
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$TEMP_FILE"
    else
        echo -e "${RED}[ERROR] Neither curl nor wget found. Please install one of them.${NC}"
        exit 1
    fi
    
    if [ ! -s "$TEMP_FILE" ]; then
        echo -e "${RED}[ERROR] Failed to download binary${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}[SUCCESS] Downloaded successfully${NC}"
}

install_binary() {
    case "$OS_NAME" in
        linux|darwin)
            INSTALL_DIR="/usr/local/bin"
            INSTALL_PATH="$INSTALL_DIR/axon"
            ;;
        windows)
            INSTALL_DIR="$PROGRAMFILES/Axon"
            INSTALL_PATH="$INSTALL_DIR/axon.exe"
            ;;
    esac
    
    echo -e "${BLUE}[INSTALL] Installing to: $INSTALL_PATH${NC}"
    
    case "$OS_NAME" in
        linux|darwin)
            if [ ! -d "$INSTALL_DIR" ]; then
                echo -e "${YELLOW}Creating directory: $INSTALL_DIR${NC}"
                sudo mkdir -p "$INSTALL_DIR"
            fi
            
            sudo cp "$TEMP_FILE" "$INSTALL_PATH"
            sudo chmod +x "$INSTALL_PATH"
            

            if [[ ":$PATH:" != *":/usr/local/bin:"* ]]; then
                echo -e "${YELLOW}[WARNING] /usr/local/bin is not in your PATH${NC}"
                echo -e "${YELLOW}Add it to your shell profile:${NC}"
                echo "export PATH=\"\$PATH:/usr/local/bin\""
            fi
            ;;
        windows)
            mkdir -p "$INSTALL_DIR"
            cp "$TEMP_FILE" "$INSTALL_PATH"
            

            if command -v setx &> /dev/null; then
                setx PATH "%PATH%;$INSTALL_DIR"
                echo -e "${YELLOW}[WARNING] Added $INSTALL_DIR to PATH. Restart your terminal.${NC}"
            fi
            ;;
    esac
    
    echo -e "${GREEN}[SUCCESS] Installation successful!${NC}"
}

verify_installation() {
    if command -v axon &> /dev/null; then
        echo -e "${GREEN}[SUCCESS] Axon is installed and available in PATH${NC}"
        INSTALLED_VERSION=$(axon -version 2>/dev/null || echo "unknown")
        echo -e "${BLUE}Version: $INSTALLED_VERSION${NC}"
        return 0
    else
        echo -e "${YELLOW}[WARNING] Axon was installed but not found in PATH${NC}"
        echo -e "${YELLOW}Try adding $INSTALL_DIR to your PATH manually${NC}"
        return 1
    fi
}

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║         Axon Server - Global Installation Script          ║"
echo "║                  Wireless Input Bridge                     ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

detect_platform
download_binary
install_binary
verify_installation

echo -e ""
echo -e "${GREEN}[DONE] Installation complete!${NC}"
echo -e "${BLUE}Run 'axon --help' to see available options${NC}"
echo -e "${BLUE}Run 'axon --version' to verify the installation${NC}"
echo -e ""

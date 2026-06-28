#!/bin/bash

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

VERSION="1.0.0"
BUILD_DIR="./build"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_header() {
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

cleanup() {
    rm -rf "$BUILD_DIR"
}

init_build() {
    mkdir -p "$BUILD_DIR"
    print_success "Build directory ready: $BUILD_DIR"
}

test_linux() {
    print_header "TESTING (Linux)"
    
    if command -v go &> /dev/null; then
        if go test -v ./...; then
            print_success "All tests passed"
        else
            print_error "Tests failed"
            return 1
        fi
    else
        print_error "Go not found"
        return 1
    fi
}

build_linux() {
    print_header "COMPILING FOR LINUX"
    
    GOOS=linux GOARCH=amd64 go build \
        -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$TIMESTAMP" \
        -o "$BUILD_DIR/axon-linux-amd64" . && \
        print_success "Linux amd64 compiled" || print_error "Linux amd64 failed"
    
    GOOS=linux GOARCH=arm64 go build \
        -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$TIMESTAMP" \
        -o "$BUILD_DIR/axon-linux-arm64" . && \
        print_success "Linux arm64 compiled" || print_error "Linux arm64 failed"
}

build_macos() {
    print_header "COMPILING FOR MACOS"
    
    if command -v clang &> /dev/null || [ "$OSTYPE" = "darwin"* ]; then
        GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build \
            -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$TIMESTAMP" \
            -o "$BUILD_DIR/axon-macos-amd64" . && \
            print_success "macOS amd64 compiled" || print_error "macOS amd64 failed"
        
        GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build \
            -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$TIMESTAMP" \
            -o "$BUILD_DIR/axon-macos-arm64" . && \
            print_success "macOS arm64 compiled" || print_error "macOS arm64 failed"
    else
        print_info "Clang not found, skipping macOS build (compile on macOS)"
        print_info "On macOS, run: GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -o axon-macos-arm64 ."
    fi
}

build_windows() {
    print_header "COMPILING FOR WINDOWS"
    
    if command -v x86_64-w64-mingw32-gcc &> /dev/null; then
        CC=x86_64-w64-mingw32-gcc GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build \
            -ldflags="-s -w -X main.Version=$VERSION -X main.BuildTime=$TIMESTAMP" \
            -o "$BUILD_DIR/axon-windows-amd64.exe" . && \
            print_success "Windows amd64 compiled" || print_error "Windows amd64 failed"
    else
        print_info "MinGW not found, skipping Windows build (compile on Windows)"
        print_info "On Windows, run: go build -o axon-windows-amd64.exe ."
    fi
}

check_syntax() {
    print_header "CHECKING SYNTAX"
    
    go build -n -o /dev/null . &> /dev/null && \
        print_success "Syntax check passed" || print_error "Syntax check failed"
}

list_builds() {
    print_header "COMPILED BINARIES"
    
    if [ -d "$BUILD_DIR" ]; then
        if [ "$(ls -A $BUILD_DIR)" ]; then
            ls -lh "$BUILD_DIR"
            echo ""
            print_success "Build complete"
        else
            print_error "No binaries found"
        fi
    else
        print_error "Build directory not found"
    fi
}

show_help() {
    cat << EOF
${BLUE}AXON Build System - Cross-Platform${NC}

Usage: $0 [command]

Commands:
    test              Run tests on Linux
    build             Compile for all platforms
    build-linux       Compile only for Linux
    build-macos       Compile only for macOS
    build-windows     Compile only for Windows
    check             Check syntax (without compiling)
    clean             Clean build directory
    help              Show this help message

Examples:
    $0 test
    $0 build
    $0 build-linux
    $0 clean

Environment variables:
    GOOS              Specify OS (linux, darwin, windows)
    GOARCH            Specify architecture (amd64, arm64)
    CGO_ENABLED       Enable CGo (1/0)
    CC                Specify C compiler
    
Supported platforms:
    ✓ Linux (amd64, arm64)
    ✓ macOS (amd64, arm64) - Requires Clang/Xcode
    ✓ Windows (amd64) - Requires MinGW or MSVC
    
EOF
}

main() {
    local cmd="${1:-help}"
    
    case "$cmd" in
        test)
            print_header "AXON Build System v$VERSION"
            test_linux
            ;;
        build)
            print_header "AXON Build System v$VERSION"
            check_syntax
            init_build
            build_linux
            build_macos
            build_windows
            list_builds
            ;;
        build-linux)
            check_syntax
            init_build
            build_linux
            list_builds
            ;;
        build-macos)
            check_syntax
            init_build
            build_macos
            list_builds
            ;;
        build-windows)
            check_syntax
            init_build
            build_windows
            list_builds
            ;;
        check)
            check_syntax
            ;;
        clean)
            print_header "CLEANING"
            cleanup
            print_success "Build directory cleaned"
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            print_error "Unknown command: $cmd"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"

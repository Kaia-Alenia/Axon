#!/bin/bash


set -e

BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

PROJECT_NAME="axon"
PACKAGE_PATH="./server"
OUTPUT_DIR="./dist"
VERSION=${VERSION:-"1.0.0"}
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

declare -a PLATFORMS=(
    "linux:amd64"
    "linux:arm64"
    "linux:arm"
    "darwin:amd64"
    "darwin:arm64"
    "windows:amd64"
)

get_binary_name() {
    local os=$1
    local arch=$2
    local name="axon-${os}-${arch}"
    
    if [ "$os" = "windows" ]; then
        echo "${name}.exe"
    else
        echo "$name"
    fi
}

build_platform() {
    local os=$1
    local arch=$2
    local binary_name=$(get_binary_name "$os" "$arch")
    local output_path="${OUTPUT_DIR}/${binary_name}"
    
    echo -e "${BLUE}Building ${GREEN}${binary_name}${NC}"
    

    LD_FLAGS="-X main.Version=${VERSION} -X main.CommitHash=${COMMIT_HASH} -X main.BuildDate=${BUILD_DATE}"
    
    GOOS=$os GOARCH=$arch CGO_ENABLED=1 go build \
        -ldflags "$LD_FLAGS" \
        -o "$output_path" \
        "$PACKAGE_PATH"
    
    if [ -f "$output_path" ]; then
        SIZE=$(du -h "$output_path" | cut -f1)
        echo -e "${GREEN}✓ Successfully built: ${binary_name} (${SIZE})${NC}"
        

        if [ "$os" != "windows" ]; then
            chmod +x "$output_path"
        fi
        
        echo "$binary_name" >> "${OUTPUT_DIR}/MANIFEST"
    else
        echo -e "${RED}✗ Failed to build ${binary_name}${NC}"
        return 1
    fi
}

echo -e "${BLUE}"
echo "╔═════════════════════════════════════════════════════════════╗"
echo "║           Axon Server - Multi-Platform Builder              ║"
echo "║                                                              ║"
echo "║ Version: ${VERSION}"
echo "║ Commit:  ${COMMIT_HASH}"
echo "║ Date:    ${BUILD_DATE}"
echo "╚═════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

mkdir -p "$OUTPUT_DIR"
rm -f "${OUTPUT_DIR}/MANIFEST"

if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go is not installed or not in PATH${NC}"
    echo -e "${YELLOW}Please install Go from https://golang.org/dl/${NC}"
    exit 1
fi

echo -e "${BLUE}Go version:${NC}"
go version
echo -e ""

FAILED=0
for platform in "${PLATFORMS[@]}"; do
    os="${platform%:*}"
    arch="${platform#*:}"
    
    if build_platform "$os" "$arch"; then
        :
    else
        FAILED=$((FAILED + 1))
    fi
done

echo -e ""
echo -e "${BLUE}════════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}Build Summary:${NC}"
echo -e "${BLUE}Output directory: ${OUTPUT_DIR}${NC}"
echo -e ""

if [ -f "${OUTPUT_DIR}/MANIFEST" ]; then
    echo -e "${GREEN}✓ Built binaries:${NC}"
    while IFS= read -r binary; do
        size=$(du -h "${OUTPUT_DIR}/${binary}" | cut -f1)
        echo -e "  • ${binary} (${size})"
    done < "${OUTPUT_DIR}/MANIFEST"
fi

echo -e ""
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All builds completed successfully!${NC}"
    echo -e "${YELLOW}Next: Create a GitHub Release and upload these files${NC}"
    exit 0
else
    echo -e "${RED}✗ ${FAILED} build(s) failed${NC}"
    exit 1
fi

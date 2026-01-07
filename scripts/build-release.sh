#!/usr/bin/env bash
#
# build-release.sh - Build and optionally sign/notarize macOS binaries
#
# This script builds Go binaries for the local platform and optionally
# signs and notarizes them for macOS distribution.
#
# Usage:
#   ./scripts/build-release.sh [OPTIONS]
#
# Options:
#   --version VERSION   Version string to embed (default: from git tag or "dev")
#   --sign              Sign binaries with Developer ID (macOS only)
#   --notarize          Notarize binaries with Apple (requires --sign)
#   --all-platforms     Build for all platforms (linux/amd64, darwin/arm64, darwin/amd64)
#   --output DIR        Output directory (default: ./bin)
#   --help              Show this help message
#
# Environment:
#   Uses 1Password CLI (op) for credential management when signing.
#   Requires OP_SERVICE_ACCOUNT_TOKEN or interactive 1Password auth.
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Defaults
VERSION=""
SIGN=false
NOTARIZE=false
ALL_PLATFORMS=false
OUTPUT_DIR="./bin"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Log functions
info() { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC} $*"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $*"; }
error() { echo -e "${RED}[ERROR]${NC} $*" >&2; }
die() { error "$*"; exit 1; }

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            VERSION="$2"
            shift 2
            ;;
        --sign)
            SIGN=true
            shift
            ;;
        --notarize)
            NOTARIZE=true
            shift
            ;;
        --all-platforms)
            ALL_PLATFORMS=true
            shift
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --help)
            head -30 "$0" | grep -E "^#" | sed 's/^# \?//'
            exit 0
            ;;
        *)
            die "Unknown option: $1"
            ;;
    esac
done

# Determine version
if [[ -z "$VERSION" ]]; then
    VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
fi
info "Building version: $VERSION"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Build function
build_binary() {
    local name=$1
    local goos=$2
    local goarch=$3
    local src_dir=$4

    local output_name="${name}-${goos}-${goarch}"
    [[ "$goos" == "windows" ]] && output_name="${output_name}.exe"

    info "Building $output_name..."

    (
        cd "$PROJECT_ROOT/$src_dir"
        GOOS="$goos" GOARCH="$goarch" go build \
            -ldflags="-s -w -X main.version=$VERSION" \
            -o "$PROJECT_ROOT/$OUTPUT_DIR/$output_name" .
    )

    success "Built $output_name"
}

# Build binaries
build_all() {
    local platforms=()

    if $ALL_PLATFORMS; then
        platforms=(
            "linux:amd64"
            "darwin:arm64"
            "darwin:amd64"
        )
    else
        # Build for current platform only
        local os=$(go env GOOS)
        local arch=$(go env GOARCH)
        platforms=("$os:$arch")
    fi

    for platform in "${platforms[@]}"; do
        IFS=':' read -r goos goarch <<< "$platform"
        build_binary "stig-processor" "$goos" "$goarch" "src/stig-processor"
        build_binary "schema-builder" "$goos" "$goarch" "src/schema-builder"
    done
}

# Sign binaries (macOS only)
sign_binaries() {
    if [[ "$(uname)" != "Darwin" ]]; then
        warn "Code signing is only available on macOS"
        return
    fi

    info "Loading signing identity from 1Password..."

    # Check for 1Password CLI
    if ! command -v op &>/dev/null; then
        die "1Password CLI (op) is required for signing. Install with: brew install --cask 1password-cli"
    fi

    # Get signing identity
    local identity
    identity=$(op read "op://Development/Apple Developer ID/identity" 2>/dev/null) || \
        die "Failed to read signing identity from 1Password"

    info "Signing with identity: $identity"

    for binary in "$OUTPUT_DIR"/*-darwin-*; do
        [[ -f "$binary" ]] || continue
        [[ "$binary" == *.sha256 ]] && continue

        info "Signing $(basename "$binary")..."
        codesign --force --options runtime --sign "$identity" "$binary"

        # Verify signature
        codesign --verify --verbose=2 "$binary" || warn "Signature verification failed"
        success "Signed $(basename "$binary")"
    done
}

# Notarize binaries (macOS only)
notarize_binaries() {
    if [[ "$(uname)" != "Darwin" ]]; then
        warn "Notarization is only available on macOS"
        return
    fi

    if ! $SIGN; then
        die "Notarization requires signed binaries (use --sign)"
    fi

    info "Loading notarization credentials from 1Password..."

    local apple_id team_id app_password
    apple_id=$(op read "op://Development/Apple Developer ID/apple_id" 2>/dev/null) || \
        die "Failed to read Apple ID from 1Password"
    team_id=$(op read "op://Development/Apple Developer ID/team_id" 2>/dev/null) || \
        die "Failed to read Team ID from 1Password"
    app_password=$(op read "op://Development/Apple Developer ID/app_password" 2>/dev/null) || \
        die "Failed to read app-specific password from 1Password"

    # Create ZIP archives for notarization
    for arch in arm64 amd64; do
        local zip_name="stig-tools-darwin-${arch}.zip"
        local binaries=()

        for binary in "$OUTPUT_DIR"/*-darwin-"$arch"; do
            [[ -f "$binary" ]] && binaries+=("$binary")
        done

        if [[ ${#binaries[@]} -eq 0 ]]; then
            warn "No darwin/$arch binaries found to notarize"
            continue
        fi

        info "Creating $zip_name..."
        (cd "$OUTPUT_DIR" && zip -j "$zip_name" "${binaries[@]##*/}")

        info "Submitting $zip_name for notarization..."
        xcrun notarytool submit "$OUTPUT_DIR/$zip_name" \
            --apple-id "$apple_id" \
            --team-id "$team_id" \
            --password "$app_password" \
            --wait

        success "Notarized $zip_name"
    done
}

# Create checksums
create_checksums() {
    info "Creating checksums..."

    for binary in "$OUTPUT_DIR"/*; do
        [[ -f "$binary" ]] || continue
        [[ "$binary" == *.sha256 ]] && continue
        [[ "$binary" == *.zip ]] && continue

        if command -v sha256sum &>/dev/null; then
            sha256sum "$binary" > "${binary}.sha256"
        else
            shasum -a 256 "$binary" > "${binary}.sha256"
        fi
    done

    success "Created checksums"
}

# Main execution
main() {
    info "Starting build process..."
    info "Output directory: $OUTPUT_DIR"

    # Check Go is installed
    command -v go &>/dev/null || die "Go is required but not installed"

    # Build
    build_all

    # Sign if requested
    if $SIGN; then
        sign_binaries
    fi

    # Notarize if requested
    if $NOTARIZE; then
        notarize_binaries
    fi

    # Create checksums
    create_checksums

    echo ""
    success "Build complete! Binaries in $OUTPUT_DIR:"
    ls -la "$OUTPUT_DIR"
}

main

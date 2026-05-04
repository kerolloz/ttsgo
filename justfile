#!/usr/bin/env -S just --justfile

set shell := ["bash", "-cu"]

# Initialize the project (install deps, etc)
init:
	cd packages/ttsgo && go mod tidy
	cd packages/nego && go mod tidy

# Build the binaries for the current platform
build:
	mkdir -p bin
	cd packages/ttsgo && go build -ldflags="-s -w" -o ../../bin/ttsgo ./cmd/ttsgo
	cd packages/nego && go build -ldflags="-s -w" -o ../../bin/nego .

# Cross-compile for all platforms (used by CI)
build-all version:
	#!/usr/bin/env bash
	set -euo pipefail
	LDFLAGS="-s -w -X main.version={{version}}"
	PLATFORMS=("darwin/arm64" "darwin/amd64" "linux/arm64" "linux/amd64" "windows/arm64" "windows/amd64")
	
	for PLATFORM in "${PLATFORMS[@]}"; do
		OS="${PLATFORM%/*}"
		ARCH="${PLATFORM#*/}"
		
		# Map GOOS/GOARCH to NPM naming conventions
		NPM_OS=$OS
		[[ "$OS" == "windows" ]] && NPM_OS="win32"
		NPM_ARCH=$ARCH
		[[ "$ARCH" == "amd64" ]] && NPM_ARCH="x64"
		
		EXT=""
		[[ "$OS" == "windows" ]] && EXT=".exe"
		
		echo "  Building ttsgo ($NPM_OS-$NPM_ARCH)..."
		mkdir -p packages/ttsgo-$NPM_OS-$NPM_ARCH/bin
		(cd packages/ttsgo && CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH \
			go build -ldflags="$LDFLAGS" \
			-o ../../packages/ttsgo-$NPM_OS-$NPM_ARCH/bin/ttsgo$EXT ./cmd/ttsgo)
		
		echo "  Building nego ($NPM_OS-$NPM_ARCH)..."
		mkdir -p packages/nego-$NPM_OS-$NPM_ARCH/bin
		(cd packages/nego && CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH \
			go build -ldflags="$LDFLAGS" \
			-o ../../packages/nego-$NPM_OS-$NPM_ARCH/bin/nego$EXT .)
		
		# Set executable bits for Unix targets
		if [[ "$OS" != "windows" ]]; then
			chmod +x packages/ttsgo-$NPM_OS-$NPM_ARCH/bin/ttsgo
			chmod +x packages/nego-$NPM_OS-$NPM_ARCH/bin/nego
		fi
	done
	echo "✓ All 12 binaries built."

# Clean build artifacts
clean:
	rm -rf bin
	find packages -name "bin" -type d -exec rm -rf {} + 2>/dev/null || true

#!/usr/bin/env -S just --justfile

set shell := ["bash", "-cu"]

# Initialize the project (install deps, etc)
init:
	go mod tidy
	cd packages/ttsgo && go mod tidy
	cd packages/nego && go mod tidy

# Build the binaries for the current platform
build:
	mkdir -p bin
	go build -ldflags="-s -w" -o bin/ttsgo ./packages/ttsgo/cmd/ttsgo
	go build -ldflags="-s -w" -o bin/nego ./packages/nego/main.go

# Cross-compile for all platforms (used by CI)
build-all version:
	#!/usr/bin/env bash
	set -e
	LDFLAGS="-s -w -X main.version={{version}}"
	PLATFORMS=("darwin/arm64" "darwin/amd64" "linux/arm64" "linux/amd64" "windows/arm64" "windows/amd64")
	
	for PLATFORM in "${PLATFORMS[@]}"; do
		OS="${PLATFORM%/*}"
		ARCH="${PLATFORM#*/}"
		
		# Map GOOS/GOARCH to NPM naming conventions
		NPM_OS=$OS
		if [ "$OS" == "windows" ]; then NPM_OS="win32"; fi
		NPM_ARCH=$ARCH
		if [ "$ARCH" == "amd64" ]; then NPM_ARCH="x64"; fi
		
		EXT=""
		if [ "$OS" == "windows" ]; then EXT=".exe"; fi
		
		echo "Building $OS-$ARCH..."
		
		# Build ttsgo
		mkdir -p packages/ttsgo-$NPM_OS-$NPM_ARCH/bin
		GOOS=$OS GOARCH=$ARCH go build -ldflags="$LDFLAGS" -o packages/ttsgo-$NPM_OS-$NPM_ARCH/bin/ttsgo$EXT ./packages/ttsgo/cmd/ttsgo
		
		# Build nego
		mkdir -p packages/nego-$NPM_OS-$NPM_ARCH/bin
		GOOS=$OS GOARCH=$ARCH go build -ldflags="$LDFLAGS" -o packages/nego-$NPM_OS-$NPM_ARCH/bin/nego$EXT ./packages/nego/main.go
		
		# Set executable bits
		if [ "$OS" != "windows" ]; then
			chmod +x packages/ttsgo-$NPM_OS-$NPM_ARCH/bin/ttsgo
			chmod +x packages/nego-$NPM_OS-$NPM_ARCH/bin/nego
		fi
	done

# Clean build artifacts
clean:
	rm -rf bin
	find packages -name "bin" -type d -exec rm -rf {} +

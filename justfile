#!/usr/bin/env -S just --justfile

set shell := ["bash", "-cu"]

# Initialize the project (tidy Go modules)
init:
	cd packages/ttsgo && go mod tidy
	cd packages/nego && go mod tidy

# Build both binaries for the current platform (used by CI + local dev)
build:
	mkdir -p bin
	cd packages/ttsgo && go build -ldflags="-s -w" -o ../../bin/ttsgo ./cmd/ttsgo
	cd packages/nego && go build -ldflags="-s -w" -o ../../bin/nego .

# Clean build artifacts
clean:
	rm -rf bin
	find packages -name "bin" -type d -exec rm -rf {} + 2>/dev/null || true

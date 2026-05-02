package compiler

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBinaryDiscoveryInNodeModules(t *testing.T) {
	dir := t.TempDir()
	binDir := filepath.Join(dir, "node_modules", ".bin")
	os.MkdirAll(binDir, 0755)
	tsgo := filepath.Join(binDir, "tsgo")
	os.WriteFile(tsgo, []byte("#!/bin/sh\necho tsgo"), 0755)

	c, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.BinaryPath() != tsgo {
		t.Errorf("BinaryPath = %q, want %q", c.BinaryPath(), tsgo)
	}
}

func TestBinaryDiscoveryWalksUp(t *testing.T) {
	parent := t.TempDir()
	binDir := filepath.Join(parent, "node_modules", ".bin")
	os.MkdirAll(binDir, 0755)
	tsgo := filepath.Join(binDir, "tsgo")
	os.WriteFile(tsgo, []byte("#!/bin/sh\necho tsgo"), 0755)

	// Call New from a subdirectory — should walk up and find parent's tsgo.
	sub := filepath.Join(parent, "packages", "app")
	os.MkdirAll(sub, 0755)

	c, err := New(sub)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.BinaryPath() != tsgo {
		t.Errorf("BinaryPath = %q, want %q", c.BinaryPath(), tsgo)
	}
}

func TestBinaryDiscoveryFallsBackToPath(t *testing.T) {
	dir := t.TempDir()
	// Create a fake tsgo on a temp PATH.
	fakeDir := t.TempDir()
	tsgo := filepath.Join(fakeDir, "tsgo")
	os.WriteFile(tsgo, []byte("#!/bin/sh\necho tsgo"), 0755)
	t.Setenv("PATH", fakeDir)

	c, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if c.BinaryPath() != tsgo {
		t.Errorf("BinaryPath = %q, want %q", c.BinaryPath(), tsgo)
	}
}

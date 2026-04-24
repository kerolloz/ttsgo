package compiler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Compiler defines the interface for running the TypeScript compiler.
type Compiler interface {
	Run(ctx context.Context, opts RunOptions) error
	Version() (string, error)
	BinaryPath() string
}

// TsGo wraps the tsgo binary.
type TsGo struct {
	binaryPath string
	cwd        string
}

// New locates the tsgo binary and returns a TsGo instance.
func New(cwd string) (*TsGo, error) {
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	// Walk up from cwd looking for node_modules/.bin/tsgo
	dir := cwd
	for {
		candidate := filepath.Join(dir, "node_modules", ".bin", "tsgo")
		if _, err := os.Stat(candidate); err == nil {
			return &TsGo{binaryPath: candidate, cwd: cwd}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fall back to PATH.
	p, err := exec.LookPath("tsgo")
	if err != nil {
		return nil, fmt.Errorf(
			"tsgo binary not found.\n"+
				"Install it with: npm install -D @typescript/native-preview\n"+
				"or ensure 'tsgo' is on your PATH.\n"+
				"Original error: %w", err,
		)
	}
	return &TsGo{binaryPath: p, cwd: cwd}, nil
}

// RunOptions contains options for a tsgo run.
type RunOptions struct {
	TsConfigPath string
	NoEmit       bool
	ExtraArgs    []string
}

// Run executes tsgo with the given options.
func (t *TsGo) Run(ctx context.Context, opts RunOptions) error {
	var args []string

	if opts.TsConfigPath != "" {
		args = append(args, "-p", opts.TsConfigPath)
	}
	if opts.NoEmit {
		args = append(args, "--noEmit")
	}
	args = append(args, opts.ExtraArgs...)

	cmd := exec.CommandContext(ctx, t.binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Dir = t.cwd

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("tsgo exited with code %d: %w", exitErr.ExitCode(), exitErr)
		}
		return fmt.Errorf("failed to run tsgo: %w", err)
	}
	return nil
}

// Version returns the version string of the located tsgo binary.
func (t *TsGo) Version() (string, error) {
	out, err := exec.Command(t.binaryPath, "--version").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// BinaryPath returns the resolved path to the tsgo binary.
func (t *TsGo) BinaryPath() string {
	return t.binaryPath
}

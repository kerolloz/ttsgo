package orchestrator

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"nego/internal/assets"
	"nego/internal/compiler"
	"nego/internal/config"
	"nego/internal/paths"
	"nego/internal/process"
)

// mockCompiler records calls and optionally writes a fake dist/src/main.js.
type mockCompiler struct {
	calls  atomic.Int32
	err    error
	outDir string // absolute path to write fake output into
}

func (m *mockCompiler) Run(_ context.Context, _ compiler.RunOptions) error {
	m.calls.Add(1)
	if m.err != nil {
		return m.err
	}
	if m.outDir != "" {
		dir := filepath.Join(m.outDir, "src")
		os.MkdirAll(dir, 0755)
		os.WriteFile(filepath.Join(dir, "main.js"), []byte(""), 0644)
	}
	return nil
}
func (m *mockCompiler) Version() (string, error) { return "mock", nil }
func (m *mockCompiler) BinaryPath() string        { return "/mock/tsgo" }

// mockRunner records Start/Kill calls.
type mockRunner struct {
	started atomic.Int32
	killed  atomic.Int32
}

func (m *mockRunner) Start() error { m.started.Add(1); return nil }
func (m *mockRunner) Kill()        { m.killed.Add(1) }
func (m *mockRunner) Wait() int    { return 0 }

func newTestOrchestrator(t *testing.T, comp compiler.Compiler, runner process.ProcessRunner) (*Orchestrator, string) {
	t.Helper()
	cwd := t.TempDir()

	nestCfg := &config.NestConfig{
		SourceRoot: "src",
		EntryFile:  "main",
		Exec:       "node",
	}
	tsCfg := &config.TsConfig{OutDir: "dist"}

	assetMgr, _ := assets.New(cwd, "src", "dist", nil)
	rewriter := paths.New(cwd, nil, "dist", "src")

	return &Orchestrator{
		Cwd:        cwd,
		NestConfig: nestCfg,
		TsConfig:   tsCfg,
		Compiler:   comp,
		Rewriter:   rewriter,
		Assets:     assetMgr,
		runner:     runner,
	}, cwd
}

func TestBuildPipeline(t *testing.T) {
	comp := &mockCompiler{}
	runner := &mockRunner{}
	o, cwd := newTestOrchestrator(t, comp, runner)
	comp.outDir = filepath.Join(cwd, "dist")

	if err := o.Build(context.Background(), false); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if comp.calls.Load() != 1 {
		t.Errorf("compiler called %d times, want 1", comp.calls.Load())
	}
	if runner.started.Load() != 1 {
		t.Errorf("runner started %d times, want 1", runner.started.Load())
	}
}

func TestBuildIsRebuild(t *testing.T) {
	comp := &mockCompiler{}
	runner := &mockRunner{}
	o, cwd := newTestOrchestrator(t, comp, runner)
	comp.outDir = filepath.Join(cwd, "dist")

	if err := o.Build(context.Background(), true); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if runner.killed.Load() != 1 {
		t.Errorf("runner killed %d times, want 1", runner.killed.Load())
	}
}

func TestBuildCompilerError(t *testing.T) {
	comp := &mockCompiler{err: errors.New("compile failed")}
	runner := &mockRunner{}
	o, _ := newTestOrchestrator(t, comp, runner)

	err := o.Build(context.Background(), false)
	if err == nil {
		t.Fatal("expected error from Build")
	}
	if runner.started.Load() != 0 {
		t.Error("runner should not start when compilation fails")
	}
}

func TestDeleteOutDirPathTraversal(t *testing.T) {
	comp := &mockCompiler{}
	o, _ := newTestOrchestrator(t, comp, nil)
	o.NestConfig.CompilerOptions.DeleteOutDir = true
	o.TsConfig.OutDir = "../../etc"

	err := o.Build(context.Background(), false)
	if err == nil {
		t.Fatal("expected error for outDir escaping project root")
	}
}

func TestWatchTriggersRebuild(t *testing.T) {
	comp := &mockCompiler{}
	o, cwd := newTestOrchestrator(t, comp, nil)
	comp.outDir = filepath.Join(cwd, "dist")

	srcDir := filepath.Join(cwd, "src")
	os.MkdirAll(srcDir, 0755)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watchDone := make(chan error, 1)
	go func() { watchDone <- o.Watch(ctx, false) }()

	// Wait for initial build.
	time.Sleep(200 * time.Millisecond)

	// Trigger a rebuild by writing a .ts file.
	os.WriteFile(filepath.Join(srcDir, "app.ts"), []byte("x"), 0644)

	// Wait for the rebuild debounce + build time.
	time.Sleep(600 * time.Millisecond)
	cancel()
	<-watchDone

	if comp.calls.Load() < 2 {
		t.Errorf("compiler called %d times, want ≥2 (initial + rebuild)", comp.calls.Load())
	}
}

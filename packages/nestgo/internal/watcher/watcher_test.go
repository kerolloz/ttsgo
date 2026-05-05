package watcher

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestDebounce(t *testing.T) {
	dir := t.TempDir()

	var count atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wt, err := New(ctx, dir, 80*time.Millisecond, func() { count.Add(1) })
	if err != nil {
		t.Fatal(err)
	}
	defer wt.Close()

	// Write several .ts files rapidly — should coalesce into one callback.
	for i := 0; i < 5; i++ {
		os.WriteFile(filepath.Join(dir, "file.ts"), []byte("x"), 0644)
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(200 * time.Millisecond)
	if n := count.Load(); n != 1 {
		t.Errorf("onChange called %d times, want 1", n)
	}
}

func TestNewDirTracking(t *testing.T) {
	dir := t.TempDir()

	fired := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wt, err := New(ctx, dir, 50*time.Millisecond, func() {
		select {
		case fired <- struct{}{}:
		default:
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer wt.Close()

	// Create a new subdirectory, then write a .ts file inside it.
	sub := filepath.Join(dir, "newpkg")
	os.Mkdir(sub, 0755)
	time.Sleep(50 * time.Millisecond) // let the watcher register the new dir
	os.WriteFile(filepath.Join(sub, "service.ts"), []byte("x"), 0644)

	select {
	case <-fired:
	case <-time.After(2 * time.Second):
		t.Error("onChange never fired after writing to new subdirectory")
	}
}

func TestDoubleClose(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wt, err := New(ctx, dir, 50*time.Millisecond, func() {})
	if err != nil {
		t.Fatal(err)
	}
	// Must not panic.
	wt.Close()
	wt.Close()
}

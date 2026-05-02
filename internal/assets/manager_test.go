package assets

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMatchesAssetIsExcluded(t *testing.T) {
	m, _ := New("/app", "src", "dist", []Asset{})
	cases := []struct {
		asset Asset
		path  string
		want  bool
	}{
		{Asset{Exclude: ""}, filepath.Join("/app", "src", "ignored.txt"), false},
		{Asset{Exclude: "**/ignored.txt"}, filepath.Join("/app", "src", "ignored.txt"), true},
		{Asset{Exclude: "ignored.txt"}, filepath.Join("/app", "src", "ignored.txt"), true},
	}
	for i, tc := range cases {
		if got := m.isExcluded(tc.asset, tc.path); got != tc.want {
			t.Errorf("case %d: isExcluded = %v; want %v", i, got, tc.want)
		}
	}
}

func TestCopyFilePreservesPermissions(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "script.sh")
	os.WriteFile(src, []byte("#!/bin/sh\necho hi"), 0755)

	dstDir := filepath.Join(dir, "out")
	os.MkdirAll(dstDir, 0755)

	if err := copyFile(src, dir, dstDir); err != nil {
		t.Fatalf("copyFile: %v", err)
	}

	dst := filepath.Join(dstDir, "script.sh")
	info, err := os.Stat(dst)
	if err != nil {
		t.Fatalf("stat dst: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Errorf("executable bit lost: mode = %v", info.Mode())
	}
}

func TestCopyAsset(t *testing.T) {
	dir := t.TempDir()
	srcRoot := filepath.Join(dir, "src")
	outDir := filepath.Join(dir, "dist")
	os.MkdirAll(filepath.Join(srcRoot, "sub"), 0755)
	os.WriteFile(filepath.Join(srcRoot, "sub", "data.json"), []byte(`{"ok":true}`), 0644)

	mgr, err := New(dir, "src", "dist", []Asset{{Glob: "**/*.json"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := mgr.Copy(); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	dst := filepath.Join(outDir, "sub", "data.json")
	if _, err := os.Stat(dst); err != nil {
		t.Errorf("expected %q to exist: %v", dst, err)
	}
}

func TestExpandGlob(t *testing.T) {
	dir := t.TempDir()
	srcRoot := filepath.Join(dir, "src")
	os.MkdirAll(srcRoot, 0755)
	os.WriteFile(filepath.Join(srcRoot, "a.proto"), []byte(""), 0644)
	os.WriteFile(filepath.Join(srcRoot, "b.json"), []byte(""), 0644)

	mgr, _ := New(dir, "src", "dist", nil)
	files, err := mgr.expandGlob(filepath.Join(srcRoot, "*.proto"), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || filepath.Base(files[0]) != "a.proto" {
		t.Errorf("expandGlob returned %v", files)
	}
}

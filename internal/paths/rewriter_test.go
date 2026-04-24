package paths

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRewriteAlias validates the source→output path mapping logic.
func TestRewriteAlias(t *testing.T) {
	cwd, _ := os.Getwd()

	paths := map[string][]string{
		"~":   {"./src"},
		"~/*": {"./src/*"},
	}
	// cwd, paths, outDir, rootDir
	r := New(cwd, paths, "dist", "src")

	cases := []struct {
		name      string
		specifier string
		fromFile  string // absolute path to the output .js file
		want      string
		wantOk    bool
	}{
		{
			name:      "~/db from dist root file",
			specifier: "~/db",
			fromFile:  filepath.Join(cwd, "dist", "app.module.js"),
			want:      "./db",
			wantOk:    true,
		},
		{
			name:      "~/core/guards from dist root file",
			specifier: "~/core/guards",
			fromFile:  filepath.Join(cwd, "dist", "app.module.js"),
			want:      "./core/guards",
			wantOk:    true,
		},
		{
			name:      "~/db from nested output file",
			specifier: "~/db",
			fromFile:  filepath.Join(cwd, "dist", "core", "guards", "auth.guard.js"),
			want:      "../../db",
			wantOk:    true,
		},
		{
			name:      "exact ~ alias from dist root file",
			specifier: "~",
			fromFile:  filepath.Join(cwd, "dist", "main.js"),
			want:      ".",
			wantOk:    true,
		},
		{
			name:      "relative import not rewritten",
			specifier: "./db",
			fromFile:  filepath.Join(cwd, "dist", "app.module.js"),
			want:      "",
			wantOk:    false,
		},
		{
			name:      "unrelated import not rewritten",
			specifier: "@nestjs/common",
			fromFile:  filepath.Join(cwd, "dist", "app.module.js"),
			want:      "",
			wantOk:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := r.resolveAlias(tc.specifier, tc.fromFile)
			if ok != tc.wantOk {
				t.Fatalf("resolveAlias(%q) ok=%v want ok=%v", tc.specifier, ok, tc.wantOk)
			}
			if ok && got != tc.want {
				t.Errorf("resolveAlias(%q) = %q, want %q", tc.specifier, got, tc.want)
			}
		})
	}
}

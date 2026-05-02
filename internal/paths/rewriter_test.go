package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRewriteAlias(t *testing.T) {
	cwd, _ := os.Getwd()
	r := New(cwd, map[string][]string{"~": {"./src"}, "~/*": {"./src/*"}}, "dist", "src")

	cases := []struct {
		name      string
		specifier string
		fromFile  string
		want      string
		wantOk    bool
	}{
		{"~/db from dist root", "~/db", filepath.Join(cwd, "dist", "app.module.js"), "./db", true},
		{"~/core/guards from dist root", "~/core/guards", filepath.Join(cwd, "dist", "app.module.js"), "./core/guards", true},
		{"~/db from nested", "~/db", filepath.Join(cwd, "dist", "core", "guards", "auth.guard.js"), "../../db", true},
		{"exact ~ alias", "~", filepath.Join(cwd, "dist", "main.js"), ".", true},
		{"relative not rewritten", "./db", filepath.Join(cwd, "dist", "app.module.js"), "", false},
		{"node_module not rewritten", "@nestjs/common", filepath.Join(cwd, "dist", "app.module.js"), "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := r.resolveAlias(tc.specifier, tc.fromFile)
			if ok != tc.wantOk {
				t.Fatalf("resolveAlias(%q) ok=%v want %v", tc.specifier, ok, tc.wantOk)
			}
			if ok && got != tc.want {
				t.Errorf("resolveAlias(%q) = %q, want %q", tc.specifier, got, tc.want)
			}
		})
	}
}

func TestRewriteFileSingleLine(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	r := New(cwd, map[string][]string{"@app/*": {"./src/*"}}, "dist", "src")

	jsFile := filepath.Join(dir, "main.js")
	os.WriteFile(jsFile, []byte(`const x = require("@app/service");`), 0644)

	if err := r.rewriteFile(jsFile); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(jsFile)
	if string(got) == `const x = require("@app/service");` {
		t.Error("file was not rewritten")
	}
}

func TestRewriteFileDynamicImport(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	r := New(cwd, map[string][]string{"@app/*": {"./src/*"}}, "dist", "src")

	jsFile := filepath.Join(dir, "lazy.js")
	os.WriteFile(jsFile, []byte(`const m = import("@app/module");`), 0644)

	if err := r.rewriteFile(jsFile); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(jsFile)
	if string(got) == `const m = import("@app/module");` {
		t.Error("dynamic import was not rewritten")
	}
}

func TestRewriteFileMultilineImport(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	r := New(cwd, map[string][]string{"@app/*": {"./src/*"}}, "dist", "src")

	src := "import {\n  Foo,\n  Bar\n} from \"@app/foo\";"
	jsFile := filepath.Join(dir, "multi.js")
	os.WriteFile(jsFile, []byte(src), 0644)

	if err := r.rewriteFile(jsFile); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(jsFile)
	if string(got) == src {
		t.Error("multiline import was not rewritten")
	}
}

func TestRewriteFileNoChange(t *testing.T) {
	dir := t.TempDir()
	cwd, _ := os.Getwd()
	r := New(cwd, map[string][]string{"@app/*": {"./src/*"}}, "dist", "src")

	original := `const x = require("./local");`
	jsFile := filepath.Join(dir, "noop.js")
	os.WriteFile(jsFile, []byte(original), 0644)

	info1, _ := os.Stat(jsFile)
	if err := r.rewriteFile(jsFile); err != nil {
		t.Fatal(err)
	}
	info2, _ := os.Stat(jsFile)
	// File should not be written if nothing changed
	if info1.ModTime() != info2.ModTime() {
		t.Error("file was written despite no alias matches")
	}
}

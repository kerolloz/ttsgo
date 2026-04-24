package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStripJSONC(t *testing.T) {
	input := `{
		// single line comment
		"foo": "bar", /* multi-line
		comment */
		"baz": "// not a comment",
		"qux": "/* not a comment */",
		"trailing": "comma",
	}`
	
	got := string(stripJSONC([]byte(input)))
	if !strings.Contains(got, `"foo": "bar"`) || strings.Contains(got, "single line comment") {
		t.Errorf("stripJSONC failed. Got:\n%s", got)
	}
	if strings.Contains(got, `"trailing": "comma",`) {
		t.Errorf("stripJSONC failed to strip trailing comma. Got:\n%s", got)
	}
	if !strings.Contains(got, `"trailing": "comma"`) {
		t.Errorf("stripJSONC missing correctly stripped comma. Got:\n%s", got)
	}
}

func TestResolveExtends(t *testing.T) {
	cwd := "/app"
	currentDir := "/app/src"
	
	cases := []struct {
		extends string
		want    string
	}{
		{"./base", "/app/src/base.json"},
		{"../tsconfig.json", "/app/tsconfig.json"},
		{"@tsconfig/node20", "/app/node_modules/@tsconfig/node20/tsconfig.json"},
	}
	
	for _, tc := range cases {
		got := resolveExtends(cwd, currentDir, tc.extends)
		if filepath.ToSlash(got) != tc.want {
			t.Errorf("resolveExtends(%q) = %q, want %q", tc.extends, got, tc.want)
		}
	}
}

func TestLoadTsConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "tsconfig-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	basePath := filepath.Join(tempDir, "tsconfig.base.json")
	baseContent := `{
		"compilerOptions": {
			"outDir": "dist-base",
			"emitDecoratorMetadata": true
		}
	}`
	if err := os.WriteFile(basePath, []byte(baseContent), 0644); err != nil {
		t.Fatal(err)
	}

	mainPath := filepath.Join(tempDir, "tsconfig.json")
	mainContent := `{
		"extends": "./tsconfig.base.json",
		"compilerOptions": {
			"rootDir": "src"
		}
	}`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadTsConfig(tempDir, "tsconfig.json")
	if err != nil {
		t.Fatalf("LoadTsConfig failed: %v", err)
	}

	if cfg.OutDir != "dist-base" {
		t.Errorf("expected OutDir dist-base, got %q", cfg.OutDir)
	}
	if cfg.RootDir != "src" {
		t.Errorf("expected RootDir src, got %q", cfg.RootDir)
	}
	if !cfg.EmitDecoratorMetadata {
		t.Error("expected EmitDecoratorMetadata true")
	}
}

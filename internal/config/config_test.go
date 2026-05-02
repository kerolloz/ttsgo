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

	got, err := stripJSONC([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(got)
	if !strings.Contains(s, `"foo": "bar"`) || strings.Contains(s, "single line comment") {
		t.Errorf("stripJSONC failed. Got:\n%s", s)
	}
	if strings.Contains(s, `"trailing": "comma",`) {
		t.Errorf("stripJSONC failed to strip trailing comma. Got:\n%s", s)
	}
}

func TestStripJSONCUnterminatedComment(t *testing.T) {
	_, err := stripJSONC([]byte(`{"foo": /* unterminated`))
	if err == nil {
		t.Error("expected error for unterminated block comment")
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
	dir := t.TempDir()

	base := filepath.Join(dir, "tsconfig.base.json")
	os.WriteFile(base, []byte(`{"compilerOptions":{"outDir":"dist-base","emitDecoratorMetadata":true}}`), 0644)

	main := filepath.Join(dir, "tsconfig.json")
	os.WriteFile(main, []byte(`{"extends":"./tsconfig.base.json","compilerOptions":{"rootDir":"src"}}`), 0644)

	cfg, err := LoadTsConfig(dir, "tsconfig.json")
	if err != nil {
		t.Fatalf("LoadTsConfig failed: %v", err)
	}
	if cfg.OutDir != "dist-base" {
		t.Errorf("OutDir = %q, want dist-base", cfg.OutDir)
	}
	if cfg.RootDir != "src" {
		t.Errorf("RootDir = %q, want src", cfg.RootDir)
	}
	if !cfg.EmitDecoratorMetadata {
		t.Error("expected EmitDecoratorMetadata true")
	}
}

func TestLoadTsConfigMissingFile(t *testing.T) {
	_, err := LoadTsConfig(t.TempDir(), "nonexistent.json")
	if err == nil {
		t.Error("expected error for missing tsconfig")
	}
}

func TestLoadNestConfig(t *testing.T) {
	dir := t.TempDir()
	content := `{
		"sourceRoot": "app",
		"entryFile": "index",
		"exec": "node",
		"compilerOptions": {
			"deleteOutDir": true,
			"assets": ["**/*.graphql", {"include": "**/*.proto", "exclude": "**/ignored"}]
		}
	}`
	os.WriteFile(filepath.Join(dir, "nest-cli.json"), []byte(content), 0644)

	cfg, err := Load(dir, "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.SourceRoot != "app" {
		t.Errorf("SourceRoot = %q, want app", cfg.SourceRoot)
	}
	if !cfg.CompilerOptions.DeleteOutDir {
		t.Error("expected DeleteOutDir true")
	}
	if len(cfg.CompilerOptions.Assets) != 2 {
		t.Errorf("expected 2 assets, got %d", len(cfg.CompilerOptions.Assets))
	}
	if cfg.CompilerOptions.Assets[0].ResolvedGlob() != "**/*.graphql" {
		t.Errorf("unexpected first asset glob: %q", cfg.CompilerOptions.Assets[0].ResolvedGlob())
	}
}

func TestLoadNestConfigDefaults(t *testing.T) {
	cfg, err := Load(t.TempDir(), "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.SourceRoot != "src" || cfg.EntryFile != "main" || cfg.Exec != "node" {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
}

func TestLoadNestConfigUnsupportedBuilder(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "nest-cli.json"), []byte(`{"compilerOptions":{"builder":"swc"}}`), 0644)
	_, err := Load(dir, "")
	if err == nil {
		t.Error("expected error for unsupported builder")
	}
}

package assets

import (
	"path/filepath"
	"testing"
)

func TestMatchesAsset(t *testing.T) {
	cwd := "/app"
	sourceRoot := "src"
	outDir := "dist"

	m, err := New(cwd, sourceRoot, outDir, []Asset{})
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		asset Asset
		path  string
		want  bool
	}{
		{
			asset: Asset{Exclude: ""},
			path:  filepath.Join(cwd, "src", "ignored.txt"),
			want:  false, // since no glob
		},
		{
			asset: Asset{Exclude: "**/ignored.txt"},
			path:  filepath.Join(cwd, "src", "ignored.txt"),
			want:  true, // excluded
		},
		{
			asset: Asset{Exclude: "ignored.txt"},
			path:  filepath.Join(cwd, "src", "ignored.txt"),
			want:  true, // excluded
		},
	}

	for i, tc := range cases {
		if got := m.isExcluded(tc.asset, tc.path); got != tc.want {
			t.Errorf("case %d: isExcluded(%v, %q) = %v; want %v", i, tc.asset, tc.path, got, tc.want)
		}
	}
}

package tspath

import (
	inner "github.com/microsoft/typescript-go/internal/tspath"
	_ "unsafe"
)

var _ = inner.ResolvePath

//go:linkname ResolvePath github.com/microsoft/typescript-go/internal/tspath.ResolvePath
func ResolvePath(path string, paths ...string) string

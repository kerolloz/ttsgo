package core

import (
	inner "github.com/microsoft/typescript-go/internal/core"
	_ "unsafe"
)

var _ = inner.CompilerOptions{}

type CompilerOptions = inner.CompilerOptions

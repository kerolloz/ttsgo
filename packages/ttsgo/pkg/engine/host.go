package engine

import (
	"github.com/microsoft/typescript-go/shim/bundled"
	shimcompiler "github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/vfs"
	"github.com/microsoft/typescript-go/shim/vfs/cachedvfs"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
)

// DefaultFS returns an OS-backed filesystem wrapped with tsgo's bundled libs.
func DefaultFS() vfs.FS {
	return bundled.WrapFS(cachedvfs.From(osvfs.FS()))
}

// DefaultHost returns a CompilerHost anchored at cwd.
func DefaultHost(cwd string, fs vfs.FS) shimcompiler.CompilerHost {
	return shimcompiler.NewCompilerHost(cwd, fs, bundled.LibPath(), nil, nil)
}

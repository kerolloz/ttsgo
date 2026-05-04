package bundled

import (
	"github.com/microsoft/typescript-go/internal/vfs"
	inner "github.com/microsoft/typescript-go/internal/bundled"
	_ "unsafe"
)

// Prevent unused import error
var _ = inner.LibPath

//go:linkname WrapFS github.com/microsoft/typescript-go/internal/bundled.WrapFS
func WrapFS(fs vfs.FS) vfs.FS

//go:linkname LibPath github.com/microsoft/typescript-go/internal/bundled.LibPath
func LibPath() string

package osvfs

import (
	"github.com/microsoft/typescript-go/internal/vfs"
	inner "github.com/microsoft/typescript-go/internal/vfs/osvfs"
	_ "unsafe"
)

var _ = inner.FS

//go:linkname FS github.com/microsoft/typescript-go/internal/vfs/osvfs.FS
func FS() vfs.FS

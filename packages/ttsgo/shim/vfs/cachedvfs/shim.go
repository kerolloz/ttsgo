package cachedvfs

import (
	"github.com/microsoft/typescript-go/internal/vfs"
	inner "github.com/microsoft/typescript-go/internal/vfs/cachedvfs"
	_ "unsafe"
)

type FS = inner.FS

//go:linkname From github.com/microsoft/typescript-go/internal/vfs/cachedvfs.From
func From(fs vfs.FS) *FS

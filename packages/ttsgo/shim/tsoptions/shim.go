package tsoptions

import (
	"github.com/microsoft/typescript-go/internal/ast"
	"github.com/microsoft/typescript-go/internal/collections"
	"github.com/microsoft/typescript-go/internal/core"
	inner "github.com/microsoft/typescript-go/internal/tsoptions"
	_ "unsafe"
)

var _ = inner.GetParsedCommandLineOfConfigFile

type ParsedCommandLine = inner.ParsedCommandLine
type ParseConfigHost = inner.ParseConfigHost
type ExtendedConfigCache = inner.ExtendedConfigCache

//go:linkname GetParsedCommandLineOfConfigFile github.com/microsoft/typescript-go/internal/tsoptions.GetParsedCommandLineOfConfigFile
func GetParsedCommandLineOfConfigFile(configFileName string, options *core.CompilerOptions, optionsRaw *collections.OrderedMap[string, any], sys inner.ParseConfigHost, extendedConfigCache inner.ExtendedConfigCache) (*inner.ParsedCommandLine, []*ast.Diagnostic)

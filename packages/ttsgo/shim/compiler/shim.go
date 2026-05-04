package compiler

import (
	"context"
	"github.com/microsoft/typescript-go/internal/ast"
	inner "github.com/microsoft/typescript-go/internal/compiler"
	"github.com/microsoft/typescript-go/internal/diagnostics"
	"github.com/microsoft/typescript-go/internal/tsoptions"
	"github.com/microsoft/typescript-go/internal/vfs"
	_ "unsafe"
)

var _ = inner.NewProgram

type CompilerHost = inner.CompilerHost
type EmitOptions = inner.EmitOptions
type EmitResult = inner.EmitResult
type Program = inner.Program
type ProgramOptions = inner.ProgramOptions
type WriteFile = inner.WriteFile
type WriteFileData = inner.WriteFileData

//go:linkname NewProgram github.com/microsoft/typescript-go/internal/compiler.NewProgram
func NewProgram(opts inner.ProgramOptions) *inner.Program

//go:linkname NewCompilerHost github.com/microsoft/typescript-go/internal/compiler.NewCompilerHost
func NewCompilerHost(currentDirectory string, fs vfs.FS, defaultLibraryPath string, extendedConfigCache tsoptions.ExtendedConfigCache, trace func(msg *diagnostics.Message, args ...any)) inner.CompilerHost

//go:linkname SortAndDeduplicateDiagnostics github.com/microsoft/typescript-go/internal/compiler.SortAndDeduplicateDiagnostics
func SortAndDeduplicateDiagnostics(diagnostics []*ast.Diagnostic) []*ast.Diagnostic

//go:linkname GetDiagnosticsOfAnyProgram github.com/microsoft/typescript-go/internal/compiler.GetDiagnosticsOfAnyProgram
func GetDiagnosticsOfAnyProgram(ctx context.Context, program inner.ProgramLike, file *ast.SourceFile, skipNoEmitCheckForDtsDiagnostics bool, getBindDiagnostics func(context.Context, *ast.SourceFile) []*ast.Diagnostic, getSemanticDiagnostics func(context.Context, *ast.SourceFile) []*ast.Diagnostic) []*ast.Diagnostic

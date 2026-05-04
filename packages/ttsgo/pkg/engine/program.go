package engine

import (
	"context"
	shimast "github.com/microsoft/typescript-go/shim/ast"
	shimcompiler "github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/tsoptions"
)

// Program wraps the native TSGo program.
type Program struct {
	TSProgram    *shimcompiler.Program
	ParsedConfig *tsoptions.ParsedCommandLine
	Host         shimcompiler.CompilerHost
}

// Close is a placeholder for any resources that need releasing.
func (p *Program) Close() error {
	return nil
}

// Diagnostics returns all semantic diagnostics for the program.
func (p *Program) Diagnostics(ctx context.Context) []*shimast.Diagnostic {
	raw := shimcompiler.GetDiagnosticsOfAnyProgram(
		ctx,
		p.TSProgram,
		nil,
		false,
		p.TSProgram.GetBindDiagnostics,
		p.TSProgram.GetSemanticDiagnostics,
	)
	return shimcompiler.SortAndDeduplicateDiagnostics(raw)
}

// Emit runs the emission pipeline with an optional WriteFile hook.
func (p *Program) Emit(ctx context.Context, writeFile shimcompiler.WriteFile) *shimcompiler.EmitResult {
	return p.TSProgram.Emit(ctx, shimcompiler.EmitOptions{
		WriteFile: writeFile,
	})
}

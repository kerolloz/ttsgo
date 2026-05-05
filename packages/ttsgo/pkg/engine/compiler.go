package engine

import "context"

// Compiler abstracts the compilation backend.
type Compiler interface {
	Compile(ctx context.Context, opts Options) (*Result, error)
}

// TsGoCompiler is the default Compiler implementation using typescript-go.
type TsGoCompiler struct{}

// Compile runs the full compilation pipeline with path alias rewriting.
func (c *TsGoCompiler) Compile(ctx context.Context, opts Options) (*Result, error) {
	return CompileWithRewrite(ctx, opts)
}

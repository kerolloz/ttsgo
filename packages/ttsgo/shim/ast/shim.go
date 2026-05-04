package ast

import (
	inner "github.com/microsoft/typescript-go/internal/ast"
	_ "unsafe"
)

var _ = inner.GetNodeAtPosition

type Diagnostic = inner.Diagnostic
type SourceFile = inner.SourceFile
type Node = inner.Node

const (
	KindFunctionDeclaration = inner.KindFunctionDeclaration
)

//go:linkname GetNodeAtPosition github.com/microsoft/typescript-go/internal/ast.GetNodeAtPosition
func GetNodeAtPosition(file *inner.SourceFile, position int, includeJSDoc bool) *inner.Node

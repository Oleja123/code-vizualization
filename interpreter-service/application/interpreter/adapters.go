package interpreter

import "github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"

type VariableDecl struct {
	converter.VariableDecl
	IsGlobal bool
}

type VariableExpr struct {
	converter.VariableExpr
	IsLvalue bool
}

type ArrayAccessExpr struct {
	converter.ArrayAccessExpr
	IsLvalue bool
}

package interfaces

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// Location представляет позицию узла в исходном коде
type Location struct {
	Line      uint32 `json:"line"`
	Column    uint32 `json:"column"`
	EndLine   uint32 `json:"endLine"`
	EndColumn uint32 `json:"endColumn"`
}

// Node является базовым интерфейсом для всех узлов AST
type Node interface{}

// Stmt представляет оператор (statement) в C
type Stmt interface {
	Node
	StmtNode() // маркерный метод для statements
	GetLocation() Location
}

// Expr представляет выражение (expression) в C
type Expr interface {
	Node
	ExprNode() // маркерный метод для expressions
	GetLocation() Location
}

// Converter определяет интерфейс для конвертации tree-sitter CST в AST
type Converter interface {
	// ConvertToProgram преобразует корневой узел tree-sitter в Program
	ConvertToProgram(tree *sitter.Tree, sourceCode []byte) (Node, error)

	// ConvertStmt преобразует узел tree-sitter в Statement
	ConvertStmt(node *sitter.Node, sourceCode []byte) (Stmt, error)

	// ConvertExpr преобразует узел tree-sitter в Expression
	ConvertExpr(node *sitter.Node, sourceCode []byte) (Expr, error)
}

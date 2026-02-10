package interfaces

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// Node является базовым интерфейсом для всех узлов AST
type Node interface{}

// Stmt представляет оператор (statement) в C
type Stmt interface {
	Node
	StmtNode() // маркерный метод для statements
}

// Expr представляет выражение (expression) в C
type Expr interface {
	Node
	ExprNode()      // маркерный метод для expressions
	IsLValue() bool // возвращает true, если выражение является lvalue (может стоять слева от =)
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

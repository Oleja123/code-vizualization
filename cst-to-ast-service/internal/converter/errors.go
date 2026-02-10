package converter

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/domain/structs"
	sitter "github.com/smacker/go-tree-sitter"
)

type ErrorCode string

const (
	ErrParseFailed                 ErrorCode = "ParseFailed"
	ErrStmtUnsupported             ErrorCode = "StmtUnsupported"
	ErrExprUnsupported             ErrorCode = "ExprUnsupported"
	ErrEmptyParenthesizedExpr      ErrorCode = "EmptyParenthesizedExpr"
	ErrInvalidDeclaration          ErrorCode = "InvalidDeclaration"
	ErrInitializerConversion       ErrorCode = "InitializerConversion"
	ErrEmptyExpressionStatement    ErrorCode = "EmptyExpressionStatement"
	ErrInvalidExpressionStatement  ErrorCode = "InvalidExpressionStatement"
	ErrInvalidReturnStatement      ErrorCode = "InvalidReturnStatement"
	ErrInvalidAssignmentExpression ErrorCode = "InvalidAssignmentExpression"
	ErrInvalidCallExpression       ErrorCode = "InvalidCallExpression"
	ErrEmptyArrayInitializer       ErrorCode = "EmptyArrayInitializer"
	ErrInvalidPostfixOperator      ErrorCode = "InvalidPostfixOperator"
	ErrInvalidIdentifier           ErrorCode = "InvalidIdentifier"
	ErrUnsupportedOperator         ErrorCode = "UnsupportedOperator"
	ErrRequiresLValue              ErrorCode = "RequiresLValue"
	ErrTreeSitterError             ErrorCode = "TreeSitterError"
	ErrIntLiteralParse             ErrorCode = "IntLiteralParse"
	ErrStmtConversion              ErrorCode = "StmtConversion"
)

type ConverterError struct {
	Code     ErrorCode        `json:"code"`
	Message  string           `json:"message"`
	NodeType string           `json:"nodeType,omitempty"`
	Loc      structs.Location `json:"location,omitempty"`
	Cause    error            `json:"-"`
}

func (e *ConverterError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *ConverterError) Unwrap() error { return e.Cause }

func newConverterError(code ErrorCode, message string, node *sitter.Node, cause error) *ConverterError {
	var loc structs.Location
	var nodeType string
	if node != nil {
		loc = structs.Location{
			Line:      node.StartPoint().Row + 1,
			Column:    node.StartPoint().Column,
			EndLine:   node.EndPoint().Row + 1,
			EndColumn: node.EndPoint().Column,
		}
		nodeType = node.Type()
	}

	return &ConverterError{
		Code:     code,
		Message:  message,
		NodeType: nodeType,
		Loc:      loc,
		Cause:    cause,
	}
}

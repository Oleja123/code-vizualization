package converter

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/domain/structs"
	sitter "github.com/smacker/go-tree-sitter"
)

// ErrorCode определяет тип ошибки парсинга/конвертации
type ErrorCode string

const (
	ErrParseFailed     ErrorCode = "ParseFailed"
	ErrStmtUnsupported ErrorCode = "StmtUnsupported"
	ErrExprUnsupported ErrorCode = "ExprUnsupported"
	ErrTreeSitterError ErrorCode = "TreeSitterError"
	ErrIntLiteralParse ErrorCode = "IntLiteralParse"
	ErrStmtConversion  ErrorCode = "StmtConversion"
)

// ConverterError представляет ошибку парсинга с полной информацией для интерпретатора
type ConverterError struct {
	// Code - код ошибки для идентификации типа проблемы
	Code ErrorCode `json:"code"`

	// Message - понятное описание ошибки (включает информацию о Cause если она есть)
	Message string `json:"message"`

	// NodeType - тип узла tree-sitter, вызвавший ошибку (если известен)
	NodeType string `json:"nodeType,omitempty"`

	// Loc - позиция в исходном коде (если известна)
	Loc structs.Location `json:"location,omitempty"`

	// Cause - оригинальная ошибка (внутреннее использование)
	Cause error `json:"-"`
}

// Error реализует интерфейс error
func (e *ConverterError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap возвращает оригинальную ошибку
func (e *ConverterError) Unwrap() error {
	return e.Cause
}

// GetLocation возвращает позицию ошибки в коде
func (e *ConverterError) GetLocation() structs.Location {
	return e.Loc
}

// GetCode возвращает код ошибки
func (e *ConverterError) GetCode() ErrorCode {
	return e.Code
}

// GetMessage возвращает сообщение об ошибке
func (e *ConverterError) GetMessage() string {
	return e.Message
}

// GetNodeType возвращает тип узла tree-sitter
func (e *ConverterError) GetNodeType() string {
	return e.NodeType
}

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
	// Дополняем сообщение информацией о Cause если она есть
	finalMessage := message
	if cause != nil {
		finalMessage = fmt.Sprintf("%s: %v", message, cause)
	}

	return &ConverterError{
		Code:     code,
		Message:  finalMessage,
		NodeType: nodeType,
		Loc:      loc,
		Cause:    cause,
	}
}

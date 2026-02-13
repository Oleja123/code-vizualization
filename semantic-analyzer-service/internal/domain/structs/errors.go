package structs

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
)

// ErrorCode представляет код семантической ошибки
type ErrorCode string

const (
	ErrInvalidVariableType     ErrorCode = "INVALID_VAR_TYPE"
	ErrInvalidParameterType    ErrorCode = "INVALID_PARAM_TYPE"
	ErrInvalidReturnType       ErrorCode = "INVALID_RETURN_TYPE"
	ErrUnsupportedAssignOp     ErrorCode = "UNSUPPORTED_ASSIGN_OP"
	ErrUnsupportedUnaryOp      ErrorCode = "UNSUPPORTED_UNARY_OP"
	ErrUnsupportedBinaryOp     ErrorCode = "UNSUPPORTED_BINARY_OP"
	ErrInvalidFunctionCall     ErrorCode = "INVALID_FUNCTION_CALL"
	ErrSemanticValidationError ErrorCode = "SEMANTIC_ERROR"
)

// SemanticError представляет семантическую ошибку
type SemanticError struct {
	Code     ErrorCode
	Message  string
	Location converter.Location
	NodeType string
	Details  string
}

// Error реализует интерфейс error
func (e *SemanticError) Error() string {
	return fmt.Sprintf("[%s] %s at line %d, column %d: %s",
		e.Code, e.Message, e.Location.Line, e.Location.Column, e.Details)
}

// GetCode возвращает код ошибки
func (e *SemanticError) GetCode() ErrorCode {
	return e.Code
}

// GetMessage возвращает сообщение ошибки
func (e *SemanticError) GetMessage() string {
	return e.Message
}

// GetLocation возвращает позицию ошибки
func (e *SemanticError) GetLocation() converter.Location {
	return e.Location
}

// GetDetails возвращает детали ошибки
func (e *SemanticError) GetDetails() string {
	return e.Details
}

// NewSemanticError создает новую семантическую ошибку
func NewSemanticError(code ErrorCode, message string, location converter.Location, nodeType string, details string) *SemanticError {
	return &SemanticError{
		Code:     code,
		Message:  message,
		Location: location,
		NodeType: nodeType,
		Details:  details,
	}
}

package interfaces

import "github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"

// Validator интерфейс для семантической валидации
type Validator interface {
	// ValidateProgram проверяет программу на семантические ошибки
	ValidateProgram(program *converter.Program) error
}

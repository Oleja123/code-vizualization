package validator

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
)

// SemanticValidator выполняет семантическую валидацию AST
type SemanticValidator struct {
	// Допустимые операторы и типы
	allowedAssignOps  map[string]bool
	allowedUnaryOps   map[string]bool
	allowedBinaryOps  map[string]bool
	allowedReturnType map[string]bool
}

// New создает новый семантический валидатор
func New() *SemanticValidator {
	return &SemanticValidator{
		allowedAssignOps: map[string]bool{
			"=":  true,
			"+=": true,
			"-=": true,
			"*=": true,
			"%=": true,
			"/=": true,
		},
		allowedUnaryOps: map[string]bool{
			"-":  true,
			"+":  true,
			"++": true,
			"--": true,
			"!":  true,
		},
		allowedBinaryOps: map[string]bool{
			"+":  true,
			"-":  true,
			"*":  true,
			"/":  true,
			"%":  true,
			"==": true,
			"!=": true,
			"<":  true,
			"<=": true,
			">":  true,
			">=": true,
			"&&": true,
			"||": true,
		},
		allowedReturnType: map[string]bool{
			"int":  true,
			"void": true,
		},
	}
}

// ValidateProgram выполняет валидацию всей программы
func (v *SemanticValidator) ValidateProgram(program *converter.Program) error {
	// Проверяем все объявления
	for _, decl := range program.Declarations {
		if funcDecl, ok := decl.(*converter.FunctionDecl); ok {
			if err := v.validateFunctionDecl(funcDecl); err != nil {
				return err
			}
		} else if varDecl, ok := decl.(*converter.VariableDecl); ok {
			if err := v.validateVariableDecl(varDecl); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateFunctionDecl проверяет объявление функции
func (v *SemanticValidator) validateFunctionDecl(fn *converter.FunctionDecl) error {
	// Проверяем возвращаемый тип
	if !v.allowedReturnType[fn.ReturnType.BaseType] {
		return NewSemanticError(
			ErrInvalidReturnType,
			fmt.Sprintf("invalid return type: %s", fn.ReturnType.BaseType),
			fn.Loc,
			"FunctionDecl",
			fmt.Sprintf("function '%s' has unsupported return type. Allowed: void, int", fn.Name),
		)
	}

	// Проверяем параметры
	for _, param := range fn.Parameters {
		if err := v.validateType(param.Type, "parameter", param.Name, param.Loc); err != nil {
			return err
		}
	}

	// Проверяем тело функции
	if fn.Body != nil {
		if err := v.validateStmt(fn.Body); err != nil {
			return err
		}
	}

	return nil
}

// validateVariableDecl проверяет объявление переменной
func (v *SemanticValidator) validateVariableDecl(varDecl *converter.VariableDecl) error {
	// Проверяем тип переменной
	if err := v.validateType(varDecl.VarType, "variable", varDecl.Name, varDecl.Loc); err != nil {
		return err
	}

	// Проверяем инициализирующее выражение
	if varDecl.InitExpr != nil {
		if err := v.validateExpr(varDecl.InitExpr); err != nil {
			return err
		}
	}

	return nil
}

// validateType проверяет, является ли тип допустимым
func (v *SemanticValidator) validateType(t converter.Type, context string, name string, loc converter.Location) error {
	// Типы должны быть только int (без указателей и массивов в этой версии)
	if t.BaseType != "int" {
		return NewSemanticError(
			ErrInvalidVariableType,
			fmt.Sprintf("invalid %s type: %s", context, t.BaseType),
			loc,
			"Type",
			fmt.Sprintf("only 'int' type is supported for %s '%s', got '%s'", context, name, t.BaseType),
		)
	}

	return nil
}

// validateStmt проверяет оператор
func (v *SemanticValidator) validateStmt(stmt interface{}) error {
	switch s := stmt.(type) {
	case *converter.VariableDecl:
		return v.validateVariableDecl(s)
	case *converter.BlockStmt:
		for _, s := range s.Statements {
			if err := v.validateStmt(s); err != nil {
				return err
			}
		}
	case *converter.ExprStmt:
		return v.validateExpr(s.Expression)
	case *converter.IfStmt:
		if err := v.validateExpr(s.Condition); err != nil {
			return err
		}
		if s.ThenBlock != nil {
			if err := v.validateStmt(s.ThenBlock); err != nil {
				return err
			}
		}
		if s.ElseBlock != nil {
			if err := v.validateStmt(s.ElseBlock); err != nil {
				return err
			}
		}
	case *converter.WhileStmt:
		if err := v.validateExpr(s.Condition); err != nil {
			return err
		}
		if err := v.validateStmt(s.Body); err != nil {
			return err
		}
	case *converter.DoWhileStmt:
		if err := v.validateStmt(s.Body); err != nil {
			return err
		}
		if err := v.validateExpr(s.Condition); err != nil {
			return err
		}
	case *converter.ForStmt:
		if s.Init != nil {
			if err := v.validateStmt(s.Init); err != nil {
				return err
			}
		}
		if s.Condition != nil {
			if err := v.validateExpr(s.Condition); err != nil {
				return err
			}
		}
		if s.Post != nil {
			if err := v.validateStmt(s.Post); err != nil {
				return err
			}
		}
		if err := v.validateStmt(s.Body); err != nil {
			return err
		}
	case *converter.ReturnStmt:
		if s.Value != nil {
			if err := v.validateExpr(s.Value); err != nil {
				return err
			}
		}
	case *converter.LabelStmt:
		if s.Statement != nil {
			if err := v.validateStmt(s.Statement); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateExpr проверяет выражение
func (v *SemanticValidator) validateExpr(expr interface{}) error {
	switch e := expr.(type) {
	case *converter.BinaryExpr:
		if !v.allowedBinaryOps[e.Operator] {
			return NewSemanticError(
				ErrUnsupportedBinaryOp,
				fmt.Sprintf("unsupported binary operator: %s", e.Operator),
				e.Loc,
				"BinaryExpr",
				fmt.Sprintf("binary operator '%s' is not supported", e.Operator),
			)
		}
		if err := v.validateExpr(e.Left); err != nil {
			return err
		}
		if err := v.validateExpr(e.Right); err != nil {
			return err
		}

	case *converter.UnaryExpr:
		if !v.allowedUnaryOps[e.Operator] {
			return NewSemanticError(
				ErrUnsupportedUnaryOp,
				fmt.Sprintf("unsupported unary operator: %s", e.Operator),
				e.Loc,
				"UnaryExpr",
				fmt.Sprintf("unary operator '%s' is not supported", e.Operator),
			)
		}
		if err := v.validateExpr(e.Operand); err != nil {
			return err
		}

	case *converter.AssignmentExpr:
		if !v.allowedAssignOps[e.Operator] {
			return NewSemanticError(
				ErrUnsupportedAssignOp,
				fmt.Sprintf("unsupported assignment operator: %s", e.Operator),
				e.Loc,
				"AssignmentExpr",
				fmt.Sprintf("assignment operator '%s' is not supported", e.Operator),
			)
		}
		if err := v.validateExpr(e.Left); err != nil {
			return err
		}
		if err := v.validateExpr(e.Right); err != nil {
			return err
		}

	case *converter.CallExpr:
		for _, arg := range e.Arguments {
			if err := v.validateExpr(arg); err != nil {
				return err
			}
		}

	case *converter.ArrayAccessExpr:
		if err := v.validateExpr(e.Array); err != nil {
			return err
		}
		if err := v.validateExpr(e.Index); err != nil {
			return err
		}

	case *converter.ArrayInitExpr:
		for _, elem := range e.Elements {
			if err := v.validateExpr(elem); err != nil {
				return err
			}
		}
	}

	return nil
}

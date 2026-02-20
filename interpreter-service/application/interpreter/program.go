package interpreter

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

func (i *Interpreter) ExecuteProgram(program *converter.Program) (*int, error) {
	if program == nil {
		return nil, runtimeerrors.NewErrUnexpectedInternalError("program is nil")
	}

	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *converter.FunctionDecl:
			_, err := i.executeStatement(d)
			if err != nil {
				return nil, err
			}
		case *converter.VariableDecl:
			_, err := i.executeStatement(&VariableDecl{VariableDecl: *d, IsGlobal: true})
			if err != nil {
				return nil, err
			}
		default:
			return nil, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("unsupported top-level declaration type %T", decl))
		}
	}

	if _, ok := i.Functions["main"]; !ok {
		return nil, runtimeerrors.NewErrUnexpectedInternalError("entrypoint function main not found")
	}

	mainCall := &converter.CallExpr{Type: "CallExpr", FunctionName: "main", Arguments: nil}
	value, err := i.executeCallExpr(mainCall)
	if err != nil {
		return nil, err
	}

	switch v := value.(type) {
	case nil:
		return nil, nil
	case *int:
		return v, nil
	case int:
		result := v
		return &result, nil
	default:
		return nil, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("unexpected main return type %T", value))
	}
}

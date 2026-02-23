package interpreter

import (
	"errors"
	"fmt"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/domain/eventdispatcher"
	"github.com/Oleja123/code-vizualization/interpreter-service/domain/events"
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

func (i *Interpreter) ExecuteProgram(program *converter.Program) (*int, []eventdispatcher.Step, int, error) {
	if program == nil {
		return nil, nil, 0, runtimeerrors.NewErrUnexpectedInternalError("program is nil")
	}

	for _, decl := range program.Declarations {
		switch d := decl.(type) {
		case *converter.FunctionDecl:
			_, err := i.executeStatement(d)
			if err != nil {
				return nil, nil, 0, err
			}
		case *converter.VariableDecl:
			_, err := i.executeStatement(&VariableDecl{VariableDecl: *d, IsGlobal: true})
			if err != nil {
				return nil, nil, 0, err
			}
		default:
			return nil, nil, 0, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("unsupported top-level declaration type %T", decl))
		}
	}

	stepBegin := i.currentStepNumber

	if _, ok := i.Functions["main"]; !ok {
		return nil, nil, 0, runtimeerrors.NewErrUnexpectedInternalError("entrypoint function main not found")
	}

	mainCall := &converter.CallExpr{Type: "CallExpr", FunctionName: "main", Arguments: nil}
	value, err := i.executeCallExpr(mainCall)

	if err != nil {
		var unfErr *runtimeerrors.ErrUndefinedBehavior

		if errors.As(err, &unfErr) {
			i.addEvents(events.UndefinedBehavior{Message: err.Error()})
			i.addStep()
			return nil, i.Steps, stepBegin, nil
		} else {
			return nil, nil, 0, err
		}
	}

	i.addEvents(events.LineChanged{Line: -1})
	i.addStep()

	switch v := value.(type) {
	case nil:
		return nil, i.Steps, stepBegin, nil
	case *int:
		return v, i.Steps, stepBegin, nil
	case int:
		result := v
		return &result, i.Steps, stepBegin, nil
	default:
		return nil, nil, 0, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("unexpected main return type %T", value))
	}
}

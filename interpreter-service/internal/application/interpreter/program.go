package interpreter

import (
	"errors"
	"fmt"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/eventdispatcher"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/events"
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/runtime/errors"
)

func (i *Interpreter) ExecuteProgram(program *converter.Program) (*int, []eventdispatcher.Step, int, error) {
	if program == nil {
		return nil, nil, 0, runtimeerrors.NewErrUnexpectedInternalError("program is nil")
	}

	i.resetExecutionState()

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
		var unfErr runtimeerrors.ErrUndefinedBehavior
		var runErr runtimeerrors.ErrRuntime

		if errors.As(err, &unfErr) {
			i.addEvents(events.UndefinedBehavior{Message: err.Error()})
			if stepErr := i.addStep(); stepErr != nil {
				return nil, nil, 0, stepErr
			}
			return nil, i.Steps, stepBegin, err
		} else if errors.As(err, &runErr) {
			i.addEvents(events.RuntimeError{Message: err.Error()})
			if stepErr := i.addStep(); stepErr != nil {
				return nil, nil, 0, stepErr
			}
			return nil, i.Steps, stepBegin, err
		} else {
			return nil, nil, 0, err
		}
	}

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

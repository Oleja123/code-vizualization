package interpreter

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime"
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

type Interpreter struct {
	CallStack   *runtime.CallStack
	GlobalScope *runtime.Scope
	Functions   map[string]*converter.FunctionDecl
}

func NewInterpreter() *Interpreter {
	globalScope := runtime.NewScope(nil)
	return &Interpreter{
		GlobalScope: globalScope,
		CallStack:   runtime.NewCallStack(globalScope),
		Functions:   make(map[string]*converter.FunctionDecl),
	}
}

func (i *Interpreter) resolveVariable(name string) (interface{}, error) {
	stackFrame := i.CallStack.GetCurrentFrame()

	valVariable, found := stackFrame.GetVariable(name)
	if found {
		return valVariable, nil
	}

	valArray, found := stackFrame.GetArray(name)
	if found {
		return valArray, nil
	}

	// valArray2D, found := stackFrame.GetArray2D(name)
	// if found {
	// 	return valArray2D, nil
	// }

	return nil, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("no variable named %s", name))
}

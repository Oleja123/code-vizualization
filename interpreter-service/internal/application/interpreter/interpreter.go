package interpreter

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/eventdispatcher"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/events"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/runtime"
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/runtime/errors"
)

type Interpreter struct {
	CallStack         *runtime.CallStack
	GlobalScope       *runtime.Scope
	Functions         map[string]*converter.FunctionDecl
	currentStepNumber int
	currentLine       int
	CurrentStep       eventdispatcher.Step
	Steps             []eventdispatcher.Step
}

func NewInterpreter() *Interpreter {
	globalScope := runtime.NewScope(nil)
	return &Interpreter{
		GlobalScope: globalScope,
		CallStack:   runtime.NewCallStack(globalScope),
		Functions:   make(map[string]*converter.FunctionDecl),
		currentLine: -1,
	}
}

func (i *Interpreter) incrementStep() {
	i.currentStepNumber++
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

	valArray2D, found := stackFrame.GetArray2D(name)
	if found {
		return valArray2D, nil
	}

	return nil, runtimeerrors.NewErrUnexpectedInternalError(fmt.Sprintf("no variable named %s", name))
}

func (i *Interpreter) addEvents(events ...events.Event) {
	i.CurrentStep.Events = append(i.CurrentStep.Events, events...)
}

func (i *Interpreter) addStep() {
	defer i.incrementStep()
	i.Steps = append(i.Steps, i.CurrentStep)
	i.CurrentStep = eventdispatcher.Step{}
}

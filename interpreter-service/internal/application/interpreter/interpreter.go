package interpreter

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/eventdispatcher"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/events"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/limitations"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/runtime"
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/runtime/errors"
)

const (
	defaultMaxAllocatedElements = 100
	defaultMaxSteps             = 1000
)

type Interpreter struct {
	CallStack         *runtime.CallStack
	GlobalScope       *runtime.Scope
	Functions         map[string]*converter.FunctionDecl
	LimitManager      limitations.LimitManager
	currentStepNumber int
	currentLine       int
	CurrentStep       eventdispatcher.Step
	Steps             []eventdispatcher.Step
	maxAllocated      int
	maxSteps          int
}

func NewInterpreter() *Interpreter {
	return NewInterpreterWithLimits(defaultMaxAllocatedElements, defaultMaxSteps)
}

func NewInterpreterWithLimits(maxAllocatedElements int, maxSteps int) *Interpreter {
	if maxAllocatedElements < 0 {
		maxAllocatedElements = defaultMaxAllocatedElements
	}

	if maxSteps < 0 {
		maxSteps = defaultMaxSteps
	}

	globalScope := runtime.NewScope(nil)
	interpreter := &Interpreter{
		GlobalScope:  globalScope,
		CallStack:    runtime.NewCallStack(globalScope),
		Functions:    make(map[string]*converter.FunctionDecl),
		currentLine:  -1,
		maxAllocated: maxAllocatedElements,
		maxSteps:     maxSteps,
	}

	interpreter.resetLimitManager()

	return interpreter
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

func (i *Interpreter) addStep() error {
	if err := i.LimitManager.MakeStep(); err != nil {
		return err
	}

	defer i.incrementStep()
	i.Steps = append(i.Steps, i.CurrentStep)
	i.CurrentStep = eventdispatcher.Step{}

	return nil
}

func (i *Interpreter) resetLimitManager() {
	i.LimitManager = limitations.LimitManager{
		AllocatedElementsRemained: i.maxAllocated,
		StepsRemained:             i.maxSteps,
	}
}

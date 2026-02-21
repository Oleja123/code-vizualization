package snapshot

import (
	"fmt"

	"github.com/Oleja123/code-vizualization/interpreter-service/domain/events"
	"github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime"
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

type Snapshot struct {
	CallStack   *runtime.CallStack
	GlobalScope *runtime.Scope
	Line        int
}

func NewSnapshot(globalScope *runtime.Scope) *Snapshot {
	return &Snapshot{
		CallStack:   runtime.NewCallStack(globalScope),
		GlobalScope: globalScope,
		Line:        0,
	}
}

func (sn *Snapshot) Apply(event events.Event, step int) error {
	switch e := event.(type) {
	case events.EnterScope:
		return sn.applyEnterScope()
	case events.ExitScope:
		return sn.applyExitScope()
	case events.DeclareVar:
		return sn.applyDeclareVar(e, step)
	case events.DeclareArray:
		return sn.applyDeclareArray(e, step)
	case events.DeclareArray2D:
		return sn.applyDeclareArray2D(e, step)
	case events.VarChanged:
		return sn.applyVarChanged(e, step)
	case events.ArrayElementChanged:
		return sn.applyArrayElementChanged(e, step)
	case events.Array2DElementChanged:
		return sn.applyArray2DElementChanged(e, step)
	case events.FunctionCall:
		return sn.applyFunctionCall(e)
	case events.FunctionReturn:
		return sn.applyFunctionReturn(e)
	case events.LineChanged:
		return sn.applyLineChanged(e)
	default:
		return runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("unknown event type: %T", e))
	}
}

func (sn *Snapshot) applyEnterScope() error {
	frame := sn.CallStack.GetCurrentFrame()
	if frame == nil {
		return runtimeerrors.NewErrUnexpectedInternalError("no current frame for enter scope")
	}
	frame.EnterScope()
	return nil
}

func (sn *Snapshot) applyExitScope() error {
	frame := sn.CallStack.GetCurrentFrame()
	if frame == nil {
		return runtimeerrors.NewErrUnexpectedInternalError("no current frame for exit scope")
	}
	return frame.ExitScope()
}

func (sn *Snapshot) applyDeclareVar(e events.DeclareVar, step int) error {
	variable := runtime.NewVariable(e.Name, e.Value, step, e.IsGlobal)
	sn.CallStack.DeclareInCurrentFrame(variable)
	return nil
}

func (sn *Snapshot) applyDeclareArray(e events.DeclareArray, step int) error {
	if e.Value == nil {
		arr := runtime.NewArray(e.Name, e.Size, nil, step, e.IsGlobal)
		sn.CallStack.DeclareInCurrentFrame(arr)
		return nil
	}
	elements := make([]runtime.ArrayElement, len(e.Value))
	for i, v := range e.Value {
		val := v
		elements[i] = *runtime.NewArrayElement(&val, step, false)
	}
	arr := runtime.NewArray(e.Name, e.Size, elements, step, false)
	sn.CallStack.DeclareInCurrentFrame(arr)
	return nil
}

func (sn *Snapshot) applyDeclareArray2D(e events.DeclareArray2D, step int) error {
	if e.Value == nil {
		arr := runtime.NewArray2D(e.Name, e.Size1, e.Size2, nil, step, e.IsGlobal)
		sn.CallStack.DeclareInCurrentFrame(arr)
		return nil
	}
	elements := make([]runtime.Array, len(e.Value))
	for i := range e.Value {
		tmp := make([]runtime.ArrayElement, len(e.Value[i]))
		for j, v := range e.Value[i] {
			val := v
			tmp[j] = *runtime.NewArrayElement(&val, step, false)
		}
		elements[i] = *runtime.NewArray("", e.Size2, tmp, step, false)
	}
	arr := runtime.NewArray2D(e.Name, e.Size1, e.Size2, elements, step, false)
	sn.CallStack.DeclareInCurrentFrame(arr)
	return nil
}

func (sn *Snapshot) applyVarChanged(e events.VarChanged, step int) error {
	variable, ok := sn.CallStack.GetVariableInCurrentFrame(e.Name)
	if !ok {
		return runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("variable %s not found", e.Name))
	}
	variable.ChangeValue(e.Value, step)
	return nil
}

func (sn *Snapshot) applyArrayElementChanged(e events.ArrayElementChanged, step int) error {
	arr, ok := sn.CallStack.GetArrayInCurrentFrame(e.Name)
	if !ok {
		return runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("array %s not found", e.Name))
	}
	return arr.ChangeElement(e.Ind, e.Value, step)
}

func (sn *Snapshot) applyArray2DElementChanged(e events.Array2DElementChanged, step int) error {
	arr, ok := sn.CallStack.GetArray2DInCurrentFrame(e.Name)
	if !ok {
		return runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("array2d %s not found", e.Name))
	}
	return arr.ChangeElement(e.Ind1, e.Ind2, e.Value, step)
}

func (sn *Snapshot) applyFunctionCall(e events.FunctionCall) error {
	if len(sn.CallStack.Frames) == 0 {
		return runtimeerrors.NewErrUnexpectedInternalError("no frames in call stack")
	}
	globalScope := sn.CallStack.Frames[0].Scopes[0]
	newFrame := runtime.NewStackFrame(e.Name, globalScope)
	sn.CallStack.PushFrame(newFrame)
	return nil
}

func (sn *Snapshot) applyFunctionReturn(e events.FunctionReturn) error {
	frame := sn.CallStack.GetCurrentFrame()
	if frame == nil {
		return runtimeerrors.NewErrUnexpectedInternalError("no current frame for function return")
	}
	if e.ReturnValue != nil {
		frame.SetReturnValue(*e.ReturnValue)
	}
	return sn.CallStack.PopFrame()
}

func (sn *Snapshot) applyLineChanged(e events.LineChanged) error {
	sn.Line = e.Line
	return nil
}

func (sn *Snapshot) Reset() {
	sn.CallStack = runtime.NewCallStack(sn.GlobalScope)
	sn.Line = 0
}

// Методы для чтения текущего состояния

func (sn *Snapshot) GetVariable(name string) (*runtime.Variable, bool) {
	return sn.CallStack.GetVariableInCurrentFrame(name)
}

func (sn *Snapshot) GetArray(name string) (*runtime.Array, bool) {
	return sn.CallStack.GetArrayInCurrentFrame(name)
}

func (sn *Snapshot) GetArray2D(name string) (*runtime.Array2D, bool) {
	return sn.CallStack.GetArray2DInCurrentFrame(name)
}

func (sn *Snapshot) GetCurrentLine() int {
	return sn.Line
}

func (sn *Snapshot) GetCurrentFrame() *runtime.StackFrame {
	return sn.CallStack.GetCurrentFrame()
}

func (sn *Snapshot) GetFramesCount() int {
	return sn.CallStack.FramesCount()
}

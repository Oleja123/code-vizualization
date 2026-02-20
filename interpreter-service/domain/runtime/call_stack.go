package runtime

import (
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

type CallStack struct {
	Frames []*StackFrame
}

func NewCallStack(globalScope *Scope) *CallStack {
	mainFrame := NewStackFrame("global", globalScope)
	return &CallStack{
		Frames: []*StackFrame{mainFrame},
	}
}

func (cs *CallStack) PushFrame(frame *StackFrame) {
	cs.Frames = append(cs.Frames, frame)
}

func (cs *CallStack) PopFrame() error {
	if len(cs.Frames) <= 1 {
		return runtimeerrors.NewErrUnexpectedInternalError("cannot pop main frame from call stack")
	}
	cs.Frames = cs.Frames[:len(cs.Frames)-1]
	return nil
}

func (cs *CallStack) GetCurrentFrame() *StackFrame {
	if len(cs.Frames) > 0 {
		return cs.Frames[len(cs.Frames)-1]
	}
	return nil
}

func (cs *CallStack) FramesCount() int {
	return len(cs.Frames)
}

func (cs *CallStack) DeclareInCurrentFrame(decl Declared) {
	cs.GetCurrentFrame().Declare(decl)
}

func (cs *CallStack) GetVariableInCurrentFrame(name string) (*Variable, bool) {
	return cs.GetCurrentFrame().GetVariable(name)
}

func (cs *CallStack) GetArrayInCurrentFrame(name string) (*Array, bool) {
	return cs.GetCurrentFrame().GetArray(name)
}

func (cs *CallStack) GetArray2DInCurrentFrame(name string) (*Array2D, bool) {
	return cs.GetCurrentFrame().GetArray2D(name)
}

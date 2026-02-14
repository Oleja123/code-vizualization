package runtime

import (
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

type CallStack struct {
	Frames []*StackFrame
}

func NewCallStack(globalScope *Scope) *CallStack {
	mainFrame := NewStackFrame("main", globalScope)
	return &CallStack{
		Frames: []*StackFrame{mainFrame},
	}
}

func (cs *CallStack) PushFrame(frame *StackFrame) {
	cs.Frames = append(cs.Frames, frame)
}

func (cs *CallStack) PopFrame() error {
	if len(cs.Frames) <= 1 {
		return runtimeerrors.NewUnexpectedInternalErr("cannot pop main frame from call stack")
	}
	cs.Frames = cs.Frames[:len(cs.Frames)-1]
	return nil
}

func (cs *CallStack) CurrentFrame() *StackFrame {
	if len(cs.Frames) > 0 {
		return cs.Frames[len(cs.Frames)-1]
	}
	return nil
}

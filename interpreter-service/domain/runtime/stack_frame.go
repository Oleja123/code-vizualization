package runtime

import (
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

type StackFrame struct {
	FuncName string
	Scopes   []*Scope
	ReturnValue *int
}

func NewStackFrame(funcName string, globalScope *Scope) *StackFrame {
	stackFrame := &StackFrame{FuncName: funcName, Scopes: make([]*Scope, 0)}
	stackFrame.Scopes = append(stackFrame.Scopes, globalScope)
	return stackFrame
}

func (sf *StackFrame) EnterScope() {
	scope := NewScope(sf.Scopes[len(sf.Scopes)-1])
	sf.Scopes = append(sf.Scopes, scope)
}

func (sf *StackFrame) ExitScope() error {
	if len(sf.Scopes) == 1 {
		return runtimeerrors.NewErrUndefinedBehavior("exit global scope in stack frame")
	}
	sf.Scopes = sf.Scopes[:len(sf.Scopes)-1]
	return nil
}

func (sf *StackFrame) GetCurrentScope() *Scope {
	if len(sf.Scopes) > 0 {
		return sf.Scopes[len(sf.Scopes)-1]
	}
	return nil
}

func (sf *StackFrame) Declare(decl Declared) {
	sf.GetCurrentScope().Declare(decl)
}

func (sf *StackFrame) SetReturnValue(val int) {
	if sf.ReturnValue == nil {
		sf.ReturnValue = new(int)
	}
	*sf.ReturnValue = val
}

func (sf *StackFrame) GetReturnValue() (*int, error) {
	if sf.ReturnValue == nil {
		return nil, runtimeerrors.NewErrUndefinedBehavior("return value not set")
	}
	return sf.ReturnValue, nil
}

func (sf *StackFrame) GetVariable(name string) (*Variable, bool) {
	current := sf.GetCurrentScope()
	for current != nil {
		if v, ok := current.Declarations.GetVariable(name); ok {
			return v, true
		}
		current = current.Parent
	}
	return nil, false
}

func (sf *StackFrame) GetArray(name string) (*Array, bool) {
	current := sf.GetCurrentScope()
	for current != nil {
		if arr, ok := current.Declarations.GetArray(name); ok {
			return arr, true
		}
		current = current.Parent
	}
	return nil, false
}

func (sf *StackFrame) GetArray2D(name string) (*Array2D, bool) {
	current := sf.GetCurrentScope()
	for current != nil {
		if arr, ok := current.Declarations.GetArray2D(name); ok {
			return arr, true
		}
		current = current.Parent
	}
	return nil, false
}

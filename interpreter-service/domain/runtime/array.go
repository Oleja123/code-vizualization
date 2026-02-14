package runtime

import (
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

type Array struct {
	Name  string
	Value []ArrayElement
}

func NewArray(name string, size int, value []ArrayElement, step int, isGlobal bool) Array {
	ret := Array{}
	ret.Name = name
	if value != nil {
		ret.Value = value
	} else {
		ret.Value = make([]ArrayElement, size)
		for i := range ret.Value {
			ret.Value[i] = NewArrayElement(nil, step, isGlobal)
		}
	}
	return ret
}

func (a *Array) ChangeElement(ind int, value int, step int) error {
	if ind < 0 || ind >= len(a.Value) {
		return runtimeerrors.NewErrArrayIndexOutOfBounds(ind, a.Name)
	}
	a.Value[ind].ChangeValue(value, step)
	return nil
}

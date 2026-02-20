package runtime

import (
	"fmt"

	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

type Array struct {
	Name   string
	Size   int
	Values []ArrayElement
}

func NewArray(name string, size int, value []ArrayElement, step int, isGlobal bool) *Array {
	ret := &Array{}
	ret.Name = name
	ret.Size = size
	if value != nil {
		ret.Values = value
	} else {
		ret.Values = make([]ArrayElement, size)
		for i := range ret.Values {
			ret.Values[i] = *NewArrayElement(nil, step, isGlobal)
		}
	}
	return ret
}

func (a *Array) ChangeElement(ind int, value int, step int) error {
	if ind < 0 || ind >= len(a.Values) {
		return runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("index out of bounds in array %s", a.Name))
	}
	a.Values[ind].ChangeValue(value, step)
	return nil
}

func (a *Array) GetElement(ind int) (*ArrayElement, error) {
	if ind < 0 || ind >= len(a.Values) {
		return nil, runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("index out of bounds in array %s", a.Name))
	}
	return &a.Values[ind], nil
}

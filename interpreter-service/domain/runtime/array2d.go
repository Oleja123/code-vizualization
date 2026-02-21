package runtime

import (
	"fmt"

	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

type Array2D struct {
	Name   string
	Size1  int
	Size2  int
	Values []Array
}

func NewArray2D(name string, size1, size2 int, value []Array, step int, isGlobal bool) *Array2D {
	ret := &Array2D{Name: name, Size1: size1, Size2: size2}
	if value != nil {
		ret.Values = value
		return ret
	}

	ret.Values = make([]Array, size1)
	for i := 0; i < size1; i++ {
		tmpArr := make([]ArrayElement, size2)
		for j := 0; j < size2; j++ {
			tmpArr[j] = *NewArrayElement(nil, step, isGlobal)
		}
		ret.Values[i] = *NewArray("", size2, tmpArr, step, isGlobal)
	}

	return ret
}

func (a *Array2D) ChangeElement(ind1, ind2 int, value int, step int) error {
	if ind1 < 0 || ind1 >= len(a.Values) || ind2 < 0 || ind2 >= a.Values[ind1].Size {
		return runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("index out of bounds in array2d %s", a.Name))
	}
	el, _ := a.Values[ind1].GetElement(ind2)
	el.ChangeValue(value, step)
	return nil
}

func (a *Array2D) GetElement(ind1, ind2 int) (int, error) {
	if ind1 < 0 || ind1 >= len(a.Values) || ind2 < 0 || ind2 >= a.Values[ind1].Size {
		return 0, runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("index out of bounds in array2d %s", a.Name))
	}
	el, _ := a.Values[ind1].GetElement(ind2)
	val, err := el.GetValue()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (a *Array2D) GetArray(ind int) (*Array, error) {
	if ind < 0 || ind >= len(a.Values) {
		return nil, runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("index out of bounds in array2d %s", a.Name))
	}
	return &a.Values[ind], nil
}

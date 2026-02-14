package runtime

import (
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

type Array2D struct {
	Name  string
	Value [][]ArrayElement
}

func NewArray2D(name string, size1, size2 int, value [][]ArrayElement, step int, isGlobal bool) Array2D {
	ret := Array2D{Name: name}
	if value != nil {
		ret.Value = value
		return ret
	}

	ret.Value = make([][]ArrayElement, size1)
	for i := 0; i < size1; i++ {
		ret.Value[i] = make([]ArrayElement, size2)
		for j := 0; j < size2; j++ {
			ret.Value[i][j] = NewArrayElement(nil, step, isGlobal)
		}
	}

	return ret
}

func (a *Array2D) ChangeElement(ind1, ind2 int, value int, step int) error {
	if ind1 < 0 || ind1 >= len(a.Value) || ind2 < 0 || ind2 >= len(a.Value[ind1]) {
		return runtimeerrors.NewErrArray2DIndexOutOfBounds(ind1, ind2, a.Name)
	}
	a.Value[ind1][ind2].ChangeValue(value, step)
	return nil
}

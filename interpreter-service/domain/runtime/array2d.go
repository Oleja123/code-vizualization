package runtime

import (
	"fmt"

	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

type Array2D struct {
	Name   string
	Size1  int
	Size2  int
	Values [][]ArrayElement
}

func NewArray2D(name string, size1, size2 int, value [][]ArrayElement, step int, isGlobal bool) *Array2D {
	ret := &Array2D{Name: name, Size1: size1, Size2: size2}
	if value != nil {
		ret.Values = value
		return ret
	}

	ret.Values = make([][]ArrayElement, size1)
	for i := 0; i < size1; i++ {
		ret.Values[i] = make([]ArrayElement, size2)
		for j := 0; j < size2; j++ {
			ret.Values[i][j] = *NewArrayElement(nil, step, isGlobal)
		}
	}

	return ret
}

func (a *Array2D) ChangeElement(ind1, ind2 int, value int, step int) error {
	if ind1 < 0 || ind1 >= len(a.Values) || ind2 < 0 || ind2 >= len(a.Values[ind1]) {
		return runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("index out of bounds in array2d %s", a.Name))
	}
	a.Values[ind1][ind2].ChangeValue(value, step)
	return nil
}

func (a *Array2D) GetElement(ind1, ind2 int) (int, error) {
	if ind1 < 0 || ind1 >= len(a.Values) || ind2 < 0 || ind2 >= len(a.Values[ind1]) {
		return 0, runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("index out of bounds in array2d %s", a.Name))
	}
	val, err := a.Values[ind1][ind2].GetValue()
	if err != nil {
		return 0, runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("undefined behavior: error getting element by index: [%d][%d]: %s", ind1, ind2, err))
	}
	return val, nil
}

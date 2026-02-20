package runtime

import (
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

type ArrayElement struct {
	Value       *int
	StepChanged int //для подстветки на фронте
}

func NewArrayElement(value *int, step int, isGlobal bool) *ArrayElement {
	var v *int
	if value != nil {
		v = value
	} else if isGlobal {
		val := 0
		v = &val
	}
	return &ArrayElement{Value: v, StepChanged: step}
}

func (ae *ArrayElement) ChangeValue(value int, step int) int {
	if ae.Value == nil {
		ae.Value = new(int)
	}
	*ae.Value = value
	ae.StepChanged = step
	return value
}

func (ae *ArrayElement) GetValue() (int, error) {
	if ae.Value == nil {
		return 0, runtimeerrors.NewErrUndefinedBehavior("getting an uninitialized array element")
	} else {
		return *ae.Value, nil
	}
}

package runtime

import (
	"fmt"

	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/domain/runtime/errors"
)

type Variable struct {
	Name        string
	Value       *int
	StepChanged int //для подстветки на фронте
}

func NewVariable(name string, value *int, step int, isGlobal bool) *Variable {
	var v *int
	if value != nil {
		v = value
	} else if isGlobal {
		val := 0
		v = &val
	}
	return &Variable{Name: name, Value: v, StepChanged: step}
}

func (v *Variable) ChangeValue(value int, step int) int {
	if v.Value == nil {
		v.Value = new(int)
	}
	*v.Value = value
	v.StepChanged = step
	return value
}

func (v *Variable) GetValue() (int, error) {
	if v.Value == nil {
		return 0, runtimeerrors.NewErrUndefinedBehavior(fmt.Sprintf("getting an uninitialized variable %s", v.Name))
	} else {
		return *v.Value, nil
	}
}

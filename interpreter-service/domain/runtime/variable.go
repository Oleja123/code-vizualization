package runtime

type Variable struct {
	Name        string
	Value       *int
	StepChanged int //для подстветки на фронте
}

func NewVariable(name string, value *int, step int, isGlobal bool) Variable {
	if value != nil {
		return Variable{Name: name, Value: value, StepChanged: step}
	} else {
		if isGlobal {
			val := 0
			return Variable{Name: name, Value: &val, StepChanged: step}
		} else {
			return Variable{Name: name, StepChanged: step}
		}
	}
}

func (v *Variable) ChangeValue(value int, step int) {
	v.Value = &value
	v.StepChanged = step
}

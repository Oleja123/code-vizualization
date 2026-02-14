package runtime

type ArrayElement struct {
	Value       *int
	StepChanged int //для подстветки на фронте
}

func NewArrayElement(value *int, step int, isGlobal bool) ArrayElement {
	if value != nil {
		return ArrayElement{Value: value, StepChanged: step}
	} else {
		if isGlobal {
			val := 0
			return ArrayElement{Value: &val, StepChanged: step}
		} else {
			return ArrayElement{StepChanged: step}
		}
	}
}

func (ae *ArrayElement) ChangeValue(value int, step int) {
	ae.Value = &value
	ae.StepChanged = step
}

package runtimeinterfaces

type Changeable interface {
	ChangeValue(value int, step int) int
	GetValue() (int, error)
}

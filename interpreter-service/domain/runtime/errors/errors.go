package runtimeerrors

import "fmt"

type ErrArrayIndexOutOfBounds struct {
	index     int
	arrayName string
}

func NewErrArrayIndexOutOfBounds(index int, arrayName string) error {
	return ErrArrayIndexOutOfBounds{
		index:     index,
		arrayName: arrayName,
	}
}

func (e ErrArrayIndexOutOfBounds) Error() string {
	return fmt.Sprintf("array %s index out of bounds: %d", e.arrayName, e.index)
}

type ErrArray2DIndexOutOfBounds struct {
	index1    int
	index2    int
	arrayName string
}

func NewErrArray2DIndexOutOfBounds(index1, index2 int, arrayName string) error {
	return ErrArray2DIndexOutOfBounds{
		index1:    index1,
		index2:    index2,
		arrayName: arrayName,
	}
}

func (e ErrArray2DIndexOutOfBounds) Error() string {
	return fmt.Sprintf("array2D %s index out of bounds: %d, %d", e.arrayName, e.index1, e.index2)
}

type UnexpectedInternalErr struct {
	reason string
}

func NewUnexpectedInternalErr(reason string) error {
	return UnexpectedInternalErr{reason: reason}
}

func (e UnexpectedInternalErr) Error() string {
	return fmt.Sprintf("unexpected internal error: %s", e.reason)
}

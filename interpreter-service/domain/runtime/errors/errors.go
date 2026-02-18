package runtimeerrors

import "fmt"

type ErrUndefinedBehavior struct {
	reason string
}

func NewErrUndefinedBehavior(reason string) error {
	return ErrUndefinedBehavior{
		reason: reason,
	}
}

func (e ErrUndefinedBehavior) Error() string {
	return fmt.Sprintf("undefined behavior: %s", e.reason)
}

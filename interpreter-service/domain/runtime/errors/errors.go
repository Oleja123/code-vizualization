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
type ErrUnexpectedInternalError struct {
	reason string
}

func NewErrUnexpectedInternalError(reason string) error {
	return ErrUnexpectedInternalError{
		reason: reason,
	}
}

func (e ErrUnexpectedInternalError) Error() string {
	return fmt.Sprintf("unexpected internal error: %s", e.reason)
}
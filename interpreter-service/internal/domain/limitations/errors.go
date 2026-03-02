package limitations

type LimitWhile int

const (
	MakingStep = iota
	AllocatingElement
)

type ErrLimitExceeded struct {
	reason LimitWhile
}

func NewErrLimitExceeded(reason LimitWhile) error {
	return ErrLimitExceeded{reason: reason}
}

func (e ErrLimitExceeded) Error() string {
	switch e.reason {
	case MakingStep:
		return "too many steps in program"
	case AllocatingElement:
		return "too many allocations in program"
	default:
		return "unknown limitation error"
	}
}

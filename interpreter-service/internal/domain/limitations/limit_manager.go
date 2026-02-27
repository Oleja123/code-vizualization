package limitations

type LimitManager struct {
	AllocatedElementsRemained int
	StepsRemained             int
}

func (lm *LimitManager) AllocateVariable() error {
	if lm.AllocatedElementsRemained <= 0 {
		return NewErrLimitExceeded(AllocatingElement)
	}
	lm.AllocatedElementsRemained -= 1
	return nil
}

func (lm *LimitManager) AllocateArray(size int) error {
	if lm.AllocatedElementsRemained < size {
		return NewErrLimitExceeded(AllocatingElement)
	}
	lm.AllocatedElementsRemained -= size
	return nil
}

func (lm *LimitManager) AllocateArray2D(size1 int, size2 int) error {
	if lm.AllocatedElementsRemained < size1*size2 {
		return NewErrLimitExceeded(AllocatingElement)
	}
	lm.AllocatedElementsRemained -= size1 * size2
	return nil
}

func (lm *LimitManager) MakeStep() error {
	if lm.StepsRemained <= 0 {
		return NewErrLimitExceeded(MakingStep)
	}
	lm.StepsRemained -= 1
	return nil
}

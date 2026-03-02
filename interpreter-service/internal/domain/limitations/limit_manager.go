package limitations

import "math"

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
	if size <= 0 {
		return NewErrLimitExceeded(AllocatingElement)
	}

	if lm.AllocatedElementsRemained < size {
		return NewErrLimitExceeded(AllocatingElement)
	}
	lm.AllocatedElementsRemained -= size
	return nil
}

func (lm *LimitManager) AllocateArray2D(size1 int, size2 int) error {
	if size1 <= 0 || size2 <= 0 {
		return NewErrLimitExceeded(AllocatingElement)
	}

	if size1 > math.MaxInt/size2 {
		return NewErrLimitExceeded(AllocatingElement)
	}

	totalSize := size1 * size2

	if lm.AllocatedElementsRemained < totalSize {
		return NewErrLimitExceeded(AllocatingElement)
	}
	lm.AllocatedElementsRemained -= totalSize
	return nil
}

func (lm *LimitManager) MakeStep() error {
	if lm.StepsRemained <= 0 {
		return NewErrLimitExceeded(MakingStep)
	}
	lm.StepsRemained -= 1
	return nil
}

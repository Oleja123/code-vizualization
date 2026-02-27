package limitations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLimitManager_AllocateVariable_Success(t *testing.T) {
	lm := &LimitManager{AllocatedElementsRemained: 2}

	err := lm.AllocateVariable()

	assert.NoError(t, err)
	assert.Equal(t, 1, lm.AllocatedElementsRemained)
}

func TestLimitManager_AllocateVariable_LimitExceeded(t *testing.T) {
	lm := &LimitManager{AllocatedElementsRemained: 0}

	err := lm.AllocateVariable()

	assert.Error(t, err)
	assert.EqualError(t, err, "too many allocations in program")
	assert.Equal(t, 0, lm.AllocatedElementsRemained)
}

func TestLimitManager_AllocateArray_Success(t *testing.T) {
	lm := &LimitManager{AllocatedElementsRemained: 5}

	err := lm.AllocateArray(3)

	assert.NoError(t, err)
	assert.Equal(t, 2, lm.AllocatedElementsRemained)
}

func TestLimitManager_AllocateArray_LimitExceeded(t *testing.T) {
	lm := &LimitManager{AllocatedElementsRemained: 2}

	err := lm.AllocateArray(3)

	assert.Error(t, err)
	assert.EqualError(t, err, "too many allocations in program")
	assert.Equal(t, 2, lm.AllocatedElementsRemained)
}

func TestLimitManager_AllocateArray2D_Success(t *testing.T) {
	lm := &LimitManager{AllocatedElementsRemained: 12}

	err := lm.AllocateArray2D(3, 4)

	assert.NoError(t, err)
	assert.Equal(t, 0, lm.AllocatedElementsRemained)
}

func TestLimitManager_AllocateArray2D_LimitExceeded(t *testing.T) {
	lm := &LimitManager{AllocatedElementsRemained: 11}

	err := lm.AllocateArray2D(3, 4)

	assert.Error(t, err)
	assert.EqualError(t, err, "too many allocations in program")
	assert.Equal(t, 11, lm.AllocatedElementsRemained)
}

func TestLimitManager_MakeStep_Success(t *testing.T) {
	lm := &LimitManager{StepsRemained: 2}

	err := lm.MakeStep()

	assert.NoError(t, err)
	assert.Equal(t, 1, lm.StepsRemained)
}

func TestLimitManager_MakeStep_LimitExceeded(t *testing.T) {
	lm := &LimitManager{StepsRemained: 0}

	err := lm.MakeStep()

	assert.Error(t, err)
	assert.EqualError(t, err, "too many steps in program")
	assert.Equal(t, 0, lm.StepsRemained)
}

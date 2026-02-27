package limitations

import (
	"math"
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

func TestLimitManager_AllocateArray_InvalidSize(t *testing.T) {
	testCases := []struct {
		name string
		size int
	}{
		{name: "zero size", size: 0},
		{name: "negative size", size: -1},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			lm := &LimitManager{AllocatedElementsRemained: 5}

			err := lm.AllocateArray(tt.size)

			assert.Error(t, err)
			assert.EqualError(t, err, "too many allocations in program")
			assert.Equal(t, 5, lm.AllocatedElementsRemained)
		})
	}
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

func TestLimitManager_AllocateArray2D_InvalidSizes(t *testing.T) {
	testCases := []struct {
		name  string
		size1 int
		size2 int
	}{
		{name: "zero first dimension", size1: 0, size2: 1},
		{name: "zero second dimension", size1: 1, size2: 0},
		{name: "negative first dimension", size1: -1, size2: 2},
		{name: "negative second dimension", size1: 2, size2: -1},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			lm := &LimitManager{AllocatedElementsRemained: 20}

			err := lm.AllocateArray2D(tt.size1, tt.size2)

			assert.Error(t, err)
			assert.EqualError(t, err, "too many allocations in program")
			assert.Equal(t, 20, lm.AllocatedElementsRemained)
		})
	}
}

func TestLimitManager_AllocateArray2D_Overflow(t *testing.T) {
	lm := &LimitManager{AllocatedElementsRemained: math.MaxInt}

	err := lm.AllocateArray2D(math.MaxInt, 2)

	assert.Error(t, err)
	assert.EqualError(t, err, "too many allocations in program")
	assert.Equal(t, math.MaxInt, lm.AllocatedElementsRemained)
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

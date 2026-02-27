package interpreter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
)

func TestInterpreter_NewLimitManagerPerRequest(t *testing.T) {
	testCases := []struct {
		name           string
		code           string
		maxAllocated   int
		maxSteps       int
		expectedResult int
	}{
		{
			name:           "variable declaration program",
			code:           "int main(){ int x = 1; return x; }",
			maxAllocated:   1,
			maxSteps:       10,
			expectedResult: 1,
		},
		{
			name:           "program with function call and parameters",
			code:           "int sum(int a, int b){ return a + b; } int main(){ return sum(2, 3); }",
			maxAllocated:   2,
			maxSteps:       20,
			expectedResult: 5,
		},
		{
			name:           "program without allocations",
			code:           "int main(){ return 42; }",
			maxAllocated:   0,
			maxSteps:       10,
			expectedResult: 42,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.New()
			program, err := conv.ParseToAST(tt.code)
			_ = err
			require.NotNil(t, program)

			runner1 := NewInterpreterWithLimits(tt.maxAllocated, tt.maxSteps)
			runner2 := NewInterpreterWithLimits(tt.maxAllocated, tt.maxSteps)

			result1, _, _, err1 := runner1.ExecuteProgram(program)
			require.NoError(t, err1)
			require.NotNil(t, result1)
			assert.Equal(t, tt.expectedResult, *result1)

			result2, _, _, err2 := runner2.ExecuteProgram(program)
			require.NoError(t, err2)
			require.NotNil(t, result2)
			assert.Equal(t, tt.expectedResult, *result2)
		})
	}
}

func TestInterpreter_StepLimit(t *testing.T) {
	testCases := []struct {
		name          string
		code          string
		maxAllocated  int
		maxSteps      int
		expectedErr   string
		expectedValue *int
	}{
		{
			name:          "fails immediately when no steps available",
			code:          "int main(){ return 1; }",
			maxAllocated:  10,
			maxSteps:      0,
			expectedErr:   "too many steps in program",
			expectedValue: nil,
		},
		{
			name:          "fails on second step",
			code:          "int main(){ int x = 1; return x; }",
			maxAllocated:  10,
			maxSteps:      1,
			expectedErr:   "too many steps in program",
			expectedValue: nil,
		},
		{
			name:         "succeeds with enough steps",
			code:         "int main(){ return 7; }",
			maxAllocated: 10,
			maxSteps:     10,
			expectedErr:  "",
			expectedValue: func() *int {
				v := 7
				return &v
			}(),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.New()
			program, err := conv.ParseToAST(tt.code)
			_ = err
			require.NotNil(t, program)

			runner := NewInterpreterWithLimits(tt.maxAllocated, tt.maxSteps)

			result, _, _, execErr := runner.ExecuteProgram(program)

			if tt.expectedErr == "" {
				require.NoError(t, execErr)
				require.NotNil(t, result)
				require.NotNil(t, tt.expectedValue)
				assert.Equal(t, *tt.expectedValue, *result)
				return
			}

			require.Error(t, execErr)
			assert.EqualError(t, execErr, tt.expectedErr)
		})
	}
}

func TestInterpreter_AllocationLimit(t *testing.T) {
	testCases := []struct {
		name          string
		code          string
		maxAllocated  int
		maxSteps      int
		expectedErr   string
		expectedValue *int
	}{
		{
			name:          "fails on first variable allocation",
			code:          "int main(){ int x = 1; return x; }",
			maxAllocated:  0,
			maxSteps:      10,
			expectedErr:   "too many allocations in program",
			expectedValue: nil,
		},
		{
			name:          "fails on array allocation size",
			code:          "int main(){ int arr[3]; return 0; }",
			maxAllocated:  2,
			maxSteps:      10,
			expectedErr:   "too many allocations in program",
			expectedValue: nil,
		},
		{
			name:          "fails on 2d array allocation size",
			code:          "int main(){ int matrix[2][2]; return 0; }",
			maxAllocated:  3,
			maxSteps:      10,
			expectedErr:   "too many allocations in program",
			expectedValue: nil,
		},
		{
			name:          "fails on function parameter allocation",
			code:          "int sum(int a, int b){ return a + b; } int main(){ return sum(1, 2); }",
			maxAllocated:  1,
			maxSteps:      20,
			expectedErr:   "too many allocations in program",
			expectedValue: nil,
		},
		{
			name:         "succeeds when no allocations are needed",
			code:         "int main(){ return 9; }",
			maxAllocated: 0,
			maxSteps:     10,
			expectedErr:  "",
			expectedValue: func() *int {
				v := 9
				return &v
			}(),
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.New()
			program, err := conv.ParseToAST(tt.code)
			_ = err
			require.NotNil(t, program)

			runner := NewInterpreterWithLimits(tt.maxAllocated, tt.maxSteps)

			result, _, _, execErr := runner.ExecuteProgram(program)

			if tt.expectedErr == "" {
				require.NoError(t, execErr)
				require.NotNil(t, result)
				require.NotNil(t, tt.expectedValue)
				assert.Equal(t, *tt.expectedValue, *result)
				return
			}

			require.Error(t, execErr)
			assert.EqualError(t, execErr, tt.expectedErr)
		})
	}
}

func TestInterpreter_InfiniteExecutionStopsByStepLimit(t *testing.T) {
	testCases := []struct {
		name         string
		code         string
		maxAllocated int
		maxSteps     int
		expectedErr  string
	}{
		{
			name:         "infinite recursion",
			code:         "int f(){ return f(); } int main(){ return f(); }",
			maxAllocated: 100,
			maxSteps:     50,
			expectedErr:  "too many steps in program",
		},
		{
			name:         "infinite while loop",
			code:         "int main(){ while(1){} return 0; }",
			maxAllocated: 10,
			maxSteps:     50,
			expectedErr:  "too many steps in program",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.New()
			program, err := conv.ParseToAST(tt.code)
			_ = err
			require.NotNil(t, program)

			runner := NewInterpreterWithLimits(tt.maxAllocated, tt.maxSteps)

			_, _, _, execErr := runner.ExecuteProgram(program)

			require.Error(t, execErr)
			assert.EqualError(t, execErr, tt.expectedErr)
		})
	}
}

func TestInterpreter_ReusedInstanceResetsStatePerExecuteProgram(t *testing.T) {
	conv := converter.New()

	program1, err1 := conv.ParseToAST("int main(){ return 1; }")
	_ = err1
	require.NotNil(t, program1)

	program2, err2 := conv.ParseToAST("int main(){ int x = 2; return x; }")
	_ = err2
	require.NotNil(t, program2)

	runner := NewInterpreterWithLimits(10, 20)

	result1, steps1, stepBegin1, execErr1 := runner.ExecuteProgram(program1)
	require.NoError(t, execErr1)
	require.NotNil(t, result1)
	assert.Equal(t, 1, *result1)
	assert.Equal(t, 0, stepBegin1)
	assert.NotEmpty(t, steps1)

	result2, steps2, stepBegin2, execErr2 := runner.ExecuteProgram(program2)
	require.NoError(t, execErr2)
	require.NotNil(t, result2)
	assert.Equal(t, 2, *result2)
	assert.Equal(t, 0, stepBegin2)
	assert.NotEmpty(t, steps2)
}

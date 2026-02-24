package eventdispatcher_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/eventdispatcher"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/application/interpreter"
	"github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/events"
	runtimeerrors "github.com/Oleja123/code-vizualization/interpreter-service/internal/domain/runtime/errors"
)

type expectedVariable struct {
	Exists bool
	Value  int
}

type expectedArray struct {
	Exists bool
	Values []int
}

type expectedArray2D struct {
	Exists bool
	Values [][]int
}

type expectedSnapshotState struct {
	FramesCount       int
	CurrentLine       int
	CurrentFrameScope int
	Error             string
	Variables         map[string]expectedVariable
	Arrays            map[string]expectedArray
	Arrays2D          map[string]expectedArray2D
}

func assertSnapshotState(t *testing.T, ed *eventdispatcher.EventDispatcher, expected expectedSnapshotState) {
	t.Helper()

	sn := ed.GetSnapshot()
	assert.Equal(t, expected.FramesCount, sn.GetFramesCount())
	assert.Equal(t, expected.CurrentLine, sn.GetCurrentLine())
	assert.Equal(t, expected.Error, sn.Error)

	currentFrame := sn.GetCurrentFrame()
	require.NotNil(t, currentFrame)
	assert.Equal(t, expected.CurrentFrameScope, len(currentFrame.Scopes))

	for name, expVar := range expected.Variables {
		v, exists := sn.GetVariable(name)
		if !expVar.Exists {
			assert.False(t, exists, "variable %s should not exist", name)
			continue
		}
		require.True(t, exists, "variable %s should exist", name)
		val, err := v.GetValue()
		require.NoError(t, err)
		assert.Equal(t, expVar.Value, val)
	}

	for name, expArr := range expected.Arrays {
		arr, exists := sn.GetArray(name)
		if !expArr.Exists {
			assert.False(t, exists, "array %s should not exist", name)
			continue
		}
		require.True(t, exists, "array %s should exist", name)
		if expArr.Values == nil {
			continue
		}
		require.Equal(t, len(expArr.Values), arr.Size)
		for i, expectedValue := range expArr.Values {
			element, err := arr.GetElement(i)
			require.NoError(t, err)
			value, err := element.GetValue()
			require.NoError(t, err)
			assert.Equal(t, expectedValue, value)
		}
	}

	for name, expArr2D := range expected.Arrays2D {
		arr2D, exists := sn.GetArray2D(name)
		if !expArr2D.Exists {
			assert.False(t, exists, "array2d %s should not exist", name)
			continue
		}
		require.True(t, exists, "array2d %s should exist", name)
		if expArr2D.Values == nil {
			continue
		}
		require.Equal(t, len(expArr2D.Values), arr2D.Size1)
		for i := range expArr2D.Values {
			require.Equal(t, len(expArr2D.Values[i]), arr2D.Size2)
			for j, expectedValue := range expArr2D.Values[i] {
				value, err := arr2D.GetElement(i, j)
				require.NoError(t, err)
				assert.Equal(t, expectedValue, value)
			}
		}
	}
}

func runDispatcherForCode(t *testing.T, code string) (*eventdispatcher.EventDispatcher, int) {
	t.Helper()

	conv := converter.New()
	program, convErr := conv.ParseToAST(code)
	if convErr != nil {
		t.Fatalf("parse error: %v", convErr)
	}
	require.NotNil(t, program)

	runner := interpreter.NewInterpreter()
	_, steps, stepBegin, err := runner.ExecuteProgram(program)
	if err != nil {
		var ubErr runtimeerrors.ErrUndefinedBehavior
		var rtErr runtimeerrors.ErrRuntime
		if !(errors.As(err, &ubErr) || errors.As(err, &rtErr)) {
			t.Fatalf("runtime error: %v", err)
		}
	}

	steps, stepBegin = compactDuplicateLineChangedSteps(steps, stepBegin)

	ed := eventdispatcher.NewEventDispatcher(stepBegin)
	ed.Steps = steps

	return ed, stepBegin
}

func compactDuplicateLineChangedSteps(steps []eventdispatcher.Step, stepBegin int) ([]eventdispatcher.Step, int) {
	if len(steps) == 0 {
		return steps, stepBegin
	}

	compacted := make([]eventdispatcher.Step, 0, len(steps))

	for i, step := range steps {
		line, singleLineChanged := extractSingleLineChanged(step)
		if singleLineChanged && len(compacted) > 0 && stepContainsLineChanged(compacted[len(compacted)-1], line) {
			if i < stepBegin {
				stepBegin--
			}
			continue
		}
		compacted = append(compacted, step)
	}

	if stepBegin < 0 {
		stepBegin = 0
	}
	if stepBegin >= len(compacted) && len(compacted) > 0 {
		stepBegin = len(compacted) - 1
	}

	return compacted, stepBegin
}

func extractSingleLineChanged(step eventdispatcher.Step) (int, bool) {
	if len(step.Events) != 1 {
		return 0, false
	}

	lineChanged, ok := step.Events[0].(events.LineChanged)
	if !ok {
		return 0, false
	}

	return lineChanged.Line, true
}

func stepContainsLineChanged(step eventdispatcher.Step, line int) bool {
	for _, event := range step.Events {
		lineChanged, ok := event.(events.LineChanged)
		if ok && lineChanged.Line == line {
			return true
		}
	}
	return false
}

func TestEventDispatcher_AssignmentProgramSnapshotByStep(t *testing.T) {
	code := `int main() {
	int x = 1;
	x = 2;
	return x;
}`

	ed, _ := runDispatcherForCode(t, code)

	checks := []struct {
		Step     int
		Expected expectedSnapshotState
	}{
		{
			Step: 0,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       2,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"x": {Exists: false},
				},
			},
		},
		{
			Step: 1,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       3,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"x": {Exists: true, Value: 1},
				},
			},
		},
		{
			Step: 2,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       4,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"x": {Exists: true, Value: 2},
				},
			},
		},
		{
			Step: 3,
			Expected: expectedSnapshotState{
				FramesCount:       1,
				CurrentLine:       -1,
				CurrentFrameScope: 1,
				Variables: map[string]expectedVariable{
					"x": {Exists: false},
				},
			},
		},
	}

	for _, check := range checks {
		require.NoError(t, ed.ApplyStep(check.Step))
		assertSnapshotState(t, ed, check.Expected)
	}
}

func TestEventDispatcher_Array2DSnapshotByStep(t *testing.T) {
	code := `int main() {
	int m[2][2] = {{1, 2}, {3, 4}};
	m[1][0] = 7;
	m[0][1] = m[1][0] + 1;
	return m[0][1];
}`

	ed, _ := runDispatcherForCode(t, code)

	checks := []struct {
		Step     int
		Expected expectedSnapshotState
	}{
		{
			Step: 0,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       2,
				CurrentFrameScope: 3,
				Arrays2D: map[string]expectedArray2D{
					"m": {Exists: false},
				},
			},
		},
		{
			Step: 1,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       3,
				CurrentFrameScope: 3,
				Arrays2D: map[string]expectedArray2D{
					"m": {Exists: true, Values: [][]int{{1, 2}, {3, 4}}},
				},
			},
		},
		{
			Step: 2,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       4,
				CurrentFrameScope: 3,
				Arrays2D: map[string]expectedArray2D{
					"m": {Exists: true, Values: [][]int{{1, 2}, {7, 4}}},
				},
			},
		},
		{
			Step: 3,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       5,
				CurrentFrameScope: 3,
				Arrays2D: map[string]expectedArray2D{
					"m": {Exists: true, Values: [][]int{{1, 8}, {7, 4}}},
				},
			},
		},
		{
			Step: 4,
			Expected: expectedSnapshotState{
				FramesCount:       1,
				CurrentLine:       -1,
				CurrentFrameScope: 1,
				Arrays2D: map[string]expectedArray2D{
					"m": {Exists: false},
				},
			},
		},
	}

	for _, check := range checks {
		require.NoError(t, ed.ApplyStep(check.Step))
		assertSnapshotState(t, ed, check.Expected)
	}
}

func TestEventDispatcher_VariableDeclareSnapshotImmutable(t *testing.T) {
	code := `int main() {
	int x = 10;
	x = 20;
	int y = 5;
	return x;
}`

	ed, _ := runDispatcherForCode(t, code)

	checks := []struct {
		Step     int
		Expected expectedSnapshotState
	}{
		{
			Step: 0,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       2,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"x": {Exists: false},
				},
			},
		},
		{
			Step: 1,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       3,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"x": {Exists: true, Value: 10},
				},
			},
		},
		{
			Step: 2,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       4,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"x": {Exists: true, Value: 20},
					"y": {Exists: false},
				},
			},
		},
		{
			Step: 3,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       5,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"x": {Exists: true, Value: 20},
					"y": {Exists: true, Value: 5},
				},
			},
		},
	}

	for _, check := range checks {
		require.NoError(t, ed.ApplyStep(check.Step), "step %d", check.Step)
		assertSnapshotState(t, ed, check.Expected)
	}
}

func TestEventDispatcher_GlobalVariablesSnapshotByStep(t *testing.T) {
	code := `int g = 5;
int h = 7;

int main() {
	int sum = g + h;
	g = sum + 1;
	return g;
}`

	ed, _ := runDispatcherForCode(t, code)

	checks := []struct {
		Step     int
		Expected expectedSnapshotState
	}{
		{
			Step: 0,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       5,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"g":   {Exists: true, Value: 5},
					"h":   {Exists: true, Value: 7},
					"sum": {Exists: false},
				},
			},
		},
		{
			Step: 1,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       6,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"g":   {Exists: true, Value: 5},
					"h":   {Exists: true, Value: 7},
					"sum": {Exists: true, Value: 12},
				},
			},
		},
		{
			Step: 2,
			Expected: expectedSnapshotState{
				FramesCount:       2,
				CurrentLine:       7,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"g":   {Exists: true, Value: 13},
					"h":   {Exists: true, Value: 7},
					"sum": {Exists: true, Value: 12},
				},
			},
		},
		{
			Step: 3,
			Expected: expectedSnapshotState{
				FramesCount:       1,
				CurrentLine:       -1,
				CurrentFrameScope: 1,
				Variables: map[string]expectedVariable{
					"g":   {Exists: true, Value: 13},
					"h":   {Exists: true, Value: 7},
					"sum": {Exists: false},
				},
			},
		},
	}

	for _, check := range checks {
		require.NoError(t, ed.ApplyStep(check.Step), "step %d", check.Step)
		assertSnapshotState(t, ed, check.Expected)
	}
}

func TestEventDispatcher_RecursiveFactorialSnapshot(t *testing.T) {
	code := `int factorial(int n) {
	if (n <= 1) {
		return 1;
	}
	return n * factorial(n - 1);
}

int main() {
	int result = factorial(3);
	return result;
}`

	ed, _ := runDispatcherForCode(t, code)

	checks := []struct {
		Step        int
		Expected    expectedSnapshotState
		Description string
	}{
		{
			Step:        0,
			Description: "entering main, before factorial call",
			Expected: expectedSnapshotState{
				FramesCount:       2, // global + main
				CurrentLine:       9,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"result": {Exists: false},
				},
			},
		},
		{
			Step:        1,
			Description: "entering factorial(3), checking condition",
			Expected: expectedSnapshotState{
				FramesCount:       3, // global + main + factorial(3)
				CurrentLine:       2,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"n": {Exists: true, Value: 3},
				},
			},
		},
		{
			Step:        2,
			Description: "in factorial(3), at return statement",
			Expected: expectedSnapshotState{
				FramesCount:       3,
				CurrentLine:       5,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"n": {Exists: true, Value: 3},
				},
			},
		},
		{
			Step:        3,
			Description: "entering factorial(2), checking condition",
			Expected: expectedSnapshotState{
				FramesCount:       4, // global + main + factorial(3) + factorial(2)
				CurrentLine:       2,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"n": {Exists: true, Value: 2},
				},
			},
		},
		{
			Step:        4,
			Description: "in factorial(2), at return statement",
			Expected: expectedSnapshotState{
				FramesCount:       4,
				CurrentLine:       5,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"n": {Exists: true, Value: 2},
				},
			},
		},
		{
			Step:        5,
			Description: "entering factorial(1), checking condition",
			Expected: expectedSnapshotState{
				FramesCount:       5, // global + main + factorial(3) + factorial(2) + factorial(1)
				CurrentLine:       2,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"n": {Exists: true, Value: 1},
				},
			},
		},
		{
			Step:        6,
			Description: "in factorial(1), condition is true, returning 1",
			Expected: expectedSnapshotState{
				FramesCount:       5,
				CurrentLine:       3,
				CurrentFrameScope: 4, // if block creates additional scope
				Variables: map[string]expectedVariable{
					"n": {Exists: true, Value: 1},
				},
			},
		},
		{
			Step:        7,
			Description: "returned from factorial(1) to factorial(2)",
			Expected: expectedSnapshotState{
				FramesCount:       4,
				CurrentLine:       5,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"n": {Exists: true, Value: 2},
				},
			},
		},
		{
			Step:        8,
			Description: "returned from factorial(2) to factorial(3)",
			Expected: expectedSnapshotState{
				FramesCount:       3,
				CurrentLine:       5,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"n": {Exists: true, Value: 3},
				},
			},
		},
		{
			Step:        9,
			Description: "returned from all factorial calls to main, result assigned",
			Expected: expectedSnapshotState{
				FramesCount:       2, // all factorial frames popped
				CurrentLine:       10,
				CurrentFrameScope: 3,
				Variables: map[string]expectedVariable{
					"result": {Exists: true, Value: 6},
				},
			},
		},
		{
			Step:        10,
			Description: "program finished",
			Expected: expectedSnapshotState{
				FramesCount:       1, // only global frame remains
				CurrentLine:       -1,
				CurrentFrameScope: 1,
				Variables: map[string]expectedVariable{
					"result": {Exists: false},
				},
			},
		},
	}

	for _, check := range checks {
		require.NoError(t, ed.ApplyStep(check.Step), "step %d: %s", check.Step, check.Description)
		assertSnapshotState(t, ed, check.Expected)
	}
}

func TestEventDispatcher_RollbackRecursiveFactorial(t *testing.T) {
	code := `int factorial(int n) {
	if (n <= 1) {
		return 1;
	}
	return n * factorial(n - 1);
}

int main() {
	int result = factorial(3);
	return result;
}`

	ed, _ := runDispatcherForCode(t, code)

	// Apply all steps forward to the end
	require.NoError(t, ed.ApplyStep(10))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       -1,
		CurrentFrameScope: 1,
		Variables: map[string]expectedVariable{
			"result": {Exists: false},
		},
	})

	// Rollback to step 9 (in main, result = 6)
	require.NoError(t, ed.ApplyStep(9))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       10,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"result": {Exists: true, Value: 6},
		},
	})

	// Rollback to step 5 (in factorial(1))
	require.NoError(t, ed.ApplyStep(5))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       5, // global + main + factorial(3) + factorial(2) + factorial(1)
		CurrentLine:       2,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"n": {Exists: true, Value: 1},
		},
	})

	// Rollback to step 3 (in factorial(2))
	require.NoError(t, ed.ApplyStep(3))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       4, // global + main + factorial(3) + factorial(2)
		CurrentLine:       2,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"n": {Exists: true, Value: 2},
		},
	})

	// Rollback to step 1 (in factorial(3), first time)
	require.NoError(t, ed.ApplyStep(1))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       3,
		CurrentLine:       2,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"n": {Exists: true, Value: 3},
		},
	})

	// Move forward to step 6 (in factorial(1), returning 1)
	require.NoError(t, ed.ApplyStep(6))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       5,
		CurrentLine:       3,
		CurrentFrameScope: 4,
		Variables: map[string]expectedVariable{
			"n": {Exists: true, Value: 1},
		},
	})
}

func TestEventDispatcher_RollbackArrayModification(t *testing.T) {
	code := `int main() {
	int arr[3] = {1, 2, 3};
	arr[0] = 10;
	arr[1] = 20;
	arr[2] = 30;
	return arr[0];
}`

	ed, _ := runDispatcherForCode(t, code)

	// Apply all steps forward to the end
	require.NoError(t, ed.ApplyStep(5))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       -1,
		CurrentFrameScope: 1,
		Arrays: map[string]expectedArray{
			"arr": {Exists: false},
		},
	})

	// Rollback to step 4 (arr[2] = 30)
	require.NoError(t, ed.ApplyStep(4))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       6,
		CurrentFrameScope: 3,
		Arrays: map[string]expectedArray{
			"arr": {Exists: true, Values: []int{10, 20, 30}},
		},
	})

	// Rollback to step 3 (arr[1] = 20)
	require.NoError(t, ed.ApplyStep(3))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       5,
		CurrentFrameScope: 3,
		Arrays: map[string]expectedArray{
			"arr": {Exists: true, Values: []int{10, 20, 3}},
		},
	})

	// Rollback to step 2 (arr[0] = 10)
	require.NoError(t, ed.ApplyStep(2))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       4,
		CurrentFrameScope: 3,
		Arrays: map[string]expectedArray{
			"arr": {Exists: true, Values: []int{10, 2, 3}},
		},
	})

	// Rollback to step 1 (arr declared with initial values)
	require.NoError(t, ed.ApplyStep(1))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       3,
		CurrentFrameScope: 3,
		Arrays: map[string]expectedArray{
			"arr": {Exists: true, Values: []int{1, 2, 3}},
		},
	})

	// Move forward again to step 4
	require.NoError(t, ed.ApplyStep(4))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       6,
		CurrentFrameScope: 3,
		Arrays: map[string]expectedArray{
			"arr": {Exists: true, Values: []int{10, 20, 30}},
		},
	})
}

func TestEventDispatcher_RollbackArray2D(t *testing.T) {
	code := `int main() {
	int m[2][2] = {{1, 2}, {3, 4}};
	m[0][0] = 10;
	m[1][1] = 40;
	return m[0][0];
}`

	ed, _ := runDispatcherForCode(t, code)

	// Apply all steps forward
	require.NoError(t, ed.ApplyStep(4))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       -1,
		CurrentFrameScope: 1,
		Arrays2D: map[string]expectedArray2D{
			"m": {Exists: false},
		},
	})

	// Rollback to step 3 (m[1][1] = 40)
	require.NoError(t, ed.ApplyStep(3))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       5,
		CurrentFrameScope: 3,
		Arrays2D: map[string]expectedArray2D{
			"m": {Exists: true, Values: [][]int{{10, 2}, {3, 40}}},
		},
	})

	// Rollback to step 2 (m[0][0] = 10)
	require.NoError(t, ed.ApplyStep(2))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       4,
		CurrentFrameScope: 3,
		Arrays2D: map[string]expectedArray2D{
			"m": {Exists: true, Values: [][]int{{10, 2}, {3, 4}}},
		},
	})

	// Rollback to step 1 (initial values)
	require.NoError(t, ed.ApplyStep(1))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       3,
		CurrentFrameScope: 3,
		Arrays2D: map[string]expectedArray2D{
			"m": {Exists: true, Values: [][]int{{1, 2}, {3, 4}}},
		},
	})

	// Move forward to step 3 again
	require.NoError(t, ed.ApplyStep(3))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       5,
		CurrentFrameScope: 3,
		Arrays2D: map[string]expectedArray2D{
			"m": {Exists: true, Values: [][]int{{10, 2}, {3, 40}}},
		},
	})
}

func TestEventDispatcher_ApplyStepOutOfRange(t *testing.T) {
	ed := eventdispatcher.NewEventDispatcher(0)
	ed.Steps = []eventdispatcher.Step{}

	err := ed.ApplyStep(0)
	assert.Error(t, err)
	assert.Equal(t, -1, ed.GetCurrentStep())

	err = ed.ApplyStep(-1)
	assert.Error(t, err)
	assert.Equal(t, -1, ed.GetCurrentStep())
}

func TestEventDispatcher_ApplyStepRollback(t *testing.T) {
	code := `int main() {
	int x = 1;
	x = 2;
	return x;
}`

	ed, _ := runDispatcherForCode(t, code)

	require.NoError(t, ed.ApplyStep(2))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       4,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"x": {Exists: true, Value: 2},
		},
	})

	require.NoError(t, ed.ApplyStep(0))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       2,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"x": {Exists: false},
		},
	})
}

func TestEventDispatcher_StepBeginOffset(t *testing.T) {
	code := `int main() {
	int x = 1;
	return x;
}`

	ed, stepBegin := runDispatcherForCode(t, code)

	require.NoError(t, ed.ApplyStep(0))
	assert.Equal(t, stepBegin, ed.GetCurrentStep())
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       2,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"x": {Exists: false},
		},
	})

	require.NoError(t, ed.ApplyStep(1))
	assert.Equal(t, stepBegin+1, ed.GetCurrentStep())
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       3,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"x": {Exists: true, Value: 1},
		},
	})
}

func TestEventDispatcher_StepWithMultipleEvents(t *testing.T) {
	code := `int main() {
	int x = 10;
	return x;
}`

	ed, stepBegin := runDispatcherForCode(t, code)

	require.Greater(t, len(ed.Steps[stepBegin].Events), 1)
	require.NoError(t, ed.ApplyStep(0))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       2,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"x": {Exists: false},
		},
	})
	assert.Equal(t, stepBegin, ed.GetCurrentStep())
}

func TestEventDispatcher_ApplyStepStopsOnError(t *testing.T) {
	code := `int main() {
	int x = 1;
	return x;
}`

	ed, stepBegin := runDispatcherForCode(t, code)

	require.NoError(t, ed.ApplyStep(1))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       3,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"x": {Exists: true, Value: 1},
		},
	})

	baseStepsCount := len(ed.Steps)
	badStepIndex := baseStepsCount - stepBegin
	ed.Steps = append(ed.Steps, eventdispatcher.Step{Events: []events.Event{
		events.VarChanged{Name: "missing", Value: 1},
	}})

	err := ed.ApplyStep(badStepIndex)
	assert.Error(t, err)
	assert.Equal(t, baseStepsCount-1, ed.GetCurrentStep())
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       -1,
		CurrentFrameScope: 1,
		Variables: map[string]expectedVariable{
			"x":       {Exists: false},
			"missing": {Exists: false},
		},
	})
}

func TestEventDispatcher_RollbackZigzag(t *testing.T) {
	code := `int main() {
	int x = 1;
	x = 2;
	x = 3;
	return x;
}`

	ed, _ := runDispatcherForCode(t, code)

	require.NoError(t, ed.ApplyStep(3))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       5,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"x": {Exists: true, Value: 3},
		},
	})

	require.NoError(t, ed.ApplyStep(1))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       3,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"x": {Exists: true, Value: 1},
		},
	})

	require.NoError(t, ed.ApplyStep(2))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       4,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"x": {Exists: true, Value: 2},
		},
	})

	require.NoError(t, ed.ApplyStep(3))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       5,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"x": {Exists: true, Value: 3},
		},
	})
}

func TestEventDispatcher_StepBeginRollback(t *testing.T) {
	code := `int main() {
	int x = 1;
	x = 2;
	return x;
}`

	ed, stepBegin := runDispatcherForCode(t, code)

	require.NoError(t, ed.ApplyStep(1))
	assert.Equal(t, stepBegin+1, ed.GetCurrentStep())
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       3,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"x": {Exists: true, Value: 1},
		},
	})

	require.NoError(t, ed.ApplyStep(0))
	assert.Equal(t, stepBegin, ed.GetCurrentStep())
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       2,
		CurrentFrameScope: 3,
		Variables: map[string]expectedVariable{
			"x": {Exists: false},
		},
	})
}

func TestEventDispatcher_GetStep(t *testing.T) {
	ed := eventdispatcher.NewEventDispatcher(0)
	ed.Steps = []eventdispatcher.Step{
		{StepNumber: 10, Events: []events.Event{events.LineChanged{Line: 1}}},
	}

	step, err := ed.GetStep(0)
	require.NoError(t, err)
	assert.Equal(t, 10, step.StepNumber)
	require.Len(t, step.Events, 1)

	_, err = ed.GetStep(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid step index")

	_, err = ed.GetStep(1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid step index")
}

func TestEventDispatcher_ApplyStepIdempotent(t *testing.T) {
	value := 10
	ed := eventdispatcher.NewEventDispatcher(0)
	ed.Steps = []eventdispatcher.Step{
		{Events: []events.Event{events.LineChanged{Line: 1}}},
		{Events: []events.Event{events.DeclareVar{Name: "x", Value: &value}, events.LineChanged{Line: 2}}},
	}

	require.NoError(t, ed.ApplyStep(1))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       2,
		CurrentFrameScope: 1,
		Variables: map[string]expectedVariable{
			"x": {Exists: true, Value: 10},
		},
	})
	assert.Equal(t, 1, ed.GetCurrentStep())

	require.NoError(t, ed.ApplyStep(1))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       2,
		CurrentFrameScope: 1,
		Variables: map[string]expectedVariable{
			"x": {Exists: true, Value: 10},
		},
	})
	assert.Equal(t, 1, ed.GetCurrentStep())
}

func TestEventDispatcher_ApplyStepEmptyStepAdvancesIndex(t *testing.T) {
	ed := eventdispatcher.NewEventDispatcher(0)
	ed.Steps = []eventdispatcher.Step{
		{Events: []events.Event{events.LineChanged{Line: 7}}},
		{Events: nil},
	}

	require.NoError(t, ed.ApplyStep(0))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       7,
		CurrentFrameScope: 1,
	})
	assert.Equal(t, 0, ed.GetCurrentStep())

	require.NoError(t, ed.ApplyStep(1))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       7,
		CurrentFrameScope: 1,
	})
	assert.Equal(t, 1, ed.GetCurrentStep())
}

func TestEventDispatcher_ApplyStepPartialFailureAndRollbackRecovery(t *testing.T) {
	initial := 1
	ed := eventdispatcher.NewEventDispatcher(0)
	ed.Steps = []eventdispatcher.Step{
		{Events: []events.Event{events.DeclareVar{Name: "x", Value: &initial}, events.LineChanged{Line: 1}}},
		{Events: []events.Event{events.LineChanged{Line: 2}}},
		{Events: []events.Event{events.VarChanged{Name: "x", Value: 5}, events.VarChanged{Name: "missing", Value: 1}}},
	}

	require.NoError(t, ed.ApplyStep(1))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       2,
		CurrentFrameScope: 1,
		Variables: map[string]expectedVariable{
			"x":       {Exists: true, Value: 1},
			"missing": {Exists: false},
		},
	})
	assert.Equal(t, 1, ed.GetCurrentStep())

	err := ed.ApplyStep(2)
	assert.Error(t, err)
	assert.Equal(t, 1, ed.GetCurrentStep())
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       2,
		CurrentFrameScope: 1,
		Variables: map[string]expectedVariable{
			"x":       {Exists: true, Value: 5},
			"missing": {Exists: false},
		},
	})

	require.NoError(t, ed.ApplyStep(0))
	assert.Equal(t, 0, ed.GetCurrentStep())
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       1,
		CurrentFrameScope: 1,
		Variables: map[string]expectedVariable{
			"x": {Exists: true, Value: 1},
		},
	})
}

func TestEventDispatcher_StepBeginNegativeExternalIndex(t *testing.T) {
	ed := eventdispatcher.NewEventDispatcher(2)
	ed.Steps = []eventdispatcher.Step{
		{Events: []events.Event{events.LineChanged{Line: 10}}},
		{Events: []events.Event{events.LineChanged{Line: 11}}},
		{Events: []events.Event{events.LineChanged{Line: 12}}},
	}

	require.NoError(t, ed.ApplyStep(-1))
	assert.Equal(t, 1, ed.GetCurrentStep())
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       11,
		CurrentFrameScope: 1,
	})
}

func TestEventDispatcher_UndefinedBehaviorByStep(t *testing.T) {
	code := `int main() {
	int x;
	return x;
}`

	ed, stepBegin := runDispatcherForCode(t, code)
	lastExternalStep := ed.GetStepsCount() - stepBegin - 1
	require.GreaterOrEqual(t, lastExternalStep, 0)

	require.NoError(t, ed.ApplyStep(lastExternalStep))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       3,
		CurrentFrameScope: 1,
		Error:             "undefined behavior: getting an uninitialized variable x",
	})
}

func TestEventDispatcher_RuntimeErrorByStep(t *testing.T) {
	code := `int main() {
	int x = 1;
	return x / 0;
}`

	ed, stepBegin := runDispatcherForCode(t, code)
	lastExternalStep := ed.GetStepsCount() - stepBegin - 1
	require.GreaterOrEqual(t, lastExternalStep, 0)

	require.NoError(t, ed.ApplyStep(lastExternalStep))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       3,
		CurrentFrameScope: 1,
		Error:             "runtime error: division by zero",
	})
}

func TestEventDispatcher_RollbackUndefinedBehavior(t *testing.T) {
	code := `int main() {
	int x;
	return x;
}`

	ed, stepBegin := runDispatcherForCode(t, code)
	lastExternalStep := ed.GetStepsCount() - stepBegin - 1
	require.GreaterOrEqual(t, lastExternalStep, 1)

	require.NoError(t, ed.ApplyStep(lastExternalStep))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       3,
		CurrentFrameScope: 1,
		Error:             "undefined behavior: getting an uninitialized variable x",
	})

	require.NoError(t, ed.ApplyStep(lastExternalStep-1))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       3,
		CurrentFrameScope: 3,
		Error:             "",
	})

	require.NoError(t, ed.ApplyStep(lastExternalStep))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       3,
		CurrentFrameScope: 1,
		Error:             "undefined behavior: getting an uninitialized variable x",
	})
}

func TestEventDispatcher_RollbackRuntimeError(t *testing.T) {
	code := `int main() {
	int x = 1;
	return x / 0;
}`

	ed, stepBegin := runDispatcherForCode(t, code)
	lastExternalStep := ed.GetStepsCount() - stepBegin - 1
	require.GreaterOrEqual(t, lastExternalStep, 1)

	require.NoError(t, ed.ApplyStep(lastExternalStep))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       3,
		CurrentFrameScope: 1,
		Error:             "runtime error: division by zero",
	})

	require.NoError(t, ed.ApplyStep(lastExternalStep-1))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       2,
		CurrentLine:       3,
		CurrentFrameScope: 3,
		Error:             "",
		Variables: map[string]expectedVariable{
			"x": {Exists: true, Value: 1},
		},
	})

	require.NoError(t, ed.ApplyStep(lastExternalStep))
	assertSnapshotState(t, ed, expectedSnapshotState{
		FramesCount:       1,
		CurrentLine:       3,
		CurrentFrameScope: 1,
		Error:             "runtime error: division by zero",
	})
}

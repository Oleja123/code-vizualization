package eventdispatcher_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/application/eventdispatcher"
	"github.com/Oleja123/code-vizualization/interpreter-service/application/interpreter"
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
	Variables         map[string]expectedVariable
	Arrays            map[string]expectedArray
	Arrays2D          map[string]expectedArray2D
}

func assertSnapshotState(t *testing.T, ed *eventdispatcher.EventDispatcher, expected expectedSnapshotState) {
	t.Helper()

	sn := ed.GetSnapshot()
	assert.Equal(t, expected.FramesCount, sn.GetFramesCount())
	assert.Equal(t, expected.CurrentLine, sn.GetCurrentLine())

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
		t.Fatalf("runtime error: %v", err)
	}

	ed := eventdispatcher.NewEventDispatcher(runner.GlobalScope, stepBegin)
	ed.Steps = steps

	return ed, stepBegin
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
				CurrentFrameScope: 2,
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
				CurrentFrameScope: 2,
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
				CurrentFrameScope: 2,
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
				CurrentFrameScope: 2,
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
				CurrentFrameScope: 2,
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
				CurrentFrameScope: 2,
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
				CurrentFrameScope: 2,
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
				CurrentFrameScope: 2,
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
				CurrentFrameScope: 2,
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
				CurrentFrameScope: 2,
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
				CurrentFrameScope: 2,
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

package interpreter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/interpreter-service/application/eventdispatcher"
	"github.com/Oleja123/code-vizualization/interpreter-service/domain/events"
)

type normalizedStep struct {
	Events []string
}

func runCodeWithSteps(t *testing.T, code string) (*int, []eventdispatcher.Step, int) {
	t.Helper()

	conv := converter.New()
	program, convErr := conv.ParseToAST(code)
	if convErr != nil {
		t.Fatalf("parse error: %v", convErr)
	}
	require.NotNil(t, program)

	runner := NewInterpreter()
	result, steps, stepBegin, err := runner.ExecuteProgram(program)
	if err != nil {
		t.Fatalf("runtime error: %v", err)
	}

	return result, steps, stepBegin
}

func runCodeWithStepsAllowError(t *testing.T, code string) (*int, []eventdispatcher.Step, int, error) {
	t.Helper()

	conv := converter.New()
	program, convErr := conv.ParseToAST(code)
	if convErr != nil {
		t.Fatalf("parse error: %v", convErr)
	}
	require.NotNil(t, program)

	runner := NewInterpreter()
	return runner.ExecuteProgram(program)
}

func normalizeEvent(event events.Event) string {
	switch e := event.(type) {
	case events.EnterScope:
		return "EnterScope"
	case events.ExitScope:
		return "ExitScope"
	case events.DeclareVar:
		return fmt.Sprintf("DeclareVar(name=%s,global=%t)", e.Name, e.IsGlobal)
	case events.DeclareArray:
		return fmt.Sprintf("DeclareArray(name=%s,size=%d,global=%t)", e.Name, e.Size, e.IsGlobal)
	case events.DeclareArray2D:
		return fmt.Sprintf("DeclareArray2D(name=%s,size1=%d,size2=%d,global=%t)", e.Name, e.Size1, e.Size2, e.IsGlobal)
	case events.VarChanged:
		return fmt.Sprintf("VarChanged(name=%s,value=%d)", e.Name, e.Value)
	case events.ArrayElementChanged:
		return fmt.Sprintf("ArrayElementChanged(name=%s,ind=%d,value=%d)", e.Name, e.Ind, e.Value)
	case events.Array2DElementChanged:
		return fmt.Sprintf("Array2DElementChanged(name=%s,ind1=%d,ind2=%d,value=%d)", e.Name, e.Ind1, e.Ind2, e.Value)
	case events.FunctionCall:
		return fmt.Sprintf("FunctionCall(name=%s)", e.Name)
	case events.FunctionReturn:
		if e.ReturnValue == nil {
			return fmt.Sprintf("FunctionReturn(name=%s,value=nil)", e.Name)
		}
		return fmt.Sprintf("FunctionReturn(name=%s,value=%d)", e.Name, *e.ReturnValue)
	case events.LineChanged:
		return fmt.Sprintf("LineChanged(line=%d)", e.Line)
	case events.UndefinedBehavior:
		return fmt.Sprintf("UndefinedBehavior(message=%s)", e.Message)
	default:
		return fmt.Sprintf("UnknownEvent(%T)", event)
	}
}

func normalizeSteps(steps []eventdispatcher.Step) []normalizedStep {
	normalized := make([]normalizedStep, 0, len(steps))
	for _, step := range steps {
		normalizedEvents := make([]string, 0, len(step.Events))
		for _, event := range step.Events {
			normalizedEvents = append(normalizedEvents, normalizeEvent(event))
		}
		normalized = append(normalized, normalizedStep{Events: normalizedEvents})
	}
	return normalized
}

func TestInterpreterSteps_MainLifecycleEvents(t *testing.T) {
	code := `int main() {
	return 1;
}`

	result, steps, stepBegin := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 1, *result)
	require.Len(t, steps, 2)
	assert.GreaterOrEqual(t, stepBegin, 0)
	assert.Less(t, stepBegin, len(steps))

	expectedSteps := []normalizedStep{
		{Events: []string{
			"FunctionCall(name=main)",
			"EnterScope",
			"LineChanged(line=2)",
		}},
		{Events: []string{
			"ExitScope",
			"FunctionReturn(name=main,value=1)",
			"LineChanged(line=-1)",
		}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_AssignmentEmitsVarChanged(t *testing.T) {
	code := `int main() {
	int x = 1;
	x = 2;
	return x;
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 2, *result)
	require.Len(t, steps, 4)

	expectedSteps := []normalizedStep{
		{Events: []string{
			"FunctionCall(name=main)",
			"EnterScope",
			"LineChanged(line=2)",
		}},
		{Events: []string{
			"DeclareVar(name=x,global=false)",
			"LineChanged(line=3)",
		}},
		{Events: []string{
			"VarChanged(name=x,value=2)",
			"LineChanged(line=4)",
		}},
		{Events: []string{
			"ExitScope",
			"FunctionReturn(name=main,value=2)",
			"LineChanged(line=-1)",
		}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_ArraysAndElementChangesWithCondition(t *testing.T) {
	code := `int main() {
	int arr[3] = {1, 2, 3};
	arr[1] = 5;
	if (arr[1] > arr[0]) {
		arr[2] = arr[1] + arr[0];
	}
	return arr[2];
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 6, *result)
	require.Len(t, steps, 6)

	expectedSteps := []normalizedStep{
		{Events: []string{
			"FunctionCall(name=main)",
			"EnterScope",
			"LineChanged(line=2)",
		}},
		{Events: []string{
			"DeclareArray(name=arr,size=3,global=false)",
			"LineChanged(line=3)",
		}},
		{Events: []string{
			"ArrayElementChanged(name=arr,ind=1,value=5)",
			"LineChanged(line=4)",
		}},
		{Events: []string{
			"EnterScope",
			"LineChanged(line=5)",
		}},
		{Events: []string{
			"ArrayElementChanged(name=arr,ind=2,value=6)",
			"ExitScope",
			"LineChanged(line=7)",
		}},
		{Events: []string{
			"ExitScope",
			"FunctionReturn(name=main,value=6)",
			"LineChanged(line=-1)",
		}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_WhileLoopWithIfAndVarChanges(t *testing.T) {
	code := `int main() {
	int i = 0;
	int sum = 0;
	while (i < 4) {
		if (i % 2 == 0) {
			sum += i;
		}
		i = i + 1;
	}
	return sum;
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 2, *result)
	require.Len(t, steps, 19)

	expectedSteps := []normalizedStep{
		{Events: []string{"FunctionCall(name=main)", "EnterScope", "LineChanged(line=2)"}},
		{Events: []string{"DeclareVar(name=i,global=false)", "LineChanged(line=3)"}},
		{Events: []string{"DeclareVar(name=sum,global=false)", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"EnterScope", "LineChanged(line=6)"}},
		{Events: []string{"VarChanged(name=sum,value=0)", "ExitScope", "LineChanged(line=8)"}},
		{Events: []string{"VarChanged(name=i,value=1)", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"LineChanged(line=8)"}},
		{Events: []string{"VarChanged(name=i,value=2)", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"EnterScope", "LineChanged(line=6)"}},
		{Events: []string{"VarChanged(name=sum,value=2)", "ExitScope", "LineChanged(line=8)"}},
		{Events: []string{"VarChanged(name=i,value=3)", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"LineChanged(line=8)"}},
		{Events: []string{"VarChanged(name=i,value=4)", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"LineChanged(line=10)"}},
		{Events: []string{"ExitScope", "FunctionReturn(name=main,value=2)", "LineChanged(line=-1)"}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_NestedFunctionCallOrder(t *testing.T) {
	code := `int foo() {
	return 7;
}
int main() {
	return foo();
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 7, *result)
	require.Len(t, steps, 3)

	expectedSteps := []normalizedStep{
		{Events: []string{
			"FunctionCall(name=main)",
			"EnterScope",
			"LineChanged(line=5)",
		}},
		{Events: []string{
			"FunctionCall(name=foo)",
			"EnterScope",
			"LineChanged(line=2)",
		}},
		{Events: []string{
			"ExitScope",
			"FunctionReturn(name=foo,value=7)",
			"ExitScope",
			"FunctionReturn(name=main,value=7)",
			"LineChanged(line=-1)",
		}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_ForLoopWithVarChanges(t *testing.T) {
	code := `int main() {
	int sum = 0;
	for (int i = 0; i < 3; i = i + 1) {
		sum += i;
	}
	return sum;
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 3, *result)
	require.Len(t, steps, 11)

	expectedSteps := []normalizedStep{
		{Events: []string{"FunctionCall(name=main)", "EnterScope", "LineChanged(line=2)"}},
		{Events: []string{"DeclareVar(name=sum,global=false)", "LineChanged(line=3)"}},
		{Events: []string{"EnterScope", "LineChanged(line=3)"}},
		{Events: []string{"DeclareVar(name=i,global=false)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"VarChanged(name=sum,value=0)", "ExitScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=1)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"VarChanged(name=sum,value=1)", "ExitScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=2)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"VarChanged(name=sum,value=3)", "ExitScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=3)", "ExitScope", "LineChanged(line=6)"}},
		{Events: []string{"ExitScope", "FunctionReturn(name=main,value=3)", "LineChanged(line=-1)"}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_DoWhileLoopWithCondition(t *testing.T) {
	code := `int main() {
	int i = 0;
	do {
		i = i + 1;
	} while (i < 3);
	return i;
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 3, *result)
	require.Len(t, steps, 9)

	expectedSteps := []normalizedStep{
		{Events: []string{"FunctionCall(name=main)", "EnterScope", "LineChanged(line=2)"}},
		{Events: []string{"DeclareVar(name=i,global=false)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"VarChanged(name=i,value=1)", "ExitScope", "LineChanged(line=5)"}},
		{Events: []string{"EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"VarChanged(name=i,value=2)", "ExitScope", "LineChanged(line=5)"}},
		{Events: []string{"EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"VarChanged(name=i,value=3)", "ExitScope", "LineChanged(line=5)"}},
		{Events: []string{"LineChanged(line=6)"}},
		{Events: []string{"ExitScope", "FunctionReturn(name=main,value=3)", "LineChanged(line=-1)"}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_MultiVarDeclarationSingleLine(t *testing.T) {
	code := `int main() {
	int a = 1, b = 2, c;
	c = a + b;
	return c;
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 3, *result)
	require.Len(t, steps, 6)

	expectedSteps := []normalizedStep{
		{Events: []string{"FunctionCall(name=main)", "EnterScope", "LineChanged(line=2)"}},
		{Events: []string{"DeclareVar(name=a,global=false)", "LineChanged(line=2)"}},
		{Events: []string{"DeclareVar(name=b,global=false)", "LineChanged(line=2)"}},
		{Events: []string{"DeclareVar(name=c,global=false)", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=c,value=3)", "LineChanged(line=4)"}},
		{Events: []string{"ExitScope", "FunctionReturn(name=main,value=3)", "LineChanged(line=-1)"}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_ContinueInWhileLoop(t *testing.T) {
	code := `int main() {
	int i = 0;
	int sum = 0;
	while (i < 5) {
		i = i + 1;
		if (i == 2) {
			continue;
		}
		sum += i;
	}
	return sum;
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 13, *result)
	require.Len(t, steps, 25)

	expectedSteps := []normalizedStep{
		{Events: []string{"FunctionCall(name=main)", "EnterScope", "LineChanged(line=2)"}},
		{Events: []string{"DeclareVar(name=i,global=false)", "LineChanged(line=3)"}},
		{Events: []string{"DeclareVar(name=sum,global=false)", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"VarChanged(name=i,value=1)", "LineChanged(line=6)"}},
		{Events: []string{"LineChanged(line=9)"}},
		{Events: []string{"VarChanged(name=sum,value=1)", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"VarChanged(name=i,value=2)", "LineChanged(line=6)"}},
		{Events: []string{"EnterScope", "LineChanged(line=7)"}},
		{Events: []string{"ExitScope", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"VarChanged(name=i,value=3)", "LineChanged(line=6)"}},
		{Events: []string{"LineChanged(line=9)"}},
		{Events: []string{"VarChanged(name=sum,value=4)", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"VarChanged(name=i,value=4)", "LineChanged(line=6)"}},
		{Events: []string{"LineChanged(line=9)"}},
		{Events: []string{"VarChanged(name=sum,value=8)", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"VarChanged(name=i,value=5)", "LineChanged(line=6)"}},
		{Events: []string{"LineChanged(line=9)"}},
		{Events: []string{"VarChanged(name=sum,value=13)", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"LineChanged(line=11)"}},
		{Events: []string{"ExitScope", "FunctionReturn(name=main,value=13)", "LineChanged(line=-1)"}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_BreakInWhileLoop(t *testing.T) {
	code := `int main() {
	int i = 0;
	int sum = 0;
	while (i < 10) {
		i = i + 1;
		if (i == 5) {
			break;
		}
		sum += i;
	}
	return sum;
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 10, *result)
	require.Len(t, steps, 24)

	expectedSteps := []normalizedStep{
		{Events: []string{"FunctionCall(name=main)", "EnterScope", "LineChanged(line=2)"}},
		{Events: []string{"DeclareVar(name=i,global=false)", "LineChanged(line=3)"}},
		{Events: []string{"DeclareVar(name=sum,global=false)", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"VarChanged(name=i,value=1)", "LineChanged(line=6)"}},
		{Events: []string{"LineChanged(line=9)"}},
		{Events: []string{"VarChanged(name=sum,value=1)", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"VarChanged(name=i,value=2)", "LineChanged(line=6)"}},
		{Events: []string{"LineChanged(line=9)"}},
		{Events: []string{"VarChanged(name=sum,value=3)", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"VarChanged(name=i,value=3)", "LineChanged(line=6)"}},
		{Events: []string{"LineChanged(line=9)"}},
		{Events: []string{"VarChanged(name=sum,value=6)", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"VarChanged(name=i,value=4)", "LineChanged(line=6)"}},
		{Events: []string{"LineChanged(line=9)"}},
		{Events: []string{"VarChanged(name=sum,value=10)", "ExitScope", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"VarChanged(name=i,value=5)", "LineChanged(line=6)"}},
		{Events: []string{"EnterScope", "LineChanged(line=7)"}},
		{Events: []string{"ExitScope", "ExitScope", "LineChanged(line=11)"}},
		{Events: []string{"ExitScope", "FunctionReturn(name=main,value=10)", "LineChanged(line=-1)"}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_ForLoopWithContinueAndBreak(t *testing.T) {
	code := `int main() {
	int sum = 0;
	for (int i = 0; i < 7; i = i + 1) {
		if (i == 2) {
			continue;
		}
		if (i == 5) {
			break;
		}
		sum += i;
	}
	return sum;
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 8, *result)
	require.Len(t, steps, 27)

	expectedSteps := []normalizedStep{
		{Events: []string{"FunctionCall(name=main)", "EnterScope", "LineChanged(line=2)"}},
		{Events: []string{"DeclareVar(name=sum,global=false)", "LineChanged(line=3)"}},
		{Events: []string{"EnterScope", "LineChanged(line=3)"}},
		{Events: []string{"DeclareVar(name=i,global=false)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"LineChanged(line=7)"}},
		{Events: []string{"LineChanged(line=10)"}},
		{Events: []string{"VarChanged(name=sum,value=0)", "ExitScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=1)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"LineChanged(line=7)"}},
		{Events: []string{"LineChanged(line=10)"}},
		{Events: []string{"VarChanged(name=sum,value=1)", "ExitScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=2)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"EnterScope", "LineChanged(line=5)"}},
		{Events: []string{"ExitScope", "ExitScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=3)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"LineChanged(line=7)"}},
		{Events: []string{"LineChanged(line=10)"}},
		{Events: []string{"VarChanged(name=sum,value=4)", "ExitScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=4)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"LineChanged(line=7)"}},
		{Events: []string{"LineChanged(line=10)"}},
		{Events: []string{"VarChanged(name=sum,value=8)", "ExitScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=5)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"LineChanged(line=7)"}},
		{Events: []string{"EnterScope", "LineChanged(line=8)"}},
		{Events: []string{"ExitScope", "ExitScope", "ExitScope", "LineChanged(line=12)"}},
		{Events: []string{"ExitScope", "FunctionReturn(name=main,value=8)", "LineChanged(line=-1)"}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_Array2DDeclarationAndElementChanges(t *testing.T) {
	code := `int main() {
	int m[2][2] = {{1, 2}, {3, 4}};
	m[1][0] = 7;
	m[0][1] = m[1][0] + 1;
	return m[0][1];
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 8, *result)
	require.Len(t, steps, 5)

	expectedSteps := []normalizedStep{
		{Events: []string{"FunctionCall(name=main)", "EnterScope", "LineChanged(line=2)"}},
		{Events: []string{"DeclareArray2D(name=m,size1=2,size2=2,global=false)", "LineChanged(line=3)"}},
		{Events: []string{"Array2DElementChanged(name=m,ind1=1,ind2=0,value=7)", "LineChanged(line=4)"}},
		{Events: []string{"Array2DElementChanged(name=m,ind1=0,ind2=1,value=8)", "LineChanged(line=5)"}},
		{Events: []string{"ExitScope", "FunctionReturn(name=main,value=8)", "LineChanged(line=-1)"}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_UndefinedBehaviorOnUninitializedVariableRead(t *testing.T) {
	code := `int main() {
	int x;
	return x;
}`

	result, steps, stepBegin, err := runCodeWithStepsAllowError(t, code)

	assert.Nil(t, result)
	assert.Nil(t, steps)
	assert.Equal(t, 0, stepBegin)
	require.Error(t, err)
	assert.Equal(t, "undefined behavior: getting an uninitialized variable x", err.Error())
}

func TestInterpreterSteps_GlobalDeclarationAndStepBegin(t *testing.T) {
	code := `int g = 5;
int main() {
	return g;
}`

	result, steps, stepBegin := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 5, *result)
	require.Len(t, steps, 3)
	assert.Equal(t, 1, stepBegin)

	expectedSteps := []normalizedStep{
		{Events: []string{"LineChanged(line=1)"}},
		{Events: []string{"DeclareVar(name=g,global=true)", "FunctionCall(name=main)", "EnterScope", "LineChanged(line=3)"}},
		{Events: []string{"ExitScope", "FunctionReturn(name=main,value=5)", "LineChanged(line=-1)"}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_IfElseFalseBranchEvents(t *testing.T) {
	code := `int main() {
	int x = 0;
	if (0) {
		x = 1;
	} else {
		x = 2;
	}
	return x;
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 2, *result)
	require.Len(t, steps, 5)

	expectedSteps := []normalizedStep{
		{Events: []string{"FunctionCall(name=main)", "EnterScope", "LineChanged(line=2)"}},
		{Events: []string{"DeclareVar(name=x,global=false)", "LineChanged(line=3)"}},
		{Events: []string{"EnterScope", "LineChanged(line=6)"}},
		{Events: []string{"VarChanged(name=x,value=2)", "ExitScope", "LineChanged(line=8)"}},
		{Events: []string{"ExitScope", "FunctionReturn(name=main,value=2)", "LineChanged(line=-1)"}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_ForContinueRunsPostExpression(t *testing.T) {
	code := `int main() {
	int i = 0;
	for (i = 0; i < 3; i = i + 1) {
		continue;
	}
	return i;
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 3, *result)
	require.Len(t, steps, 11)

	expectedSteps := []normalizedStep{
		{Events: []string{"FunctionCall(name=main)", "EnterScope", "LineChanged(line=2)"}},
		{Events: []string{"DeclareVar(name=i,global=false)", "LineChanged(line=3)"}},
		{Events: []string{"EnterScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=0)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"ExitScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=1)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"ExitScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=2)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"ExitScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=3)", "ExitScope", "LineChanged(line=6)"}},
		{Events: []string{"ExitScope", "FunctionReturn(name=main,value=3)", "LineChanged(line=-1)"}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_ForBreakSkipsPostExpression(t *testing.T) {
	code := `int main() {
	int i = 0;
	for (i = 0; i < 3; i = i + 1) {
		break;
	}
	return i;
}`

	result, steps, _ := runCodeWithSteps(t, code)

	require.NotNil(t, result)
	assert.Equal(t, 0, *result)
	require.Len(t, steps, 6)

	expectedSteps := []normalizedStep{
		{Events: []string{"FunctionCall(name=main)", "EnterScope", "LineChanged(line=2)"}},
		{Events: []string{"DeclareVar(name=i,global=false)", "LineChanged(line=3)"}},
		{Events: []string{"EnterScope", "LineChanged(line=3)"}},
		{Events: []string{"VarChanged(name=i,value=0)", "EnterScope", "LineChanged(line=4)"}},
		{Events: []string{"ExitScope", "ExitScope", "LineChanged(line=6)"}},
		{Events: []string{"ExitScope", "FunctionReturn(name=main,value=0)", "LineChanged(line=-1)"}},
	}

	assert.Equal(t, expectedSteps, normalizeSteps(steps))
}

func TestInterpreterSteps_RuntimeErrorDivisionByZero(t *testing.T) {
	code := `int main() {
	int x = 1;
	return x / 0;
}`

	result, steps, stepBegin, err := runCodeWithStepsAllowError(t, code)

	assert.Nil(t, result)
	assert.Nil(t, steps)
	assert.Equal(t, 0, stepBegin)
	require.Error(t, err)
	assert.Equal(t, "runtime error: division by zero", err.Error())
}

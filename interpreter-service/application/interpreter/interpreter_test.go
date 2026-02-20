package interpreter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
)

// Helper function to parse and execute C code
func runCode(t *testing.T, code string) *int {
	conv := converter.New()
	program, convErr := conv.ParseToAST(code)
	if convErr != nil {
		t.Fatalf("parse error: %v", convErr)
	}
	require.NotNil(t, program, "program should not be nil")

	runner := NewInterpreter()
	result, err := runner.ExecuteProgram(program)
	if err != nil {
		t.Fatalf("runtime error: %v", err)
	}

	return result
}

// Helper to get variable value from executed code
func getVariableValue(t *testing.T, code string, varName string) int {
	conv := converter.New()
	program, convErr := conv.ParseToAST(code)
	require.NoError(t, convErr, "parse should succeed")

	runner := NewInterpreter()
	_, err := runner.ExecuteProgram(program)
	require.NoError(t, err, "execution should succeed")

	frame := runner.CallStack.GetCurrentFrame()
	variable, found := frame.GetVariable(varName)
	require.True(t, found, "variable %s should exist", varName)

	value, err := variable.GetValue()
	require.NoError(t, err, "should be able to get variable value")

	return value
}

// ============ Arithmetic Expressions ============

func TestArithmeticAddition(t *testing.T) {
	code := `
int main() {
    return 5 + 3;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 8, *result)
}

func TestArithmeticSubtraction(t *testing.T) {
	code := `
int main() {
    return 10 - 4;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 6, *result)
}

func TestArithmeticMultiplication(t *testing.T) {
	code := `
int main() {
    return 6 * 7;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 42, *result)
}

func TestArithmeticDivision(t *testing.T) {
	code := `
int main() {
    return 20 / 4;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 5, *result)
}

func TestArithmeticModulo(t *testing.T) {
	code := `
int main() {
    return 17 % 5;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 2, *result)
}

func TestArithmeticOrderOfOperations(t *testing.T) {
	code := `
int main() {
    return 2 + 3 * 4;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 14, *result)
}

// ============ Comparison Expressions ============

func TestComparisonEqual(t *testing.T) {
	code := `
int main() {
    return 5 == 5;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 1, *result)
}

func TestComparisonNotEqual(t *testing.T) {
	code := `
int main() {
    return 5 != 3;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 1, *result)
}

func TestComparisonLessThan(t *testing.T) {
	code := `
int main() {
    return 3 < 5;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 1, *result)
}

func TestComparisonLessThanFalse(t *testing.T) {
	code := `
int main() {
    return 5 < 3;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 0, *result)
}

func TestComparisonGreaterThan(t *testing.T) {
	code := `
int main() {
    return 7 > 2;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 1, *result)
}

// ============ Logical Expressions ============

func TestLogicalAnd(t *testing.T) {
	code := `
int main() {
    return 1 && 1;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 1, *result)
}

func TestLogicalAndFalse(t *testing.T) {
	code := `
int main() {
    return 1 && 0;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 0, *result)
}

func TestLogicalOr(t *testing.T) {
	code := `
int main() {
    return 0 || 1;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 1, *result)
}

func TestLogicalOrBothFalse(t *testing.T) {
	code := `
int main() {
    return 0 || 0;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 0, *result)
}

func TestLogicalNot(t *testing.T) {
	code := `
int main() {
    return !0;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 1, *result)
}

func TestLogicalNotTrue(t *testing.T) {
	code := `
int main() {
    return !1;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 0, *result)
}

// ============ Assignment and Variable Declarations ============

func TestVariableDeclaration(t *testing.T) {
	code := `
int main() {
    int x = 42;
    return x;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 42, *result)
}

func TestVariableAssignment(t *testing.T) {
	code := `
int main() {
    int x = 0;
    x = 100;
    return x;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 100, *result)
}

func TestCompoundAssignmentAddition(t *testing.T) {
	code := `
int main() {
    int x = 10;
    x += 5;
    return x;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 15, *result)
}

func TestCompoundAssignmentSubtraction(t *testing.T) {
	code := `
int main() {
    int x = 20;
    x -= 7;
    return x;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 13, *result)
}

func TestIncrementPrefix(t *testing.T) {
	code := `
int main() {
    int x = 5;
    ++x;
    return x;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 6, *result)
}

func TestIncrementPostfix(t *testing.T) {
	code := `
int main() {
    int x = 5;
    int y = x++;
    return y;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 5, *result)
}

func TestDecrementPrefix(t *testing.T) {
	code := `
int main() {
    int x = 10;
    --x;
    return x;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 9, *result)
}

// ============ If Statements ============

func TestIfStatementTrue(t *testing.T) {
	code := `
int main() {
    int x = 5;
    if (x > 0) {
        return 100;
    }
    return 0;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 100, *result)
}

func TestIfStatementFalse(t *testing.T) {
	code := `
int main() {
    int x = -5;
    if (x > 0) {
        return 100;
    }
    return 0;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 0, *result)
}

func TestIfElseStatement(t *testing.T) {
	code := `
int main() {
    int x = -5;
    if (x > 0) {
        return 100;
    } else {
        return 200;
    }
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 200, *result)
}

func TestNestedIfStatements(t *testing.T) {
	code := `
int main() {
    int x = 5;
    int y = 3;
    if (x > 0) {
        if (y > 0) {
            return 42;
        }
    }
    return 0;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 42, *result)
}

// ============ While Loops ============

func TestWhileLoop(t *testing.T) {
	code := `
int main() {
    int sum = 0;
    int i = 1;
    while (i <= 5) {
        sum += i;
        i++;
    }
    return sum;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 15, *result)
}

func TestWhileLoopBreak(t *testing.T) {
	code := `
int main() {
    int i = 0;
    while (i < 10) {
        if (i == 5) {
            break;
        }
        i++;
    }
    return i;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 5, *result)
}

func TestWhileLoopContinue(t *testing.T) {
	code := `
int main() {
    int sum = 0;
    int i = 0;
    while (i < 5) {
        i++;
        if (i == 3) {
            continue;
        }
        sum += i;
    }
    return sum;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 12, *result) // 1 + 2 + 4 + 5
}

// ============ For Loops ============

func TestForLoop(t *testing.T) {
	code := `
int main() {
    int sum = 0;
    for (int i = 1; i <= 5; i++) {
        sum += i;
    }
    return sum;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 15, *result)
}

func TestForLoopBreak(t *testing.T) {
	code := `
int main() {
    for (int i = 0; i < 10; i++) {
        if (i == 3) {
            break;
        }
    }
    return 0;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 0, *result)
}

// ============ Do-While Loops ============

func TestDoWhileLoop(t *testing.T) {
	code := `
int main() {
    int i = 0;
    int count = 0;
    do {
        count++;
        i++;
    } while (i < 3);
    return count;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 3, *result)
}

func TestDoWhileLoopExecutesOnce(t *testing.T) {
	code := `
int main() {
    int count = 0;
    do {
        count++;
    } while (0);
    return count;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 1, *result)
}

// ============ Arrays ============

func TestArrayDeclaration(t *testing.T) {
	code := `
int main() {
    int arr[3] = {10, 20, 30};
    return arr[1];
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 20, *result)
}

func TestArrayAccess(t *testing.T) {
	code := `
int main() {
    int arr[5] = {1, 2, 3, 4, 5};
    return arr[2] + arr[4];
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 8, *result)
}

func TestArrayModification(t *testing.T) {
	code := `
int main() {
    int arr[3] = {10, 20, 30};
    arr[1] = 99;
    return arr[1];
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 99, *result)
}

func TestArrayIteration(t *testing.T) {
	code := `
int main() {
    int arr[4] = {1, 2, 3, 4};
    int sum = 0;
    for (int i = 0; i < 4; i++) {
        sum += arr[i];
    }
    return sum;
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 10, *result)
}

// ============ Functions ============

func TestFunctionCall(t *testing.T) {
	code := `
int add(int a, int b) {
    return a + b;
}

int main() {
    return add(3, 5);
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 8, *result)
}

func TestFunctionWithLocalVariables(t *testing.T) {
	code := `
int multiply(int x, int y) {
    int result = x * y;
    return result;
}

int main() {
    return multiply(6, 7);
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 42, *result)
}

func TestNestedFunctionCalls(t *testing.T) {
	code := `
int add(int a, int b) {
    return a + b;
}

int multiply(int a, int b) {
    return a * b;
}

int main() {
    return add(multiply(2, 3), multiply(4, 5));
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 26, *result) // (2*3) + (4*5) = 6 + 20 = 26
}

func TestRecursiveFunction(t *testing.T) {
	code := `
int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

int main() {
    return factorial(5);
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 120, *result)
}

// ============ Complex Scenarios ============

func TestSumToN(t *testing.T) {
	code := `
int sum_to_n(int n) {
    int sum = 0;
    int i = 1;
    while (i <= n) {
        sum += i;
        i++;
    }
    return sum;
}

int main() {
    return sum_to_n(10);
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 55, *result)
}

func TestBubbleSort(t *testing.T) {
	code := `
int main() {
    int arr[5] = {5, 2, 8, 1, 9};
    
    for (int i = 0; i < 5; i++) {
        for (int j = 0; j < 4; j++) {
            if (arr[j] > arr[j + 1]) {
                int temp = arr[j];
                arr[j] = arr[j + 1];
                arr[j + 1] = temp;
            }
        }
    }
    
    return arr[0];
}
`
	result := runCode(t, code)
	require.NotNil(t, result)
	assert.Equal(t, 1, *result)
}

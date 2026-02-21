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

// Helper function to parse and execute C code, expecting an error
func runCodeExpectError(t *testing.T, code string) error {
	conv := converter.New()
	program, convErr := conv.ParseToAST(code)
	if convErr != nil {
		return convErr
	}

	runner := NewInterpreter()
	_, err := runner.ExecuteProgram(program)
	return err
}

// testCase represents a single test case
type testCase struct {
	name     string
	code     string
	expected int
}

// TestArithmetic tests arithmetic operations
func TestArithmetic(t *testing.T) {
	tests := []testCase{
		{
			name:     "addition",
			code:     `int main() { return 5 + 3; }`,
			expected: 8,
		},
		{
			name:     "subtraction",
			code:     `int main() { return 10 - 4; }`,
			expected: 6,
		},
		{
			name:     "multiplication",
			code:     `int main() { return 6 * 7; }`,
			expected: 42,
		},
		{
			name:     "division",
			code:     `int main() { return 20 / 4; }`,
			expected: 5,
		},
		{
			name:     "modulo",
			code:     `int main() { return 17 % 5; }`,
			expected: 2,
		},
		{
			name:     "order of operations",
			code:     `int main() { return 2 + 3 * 4; }`,
			expected: 14,
		},
		{
			name:     "parentheses override precedence",
			code:     `int main() { return (2 + 3) * 4; }`,
			expected: 20,
		},
		{
			name:     "nested parentheses",
			code:     `int main() { return ((5 + 3) * 2) - 6; }`,
			expected: 10,
		},
		{
			name:     "parentheses with division",
			code:     `int main() { return (20 + 4) / 4; }`,
			expected: 6,
		},
		{
			name:     "complex expression with parentheses",
			code:     `int main() { return (3 + 4) * (5 - 2); }`,
			expected: 21,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestComparison tests comparison operations
func TestComparison(t *testing.T) {
	tests := []testCase{
		{
			name:     "equal",
			code:     `int main() { return 5 == 5; }`,
			expected: 1,
		},
		{
			name:     "not equal",
			code:     `int main() { return 5 != 3; }`,
			expected: 1,
		},
		{
			name:     "less than true",
			code:     `int main() { return 3 < 5; }`,
			expected: 1,
		},
		{
			name:     "less than false",
			code:     `int main() { return 5 < 3; }`,
			expected: 0,
		},
		{
			name:     "greater than",
			code:     `int main() { return 7 > 2; }`,
			expected: 1,
		},
		{
			name:     "compare expressions with parentheses",
			code:     `int main() { return (5 + 3) == (4 + 4); }`,
			expected: 1,
		},
		{
			name:     "compare computed results",
			code:     `int main() { return (10 - 2) > (3 * 2); }`,
			expected: 1,
		},
		{
			name:     "chained with parentheses",
			code:     `int main() { return ((5 + 3) == 8) == (7 > 5); }`,
			expected: 1,
		},
		{
			name:     "complex nested comparison",
			code:     `int main() { return ((2 * 5) < (3 * 4)) && ((10 - 3) == 7); }`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestLogical tests logical operations
func TestLogical(t *testing.T) {
	tests := []testCase{
		{
			name:     "and true",
			code:     `int main() { return 1 && 1; }`,
			expected: 1,
		},
		{
			name:     "and false",
			code:     `int main() { return 1 && 0; }`,
			expected: 0,
		},
		{
			name:     "or true",
			code:     `int main() { return 0 || 1; }`,
			expected: 1,
		},
		{
			name:     "or false",
			code:     `int main() { return 0 || 0; }`,
			expected: 0,
		},
		{
			name:     "not false",
			code:     `int main() { return !0; }`,
			expected: 1,
		},
		{
			name:     "not true",
			code:     `int main() { return !1; }`,
			expected: 0,
		},
		{
			name:     "logical with parentheses",
			code:     `int main() { return (1 && 1) || 0; }`,
			expected: 1,
		},
		{
			name:     "nested logical",
			code:     `int main() { return ((1 && 1) || 0) && (1 || 0); }`,
			expected: 1,
		},
		{
			name:     "short circuit and false left",
			code:     `int main() { int x = 0; if (0 && (x = 5)) { x = 10; } return x; }`,
			expected: 0,
		},
		{
			name:     "short circuit or true left",
			code:     `int main() { int x = 0; if (1 || (x = 5)) { x = 10; } return x; }`,
			expected: 10,
		},
		{
			name:     "short circuit and requires right",
			code:     `int main() { return 1 && 5; }`,
			expected: 1,
		},
		{
			name:     "short circuit or requires right",
			code:     `int main() { return 0 || 5; }`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestVariables tests variable declarations and assignments
func TestVariables(t *testing.T) {
	tests := []testCase{
		{
			name:     "declaration",
			code:     `int main() { int x = 42; return x; }`,
			expected: 42,
		},
		{
			name:     "assignment",
			code:     `int main() { int x = 0; x = 100; return x; }`,
			expected: 100,
		},
		{
			name:     "compound add",
			code:     `int main() { int x = 10; x += 5; return x; }`,
			expected: 15,
		},
		{
			name:     "compound subtract",
			code:     `int main() { int x = 20; x -= 7; return x; }`,
			expected: 13,
		},
		{
			name:     "compound multiply",
			code:     `int main() { int x = 6; x *= 7; return x; }`,
			expected: 42,
		},
		{
			name:     "compound divide",
			code:     `int main() { int x = 30; x /= 5; return x; }`,
			expected: 6,
		},
		{
			name:     "compound modulo",
			code:     `int main() { int x = 17; x %= 5; return x; }`,
			expected: 2,
		},
		{
			name:     "pre increment",
			code:     `int main() { int x = 5; ++x; return x; }`,
			expected: 6,
		},
		{
			name:     "post increment",
			code:     `int main() { int x = 5; int y = x++; return y; }`,
			expected: 5,
		},
		{
			name:     "pre decrement",
			code:     `int main() { int x = 10; --x; return x; }`,
			expected: 9,
		},
		{
			name:     "post decrement",
			code:     `int main() { int x = 10; int y = x--; return y; }`,
			expected: 10,
		},
		{
			name:     "pre increment in expression",
			code:     `int main() { int x = 5; return ++x + 2; }`,
			expected: 8,
		},
		{
			name:     "post increment in expression",
			code:     `int main() { int x = 5; return x++ + 2; }`,
			expected: 7,
		},
		{
			name:     "multiple assignments",
			code:     `int main() { int x = 5; int y = x; int z = y; return z; }`,
			expected: 5,
		},
		{
			name:     "chained assignments",
			code:     `int main() { int x = 10; x += 5; x *= 2; return x; }`,
			expected: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestIfElse tests if/else statements
func TestIfElse(t *testing.T) {
	tests := []testCase{
		{
			name:     "if true",
			code:     `int main() { int x = 5; if (x > 0) { return 100; } return 0; }`,
			expected: 100,
		},
		{
			name:     "if false",
			code:     `int main() { int x = -5; if (x > 0) { return 100; } return 0; }`,
			expected: 0,
		},
		{
			name:     "if else",
			code:     `int main() { int x = -5; if (x > 0) { return 100; } else { return 200; } }`,
			expected: 200,
		},
		{
			name:     "nested if",
			code:     `int main() { int x = 5; int y = 3; if (x > 0) { if (y > 0) { return 42; } } return 0; }`,
			expected: 42,
		},
		{
			name:     "else if single",
			code:     `int main() { int x = 15; if (x < 10) { return 1; } else if (x < 20) { return 2; } else { return 3; } }`,
			expected: 2,
		},
		{
			name:     "else if multiple first",
			code:     `int main() { int x = 5; if (x < 10) { return 1; } else if (x < 20) { return 2; } else if (x < 30) { return 3; } else { return 4; } }`,
			expected: 1,
		},
		{
			name:     "else if multiple second",
			code:     `int main() { int x = 15; if (x < 10) { return 1; } else if (x < 20) { return 2; } else if (x < 30) { return 3; } else { return 4; } }`,
			expected: 2,
		},
		{
			name:     "else if multiple third",
			code:     `int main() { int x = 25; if (x < 10) { return 1; } else if (x < 20) { return 2; } else if (x < 30) { return 3; } else { return 4; } }`,
			expected: 3,
		},
		{
			name:     "else if multiple final else",
			code:     `int main() { int x = 35; if (x < 10) { return 1; } else if (x < 20) { return 2; } else if (x < 30) { return 3; } else { return 4; } }`,
			expected: 4,
		},
		{
			name:     "else if with complex conditions",
			code:     `int main() { int x = 15; if ((x > 0) && (x < 10)) { return 1; } else if ((x >= 10) && (x < 20)) { return 2; } else { return 3; } }`,
			expected: 2,
		},
		{
			name:     "else if many branches",
			code:     `int main() { int x = 50; if (x < 10) { return 1; } else if (x < 20) { return 2; } else if (x < 30) { return 3; } else if (x < 40) { return 4; } else if (x < 60) { return 5; } else { return 6; } }`,
			expected: 5,
		},
		{
			name:     "else if no match final",
			code:     `int main() { int x = 100; if (x == 10) { return 1; } else if (x == 20) { return 2; } else if (x == 30) { return 3; } else { return 4; } }`,
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestWhileLoop tests while loops
func TestWhileLoop(t *testing.T) {
	tests := []testCase{
		{
			name:     "simple",
			code:     `int main() { int sum = 0; int i = 1; while (i <= 5) { sum += i; i++; } return sum; }`,
			expected: 15,
		},
		{
			name:     "break",
			code:     `int main() { int i = 0; while (i < 10) { if (i == 5) { break; } i++; } return i; }`,
			expected: 5,
		},
		{
			name:     "continue",
			code:     `int main() { int sum = 0; int i = 0; while (i < 5) { i++; if (i == 3) { continue; } sum += i; } return sum; }`,
			expected: 12,
		},
		{
			name:     "nested while simple",
			code:     `int main() { int count = 0; int i = 0; while (i < 3) { int j = 0; while (j < 2) { count++; j++; } i++; } return count; }`,
			expected: 6,
		},
		{
			name:     "nested while accumulation",
			code:     `int main() { int sum = 0; int i = 1; while (i <= 3) { int j = 1; while (j <= 2) { sum += i * j; j++; } i++; } return sum; }`,
			expected: 18,
		},
		{
			name:     "nested while with break inner",
			code:     `int main() { int count = 0; int i = 0; while (i < 3) { int j = 0; while (j < 5) { if (j == 2) { break; } count++; j++; } i++; } return count; }`,
			expected: 6,
		},
		{
			name:     "nested while with break outer",
			code:     `int main() { int count = 0; int i = 0; while (i < 5) { int j = 0; while (j < 3) { count++; j++; } if (i == 1) { break; } i++; } return count; }`,
			expected: 6,
		},
		{
			name:     "triple nested while",
			code:     `int main() { int count = 0; int i = 0; while (i < 2) { int j = 0; while (j < 2) { int k = 0; while (k < 2) { count++; k++; } j++; } i++; } return count; }`,
			expected: 8,
		},
		{
			name:     "nested while different conditions",
			code:     `int main() { int sum = 0; int i = 0; while (i < 3) { int j = 10; while (j > 8) { sum += j; j--; } i++; } return sum; }`,
			expected: 57,
		},
		{
			name:     "nested while matrix iteration",
			code:     `int main() { int result = 0; int i = 1; while (i <= 2) { int j = 1; while (j <= 3) { result += (i * 10) + j; j++; } i++; } return result; }`,
			expected: 102,
		},
		{
			name:     "nested while continue in inner",
			code:     `int main() { int sum = 0; int i = 0; while (i < 3) { int j = 0; while (j < 4) { if (j == 2) { j++; continue; } sum += j; j++; } i++; } return sum; }`,
			expected: 12,
		},
		{
			name:     "nested while continue in outer",
			code:     `int main() { int sum = 0; int i = 0; while (i < 4) { if (i == 2) { i++; continue; } int j = 0; while (j < 2) { sum += 1; j++; } i++; } return sum; }`,
			expected: 6,
		},
		{
			name:     "nested while continue both loops",
			code:     `int main() { int count = 0; int i = 0; while (i < 3) { if (i == 1) { i++; continue; } int j = 0; while (j < 3) { if (j == 1) { j++; continue; } count++; j++; } i++; } return count; }`,
			expected: 4,
		},
		{
			name:     "nested while continue skip accumulation",
			code:     `int main() { int sum = 0; int i = 1; while (i <= 3) { int j = 1; while (j <= 3) { if ((i == 2) && (j == 2)) { j++; continue; } sum += i; j++; } i++; } return sum; }`,
			expected: 16,
		},
		{
			name:     "nested while continue all inner iterations",
			code:     `int main() { int count = 0; int i = 0; while (i < 2) { int j = 0; while (j < 3) { count++; j++; if (j < 3) { continue; } } i++; } return count; }`,
			expected: 6,
		},
		{
			name:     "return from while loop early",
			code:     `int main() { int i = 0; while (i < 10) { if (i == 5) { return 42; } i++; } return 0; }`,
			expected: 42,
		},
		{
			name:     "return from nested while loop",
			code:     `int main() { int i = 0; while (i < 3) { int j = 0; while (j < 3) { if (j == 1) { return 100; } j++; } i++; } return 0; }`,
			expected: 100,
		},
		{
			name:     "return from while with accumulation",
			code:     `int main() { int sum = 0; int i = 0; while (i < 5) { sum += i; if (sum > 5) { return sum; } i++; } return 0; }`,
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestForLoop tests for loops
func TestForLoop(t *testing.T) {
	tests := []testCase{
		{
			name:     "simple",
			code:     `int main() { int sum = 0; for (int i = 1; i <= 5; i++) { sum += i; } return sum; }`,
			expected: 15,
		},
		{
			name:     "break",
			code:     `int main() { for (int i = 0; i < 10; i++) { if (i == 3) { break; } } return 0; }`,
			expected: 0,
		},
		{
			name:     "continue",
			code:     `int main() { int sum = 0; for (int i = 0; i < 5; i++) { if (i == 2) { continue; } sum += i; } return sum; }`,
			expected: 8,
		},
		{
			name:     "nested for simple",
			code:     `int main() { int count = 0; for (int i = 0; i < 3; i++) { for (int j = 0; j < 2; j++) { count++; } } return count; }`,
			expected: 6,
		},
		{
			name:     "nested for accumulation",
			code:     `int main() { int sum = 0; for (int i = 1; i <= 3; i++) { for (int j = 1; j <= 2; j++) { sum += i * j; } } return sum; }`,
			expected: 18,
		},
		{
			name:     "nested for with break inner",
			code:     `int main() { int count = 0; for (int i = 0; i < 3; i++) { for (int j = 0; j < 5; j++) { if (j == 2) { break; } count++; } } return count; }`,
			expected: 6,
		},
		{
			name:     "nested for continue in inner",
			code:     `int main() { int sum = 0; for (int i = 0; i < 3; i++) { for (int j = 0; j < 4; j++) { if (j == 2) { continue; } sum += j; } } return sum; }`,
			expected: 12,
		},
		{
			name:     "nested for continue in outer",
			code:     `int main() { int sum = 0; for (int i = 0; i < 4; i++) { if (i == 2) { continue; } for (int j = 0; j < 2; j++) { sum++; } } return sum; }`,
			expected: 6,
		},
		{
			name:     "nested for continue both",
			code:     `int main() { int count = 0; for (int i = 0; i < 3; i++) { if (i == 1) { continue; } for (int j = 0; j < 3; j++) { if (j == 1) { continue; } count++; } } return count; }`,
			expected: 4,
		},
		{
			name:     "nested for continue skip accumulation",
			code:     `int main() { int sum = 0; for (int i = 1; i <= 3; i++) { for (int j = 1; j <= 3; j++) { if ((i == 2) && (j == 2)) { continue; } sum += i; } } return sum; }`,
			expected: 16,
		},
		{
			name:     "nested for triple nesting",
			code:     `int main() { int count = 0; for (int i = 0; i < 2; i++) { for (int j = 0; j < 2; j++) { for (int k = 0; k < 2; k++) { count++; } } } return count; }`,
			expected: 8,
		},
		{
			name:     "for init is assignment not declaration",
			code:     `int main() { int i; for (i = 0; i < 5; i++) { } return i; }`,
			expected: 5,
		},
		{
			name:     "for init with pre-declared variable",
			code:     `int main() { int sum = 0; int i; for (i = 1; i <= 5; i++) { sum += i; } return sum; }`,
			expected: 15,
		},
		{
			name:     "for init with existing variable assignment",
			code:     `int main() { int x = 10; for (x = 0; x < 3; x++) { } return x; }`,
			expected: 3,
		},
		{
			name:     "for init with compound assignment",
			code:     `int main() { int sum = 100; for (sum = 0; sum < 5; sum++) { } return sum; }`,
			expected: 5,
		},
		{
			name:     "for empty init",
			code:     `int main() { int i = 0; for (; i < 5; i++) { } return i; }`,
			expected: 5,
		},
		{
			name:     "for empty init with accumulation",
			code:     `int main() { int i = 0; int sum = 0; for (; i < 5; i++) { sum += i; } return sum; }`,
			expected: 10,
		},
		{
			name:     "for init expression with addition",
			code:     `int main() { int x = 5; for (x = x + 1; x < 10; x++) { } return x; }`,
			expected: 10,
		},
		{
			name:     "for init with function-like expression",
			code:     `int main() { int count = 0; int i = 0; for (count = 2; count < 4; count++) { i += count; } return i; }`,
			expected: 5,
		},
		{
			name:     "return from for loop early",
			code:     `int main() { for (int i = 0; i < 10; i++) { if (i == 5) { return 42; } } return 0; }`,
			expected: 42,
		},
		{
			name:     "return from nested for loop",
			code:     `int main() { for (int i = 0; i < 3; i++) { for (int j = 0; j < 3; j++) { if (j == 1) { return 100; } } } return 0; }`,
			expected: 100,
		},
		{
			name:     "return from for with accumulation",
			code:     `int main() { int sum = 0; for (int i = 0; i < 5; i++) { sum += i; if (sum > 5) { return sum; } } return 0; }`,
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestDoWhileLoop tests do-while loops
func TestDoWhileLoop(t *testing.T) {
	tests := []testCase{
		{
			name:     "simple",
			code:     `int main() { int i = 0; int count = 0; do { count++; i++; } while (i < 3); return count; }`,
			expected: 3,
		},
		{
			name:     "executes once",
			code:     `int main() { int count = 0; do { count++; } while (0); return count; }`,
			expected: 1,
		},
		{
			name:     "continue",
			code:     `int main() { int sum = 0; int i = 0; do { if (i == 2) { i++; continue; } sum += i; i++; } while (i < 5); return sum; }`,
			expected: 8,
		},
		{
			name:     "nested do-while simple",
			code:     `int main() { int count = 0; int i = 0; do { int j = 0; do { count++; j++; } while (j < 2); i++; } while (i < 3); return count; }`,
			expected: 6,
		},
		{
			name:     "nested do-while accumulation",
			code:     `int main() { int sum = 0; int i = 1; do { int j = 1; do { sum += i * j; j++; } while (j <= 2); i++; } while (i <= 3); return sum; }`,
			expected: 18,
		},
		{
			name:     "nested do-while with break inner",
			code:     `int main() { int count = 0; int i = 0; do { int j = 0; do { if (j == 2) { break; } count++; j++; } while (j < 5); i++; } while (i < 3); return count; }`,
			expected: 6,
		},
		{
			name:     "nested do-while continue in inner",
			code:     `int main() { int sum = 0; int i = 0; do { int j = 0; do { if (j == 2) { j++; continue; } sum += j; j++; } while (j < 4); i++; } while (i < 3); return sum; }`,
			expected: 12,
		},
		{
			name:     "nested do-while continue in outer",
			code:     `int main() { int sum = 0; int i = 0; do { if (i == 2) { i++; continue; } int j = 0; do { sum++; j++; } while (j < 2); i++; } while (i < 4); return sum; }`,
			expected: 6,
		},
		{
			name:     "nested do-while continue both",
			code:     `int main() { int count = 0; int i = 0; do { if (i == 1) { i++; continue; } int j = 0; do { if (j == 1) { j++; continue; } count++; j++; } while (j < 3); i++; } while (i < 3); return count; }`,
			expected: 4,
		},
		{
			name:     "nested do-while continue skip accumulation",
			code:     `int main() { int sum = 0; int i = 1; do { int j = 1; do { if ((i == 2) && (j == 2)) { j++; continue; } sum += i; j++; } while (j <= 3); i++; } while (i <= 3); return sum; }`,
			expected: 16,
		},
		{
			name:     "nested do-while triple nesting",
			code:     `int main() { int count = 0; int i = 0; do { int j = 0; do { int k = 0; do { count++; k++; } while (k < 2); j++; } while (j < 2); i++; } while (i < 2); return count; }`,
			expected: 8,
		},
		{
			name:     "return from do-while loop early",
			code:     `int main() { int i = 0; do { if (i == 5) { return 42; } i++; } while (i < 10); return 0; }`,
			expected: 42,
		},
		{
			name:     "return from nested do-while loop",
			code:     `int main() { int i = 0; do { int j = 0; do { if (j == 1) { return 100; } j++; } while (j < 3); i++; } while (i < 3); return 0; }`,
			expected: 100,
		},
		{
			name:     "return from do-while with accumulation",
			code:     `int main() { int sum = 0; int i = 0; do { sum += i; if (sum > 5) { return sum; } i++; } while (i < 5); return 0; }`,
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestArray tests array operations
func TestArray(t *testing.T) {
	tests := []testCase{
		{
			name:     "declaration",
			code:     `int main() { int arr[3] = {10, 20, 30}; return arr[1]; }`,
			expected: 20,
		},
		{
			name:     "access",
			code:     `int main() { int arr[5] = {1, 2, 3, 4, 5}; return arr[2] + arr[4]; }`,
			expected: 8,
		},
		{
			name:     "modification",
			code:     `int main() { int arr[3] = {10, 20, 30}; arr[1] = 99; return arr[1]; }`,
			expected: 99,
		},
		{
			name:     "iteration",
			code:     `int main() { int arr[4] = {1, 2, 3, 4}; int sum = 0; for (int i = 0; i < 4; i++) { sum += arr[i]; } return sum; }`,
			expected: 10,
		},
		{
			name:     "multiple modifications",
			code:     `int main() { int arr[5] = {1, 2, 3, 4, 5}; arr[0] = 10; arr[2] = 30; arr[4] = 50; return arr[0] + arr[2] + arr[4]; }`,
			expected: 90,
		},
		{
			name:     "modification affects subsequent access",
			code:     `int main() { int arr[3] = {5, 10, 15}; arr[1] = 20; int x = arr[1]; arr[1] = 25; return x + arr[1]; }`,
			expected: 45,
		},
		{
			name:     "first and last elements",
			code:     `int main() { int arr[5] = {100, 2, 3, 4, 200}; return arr[0] + arr[4]; }`,
			expected: 300,
		},
		{
			name:     "modification in loop",
			code:     `int main() { int arr[5] = {1, 2, 3, 4, 5}; for (int i = 0; i < 5; i++) { arr[i] = arr[i] * 2; } int sum = 0; for (int i = 0; i < 5; i++) { sum += arr[i]; } return sum; }`,
			expected: 30,
		},
		{
			name:     "array element arithmetic",
			code:     `int main() { int arr[4] = {10, 20, 30, 40}; int result = arr[0] + arr[1] * arr[2] - arr[3]; return result; }`,
			expected: 570,
		},
		{
			name:     "conditional modification",
			code:     `int main() { int arr[3] = {5, 10, 15}; for (int i = 0; i < 3; i++) { if (arr[i] > 7) { arr[i] = 100; } } return arr[0] + arr[1] + arr[2]; }`,
			expected: 205,
		},
		{
			name:     "read modify write sequence",
			code:     `int main() { int arr[3] = {10, 20, 30}; arr[0] = arr[1] + arr[2]; return arr[0]; }`,
			expected: 50,
		},
		{
			name:     "accumulation from array",
			code:     `int main() { int arr[6] = {10, 20, 30, 40, 50, 60}; int sum = 0; for (int i = 0; i < 6; i++) { sum += arr[i]; } return sum; }`,
			expected: 210,
		},
		{
			name:     "zero initialization",
			code:     `int main() { int arr[5] = {0, 0, 0, 0, 0}; arr[2] = 42; return arr[2]; }`,
			expected: 42,
		},
		{
			name:     "nested array access modifications",
			code:     `int main() { int arr[4] = {1, 2, 3, 4}; int idx = 1; arr[idx] = arr[idx] + 10; idx = 3; arr[idx] = arr[idx] * 5; return arr[1] + arr[3]; }`,
			expected: 32,
		},
		{
			name:     "array element assignment chain",
			code:     `int main() { int arr[5] = {5, 5, 5, 5, 5}; arr[0] = 10; arr[1] = arr[0] + 5; arr[2] = arr[1] + 5; return arr[2]; }`,
			expected: 20,
		},
		{
			name:     "large array access",
			code:     `int main() { int arr[10] = {1, 2, 3, 4, 5, 6, 7, 8, 9, 10}; return arr[0] + arr[9]; }`,
			expected: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestArray2D tests 2D array operations
func TestArray2D(t *testing.T) {
	tests := []testCase{
		{
			name:     "declaration and access",
			code:     `int main() { int matrix[2][3] = {{1, 2, 3}, {4, 5, 6}}; return matrix[1][2]; }`,
			expected: 6,
		},
		{
			name:     "sum of two elements",
			code:     `int main() { int matrix[2][2] = {{10, 20}, {30, 40}}; return matrix[0][1] + matrix[1][0]; }`,
			expected: 50,
		},
		{
			name:     "modification",
			code:     `int main() { int matrix[2][2] = {{1, 2}, {3, 4}}; matrix[1][0] = 99; return matrix[1][0]; }`,
			expected: 99,
		},
		{
			name:     "multiple modifications",
			code:     `int main() { int matrix[2][3] = {{1, 2, 3}, {4, 5, 6}}; matrix[0][2] = 30; matrix[1][1] = 50; return matrix[0][2] + matrix[1][1]; }`,
			expected: 80,
		},
		{
			name:     "iteration with nested loops",
			code:     `int main() { int matrix[2][3] = {{1, 2, 3}, {4, 5, 6}}; int sum = 0; for (int i = 0; i < 2; i++) { for (int j = 0; j < 3; j++) { sum += matrix[i][j]; } } return sum; }`,
			expected: 21,
		},
		{
			name:     "write in nested loops",
			code:     `int main() { int matrix[2][2] = {{0, 0}, {0, 0}}; for (int i = 0; i < 2; i++) { for (int j = 0; j < 2; j++) { matrix[i][j] = i + j; } } return matrix[0][0] + matrix[0][1] + matrix[1][0] + matrix[1][1]; }`,
			expected: 4,
		},
		{
			name:     "index expression access",
			code:     `int main() { int matrix[2][3] = {{10, 20, 30}, {40, 50, 60}}; int i = 1; int j = 2; return matrix[i - 1][j - 1] + matrix[i][j - 2]; }`,
			expected: 60,
		},
		{
			name:     "read modify write",
			code:     `int main() { int matrix[2][2] = {{5, 6}, {7, 8}}; matrix[0][1] = matrix[1][0] + matrix[1][1]; return matrix[0][1]; }`,
			expected: 15,
		},
		{
			name:     "global 2d array access",
			code:     `int matrix[2][2] = {{1, 2}, {3, 4}}; int main() { return matrix[0][0] + matrix[1][1]; }`,
			expected: 5,
		},
		{
			name:     "global 2d array modification",
			code:     `int matrix[2][2] = {{1, 2}, {3, 4}}; int main() { matrix[0][1] = 20; return matrix[0][1] + matrix[1][0]; }`,
			expected: 23,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestFunction tests function calls
func TestFunction(t *testing.T) {
	tests := []testCase{
		{
			name:     "simple",
			code:     `int add(int a, int b) { return a + b; } int main() { return add(3, 5); }`,
			expected: 8,
		},
		{
			name:     "local variables",
			code:     `int multiply(int x, int y) { int result = x * y; return result; } int main() { return multiply(6, 7); }`,
			expected: 42,
		},
		{
			name:     "nested calls",
			code:     `int add(int a, int b) { return a + b; } int multiply(int a, int b) { return a * b; } int main() { return add(multiply(2, 3), multiply(4, 5)); }`,
			expected: 26,
		},
		{
			name:     "parameter shadowing basic",
			code:     `int test(int x) { return x * 2; } int main() { int x = 10; int result = test(5); return x; }`,
			expected: 10,
		},
		{
			name:     "parameter shadows main variable",
			code:     `int modify(int x) { x = 100; return x; } int main() { int x = 5; int result = modify(99); return x; }`,
			expected: 5,
		},
		{
			name:     "local variable shadows parameter",
			code:     `int test(int a) { int a = 20; return a; } int main() { return test(5); }`,
			expected: 20,
		},
		{
			name:     "multiple parameters shadowing",
			code:     `int add(int x, int y) { return x + y; } int main() { int x = 100; int y = 200; return add(3, 4); }`,
			expected: 7,
		},
		{
			name:     "shadowing preserves caller scope",
			code:     `int change(int a, int b) { a = 50; b = 60; return a + b; } int main() { int a = 1; int b = 2; int result = change(a, b); return a + b; }`,
			expected: 3,
		},
		{
			name:     "local var shadows param shadows outer var",
			code:     `int test(int x) { int x = 50; return x; } int main() { int x = 10; return test(20) + x; }`,
			expected: 60,
		},
		{
			name:     "multiple functions with same variable names",
			code:     `int func1(int val) { return val + 10; } int func2(int val) { return val * 2; } int main() { int val = 5; return func1(val) + func2(val); }`,
			expected: 25,
		},
		{
			name:     "function return preserves outer variable",
			code:     `int increment(int n) { int n = n + 1; return n; } int main() { int n = 5; int newN = increment(n); return n + newN; }`,
			expected: 11,
		},
		{
			name:     "shadowing in conditional",
			code:     `int test(int x) { if (x > 10) { int x = 100; return x; } return x; } int main() { return test(5) + test(15); }`,
			expected: 105,
		},
		{
			name:     "void function modifies global",
			code:     `int g = 0; void set_value(int val) { g = val; } int main() { set_value(42); return g; }`,
			expected: 42,
		},
		{
			name:     "void function no params",
			code:     `int counter = 0; void increment() { counter = counter + 1; } int main() { increment(); increment(); increment(); return counter; }`,
			expected: 3,
		},
		{
			name:     "void function with multiple params",
			code:     `int a = 0, b = 0; void set_both(int x, int y) { a = x; b = y; } int main() { set_both(10, 20); return a + b; }`,
			expected: 30,
		},
		{
			name:     "void function calls int function",
			code:     `int g = 0; int get_value() { return 100; } void update() { g = get_value(); } int main() { update(); return g; }`,
			expected: 100,
		},
		{
			name:     "void function with conditional",
			code:     `int result = 0; void check(int x) { if (x > 5) { result = 10; } else { result = 20; } } int main() { check(3); int r1 = result; check(8); return r1 + result; }`,
			expected: 30,
		},
		{
			name:     "void function with loop",
			code:     `int sum = 0; void add_range(int n) { int i; for (i = 0; i < n; i++) { sum = sum + i; } } int main() { add_range(5); return sum; }`,
			expected: 10,
		},
		{
			name:     "void function with early return",
			code:     `int val = 0; void set_if_positive(int x) { if (x < 0) { return; } val = x; } int main() { set_if_positive(-5); int r1 = val; set_if_positive(25); return r1 + val; }`,
			expected: 25,
		},
		{
			name:     "void function called multiple times",
			code:     `int acc = 0; void add(int x) { acc = acc + x; } int main() { add(5); add(10); add(15); return acc; }`,
			expected: 30,
		},
		{
			name:     "void function modifies multiple globals",
			code:     `int x = 0, y = 0, z = 0; void init() { x = 1; y = 2; z = 3; } int main() { init(); return x + y + z; }`,
			expected: 6,
		},
		{
			name:     "void function with nested calls",
			code:     `int value = 0; void set_ten() { value = 10; } void double_it() { set_ten(); value = value * 2; } int main() { double_it(); return value; }`,
			expected: 20,
		},
		{
			name:     "void function in conditional",
			code:     `int flag = 0; void set_flag() { flag = 1; } int main() { if (5 > 3) { set_flag(); } return flag; }`,
			expected: 1,
		},
		{
			name:     "void function in loop",
			code:     `int total = 0; void add_five() { total = total + 5; } int main() { int i; for (i = 0; i < 3; i++) { add_five(); } return total; }`,
			expected: 15,
		},
		{
			name:     "void function with while loop inside",
			code:     `int count = 0; void count_to(int n) { int i = 0; while (i < n) { count = count + 1; i++; } } int main() { count_to(7); return count; }`,
			expected: 7,
		},
		{
			name:     "void function return in do-while",
			code:     `int result = 0; void process(int n) { int i = 0; do { result = result + i; i++; if (result > 10) { return; } } while (i < n); } int main() { process(20); return result; }`,
			expected: 15,
		},
		{
			name:     "void function sequential calls",
			code:     `int n = 0; void add_one() { n = n + 1; } void add_two() { n = n + 2; } void add_three() { n = n + 3; } int main() { add_one(); add_two(); add_three(); return n; }`,
			expected: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestRecursion tests recursive functions
func TestRecursion(t *testing.T) {
	tests := []testCase{
		{
			name:     "simple parameter",
			code:     `int add(int x, int y) { return x + y; } int main() { return add(3, 4); }`,
			expected: 7,
		},
		{
			name:     "factorial",
			code:     `int factorial(int n) { if (n <= 1) { return 1; } return n * factorial(n - 1); } int main() { return factorial(5); }`,
			expected: 120,
		},
		{
			name:     "recursive shadowing preserves parameter",
			code:     `int countdown(int n) { if (n <= 0) { return n; } return countdown(n - 1); } int main() { return countdown(3); }`,
			expected: 0,
		},
		{
			name:     "fibonacci with parameter shadowing",
			code:     `int fib(int n) { if (n <= 1) { return n; } return fib(n - 1) + fib(n - 2); } int main() { return fib(6); }`,
			expected: 8,
		},
		{
			name:     "recursive sum with shadowing",
			code:     `int sum(int n) { if (n <= 0) { return 0; } return n + sum(n - 1); } int main() { return sum(5); }`,
			expected: 15,
		},
		{
			name:     "recursive accumulation shadows variable",
			code:     `int power(int base, int exp) { if (exp == 0) { return 1; } return base * power(base, exp - 1); } int main() { return power(2, 4); }`,
			expected: 16,
		},
		{
			name:     "local var shadows param in recursion",
			code:     `int test(int n) { int n = 100; if (n <= 100) { return n; } return test(n - 1); } int main() { return test(5); }`,
			expected: 100,
		},
		{
			name:     "multiple recursive parameters",
			code:     `int gcd(int a, int b) { if (b == 0) { return a; } return gcd(b, a % b); } int main() { return gcd(12, 8); }`,
			expected: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestComplex tests complex scenarios
func TestComplex(t *testing.T) {
	tests := []testCase{
		{
			name:     "sum to n",
			code:     `int sum_to_n(int n) { int sum = 0; int i = 1; while (i <= n) { sum += i; i++; } return sum; } int main() { return sum_to_n(10); }`,
			expected: 55,
		},
		{
			name:     "bubble sort",
			code:     `int main() { int arr[5] = {5, 2, 8, 1, 9}; for (int i = 0; i < 5; i++) { for (int j = 0; j < 4; j++) { if (arr[j] > arr[j + 1]) { int temp = arr[j]; arr[j] = arr[j + 1]; arr[j + 1] = temp; } } } return arr[0]; }`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestScopes tests variable scoping and shadowing
func TestScopes(t *testing.T) {
	tests := []testCase{
		{
			name:     "block scope hides outer variable",
			code:     `int main() { int x = 10; { int x = 20; } return x; }`,
			expected: 10,
		},
		{
			name:     "block scope access variable before shadowing",
			code:     `int main() { int x = 10; int y = x; { int x = 20; } return y; }`,
			expected: 10,
		},
		{
			name:     "nested block scopes",
			code:     `int main() { int x = 1; { int x = 2; { int x = 3; } } return x; }`,
			expected: 1,
		},
		{
			name:     "block scope with arithmetic",
			code:     `int main() { int x = 5; { int x = 10; x = x + 5; } return x; }`,
			expected: 5,
		},
		{
			name:     "for loop variable shadows outer",
			code:     `int main() { int i = 100; for (int i = 0; i < 1; i++) { } return i; }`,
			expected: 100,
		},
		{
			name:     "for loop variable scope ends after loop",
			code:     `int main() { int sum = 0; for (int i = 0; i < 5; i++) { sum += i; } return sum; }`,
			expected: 10,
		},
		{
			name:     "while loop variable in block",
			code:     `int main() { int x = 10; while (0) { int x = 20; } return x; }`,
			expected: 10,
		},
		{
			name:     "if block shadows variable",
			code:     `int main() { int x = 5; if (1) { int x = 50; } return x; }`,
			expected: 5,
		},
		{
			name:     "else block shadows variable",
			code:     `int main() { int x = 5; if (0) { int x = 10; } else { int x = 20; } return x; }`,
			expected: 5,
		},
		{
			name:     "multiple block scopes independent",
			code:     `int main() { int x = 1; { int x = 2; } { int x = 3; } return x; }`,
			expected: 1,
		},
		{
			name:     "function parameter shadows global in block",
			code:     `int test(int x) { int y = x; { int x = 100; } return y; } int main() { return test(5); }`,
			expected: 5,
		},
		{
			name:     "shadowing in nested if",
			code:     `int main() { int x = 1; if (1) { int x = 2; if (1) { int x = 3; } int y = x; return y; } }`,
			expected: 2,
		},
		{
			name:     "block scope modification doesn't affect outer",
			code:     `int main() { int x = 10; { int x = 20; x = 30; } return x; }`,
			expected: 10,
		},
		{
			name:     "variable shadowing in for loop body",
			code:     `int main() { int x = 50; int sum = 0; for (int i = 0; i < 3; i++) { int x = 10; sum += x; } return sum + x; }`,
			expected: 80,
		},
		{
			name:     "nested for with variable shadowing",
			code:     `int main() { int i = 100; int count = 0; for (int i = 0; i < 2; i++) { for (int j = 0; j < 2; j++) { count++; } } return i + count; }`,
			expected: 104,
		},
		{
			name:     "while loop variable scope",
			code:     `int main() { int count = 0; int i = 0; while (i < 3) { int temp = i; count += temp; i++; } int x = count; return x; }`,
			expected: 3,
		},
		{
			name:     "block scope in else if chain",
			code:     `int main() { int result = 0; int x = 5; if (x < 0) { int x = 1; result = x; } else if (x > 0) { int x = 2; result = x; } else { int x = 3; result = x; } return result; }`,
			expected: 2,
		},
		{
			name:     "shadowing with different types in same scope",
			code:     `int main() { int x = 10; { int x = 20; int y = x; } return x; }`,
			expected: 10,
		},
		{
			name:     "multiple variables multiple scopes",
			code:     `int main() { int a = 1; int b = 2; { int a = 10; int c = 3; } return a + b; }`,
			expected: 3,
		},
		{
			name:     "deeply nested block scopes",
			code:     `int main() { int x = 1; { { { { int x = 100; } } } } return x; }`,
			expected: 1,
		},
		{
			name:     "scope with shadowing and accumulation",
			code:     `int main() { int sum = 0; for (int i = 0; i < 3; i++) { int x = i; sum += x; } return sum; }`,
			expected: 3,
		},
		{
			name:     "block variable doesn't leak to outer scope",
			code:     `int main() { { int x = 50; } int y = 10; return y; }`,
			expected: 10,
		},
		{
			name:     "shadowing in conditional with operators",
			code:     `int main() { int x = 5; if (x > 0) { int x = 10; int y = x * 2; return y; } }`,
			expected: 20,
		},
		{
			name:     "nested block with parameter",
			code:     `int test(int x) { { int x = 50; int y = x + 10; return y; } } int main() { return test(5); }`,
			expected: 60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestErrors tests error cases for runtime errors
func TestErrors(t *testing.T) {
	testCases := []struct {
		name string
		code string
	}{
		{
			name: "uninitialized variable access",
			code: `int main() { int x; return x; }`,
		},
		{
			name: "uninitialized variable in expression",
			code: `int main() { int x; int y = x + 5; return y; }`,
		},
		{
			name: "uninitialized variable in comparison",
			code: `int main() { int x; if (x > 0) { return 1; } return 0; }`,
		},
		{
			name: "array out of bounds positive",
			code: `int main() { int arr[3] = {1, 2, 3}; return arr[5]; }`,
		},
		{
			name: "array out of bounds negative",
			code: `int main() { int arr[3] = {1, 2, 3}; return arr[-1]; }`,
		},
		{
			name: "array out of bounds large",
			code: `int main() { int arr[2] = {1, 2}; return arr[1000]; }`,
		},
		{
			name: "uninitialized array element access",
			code: `int main() { int arr[5]; return arr[2]; }`,
		},
		{
			name: "uninitialized variable with modification",
			code: `int main() { int x; x += 10; return x; }`,
		},
		{
			name: "uninitialized variable in loop",
			code: `int main() { int sum; while (0) { sum += 1; } return sum; }`,
		},
		{
			name: "uninitialized variable in array index",
			code: `int main() { int arr[5] = {1, 2, 3, 4, 5}; int idx; return arr[idx]; }`,
		},
		{
			name: "multiple uninitialized variables",
			code: `int main() { int x; int y; int z = x + y; return z; }`,
		},
		{
			name: "uninitialized in conditional block",
			code: `int main() { int x; if (1) { int y = x; return y; } }`,
		},
		{
			name: "uninitialized after declaration no init",
			code: `int main() { int a; int b; int c = a; return c; }`,
		},
		{
			name: "array access with uninitialized index",
			code: `int main() { int arr[4] = {10, 20, 30, 40}; int i; return arr[i]; }`,
		},
		{
			name: "2d array out of bounds row",
			code: `int main() { int matrix[2][2] = {{1, 2}, {3, 4}}; return matrix[2][0]; }`,
		},
		{
			name: "2d array out of bounds col",
			code: `int main() { int matrix[2][2] = {{1, 2}, {3, 4}}; return matrix[0][2]; }`,
		},
		{
			name: "2d uninitialized element access",
			code: `int main() { int matrix[2][2]; return matrix[1][1]; }`,
		},
		{
			name: "2d array with uninitialized index",
			code: `int main() { int matrix[2][2] = {{1, 2}, {3, 4}}; int i; return matrix[i][1]; }`,
		},
		{
			name: "uninitialized variable in function call",
			code: `int test(int x) { return x * 2; } int main() { int y; return test(y); }`,
		},
		{
			name: "undefined variable access at global",
			code: `int main() { return undefined_var; }`,
		},
		{
			name: "division by zero in binary expression",
			code: `int main() { return 10 / 0; }`,
		},
		{
			name: "modulo by zero in binary expression",
			code: `int main() { return 17 % 0; }`,
		},
		{
			name: "division by zero in compound assignment",
			code: `int main() { int x = 20; x /= 0; return x; }`,
		},
		{
			name: "modulo by zero in compound assignment",
			code: `int main() { int x = 17; x %= 0; return x; }`,
		},
		{
			name: "pre increment on literal",
			code: `int main() { int x = ++5; return x; }`,
		},
		{
			name: "pre decrement on literal",
			code: `int main() { int x = --10; return x; }`,
		},
		{
			name: "post increment on literal",
			code: `int main() { int x = 5++; return x; }`,
		},
		{
			name: "post decrement on literal",
			code: `int main() { int x = 10--; return x; }`,
		},
		{
			name: "post increment on expression result",
			code: `int main() { return (5 + 2)++; }`,
		},
		{
			name: "pre decrement on expression result",
			code: `int main() { return --(3 * 2); }`,
		},
		{
			name: "call undefined function",
			code: `int main() { return undefined_function(5); }`,
		},
		{
			name: "call function with too few arguments",
			code: `int add(int a, int b) { return a + b; } int main() { return add(5); }`,
		},
		{
			name: "call function with too many arguments",
			code: `int add(int a, int b) { return a + b; } int main() { return add(5, 10, 15); }`,
		},
		{
			name: "call function expecting no args with args",
			code: `int get_constant() { return 42; } int main() { return get_constant(5); }`,
		},
		{
			name: "duplicate function definition",
			code: `int get_value() { return 10; } int get_value() { return 20; } int main() { return get_value(); }`,
		},
		{
			name: "missing main function",
			code: `int helper() { return 42; }`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			err := runCodeExpectError(t, tt.code)
			assert.NotNil(t, err, "expected an error for: %s", tt.name)
		})
	}
}

// TestGlobal tests global variables and arrays
func TestGlobal(t *testing.T) {
	tests := []testCase{
		{
			name:     "global variable simple",
			code:     `int x = 42; int main() { return x; }`,
			expected: 42,
		},
		{
			name:     "global variable modification",
			code:     `int x = 10; int main() { x = 20; return x; }`,
			expected: 20,
		},
		{
			name:     "global variable accessed in function",
			code:     `int x = 5; int test() { return x * 2; } int main() { return test(); }`,
			expected: 10,
		},
		{
			name:     "global variable modified by function",
			code:     `int x = 5; void modify() { x = 100; } int main() { modify(); return x; }`,
			expected: 100,
		},
		{
			name:     "multiple global variables",
			code:     `int x = 10; int y = 20; int main() { return x + y; }`,
			expected: 30,
		},
		{
			name:     "global variable in loop",
			code:     `int sum = 0; int main() { for (int i = 0; i < 5; i++) { sum += i; } return sum; }`,
			expected: 10,
		},
		{
			name:     "global variable used in condition",
			code:     `int threshold = 10; int main() { int x = 5; if (x < threshold) { return 1; } return 0; }`,
			expected: 1,
		},
		{
			name:     "global array declaration and access",
			code:     `int arr[3] = {10, 20, 30}; int main() { return arr[1]; }`,
			expected: 20,
		},
		{
			name:     "global array modification",
			code:     `int arr[3] = {1, 2, 3}; int main() { arr[0] = 100; return arr[0]; }`,
			expected: 100,
		},
		{
			name:     "global array accessed in function",
			code:     `int arr[4] = {10, 20, 30, 40}; int sum() { int s = 0; for (int i = 0; i < 4; i++) { s += arr[i]; } return s; } int main() { return sum(); }`,
			expected: 100,
		},
		{
			name:     "global array zero initialization",
			code:     `int arr[5]; int main() { return arr[0] + arr[2] + arr[4]; }`,
			expected: 0,
		},
		{
			name:     "global 2d array declaration and access",
			code:     `int matrix[2][3] = {{1, 2, 3}, {4, 5, 6}}; int main() { return matrix[1][2]; }`,
			expected: 6,
		},
		{
			name:     "global 2d array declaration without init",
			code:     `int matrix[2][2]; int main() { matrix[0][0] = 7; matrix[1][1] = 9; return matrix[0][0] + matrix[1][1]; }`,
			expected: 16,
		},
		{
			name:     "global 2d array zero initialization",
			code:     `int matrix[2][3]; int main() { return matrix[0][0] + matrix[0][2] + matrix[1][1]; }`,
			expected: 0,
		},
		{
			name:     "global 2d array declaration used in function",
			code:     `int matrix[2][2] = {{10, 20}, {30, 40}}; int get_sum() { return matrix[0][1] + matrix[1][0]; } int main() { return get_sum(); }`,
			expected: 50,
		},
		{
			name:     "global variable shadowed by local",
			code:     `int x = 100; int main() { int x = 50; return x; }`,
			expected: 50,
		},
		{
			name:     "global variable accessible after local scope",
			code:     `int x = 100; int main() { { int x = 50; } return x; }`,
			expected: 100,
		},
		{
			name:     "global array shadowed by local",
			code:     `int arr[2] = {10, 20}; int main() { int arr[2] = {100, 200}; return arr[0]; }`,
			expected: 100,
		},
		{
			name:     "multiple globals with functions",
			code:     `int x = 5; int y = 3; int multiply() { return x * y; } int add() { return x + y; } int main() { return multiply() + add(); }`,
			expected: 23,
		},
		{
			name:     "global array in multiple functions",
			code:     `int arr[3] = {1, 2, 3}; int first() { return arr[0]; } int second() { return arr[2]; } int main() { return first() + second(); }`,
			expected: 4,
		},
		{
			name:     "global variable with zero initialization",
			code:     `int x; int main() { x = 42; return x; }`,
			expected: 42,
		},
		{
			name:     "global array accessed via function",
			code:     `int arr[5] = {1, 2, 3, 4, 5}; int get(int idx) { return arr[idx]; } int main() { return get(2) + get(4); }`,
			expected: 8,
		},
		{
			name:     "recursive function with global variable",
			code:     `int factor = 2; int power(int n) { if (n <= 0) { return 1; } return factor * power(n - 1); } int main() { return power(3); }`,
			expected: 8,
		},
		{
			name:     "global variable modified in nested functions",
			code:     `int counter = 0; void inc() { counter++; } void add_two() { inc(); inc(); } int main() { add_two(); return counter; }`,
			expected: 2,
		},
		{
			name:     "global array with function returning element",
			code:     `int values[3] = {100, 200, 300}; int get_double(int i) { return values[i] * 2; } int main() { return get_double(1); }`,
			expected: 400,
		},
		{
			name:     "global variable in recursive addition",
			code:     `int multiplier = 10; int add_recursive(int n) { if (n <= 0) { return 0; } return multiplier + add_recursive(n - 1); } int main() { return add_recursive(3); }`,
			expected: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestDeclarations тестирует разные виды объявлений переменных
func TestDeclarations(t *testing.T) {
	tests := []testCase{
		{
			name:     "multiple variables no init",
			code:     `int main() { int a, b, c; a = 5; b = 10; c = 15; return a + b + c; }`,
			expected: 30,
		},
		{
			name:     "multiple variables with init",
			code:     `int main() { int x = 1, y = 2, z = 3; return x + y + z; }`,
			expected: 6,
		},
		{
			name:     "multiple variables mixed init",
			code:     `int main() { int a = 10, b, c = 30; b = 20; return a + b + c; }`,
			expected: 60,
		},
		{
			name:     "multiple variables in loop",
			code:     `int main() { int i, sum, product; i = 0; sum = 0; product = 1; for (i = 1; i <= 5; i++) { sum += i; product *= i; } return sum; }`,
			expected: 15,
		},
		{
			name:     "multiple variables in condition",
			code:     `int main() { int x, y, result; x = 5; y = 10; if (x < y) { result = x; } else { result = y; } return result; }`,
			expected: 5,
		},
		{
			name:     "global multiple variables",
			code:     `int x = 10, y = 20, z = 30; int main() { return x + y + z; }`,
			expected: 60,
		},
		{
			name:     "global multiple with no init",
			code:     `int a, b, c; int main() { a = 5; b = 10; c = 15; return a + b + c; }`,
			expected: 30,
		},
		{
			name:     "global multiple mixed init",
			code:     `int first = 100, second, third = 300; int main() { second = 200; return first + second + third; }`,
			expected: 600,
		},
		{
			name:     "multiple in function parameters and locals",
			code:     `int add(int a, int b) { int x, y, sum; x = a; y = b; sum = x + y; return sum; } int main() { return add(3, 4); }`,
			expected: 7,
		},
		{
			name:     "multiple declarations in nested scope",
			code:     `int main() { int a = 5, b = 10; { int x, y; x = a; y = b; return x + y; } }`,
			expected: 15,
		},
		{
			name:     "multiple variables in while loop",
			code:     `int main() { int i, sum; i = 0; sum = 0; while (i < 5) { sum += i; i++; } return sum; }`,
			expected: 10,
		},
		{
			name:     "multiple variables in do-while",
			code:     `int main() { int i, count; i = 1; count = 0; do { count++; i++; } while (i <= 3); return count; }`,
			expected: 3,
		},
		{
			name:     "multiple declarations combined assignment",
			code:     `int main() { int a, b, c; c = 5; b = c; a = b; return a + b + c; }`,
			expected: 15,
		},
		{
			name:     "multiple with compound operators",
			code:     `int main() { int x = 10, y = 5, z = 2; x += y; y *= z; z -= 1; return x + y + z; }`,
			expected: 26,
		},
		{
			name:     "multiple variables used in expression",
			code:     `int main() { int a = 2, b = 3, c = 4; return a * b + c; }`,
			expected: 10,
		},
		{
			name:     "global multiple in functions",
			code:     `int a = 2, b = 3, c = 4; int calc() { return a * b + c; } int main() { return calc(); }`,
			expected: 10,
		},
		{
			name:     "multiple declarations with recursion",
			code:     `int factorial(int n) { int result; if (n <= 1) { result = 1; } else { result = n * factorial(n - 1); } return result; } int main() { int x, y; x = 5; y = factorial(x); return y; }`,
			expected: 120,
		},
		{
			name:     "multiple variables in array initialization",
			code:     `int main() { int i, sum, arr[3]; i = 0; sum = 0; arr[0] = 10; arr[1] = 20; arr[2] = 30; for (i = 0; i < 3; i++) { sum += arr[i]; } return sum; }`,
			expected: 60,
		},
		{
			name:     "multiple in complex expression",
			code:     `int main() { int x = 1, y = 2, z = 3; int result = (x + y) * (z - 1) + x; return result; }`,
			expected: 7,
		},
		{
			name:     "multiple with logical operators",
			code:     `int main() { int a = 5, b = 10, c = 15; int result; if ((a < b) && (b < c)) { result = 1; } else { result = 0; } return result; }`,
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

// TestMultipleArrayDeclarations тестирует объявления нескольких массивов в одном операторе
func TestMultipleArrayDeclarations(t *testing.T) {
	tests := []testCase{
		{
			name:     "multiple arrays no init",
			code:     `int main() { int a[2], b[3]; a[0] = 10; a[1] = 20; b[0] = 30; b[1] = 40; b[2] = 50; return a[0] + a[1] + b[0] + b[1] + b[2]; }`,
			expected: 150,
		},
		{
			name:     "multiple arrays with init",
			code:     `int main() { int a[2] = {1, 2}, b[3] = {3, 4, 5}; return a[0] + a[1] + b[0] + b[1] + b[2]; }`,
			expected: 15,
		},
		{
			name:     "multiple arrays mixed init",
			code:     `int main() { int a[2] = {10, 20}, b[2], c[2] = {30, 40}; b[0] = 25; b[1] = 35; return a[0] + b[0] + c[0] + c[1]; }`,
			expected: 105,
		},
		{
			name:     "multiple arrays in loop",
			code:     `int main() { int a[3], b[3]; int i, sum = 0; for (i = 0; i < 3; i++) { a[i] = i + 1; b[i] = (i + 1) * 10; sum += a[i] + b[i]; } return sum; }`,
			expected: 66,
		},
		{
			name:     "global multiple arrays",
			code:     `int arr[2] = {5, 10}, brr[3] = {1, 2, 3}; int main() { return arr[0] + arr[1] + brr[0] + brr[1] + brr[2]; }`,
			expected: 21,
		},
		{
			name:     "global multiple no init",
			code:     `int a[2], b[3]; int main() { a[0] = 100; a[1] = 200; b[0] = 1; b[1] = 2; b[2] = 3; return a[0] + a[1] + b[0] + b[1] + b[2]; }`,
			expected: 306,
		},
		{
			name:     "variable and array together",
			code:     `int main() { int x = 10, arr[2] = {5, 15}; return x + arr[0] + arr[1]; }`,
			expected: 30,
		},
		{
			name:     "variable arrays and variables",
			code:     `int main() { int x = 5, a[2] = {10, 20}, y = 15; return x + a[0] + a[1] + y; }`,
			expected: 50,
		},
		{
			name:     "global variables and arrays",
			code:     `int x = 10, arr[2] = {5, 15}, y = 20; int main() { return x + arr[0] + arr[1] + y; }`,
			expected: 50,
		},
		{
			name:     "multiple arrays accessed in function",
			code:     `int get_sum() { int a[2] = {1, 2}, b[2] = {3, 4}; return a[0] + a[1] + b[0] + b[1]; } int main() { return get_sum(); }`,
			expected: 10,
		},
		{
			name:     "2d and 1d arrays together",
			code:     `int main() { int a[2] = {10, 20}, m[2][2] = {{1, 2}, {3, 4}}; return a[0] + m[1][1]; }`,
			expected: 14,
		},
		{
			name:     "multiple 2d arrays",
			code:     `int main() { int m1[2][2] = {{1, 2}, {3, 4}}, m2[2][2] = {{5, 6}, {7, 8}}; return m1[0][0] + m2[1][1]; }`,
			expected: 9,
		},
		{
			name:     "multiple 2d no init",
			code:     `int main() { int m1[2][2], m2[2][2]; m1[0][0] = 5; m1[1][1] = 10; m2[0][0] = 20; m2[1][1] = 30; return m1[0][0] + m1[1][1] + m2[0][0] + m2[1][1]; }`,
			expected: 65,
		},
		{
			name:     "global multiple 2d arrays",
			code:     `int m1[2][2] = {{10, 20}, {30, 40}}, m2[2][2] = {{1, 2}, {3, 4}}; int main() { return m1[0][0] + m2[1][1]; }`,
			expected: 14,
		},
		{
			name:     "multiple arrays in nested scopes",
			code:     `int main() { int a[2] = {1, 2}; { int b[2] = {3, 4}; int c[2]; c[0] = 5; c[1] = 6; return a[0] + b[0] + c[1]; } }`,
			expected: 10,
		},
		{
			name:     "multiple arrays modification in loop",
			code:     `int main() { int a[2] = {1, 1}, b[2] = {1, 1}; int i; for (i = 0; i < 2; i++) { a[i] *= 5; b[i] *= 10; } return a[0] + a[1] + b[0] + b[1]; }`,
			expected: 30,
		},
		{
			name:     "multiple with array expression",
			code:     `int main() { int a[3] = {2, 3, 4}, b[3] = {1, 1, 1}; int i, sum = 0; for (i = 0; i < 3; i++) { sum += a[i] * b[i]; } return sum; }`,
			expected: 9,
		},
		{
			name:     "global mixed multiple declarations",
			code:     `int x = 5, arr[2] = {10, 20}, y = 15, brr[3] = {1, 2, 3}; int main() { return x + arr[0] + y + brr[2]; }`,
			expected: 33,
		},
		{
			name:     "multiple arrays in condition",
			code:     `int main() { int a[2] = {5, 10}, b[2] = {7, 3}; int result; if (a[0] < a[1] && b[0] > b[1]) { result = 100; } else { result = 0; } return result; }`,
			expected: 100,
		},
		{
			name:     "multiple arrays parameter and local",
			code:     `int process(int val) { int a[2] = {1, 2}, b[2] = {3, 4}; return val + a[0] + b[1]; } int main() { return process(10); }`,
			expected: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := runCode(t, tt.code)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

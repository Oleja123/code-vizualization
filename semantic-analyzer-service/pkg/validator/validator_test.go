package validator

import (
	"testing"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariableTypes(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantError bool
		errCode   ErrorCode
	}{
		   {
			   name: "valid int variable",
			   code: `int main() { int a = 0; int b[2][2]; return a; }`,
			   wantError: false,
		   },
		   {
			   name: "valid 2d array",
			   code: `int main() { int arr[2][2]; arr[0][1] = 5; return arr[0][1]; }`,
			   wantError: false,
		   },
		   {
			   name: "invalid variable type (pointer)",
			   code: `int main() { int* a; return 0; }`,
			   wantError: true,
			   errCode: ErrInvalidType,
		   },
		   {
			   name: "invalid variable type (array 3d)",
			   code: `int main() { int a[2][2][2]; return 0; }`,
			   wantError: true,
			   errCode: ErrInvalidType,
		   },
		   {
			   name: "invalid variable type (float)",
			   code: `int main() { float a = 1; return 0; }`,
			   wantError: true,
			   errCode: ErrInvalidType,
		   },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.NewCConverter()
			tree, err := conv.Parse([]byte(tt.code))
			require.NoError(t, err, "Parse failed")
			program, err := conv.ConvertToProgram(tree, []byte(tt.code))
			require.NoError(t, err, "ConvertToProgram failed")
			prog := program.(*converter.Program)
			validator := New()
			err = validator.ValidateProgram(prog)
			if tt.wantError {
				assert.Error(t, err)
				require.NotNil(t, err)
				semErr, ok := err.(*SemanticError)
				assert.True(t, ok)
				if ok {
					assert.Equal(t, tt.errCode, semErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAssignmentOperators(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantError bool
		errCode   ErrorCode
	}{
		   {
			   name: "valid assignment operators",
			   code: `int main() { int a = 5; a += 10; a -= 3; a /= 4; a %= 3; a = 1; return a; }`,
			   wantError: false,
		   },
		   {
			   name: "invalid assignment operator (^=)",
			   code: `int main() { int a = 5; a ^= 2; return a; }`,
			   wantError: true,
			   errCode: ErrUnsupportedAssignOp,
		   },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.NewCConverter()
			tree, err := conv.Parse([]byte(tt.code))
			require.NoError(t, err, "Parse failed")
			program, err := conv.ConvertToProgram(tree, []byte(tt.code))
			require.NoError(t, err, "ConvertToProgram failed")
			prog := program.(*converter.Program)
			validator := New()
			err = validator.ValidateProgram(prog)
			if tt.wantError {
				assert.Error(t, err)
				require.NotNil(t, err)
				semErr, ok := err.(*SemanticError)
				assert.True(t, ok)
				if ok {
					assert.Equal(t, tt.errCode, semErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBinaryOperators(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantError bool
		errCode   ErrorCode
	}{
		   {
			   name: "valid binary operators",
			   code: `int main() { int a = 5; int b = 3; int c1 = a + b; int c2 = a - b; int c3 = a * b; int c4 = a / b; int c5 = a % b; int c6 = (a == b); int c7 = (a != b); int c8 = (a < b); int c9 = (a <= b); int c10 = (a > b); int c11 = (a >= b); int c12 = (a && b); int c13 = (a || b); return c1; }`,
			   wantError: false,
		   },
		   {
			   name: "valid binary with assignment",
			   code: `int main() { int a = 5; int b = 3; a = a + b; a = a - b; a = a * b; a = a / b; a = a % b; return a; }`,
			   wantError: false,
		   },
		   {
			   name: "unsupported binary operator (&)",
			   code: `int main() { int a = 5; int b = a & 3; return b; }`,
			   wantError: true,
			   errCode: ErrUnsupportedBinaryOp,
		   },
		   {
			   name: "bitwise_and",
			   code: `int main() { int a = 5; int b = 3; int c = a & b; return c; }`,
			   wantError: true,
			   errCode: ErrUnsupportedBinaryOp,
		   },
		   {
			   name: "bitwise_or",
			   code: `int main() { int a = 5; int b = 3; int c = a | b; return c; }`,
			   wantError: true,
			   errCode: ErrUnsupportedBinaryOp,
		   },
		   {
			   name: "bitwise_xor",
			   code: `int main() { int a = 5; int b = 3; int c = a ^ b; return c; }`,
			   wantError: true,
			   errCode: ErrUnsupportedBinaryOp,
		   },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.NewCConverter()
			tree, err := conv.Parse([]byte(tt.code))
			require.NoError(t, err, "Parse failed")
			program, err := conv.ConvertToProgram(tree, []byte(tt.code))
			require.NoError(t, err, "ConvertToProgram failed")
			prog := program.(*converter.Program)
			validator := New()
			err = validator.ValidateProgram(prog)
			if tt.wantError {
				assert.Error(t, err)
				require.NotNil(t, err)
				semErr, ok := err.(*SemanticError)
				assert.True(t, ok)
				if ok {
					assert.Equal(t, tt.errCode, semErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUnaryOperators(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantError bool
		errCode   ErrorCode
	}{
		   {
			   name: "valid unary operators",
			   code: `int main() { int a = 5; int b = -a; int d = !0; int e = ++a; int f = a++; int g = --a; int h = a--; return b; }`,
			   wantError: false,
		   },
		   {
			   name: "valid unary with assignment",
			   code: `int main() { int a = 5; a++; ++a; a--; --a; return a; }`,
			   wantError: false,
		   },
		   {
			   name: "invalid unary operator (~)",
			   code: `int main() { int a = 5; int b = ~a; return b; }`,
			   wantError: true,
			   errCode: ErrUnsupportedUnaryOp,
		   },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.NewCConverter()
			tree, err := conv.Parse([]byte(tt.code))
			require.NoError(t, err, "Parse failed")
			program, err := conv.ConvertToProgram(tree, []byte(tt.code))
			require.NoError(t, err, "ConvertToProgram failed")
			prog := program.(*converter.Program)
			validator := New()
			err = validator.ValidateProgram(prog)
			if tt.wantError {
				assert.Error(t, err)
				require.NotNil(t, err)
				semErr, ok := err.(*SemanticError)
				assert.True(t, ok)
				if ok {
					assert.Equal(t, tt.errCode, semErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestControlStructures(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantError bool
		errCode   ErrorCode
	}{
		   {
			   name: "valid control structures (if/else)",
			   code: `int main() { int x = 0; if (x < 10) { x++; } else { x--; } return x; }`,
			   wantError: false,
		   },
		   {
			   name: "valid control structures (for)",
			   code: `int main() { int x = 0; for (int i = 0; i < 5; i++) { x += i; } return x; }`,
			   wantError: false,
		   },
		   {
			   name: "valid control structures (while)",
			   code: `int main() { int x = 0; while (x < 10) { x++; } return x; }`,
			   wantError: false,
		   },
		   {
			   name: "valid control structures (do while)",
			   code: `int main() { int x = 0; do { x++; } while (x < 10); return x; }`,
			   wantError: false,
		   },
		   {
			   name: "valid control structures (continue/break/goto)",
			   code: `int main() { int x = 0; for (int i = 0; i < 5; i++) { if (i == 2) continue; if (i == 4) break; } goto label; label: x++; return x; }`,
			   wantError: false,
		   },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.NewCConverter()
			tree, err := conv.Parse([]byte(tt.code))
			require.NoError(t, err, "Parse failed")
			program, err := conv.ConvertToProgram(tree, []byte(tt.code))
			require.NoError(t, err, "ConvertToProgram failed")
			prog := program.(*converter.Program)
			validator := New()
			err = validator.ValidateProgram(prog)
			if tt.wantError {
				assert.Error(t, err)
				require.NotNil(t, err)
				semErr, ok := err.(*SemanticError)
				assert.True(t, ok)
				if ok {
					assert.Equal(t, tt.errCode, semErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFunctionTypes(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantError bool
		errCode   ErrorCode
	}{
		   {
			   name: "valid function return types",
			   code: `int getNumber() { return 42; } void printNumber() { return; } int main() { int x = getNumber(); printNumber(); return x; }`,
			   wantError: false,
		   },
		   {
			   name: "valid void return",
			   code: `void foo() { return; } int main() { foo(); return 0; }`,
			   wantError: false,
		   },
		   {
			   name: "valid int return",
			   code: `int bar() { return 1; } int main() { return bar(); }`,
			   wantError: false,
		   },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.NewCConverter()
			tree, err := conv.Parse([]byte(tt.code))
			require.NoError(t, err, "Parse failed")
			program, err := conv.ConvertToProgram(tree, []byte(tt.code))
			require.NoError(t, err, "ConvertToProgram failed")
			prog := program.(*converter.Program)
			validator := New()
			err = validator.ValidateProgram(prog)
			if tt.wantError {
				assert.Error(t, err)
				require.NotNil(t, err)
				semErr, ok := err.(*SemanticError)
				assert.True(t, ok)
				if ok {
					assert.Equal(t, tt.errCode, semErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFunctionParameters(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantError bool
		errCode   ErrorCode
	}{
		   {
			   name: "valid function parameter",
			   code: `int sum(int a, int b) { return a + b; } int main() { int x = sum(1,2); return x; }`,
			   wantError: false,
		   },
		   {
			   name: "valid function parameter (single)",
			   code: `int inc(int a) { return a + 1; } int main() { return inc(5); }`,
			   wantError: false,
		   },
		   {
			   name: "valid function parameter (multiple calls)",
			   code: `int sum(int a, int b) { return a + b; } int main() { int x = sum(2,3); int y = sum(4,5); return x + y; }`,
			   wantError: false,
		   },
		   {
			   name: "invalid function parameter (pointer)",
			   code: `int sum(int* arr) { return arr[0] + arr[1]; } int main() {  }`,
			   wantError: true,
			   errCode: ErrInvalidType,
		   },
		   {
			   name: "invalid function parameter (array)",
			   code: `int sum(int arr[2]) { return arr[0] + arr[1]; } int main() {  }`,
			   wantError: true,
			   errCode: ErrInvalidType,
		   },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.NewCConverter()
			tree, err := conv.Parse([]byte(tt.code))
			require.NoError(t, err, "Parse failed")
			program, err := conv.ConvertToProgram(tree, []byte(tt.code))
			require.NoError(t, err, "ConvertToProgram failed")
			prog := program.(*converter.Program)
			validator := New()
			err = validator.ValidateProgram(prog)
			if tt.wantError {
				assert.Error(t, err)
				require.NotNil(t, err)
				semErr, ok := err.(*SemanticError)
				assert.True(t, ok)
				if ok {
					assert.Equal(t, tt.errCode, semErr.Code)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

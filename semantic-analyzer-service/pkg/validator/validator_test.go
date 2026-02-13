package validator

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
	"github.com/Oleja123/code-vizualization/semantic-analyzer-service/internal/domain/structs"
)

func TestVariableTypes(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		wantError bool
		errCode   structs.ErrorCode
	}{
		{
			name: "valid int variable",
			code: `int main() { int a = 0; return a; }`,
			wantError: false,
		},
		{
			name: "invalid variable type (pointer)",
			code: `int main() { int* a; return 0; }`,
			wantError: true,
			errCode: structs.ErrInvalidVariableType,
		},
		{
			name: "invalid variable type (array 3d)",
			code: `int main() { int a[2][2][2]; return 0; }`,
			wantError: true,
			errCode: structs.ErrInvalidVariableType,
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
				semErr, ok := err.(*structs.SemanticError)
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
		errCode   structs.ErrorCode
	}{
		{
			name: "valid assignment operators",
			code: `int main() { int a = 5; a += 10; a -= 3; a /= 4; a %= 3; return a; }`,
			wantError: false,
		},
		{
			name: "invalid assignment operator",
			code: `int main() { int a = 5; a *= 2; return a; }`,
			wantError: true,
			errCode: structs.ErrUnsupportedAssignOp,
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
				semErr, ok := err.(*structs.SemanticError)
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
		errCode   structs.ErrorCode
	}{
		{
			name: "valid binary operators",
			code: `int main() { int a = 5; int b = 3; int c1 = a + b; int c2 = a - b; int c3 = a * b; int c4 = a / b; int c5 = a % b; int c6 = (a == b); int c7 = (a != b); int c8 = (a < b); int c9 = (a <= b); int c10 = (a > b); int c11 = (a >= b); int c12 = (a && b); int c13 = (a || b); return c1; }`,
			wantError: false,
		},
		{
			name: "unsupported binary operator (&)",
			code: `int main() { int a = 5; int b = a & 3; return b; }`,
			wantError: true,
			errCode: structs.ErrUnsupportedBinaryOp,
		},
		{
			name: "bitwise_and",
			code: `int main() { int a = 5; int b = 3; int c = a & b; return c; }`,
			wantError: true,
			errCode: structs.ErrUnsupportedBinaryOp,
		},
		{
			name: "bitwise_or",
			code: `int main() { int a = 5; int b = 3; int c = a | b; return c; }`,
			wantError: true,
			errCode: structs.ErrUnsupportedBinaryOp,
		},
		{
			name: "bitwise_xor",
			code: `int main() { int a = 5; int b = 3; int c = a ^ b; return c; }`,
			wantError: true,
			errCode: structs.ErrUnsupportedBinaryOp,
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
				semErr, ok := err.(*structs.SemanticError)
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
		errCode   structs.ErrorCode
	}{
		{
			name: "valid unary operators",
			code: `int main() { int a = 5; int b = -a; int d = !0; int e = ++a; int f = a++; int g = --a; int h = a--; return b; }`,
			wantError: false,
		},
		{
			name: "invalid unary operator (~)",
			code: `int main() { int a = 5; int b = ~a; return b; }`,
			wantError: true,
			errCode: structs.ErrUnsupportedUnaryOp,
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
				semErr, ok := err.(*structs.SemanticError)
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
		errCode   structs.ErrorCode
	}{
		{
			name: "valid control structures",
			code: `int main() { int x = 0; if (x < 10) { x++; } else { x--; } for (int i = 0; i < 5; i++) { x += i; } while (x > 0) { x--; } do { x++; } while (x < 10); continue; break; goto label; label: x++; return x; }`,
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
				semErr, ok := err.(*structs.SemanticError)
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
		errCode   structs.ErrorCode
	}{
		{
			name: "valid function return types",
			code: `int getNumber() { return 42; } void printNumber() { return; } int main() { int x = getNumber(); printNumber(); return x; }`,
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
				semErr, ok := err.(*structs.SemanticError)
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
		errCode   structs.ErrorCode
	}{
		{
			name: "valid function parameter",
			code: `int sum(int a, int b) { return a + b; } int main() { int x = sum(1,2); return x; }`,
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
				semErr, ok := err.(*structs.SemanticError)
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

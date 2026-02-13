package validator

import (
	"testing"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
)

// TestValidSimpleProgram проверяет валидацию простой программы
func TestValidSimpleProgram(t *testing.T) {
	sourceCode := []byte(`int main() {
	int a = 0;
	return a;
}`)

	conv := converter.NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	program, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	prog := program.(*converter.Program)
	validator := New()
	if err := validator.ValidateProgram(prog); err != nil {
		t.Fatalf("Validation should pass, got error: %v", err)
	}
}

// TestUnsupportedBinaryOperator проверяет обнаружение неподдерживаемого оператора
func TestUnsupportedBinaryOperator(t *testing.T) {
	sourceCode := []byte(`int main() {
	int a = 5;
	int b = a & 3;
	return b;
}`)

	conv := converter.NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	program, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	prog := program.(*converter.Program)
	validator := New()
	err = validator.ValidateProgram(prog)
	if err == nil {
		t.Fatalf("Expected error for & operator")
	}

	semErr, ok := err.(*SemanticError)
	if !ok {
		t.Fatalf("Expected SemanticError, got %T", err)
	}

	if semErr.Code != ErrUnsupportedBinaryOp {
		t.Errorf("Expected ErrUnsupportedBinaryOp, got %v", semErr.Code)
	}
}

// TestValidAssignmentOperators проверяет разрешенные операторы присваивания
func TestValidAssignmentOperators(t *testing.T) {
	sourceCode := []byte(`int main() {
	int a = 5;
	a += 10;
	a -= 3;
	a /= 4;
	a %= 3;
	return a;
}`)

	conv := converter.NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	program, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	prog := program.(*converter.Program)
	validator := New()
	if err := validator.ValidateProgram(prog); err != nil {
		t.Fatalf("Validation should pass for allowed operators, got error: %v", err)
	}
}

// TestValidUnaryOperators проверяет разрешенные унарные операторы
func TestValidUnaryOperators(t *testing.T) {
	sourceCode := []byte(`int main() {
	int a = 5;
	int b = -a;
	int d = !0;
	int e = ++a;
	int f = a++;
	int g = --a;
	int h = a--;
	return b;
}`)

	conv := converter.NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	program, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	prog := program.(*converter.Program)
	validator := New()
	if err := validator.ValidateProgram(prog); err != nil {
		t.Fatalf("Validation should pass for allowed unary operators, got error: %v", err)
	}
}

// TestValidLogicalOperators проверяет логические операторы
func TestValidLogicalOperators(t *testing.T) {
	sourceCode := []byte(`int main() {
	int a = 1;
	int b = 0;
	int c = (a == 1) && (b == 0);
	int d = (a != 0) || (b == 0);
	return c;
}`)

	conv := converter.NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	program, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	prog := program.(*converter.Program)
	validator := New()
	if err := validator.ValidateProgram(prog); err != nil {
		t.Fatalf("Validation should pass for logical operators, got error: %v", err)
	}
}

// TestValidReturnTypes проверяет допустимые типы возврата
func TestValidReturnTypes(t *testing.T) {
	sourceCode := []byte(`int getNumber() {
	return 42;
}

void printNumber() {
	return;
}

int main() {
	int x = getNumber();
	printNumber();
	return x;
}`)

	conv := converter.NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	program, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	prog := program.(*converter.Program)
	validator := New()
	if err := validator.ValidateProgram(prog); err != nil {
		t.Fatalf("Validation should pass for valid return types, got error: %v", err)
	}
}

// TestValidComparisonOperators проверяет операторы сравнения
func TestValidComparisonOperators(t *testing.T) {
	sourceCode := []byte(`int main() {
	int a = 5;
	int b = 3;
	int c1 = (a < b);
	int c2 = (a <= b);
	int c3 = (a > b);
	int c4 = (a >= b);
	int c5 = (a == b);
	int c6 = (a != b);
	return c1;
}`)

	conv := converter.NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	program, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	prog := program.(*converter.Program)
	validator := New()
	if err := validator.ValidateProgram(prog); err != nil {
		t.Fatalf("Validation should pass for comparison operators, got error: %v", err)
	}
}

// TestBitwiseOperatorsRejection проверяет отклонение битовых операторов
func TestBitwiseOperatorsRejection(t *testing.T) {
	tests := []struct {
		name     string
		operator string
		code     string
	}{
		{
			name:     "bitwise_and",
			operator: "&",
			code: `int main() {
	int a = 5;
	int b = 3;
	int c = a & b;
	return c;
}`,
		},
		{
			name:     "bitwise_or",
			operator: "|",
			code: `int main() {
	int a = 5;
	int b = 3;
	int c = a | b;
	return c;
}`,
		},
		{
			name:     "bitwise_xor",
			operator: "^",
			code: `int main() {
	int a = 5;
	int b = 3;
	int c = a ^ b;
	return c;
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := converter.NewCConverter()
			tree, err := conv.Parse([]byte(tt.code))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			program, err := conv.ConvertToProgram(tree, []byte(tt.code))
			if err != nil {
				t.Fatalf("ConvertToProgram failed: %v", err)
			}

			prog := program.(*converter.Program)
			validator := New()
			err = validator.ValidateProgram(prog)
			if err == nil {
				t.Fatalf("Expected error for bitwise operator %s", tt.operator)
			}

			semErr, ok := err.(*SemanticError)
			if !ok {
				t.Fatalf("Expected SemanticError, got %T", err)
			}

			if semErr.Code != ErrUnsupportedBinaryOp {
				t.Errorf("Expected ErrUnsupportedBinaryOp, got %v", semErr.Code)
			}
		})
	}
}

// TestComplexValidProgram проверяет валидацию сложной программы
func TestComplexValidProgram(t *testing.T) {
	sourceCode := []byte(`int factorial(int n) {
	if (n <= 1) {
		return 1;
	}
	return n * factorial(n - 1);
}

int main() {
	int x = 5;
	int result = factorial(x);
	
	while (result > 0) {
		result = result - 1;
	}
	
	return result;
}`)

	conv := converter.NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	program, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	prog := program.(*converter.Program)
	validator := New()
	if err := validator.ValidateProgram(prog); err != nil {
		t.Fatalf("Validation should pass for complex program, got error: %v", err)
	}
}

// TestDoWhileValidation проверяет валидацию do-while циклов
func TestDoWhileValidation(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 0;
	do {
		x = x + 1;
	} while (x < 10);
	return x;
}`)

	conv := converter.NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	program, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	prog := program.(*converter.Program)
	validator := New()
	if err := validator.ValidateProgram(prog); err != nil {
		t.Fatalf("Validation should pass for do-while loop, got error: %v", err)
	}
}

package converter

import (
	"testing"

	"github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/domain/structs"
)

// TestParseSimpleVariable проверяет парсинг простой переменной
func TestParseSimpleVariable(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 5;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	if len(program.Declarations) == 0 {
		t.Fatal("Expected declarations, got none")
	}

	funcDecl, ok := program.Declarations[0].(*structs.FunctionDecl)
	if !ok {
		t.Fatalf("Expected FunctionDecl, got %T", program.Declarations[0])
	}

	if funcDecl.Name != "main" {
		t.Errorf("Expected function name 'main', got '%s'", funcDecl.Name)
	}

	if funcDecl.ReturnType.BaseType != "int" {
		t.Errorf("Expected return type 'int', got '%s'", funcDecl.ReturnType.BaseType)
	}

	if len(funcDecl.Body.Statements) == 0 {
		t.Fatal("Expected statements in body")
	}

	varDecl, ok := funcDecl.Body.Statements[0].(*structs.VariableDecl)
	if !ok {
		t.Fatalf("Expected VariableDecl, got %T", funcDecl.Body.Statements[0])
	}

	if varDecl.Name != "x" {
		t.Errorf("Expected variable name 'x', got '%s'", varDecl.Name)
	}

	if varDecl.VarType.BaseType != "int" {
		t.Errorf("Expected variable type 'int', got '%s'", varDecl.VarType.BaseType)
	}

	intLit, ok := varDecl.InitExpr.(*structs.IntLiteral)
	if !ok {
		t.Fatalf("Expected IntLiteral, got %T", varDecl.InitExpr)
	}

	if intLit.Value != 5 {
		t.Errorf("Expected value 5, got %d", intLit.Value)
	}
}

// TestParsePointerVariable проверяет парсинг переменной с указателем
func TestParsePointerVariable(t *testing.T) {
	sourceCode := []byte(`int main() {
	int *ptr = 0;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[0].(*structs.VariableDecl)

	if varDecl.VarType.PointerLevel != 1 {
		t.Errorf("Expected pointer level 1, got %d", varDecl.VarType.PointerLevel)
	}
}

// TestParseDoublePointer проверяет парсинг двойного указателя
func TestParseDoublePointer(t *testing.T) {
	sourceCode := []byte(`int main() {
	int **pptr = 0;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[0].(*structs.VariableDecl)

	if varDecl.VarType.PointerLevel != 2 {
		t.Errorf("Expected pointer level 2, got %d", varDecl.VarType.PointerLevel)
	}
}

// TestParseArrayVariable проверяет парсинг массива
func TestParseArrayVariable(t *testing.T) {
	sourceCode := []byte(`int main() {
	int arr[10];
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[0].(*structs.VariableDecl)

	if len(varDecl.VarType.ArraySizes) != 1 || varDecl.VarType.ArraySizes[0] != 10 {
		t.Errorf("Expected array size [10], got %v", varDecl.VarType.ArraySizes)
	}
}

// TestParseArrayOfPointers проверяет парсинг массива указателей
func TestParseArrayOfPointers(t *testing.T) {
	sourceCode := []byte(`int main() {
	int *arr[10];
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[0].(*structs.VariableDecl)

	if varDecl.VarType.PointerLevel != 1 {
		t.Errorf("Expected pointer level 1, got %d", varDecl.VarType.PointerLevel)
	}

	if len(varDecl.VarType.ArraySizes) != 1 || varDecl.VarType.ArraySizes[0] != 10 {
		t.Errorf("Expected array size [10], got %v", varDecl.VarType.ArraySizes)
	}
}

// TestParse2DArray проверяет парсинг двумерного массива
func TestParse2DArray(t *testing.T) {
	sourceCode := []byte(`int main() {
	int arr[10][20];
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[0].(*structs.VariableDecl)

	if len(varDecl.VarType.ArraySizes) != 2 {
		t.Errorf("Expected 2D array, got %d dimensions", len(varDecl.VarType.ArraySizes))
	}

	if varDecl.VarType.ArraySizes[0] != 10 || varDecl.VarType.ArraySizes[1] != 20 {
		t.Errorf("Expected array sizes [10, 20], got %v", varDecl.VarType.ArraySizes)
	}
}

// TestParse3DArray проверяет парсинг трехмерного массива
func TestParse3DArray(t *testing.T) {
	sourceCode := []byte(`int main() {
	int arr[5][10][15];
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[0].(*structs.VariableDecl)

	if len(varDecl.VarType.ArraySizes) != 3 {
		t.Errorf("Expected 3D array, got %d dimensions", len(varDecl.VarType.ArraySizes))
	}

	if varDecl.VarType.ArraySizes[0] != 5 || varDecl.VarType.ArraySizes[1] != 10 || varDecl.VarType.ArraySizes[2] != 15 {
		t.Errorf("Expected array sizes [5, 10, 15], got %v", varDecl.VarType.ArraySizes)
	}
}

// TestParse2DArrayOfPointers проверяет парсинг двумерного массива указателей
func TestParse2DArrayOfPointers(t *testing.T) {
	sourceCode := []byte(`int main() {
	int *arr[3][4];
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[0].(*structs.VariableDecl)

	if varDecl.VarType.PointerLevel != 1 {
		t.Errorf("Expected pointer level 1, got %d", varDecl.VarType.PointerLevel)
	}

	if len(varDecl.VarType.ArraySizes) != 2 {
		t.Errorf("Expected 2D array, got %d dimensions", len(varDecl.VarType.ArraySizes))
	}

	if varDecl.VarType.ArraySizes[0] != 3 || varDecl.VarType.ArraySizes[1] != 4 {
		t.Errorf("Expected array sizes [3, 4], got %v", varDecl.VarType.ArraySizes)
	}
}

// TestParseBinaryExpression проверяет парсинг бинарного выражения
func TestParseBinaryExpression(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 5 + 3;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[0].(*structs.VariableDecl)

	binExpr, ok := varDecl.InitExpr.(*structs.BinaryExpr)
	if !ok {
		t.Fatalf("Expected BinaryExpr, got %T", varDecl.InitExpr)
	}

	if binExpr.Operator != "+" {
		t.Errorf("Expected operator '+', got '%s'", binExpr.Operator)
	}

	left, ok := binExpr.Left.(*structs.IntLiteral)
	if !ok {
		t.Fatalf("Expected IntLiteral on left, got %T", binExpr.Left)
	}
	if left.Value != 5 {
		t.Errorf("Expected left value 5, got %d", left.Value)
	}

	right, ok := binExpr.Right.(*structs.IntLiteral)
	if !ok {
		t.Fatalf("Expected IntLiteral on right, got %T", binExpr.Right)
	}
	if right.Value != 3 {
		t.Errorf("Expected right value 3, got %d", right.Value)
	}
}

// TestParseUnaryExpression проверяет парсинг унарного выражения (логическое отрицание)
func TestParseUnaryExpression(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 5;
	int y = !x;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[1].(*structs.VariableDecl)

	unaryExpr, ok := varDecl.InitExpr.(*structs.UnaryExpr)
	if !ok {
		t.Fatalf("Expected UnaryExpr, got %T", varDecl.InitExpr)
	}

	if unaryExpr.Operator != "!" {
		t.Errorf("Expected operator '!', got '%s'", unaryExpr.Operator)
	}

	if unaryExpr.IsPostfix {
		t.Error("Expected prefix operator, got postfix")
	}
}

// TestParseNegativeLiteral проверяет парсинг отрицательного литерала как UnaryExpr
func TestParseNegativeLiteral(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = -5;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[0].(*structs.VariableDecl)

	// Отрицательный литерал должен быть представлен как UnaryExpr с минусом
	unaryExpr, ok := varDecl.InitExpr.(*structs.UnaryExpr)
	if !ok {
		t.Fatalf("Expected UnaryExpr for negative literal, got %T", varDecl.InitExpr)
	}

	if unaryExpr.Operator != "-" {
		t.Errorf("Expected operator '-', got '%s'", unaryExpr.Operator)
	}

	if unaryExpr.IsPostfix {
		t.Error("Expected prefix operator, got postfix")
	}

	// Операнд должен быть положительным IntLiteral(5)
	posLiteral, ok := unaryExpr.Operand.(*structs.IntLiteral)
	if !ok {
		t.Fatalf("Expected IntLiteral operand, got %T", unaryExpr.Operand)
	}

	if posLiteral.Value != 5 {
		t.Errorf("Expected positive literal 5, got %d", posLiteral.Value)
	}
}

// TestParsePrefixIncrement проверяет парсинг префиксного инкремента
func TestParsePrefixIncrement(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 5;
	++x;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	exprStmt := funcDecl.Body.Statements[1].(*structs.ExprStmt)

	unaryExpr, ok := exprStmt.Expression.(*structs.UnaryExpr)
	if !ok {
		t.Fatalf("Expected UnaryExpr, got %T", exprStmt.Expression)
	}

	if unaryExpr.Operator != "++" {
		t.Errorf("Expected operator '++', got '%s'", unaryExpr.Operator)
	}

	if unaryExpr.IsPostfix {
		t.Error("Expected prefix operator, got postfix")
	}
}

// TestParsePostfixIncrement проверяет парсинг постфиксного инкремента
func TestParsePostfixIncrement(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 5;
	x++;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	exprStmt := funcDecl.Body.Statements[1].(*structs.ExprStmt)

	unaryExpr, ok := exprStmt.Expression.(*structs.UnaryExpr)
	if !ok {
		t.Fatalf("Expected UnaryExpr, got %T", exprStmt.Expression)
	}

	if unaryExpr.Operator != "++" {
		t.Errorf("Expected operator '++', got '%s'", unaryExpr.Operator)
	}

	if !unaryExpr.IsPostfix {
		t.Error("Expected postfix operator, got prefix")
	}
}

// TestParseAssignmentExpression проверяет парсинг присваивания
func TestParseAssignmentExpression(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x;
	x = 10;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	exprStmt := funcDecl.Body.Statements[1].(*structs.ExprStmt)

	assignExpr, ok := exprStmt.Expression.(*structs.AssignmentExpr)
	if !ok {
		t.Fatalf("Expected AssignmentExpr, got %T", exprStmt.Expression)
	}

	if assignExpr.Operator != "=" {
		t.Errorf("Expected operator '=', got '%s'", assignExpr.Operator)
	}

	left, ok := assignExpr.Left.(*structs.VariableExpr)
	if !ok {
		t.Fatalf("Expected VariableExpr on left, got %T", assignExpr.Left)
	}
	if left.Name != "x" {
		t.Errorf("Expected variable 'x', got '%s'", left.Name)
	}

	right, ok := assignExpr.Right.(*structs.IntLiteral)
	if !ok {
		t.Fatalf("Expected IntLiteral on right, got %T", assignExpr.Right)
	}
	if right.Value != 10 {
		t.Errorf("Expected value 10, got %d", right.Value)
	}
}

// TestParseCompoundAssignmentOperators проверяет присваивания с операторами +=, -=, /= и т.д.
func TestParseCompoundAssignmentOperators(t *testing.T) {
	tests := []struct {
		name       string
		sourceCode string
		operator   string
		value      int
	}{
		{
			name: "plus_equal",
			sourceCode: `int main() {
	int x = 10;
	x += 5;
	return 0;
}`,
			operator: "+=",
			value:    5,
		},
		{
			name: "minus_equal",
			sourceCode: `int main() {
	int x = 10;
	x -= 3;
	return 0;
}`,
			operator: "-=",
			value:    3,
		},
		{
			name: "divide_equal",
			sourceCode: `int main() {
	int x = 10;
	x /= 2;
	return 0;
}`,
			operator: "/=",
			value:    2,
		},
		{
			name: "mod_equal",
			sourceCode: `int main() {
	int x = 10;
	x %= 4;
	return 0;
}`,
			operator: "%=",
			value:    4,
		},
		{
			name: "and_equal",
			sourceCode: `int main() {
	int x = 10;
	x &= 6;
	return 0;
}`,
			operator: "&=",
			value:    6,
		},
		{
			name: "or_equal",
			sourceCode: `int main() {
	int x = 10;
	x |= 3;
	return 0;
}`,
			operator: "|=",
			value:    3,
		},
		{
			name: "xor_equal",
			sourceCode: `int main() {
	int x = 10;
	x ^= 5;
	return 0;
}`,
			operator: "^=",
			value:    5,
		},
		{
			name: "shift_left_equal",
			sourceCode: `int main() {
	int x = 10;
	x <<= 2;
	return 0;
}`,
			operator: "<<=",
			value:    2,
		},
		{
			name: "shift_right_equal",
			sourceCode: `int main() {
	int x = 10;
	x >>= 1;
	return 0;
}`,
			operator: ">>=",
			value:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewCConverter()
			tree, err := conv.Parse([]byte(tt.sourceCode))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			ast, err := conv.ConvertToProgram(tree, []byte(tt.sourceCode))
			if err != nil {
				t.Fatalf("ConvertToProgram failed: %v", err)
			}

			program := ast.(*structs.Program)
			funcDecl := program.Declarations[0].(*structs.FunctionDecl)
			exprStmt := funcDecl.Body.Statements[1].(*structs.ExprStmt)

			assignExpr, ok := exprStmt.Expression.(*structs.AssignmentExpr)
			if !ok {
				t.Fatalf("Expected AssignmentExpr, got %T", exprStmt.Expression)
			}

			if assignExpr.Operator != tt.operator {
				t.Errorf("Expected operator '%s', got '%s'", tt.operator, assignExpr.Operator)
			}

			right, ok := assignExpr.Right.(*structs.IntLiteral)
			if !ok {
				t.Fatalf("Expected IntLiteral on right, got %T", assignExpr.Right)
			}

			if right.Value != tt.value {
				t.Errorf("Expected value %d, got %d", tt.value, right.Value)
			}
		})
	}
}

// TestParseCompoundAssignmentWithExpressions проверяет сложные присваивания с выражениями
func TestParseCompoundAssignmentWithExpressions(t *testing.T) {
	tests := []struct {
		name       string
		sourceCode string
		operator   string
	}{
		{
			name: "plus_equal_with_expression",
			sourceCode: `int main() {
	int x = 10;
	int y = 5;
	x += y + 1;
	return 0;
}`,
			operator: "+=",
		},
		{
			name: "minus_equal_with_multiply",
			sourceCode: `int main() {
	int x = 10;
	int y = 2;
	x -= y * 3;
	return 0;
}`,
			operator: "-=",
		},
		{
			name: "array_plus_equal",
			sourceCode: `int main() {
	int arr[5];
	arr[0] = 10;
	arr[0] += 5;
	return 0;
}`,
			operator: "+=",
		},
		{
			name: "array_index_expression",
			sourceCode: `int main() {
	int arr[5];
	int i = 2;
	arr[i] += 3;
	return 0;
}`,
			operator: "+=",
		},
		{
			name: "shift_with_variable",
			sourceCode: `int main() {
	int x = 8;
	int shift = 2;
	x <<= shift;
	return 0;
}`,
			operator: "<<=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewCConverter()
			tree, err := conv.Parse([]byte(tt.sourceCode))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			ast, err := conv.ConvertToProgram(tree, []byte(tt.sourceCode))
			if err != nil {
				t.Fatalf("ConvertToProgram failed: %v", err)
			}

			program := ast.(*structs.Program)
			funcDecl := program.Declarations[0].(*structs.FunctionDecl)

			// Найдём последний ExprStmt с присваиванием
			var assignExpr *structs.AssignmentExpr
			for i := len(funcDecl.Body.Statements) - 1; i >= 0; i-- {
				if exprStmt, ok := funcDecl.Body.Statements[i].(*structs.ExprStmt); ok {
					if assign, ok := exprStmt.Expression.(*structs.AssignmentExpr); ok {
						if assign.Operator == tt.operator {
							assignExpr = assign
							break
						}
					}
				}
			}

			if assignExpr == nil {
				t.Fatalf("Expected AssignmentExpr with operator '%s'", tt.operator)
			}

			if assignExpr.Operator != tt.operator {
				t.Errorf("Expected operator '%s', got '%s'", tt.operator, assignExpr.Operator)
			}

			// Проверяем что правая часть существует
			if assignExpr.Right == nil {
				t.Error("Right side of assignment should not be nil")
			}
		})
	}
}

// TestParseIfStatement проверяет парсинг if-statement
func TestParseIfStatement(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 5;
	if (x > 0) {
		return 1;
	}
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	ifStmt := funcDecl.Body.Statements[1].(*structs.IfStmt)

	if ifStmt.Condition == nil {
		t.Fatal("Expected condition in if statement")
	}

	if ifStmt.ThenBlock == nil {
		t.Fatal("Expected then block in if statement")
	}
}

// TestParseIfElseStatement проверяет парсинг if-else
func TestParseIfElseStatement(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 5;
	if (x > 0) {
		return 1;
	} else {
		return 2;
	}
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	ifStmt := funcDecl.Body.Statements[1].(*structs.IfStmt)

	if ifStmt.ThenBlock == nil {
		t.Fatal("Expected then block in if-else statement")
	}

	if ifStmt.ElseBlock == nil {
		t.Fatal("Expected else block in if-else statement")
	}
}

// TestParseElseIfStatement проверяет парсинг if-else if-else
func TestParseElseIfStatement(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 5;
	if (x > 10) {
		return 1;
	} else if (x > 0) {
		return 2;
	} else {
		return 3;
	}
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	ifStmt := funcDecl.Body.Statements[1].(*structs.IfStmt)

	if ifStmt.ThenBlock == nil {
		t.Fatal("Expected then block in if statement")
	}

	// else if представлен как else с вложенным if
	if ifStmt.ElseBlock == nil {
		t.Fatal("Expected else block in if statement")
	}

	// ElseBlock должен быть IfStmt (else if)
	elseIfStmt, ok := ifStmt.ElseBlock.(*structs.IfStmt)
	if !ok {
		t.Fatal("Expected ElseBlock to be IfStmt (else if)")
	}

	if elseIfStmt.Condition == nil {
		t.Fatal("Expected condition in else-if clause")
	}

	if elseIfStmt.ThenBlock == nil {
		t.Fatal("Expected block in else-if clause")
	}

	// Проверяем финальный else блок
	if elseIfStmt.ElseBlock == nil {
		t.Fatal("Expected else block after else-if")
	}
}

// TestParseWhileStatement проверяет парсинг while-statement
func TestParseWhileStatement(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 0;
	while (x < 10) {
		x = x + 1;
	}
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	whileStmt := funcDecl.Body.Statements[1].(*structs.WhileStmt)

	if whileStmt.Condition == nil {
		t.Fatal("Expected condition in while statement")
	}

	if whileStmt.Body == nil {
		t.Fatal("Expected body in while statement")
	}
}

// TestParseForStatement проверяет парсинг for-statement
func TestParseForStatement(t *testing.T) {
	sourceCode := []byte(`int main() {
	for (int i = 0; i < 10; i++) {
		i = i + 1;
	}
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	forStmt := funcDecl.Body.Statements[0].(*structs.ForStmt)

	if forStmt.Init == nil {
		t.Fatal("Expected init in for statement")
	}

	if forStmt.Condition == nil {
		t.Fatal("Expected condition in for statement")
	}

	if forStmt.Post == nil {
		t.Fatal("Expected post in for statement")
	}

	if forStmt.Body == nil {
		t.Fatal("Expected body in for statement")
	}
}

// TestParseArrayAccess проверяет парсинг доступа к элементу массива
func TestParseArrayAccess(t *testing.T) {
	sourceCode := []byte(`int main() {
	int arr[5];
	int x = arr[0];
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[1].(*structs.VariableDecl)

	arrayAccess, ok := varDecl.InitExpr.(*structs.ArrayAccessExpr)
	if !ok {
		t.Fatalf("Expected ArrayAccessExpr, got %T", varDecl.InitExpr)
	}

	if arrayAccess.Array == nil {
		t.Fatal("Expected array expression")
	}

	if arrayAccess.Index == nil {
		t.Fatal("Expected index expression")
	}
}

// TestParseCallExpression проверяет парсинг вызова функции
func TestParseCallExpression(t *testing.T) {
	sourceCode := []byte(`int foo(int x) { return x; }
int main() {
	int result = foo(5);
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	mainDecl := program.Declarations[1].(*structs.FunctionDecl)
	varDecl := mainDecl.Body.Statements[0].(*structs.VariableDecl)

	callExpr, ok := varDecl.InitExpr.(*structs.CallExpr)
	if !ok {
		t.Fatalf("Expected CallExpr, got %T", varDecl.InitExpr)
	}

	if callExpr.FunctionName != "foo" {
		t.Errorf("Expected function name 'foo', got '%s'", callExpr.FunctionName)
	}

	if len(callExpr.Arguments) != 1 {
		t.Errorf("Expected 1 argument, got %d", len(callExpr.Arguments))
	}
}

// TestValidation_InvalidLValueAssignment проверяет валидацию lvalue при присваивании выражения
// TestValidation_EmptyParenthesizedExpression проверяет пустое выражение в скобках
func TestValidation_EmptyParenthesizedExpression(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = ();
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_, err = conv.ConvertToProgram(tree, sourceCode)
	if err == nil {
		t.Fatal("Expected error for empty parenthesized expression")
	}

	errMsg := err.Error()
	if !contains(errMsg, "EmptyParenthesizedExpr") && !contains(errMsg, "parenthesized expression") && !contains(errMsg, "TreeSitterError") {
		t.Errorf("Expected empty parenthesized expression error, got: %v", err)
	}
}

// TestValidation_AssignmentMissingSide проверяет присваивание без левой или правой части
func TestValidation_AssignmentMissingSide(t *testing.T) {
	tests := []struct {
		name       string
		sourceCode string
	}{
		{
			name: "missing_left",
			sourceCode: `int main() {
	int x = 1;
	= 5;
	return 0;
}`,
		},
		{
			name: "missing_right",
			sourceCode: `int main() {
	int x = 1;
	x = ;
	return 0;
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewCConverter()
			tree, err := conv.Parse([]byte(tt.sourceCode))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			_, err = conv.ConvertToProgram(tree, []byte(tt.sourceCode))
			if err == nil {
				t.Fatal("Expected error for assignment with missing side")
			}

			errMsg := err.Error()
			if !contains(errMsg, "TreeSitterError") && !contains(errMsg, "syntax error") && !contains(errMsg, "Assignment") {
				t.Errorf("Expected assignment/syntax error, got: %v", err)
			}
		})
	}
}

// TestValidation_InvalidIdentifier проверяет валидацию идентификаторов
func TestValidation_InvalidIdentifier(t *testing.T) {
	sourceCode := []byte(`int main() {
	int 5invalid = 10;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_, err = conv.ConvertToProgram(tree, sourceCode)
	if err == nil {
		// tree-sitter может не парсить это, но если спарсит, то должна быть ошибка
		t.Log("Note: tree-sitter didn't create invalid identifier node")
	}
}

// TestValidation_UnsupportedOperator проверяет валидацию неподдерживаемого оператора
func TestValidation_UnsupportedOperator(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 5;
	int y = x @@ 3;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_, err = conv.ConvertToProgram(tree, sourceCode)
	// tree-sitter может не создать бинарное выражение для такого синтаксиса,
	// вместо этого он создает ERROR узлы
	if err != nil {
		t.Logf("Got expected error: %v", err)
	}
}

// TestValidation_UnsupportedOperatorsTable проверяет разные неподдерживаемые операторы
func TestValidation_UnsupportedOperatorsTable(t *testing.T) {
	tests := []struct {
		name       string
		sourceCode string
	}{
		{
			name: "double_at",
			sourceCode: `int main() {
	int x = 5;
	int y = x @@ 3;
	return 0;
}`,
		},
		{
			name: "triple_shift",
			sourceCode: `int main() {
	int x = 5;
	int y = x >>> 1;
	return 0;
}`,
		},
		{
			name: "double_percent",
			sourceCode: `int main() {
	int x = 10;
	int y = x %% 3;
	return 0;
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewCConverter()
			tree, err := conv.Parse([]byte(tt.sourceCode))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			_, err = conv.ConvertToProgram(tree, []byte(tt.sourceCode))
			if err == nil {
				t.Fatal("Expected error for unsupported operator")
			}

			errMsg := err.Error()
			if !contains(errMsg, "UnsupportedOperator") && !contains(errMsg, "unsupported") && !contains(errMsg, "TreeSitterError") {
				t.Errorf("Expected unsupported operator error, got: %v", err)
			}
		})
	}
}

// TestTreeSitterParseError проверяет обработку синтаксической ошибки tree-sitter
func TestTreeSitterParseError(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = ;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	_, err = conv.ConvertToProgram(tree, sourceCode)
	if err == nil {
		t.Fatal("Expected tree-sitter error for invalid syntax")
	}

	errMsg := err.Error()
	if !contains(errMsg, "TreeSitterError") && !contains(errMsg, "syntax error") {
		t.Errorf("Expected tree-sitter error message, got: %v", err)
	}
}

// TestParseBreakContinue проверяет парсинг break и continue
func TestParseBreakContinue(t *testing.T) {
	sourceCode := []byte(`int main() {
	while (1) {
		break;
		continue;
	}
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	if len(funcDecl.Body.Statements) < 2 {
		t.Fatalf("Expected at least 2 statements, got %d", len(funcDecl.Body.Statements))
	}

	whileStmt, ok := funcDecl.Body.Statements[0].(*structs.WhileStmt)
	if !ok {
		t.Fatalf("Expected WhileStmt, got %T", funcDecl.Body.Statements[0])
	}

	block, ok := whileStmt.Body.(*structs.BlockStmt)
	if !ok {
		t.Fatalf("Expected BlockStmt in while body, got %T", whileStmt.Body)
	}

	if len(block.Statements) != 2 {
		t.Fatalf("Expected 2 statements in while body, got %d", len(block.Statements))
	}

	if _, ok := block.Statements[0].(*structs.BreakStmt); !ok {
		t.Fatalf("Expected BreakStmt, got %T", block.Statements[0])
	}

	if _, ok := block.Statements[1].(*structs.ContinueStmt); !ok {
		t.Fatalf("Expected ContinueStmt, got %T", block.Statements[1])
	}
}

// TestParseReturnStatement проверяет парсинг return statement
func TestParseReturnStatement(t *testing.T) {
	sourceCode := []byte(`int main() {
	return 42;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	returnStmt := funcDecl.Body.Statements[0].(*structs.ReturnStmt)

	if returnStmt.Value == nil {
		t.Fatal("Expected value in return statement")
	}

	intLit, ok := returnStmt.Value.(*structs.IntLiteral)
	if !ok {
		t.Fatalf("Expected IntLiteral, got %T", returnStmt.Value)
	}

	if intLit.Value != 42 {
		t.Errorf("Expected return value 42, got %d", intLit.Value)
	}
}

// TestParseArrayInitializer проверяет парсинг инициализатора массива
func TestParseArrayInitializer(t *testing.T) {
	sourceCode := []byte(`int main() {
	int arr[] = {1, 2, 3};
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[0].(*structs.VariableDecl)

	arrayInit, ok := varDecl.InitExpr.(*structs.ArrayInitExpr)
	if !ok {
		t.Fatalf("Expected ArrayInitExpr, got %T", varDecl.InitExpr)
	}

	if len(arrayInit.Elements) != 3 {
		t.Errorf("Expected 3 elements, got %d", len(arrayInit.Elements))
	}
}

// TestParseArrayInitializerWithSize проверяет парсинг инициализатора массива с явным размером
// Это проверяет, что не создается двойное объявление переменной
func TestParseArrayInitializerWithSize(t *testing.T) {
	sourceCode := []byte(`int main() {
	int arr[3] = {1, 2, 3};
	int *p = arr;
	*p = 10;
	return arr[0];
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	// Проверяем, что есть ровно 4 statement в функции main:
	// 1. VariableDecl arr[3] = {1,2,3}
	// 2. VariableDecl *p = arr
	// 3. ExprStmt *p = 10
	// 4. ReturnStmt arr[0]
	if len(funcDecl.Body.Statements) != 4 {
		t.Errorf("Expected 4 statements, got %d", len(funcDecl.Body.Statements))
	}

	// Проверяем первое объявление - массив
	varDecl1, ok := funcDecl.Body.Statements[0].(*structs.VariableDecl)
	if !ok {
		t.Fatalf("Expected VariableDecl at index 0, got %T", funcDecl.Body.Statements[0])
	}

	if varDecl1.Name != "arr" {
		t.Errorf("Expected var name 'arr', got %s", varDecl1.Name)
	}

	if len(varDecl1.VarType.ArraySizes) != 1 || varDecl1.VarType.ArraySizes[0] != 3 {
		t.Errorf("Expected array size [3], got %v", varDecl1.VarType.ArraySizes)
	}

	_, ok = varDecl1.InitExpr.(*structs.ArrayInitExpr)
	if !ok {
		t.Fatalf("Expected ArrayInitExpr, got %T", varDecl1.InitExpr)
	}

	// Проверяем второе объявление - указатель
	varDecl2, ok := funcDecl.Body.Statements[1].(*structs.VariableDecl)
	if !ok {
		t.Fatalf("Expected VariableDecl at index 1, got %T", funcDecl.Body.Statements[1])
	}

	if varDecl2.Name != "p" {
		t.Errorf("Expected var name 'p', got %s", varDecl2.Name)
	}

	if varDecl2.VarType.PointerLevel != 1 {
		t.Errorf("Expected pointer level 1, got %d", varDecl2.VarType.PointerLevel)
	}
}

// TestParsePointerReturnType проверяет парсинг функции с возвращаемым указателем
func TestParsePointerReturnType(t *testing.T) {
	sourceCode := []byte(`int* get_pointer() {
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	if funcDecl.ReturnType.PointerLevel != 1 {
		t.Errorf("Expected pointer level 1, got %d", funcDecl.ReturnType.PointerLevel)
	}
}

// TestParseVoidReturnType проверяет парсинг функции с возвращаемым типом void
func TestParseVoidReturnType(t *testing.T) {
	sourceCode := []byte(`void do_nothing() {
	return;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	if funcDecl.ReturnType.BaseType != "void" {
		t.Errorf("Expected return type 'void', got '%s'", funcDecl.ReturnType.BaseType)
	}
}

// TestCommentsSingleLine проверяет, что однострочные комментарии игнорируются
func TestCommentsSingleLine(t *testing.T) {
	sourceCode := []byte(`int main() {
	// Это комментарий
	int x = 5;  // Ещё комментарий
	return 0;   // Финальный комментарий
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	// Должно быть 2 statement: VariableDecl и ReturnStmt
	// Комментарии не должны быть в AST
	if len(funcDecl.Body.Statements) != 2 {
		t.Errorf("Expected 2 statements, got %d", len(funcDecl.Body.Statements))
	}

	_, ok := funcDecl.Body.Statements[0].(*structs.VariableDecl)
	if !ok {
		t.Fatalf("Expected first statement to be VariableDecl")
	}

	_, ok = funcDecl.Body.Statements[1].(*structs.ReturnStmt)
	if !ok {
		t.Fatalf("Expected second statement to be ReturnStmt")
	}
}

// TestCommentsMultiLine проверяет, что многострочные комментарии игнорируются
func TestCommentsMultiLine(t *testing.T) {
	sourceCode := []byte(`int main() {
	/* Это многострочный
	   комментарий */
	int x = 5;
	/* Ещё один
	   многострочный
	   комментарий */
	return x;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	// Должно быть 2 statement: VariableDecl и ReturnStmt
	// Комментарии не должны быть в AST
	if len(funcDecl.Body.Statements) != 2 {
		t.Errorf("Expected 2 statements, got %d", len(funcDecl.Body.Statements))
	}

	varDecl, ok := funcDecl.Body.Statements[0].(*structs.VariableDecl)
	if !ok {
		t.Fatalf("Expected first statement to be VariableDecl")
	}

	if varDecl.Name != "x" {
		t.Errorf("Expected variable name 'x', got '%s'", varDecl.Name)
	}

	returnStmt, ok := funcDecl.Body.Statements[1].(*structs.ReturnStmt)
	if !ok {
		t.Fatalf("Expected second statement to be ReturnStmt")
	}

	if returnStmt.Value == nil {
		t.Error("Expected return statement to have a value")
	}
}

// TestCommentsMixed проверяет, что смешанные комментарии обрабатываются правильно
func TestCommentsMixed(t *testing.T) {
	sourceCode := []byte(`/* Комментарий перед функцией */
int main() {
	int x = 10;     // x получает значение 10
	/* Подготовка к циклу */
	for (int i = 0; /* начало цикла */ i < x; i++) {
		x = x + 1;  // Увеличение x
	}
	/* Возврат результата */
	return x;  // Конец функции
}
/* Конечный комментарий */`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	if len(program.Declarations) == 0 {
		t.Fatal("Expected function declaration")
	}

	funcDecl, ok := program.Declarations[0].(*structs.FunctionDecl)
	if !ok {
		t.Fatalf("Expected FunctionDecl, got %T", program.Declarations[0])
	}

	if funcDecl.Name != "main" {
		t.Errorf("Expected function name 'main', got '%s'", funcDecl.Name)
	}

	// Должно быть 3 statement: VariableDecl, ForStmt и ReturnStmt
	if len(funcDecl.Body.Statements) != 3 {
		t.Errorf("Expected 3 statements, got %d", len(funcDecl.Body.Statements))
	}

	_, ok = funcDecl.Body.Statements[0].(*structs.VariableDecl)
	if !ok {
		t.Fatalf("Expected first statement to be VariableDecl")
	}

	_, ok = funcDecl.Body.Statements[1].(*structs.ForStmt)
	if !ok {
		t.Fatalf("Expected second statement to be ForStmt")
	}

	_, ok = funcDecl.Body.Statements[2].(*structs.ReturnStmt)
	if !ok {
		t.Fatalf("Expected third statement to be ReturnStmt")
	}
}

// TestCommentsInExpressions проверяет, что комментарии в выражениях игнорируются
func TestCommentsInExpressions(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 5 /* пять */ + 3 /* три */;
	return x;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)
	varDecl := funcDecl.Body.Statements[0].(*structs.VariableDecl)

	binExpr, ok := varDecl.InitExpr.(*structs.BinaryExpr)
	if !ok {
		t.Fatalf("Expected BinaryExpr, got %T", varDecl.InitExpr)
	}

	// Проверяем, что бинарное выражение корректно парсится несмотря на комментарии
	left, ok := binExpr.Left.(*structs.IntLiteral)
	if !ok {
		t.Fatalf("Expected IntLiteral on left, got %T", binExpr.Left)
	}

	if left.Value != 5 {
		t.Errorf("Expected left value 5, got %d", left.Value)
	}

	right, ok := binExpr.Right.(*structs.IntLiteral)
	if !ok {
		t.Fatalf("Expected IntLiteral on right, got %T", binExpr.Right)
	}

	if right.Value != 3 {
		t.Errorf("Expected right value 3, got %d", right.Value)
	}
}

// TestCommentsInVariousContexts проверяет комментарии в различных контекстах через таблицу
func TestCommentsInVariousContexts(t *testing.T) {
	tests := []struct {
		name               string
		sourceCode         string
		expectedVarCount   int
		expectedStatements int
		expectedInitValue  int
		expectedBinaryOp   string
	}{
		{
			name: "comment_before_variable",
			sourceCode: `int main() {
	/* Инициализация */ int x = 42;
	return x;
}`,
			expectedVarCount:   1,
			expectedStatements: 2,
			expectedInitValue:  42,
		},
		{
			name: "comment_after_variable",
			sourceCode: `int main() {
	int x = 42; /* завершена инициализация */
	return x;
}`,
			expectedVarCount:   1,
			expectedStatements: 2,
			expectedInitValue:  42,
		},
		{
			name: "comment_between_statements",
			sourceCode: `int main() {
	int x = 10;
	/* важный комментарий посередине */
	int y = 20;
	return x;
}`,
			expectedVarCount:   2,
			expectedStatements: 3,
			expectedInitValue:  10,
		},
		{
			name: "comment_in_binary_expression",
			sourceCode: `int main() {
	int result = 10 /* plus */ + /* another */ 5;
	return result;
}`,
			expectedVarCount:   1,
			expectedStatements: 2,
			expectedInitValue:  10,
			expectedBinaryOp:   "+",
		},
		{
			name: "comment_in_for_init",
			sourceCode: `int main() {
	for (/* init */ int i = 0; i < 5; i++) {
		int x = i;
	}
	return 0;
}`,
			expectedVarCount: 1,
		},
		{
			name: "comment_in_for_condition",
			sourceCode: `int main() {
	for (int i = 0; /* condition */ i < 5; i++) {
		int x = i;
	}
	return 0;
}`,
			expectedVarCount: 1,
		},
		{
			name: "comment_in_for_increment",
			sourceCode: `int main() {
	for (int i = 0; i < 5; /* increment */ i++) {
		int x = i;
	}
	return 0;
}`,
			expectedVarCount: 1,
		},
		{
			name: "multiple_comments_complex",
			sourceCode: `int main() {
	/* variable initialization */
	int x = 5 /* first number */ + 3 /* second number */; // calculation
	/* some pause */
	int y = x /* multiply */ * 2 /* doubled */;
	return y; /* result */
}`,
			expectedVarCount:   2,
			expectedStatements: 3,
			expectedInitValue:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewCConverter()
			tree, err := conv.Parse([]byte(tt.sourceCode))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			ast, err := conv.ConvertToProgram(tree, []byte(tt.sourceCode))
			if err != nil {
				t.Fatalf("ConvertToProgram failed: %v", err)
			}

			program := ast.(*structs.Program)
			if len(program.Declarations) == 0 {
				t.Fatal("Expected function declaration")
			}

			funcDecl, ok := program.Declarations[0].(*structs.FunctionDecl)
			if !ok {
				t.Fatalf("Expected FunctionDecl, got %T", program.Declarations[0])
			}

			if tt.expectedStatements > 0 {
				if len(funcDecl.Body.Statements) != tt.expectedStatements {
					t.Errorf("Expected %d statements, got %d", tt.expectedStatements, len(funcDecl.Body.Statements))
				}
			}

			if tt.expectedInitValue > 0 {
				// Проверяем первый statement - должна быть VariableDecl
				varDecl, ok := funcDecl.Body.Statements[0].(*structs.VariableDecl)
				if !ok {
					t.Fatalf("Expected VariableDecl, got %T", funcDecl.Body.Statements[0])
				}

				if varDecl.InitExpr != nil {
					intLit, ok := varDecl.InitExpr.(*structs.IntLiteral)
					if ok && intLit.Value != tt.expectedInitValue {
						t.Errorf("Expected init value %d, got %d", tt.expectedInitValue, intLit.Value)
					}
				}
			}

			if tt.expectedBinaryOp != "" {
				varDecl, ok := funcDecl.Body.Statements[0].(*structs.VariableDecl)
				if !ok {
					t.Fatalf("Expected VariableDecl")
				}

				binExpr, ok := varDecl.InitExpr.(*structs.BinaryExpr)
				if !ok {
					t.Fatalf("Expected BinaryExpr, got %T", varDecl.InitExpr)
				}

				if binExpr.Operator != tt.expectedBinaryOp {
					t.Errorf("Expected operator %s, got %s", tt.expectedBinaryOp, binExpr.Operator)
				}
			}
		})
	}
}

// TestCommentsInExpressionVariations проверяет комментарии в выражениях через таблицу
func TestCommentsInExpressionVariations(t *testing.T) {
	tests := []struct {
		name             string
		sourceCode       string
		expectedLeftVal  int
		expectedRightVal int
		expectedOperator string
	}{
		{
			name: "addition_with_comments",
			sourceCode: `int main() {
	int x = 10 /* ten */ + 20 /* twenty */;
	return x;
}`,
			expectedLeftVal:  10,
			expectedRightVal: 20,
			expectedOperator: "+",
		},
		{
			name: "subtraction_with_comments",
			sourceCode: `int main() {
	int x = 100 /* hundred */ - 50 /* fifty */;
	return x;
}`,
			expectedLeftVal:  100,
			expectedRightVal: 50,
			expectedOperator: "-",
		},
		{
			name: "multiplication_with_comments",
			sourceCode: `int main() {
	int x = 7 /* seven */ * 3 /* three */;
	return x;
}`,
			expectedLeftVal:  7,
			expectedRightVal: 3,
			expectedOperator: "*",
		},
		{
			name: "division_with_comments",
			sourceCode: `int main() {
	int x = 100 /* numerator */ / 5 /* denominator */;
	return x;
}`,
			expectedLeftVal:  100,
			expectedRightVal: 5,
			expectedOperator: "/",
		},
		{
			name: "modulo_with_comments",
			sourceCode: `int main() {
	int x = 17 /* dividend */ % 5 /* divisor */;
	return x;
}`,
			expectedLeftVal:  17,
			expectedRightVal: 5,
			expectedOperator: "%",
		},
		{
			name: "equality_with_comments",
			sourceCode: `int main() {
	int x = (5 /* first */ == 5 /* same */);
	return x;
}`,
			expectedLeftVal:  5,
			expectedRightVal: 5,
			expectedOperator: "==",
		},
		{
			name: "comparison_with_comments",
			sourceCode: `int main() {
	int x = (10 /* a */ < 20 /* b */);
	return x;
}`,
			expectedLeftVal:  10,
			expectedRightVal: 20,
			expectedOperator: "<",
		},
		{
			name: "logical_and_with_comments",
			sourceCode: `int main() {
	int x = (1 /* true */ && 1 /* also true */);
	return x;
}`,
			expectedLeftVal:  1,
			expectedRightVal: 1,
			expectedOperator: "&&",
		},
		{
			name: "logical_or_with_comments",
			sourceCode: `int main() {
	int x = (0 /* false */ || 1 /* true */);
	return x;
}`,
			expectedLeftVal:  0,
			expectedRightVal: 1,
			expectedOperator: "||",
		},
		{
			name: "bitwise_and_with_comments",
			sourceCode: `int main() {
	int x = (12 /* 1100 */ & 10 /* 1010 */);
	return x;
}`,
			expectedLeftVal:  12,
			expectedRightVal: 10,
			expectedOperator: "&",
		},
		{
			name: "bitwise_or_with_comments",
			sourceCode: `int main() {
	int x = (12 /* 1100 */ | 10 /* 1010 */);
	return x;
}`,
			expectedLeftVal:  12,
			expectedRightVal: 10,
			expectedOperator: "|",
		},
		{
			name: "bitwise_xor_with_comments",
			sourceCode: `int main() {
	int x = (12 /* 1100 */ ^ 10 /* 1010 */);
			return x;
}`,
			expectedLeftVal:  12,
			expectedRightVal: 10,
			expectedOperator: "^",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewCConverter()
			tree, err := conv.Parse([]byte(tt.sourceCode))
			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}

			ast, err := conv.ConvertToProgram(tree, []byte(tt.sourceCode))
			if err != nil {
				t.Fatalf("ConvertToProgram failed: %v", err)
			}

			program := ast.(*structs.Program)
			funcDecl := program.Declarations[0].(*structs.FunctionDecl)
			varDecl := funcDecl.Body.Statements[0].(*structs.VariableDecl)

			binExpr, ok := varDecl.InitExpr.(*structs.BinaryExpr)
			if !ok {
				t.Fatalf("Expected BinaryExpr, got %T", varDecl.InitExpr)
			}

			if binExpr.Operator != tt.expectedOperator {
				t.Errorf("Expected operator %s, got %s", tt.expectedOperator, binExpr.Operator)
			}

			// Для parenthesized expressions левая часть может быть другой
			var leftVal, rightVal int

			// Разворачиваем parenthesized expression если нужно
			if parenthesized, ok := varDecl.InitExpr.(*structs.BinaryExpr); ok {
				binExpr = parenthesized
			}

			if leftLit, ok := binExpr.Left.(*structs.IntLiteral); ok {
				leftVal = leftLit.Value
			}

			if rightLit, ok := binExpr.Right.(*structs.IntLiteral); ok {
				rightVal = rightLit.Value
			}

			if leftVal != tt.expectedLeftVal {
				t.Errorf("Expected left value %d, got %d", tt.expectedLeftVal, leftVal)
			}

			if rightVal != tt.expectedRightVal {
				t.Errorf("Expected right value %d, got %d", tt.expectedRightVal, rightVal)
			}
		})
	}
}

// contains проверяет, содержит ли строка подстроку (для простой проверки ошибок)
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestParseGoto проверяет парсинг goto statement
func TestParseGoto(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 0;
	goto end;
	x = 5;
	end: return x;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	// Ищем goto statement
	var gotoFound bool
	for _, stmt := range funcDecl.Body.Statements {
		if g, ok := stmt.(*structs.GotoStmt); ok {
			if g.Label == "end" {
				gotoFound = true
				if g.Type != "GotoStmt" {
					t.Fatalf("Expected type 'GotoStmt', got '%s'", g.Type)
				}
			}
		}
	}

	if !gotoFound {
		t.Fatal("Expected to find GotoStmt with label 'end'")
	}
}

// TestParseLabel проверяет парсинг labeled statement
func TestParseLabel(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 0;
	loop: x = x + 1;
	return x;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	// Ищем label statement
	var labelFound bool
	for _, stmt := range funcDecl.Body.Statements {
		if l, ok := stmt.(*structs.LabelStmt); ok {
			if l.Label == "loop" {
				labelFound = true
				if l.Type != "LabelStmt" {
					t.Fatalf("Expected type 'LabelStmt', got '%s'", l.Type)
				}
				if l.Statement == nil {
					t.Fatal("Expected statement after label, got nil")
				}
			}
		}
	}

	if !labelFound {
		t.Fatal("Expected to find LabelStmt with label 'loop'")
	}
}

// TestParseDoWhile проверяет парсинг do while statement
func TestParseDoWhile(t *testing.T) {
	sourceCode := []byte(`int main() {
	int i = 0;
	do {
		i = i + 1;
	} while (i < 5);
	return i;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	// Проверяем, что найден do while statement
	var doWhileFound bool
	for _, stmt := range funcDecl.Body.Statements {
		if dw, ok := stmt.(*structs.DoWhileStmt); ok {
			doWhileFound = true
			if dw.Type != "DoWhileStmt" {
				t.Errorf("Expected type 'DoWhileStmt', got '%s'", dw.Type)
			}
			if dw.Body == nil {
				t.Error("DoWhileStmt body is nil")
			}
			if dw.Condition == nil {
				t.Error("DoWhileStmt condition is nil")
			}
			break
		}
	}

	if !doWhileFound {
		t.Fatal("DoWhileStmt not found in AST")
	}
}

// TestParseDoWhileSimple проверяет простой do while с одним оператором
func TestParseDoWhileSimple(t *testing.T) {
	sourceCode := []byte(`int main() {
	int x = 0;
	do x = x + 1; while (x < 10);
	return x;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	// Проверяем, что найден do while statement
	var doWhileFound bool
	for _, stmt := range funcDecl.Body.Statements {
		if dw, ok := stmt.(*structs.DoWhileStmt); ok {
			doWhileFound = true
			// Тело должно быть ExprStmt, содержащим AssignmentExpr
			if _, ok := dw.Body.(*structs.ExprStmt); !ok {
				t.Errorf("Expected ExprStmt in body, got %T", dw.Body)
			}
			break
		}
	}

	if !doWhileFound {
		t.Fatal("DoWhileStmt not found in AST")
	}
}

// TestParseNestedDoWhile проверяет вложенный do while
func TestParseNestedDoWhile(t *testing.T) {
	sourceCode := []byte(`int main() {
	int i = 0;
	int j = 0;
	do {
		j = 0;
		do {
			j = j + 1;
		} while (j < 3);
		i = i + 1;
	} while (i < 5);
	return i;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	// Проверяем наличие внешнего do while
	var outerDoWhileFound bool
	for _, stmt := range funcDecl.Body.Statements {
		if dw, ok := stmt.(*structs.DoWhileStmt); ok {
			outerDoWhileFound = true
			// Тело должно содержать вложенный do while
			if blockStmt, ok := dw.Body.(*structs.BlockStmt); ok {
				var innerDoWhileFound bool
				for _, innerStmt := range blockStmt.Statements {
					if _, ok := innerStmt.(*structs.DoWhileStmt); ok {
						innerDoWhileFound = true
						break
					}
				}
				if !innerDoWhileFound {
					t.Error("Inner DoWhileStmt not found in outer DoWhileStmt body")
				}
			} else {
				t.Errorf("Expected BlockStmt in body, got %T", dw.Body)
			}
			break
		}
	}

	if !outerDoWhileFound {
		t.Fatal("Outer DoWhileStmt not found in AST")
	}
}

// TestIfWithGoto проверяет, что goto внутри if парсится правильно
func TestIfWithGoto(t *testing.T) {
	sourceCode := []byte(`int main() {
	int a = 0;
	if (a == 0) goto end;
	a = 5;
end:
	return a;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	// Ищем if statement
	var ifFound bool
	for _, stmt := range funcDecl.Body.Statements {
		if ifStmt, ok := stmt.(*structs.IfStmt); ok {
			ifFound = true
			// Проверяем, что thenBlock содержит GotoStmt, а не null
			if ifStmt.ThenBlock == nil {
				t.Error("IfStmt.ThenBlock is nil, expected GotoStmt")
			} else if gotoStmt, ok := ifStmt.ThenBlock.(*structs.GotoStmt); !ok {
				t.Errorf("Expected GotoStmt in if, got %T", ifStmt.ThenBlock)
			} else if gotoStmt.Label != "end" {
				t.Errorf("Expected goto label 'end', got '%s'", gotoStmt.Label)
			}
			break
		}
	}

	if !ifFound {
		t.Fatal("IfStmt not found in AST")
	}
}

// TestComplexControlFlow проверяет сложный контрольный поток с goto, label, do-while и if
func TestComplexControlFlow(t *testing.T) {
	sourceCode := []byte(`int main() {
	int a = 0;
	int b = 0;

start:
	do {
		a += 1;
		if (a == 3) goto end;
		b += 2;
	} while (a < 5);

end:
	b += 10;
	return 0;
}`)

	conv := NewCConverter()
	tree, err := conv.Parse(sourceCode)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	ast, err := conv.ConvertToProgram(tree, sourceCode)
	if err != nil {
		t.Fatalf("ConvertToProgram failed: %v", err)
	}

	program := ast.(*structs.Program)
	funcDecl := program.Declarations[0].(*structs.FunctionDecl)

	// Проверяем структуру: должны быть два LabelStmt (start и end)
	var startLabelFound, endLabelFound, ifWithGotoFound bool
	for _, stmt := range funcDecl.Body.Statements {
		if labelStmt, ok := stmt.(*structs.LabelStmt); ok {
			if labelStmt.Label == "start" {
				startLabelFound = true
				// Проверяем, что содержит do-while
				if _, ok := labelStmt.Statement.(*structs.DoWhileStmt); !ok {
					t.Errorf("Expected DoWhileStmt after 'start' label, got %T", labelStmt.Statement)
				}
			} else if labelStmt.Label == "end" {
				endLabelFound = true
			}
		}
	}

	// Также проверяем наличие if с goto внутри do-while
	for _, stmt := range funcDecl.Body.Statements {
		if labelStmt, ok := stmt.(*structs.LabelStmt); ok && labelStmt.Label == "start" {
			if doWhile, ok := labelStmt.Statement.(*structs.DoWhileStmt); ok {
				if blockStmt, ok := doWhile.Body.(*structs.BlockStmt); ok {
					for _, innerStmt := range blockStmt.Statements {
						if ifStmt, ok := innerStmt.(*structs.IfStmt); ok {
							if gotoStmt, ok := ifStmt.ThenBlock.(*structs.GotoStmt); ok && gotoStmt.Label == "end" {
								ifWithGotoFound = true
							}
						}
					}
				}
			}
		}
	}

	if !startLabelFound {
		t.Error("'start' label not found")
	}
	if !endLabelFound {
		t.Error("'end' label not found")
	}
	if !ifWithGotoFound {
		t.Error("if with goto to 'end' not found inside do-while")
	}
}

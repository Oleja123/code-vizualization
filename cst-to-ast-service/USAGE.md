# Краткая справка по использованию CST-to-AST конвертера

## Быстрый старт

### 1. Импорт пакета
```go
import "github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/converter"
```

### 2. Создание конвертера
```go
conv := converter.NewCConverter()
```

### 3. Парсинг исходного кода
```go
sourceCode := []byte(`
int main() {
    int x = 42;
    return x;
}
`)

tree, err := conv.Parse(sourceCode)
if err != nil {
    log.Fatal(err)
}
```

### 4. Конвертация в AST
```go
ast, err := conv.ConvertToProgram(tree, sourceCode)
if err != nil {
    log.Fatal(err)
}
```

### 5. Использование AST
```go
program := ast.(*structs.Program)

for _, decl := range program.Declarations {
    switch node := decl.(type) {
    case *structs.FunctionDecl:
        fmt.Printf("Function: %s\n", node.Name)
    case *structs.VariableDecl:
        fmt.Printf("Variable: %s\n", node.Name)
    }
}
```

### 6. Сериализация в JSON
```go
import "encoding/json"

jsonData, err := json.MarshalIndent(ast, "", "  ")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(jsonData))
```

## Запуск примеров

```bash
# Простой пример (факториал, рекурсия)
go run cmd/example/main.go

# Продвинутый пример (массивы, указатели, циклы)
go run cmd/advanced-example/main.go
```

## Основные типы узлов

### Statements
```go
// Объявление переменной
*structs.VariableDecl {
    Type:     structs.Type{BaseType: "int", PointerLevel: 0, ArraySize: 0},
    Name:     "x",
    InitExpr: &structs.IntLiteral{Value: 42},
}

// Функция
*structs.FunctionDecl {
    Name:       "main",
    Parameters: []structs.Parameter{},
    Body:       &structs.BlockStmt{...},
}

// if/else
*structs.IfStmt {
    Condition:  interfaces.Expr,
    ThenBlock:  interfaces.Stmt,
    ElseBlock:  interfaces.Stmt,
}

// while
*structs.WhileStmt {
    Condition: interfaces.Expr,
    Body:      interfaces.Stmt,
}

// for
*structs.ForStmt {
    Init:      interfaces.Stmt,
    Condition: interfaces.Expr,
    Post:      interfaces.Expr,
    Body:      interfaces.Stmt,
}

// return
*structs.ReturnStmt {
    Value: interfaces.Expr,
}

// Блок {...}
*structs.BlockStmt {
    Statements: []interfaces.Stmt,
}
```

### Expressions
```go
// Идентификатор
*structs.Identifier {Name: "x"}

// Число
*structs.IntLiteral {Value: 42}

// Бинарная операция
*structs.BinaryExpr {
    Left:     interfaces.Expr,
    Operator: "+",
    Right:    interfaces.Expr,
}

// Унарная операция
*structs.UnaryExpr {
    Operator: "-",
    Operand:  interfaces.Expr,
}

// Присваивание
*structs.AssignmentExpr {
    Left:  interfaces.Expr,
    Right: interfaces.Expr,
}

// Вызов функции
*structs.CallExpr {
    FunctionName: "printf",
    Arguments:    []interfaces.Expr{},
}

// Доступ к элементу массива
*structs.ArrayAccessExpr {
    Array: interfaces.Expr,
    Index: interfaces.Expr,
}
```

## Примеры конкретного использования

### Работа с функциями
```go
ast := /* ... полученный AST ... */
program := ast.(*structs.Program)

for _, decl := range program.Declarations {
    if fn, ok := decl.(*structs.FunctionDecl); ok {
        fmt.Printf("Function: %s\n", fn.Name)
        fmt.Printf("Parameters: %d\n", len(fn.Parameters))
        for _, param := range fn.Parameters {
            fmt.Printf("  - %s (int)\n", param.Name)
        }
    }
}
```

### Обход всех переменных
```go
var visitVars func(interfaces.Node)
visitVars = func(node interfaces.Node) {
    switch n := node.(type) {
    case *structs.VariableDecl:
        fmt.Printf("Variable: %s\n", n.Name)
    case *structs.BlockStmt:
        for _, stmt := range n.Statements {
            visitVars(stmt)
        }
    }
}

program := ast.(*structs.Program)
for _, decl := range program.Declarations {
    visitVars(decl)
}
```

### Поиск функций с конкретным именем
```go
func findFunction(program *structs.Program, name string) *structs.FunctionDecl {
    for _, decl := range program.Declarations {
        if fn, ok := decl.(*structs.FunctionDecl); ok && fn.Name == name {
            return fn
        }
    }
    return nil
}

main := findFunction(program, "main")
if main != nil {
    fmt.Printf("Found main function with %d statements\n", len(main.Body.Statements))
}
```

## Информация о позициях в коде

Каждый узел содержит информацию о позиции в исходном коде:

```go
node := &structs.IntLiteral{
    Value: 42,
    Loc: structs.Location{
        Line:      5,
        Column:    12,
        EndLine:   5,
        EndColumn: 14,
    },
}

fmt.Printf("Узел на строке %d, колонка %d\n", node.Loc.Line, node.Loc.Column)
```

## Типы данных

### Система типов
```go
type Type struct {
    BaseType     string // только "int"
    PointerLevel int    // 0 = int, 1 = int*, 2 = int**, и т.д.
    ArraySize    int    // 0 если не массив, N для int[N]
}

// Примеры:
// int x -> Type{BaseType: "int", PointerLevel: 0, ArraySize: 0}
// int *p -> Type{BaseType: "int", PointerLevel: 1, ArraySize: 0}
// int arr[10] -> Type{BaseType: "int", PointerLevel: 0, ArraySize: 10}
// int **pp -> Type{BaseType: "int", PointerLevel: 2, ArraySize: 0}
```

## Обработка ошибок

```go
tree, err := conv.Parse(sourceCode)
if err != nil {
    fmt.Printf("Parse error: %v\n", err)
    return
}

ast, err := conv.ConvertToProgram(tree, sourceCode)
if err != nil {
    fmt.Printf("Conversion error: %v\n", err)
    return
}
```

Ошибки содержат информацию о строке, на которой произошла проблема.

## Производительность

- Для небольших программ (< 1000 строк): < 10ms
- Для средних программ (1000-10000 строк): 10-100ms
- Для больших программ (> 10000 строк): > 100ms

Tree-sitter использует инкрементальное парсинга, что позволяет эффективно обновлять AST.

## Ограничения текущей версии

- Поддерживается только тип `int`
- Нет структур и объединений
- Нет typedef и enum
- Нет препроцессора
- Нет обработки комментариев
- Одномерные массивы только

## Дальнейшее развитие

Проект готов к:
- Созданию интерпретатора для C
- Визуализации структуры программы
- Статического анализа кода
- Оптимизаций программ
- Трансформации AST

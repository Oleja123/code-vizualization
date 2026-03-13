# Архитектура `cst-to-ast-service`

## Назначение

`cst-to-ast-service` парсит C-код через tree-sitter и конвертирует CST в AST, который используется другими сервисами проекта (интерпретатор, семантический анализатор, визуализация).

Сервис **не выполняет** семантический анализ (типы, символы, правила языка) и **не исполняет** программу.

---

## Структура пакетов

- `pkg/converter` — публичный API и реализация конвертера.
- `internal/domain/interfaces` — базовые интерфейсы `Stmt`, `Expr`, `Location`.
- `internal/domain/structs` — AST-структуры.

---

## Публичный API

```go
import "github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"

conv := converter.New()
program, err := conv.ParseToAST(sourceCode)
if err != nil {
    // err имеет тип *converter.ConverterError
    _ = err.GetCode()
    _ = err.GetMessage()
    _ = err.GetLocation()
    _ = err.GetNodeType()
}
```

### Основные методы `CConverter`

- `New()` / `NewCConverter()` — создать конвертер.
- `ParseToAST(sourceCode string) (*Program, *ConverterError)` — основной публичный путь.
- `Parse(sourceCode []byte) (*sitter.Tree, error)` — получить CST.
- `ConvertToProgram(tree *sitter.Tree, sourceCode []byte) (interfaces.Node, error)` — CST → AST.
- `ConvertStmt(...)`, `ConvertExpr(...)` — конвертация отдельных узлов.

---

## Ошибки конвертации

`ParseToAST` возвращает `*ConverterError` с полями:

- `Code ErrorCode`
- `Message string`
- `NodeType string`
- `Loc Location`
- `Cause error` (внутреннее)

Коды ошибок (`ErrorCode`):

- `ParseFailed`
- `StmtUnsupported`
- `ExprUnsupported`
- `TreeSitterError`
- `IntLiteralParse`
- `StmtConversion`

---

## AST: актуальные структуры

Ниже — фактический контракт из `internal/domain/structs/ast.go`.

### Базовые типы

```go
type Location struct {
    Line      uint32 `json:"line"`
    Column    uint32 `json:"column"`
    EndLine   uint32 `json:"endLine"`
    EndColumn uint32 `json:"endColumn"`
}

type Type struct {
    BaseType     string `json:"baseType"`
    PointerLevel int    `json:"pointerLevel"`
    ArraySizes   []int  `json:"arraySizes"`
}

type Parameter struct {
    Type Type     `json:"type"`
    Name string   `json:"name"`
    Loc  Location `json:"location"`
}
```

### Корень AST

```go
type Program struct {
    Type         string   `json:"type"`        // "Program"
    Declarations []Stmt   `json:"declarations"`
    Loc          Location `json:"location"`
}
```

### Statements

```go
type VariableDecl struct {
    Type     string   `json:"type"`    // "VariableDecl"
    VarType  Type     `json:"varType"`
    Name     string   `json:"name"`
    InitExpr Expr     `json:"initExpr,omitempty"`
    Loc      Location `json:"location"`
}

type FunctionDecl struct {
    Type       string      `json:"type"` // "FunctionDecl"
    Name       string      `json:"name"`
    ReturnType Type        `json:"returnType"`
    Parameters []Parameter `json:"parameters"`
    Body       *BlockStmt  `json:"body"`
    Loc        Location    `json:"location"`
}

type IfStmt struct {
    Type      string   `json:"type"` // "IfStmt"
    Condition Expr     `json:"condition"`
    ThenBlock Stmt     `json:"thenBlock"`
    ElseBlock Stmt     `json:"elseBlock,omitempty"`
    Loc       Location `json:"location"`
}

type WhileStmt struct {
    Type      string   `json:"type"` // "WhileStmt"
    Condition Expr     `json:"condition"`
    Body      Stmt     `json:"body"`
    Loc       Location `json:"location"`
}

type DoWhileStmt struct {
    Type      string   `json:"type"` // "DoWhileStmt"
    Body      Stmt     `json:"body"`
    Condition Expr     `json:"condition"`
    Loc       Location `json:"location"`
}

type ForStmt struct {
    Type      string   `json:"type"` // "ForStmt"
    Init      Stmt     `json:"init,omitempty"`
    Condition Expr     `json:"condition,omitempty"`
    Post      Stmt     `json:"post,omitempty"`
    Body      Stmt     `json:"body"`
    Loc       Location `json:"location"`
}

type ReturnStmt struct {
    Type  string   `json:"type"` // "ReturnStmt"
    Value Expr     `json:"value,omitempty"`
    Loc   Location `json:"location"`
}

type BlockStmt struct {
    Type       string `json:"type"` // "BlockStmt"
    Statements []Stmt `json:"statements"`
    Loc        Location `json:"location"`
}

type ExprStmt struct {
    Type       string `json:"type"` // "ExprStmt"
    Expression Expr   `json:"expression"`
    Loc        Location `json:"location"`
}

type BreakStmt struct {
    Type string   `json:"type"` // "BreakStmt"
    Loc  Location `json:"location"`
}

type ContinueStmt struct {
    Type string   `json:"type"` // "ContinueStmt"
    Loc  Location `json:"location"`
}

type GotoStmt struct {
    Type  string   `json:"type"` // "GotoStmt"
    Label string   `json:"label"`
    Loc   Location `json:"location"`
}

type LabelStmt struct {
    Type      string `json:"type"` // "LabelStmt"
    Label     string `json:"label"`
    Statement Stmt   `json:"statement"`
    Loc       Location `json:"location"`
}
```

### Expressions (публичные)

```go
type VariableExpr struct {
    Type string   `json:"type"` // "VariableExpr"
    Name string   `json:"name"`
    Loc  Location `json:"location"`
}

type IntLiteral struct {
    Type  string   `json:"type"` // "IntLiteral"
    Value int      `json:"value"`
    Loc   Location `json:"location"`
}

type BinaryExpr struct {
    Type     string `json:"type"` // "BinaryExpr"
    Left     Expr   `json:"left"`
    Operator string `json:"operator"`
    Right    Expr   `json:"right"`
    Loc      Location `json:"location"`
}

type UnaryExpr struct {
    Type      string `json:"type"` // "UnaryExpr"
    Operator  string `json:"operator"`
    Operand   Expr   `json:"operand"`
    IsPostfix bool   `json:"isPostfix"`
    Loc       Location `json:"location"`
}

type AssignmentExpr struct {
    Type     string `json:"type"` // "AssignmentExpr"
    Left     Expr   `json:"left"`
    Operator string `json:"operator"`
    Right    Expr   `json:"right"`
    Loc      Location `json:"location"`
}

type CallExpr struct {
    Type         string `json:"type"` // "CallExpr"
    FunctionName string `json:"functionName"`
    Arguments    []Expr `json:"arguments"`
    Loc          Location `json:"location"`
}

type ArrayAccessExpr struct {
    Type  string `json:"type"` // "ArrayAccessExpr"
    Array Expr   `json:"array"`
    Index Expr   `json:"index"`
    Loc   Location `json:"location"`
}

type ArrayInitExpr struct {
    Type     string `json:"type"` // "ArrayInitExpr"
    Elements []Expr `json:"elements"`
    Loc      Location `json:"location"`
}
```

> Внутренний тип `Identifier` существует в `internal/domain/structs`, но в публичном API используется `VariableExpr`.

---

## Что реально конвертируется из CST

`ConvertStmt` поддерживает:

- `declaration`
- `function_definition`
- `if_statement`
- `while_statement`
- `do_statement`
- `for_statement`
- `return_statement`
- `compound_statement`
- `expression_statement`
- `assignment_expression` (как `ExprStmt`)
- `break_statement`
- `continue_statement`
- `goto_statement`
- `labeled_statement`
- `comment` (игнорируется)

`ConvertExpr` поддерживает:

- `identifier`
- `number_literal`
- `binary_expression`
- `unary_expression`
- `update_expression`
- `assignment_expression`
- `call_expression`
- `subscript_expression`
- `initializer_list`
- `pointer_expression`
- `parenthesized_expression` (разворачивается)
- `comment` (игнорируется)

---

## Важные детали поведения

- `else if` представляется как `ElseBlock` с вложенным `IfStmt`.
- Комментарии не попадают в AST.
- Отрицательные числовые литералы представляются как `UnaryExpr("-", IntLiteral(...))`, а не как `IntLiteral` с отрицательным `Value`.
- В `Program.Declarations` могут быть и глобальные переменные, и функции.

---

## Совместимость с другими сервисами проекта

- `interpreter-service` и `semantic-analyzer-service` используют типы из `pkg/converter`.
- Для внешнего кода опирайтесь на переэкспортированные типы из `pkg/converter`, а не на `internal/domain/structs` напрямую.

---

## Быстрая проверка

```bash
cd cst-to-ast-service
go test ./pkg/converter -v
```

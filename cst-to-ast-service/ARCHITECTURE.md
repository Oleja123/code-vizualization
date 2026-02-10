# Справочник AST для CST-to-AST конвертера

## Обзор

Этот документ описывает структуру Abstract Syntax Tree (AST), генерируемую CST-to-AST конвертером. AST представляет упрощенное подмножество языка C, оптимизированное для образовательных интерпретаторов и визуализаторов.

**Область применения**: Написание интерпретаторов, анализаторов кода, визуализаторов выполнения.

---

## API конвертера

### Точка входа

```go
import "github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/converter"

conv := converter.NewCConverter()
ast, err := conv.ConvertToAST(sourceCode)
if err != nil {
    // Обработка ошибок парсинга или конвертации
}
```

**Возвращает**: `*structs.Program` - корневой узел AST

---

## Структура AST

### Иерархия узлов

```
Node (интерфейс)
├── Stmt (операторы)
│   ├── VariableDecl
│   ├── FunctionDecl
│   ├── IfStmt
│   ├── WhileStmt
│   ├── ForStmt
│   ├── ReturnStmt
│   ├── BlockStmt
│   ├── ExprStmt
│   ├── BreakStmt
│   └── ContinueStmt
└── Expr (выражения)
    ├── VariableExpr
    ├── IntLiteral
    ├── BinaryExpr
    ├── UnaryExpr
    ├── AssignmentExpr
    ├── CallExpr
    ├── ArrayAccessExpr
    └── ArrayInitExpr
```

---

## Базовые типы

### Location

Позиция в исходном коде (для каждого узла):

```go
type Location struct {
    Line      uint32 `json:"line"`
    Column    uint32 `json:"column"`
    EndLine   uint32 `json:"endLine"`
    EndColumn uint32 `json:"endColumn"`
}
```

**Использование в интерпретаторе**:
- Отображение текущей строки выполнения
- Генерация понятных сообщений об ошибках
- Подсветка кода при визуализации

### Type

Система типов (только `int` с модификаторами):

```go
type Type struct {
    BaseType     string `json:"baseType"`     // Всегда "int"
    PointerLevel int    `json:"pointerLevel"` // 0=int, 1=int*, 2=int**
    ArraySize    *int   `json:"arraySize"`    // nil или размер массива
}
```

**Примеры**:
- `int x` → `{BaseType: "int", PointerLevel: 0, ArraySize: nil}`
- `int *p` → `{BaseType: "int", PointerLevel: 1, ArraySize: nil}`
- `int **pp` → `{BaseType: "int", PointerLevel: 2, ArraySize: nil}`
- `int arr[10]` → `{BaseType: "int", PointerLevel: 0, ArraySize: &10}`
- `int *arr[5]` → `{BaseType: "int", PointerLevel: 1, ArraySize: &5}`

**Использование в интерпретаторе**:
- Определение размера аллокации памяти
- Разыменование указателей
- Валидация операций с типами

---

## Операторы (Statements)

### 1. Program

Корневой узел программы:

```go
type Program struct {
    Declarations []Stmt   `json:"declarations"` // Переменные и функции
    Loc          Location `json:"loc"`
}
```

**Содержит**: Глобальные `VariableDecl` и `FunctionDecl`

**Интерпретация**: Точка входа, загрузка глобальных переменных перед выполнением `main()`

---

### 2. VariableDecl

Объявление переменной:

```go
type VariableDecl struct {
    Name  string   `json:"name"`
    Type  Type     `json:"type"`
    Init  Expr     `json:"init,omitempty"` // nil если без инициализации
    Loc   Location `json:"loc"`
}
```

**Примеры**:
```c
int x;              // Init = nil
int y = 42;         // Init = IntLiteral(42)
int *p = &x;        // Init = UnaryExpr("&", VariableExpr("x"))
int arr[5] = {1,2}; // Init = ArrayInitExpr([1,2])
```

**Интерпретация**:
- Аллоцировать память в текущей области видимости
- Если `Init != nil`, вычислить и присвоить значение
- Для массивов: аллоцировать `ArraySize * sizeof(int)` байт

---

### 3. FunctionDecl

Объявление функции:

```go
type FunctionDecl struct {
    Name       string        `json:"name"`
    ReturnType Type          `json:"returnType"`
    Params     []VariableDecl `json:"params"`
    Body       *BlockStmt    `json:"body"`
    Loc        Location      `json:"loc"`
}
```

**Примеры**:
```c
void main() { }                    // ReturnType.BaseType = "void"
int add(int a, int b) { return a+b; }
int *createArray(int size) { }     // ReturnType.PointerLevel = 1
```

**Интерпретация**:
- Сохранить в таблице функций
- При вызове: создать новый scope, привязать параметры, выполнить `Body`
- При `return`: вычислить выражение и вернуться к месту вызова

---

### 4. IfStmt

Условный оператор с поддержкой `else if`:

```go
type IfStmt struct {
    Condition  Expr           `json:"condition"`
    ThenBlock  Stmt           `json:"thenBlock"`
    ElseIfList []ElseIfClause `json:"elseIf,omitempty"`
    ElseBlock  Stmt           `json:"elseBlock,omitempty"` // nil если нет else
    Loc        Location       `json:"loc"`
}

type ElseIfClause struct {
    Condition Expr     `json:"condition"`
    Block     Stmt     `json:"block"`
    Loc       Location `json:"loc"`
}
```

**Примеры**:
```c
if (x > 0) { return 1; }
// IfStmt{Condition: ..., ThenBlock: ..., ElseIfList: nil, ElseBlock: nil}

if (x > 0) {
    return 1;
} else if (x < 0) {
    return -1;
} else {
    return 0;
}
// IfStmt{
//   Condition: BinaryExpr(">", x, 0),
//   ThenBlock: ...,
//   ElseIfList: [ElseIfClause{Condition: BinaryExpr("<", x, 0), Block: ...}],
//   ElseBlock: BlockStmt(...)
// }
```

**Интерпретация**:
1. Вычислить `Condition`
2. Если `true` → выполнить `ThenBlock`, выйти
3. Иначе: для каждого `ElseIfClause`:
   - Вычислить `Condition`
   - Если `true` → выполнить `Block`, выйти
4. Если все `false` и `ElseBlock != nil` → выполнить `ElseBlock`

---

### 5. WhileStmt

Цикл `while`:

```go
type WhileStmt struct {
    Condition Expr     `json:"condition"`
    Body      Stmt     `json:"body"`
    Loc       Location `json:"loc"`
}
```

**Интерпретация**:
1. Вычислить `Condition`
2. Если `true` → выполнить `Body`, вернуться к шагу 1
3. Если `false` → выйти из цикла
4. `break` в `Body` → немедленный выход
5. `continue` в `Body` → переход к шагу 1

---

### 6. ForStmt

Цикл `for`:

```go
type ForStmt struct {
    Init      Stmt     `json:"init,omitempty"`      // VariableDecl или ExprStmt
    Condition Expr     `json:"condition,omitempty"`
    Update    Expr     `json:"update,omitempty"`
    Body      Stmt     `json:"body"`
    Loc       Location `json:"loc"`
}
```

**Примеры**:
```c
for (int i = 0; i < 10; i++) { sum += i; }
for (; x < 100; x *= 2) { }         // Init = nil
for (;;) { }                        // Бесконечный цикл
```

**Интерпретация**:
1. Если `Init != nil` → выполнить один раз
2. Вычислить `Condition` (если nil, считать `true`)
3. Если `false` → выйти
4. Выполнить `Body`
5. Если `Update != nil` → вычислить
6. Вернуться к шагу 2

---

### 7. ReturnStmt

Оператор возврата:

```go
type ReturnStmt struct {
    Value Expr     `json:"value,omitempty"` // nil для void функций
    Loc   Location `json:"loc"`
}
```

**Интерпретация**:
- Если `Value != nil` → вычислить, поместить в регистр возврата
- Восстановить предыдущий stack frame
- Вернуть управление в место вызова

---

### 8. BlockStmt

Блок операторов `{ ... }`:

```go
type BlockStmt struct {
    Statements []Stmt   `json:"statements"`
    Loc        Location `json:"loc"`
}
```

**Интерпретация**:
- Создать новую область видимости (scope)
- Выполнить `Statements` по порядку
- Уничтожить scope при выходе

---

### 9. ExprStmt

Выражение как оператор:

```go
type ExprStmt struct {
    Expr Expr     `json:"expr"`
    Loc  Location `json:"loc"`
}
```

**Примеры**: `x = 5;`, `foo();`, `x++;`

**Интерпретация**: Вычислить `Expr`, игнорировать результат

---

### 10. BreakStmt / ContinueStmt

```go
type BreakStmt struct {
    Loc Location `json:"loc"`
}

type ContinueStmt struct {
    Loc Location `json:"loc"`
}
```

**Интерпретация**: Управление потоком в циклах (требует context цикла)

---

## Выражения (Expressions)

### 1. VariableExpr

Ссылка на переменную:

```go
type VariableExpr struct {
    Name string   `json:"name"`
    Loc  Location `json:"loc"`
}
```

**Интерпретация**: Найти в scope, вернуть значение (или адрес для lvalue)

---

### 2. IntLiteral

Целочисленная константа:

```go
type IntLiteral struct {
    Value int      `json:"value"`
    Loc   Location `json:"loc"`
}
```

**Интерпретация**: Вернуть `Value`

---

### 3. BinaryExpr

Бинарная операция:

```go
type BinaryExpr struct {
    Op    string   `json:"op"` // +, -, *, /, %, ==, !=, <, >, <=, >=, &&, ||, &, |, ^, <<, >>
    Left  Expr     `json:"left"`
    Right Expr     `json:"right"`
    Loc   Location `json:"loc"`
}
```

**Поддерживаемые операторы**:
- Арифметика: `+`, `-`, `*`, `/`, `%`
- Сравнение: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Логические: `&&`, `||`
- Битовые: `&`, `|`, `^`, `<<`, `>>`

**Интерпретация**: Вычислить `Left` и `Right`, применить `Op`

---

### 4. UnaryExpr

Унарная операция:

```go
type UnaryExpr struct {
    Op      string   `json:"op"` // -, !, *, &, ++, --
    Operand Expr     `json:"operand"`
    IsPostfix bool   `json:"isPostfix,omitempty"` // для ++ и --
    Loc     Location `json:"loc"`
}
```

**Операторы**:
- `-` (унарный минус), `!` (логическое НЕ)
- `*` (разыменование), `&` (взятие адреса)
- `++`, `--` (инкремент/декремент, prefix или postfix)

**Интерпретация**:
- `*ptr` → разыменовать указатель
- `&var` → вернуть адрес переменной
- `++x` → инкремент, вернуть новое значение
- `x++` → вернуть старое значение, затем инкремент

---

### 5. AssignmentExpr

Присваивание:

```go
type AssignmentExpr struct {
    Operator string   `json:"operator"` // =, +=, -=, /=, %=, &=, |=, ^=, <<=, >>=
    Left     Expr     `json:"left"`     // Должно быть lvalue
    Right    Expr     `json:"right"`
    Loc      Location `json:"loc"`
}
```

**Операторы**:
- Простое: `=`
- Составные: `+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=`

**Примеры**:
```c
x = 5;          // Operator="=", Left=VariableExpr("x"), Right=IntLiteral(5)
arr[i] = 10;    // Left=ArrayAccessExpr(...)
*ptr = 20;      // Left=UnaryExpr("*", ...)
x += 3;         // Operator="+=", эквивалентно x = x + 3
```

**Интерпретация**:
1. Вычислить `Right`
2. Вычислить адрес `Left` (lvalue)
3. Для составных операторов: вычислить `Left op Right`
4. Записать результат по адресу `Left`

**Валидация lvalue**: `Left` должно быть `VariableExpr`, `ArrayAccessExpr` или `UnaryExpr("*", ...)`

---

### 6. CallExpr

Вызов функции:

```go
type CallExpr struct {
    Callee    Expr     `json:"callee"` // Обычно VariableExpr
    Arguments []Expr   `json:"arguments"`
    Loc       Location `json:"loc"`
}
```

**Интерпретация**:
1. Найти функцию по имени (`Callee`)
2. Вычислить все `Arguments`
3. Создать новый stack frame
4. Привязать параметры к аргументам
5. Выполнить тело функции
6. Вернуть результат

---

### 7. ArrayAccessExpr

Доступ к элементу массива:

```go
type ArrayAccessExpr struct {
    Array Expr     `json:"array"` // VariableExpr или другой ArrayAccessExpr
    Index Expr     `json:"index"`
    Loc   Location `json:"loc"`
}
```

**Примеры**:
```c
arr[i]       // Array=VariableExpr("arr"), Index=VariableExpr("i")
matrix[i][j] // Array=ArrayAccessExpr(VariableExpr("matrix"), i), Index=j
```

**Интерпретация**:
1. Вычислить `Array` (получить базовый адрес)
2. Вычислить `Index`
3. Вычислить адрес: `base + index * sizeof(element)`
4. Вернуть значение или адрес (для lvalue)

---

### 8. ArrayInitExpr

Инициализация массива:

```go
type ArrayInitExpr struct {
    Elements []Expr   `json:"elements"`
    Loc      Location `json:"loc"`
}
```

**Примеры**:
```c
int arr[5] = {1, 2, 3};     // Elements=[IntLiteral(1), IntLiteral(2), IntLiteral(3)]
int mat[2][2] = {{1,2},{3,4}}; // Вложенные ArrayInitExpr
```

**Интерпретация**:
- Вычислить все `Elements` по порядку
- Записать в память массива последовательно
- Оставшиеся элементы заполнить нулями

---

## Обработка комментариев

**Важно**: Комментарии (`//` и `/* */`) **фильтруются** на этапе конвертации и **не присутствуют в AST**.

Интерпретатору не нужно их обрабатывать.

---

## Примеры использования в интерпретаторе

### Пример 1: Обход программы

```go
func Execute(program *structs.Program) error {
    // 1. Загрузить глобальные переменные
    for _, decl := range program.Declarations {
        switch d := decl.(type) {
        case *structs.VariableDecl:
            globalScope.Declare(d.Name, d.Type, d.Init)
        case *structs.FunctionDecl:
            functions[d.Name] = d
        }
    }
    
    // 2. Вызвать main()
    mainFunc := functions["main"]
    return executeFunction(mainFunc, []Value{})
}
```

### Пример 2: Вычисление выражения

```go
func Eval(expr structs.Expr, scope *Scope) (Value, error) {
    switch e := expr.(type) {
    case *structs.IntLiteral:
        return IntValue(e.Value), nil
    
    case *structs.VariableExpr:
        return scope.Get(e.Name)
    
    case *structs.BinaryExpr:
        left, _ := Eval(e.Left, scope)
        right, _ := Eval(e.Right, scope)
        return applyOp(e.Op, left, right), nil
    
    case *structs.AssignmentExpr:
        value, _ := Eval(e.Right, scope)
        addr := getAddress(e.Left, scope)
        memory.Write(addr, value)
        return value, nil
        
    // ... остальные типы
    }
}
```

### Пример 3: Выполнение if-else-if

```go
func ExecuteIf(stmt *structs.IfStmt, scope *Scope) error {
    // Проверить основное условие
    cond, _ := Eval(stmt.Condition, scope)
    if cond.IsTrue() {
        return Execute(stmt.ThenBlock, scope)
    }
    
    // Проверить else if цепочку
    for _, elseIf := range stmt.ElseIfList {
        cond, _ := Eval(elseIf.Condition, scope)
        if cond.IsTrue() {
            return Execute(elseIf.Block, scope)
        }
    }
    
    // Выполнить else блок
    if stmt.ElseBlock != nil {
        return Execute(stmt.ElseBlock, scope)
    }
    
    return nil
}
```

---

## Рекомендации для интерпретатора

### Необходимые компоненты

1. **Memory Manager**: Управление heap/stack для указателей и массивов
2. **Scope Manager**: Иерархия областей видимости (global → function → block)
3. **Call Stack**: Трассировка вызовов функций для рекурсии
4. **Value System**: Представление значений (int, pointer, array)

### Визуализация выполнения

Используйте `Location` для:
- Подсветки текущей исполняемой строки
- Отображения stack trace с номерами строк
- Связывания значений переменных с местом объявления

### Обработка ошибок

- **Runtime errors**: Деление на ноль, выход за границы массива, null pointer
- **Type errors**: Несоответствие типов (можно проверить на этапе интерпретации)
- **Location**: Используйте `stmt.Loc` или `expr.Loc` для сообщений об ошибках

---

## Ограничения текущей версии

- **Типы**: Только `int` (нет `float`, `char`, `struct`)
- **Возвращаемые типы**: `int`, `int*`, `int**`, `void` (нет `float`, `double`, `char`)
- **Строки**: Не поддерживаются
- **Динамическая память**: Нет `malloc`/`free`
- **Многомерные массивы**: Только синтаксис `arr[i][j]`, не `int arr[2][3]`

---

## API Summary

### Основные структуры

```go
// Точка входа
type Program struct {
    Declarations []Stmt
    Loc Location
}

// Операторы (12 типов)
VariableDecl, FunctionDecl, IfStmt, WhileStmt, ForStmt,
ReturnStmt, BlockStmt, ExprStmt, BreakStmt, ContinueStmt

// Выражения (8 типов)
VariableExpr, IntLiteral, BinaryExpr, UnaryExpr,
AssignmentExpr, CallExpr, ArrayAccessExpr, ArrayInitExpr

// Вспомогательные
Location, Type, ElseIfClause
```

### Методы type assertion

```go
// Проверка типа узла
if stmt, ok := node.(structs.Stmt); ok { }
if expr, ok := node.(structs.Expr); ok { }

// Конкретный тип
if ifStmt, ok := stmt.(*structs.IfStmt); ok {
    // Работа с ifStmt.Condition, ifStmt.ElseIfList, ...
}

```

---

## Тестирование и запуск

### Запуск тестов

```bash
cd cst-to-ast-service
go test ./internal/converter/... -v -cover
```

**Покрытие**: 81.7% (47 тестов)

### Примеры использования

```bash
# Базовый пример (факториал)
go run cmd/example/main.go

# Массивы и указатели
go run cmd/advanced-example/main.go

# Условные операторы
go run cmd/else-if-example/main.go
```

---

## Технические детали (для расширения конвертера)

### Архитектура конвертера

```
C source → tree-sitter parser → CST → CConverter → AST
```

**Ключевые принципы**:
1. Маркер методы (`StmtNode()`, `ExprNode()`) для type safety
2. Рекурсивная обработка `else if` цепочек (преобразование вложенной структуры в плоский список)
3. Фильтрация комментариев на уровне конвертации

### Расширение функциональности

Для добавления новых конструкций:
1. Добавить struct в `internal/domain/structs/ast.go`
2. Реализовать интерфейс `Stmt` или `Expr`
3. Добавить метод `convertXxx()` в `internal/converter/converter.go`
4. Обновить dispatcher (`ConvertStmt` или `ConvertExpr`)

### Зависимости

- `github.com/smacker/go-tree-sitter` - парсер C
- Стандартная библиотека Go

# Справочник AST для CST-to-AST конвертера

## Обзор

Этот документ описывает структуру Abstract Syntax Tree (AST), генерируемую CST-to-AST конвертером. AST представляет упрощенное подмножество языка C, оптимизированное для образовательных интерпретаторов и визуализаторов.

**Область применения**: Написание интерпретаторов, визуализаторов выполнения, отладчиков кода на C.

**Примечание**: Семантический анализ (проверка типов, таблица символов, и т.д.) выполняется **отдельным микросервисом** (`semantic-analyzer-service`). Этот сервис занимается исключительно конвертацией CST в AST.

---

## API конвертера

### Точка входа

```go
import "github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"

// Создать конвертер
conv := converter.New()

// Основной метод: парсинг кода в AST
program, err := conv.ParseToAST(sourceCode)
if err != nil {
    // Обработка ошибки парсинга
    loc := err.GetLocation()
    code := err.GetCode()
    msg := err.GetMessage()
    nodeType := err.GetNodeType()
}
```

**Возвращает**: `(*Program, *ConverterError)` где:
- `Program` - корневой узел AST (nil при ошибке)
- `ConverterError` - детальная информация об ошибке (nil при успехе)

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
│   ├── DoWhileStmt
│   ├── ForStmt
│   ├── ReturnStmt
│   ├── BlockStmt
│   ├── ExprStmt
│   ├── BreakStmt
│   ├── ContinueStmt
│   ├── GotoStmt
│   └── LabelStmt
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
    ArraySizes   []int  `json:"arraySizes"`   // Поддержка многомерных массивов
}
```

**Примеры**:
- `int x` → `{BaseType: "int", PointerLevel: 0, ArraySizes: []}`
- `int *p` → `{BaseType: "int", PointerLevel: 1, ArraySizes: []}`
- `int **pp` → `{BaseType: "int", PointerLevel: 2, ArraySizes: []}`
- `int arr[10]` → `{BaseType: "int", PointerLevel: 0, ArraySizes: [10]}`
- `int arr[2][3]` → `{BaseType: "int", PointerLevel: 0, ArraySizes: [2, 3]}`
- `int *arr[5]` → `{BaseType: "int", PointerLevel: 1, ArraySizes: [5]}`

**Использование в интерпретаторе**:
- Определение размера аллокации памяти
- Разыменование указателей
- Навигация по многомерным массивам

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

Условный оператор с поддержкой `else if`. Else if представляется как else с вложенным if (как в C):

```go
type IfStmt struct {
    Condition Expr     `json:"condition"`
    ThenBlock Stmt     `json:"thenBlock"`
    ElseBlock Stmt     `json:"elseBlock,omitempty"` // nil если нет else; может быть вложенный IfStmt для else if
    Loc       Location `json:"loc"`
}
```

**Примеры**:
```c
if (x > 0) { return 1; }
// IfStmt{Condition: ..., ThenBlock: ..., ElseBlock: nil}

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
//   ElseBlock: IfStmt{  // Else-if представлен как вложенный if
//     Condition: BinaryExpr("<", x, 0),
//     ThenBlock: ...,
//     ElseBlock: BlockStmt(...)
//   }
// }
```

**Интерпретация**:
1. Вычислить `Condition`
2. Если `true` → выполнить `ThenBlock`, выйти
3. Иначе: если `ElseBlock != nil`:
   - Если это IfStmt (else if) → рекурсивно интерпретировать
   - Иначе (else) → выполнить `ElseBlock`

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

### 5.1 DoWhileStmt

Цикл `do while` (выполняется минимум один раз):

```go
type DoWhileStmt struct {
    Body      Stmt     `json:"body"`
    Condition Expr     `json:"condition"`
    Loc       Location `json:"loc"`
}
```

**Интерпретация**:
1. Выполнить `Body`
2. Вычислить `Condition`
3. Если `true` → вернуться к шагу 1
4. Если `false` → выйти из цикла
5. `break` в `Body` → немедленный выход
6. `continue` в `Body` → переход к шагу 2

**Ключевое отличие от while**: Тело выполняется **минимум один раз** перед проверкой условия.

**Пример:**
```c
int x = 0;
do {
    x = x + 1;
    printf("%d\n", x);
} while (x < 5);  // Выведет 1, 2, 3, 4, 5
```

---

### 6. ForStmt

Цикл `for`:

```go
type ForStmt struct {
    Init      Stmt     `json:"init,omitempty"`      // VariableDecl или ExprStmt
    Condition Expr     `json:"condition,omitempty"`
    Post      Stmt     `json:"post,omitempty"`      // ExprStmt для Update
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
5. Если `Post != nil` → выполнить
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

### 11. GotoStmt

Безусловный переход на метку:

```go
type GotoStmt struct {
    Type  string   `json:"type"`    // "GotoStmt"
    Label string   `json:"label"`   // Имя метки для перехода
    Loc   Location `json:"loc"`
}
```

**Интерпретация**: Выполнить скачок на соответствующий LabelStmt. Требует корректной навигации по метке.

**Пример:**
```c
int i = 0;
loop: i = i + 1;
if (i < 10) goto loop;
```

---

### 12. LabelStmt

Метка с следующим оператором:

```go
type LabelStmt struct {
    Type      string `json:"type"`      // "LabelStmt"
    Label     string `json:"label"`     // Имя метки
    Statement Stmt   `json:"statement"` // Оператор, связанный с меткой
    Loc       Location `json:"loc"`
}
```

**Интерпретация**: Точка назначения для goto. При выполнении: выполнить оператор Statement.

**Пример:**
```c
end: return 0;  // LabelStmt("end", ReturnStmt(0))
```

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
matrix[0][1] // Вложенные ArrayAccessExpr
```

**Интерпретация** (для многомерных массивов):
1. Вычислить `Array` (получить базовый адрес или подмассив)
2. Вычислить `Index`
3. Для каждого уровня доступа: вычислить смещение `index * stride`
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
int arr[5] = {1, 2, 3};              // Elements=[IntLiteral(1), IntLiteral(2), IntLiteral(3)]
int mat[2][2] = {{1,2},{3,4}};       // Вложенные ArrayInitExpr
int vec[] = {10, 20, 30};            // Size выводится из Elements
```

**Интерпретация**:
- Вычислить все `Elements` по порядку
- Записать в память массива последовательно
- Оставшиеся элементы (если размер известен) заполнить нулями

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

## Рекомендации для интерпретатора / Визуализатора

### Архитектура интерпретатора

```
CST-to-AST Service
        ↓
      AST
        ↓
Semantic Analyzer Service (опционально)
        ↓
Annotated AST (типы, таблица символов)
        ↓
    Интерпретатор
```

### Необходимые компоненты в интерпретаторе

1. **Memory Manager**: Управление heap/stack для указателей и массивов
2. **Scope Manager**: Иерархия областей видимости (global → function → block)
3. **Call Stack**: Трассировка вызовов функций для рекурсии
4. **Value System**: Представление значений (int, pointer, array)
5. **Executor**: Обход AST и выполнение операций

### Визуализация выполнения

Используйте `Location` в каждом узле для:
- Подсветки текущей исполняемой строки в IDE
- Отображения stack trace с номерами строк и имен функций
- Связывания значений переменных с местом объявления
- Визуализации состояния памяти по шагам

### Обработка ошибок

- **Runtime errors**: Деление на ноль, выход за границы массива, null pointer dereference
- **Semantic errors**: Проверка типов, таблица символов — обрабатывает **отдельный сервис** (semantic-analyzer-service)
- **Location**: Используйте `stmt.Loc` или `expr.Loc` для сообщений об ошибках

---

## Возможности и ограничения

### Поддерживается
- **Типы**: `int`, указатели (`int*`, `int**`), массивы (`int[10]`), многомерные массивы (`int[2][3]`)
- **Возвращаемые типы**: `int`, `int*`, `int**`, `void`
- **Циклы**: `for`, `while`, `do-while` (поддержка `break`/`continue`)
- **Условные операторы**: `if/else if/else`
- **Функции**: Рекурсия, вложенные вызовы
- **Операции**: Арифметика, сравнение, логические, битовые, указатели

### Не поддерживается
- **Типы**: `float`, `double`, `char` (только `int`)
- **Структуры**: Нет `struct`, `union`, `enum`
- **Строки**: Нет встроенной поддержки (можно использовать массивы)
- **Динамическая память**: Нет `malloc`/`free` (управление памятью за интерпретатором)
- **Стандартная библиотека**: Функции `printf`, `scanf` и т.д. на уровне CST-to-AST не обрабатываются

---

## API Summary

### Основные структуры

```go
// Точка входа
type Program struct {
    Declarations []Stmt  // Глобальные переменные и функции
    Loc          Location
}

// Операторы (13 типов)
VariableDecl    // Объявление переменной
FunctionDecl    // Объявление функции
IfStmt          // Условный оператор с else if цепочкой
WhileStmt       // Цикл while
DoWhileStmt     // Цикл do while
ForStmt         // Цикл for
ReturnStmt      // Оператор return
BlockStmt       // Блок { ... }
ExprStmt        // Выражение-оператор
BreakStmt       // break
ContinueStmt    // continue
GotoStmt        // goto
LabelStmt       // Метка с оператором

// Выражения (8 типов)
VariableExpr    // Ссылка на переменную
IntLiteral      // Целочисленная константа
BinaryExpr      // Бинарная операция (+, -, *, /, %,  ==, !=, <, >, <=, >=, &&, ||, &, |, ^, <<, >>)
UnaryExpr       // Унарная операция (-, !, *, &, ++, --)
AssignmentExpr  // Присваивание (=, +=, -=, и т.д.)
CallExpr        // Вызов функции
ArrayAccessExpr // Доступ к элементу массива
ArrayInitExpr   // Инициализация массива

// Вспомогательные
Location        // Позиция в коде
Type            // Описание типа
Parameter       // Параметр функции
```

### Методы type assertion

```go
// Проверка типа узла
if stmt, ok := node.(structs.Stmt); ok { }
if expr, ok := node.(structs.Expr); ok { }

// Конкретный тип
if ifStmt, ok := stmt.(*structs.IfStmt); ok {
    // Работа с ifStmt.Condition, ifStmt.ElseBlock
    // Если ElseBlock это IfStmt, то это else if
}

```

---

## Тестирование и запуск

### Запуск тестов

```bash
cd cst-to-ast-service
go test ./pkg/converter -v -cover
```

**Текущее состояние**: 80+ тестов, полное покрытие основных конструкций C

### Использование как библиотеки

```bash
# Добавить зависимость в ваш проект
go get github.com/Oleja123/code-vizualization/cst-to-ast-service

# Использование в коде
conv := converter.New()
ast, err := conv.ParseToAST("int main() { return 0; }")
if err != nil {
    log.Fatal(err)
}
// Использовать ast...
```

---

## Технические детали (для разработчиков)

### Архитектура конвертера

```
C source code
        ↓
  tree-sitter parser (Concrete Syntax Tree)
        ↓
  CConverter (pkg/converter)
        ↓
AST (internal/domain/structs)
        ↓
Интерпретатор / Анализатор
```

**Ключевые принципы реализации**:
1. **Type safety**: Маркер методы (`StmtNode()`, `ExprNode()`) через интерфейсы
2. **Else-if обработка**: Else if представлен как else с вложенным if (как в C), вместо плоского списка
3. **Комментарии**: Фильтруются на уровне tree-sitter парсера (не присутствуют в CST)
4. **Многомерные массивы**: Поддержка через `ArraySizes []int` в структуре `Type`

### Расширение функциональности

Для добавления новых конструкций:
1. Добавить struct в `internal/domain/structs/ast.go`
2. Реализовать интерфейс `Stmt` или `Expr`
3. Добавить метод `convertXxx()` в `internal/converter/converter.go`
4. Обновить dispatcher (`ConvertStmt` или `ConvertExpr`)

### Зависимости

- `github.com/smacker/go-tree-sitter` - парсер C (tree-sitter)
- `github.com/tree-sitter/tree-sitter` - C language definitions
- Стандартная библиотека Go (no external deps besides tree-sitter)

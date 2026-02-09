# Архитектура CST-to-AST конвертера

## Обзор

CST-to-AST конвертер преобразует Concrete Syntax Tree (получаемое от tree-sitter) в упрощенное Abstract Syntax Tree для языка C. Архитектура следует принципам Domain-Driven Design с четким разделением интерфейсов и реализации.

## Слои архитектуры

### 1. Domain Layer (`internal/domain/`)

#### Interfaces (`internal/domain/interfaces/interfaces.go`)

Определяет основные интерфейсы для типизации AST узлов:

```go
type Node interface {
    // Маркер для идентификации узла как базового узла
}

type Stmt interface {
    Node
    StmtNode() // Маркер метод для типизации
}

type Expr interface {
    Node
    ExprNode() // Маркер метод для типизации
}

type Converter interface {
    Parse(sourceCode []byte) (*sitter.Tree, error)
    ConvertToProgram(tree *sitter.Tree, sourceCode []byte) (*Program, error)
    ConvertStmt(node *sitter.Node, sourceCode []byte) Stmt
    ConvertExpr(node *sitter.Node, sourceCode []byte) Expr
}
```

**Назначение**: Определяет контракты для работы с AST узлами. Маркер методы обеспечивают типобезопасность в Go (так как Go не имеет наследования).

#### Структуры (`internal/domain/structs/ast.go`)

Содержит определения всех AST узлов. Основные типы:

##### Базовые типы

```go
type Location struct {
    Line    uint32
    Column  uint32
    EndLine   uint32
    EndColumn uint32
}
```

**Назначение**: Отслеживание позиции узла в исходном коде для:
- Генерации сообщений об ошибках
- Отладки
- Визуализации с указанием строк
- Привязки к исходному коду в образовательных интерпретаторах

```go
type Type struct {
    BaseType     string   // "int"
    PointerLevel int      // 0 for int, 1 for int*, 2 for int**
    ArraySize    *int     // nil for non-array, pointer to size for arrays
}
```

**Назначение**: Представление типов в упрощенной системе (только int с указателями и массивами).

##### Операторы (Statements) - 10 типов

1. **Program**: Корневой узел, содержит список объявлений функций и переменных
2. **VariableDecl**: Объявление переменной с опциональной инициализацией
3. **FunctionDecl**: Определение функции
4. **IfStmt**: Условный оператор с явной поддержкой else if
   - Использует `ElseIfClause` для каждого блока else if
   - Отделяет final `else` от цепочки else if
5. **ElseIfClause**: Представляет один блок else if (NEW)
   - Condition: условное выражение
   - Block: тело блока
   - Loc: позиция в коде
6. **WhileStmt**: Цикл while
7. **ForStmt**: Цикл for (инициализация, условие, инкремент)
8. **ReturnStmt**: Оператор return с опциональным выражением
9. **BlockStmt**: Блок операторов (составной оператор)
10. **ExprStmt**: Оператор-выражение
11. **BreakStmt**: Оператор break
12. **ContinueStmt**: Оператор continue

##### Выражения (Expressions) - 8 типов

1. **Identifier**: Идентификатор переменной
2. **IntLiteral**: Целочисленная константа
3. **BinaryExpr**: Бинарная операция (+, -, *, /, %, ==, !=, <, >, <=, >=, &&, ||)
4. **UnaryExpr**: Унарная операция (-, !, *, &)
5. **AssignmentExpr**: Присваивание (=)
6. **CallExpr**: Вызов функции
7. **ArrayAccessExpr**: Доступ к элементу массива (arr[index])
8. **ArrayInitExpr**: Инициализация массива ({val1, val2, ...})

### 2. Converter Layer (`internal/converter/converter.go`)

#### CConverter struct

```go
type CConverter struct {
    // Содержит reference на parser tree-sitter
}
```

#### Основные методы

##### Parse(sourceCode []byte) (*sitter.Tree, error)
- Создает parser tree-sitter для языка C
- Парсит исходный код в CST
- Возвращает дерево для последующей конвертации

##### ConvertToProgram(tree *sitter.Tree, sourceCode []byte) (*Program, error)
- Точка входа для конвертации
- Обходит корневой узел (program) CST
- Преобразует все объявления (переменные, функции)
- Собирает в структуру Program

##### ConvertStmt(node *sitter.Node, sourceCode []byte) Stmt
- Диспетчер для операторов
- На основе типа узла вызывает специализированный конвертер:
  - `convertVariableDeclaration()`
  - `convertFunctionDefinition()`
  - `convertIfStatement()`
  - `convertWhileStatement()`
  - `convertForStatement()`
  - `convertReturnStatement()`
  - `convertBlock()`
  - `convertExprStatement()`
  - `convertBreakStatement()`
  - `convertContinueStatement()`

##### ConvertExpr(node *sitter.Node, sourceCode []byte) Expr
- Диспетчер для выражений
- На основе типа узла вызывает специализированный конвертер:
  - `convertIdentifier()`
  - `convertNumber()`
  - `convertBinaryExpression()`
  - `convertUnaryExpression()`
  - `convertAssignmentExpression()`
  - `convertCallExpression()`
  - `convertArrayAccess()`
  - `convertArrayInitializer()`

#### Ключевые методы конвертации

##### convertIfStatement() - Обработка условных операторов

**Проблема**: tree-sitter использует позиционные дочерние узлы вместо именованных полей.

**Решение**:
```
if_statement children:
  [0] "if" keyword
  [1] parenthesized_expression (condition)
  [2] compound_statement (body)
  [3] else_clause (optional)
```

**Алгоритм**:
1. Ищем первое parenthesized_expression - это условие
2. Ищем первое compound_statement после условия - это тело
3. Если есть else_clause - передаем на обработку processElseClause()
4. processElseClause() рекурсивно обрабатывает else if цепочку

**Результат**: ElseIfList содержит все else if блоки, ElseBlock содержит final else (или nil)

##### processElseClause() - Рекурсивная обработка else цепочки (NEW)

```go
func (c *CConverter) processElseClause(elseClauseNode *sitter.Node, 
                                       sourceCode []byte) ([]ElseIfClause, Stmt)
```

**Назначение**: Превращает вложенные if-else-if структуры в плоский список ElseIfClause.

**Алгоритм**:
- Если содержит if_statement (else if):
  - Извлекаем условие и блок
  - Добавляем в elseIfList
  - Рекурсивно вызываем для nested else_clause
- Если содержит compound_statement (else):
  - Это final else блок

**Преимущества**: Явное представление else if упрощает интерпретацию и визуализацию.

#### Вспомогательные методы

##### extractText(node *sitter.Node, sourceCode []byte) string
- Извлекает текст из исходного кода по позиции узла

##### getType(declNode *sitter.Node, sourceCode []byte) (*Type, error)
- Парсит тип переменной/параметра
- Определяет уровень указателя и размер массива

##### getChildByType(node *sitter.Node, childType string) *sitter.Node
- Находит первого дочернего узла с указанным типом

##### getChildByContent(node *sitter.Node, content string) *sitter.Node
- Находит первого дочернего узла с указанным содержимым

## Поток данных

```
C source code
      ↓
tree-sitter parser (Parse)
      ↓
CST (Concrete Syntax Tree)
      ↓
CConverter.ConvertToProgram()
      ↓
ConvertStmt() / ConvertExpr() dispatchers
      ↓
Specialized converters (convertIfStatement, etc)
      ↓
AST Nodes (Program, IfStmt, ElseIfClause, etc)
      ↓
JSON serialization (encoding/json)
      ↓
Output
```

## Примеры конвертации

### Пример 1: Простой if-else if-else

**C код**:
```c
int grade(int score) {
    if (score >= 90) {
        return 5;
    } else if (score >= 80) {
        return 4;
    } else if (score >= 70) {
        return 3;
    } else {
        return 2;
    }
}
```

**CST структура (tree-sitter)**:
```
if_statement
├─ if
├─ parenthesized_expression (score >= 90)
├─ compound_statement ({ return 5; })
└─ else_clause
   └─ if_statement
      ├─ if
      ├─ parenthesized_expression (score >= 80)
      ├─ compound_statement ({ return 4; })
      └─ else_clause
         └─ if_statement
            ├─ if
            ├─ parenthesized_expression (score >= 70)
            ├─ compound_statement ({ return 3; })
            └─ else_clause
               └─ compound_statement ({ return 2; })
```

**AST структура (после конвертации)**:
```go
IfStmt{
    Condition: BinaryExpr{Op: ">=", Left: Identifier("score"), Right: IntLiteral(90)},
    ThenBlock: BlockStmt{...},
    ElseIfList: [
        ElseIfClause{
            Condition: BinaryExpr{Op: ">=", Left: Identifier("score"), Right: IntLiteral(80)},
            Block: BlockStmt{...}
        },
        ElseIfClause{
            Condition: BinaryExpr{Op: ">=", Left: Identifier("score"), Right: IntLiteral(70)},
            Block: BlockStmt{...}
        }
    ],
    ElseBlock: BlockStmt{...}
}
```

**JSON вывод**:
```json
{
    "condition": {...},
    "thenBlock": {...},
    "elseIf": [
        {
            "condition": {...},
            "block": {...}
        },
        {
            "condition": {...},
            "block": {...}
        }
    ],
    "elseBlock": {...}
}
```

## Ключевые решения архитектуры

### 1. Маркер методы вместо наследования
**Почему**: Go не имеет классического наследования. Маркер методы обеспечивают типобезопасность.
```go
func (i *IfStmt) StmtNode() {}     // IfStmt реализует Stmt
func (bl *IntLiteral) ExprNode() {} // IntLiteral реализует Expr
```

### 2. Явные Location для каждого узла
**Почему**: Критично для образовательных приложений (отладка, визуализация, ошибки).

### 3. Упрощение типов (только int)
**Почему**: Целевое образовательное приложение, сложность не требуется.

### 4. Явное представление else if (ElseIfClause)
**Почему**: 
- Отражает намерение программиста (else if - это одна конструкция)
- Упрощает интерпретацию (не нужно разбирать вложенные if)
- Улучшает визуализацию (понятная цепочка условий)

### 5. Рекурсивная обработка else цепочки (processElseClause)
**Почему**: Преобразует глубоко вложенную структуру в плоский список для удобства.

## Обработка ошибок

### Ошибки парсинга
```go
tree, err := conv.Parse(sourceCode)
if err != nil {
    // tree-sitter ошибки парсинга
}
```

### Ошибки конвертации
```go
ast, err := conv.ConvertToProgram(tree, sourceCode)
if err != nil {
    // Ошибки при конвертации (неподдерживаемые узлы и т.д.)
}
```

Каждый узел содержит Location для указания точного места ошибки.

## Расширяемость

### Добавление новой конструкции C

1. Добавить struct в `internal/domain/structs/ast.go`
2. Реализовать интерфейс Stmt или Expr
3. Добавить метод convertXxx в `internal/converter/converter.go`
4. Добавить case в ConvertStmt или ConvertExpr dispatcher
5. Добавить пример в `cmd/`

### Добавление новых операций

Обновить списки в методах:
- BinaryExpr для новых бинарных операций
- UnaryExpr для новых унарных операций

## Тестирование

Проект включает три рабочих примера:

1. **cmd/example/main.go** - Базовая функциональность (факториал)
2. **cmd/advanced-example/main.go** - Массивы, указатели, циклы
3. **cmd/else-if-example/main.go** - Else if конструкции (NEW)

Каждый пример:
- Содержит C код
- Парсит в CST
- Конвертирует в AST
- Выводит JSON для проверки структуры

## Зависимости

- `github.com/smacker/go-tree-sitter` - Парсер CST
- Стандартная библиотека Go (context, encoding/json, fmt, errors)

## Запуск и отладка

### Просмотр CST структуры
```bash
go run cmd/debug-cst/main.go
```

### Просмотр else_clause структуры
```bash
go run cmd/debug-else-if/main.go
```

### Сборка и запуск
```bash
go build ./...
go test ./...
```

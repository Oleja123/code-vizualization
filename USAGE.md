# Руководство по использованию CST-to-AST конвертера

## Быстрый старт

### Установка

```bash
# Клонировать репозиторий
git clone https://github.com/Oleja123/code-vizualization.git
cd code-vizualization/cst-to-ast-service

# Убедиться что установлен Go 1.21+
go version
```

### Запуск примеров

```bash
# Пример 1: Простая функция с рекурсией
go run cmd/example/main.go

# Пример 2: Массивы, указатели, циклы
go run cmd/advanced-example/main.go

# Пример 3: Конструкции else if
go run cmd/else-if-example/main.go
```

## Программное использование

### Базовый пример

```go
package main

import (
    "fmt"
    "log"
    "encoding/json"
    
    "github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/converter"
)

func main() {
    // Исходный код C
    sourceCode := []byte(`
        int add(int a, int b) {
            return a + b;
        }
    `)
    
    // Создаем конвертер
    conv := converter.NewCConverter()
    
    // Парсим в CST
    tree, err := conv.Parse(sourceCode)
    if err != nil {
        log.Fatalf("Parse error: %v", err)
    }
    
    // Конвертируем в AST
    ast, err := conv.ConvertToProgram(tree, sourceCode)
    if err != nil {
        log.Fatalf("Conversion error: %v", err)
    }
    
    // Сериализуем в JSON для просмотра
    jsonData, err := json.MarshalIndent(ast, "", "  ")
    if err != nil {
        log.Fatalf("JSON error: %v", err)
    }
    
    fmt.Println(string(jsonData))
}
```

## Структура AST

### Program (корневой узел)

```go
type Program struct {
    Declarations []interface{} // []Stmt содержит VariableDecl и FunctionDecl
}
```

**JSON пример**:
```json
{
    "declarations": [
        {
            "type": "FunctionDecl",
            "name": "add",
            "parameters": [...],
            "body": {...},
            "location": {...}
        }
    ]
}
```

### VariableDecl (объявление переменной)

```go
type VariableDecl struct {
    Name        string
    Type        *Type
    Initializer interfaces.Expr  // может быть nil
    Location    Location
}
```

**Примеры**:
```c
int x;              // Без инициализации
int y = 10;         // С инициализацией
int arr[5];         // Массив
int *ptr;           // Указатель
```

**JSON**:
```json
{
    "type": "VariableDecl",
    "name": "x",
    "varType": {
        "baseType": "int",
        "pointerLevel": 0,
        "arraySize": null
    },
    "initializer": null,
    "location": {
        "line": 1,
        "column": 0,
        "endLine": 1,
        "endColumn": 6
    }
}
```

### FunctionDecl (определение функции)

```go
type FunctionDecl struct {
    Name       string
    ReturnType *Type
    Parameters []*Parameter
    Body       Stmt
    Location   Location
}

type Parameter struct {
    Name string
    Type *Type
}
```

**Пример**:
```c
int calculate(int a, int b) {
    return a + b;
}
```

**JSON**:
```json
{
    "type": "FunctionDecl",
    "name": "calculate",
    "returnType": {
        "baseType": "int",
        "pointerLevel": 0,
        "arraySize": null
    },
    "parameters": [
        {
            "name": "a",
            "type": {
                "baseType": "int",
                "pointerLevel": 0,
                "arraySize": null
            }
        },
        {
            "name": "b",
            "type": {
                "baseType": "int",
                "pointerLevel": 0,
                "arraySize": null
            }
        }
    ],
    "body": {...},
    "location": {...}
}
```

### IfStmt (условный оператор)

```go
type IfStmt struct {
    Condition   Expr
    ThenBlock   Stmt
    ElseIfList  []ElseIfClause  // NEW: явное представление else if
    ElseBlock   Stmt            // может быть nil
    Location    Location
}

type ElseIfClause struct {
    Condition Expr
    Block     Stmt
    Location  Location
}
```

**Пример**:
```c
if (x > 10) {
    printf("big");
} else if (x > 5) {
    printf("medium");
} else if (x > 0) {
    printf("small");
} else {
    printf("negative");
}
```

**JSON**:
```json
{
    "type": "IfStmt",
    "condition": {
        "type": "BinaryExpr",
        "operator": ">",
        "left": {"type": "Identifier", "name": "x"},
        "right": {"type": "IntLiteral", "value": 10}
    },
    "thenBlock": {...},
    "elseIf": [
        {
            "condition": {
                "type": "BinaryExpr",
                "operator": ">",
                "left": {"type": "Identifier", "name": "x"},
                "right": {"type": "IntLiteral", "value": 5}
            },
            "block": {...}
        },
        {
            "condition": {
                "type": "BinaryExpr",
                "operator": ">",
                "left": {"type": "Identifier", "name": "x"},
                "right": {"type": "IntLiteral", "value": 0}
            },
            "block": {...}
        }
    ],
    "elseBlock": {...},
    "location": {...}
}
```

### WhileStmt (цикл while)

```go
type WhileStmt struct {
    Condition Expr
    Body      Stmt
    Location  Location
}
```

**Пример**:
```c
while (i < 10) {
    i = i + 1;
}
```

### ForStmt (цикл for)

```go
type ForStmt struct {
    Init      Expr  // может быть nil
    Condition Expr  // может быть nil
    Update    Expr  // может быть nil
    Body      Stmt
    Location  Location
}
```

**Пример**:
```c
for (i = 0; i < 10; i = i + 1) {
    printf("i");
}
```

### ReturnStmt (оператор return)

```go
type ReturnStmt struct {
    Value    Expr  // может быть nil для return без значения
    Location Location
}
```

### BlockStmt (блок операторов)

```go
type BlockStmt struct {
    Statements []Stmt
    Location   Location
}
```

### Выражения

#### BinaryExpr (бинарное выражение)

```go
type BinaryExpr struct {
    Operator string       // "+", "-", "*", "/", "%", "==", "!=", "<", ">", "<=", ">=", "&&", "||"
    Left     Expr
    Right    Expr
    Location Location
}
```

#### UnaryExpr (унарное выражение)

```go
type UnaryExpr struct {
    Operator string  // "-", "!", "*", "&"
    Operand  Expr
    Location Location
}
```

**Примеры**:
```c
int x = -5;        // Унарный минус
int *ptr = &x;     // Адрес переменной (&)
int val = *ptr;    // Разыменование (*)
if (!condition) {} // Логическое отрицание (!)
```

#### AssignmentExpr (присваивание)

```go
type AssignmentExpr struct {
    Target   Expr  // Левая часть (переменная или доступ к массиву)
    Value    Expr  // Правая часть
    Location Location
}
```

#### CallExpr (вызов функции)

```go
type CallExpr struct {
    Function  Expr      // Identifier функции
    Arguments []Expr
    Location  Location
}
```

**Пример**:
```c
result = factorial(5);
```

#### ArrayAccessExpr (доступ к элементу массива)

```go
type ArrayAccessExpr struct {
    Array    Expr  // Выражение для массива
    Index    Expr  // Индекс
    Location Location
}
```

**Пример**:
```c
int val = arr[5];
arr[i] = 10;
```

#### ArrayInitExpr (инициализация массива)

```go
type ArrayInitExpr struct {
    Elements []Expr
    Location Location
}
```

**Пример**:
```c
int arr[3] = {1, 2, 3};
```

## Полные примеры

### Пример 1: Факториал (рекурсия)

```c
int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

int main() {
    int result = factorial(5);
    return result;
}
```

Демонстрирует:
- Объявление функции с параметром
- Условный оператор (if)
- Вызов функции (рекурсия)
- Бинарные выражения
- Return с значением

### Пример 2: Массивы и указатели

```c
int sum_array(int *arr, int size) {
    int sum = 0;
    int i = 0;
    while (i < size) {
        sum = sum + arr[i];
        i = i + 1;
    }
    return sum;
}

int main() {
    int data[5] = {1, 2, 3, 4, 5};
    int total = sum_array(data, 5);
    return total;
}
```

Демонстрирует:
- Объявление и инициализацию массива
- Указатели на int
- Доступ к элементам массива через указатель
- Цикл while
- Проход по массиву

### Пример 3: Else if конструкции

```c
int get_grade(int score) {
    if (score >= 90) {
        return 5;
    } else if (score >= 80) {
        return 4;
    } else if (score >= 70) {
        return 3;
    } else if (score >= 60) {
        return 2;
    } else {
        return 1;
    }
}
```

Демонстрирует:
- Цепочка else if блоков
- Явное представление в ElseIfList
- Финальный else блок
- Сравнительные операции

## Отладка

### Просмотр CST структуры

Для понимания как tree-sitter парсит код:

```bash
go run cmd/debug-cst/main.go
```

Выведет всю структуру дерева с типами узлов.

### Просмотр AST в JSON

Все примеры выводят JSON:

```bash
go run cmd/example/main.go > output.json
# Затем откройте output.json в редакторе или визуализаторе JSON
```

### Проверка структуры узла

Для проверки что поле существует в JSON:

```bash
go run cmd/example/main.go | jq '.declarations[0]'
```

## API справка

### NewCConverter() *CConverter

Создает новый экземпляр конвертера.

```go
conv := converter.NewCConverter()
```

### Parse(sourceCode []byte) (*sitter.Tree, error)

Парсит исходный код C в CST.

```go
tree, err := conv.Parse(sourceCode)
if err != nil {
    log.Fatal(err)
}
```

### ConvertToProgram(tree *sitter.Tree, sourceCode []byte) (*Program, error)

Конвертирует CST в AST.

```go
ast, err := conv.ConvertToProgram(tree, sourceCode)
if err != nil {
    log.Fatal(err)
}
```

### JSON сериализация

Все AST узлы автоматически сериализуются в JSON:

```go
jsonData, err := json.MarshalIndent(ast, "", "  ")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(jsonData))
```

## Ограничения

1. **Только int тип**: Не поддерживаются float, double, char и другие типы
2. **Указатели и массивы**: Только для int
3. **Функции**: Только определения в глобальной области
4. **Операторы**: Не поддерживаются switch, do-while, goto
5. **Препроцессор**: Не парсятся #include, #define и т.д.
6. **Комментарии**: Игнорируются (не включаются в AST)

## Частые вопросы

### Как добавить поддержку нового типа оператора?

1. Создайте struct в `internal/domain/structs/ast.go`
2. Реализуйте интерфейс `Stmt` (добавьте метод `StmtNode()`)
3. Добавьте case в `ConvertStmt()` методе
4. Реализуйте `convertXxxStatement()` метод

### Как добавить поддержку новой операции?

1. Для бинарной операции добавьте оператор в список в `convertBinaryExpression()`
2. Для унарной - в `convertUnaryExpression()`

### Почему вся информация о типе в одном объекте Type?

Потому что система типов упрощена - нужно только представить int, int*, int**, int[N], int*[N] и т.д.

### Как работает Location?

Location содержит позицию узла в исходном коде (Line, Column, EndLine, EndColumn).
Это нужно для:
- Отладки (знать где в коде произошла ошибка)
- Визуализации (подсвечивать в редакторе)
- Генерации сообщений об ошибках

## Производительность

Конвертер эффективен для образовательных целей:
- Быстрое парсинг (tree-sitter оптимизирован)
- Быстрая конвертация (линейный обход дерева)
- Малая память (структуры не содержат лишних данных)

Для больших файлов (>10K строк) рекомендуется асинхронная обработка.

## Лицензия

Проект лицензирован под MIT.

## Помощь и вопросы

Для вопросов и предложений см. репозиторий GitHub.

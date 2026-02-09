# Архитектура CST-to-AST конвертера

## Обзор

Проект преобразует Concrete Syntax Tree (CST), полученное от парсера tree-sitter, в упрощённое Abstract Syntax Tree (AST), подходящее для интерпретатора или визуализации программ на языке C.

## Структура пакетов

### `internal/domain/interfaces`
Определяет интерфейсы системы:

- **Node** - базовый интерфейс для всех узлов AST
- **Stmt** - интерфейс для операторов (statement nodes)
- **Expr** - интерфейс для выражений (expression nodes)
- **Converter** - интерфейс для конвертации CST в AST

Каждый узел AST реализует маркерные методы `StmtNode()` или `ExprNode()`, что позволяет компилятору Go проверить соответствие интерфейсам.

### `internal/domain/structs`
Содержит конкретные структуры для всех типов узлов AST:

#### Типы данных
- **Type** - представляет тип переменной (базовый тип, уровень указателей, размер массива)
- **Location** - информация о позиции узла в исходном коде
- **Parameter** - параметр функции

#### Statements
- **VariableDecl** - объявление переменной
- **FunctionDecl** - объявление/определение функции
- **IfStmt** - условный оператор if/else
- **WhileStmt** - цикл while
- **ForStmt** - цикл for
- **ReturnStmt** - оператор return
- **BlockStmt** - блок операторов в фигурных скобках
- **ExprStmt** - выражение как оператор
- **BreakStmt** - оператор break
- **ContinueStmt** - оператор continue

#### Expressions
- **Identifier** - имя переменной/функции
- **IntLiteral** - целочисленный литерал
- **BinaryExpr** - бинарное выражение (a + b, a == b, и т.д.)
- **UnaryExpr** - унарное выражение (-a, !a, *ptr, &var, и т.д.)
- **AssignmentExpr** - присваивание (a = b)
- **CallExpr** - вызов функции (func(args))
- **ArrayAccessExpr** - доступ к элементу массива (arr[i])
- **ArrayInitExpr** - инициализатор массива ({1, 2, 3})

#### Корневой узел
- **Program** - корневой узел программы, содержит список деклараций

### `internal/converter`
Основная логика конвертации:

- **CConverter** - основной конвертер, реализует интерфейс Converter
- **NewCConverter()** - создаёт новый конвертер с инициализированным парсером
- **Parse(sourceCode)** - парсит C код в CST (tree-sitter)
- **ConvertToProgram()** - преобразует корневой узел CST в AST Program
- **ConvertStmt()** - преобразует узел CST в Stmt
- **ConvertExpr()** - преобразует узел CST в Expr

Каждый метод содержит специализированные функции для каждого типа узла:
- `convertDeclaration()`
- `convertFunctionDefinition()`
- `convertIfStatement()`
- `convertWhileStatement()`
- `convertForStatement()`
- `convertReturnStatement()`
- `convertBlockStatement()`
- и так далее...

## Процесс конвертации

```
C исходный код
       ↓
  tree-sitter Parser
       ↓
    CST (tree-sitter)
       ↓
   CConverter
       ↓
   AST (наше представление)
       ↓
     JSON
```

## Примеры использования

### Простой пример (cmd/example)
```c
int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}

int main() {
    int x = 5;
    int result = factorial(x);
    return result;
}
```

Демонстрирует:
- Объявление функций с параметрами
- Условные операторы (if/else)
- Возвращаемые значения
- Вызовы функций
- Бинарные выражения

### Расширенный пример (cmd/advanced-example)
```c
int sum_array(int *arr, int size) {
    int total = 0;
    for (int i = 0; i < size; i++) {
        total = total + arr[i];
    }
    return total;
}

int main() {
    int numbers[5];
    int i = 0;
    
    while (i < 5) {
        numbers[i] = i * 2;
        i = i + 1;
    }
    
    int *ptr = &numbers[0];
    int result = sum_array(ptr, 5);
    
    return result;
}
```

Демонстрирует:
- Объявление массивов
- Указатели (int*)
- Циклы while и for
- Доступ к элементам массива (arr[i])
- Унарные операции (&, *)
- Обновляющие выражения (i++)

## Поддерживаемые типы узлов tree-sitter

### Statements
- `declaration` → VariableDecl
- `function_definition` → FunctionDecl
- `if_statement` → IfStmt
- `while_statement` → WhileStmt
- `for_statement` → ForStmt
- `return_statement` → ReturnStmt
- `compound_statement` → BlockStmt
- `expression_statement` → ExprStmt
- `break_statement` → BreakStmt
- `continue_statement` → ContinueStmt

### Expressions
- `identifier` → Identifier
- `number_literal` → IntLiteral
- `binary_expression` → BinaryExpr
- `unary_expression` → UnaryExpr
- `update_expression` → UnaryExpr (i++, i--)
- `assignment_expression` → AssignmentExpr
- `call_expression` → CallExpr
- `subscript_expression` → ArrayAccessExpr
- `initializer_list` → ArrayInitExpr
- `pointer_expression` → UnaryExpr
- `parenthesized_expression` → распаковывается

## Ограничения и будущие расширения

### Текущие ограничения
- Только тип `int` (может быть расширено на float, double, char)
- Нет поддержки структур и объединений
- Нет поддержки typedef
- Нет поддержки препроцессора (#include, #define)
- Нет поддержки многомерных массивов сверх одного уровня
- Нет обработки комментариев в AST

### Возможные расширения
- Поддержка дополнительных типов данных
- Структуры и объединения
- Перечисления
- Глобальные переменные с видимостью
- Статические переменные
- Функции с переменным числом аргументов
- Сохранение комментариев
- Оптимизация AST (постоянное сворачивание выражений)

## JSON представление AST

AST сериализуется в JSON следующим образом:

```json
{
  "declarations": [
    {
      "type": "FunctionDecl",
      "name": "main",
      "parameters": [],
      "body": {
        "type": "BlockStmt",
        "statements": [
          {
            "type": "VariableDecl",
            "name": "x",
            "type": {
              "baseType": "int",
              "pointerLevel": 0,
              "arraySize": 0
            },
            "initExpr": {
              "type": "IntLiteral",
              "value": 42
            }
          }
        ]
      }
    }
  ]
}
```

## Обработка ошибок

Конвертер возвращает подробные ошибки с информацией о строке и типе узла, на котором произошла ошибка:

```go
ast, err := conv.ConvertToProgram(tree, sourceCode)
if err != nil {
    log.Fatalf("Ошибка конвертации: %v", err)
}
```

## Производительность

- Парсинг: O(n), где n - размер исходного кода
- Конвертация: O(m), где m - количество узлов в CST
- Общая сложность: O(n)

Библиотека tree-sitter использует инкрементальное парсинга, что позволяет эффективно обновлять AST при изменении исходного кода.

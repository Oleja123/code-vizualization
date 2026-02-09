# code-vizualization
Разработка программных модулей в приложении для обучения основам программирования на языках Go, Java, JavaScript (Vue)

## cst-to-ast-service

Сервис для преобразования CST (Concrete Syntax Tree) в AST (Abstract Syntax Tree) для языка C.

### Описание

Проект использует библиотеку [go-tree-sitter](https://github.com/smacker/go-tree-sitter) для парсинга исходного кода C в CST, 
а затем преобразует его в упрощенное AST, подходящее для использования в интерпретаторе или визуализации структуры программы.

### Поддерживаемые конструкции C

- **Типы данных**: `int` с поддержкой указателей (`int*`, `int**`) и массивов (`int[N]`)
- **Переменные**: объявление с инициализацией и без
- **Функции**: объявление и определение с параметрами
- **Выражения**: 
  - Литералы (целые числа)
  - Идентификаторы
  - Бинарные операции (+, -, *, /, %, ==, !=, <, >, <=, >=, &&, ||)
  - Унарные операции (-, !, *, &)
  - Присваивание
  - Вызов функций
  - Доступ к элементам массива
  - Инициализация массивов
- **Операторы**:
  - if/else if/else (с поддержкой явного представления else if через ElseIfClause)
  - while
  - for
  - return
  - break/continue
  - Блоки операторов

### Структура проекта

```
cst-to-ast-service/
├── cmd/
│   ├── example/          # Простой пример использования
│   └── advanced-example/ # Пример с массивами, указателями, циклами
├── internal/
│   ├── converter/        # Логика конвертации CST -> AST
│   └── domain/
│       ├── interfaces/   # Интерфейсы Node, Stmt, Expr, Converter
│       └── structs/      # Структуры AST узлов
└── go.mod
```

### Использование

```go
import "github.com/Oleja123/code-vizualization/cst-to-ast-service/internal/converter"

// Исходный код C
sourceCode := []byte(`
int factorial(int n) {
    if (n <= 1) {
        return 1;
    }
    return n * factorial(n - 1);
}
`)

// Создаем конвертер
conv := converter.NewCConverter()

// Парсим в CST
tree, err := conv.Parse(sourceCode)

// Конвертируем в AST
ast, err := conv.ConvertToProgram(tree, sourceCode)
```

### Запуск примеров

```bash
# Простой пример
go run cmd/example/main.go

# Расширенный пример с массивами и указателями
go run cmd/advanced-example/main.go

# Пример с else if конструкциями
go run cmd/else-if-example/main.go
```

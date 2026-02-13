# Semantic Analyzer Service

Микросервис для семантической валидации AST кода на C, генерируемого `cst-to-ast-service`.

## Функции

✅ **Валидация типов переменных и параметров**: Только `int`
✅ **Проверка операторов присваивания**: `=`, `+=`, `-=`, `*=`, `%=`, `/=`
✅ **Проверка унарных операторов**: `-`, `+`, `++`, `--`, `!` (префиксные и постфиксные)
✅ **Проверка бинарных операторов**: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||`
✅ **Валидация типов функций**: Возвращаемый тип только `void` или `int`
✅ **Проверка параметров функций**: Только тип `int`
✅ **Детальные ошибки**: С указанием строки, колонки и типа ошибки

## Структура проекта

```
semantic-analyzer-service/
├── cmd/
│   └── example/
│       └── main.go              # CLI пример использования
├── internal/
│   └── domain/
│       ├── interfaces/
│       │   └── validator.go     # Интерфейсы валидатора
│       └── structs/
│           └── errors.go        # Структуры ошибок
├── pkg/
│   └── validator/
│       ├── validator.go         # Реализация валидатора
│       └── validator_test.go    # Тесты
├── go.mod
└── README.md
```

## Использование как библиотеки

```go
import (
    "github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
    "github.com/Oleja123/code-vizualization/semantic-analyzer-service/pkg/validator"
)

// Парсим код в AST
conv := converter.NewCConverter()
tree, err := conv.Parse(sourceCode)
program, err := conv.ConvertToProgram(tree, sourceCode)

// Проверяем семантику
val := validator.New()
if err := val.ValidateProgram(program); err != nil {
    log.Printf("Semantic error: %v", err)
}
```

## CLI пример

```bash
# Валидация файла
go run ./cmd/example/main.go -file code.c

# Сохранение результата
go run ./cmd/example/main.go -file code.c -out result.json

# Запуск на примере
go run ./cmd/example/main.go
```

## Тесты

```bash
go test ./pkg/validator -v
```

## Ошибки

Валидатор возвращает `SemanticError` со следующими кодами:

- `INVALID_VAR_TYPE` - неправильный тип переменной
- `INVALID_PARAM_TYPE` - неправильный тип параметра
- `INVALID_RETURN_TYPE` - неправильный возвращаемый тип
- `UNSUPPORTED_ASSIGN_OP` - неподдерживаемый оператор присваивания
- `UNSUPPORTED_UNARY_OP` - неподдерживаемый унарный оператор
- `UNSUPPORTED_BINARY_OP` - неподдерживаемый бинарный оператор
- `INVALID_FUNCTION_CALL` - ошибка вызова функции
- `SEMANTIC_ERROR` - другая семантическая ошибка

## Зависимости

- `cst-to-ast-service` (как replace в go.mod)
- `github.com/smacker/go-tree-sitter`

## Автор

Code Visualization Project

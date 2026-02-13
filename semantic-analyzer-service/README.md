# Semantic Analyzer Service

HTTP-сервис для семантической валидации C кода с опциональной проверкой компиляции через OneCompiler API.

## Возможности

✅ **HTTP REST API**: Готовый к продакшену веб-сервис
✅ **Валидация типов переменных и параметров**: Только `int`
✅ **Проверка операторов присваивания**: `=`, `+=`, `-=`, `*=`, `%=`, `/=`
✅ **Проверка унарных операторов**: `-`, `+`, `++`, `--`, `!` (префиксные и постфиксные)
✅ **Проверка бинарных операторов**: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||`
✅ **Валидация типов функций**: Возвращаемый тип только `void` или `int`
✅ **Проверка параметров функций**: Только тип `int`
✅ **Детальные ошибки**: С указанием строки, колонки и типа ошибки
✅ **Проверка компиляции**: Опциональная интеграция с OneCompiler API
✅ **Структурированное логирование**: JSON-формат с помощью slog
✅ **Гибкая конфигурация**: YAML + переменные окружения + флаги командной строки

## Структура проекта

```
semantic-analyzer-service/
├── cmd/
│   └── server/
│       └── main.go              # HTTP сервер
├── internal/
│   ├── domain/
│   │   └── interfaces/
│   │       └── validator.go     # Интерфейсы валидатора
│   └── infrastructure/
│       ├── config/
│       │   └── config.go        # Управление конфигурацией
│       └── onecompiler/
│           └── client.go        # Клиент для OneCompiler API
├── pkg/
│   └── validator/
│       ├── validator.go         # Реализация валидатора
│       ├── validator_test.go    # Тесты
│       └── errors.go            # Типы ошибок
├── config.yaml                  # Конфигурация по умолчанию
├── HTTP_API.md                  # Документация HTTP API
├── go.mod
└── README.md
```

## Быстрый старт

### HTTP Сервер

```bash
# Запуск с конфигурацией по умолчанию
go run ./cmd/server/main.go

# Сервер слушает на :8080
# Проверить статус:
curl http://localhost:8080/health

# Отправить код на валидацию:
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"int main() { return 0; }"}'
```

### Конфигурация

Создайте `config.yaml`:

```yaml
server:
  port: 8080

onecompiler:
  api_url: "https://api.onecompiler.com/api/v1"
  api_key: ""  # или через ONECOMPILER_API_KEY
  enabled: true
  timeout_seconds: 10
```

Запуск с кастомной конфигурацией:

```bash
# С указанием файла конфигурации
./semantic-analyzer-service -config config.yaml

# С переопределением порта
./semantic-analyzer-service -port 9000

# С API ключом OneCompiler через переменную окружения
ONECOMPILER_API_KEY="your-key" ./semantic-analyzer-service
```

### Использование как библиотеки

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

## HTTP API

Сервис предоставляет REST API для валидации C кода:

### Endpoints

- **POST /validate** - Валидирует C код
- **GET /health** - Проверка статуса сервиса
- **GET /info** - Информация о сервисе и поддерживаемых функциях

Подробная документация: [`HTTP_API.md`](HTTP_API.md)

### Примеры запросов

```bash
# Валидный код
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"int add(int a, int b) { return a + b; }"}'

# Код с ошибкой типа
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"float x = 3.14;"}'

# Проверка здоровья
curl http://localhost:8080/health

# Информация о сервисе
curl http://localhost:8080/info
```

## Тесты

```bash
go test ./pkg/validator -v
```

## Ошибки

Валидатор возвращает `SemanticError` со следующими кодами:

- `INVALID_TYPE` - неправильный тип переменной/параметра/возврата
- `UNSUPPORTED_ASSIGN_OP` - неподдерживаемый оператор присваивания
- `UNSUPPORTED_UNARY_OP` - неподдерживаемый унарный оператор
- `UNSUPPORTED_BINARY_OP` - неподдерживаемый бинарный оператор
- `INVALID_FUNCTION_CALL` - ошибка вызова функции
- `SEMANTIC_ERROR` - другая семантическая ошибка
- `UNKNOWN_STMT` - неизвестный тип оператора
- `UNKNOWN_EXPR` - неизвестный тип выражения

## Возможности OneCompiler

Когда `onecompiler.enabled: true`, сервис дополнительно проверяет код через компиляцию:

1. Выполняется семантическая валидация
2. Если семантика корректна, код отправляется на OneCompiler API
3. Если компиляция не удаётся, возвращается ошибка компиляции

Это позволяет обнаружить ошибки, которые не покрыты семантическим анализом.

## Логирование

Сервис использует структурированное логирование (slog) в формате JSON:

```json
{"time":"2026-02-13T12:00:00Z","level":"INFO","msg":"Starting Semantic Analyzer Server","address":":8080"}
{"time":"2026-02-13T12:00:01Z","level":"INFO","msg":"OneCompiler client initialized","timeout_seconds":10}
```

## Зависимости

- Go 1.21+
- `cst-to-ast-service` (как replace в go.mod)
- `github.com/smacker/go-tree-sitter`
- `gopkg.in/yaml.v2` (для конфигурации)

## Автор

Code Visualization Project

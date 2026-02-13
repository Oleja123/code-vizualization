# Code Visualization

Приложение для обучения основам программирования с визуализацией выполнения кода.

## 📦 Сервисы

### 1. CST-to-AST Service

Сервис парсинга и конвертации C кода в Abstract Syntax Tree с REST API.

**Технологии:** Go, tree-sitter, HTTP REST API

### 2. Semantic Analyzer Service

HTTP-сервис семантической валидации C кода с опциональной проверкой компиляции.

**Технологии:** Go, slog, OneCompiler API, YAML конфигурация

## CST-to-AST Service

### Возможности

✅ **Полная поддержка подмножества C**:
- Переменные: `int`, указатели (`int*`, `int**`), массивы (`int arr[10]`, `int *arr[5]`)
- Функции с параметрами и возвратом значений
- Управляющие конструкции: `if`/`else if`/`else`, `while`, `do while`, `for`, `break`, `continue`, `return`
- Переходы: `goto`, меченые операторы (labeled statements)
- Выражения: арифметика, логика, битовые операции, вызовы функций, доступ к массивам
- Присваивание: `=`, `+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=`
- Комментарии: автоматическая фильтрация `//` и `/* */`

✅ **Оптимизировано для интерпретаторов**:
- Else if представлен как else с вложенным if (как в C)
- Полная информация о позиции в коде (`Location`) для каждого узла
- Упрощенная система типов (только `int` с модификаторами)
- JSON-сериализация для межсервисного взаимодействия

✅ **Высокое качество**:
- 85 unit-тестов (все проходят ✅)
- Валидация ошибок: синтаксис, неподдерживаемые операторы, некорректные lvalue
- Готово к production

### Поддерживаемые конструкции

**Операторы** (15 типов):
- `Program`, `VariableDecl`, `FunctionDecl`
- `IfStmt` (else if представлен как вложенный if), `WhileStmt`, `DoWhileStmt`, `ForStmt`
- `ReturnStmt`, `BlockStmt`, `ExprStmt`
- `BreakStmt`, `ContinueStmt`, `GotoStmt`, `LabelStmt`

**Выражения** (8 типов):
- `VariableExpr`, `IntLiteral`
- `BinaryExpr` (20+ операторов: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `>`, `<=`, `>=`, `&&`, `||`, `&`, `|`, `^`, `<<`, `>>`)
- `UnaryExpr` (`-`, `!`, `*`, `&`, `++`, `--`)
- `AssignmentExpr` (10 операторов: `=`, `+=`, `-=`, `*=`, `/=`, `%=`, `&=`, `|=`, `^=`, `<<=`, `>>=`)
- `CallExpr`, `ArrayAccessExpr`, `ArrayInitExpr`

### Быстрый старт

**Как библиотека:**

```go
import "github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"

// Парсинг C кода в AST
conv := converter.New()
program, err := conv.ParseToAST(`
    int factorial(int n) {
        if (n <= 1) return 1;
        return n * factorial(n - 1);
    }
`)

if err != nil {
    fmt.Printf("Parse error at %d:%d\n", err.GetLocation().Line, err.GetLocation().Column)
    fmt.Printf("Code: %s\n", err.GetCode())
    fmt.Printf("Message: %s\n", err.GetMessage())
    return
}

// program имеет тип *converter.Program
for _, decl := range program.Declarations {
    if fn, ok := decl.(*converter.FunctionDecl); ok {
        fmt.Printf("Function: %s\n", fn.Name)
    }
}
```

**Как HTTP API:**

```bash
# Запустить сервер
go run cmd/server/main.go

# В другом терминале отправить код на парсинг
curl -X POST http://localhost:8080/parse \
  -H "Content-Type: application/json" \
  -d '{"code":"int factorial(int n) { if (n <= 1) return 1; return n * factorial(n - 1); }"}'
```

### Обработка ошибок

Все ошибки парсинга имеют тип `*converter.ConverterError` с полной информацией:

```go
if err != nil {
    convErr := err.(*converter.ConverterError)
    
    // Location в коде
    loc := convErr.GetLocation()
    fmt.Printf("at %d:%d\n", loc.Line, loc.Column)
    
    // Тип ошибки (ParseFailed, StmtConversion, ExprUnsupported, etc.)
    fmt.Printf("error code: %s\n", convErr.GetCode())
    
    // Понятное описание
    fmt.Printf("message: %s\n", convErr.GetMessage())
    
    // Тип узла tree-sitter (для отладки)
    fmt.Printf("node type: %s\n", convErr.GetNodeType())
}
```

### Структура проекта

```
cst-to-ast-service/
├── ARCHITECTURE.md       # 📖 Полный справочник AST для разработки интерпретаторов
├── API.md               # 📚 REST API документация
├── pkg/
│   └── converter/       # Публичный API
│       ├── converter.go      # Основной конвертер (ParseToAST)
│       ├── errors.go         # Структура ошибок (ConverterError)
│       ├── converter_test.go # 47 тестов (81.7% покрытие)
│       └── doc.go            # Документация пакета
├── cmd/
│   └── server/
│       └── main.go      # HTTP REST API сервер
├── internal/
│   └── domain/
│       ├── interfaces/  # Интерфейсы (Node, Stmt, Expr)
│       └── structs/     # Определения типов AST
└── go.mod
```

### REST HTTP API

AST-сервис предоставляет REST API для парсинга C кода.

**Запуск:**
```bash
go run cmd/server/main.go
```
Сервер доступен на `http://localhost:8080`

**Endpoints:**

- **POST /parse** — Парсит C код и возвращает AST или ошибку
- **GET /health** — Проверка статуса сервиса
- **GET /info** — Информация об API и поддерживаемых конструкциях

**Примеры:**
```bash
# Парсинг переменной
curl -X POST http://localhost:8080/parse \
  -H "Content-Type: application/json" \
  -d '{"code":"int x = 42;"}'

# Проверка здоровья
curl http://localhost:8080/health

# Информация об API
curl http://localhost:8080/info
```

Полная документация API: [`API.md`](cst-to-ast-service/API.md)

### Для разработчиков интерпретаторов

Читайте **[ARCHITECTURE.md](cst-to-ast-service/ARCHITECTURE.md)** — полный справочник с:
- Описанием всех 20 типов узлов AST
- Примерами интерпретации каждого узла
- Рекомендациями по архитектуре интерпретатора
- Примерами обхода и вычисления AST
- Информацией о системе типов и `Location`

Все типы AST доступны из пакета `converter`:

```go
import "github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"

// Основная точка входа
conv := converter.New()
program, err := conv.ParseToAST(sourceCode)

// Типы операторов
_ = (*converter.FunctionDecl)(nil)
_ = (*converter.IfStmt)(nil)
_ = (*converter.ForStmt)(nil)
// ... остальные 20 типов

// Базовые типы
_ = (*converter.Location)(nil)
_ = (*converter.Type)(nil)
```

### Ограничения

- Типы: только `int` (нет `float`, `char`, `struct`)
- Возвращаемые типы функций: `int`, `int*`, `int**`, `void`
- Строки и динамическая память (`malloc`/`free`) не поддерживаются
- Многомерные массивы: синтаксис `arr[i][j]` работает, объявление `int arr[2][3]` — нет
- Битовые операторы не поддерживаются (валидируются semantic-analyzer)

### Технологии

- Go 1.21+
- [go-tree-sitter](https://github.com/smacker/go-tree-sitter) для парсинга C
- Domain-Driven Design архитектура

---

## Semantic Analyzer Service

HTTP-сервис семантической валидации C кода с проверкой типов, операторов и опциональной компиляцией через OneCompiler API.

### Возможности

✅ **HTTP REST API**: Готовый к продакшену веб-сервис на порту 8080
✅ **Валидация типов**: Переменные и параметры только `int`, возврат `int` или `void`
✅ **Проверка операторов**:
- Присваивание: `=`, `+=`, `-=`, `/=`, `%=`
- Унарные: `-`, `!`, `++`, `--` (префикс/постфикс)
- Бинарные: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||`
- Отклонение битовых операторов: `&`, `|`, `^`, `<<`, `>>`

✅ **Проверка компиляции**: Опциональная интеграция с OneCompiler API
✅ **Детальные сообщения об ошибках**: С указанием линии, колонки и кода ошибки
✅ **Структурированное логирование**: JSON-формат с помощью slog
✅ **Гибкая конфигурация**: YAML + переменные окружения + флаги командной строки

### Структура проекта

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
│       ├── validator_test.go    # Unit тесты
│       └── errors.go            # Типы ошибок
├── config.yaml                  # Конфигурация по умолчанию
├── HTTP_API.md                  # Документация HTTP API
├── go.mod
└── README.md
```

### Быстрый старт

**Запуск HTTP сервера:**

```bash
cd semantic-analyzer-service

# Запуск с конфигурацией по умолчанию
go run ./cmd/server/main.go

# Сервер слушает на :8080
# Проверить статус:
curl http://localhost:8080/health

# Валидация кода:
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"int main() { return 0; }"}'
```

**Конфигурация (`config.yaml`):**

```yaml
server:
  port: 8080

onecompiler:
  api_url: "https://api.onecompiler.com/api/v1"
  api_key: ""  # или через ONECOMPILER_API_KEY
  enabled: true
  timeout_seconds: 10
```

**Запуск с кастомной конфигурацией:**

```bash
# С указанием файла конфигурации
./semantic-analyzer-service -config config.yaml

# С переопределением порта
./semantic-analyzer-service -port 9000

# С API ключом OneCompiler через переменную окружения
ONECOMPILER_API_KEY="your-key" ./semantic-analyzer-service
```

**HTTP API:**

- **POST /validate** — Валидирует C код и возвращает AST или ошибку
- **GET /health** — Проверка статуса сервиса
- **GET /info** — Информация об API и поддерживаемых конструкциях

Полная документация API: [`HTTP_API.md`](semantic-analyzer-service/HTTP_API.md)

**Как библиотека:**

```go
import (
    "github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
    "github.com/Oleja123/code-vizualization/semantic-analyzer-service/pkg/validator"
)

// Парсим
conv := converter.NewCConverter()
tree, _ := conv.Parse(sourceCode)
program, _ := conv.ConvertToProgram(tree, sourceCode)

// Валидируем
val := validator.New()
if err := val.ValidateProgram(program); err != nil {
    log.Fatalf("Semantic error: %v", err)
}
```

**Тестирование:**

```bash
cd semantic-analyzer-service
go test ./pkg/validator -v
```

**Примеры с curl:**

```bash
# Валидный код
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"int add(int a, int b) { return a + b; }"}'

# Ошибка типа
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"float x = 3.14;"}'

# Проверка здоровья
curl http://localhost:8080/health

# Информация о сервисе
curl http://localhost:8080/info
```

### Возможности OneCompiler

Когда `onecompiler.enabled: true`, сервис:

1. Выполняет семантическую валидацию
2. Если семантика корректна, отправляет код на OneCompiler API
3. Возвращает ошибку компиляции, если компиляция не удалась

Это позволяет обнаружить ошибки, не покрытые семантическим анализом.

### Логирование

Сервис использует структурированное логирование (slog) в формате JSON:

```json
{"time":"2026-02-13T12:00:00Z","level":"INFO","msg":"Starting Semantic Analyzer Server","address":":8080"}
{"time":"2026-02-13T12:00:01Z","level":"INFO","msg":"OneCompiler client initialized","timeout_seconds":10}
```

---

## Интеграция сервисов

Полный процесс парсинга и валидации:

```
C код → cst-to-ast-service (HTTP API) → AST → semantic-analyzer-service (HTTP API) → результат
```

**Пример интеграции:**

```bash
# 1. Запустить оба сервиса
cd cst-to-ast-service && go run cmd/server/main.go &
cd semantic-analyzer-service && go run cmd/server/main.go &

# 2. Парсинг кода
curl -X POST http://localhost:8080/parse \
  -H "Content-Type: application/json" \
  -d '{"code":"int factorial(int n) { if (n <= 1) return 1; return n * factorial(n - 1); }"}' \
  > ast.json

# 3. Валидация кода
curl -X POST http://localhost:8081/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"int factorial(int n) { if (n <= 1) return 1; return n * factorial(n - 1); }"}'
```

**Библиотечная интеграция:**

```go
import (
    "github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
    "github.com/Oleja123/code-vizualization/semantic-analyzer-service/pkg/validator"
)

// Парсинг
conv := converter.New()
program, err := conv.ParseToAST(code)
if err != nil {
    log.Fatalf("Parse error: %v", err)
}

// Валидация
val := validator.New()
if err := val.ValidateProgram(program); err != nil {
    log.Fatalf("Semantic error: %v", err)
}

// Успех - можно передавать в интерпретатор
```

### Технологии

- Go 1.21+
- [go-tree-sitter](https://github.com/smacker/go-tree-sitter) для парсинга C
- Domain-Driven Design архитектура

---

## Roadmap

- [ ] Interpreter Service (отдельный сервис для выполнения AST)
- [ ] Visualization Service (отрисовка состояния переменных, call stack)
- [ ] Web UI (Vue.js фронтенд)

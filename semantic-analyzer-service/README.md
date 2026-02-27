# semantic-analyzer-service

HTTP-сервис для валидации C-кода:

1. парсинг в AST (`cst-to-ast-service`),
2. семантическая проверка,
3. опциональная compile-check проверка через OneCompiler.

## Что делает сервис

- Валидирует AST, полученный из C-кода.
- Проверяет допустимые типы и операторы.
- При включённом OneCompiler дополнительно проверяет компилируемость.
- Возвращает AST при успешной валидации.

## HTTP endpoints

- `POST /validate` — валидация кода.
- `GET /health` — health-check.
- `GET /info` — информация о сервисе и поддерживаемых операторах.

Подробный контракт: [HTTP_API.md](HTTP_API.md).

## Быстрый запуск

```bash
cd semantic-analyzer-service
go run ./cmd/server/main.go
```

Сервис слушает порт из конфига (`server.port`, обычно `8080`).

Пример запроса:

```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"int main() { return 0; }"}'
```

## Конфигурация

Файл: `config.yaml`

```yaml
server:
  port: 8080

onecompiler:
  api_url: "https://api.onecompiler.com/api/v1"
  api_key: ""
  enabled: true
  timeout_seconds: 10
```

### Источники конфигурации

- `-config` — путь к YAML (по умолчанию `config.yaml`).
- `-port` — переопределяет `server.port`.
- `ONECOMPILER_API_KEY` — переопределяет `onecompiler.api_key`.

Если конфиг не загрузился, применяется fallback:

- `server.port = 8080`
- `onecompiler.api_url = https://api.onecompiler.com/api/v1`
- `onecompiler.enabled = true`
- `onecompiler.timeout_seconds = 10`

## Использование как библиотеки (`pkg/validator`)

```go
import (
    "github.com/Oleja123/code-vizualization/cst-to-ast-service/pkg/converter"
    "github.com/Oleja123/code-vizualization/semantic-analyzer-service/pkg/validator"
)

conv := converter.New()
program, err := conv.ParseToAST("int main(){ return 0; }")
if err != nil {
    panic(err)
}

val := validator.New()
if err := val.ValidateProgram(program); err != nil {
    panic(err)
}
```

С OneCompiler:

```go
import (
    "github.com/Oleja123/code-vizualization/semantic-analyzer-service/pkg/onecompiler"
    "github.com/Oleja123/code-vizualization/semantic-analyzer-service/pkg/validator"
)

client := onecompiler.NewClient("https://api.onecompiler.com/api/v1", "API_KEY", 10)
val := validator.NewWithOneCompilerClient(client)
err := val.ValidateProgram(program, sourceCode)
```

## Поддерживаемые правила (текущее состояние)

- Базовый тип: `int`.
- Возвращаемые типы функций: `int` и `void`.
- Максимальная размерность массива: `2`.
- Указатели запрещены (`pointerLevel` должен быть `0`).
- Массивы в параметрах и возвращаемом типе функций запрещены.

Операторы:

- Присваивание: `=`, `+=`, `-=`, `*=`, `%=` ,`/=`
- Унарные: `-`, `+`, `++`, `--`, `!`
- Бинарные: `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||`

## Коды ошибок валидатора

`pkg/validator/errors.go`:

- `INVALID_TYPE`
- `UNSUPPORTED_ASSIGN_OP`
- `UNSUPPORTED_UNARY_OP`
- `UNSUPPORTED_BINARY_OP`
- `INVALID_FUNCTION_CALL`
- `SEMANTIC_ERROR`
- `UNKNOWN_STMT`
- `UNKNOWN_EXPR`

Также используются ошибки compile-check:

- `compilation error: ...`
- `compilation check unavailable: ...`

## Тесты

```bash
go test ./pkg/validator -v
go test ./... -v
```

## Зависимости

- Go `1.23+`
- `cst-to-ast-service`
- `github.com/smacker/go-tree-sitter`
- `gopkg.in/yaml.v2`

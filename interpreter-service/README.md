# interpreter-service

HTTP-сервис для пошаговой визуализации выполнения C-кода.

Сервис:

1. парсит C-код в AST,
2. запускает семантическую проверку,
3. интерпретирует программу,
4. возвращает snapshot состояния на выбранном шаге.

## Что делает сервис

- Предоставляет endpoint `POST /snapshot`.
- Возвращает состояние call stack, scopes, переменных и текущей строки.
- Поддерживает rewind/forward по шагам через `eventdispatcher`.
- Ограничивает выполнение через лимиты:
  - `max_allocated_elements` (аллоцируемые элементы),
  - `max_steps` (число шагов интерпретации).

Подробный HTTP-контракт: [HTTP_API.md](HTTP_API.md).

## Быстрый запуск

```bash
cd interpreter-service
go run ./cmd/main.go
```

Сервис читает `config.yaml` по умолчанию.

Запуск с явным конфигом:

```bash
go run ./cmd/main.go -config ./config.yaml
```

## Конфигурация

Файл: `config.yaml`

```yaml
server:
  port: 8080

onecompiler:
  api_url: "https://api.onecompiler.com/api/v1"
  api_key: ""
  enabled: false
  timeout_seconds: 10

limitations:
  max_allocated_elements: 100
  max_steps: 1000
```

### Источники значений

- Флаг `-config` — путь к YAML (по умолчанию `config.yaml`).
- Если конфиг не читается, используется `LoadOrDefault` с дефолтами.

### Fallback defaults

- `server.port = 8080`
- `onecompiler.api_url = https://api.onecompiler.com/api/v1`
- `onecompiler.enabled = false`
- `onecompiler.timeout_seconds = 10`
- `limitations.max_allocated_elements = 100`
- `limitations.max_steps = 1000`

## Пример запроса

```bash
curl -X POST http://localhost:8080/snapshot \
  -H "Content-Type: application/json" \
  -d '{
    "code": "int main(){ int x = 1; return x; }",
    "step": 1
  }'
```

## Структура ответа (кратко)

- `success` — успешность запроса.
- `step` — шаг из запроса.
- `current_step` — применённый внешний шаг.
- `steps_count` — доступное число внешних шагов.
- `result` — `return` из `main` (если есть).
- `snapshot` — состояние runtime (stack/scopes/line/error).

## Коды ответов

- `200` — snapshot успешно возвращён.
- `400` — невалидный запрос, parse/semantic/interpreter ошибка, невалидный шаг.
- `405` — метод отличен от `POST`.
- `503` — OneCompiler compile-check недоступен (когда включён).

## Тесты

```bash
go test ./internal/handler -v
go test ./... -v
```

## Безопасность

- Не храните реальный `onecompiler.api_key` в репозитории.
- Используйте переменные окружения/секрет-хранилище для production-конфигураций.

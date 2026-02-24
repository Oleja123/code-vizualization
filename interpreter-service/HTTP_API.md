# Interpreter Service - HTTP API

## Обзор

`interpreter-service` предоставляет HTTP API для пошагового выполнения C-программы и получения снимка состояния интерпретатора на указанном шаге.

Сервис выполняет:
1. Парсинг C-кода в AST.
2. Семантическую валидацию (и опционально compile-check через OneCompiler).
3. Интерпретацию программы с построением шагов.
4. Возврат snapshot состояния по шагу `step`.

## Конфигурация и запуск

Сервис запускается из `cmd/main.go` и регистрирует endpoint:
- `POST /snapshot`

Флаги запуска:
- `-port` (по умолчанию `8080`) — HTTP порт.
- `-onecompiler-config` (по умолчанию `config.yaml`) — путь к YAML-конфигу OneCompiler.

Пример:

```bash
go run ./cmd/main.go -port 8080 -onecompiler-config ./config.yaml
```

## Endpoint

### POST /snapshot

Возвращает snapshot состояния выполнения программы для заданного шага.

### Request

`Content-Type: application/json`

```json
{
  "code": "int main() { int x = 1; return x; }",
  "step": 0
}
```

Поля:
- `code` (string, required) — исходный C-код.
- `step` (int, required) — неотрицательный индекс внешнего шага.

### Response (200 OK)

```json
{
  "success": true,
  "step": 0,
  "current_step": 0,
  "steps_count": 3,
  "result": 1,
  "snapshot": {
    "call_stack": {
      "frames": [
        {
          "func_name": "global",
          "scopes": [
            {
              "declarations": {
                "declarations": []
              }
            }
          ]
        },
        {
          "func_name": "main",
          "scopes": [
            {
              "declarations": {
                "declarations": []
              }
            },
            {
              "declarations": {
                "declarations": [
                  {
                    "name": "x",
                    "value": 1,
                    "step_changed": 1
                  }
                ]
              }
            }
          ]
        }
      ]
    },
    "global_scope": {
      "declarations": {
        "declarations": []
      }
    },
    "line": 2,
    "error": ""
  }
}
```

Поля верхнего уровня:
- `success` (bool) — признак успешного запроса.
- `step` (int) — шаг, который запрошен клиентом.
- `current_step` (int) — текущий применённый шаг (внешняя нумерация).
- `steps_count` (int) — количество доступных внешних шагов.
- `result` (int|null) — итоговый `return` из `main`, если вычислен.
- `snapshot` (object) — снимок runtime-состояния.

### Snapshot structure

`snapshot`:
- `call_stack.frames[]`
  - `func_name` (string)
  - `scopes[]`
    - `declarations.declarations[]` — массив деклараций в текущем scope.
      - Возможные элементы:
        1. Variable:
           - `name` (string)
           - `value` (int|null)
           - `step_changed` (int)
        2. Array:
           - `name` (string)
           - `size` (int)
           - `values[]`:
             - `value` (int|null)
             - `step_changed` (int)
        3. Array2D:
           - `name` (string)
           - `size1` (int)
           - `size2` (int)
           - `values[]` (массив строк, каждая как `Array`)
- `global_scope` — глобальный scope.
- `line` (int) — текущая строка выполнения.
- `error` (string) — runtime/undefined behavior ошибка в контексте snapshot (если была).

## Ошибки

Все ошибки возвращаются в формате:

```json
{
  "success": false,
  "error": "..."
}
```

Коды и причины:
- `400 Bad Request`
  - невалидный JSON body;
  - `code` пустой;
  - `step < 0`;
  - parse error;
  - semantic error;
  - ошибка интерпретации;
  - шаг вне диапазона.
- `405 Method Not Allowed`
  - используется метод, отличный от `POST`.
- `503 Service Unavailable`
  - compile-check включён, но OneCompiler недоступен (`CompileUnavailableError`).

## Примеры

### Успешный запрос

```bash
curl -X POST http://localhost:8080/snapshot \
  -H "Content-Type: application/json" \
  -d '{
    "code":"int main(){ int x=1; x=2; return x; }",
    "step": 2
  }'
```

### Ошибка валидации

```bash
curl -X POST http://localhost:8080/snapshot \
  -H "Content-Type: application/json" \
  -d '{
    "code":"float main(){ return 0; }",
    "step": 0
  }'
```

Пример ответа:

```json
{
  "success": false,
  "error": "semantic error: ..."
}
```

### Неверный шаг

```bash
curl -X POST http://localhost:8080/snapshot \
  -H "Content-Type: application/json" \
  -d '{
    "code":"int main(){ return 0; }",
    "step": 999
  }'
```

Пример ответа:

```json
{
  "success": false,
  "error": "step 999 out of range [0, N]"
}
```

## Примечания

- API использует **внешнюю нумерацию шагов** (`step`, `current_step`, `steps_count`) с учётом внутреннего `step_begin`.
- `result` может быть `null`, если вычисление завершилось ошибкой до возврата из `main`.
- Поле `error` внутри `snapshot` используется для отображения runtime/undefined behavior ошибок на шаге.
- Не храните реальные секреты (например, API-ключи) в репозитории; используйте переменные окружения/секрет-хранилище.

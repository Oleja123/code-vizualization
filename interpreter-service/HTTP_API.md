# interpreter-service HTTP API

## Base URL

По умолчанию:

```text
http://localhost:8080
```

Порт задаётся в `config.yaml` (`server.port`).

## Endpoint

### `POST /snapshot`

Выполняет C-программу, строит список шагов интерпретации и возвращает snapshot для запрошенного шага.

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
- `step` (int, required) — индекс шага (должен быть `>= 0`).

### Success — `200 OK`

```json
{
  "success": true,
  "step": 0,
  "current_step": 0,
  "steps_count": 3,
  "result": 1,
  "snapshot": {
    "call_stack": {
      "frames": []
    },
    "global_scope": {
      "declarations": {
        "declarations": []
      }
    },
    "line": 1,
    "error": ""
  }
}
```

Поля ответа:

- `success` (bool)
- `step` (int) — шаг из запроса.
- `current_step` (int) — реально применённый шаг (внешняя нумерация относительно начала `main`).
- `steps_count` (int) — количество доступных внешних шагов.
- `result` (*int, optional) — значение `return` из `main`, если вычислено.
- `snapshot` (object) — снимок runtime-состояния после применения шага.

### Error format

Все ошибки возвращаются в JSON:

```json
{
  "success": false,
  "error": "..."
}
```

### Errors

- `400 Bad Request`
  - невалидный JSON (`invalid request body: ...`),
  - неизвестные поля в JSON,
  - пустой `code` (`code is required`),
  - `step < 0` (`step must be non-negative`),
  - parse error,
  - semantic error,
  - ошибка выполнения интерпретатора,
  - шаг вне диапазона (`invalid step index: ...`).
- `405 Method Not Allowed`
  - метод отличен от `POST`.
- `503 Service Unavailable`
  - включён compile-check и OneCompiler недоступен (`compilation check unavailable: ...`).

## Snapshot model (кратко)

- `snapshot.call_stack.frames[]`
  - `func_name`
  - `scopes[]`
    - `declarations.declarations[]` — переменные/массивы/2D-массивы.
- `snapshot.global_scope` — глобальный scope.
- `snapshot.line` — текущая строка интерпретации.
- `snapshot.error` — runtime/undefined behavior ошибка на текущем состоянии (если есть).

`parent`-ссылки scope не сериализуются в JSON.

## Examples

### Success

```bash
curl -X POST http://localhost:8080/snapshot \
  -H "Content-Type: application/json" \
  -d '{
    "code": "int main(){ int x = 1; x = 2; return x; }",
    "step": 2
  }'
```

### Step out of range

```bash
curl -X POST http://localhost:8080/snapshot \
  -H "Content-Type: application/json" \
  -d '{
    "code": "int main(){ return 0; }",
    "step": 999
  }'
```

### Method not allowed

```bash
curl -X GET http://localhost:8080/snapshot
```

## Execution pipeline

`/snapshot` обрабатывается по этапам:

1. Parse C-кода в AST (`cst-to-ast-service/pkg/converter`).
2. Семантическая валидация (`semantic-analyzer-service/pkg/validator`).
3. Интерпретация (`internal/application/interpreter`) с лимитами:
   - `limitations.max_allocated_elements`,
   - `limitations.max_steps`.
4. Применение шагов в `eventdispatcher` и возврат snapshot.

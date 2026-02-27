# semantic-analyzer-service HTTP API

## Base URL

По умолчанию:

```text
http://localhost:8080
```

Порт задаётся через `config.yaml` (`server.port`) или флаг `-port`.

---

## Endpoints

### 1) `POST /validate`

Валидирует C-код.

#### Request

```json
{
  "code": "int main() { return 0; }"
}
```

#### Success — `200 OK`

```json
{
  "success": true,
  "program": {
    "type": "Program",
    "declarations": []
  }
}
```

`program` — AST из `cst-to-ast-service/pkg/converter`.

#### Validation/Parse/Conversion error — `400 Bad Request`

```json
{
  "success": false,
  "error": "..."
}
```

Возможные префиксы ошибок:

- `Invalid request body: ...`
- `Parse error: ...`
- `Conversion error: ...`
- ошибки `validator` (например `[INVALID_TYPE] ...`)
- `compilation error: ...` (если включён OneCompiler и код не компилируется)

#### Compile-check unavailable — `503 Service Unavailable`

```json
{
  "success": false,
  "error": "compilation check unavailable: ..."
}
```

#### Method not allowed — `405 Method Not Allowed`

Для `/validate` при методе, отличном от `POST`, сервер возвращает plain-text ответ:

```text
Method not allowed
```

---

### 2) `GET /health`

Проверка состояния сервиса.

#### `/info` response — `200 OK`

```json
{
  "status": "healthy",
  "service": "semantic-analyzer-service"
}
```

---

### 3) `GET /info`

Информация о сервисе и поддерживаемых операторах.

#### Response — `200 OK`

```json
{
  "service": "Semantic Analyzer Service",
  "version": "1.0.0",
  "endpoints": {
    "POST /validate": "Validate C code",
    "GET /health": "Health check",
    "GET /info": "Service information"
  },
  "supported_types": ["int", "void"],
  "supported_operators": {
    "assignment": ["=", "+=", "-=", "/=", "%="],
    "unary": ["-", "!", "++", "--"],
    "binary": ["+", "-", "*", "/", "%", "==", "!=", "<", "<=", ">", ">=", "&&", "||"]
  },
  "unsupported_operators": ["&", "|", "^", "<<", ">>"]
}
```

---

## Конфигурация

Формат `config.yaml`:

```yaml
server:
  port: 8080

onecompiler:
  api_url: "https://api.onecompiler.com/api/v1"
  api_key: ""
  enabled: true
  timeout_seconds: 10
```

### Переопределения

- Флаг `-config` — путь к конфигу.
- Флаг `-port` — переопределяет `server.port`.
- Переменная `ONECOMPILER_API_KEY` — переопределяет `onecompiler.api_key`.

### Fallback defaults (если конфиг не прочитан)

- `server.port = 8080`
- `onecompiler.api_url = https://api.onecompiler.com/api/v1`
- `onecompiler.enabled = true`
- `onecompiler.timeout_seconds = 10`

---

## Примеры

### Validate (ok)

```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"int main() { return 0; }"}'
```

### Validate (semantic error)

```bash
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"float x = 1.0;"}'
```

### Health

```bash
curl http://localhost:8080/health
```

### Info

```bash
curl http://localhost:8080/info
```

---

## Слои валидации

Запрос `/validate` проходит этапы:

1. Parse C-кода в CST (`converter.Parse`).
2. Convert CST → AST (`converter.ConvertToProgram`).
3. Семантическая проверка (`validator.ValidateProgram`).
4. Опционально: compile-check через OneCompiler (если включён).

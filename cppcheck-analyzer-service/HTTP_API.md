# cppcheck-analyzer-service HTTP API

## Base URL

По умолчанию:

```text
http://localhost:8086
```

Порт задаётся через `config.yaml` (`server.port`) или флаг `-port`.

## Endpoints

### 1) `POST /analyze`

Выполняет статический анализ C-кода через `cppcheck`.

#### Request

```json
{
  "code": "int main() { int x; return 0; }"
}
```

Поля:

- `code` (string, required) — исходный код на C.

#### Success — `200 OK`

```json
{
  "success": true,
  "passed": false,
  "count": 1,
  "analyzer": "cppcheck",
  "issues": [
    {
      "severity": "style",
      "id": "unreadVariable",
      "line": 1,
      "message": "Variable 'x' is assigned a value that is never used."
    }
  ]
}
```

`passed=true` означает, что замечаний не найдено (`issues` пуст).

Если найдено больше замечаний, чем `cppcheck.max_issues`, в `issues` возвращаются только первые `max_issues` элементов.

В `issues` возвращаются только сообщения, которые относятся к конкретным строкам исходного кода.

#### Bad request — `400 Bad Request`

```json
{
  "success": false,
  "error": "..."
}
```

Возможные причины:

- невалидный JSON (`invalid request body: ...`),
- пустой `code` (`code is required`),

#### Internal error — `500 Internal Server Error`

```json
{
  "success": false,
  "error": "failed to execute cppcheck: ..."
}
```

### 2) `GET /health`

Проверка состояния сервиса.

#### Response — `200 OK`

```json
{
  "status": "healthy",
  "service": "cppcheck-analyzer-service"
}
```

### 3) `GET /info`

Информация о сервисе.

#### Response — `200 OK`

```json
{
  "service": "Cppcheck Analyzer Service",
  "version": "1.0.0",
  "analyzer": "cppcheck",
  "language": "c",
  "endpoints": {
    "POST /analyze": "Analyze C code with cppcheck",
    "GET /health": "Health check",
    "GET /info": "Service information"
  }
}
```

## Примеры

## Конфигурация

```yaml
server:
  port: 8086

cppcheck:
  path: "cppcheck"
  std: "c11"
  enable: "warning,style,performance,portability,information"
  timeout_seconds: 15
  inconclusive: false
  max_issues: 100
```

Env override:

- `CPPCHECK_MAX_ISSUES`

### Анализ

```bash
curl -X POST http://localhost:8086/analyze \
  -H "Content-Type: application/json" \
  -d '{"code":"int main(){int x; return 0;}"}'
```

### Health

```bash
curl http://localhost:8086/health
```

### Info

```bash
curl http://localhost:8086/info
```

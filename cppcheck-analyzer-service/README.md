# cppcheck-analyzer-service

HTTP-сервис для статического анализа C-кода через `cppcheck`.

Сервис принимает исходный код, запускает `cppcheck` c режимом `--language=c` и возвращает список найденных замечаний.

## Что делает сервис

- Анализирует только C-код.
- Запускает `cppcheck` локально в изолированном временном файле.
- Возвращает структурированные замечания: severity, id, line, message.

## HTTP endpoints

- `POST /analyze` — анализ C-кода.
- `GET /health` — health-check.
- `GET /info` — информация о сервисе.

Подробный контракт: [HTTP_API.md](HTTP_API.md).

## Быстрый запуск

```bash
cd cppcheck-analyzer-service
go run ./cmd/server/main.go
```

Сервис слушает порт из конфига (`server.port`, по умолчанию `8086`).

Пример запроса:

```bash
curl -X POST http://localhost:8086/analyze \
  -H "Content-Type: application/json" \
  -d '{"code":"int main(){int x; return 0;}"}'
```

## Конфигурация

Файл: `config.yaml`

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

### Источники конфигурации

- `-config` — путь к YAML (по умолчанию `config.yaml`).
- `-port` — переопределяет `server.port`.
- `CPPCHECK_PATH` — переопределяет `cppcheck.path`.
- `CPPCHECK_STD` — переопределяет `cppcheck.std`.
- `CPPCHECK_ENABLE` — переопределяет `cppcheck.enable`.
- `CPPCHECK_TIMEOUT_SECONDS` — переопределяет `cppcheck.timeout_seconds`.
- `CPPCHECK_INCONCLUSIVE` — переопределяет `cppcheck.inconclusive`.
- `CPPCHECK_MAX_ISSUES` — переопределяет `cppcheck.max_issues`.

## Тесты и сборка

```bash
go test ./...
go build ./cmd/server/main.go
```

## Зависимости

- Go `1.25.6+`
- `cppcheck` в окружении запуска
- `gopkg.in/yaml.v2`

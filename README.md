# Code Visualization

Монорепозиторий сервиса визуализации пошагового выполнения C-кода для учебных сценариев.

## Состав проекта

- `cst-to-ast-service` — парсинг C-кода и конвертация в AST.
- `semantic-analyzer-service` — семантическая проверка AST и опциональный compile-check.
- `interpreter-service` — интерпретация программы и выдача snapshot по шагам.
- `frontend` — Vue UI для запуска кода и навигации по шагам выполнения.
- `designing` — проектная документация.

## Как работает система

1. Frontend отправляет код и номер шага.
2. `interpreter-service` получает `POST /snapshot`.
3. Код парсится через `cst-to-ast-service`.
4. AST валидируется через `semantic-analyzer-service`.
5. Интерпретатор выполняет программу, формирует шаги и возвращает snapshot выбранного шага.

## Быстрый запуск (локально)

### 1) interpreter-service

```bash
cd interpreter-service
go run ./cmd/main.go -config ./config.yaml
```

Сервис по умолчанию слушает `:8080` (из `server.port`).

### 2) frontend

```bash
cd frontend
npm install
npm run dev
```

Dev-сервер frontend: `http://localhost:3000`.

## Конфигурация

Основные параметры для визуализации находятся в `interpreter-service/config.yaml`:

- `server.port`
- `onecompiler.enabled`, `onecompiler.api_url`, `onecompiler.api_key`, `onecompiler.timeout_seconds`
- `limitations.max_allocated_elements`
- `limitations.max_steps`

## Документация по сервисам

- AST/конвертер: [cst-to-ast-service/ARCHITECTURE.md](cst-to-ast-service/ARCHITECTURE.md)
- semantic-analyzer API: [semantic-analyzer-service/HTTP_API.md](semantic-analyzer-service/HTTP_API.md)
- semantic-analyzer обзор: [semantic-analyzer-service/README.md](semantic-analyzer-service/README.md)
- interpreter API: [interpreter-service/HTTP_API.md](interpreter-service/HTTP_API.md)
- interpreter обзор: [interpreter-service/README.md](interpreter-service/README.md)

## Тесты

Запуск тестов по сервисам:

```bash
cd cst-to-ast-service && go test ./... -v
cd ../semantic-analyzer-service && go test ./... -v
cd ../interpreter-service && go test ./... -v
```

## Технологии

- Go (backend-сервисы)
- Vue 3 + Vite (frontend)
- tree-sitter (парсинг C)

## Статус

Проект в активной разработке; актуальные контракты и ограничения фиксируются в документации каждого сервиса.

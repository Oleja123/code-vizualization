# Code Visualization Frontend

Vue 3 приложение для:

- пошаговой трассировки выполнения C-кода,
- визуализации runtime-состояния,
- генерации блок-схем,
- статического анализа C-кода через cppcheck,
- авторизации пользователя (или dev-режима без auth).

## Основные возможности

### 1) Трассировка выполнения

- Редактор C-кода с примерами.
- Запуск интерпретации через `POST /api/snapshot`.
- Навигация по шагам: в начало, назад, вперёд, в конец.
- Подсветка текущей строки и недавно изменённых значений.
- Отображение call stack, scope, переменных, массивов и 2D-массивов.

### 2) Режим блок-схем

- Отдельная вкладка «Блок-схема».
- Генерация SVG-схемы из C-кода через flowchart-сервис.

### 3) Авторизация

- Экран входа/регистрации в приложении.
- Проверка сессии, login/register/logout.
- Поддержка dev-режима без реального auth backend.

### 4) Анализ кода (cppcheck)

- Отдельная вкладка «Анализ кода».
- Отправка C-кода в `cppcheck-analyzer-service` (`POST /api/analyze`).
- Список найденных issues c сообщением и строкой.
- Подсветка строк с проблемами и фокус по клику на issue.

## Установка и запуск

```bash
cd frontend
npm install
npm run dev
```

Приложение доступно на `http://localhost:3000`.

## Переменные окружения

Поддерживаются переменные Vite (`.env`, `.env.local`):

- `VITE_AUTH_ENABLED` — `false` отключает реальную авторизацию (по умолчанию включена).
- `VITE_AUTH_SERVICE_URL` — URL auth-сервиса (по умолчанию `http://localhost:8083`).
- `VITE_FLOWCHART_SERVICE_URL` — URL flowchart-сервиса (по умолчанию `http://localhost:8081`).

Пример:

```env
VITE_AUTH_ENABLED=false
VITE_AUTH_SERVICE_URL=http://localhost:8083
VITE_FLOWCHART_SERVICE_URL=http://localhost:8081
```

## Proxy в dev-режиме

`vite.config.js` проксирует:

- `/api/snapshot` → `http://localhost:8080/snapshot`
- `/api/analyze` → `http://localhost:8086/analyze`
- `/api/auth/*` → `http://localhost:8083/api/auth/*`
- `/api/flowchart/*` → `http://localhost:8081/api/flowchart/*`

## Структура проекта

```text
frontend/
├── src/
│   ├── api/
│   │   ├── interpreter.js
│   │   ├── cppcheck.js
│   │   ├── auth.js
│   │   └── flowchart.js
│   ├── components/
│   │   ├── CodeEditor.vue
│   │   ├── RuntimeVisualization.vue
│   │   ├── FlowchartBuilder.vue
│   │   ├── StackFrame.vue
│   │   ├── Scope.vue
│   │   ├── Variable.vue
│   │   ├── Array.vue
│   │   └── Array2D.vue
│   ├── views/
│   │   ├── VisualizationView.vue
│   │   └── CodeAnalysisView.vue
│   ├── App.vue
│   ├── main.js
│   └── style.css
├── index.html
├── package.json
└── vite.config.js
```

## Сборка

```bash
npm run build
npm run preview
```

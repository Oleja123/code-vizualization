# Semantic Analyzer Service - HTTP API

## Обзор

Сервис семантического анализа предоставляет HTTP endpoints для валидации кода на C с проверкой семантики и опциональной проверкой компиляции.

## Конфигурация

Конфигурация управляется через `config.yaml` с опциональным переопределением через переменные окружения:

```yaml
server:
  port: 8080

onecompiler:
  api_url: "https://api.onecompiler.com/api/v1"
  api_key: ""
  enabled: false
  timeout_seconds: 10
```

### Опции конфигурации

- **server.port**: Порт для прослушивания (по умолчанию: 8080)
  - Может быть переопределён флагом `-port`

- **onecompiler.api_url**: Endpoint OneCompiler API
  - По умолчанию: `https://api.onecompiler.com/api/v1`

- **onecompiler.api_key**: API ключ OneCompiler
  - Устанавливается через конфиг файл или переменную окружения `ONECOMPILER_API_KEY`
  - Переменная окружения имеет приоритет над конфиг файлом

- **onecompiler.enabled**: Включить/выключить проверку компиляции
  - `true` - проверять компиляцию после семантической валидации
  - `false` - пропускать проверку компиляции (быстрее)

- **onecompiler.timeout_seconds**: Timeout для API запросов к OneCompiler
  - По умолчанию: 10 секунд

## Запуск сервера

### Базовое использование

```bash
./semantic-analyzer-service
# Слушает на :8080 (или порт из config.yaml)
```

### С кастомной конфигурацией

```bash
./semantic-analyzer-service -config /path/to/config.yaml -port 9000
```

### С API ключом OneCompiler

Установите переменную окружения перед запуском:

```bash
ONECOMPILER_API_KEY="your-api-key-here" ./semantic-analyzer-service
```

## Endpoints

### POST /validate

Валидирует C код на синтаксические и семантические ошибки, опционально проверяет компиляцию.

**Запрос:**

```json
{
  "code": "int main() { return 0; }"
}
```

**Успешный ответ (200 OK):**

```json
{
  "success": true,
  "program": {
    "type": "Program",
    "declarations": [
      {
        "type": "FunctionDecl",
        "name": "main",
        "returnType": {
          "baseType": "int",
          "pointerLevel": 0,
          "arraySizes": []
        },
        "parameters": [],
        "body": {
          "type": "BlockStmt",
          "statements": [
            {
              "type": "ReturnStmt",
              "value": {
                "type": "IntLiteral",
                "value": 0
              }
            }
          ]
        }
      }
    ]
  }
}
```

**Ошибка семантики (400 Bad Request):**

```json
{
  "success": false,
  "error": "Semantic error: [INVALID_VARIABLE_TYPE] неподдерживаемый тип переменной: float на строке 1, колонка 0"
}
```

**Ошибка компиляции (400 Bad Request):**

Когда OneCompiler включён и код не компилируется:

```json
{
  "success": false,
  "error": "Compilation error: ..."
}
```

**Ошибка парсинга (400 Bad Request):**

```json
{
  "success": false,
  "error": "Parse error: ..."
}
```

### GET /health

Проверка здоровья сервиса.

**Ответ (200 OK):**

```json
{
  "status": "healthy",
  "service": "semantic-analyzer-service"
}
```

### GET /info

Информация о сервисе и документация API.

**Ответ (200 OK):**

```json
{
  "service": "Semantic Analyzer Service",
  "version": "1.0.0",
  "endpoints": {
    "POST /validate": "Валидация C кода",
    "GET /health": "Проверка здоровья",
    "GET /info": "Информация о сервисе"
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

## Примеры

### Тестирование с curl

```bash
# Валидный код
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"int main() { return 0; }"}'

# Ошибка типа
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"float main() { return 0; }"}'

# Проверка здоровья
curl http://localhost:8080/health

# Информация о сервисе
curl http://localhost:8080/info
```

## Правила валидации

### Типы данных

Поддерживаются только типы `int` и `void`:
- Переменные должны быть объявлены как `int`
- Возвращаемый тип функции должен быть `int` или `void`
- Параметры функции должны быть `int`

Пример валидного кода:
```c
int add(int a, int b) {
  int result = a + b;
  return result;
}
```

### Операторы

**Поддерживаемые бинарные операторы:** `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||`

**Поддерживаемые унарные операторы:** `-`, `!`, `++`, `--`

**Поддерживаемые операторы присваивания:** `=`, `+=`, `-=`, `/=`, `%=`

**Неподдерживаемые операторы:** `&` (побитовое И), `|` (побитовое ИЛИ), `^` (XOR), `<<` (левый сдвиг), `>>` (правый сдвиг)

### Поддерживаемые операторы

Поддерживаются следующие типы операторов:
- Объявления функций
- Объявления и присваивания переменных
- Условные операторы if/else (else-if как вложенный if)
- Циклы while
- Циклы do-while
- Циклы for
- Операторы return
- Операторы goto и метки (label)
- Блочные операторы

## Проверка компиляции

Когда `onecompiler.enabled: true`, сервис:

1. Сначала проверяет семантику
2. Отправляет код на OneCompiler API для проверки компиляции
3. Возвращает ошибку если компиляция не удалась

Если проверка компиляции включена, но OneCompiler недоступен:
- Ответ содержит поле `error` с информацией об ошибке
- Семантическая валидация всё равно работает

### Включение проверки компиляции

1. Установите `onecompiler.enabled: true` в `config.yaml`
2. Опционально установите `onecompiler.api_key` если требуется API ключ
3. Перезагрузите сервер

Пример с API ключом:
```yaml
onecompiler:
  api_url: "https://api.onecompiler.com/api/v1"
  api_key: "your-key-here"
  enabled: true
  timeout_seconds: 10
```

Или используйте переменную окружения:
```bash
ONECOMPILER_API_KEY="your-key-here" ./semantic-analyzer-service
```

## Обработка ошибок

Все ошибки содержат описательные сообщения:

- **Ошибки парсинга**: Структура кода невалидна
- **Семантические ошибки**: Структура кода валидна, но нарушены семантические правила
- **Ошибки компиляции**: Код семантически валиден, но не компилируется
- **Сетевые ошибки**: OneCompiler API недоступен или истёк timeout

## Производительность

- Семантическая валидация быстрая (< 10мс для типичного кода)
- Проверка компиляции добавляет сетевую задержку (~1-2 сек в зависимости от OneCompiler API)
- Отключите проверку компиляции (`onecompiler.enabled: false`) для более быстрых ответов
- Timeout для проверки компиляции настраивается через `onecompiler.timeout_seconds`

## Архитектура

```
HTTP Запрос
    ↓
[Парсинг кода] → Ошибка парсинга
    ↓
[Семантическая валидация] → Семантическая ошибка
    ↓
[Проверка компиляции] (опционально) → Ошибка компиляции
    ↓
[Успешный ответ]
```

Сервис валидирует код в несколько слоёв:
1. **Синтаксис**: Парсинг C кода в AST
2. **Семантика**: Проверка типов и операторов
3. **Компиляция**: Опциональная проверка компиляции через OneCompiler

```yaml
server:
  port: 8080

onecompiler:
  api_url: "https://api.onecompiler.com/api/v1"
  api_key: ""
  enabled: false
  timeout_seconds: 10
```

### Configuration Options

- **server.port**: Port to listen on (default: 8080)
  - Can be overridden via `-port` flag

- **onecompiler.api_url**: OneCompiler API endpoint
  - Default: `https://api.onecompiler.com/api/v1`

- **onecompiler.api_key**: OneCompiler API key
  - Set via config file or `ONECOMPILER_API_KEY` environment variable
  - Environment variable takes precedence over config file

- **onecompiler.enabled**: Enable/disable compilation verification
  - Set to `true` to check code compilation after semantic validation
  - Set to `false` to skip compilation checks (faster responses)

- **onecompiler.timeout_seconds**: Timeout for OneCompiler API calls
  - Default: 10 seconds

## Running the Server

### Basic Usage

```bash
./semantic-analyzer-service
# Listens on :8080 (or port from config.yaml)
```

### With Custom Configuration

```bash
./semantic-analyzer-service -config /path/to/config.yaml -port 9000
```

### With OneCompiler API Key

Set the environment variable before running:

```bash
ONECOMPILER_API_KEY="your-api-key-here" ./semantic-analyzer-service
```

## Endpoints

### POST /validate

Validates C code for semantic errors and optionally checks compilation.

**Request:**

```json
{
  "code": "int main() { return 0; }"
}
```

**Successful Response (200 OK):**

```json
{
  "success": true,
  "program": {
    "type": "Program",
    "declarations": [
      {
        "type": "FunctionDecl",
        "name": "main",
        "returnType": {
          "baseType": "int",
          "pointerLevel": 0,
          "arraySizes": []
        },
        "parameters": [],
        "body": {
          "type": "BlockStmt",
          "statements": [
            {
              "type": "ReturnStmt",
              "value": {
                "type": "IntLiteral",
                "value": 0
              }
            }
          ]
        }
      }
    ]
  }
}
```

**Semantic Error Response (400 Bad Request):**

```json
{
  "success": false,
  "error": "Semantic error: [INVALID_VARIABLE_TYPE] unsupported variable type: float at line 1, column 0"
}
```

**Compilation Error Response (400 Bad Request):**

When OneCompiler is enabled and compilation fails:

```json
{
  "success": false,
  "error": "Compilation error: ..."
}
```

**Invalid Code Response (400 Bad Request):**

```json
{
  "success": false,
  "error": "Parse error: ..."
}
```

### GET /health

Health check endpoint.

**Response (200 OK):**

```json
{
  "status": "healthy",
  "service": "semantic-analyzer-service"
}
```

### GET /info

Service information and API documentation.

**Response (200 OK):**

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

## Examples

### Test with curl

```bash
# Valid code
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"int main() { return 0; }"}'

# Invalid type
curl -X POST http://localhost:8080/validate \
  -H "Content-Type: application/json" \
  -d '{"code":"float main() { return 0; }"}'

# Health check
curl http://localhost:8080/health

# Service info
curl http://localhost:8080/info
```

## Validation Rules

### Type Validation

Only `int` and `void` types are supported:
- Variables must be declared as `int`
- Function return types must be `int` or `void`
- Parameters must be `int`

Example valid code:
```c
int add(int a, int b) {
  int result = a + b;
  return result;
}
```

### Operator Validation

**Supported binary operators:** `+`, `-`, `*`, `/`, `%`, `==`, `!=`, `<`, `<=`, `>`, `>=`, `&&`, `||`

**Supported unary operators:** `-`, `!`, `++`, `--`

**Supported assignment operators:** `=`, `+=`, `-=`, `/=`, `%=`

**Unsupported operators:** `&` (bitwise AND), `|` (bitwise OR), `^` (XOR), `<<` (left shift), `>>` (right shift)

### Statement Support

Supported statements include:
- Function declarations
- Variable declarations and assignments
- If/else statements (else-if as nested if)
- While loops
- Do-while loops
- For loops
- Return statements
- Goto and label statements
- Block statements

## Compilation Verification

When `onecompiler.enabled: true`, the service will:

1. Pass semantic validation first
2. Send code to OneCompiler API for compilation check
3. Return error if compilation fails

If compilation verification is enabled but OneCompiler is unavailable:
- Response includes `error` field with compilation check failure details
- Semantic validation still works

### Enabling Compilation Verification

1. Set `onecompiler.enabled: true` in `config.yaml`
2. Optionally set `onecompiler.api_key` if API key is required
3. Restart the server

Example with API key:
```yaml
onecompiler:
  api_url: "https://api.onecompiler.com/api/v1"
  api_key: "your-key-here"
  enabled: true
  timeout_seconds: 10
```

Or use environment variable:
```bash
ONECOMPILER_API_KEY="your-key-here" ./semantic-analyzer-service
```

## Error Handling

All errors include descriptive messages:

- **Parse errors**: Code structure is invalid
- **Semantic errors**: Code structure is valid but violates semantic rules
- **Compilation errors**: Code is semantically valid but fails compilation
- **Network errors**: OneCompiler API is unreachable or times out

## Performance Considerations

- Semantic validation is fast (< 10ms for typical code)
- Compilation verification adds network latency (~1-2 seconds depending on OneCompiler API)
- Disable compilation verification (`onecompiler.enabled: false`) for faster responses
- Compilation check timeout is configurable via `onecompiler.timeout_seconds`

## Architecture

```
HTTP Request
    ↓
[Parse Code] → Syntax/Parse Error
    ↓
[Semantic Validation] → Semantic Error
    ↓
[Compilation Check] (optional) → Compilation Error
    ↓
[Success Response]
```

The service validates code in layers:
1. **Syntax**: Parse C code into AST
2. **Semantics**: Validate types and operators
3. **Compilation**: Optionally verify code compiles (via OneCompiler)

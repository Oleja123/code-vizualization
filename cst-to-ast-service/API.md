# HTTP API Documentation

## Overview

CST-to-AST Service предоставляет REST API для парсинга C кода в Abstract Syntax Tree.

## Base URL

```
http://localhost:8080
```

## Endpoints

### POST /parse

Парсит C код и возвращает AST или ошибку.

**Request:**
```json
{
  "code": "int factorial(int n) { if (n <= 1) return 1; return n * factorial(n - 1); }"
}
```

**Success Response (200):**
```json
{
  "ast": {
    "type": "Program",
    "declarations": [
      {
        "type": "FunctionDecl",
        "name": "factorial",
        "returnType": {
          "baseType": "int",
          "pointerLevel": 0,
          "arraySize": 0
        },
        "parameters": [
          {
            "type": { "baseType": "int", "pointerLevel": 0, "arraySize": 0 },
            "name": "n",
            "location": { "line": 1, "column": 14, "endLine": 1, "endColumn": 19 }
          }
        ],
        "body": { ... },
        "location": { "line": 1, "column": 0, "endLine": 1, "endColumn": 72 }
      }
    ],
    "location": { "line": 1, "column": 0, "endLine": 1, "endColumn": 72 }
  }
}
```

**Error Response (400):**
```json
{
  "error": "Parse error",
  "code": "StmtConversion",
  "message": "failed to convert statement",
  "location": {
    "line": 1,
    "column": 0,
    "endLine": 1,
    "endColumn": 9
  },
  "nodeType": "declaration"
}
```

**Error Codes:**
- `ParseFailed` - tree-sitter парсинг провалился
- `StmtConversion` - ошибка конвертации оператора
- `ExprUnsupported` - неподдерживаемое выражение
- `UnsupportedOperator` - неподдерживаемый оператор
- `RequiresLValue` - требуется левое значение (lvalue)
- `InvalidType` - невалидный тип данных
- `TreeSitterError` - синтаксическая ошибка в коде

### GET /health

Проверка статуса сервиса.

**Response (200):**
```json
{
  "status": "ok",
  "service": "cst-to-ast-service",
  "version": "1.0.0"
}
```

### GET /info

Информация об API и поддерживаемых конструкциях.

**Response (200):**
```json
{
  "name": "CST-to-AST Converter",
  "description": "Converts C code to Abstract Syntax Tree",
  "endpoints": {
    "POST /parse": "Parse C code and return AST or error",
    "GET /health": "Health check",
    "GET /info": "API information"
  },
  "supported_constructs": {
    "types": ["int", "int*", "int**", "int[N]"],
    "statements": [
      "variable declaration",
      "function declaration",
      "if/else if/else",
      "while",
      "for",
      "return",
      "break",
      "continue"
    ],
    "expressions": [
      "variables",
      "integer literals",
      "binary operations",
      "unary operations",
      "assignments",
      "function calls",
      "array access",
      "array initialization"
    ],
    "operators": {
      "binary": ["+", "-", "*", "/", "%", "==", "!=", "<", ">", "<=", ">=", "&&", "||", "&", "|", "^", "<<", ">>"],
      "unary": ["-", "!", "*", "&", "++", "--"],
      "assignment": ["=", "+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "<<=", ">>="]
    }
  }
}
```

## Examples

### curl

**Парсинг простой переменной:**
```bash
curl -X POST http://localhost:8080/parse \
  -H "Content-Type: application/json" \
  -d '{"code":"int x = 42;"}'
```

**Парсинг функции:**
```bash
curl -X POST http://localhost:8080/parse \
  -H "Content-Type: application/json" \
  -d '{"code":"int add(int a, int b) { return a + b; }"}'
```

**Обработка ошибки:**
```bash
curl -X POST http://localhost:8080/parse \
  -H "Content-Type: application/json" \
  -d '{"code":"int x = ;"}'
```

**Проверка здоровья:**
```bash
curl http://localhost:8080/health
```

**Получение информации об API:**
```bash
curl http://localhost:8080/info
```

## HTTP Status Codes

- `200 OK` - Успешный парс кода
- `400 Bad Request` - Синтаксическая ошибка в коде или невалидный запрос
- `405 Method Not Allowed` - Используется неправильный HTTP метод
- `500 Internal Server Error` - Внутренняя ошибка сервиса

## Supported C Subset

### Types
- `int` - целые числа
- `int*`, `int**` - указатели
- `int[N]` - массивы фиксированного размера

### Statements
- Объявления переменных: `int x = 5;`
- Объявления функций: `int add(int a, int b) { ... }`
- Условные операторы: `if`, `else if`, `else`
- Циклы: `while`, `for`
- Возврат: `return`
- Управление потоком: `break`, `continue`

### Expressions
- Переменные: `x`
- Литералы: `42`
- Бинарные операции: `a + b`, `x * y`, `a == b`
- Унарные операции: `-x`, `!flag`, `*ptr`, `&var`, `++i`, `i++`
- Присваивание: `x = 5`, `x += 3`, `x *= 2`
- Вызовы функций: `factorial(n)`
- Доступ к массивам: `arr[i]`
- Инициализация массивов: `{1, 2, 3}`

## Running the Server

```bash
cd cst-to-ast-service
go run cmd/server/main.go
```

Server starts at `http://localhost:8080`

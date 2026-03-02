# 🔷 Визуализатор блок-схем по ГОСТ 19.701-90

Java-сервис для генерации блок-схем из AST (Abstract Syntax Tree), создаваемого Go сервисом `cst-to-ast-service`.

## 🎯 Возможности

- ✅ Генерация блок-схем по ГОСТ 19.701-90
- ✅ Поддержка всех конструкций упрощённого C (if/else if/else, for, while, функции)
- ✅ Экспорт в SVG (векторный формат для веба)
- ✅ REST API для интеграции с веб-приложением
- ✅ Каждый блок имеет уникальный ID для последующей трассировки
- ✅ Связь блоков с исходным кодом через Location

## 📋 Требования

- Java 17+
- Maven 3.6+
- Go CST-to-AST сервис (для генерации AST)

## 🚀 Быстрый старт

### 1. Сборка проекта

```bash
cd flowchart-visualizer
mvn clean install
```

### 2. Запуск сервера

```bash
mvn spring-boot:run
```

Сервер запустится на `http://localhost:8080`

### 3. Проверка работоспособности

```bash
curl http://localhost:8080/api/flowchart/health
```

## 📡 API

### POST `/api/flowchart/generate`

Генерация SVG блок-схемы из AST.

**Request:**
```json
{
  "ast": {
    "type": "Program",
    "declarations": [...]
  }
}
```

**Response:**
```json
{
  "svg": "<svg>...</svg>",
  "metadata": {
    "success": true,
    "svgLength": 4521
  }
}
```

### GET `/api/flowchart/health`

Проверка работоспособности сервиса.

## 🎨 Веб-интерфейс (демо)

Откройте `demo/index.html` в браузере для интерактивной демонстрации.

1. Запустите Java сервер (`mvn spring-boot:run`)
2. Откройте `demo/index.html` в браузере
3. Нажмите "Загрузить пример" или вставьте свой AST
4. Нажмите "Создать блок-схему"

## 🏗️ Архитектура

```
┌─────────────┐
│ Go CST→AST  │
│  Service    │
└──────┬──────┘
       │ JSON AST
       ▼
┌─────────────┐
│   Java      │
│ Flowchart   │ ──→ SVG
│ Visualizer  │
└─────────────┘
       │
       ▼
┌─────────────┐
│  Web Client │
│ (Browser)   │
└─────────────┘
```

### Основные компоненты:

1. **AST Model** (`flowchart.ast`) - Модель данных AST (соответствует Go структурам)
2. **Flowchart Builder** (`flowchart.builder`) - Конвертация AST → граф блок-схемы
3. **SVG Renderer** (`flowchart.renderer`) - Генерация SVG по ГОСТ
4. **REST API** (`flowchart.api`) - Веб-интерфейс для клиента

## 📦 Структура проекта

```
flowchart-visualizer/
├── src/main/java/flowchart/
│   ├── model/              # Модель блоков блок-схемы
│   │   ├── FlowchartNode.java
│   │   ├── Nodes.java      # Терминатор, Процесс, Решение...
│   │   └── Location.java
│   ├── ast/                # Модель AST (JSON десериализация)
│   │   └── ASTModel.java
│   ├── builder/            # AST → Flowchart граф
│   │   └── FlowchartBuilder.java
│   ├── renderer/           # Flowchart → SVG
│   │   └── SVGRenderer.java
│   ├── api/                # REST контроллеры
│   │   └── FlowchartController.java
│   ├── FlowchartGenerator.java         # Главный API класс
│   └── FlowchartVisualizerApplication.java
├── demo/
│   └── index.html          # Интерактивное демо
├── pom.xml
└── README.md
```

## 🎯 Использование в коде

### Вариант 1: REST API

```javascript
// JavaScript (в вашем веб-приложении)
const ast = { /* JSON от Go сервиса */ };

fetch('http://localhost:8080/api/flowchart/generate', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ ast: ast })
})
.then(res => res.json())
.then(data => {
  document.getElementById('output').innerHTML = data.svg;
});
```

### Вариант 2: Прямое использование в Java

```java
import flowchart.FlowchartGenerator;

FlowchartGenerator generator = new FlowchartGenerator();

// Из JSON строки
String astJson = "{ ... }";
String svg = generator.generateSVG(astJson);

// Из файла
String svg = generator.generateSVGFromFile("path/to/ast.json");

// Сохранение в файл
generator.generateSVGToFile(astJson, "output.svg");
```

## 🔄 Интеграция с Go сервисом

Типичный workflow:

```
1. Пользователь вводит C код
   ↓
2. Go: ParseToAST(code) → AST JSON
   ↓
3. Java: generateSVG(ast) → SVG строка
   ↓
4. Браузер отображает SVG
```

## 🎨 Блоки по ГОСТ 19.701-90

| Тип блока | ГОСТ форма | Использование |
|-----------|-----------|---------------|
| Терминатор | Скруглённый прямоугольник | Начало/конец функции |
| Процесс | Прямоугольник | Присваивание, вызов функции |
| Решение | Ромб | if/else, условия |
| Цикл | Шестиугольник | for/while |
| Соединитель | Круг | break/continue |

## 🔮 Будущие возможности (для трассировки)

Каждый SVG блок имеет уникальный `id` и связь с исходным кодом через `Location`.

Для пошаговой визуализации выполнения:

```javascript
// Подсветить блок во время выполнения
function highlightNode(nodeId) {
  const node = document.getElementById(nodeId);
  node.classList.add('highlight');
}

// CSS
.highlight {
  fill: #ffeb3b;
  stroke: #f57c00;
  stroke-width: 3;
}
```

## 📝 Пример AST

```json
{
  "type": "Program",
  "declarations": [
    {
      "type": "FunctionDecl",
      "name": "main",
      "returnType": { "baseType": "int", "pointerLevel": 0, "arraySizes": [] },
      "parameters": [],
      "body": {
        "type": "BlockStmt",
        "statements": [
          {
            "type": "VariableDecl",
            "varType": { "baseType": "int", "pointerLevel": 0, "arraySizes": [] },
            "name": "x",
            "initExpr": {
              "type": "IntLiteral",
              "value": 10,
              "location": { "line": 1, "column": 13, "endLine": 1, "endColumn": 15 }
            },
            "location": { "line": 1, "column": 5, "endLine": 1, "endColumn": 16 }
          },
          {
            "type": "ReturnStmt",
            "value": {
              "type": "VariableExpr",
              "name": "x",
              "location": { "line": 2, "column": 12, "endLine": 2, "endColumn": 13 }
            },
            "location": { "line": 2, "column": 5, "endLine": 2, "endColumn": 14 }
          }
        ],
        "location": { "line": 0, "column": 16, "endLine": 3, "endColumn": 2 }
      },
      "location": { "line": 0, "column": 1, "endLine": 3, "endColumn": 2 }
    }
  ],
  "location": { "line": 0, "column": 1, "endLine": 3, "endColumn": 2 }
}
```

## 🐛 Отладка

### Включить debug логи:

В `application.properties`:
```properties
logging.level.flowchart=DEBUG
```

### Проверить AST:

```bash
# Сохранить AST от Go сервиса в файл
echo '{ "type": "Program", ... }' > test.json

# Сгенерировать SVG
curl -X POST http://localhost:8080/api/flowchart/generate \
  -H "Content-Type: application/json" \
  -d @test.json
```

## 🤝 Вклад

1. Fork репозиторий
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add some amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Создайте Pull Request

## 📄 Лицензия

MIT License - используйте свободно!

## 🙏 Благодарности

- Go CST-to-AST сервис от вашего товарища
- ГОСТ 19.701-90 стандарт для блок-схем
- Spring Boot за отличный REST framework

## 📧 Контакты

Вопросы? Создайте Issue в репозитории!

---

**Важно**: Этот сервис работает в паре с Go CST-to-AST сервисом. Убедитесь, что Go сервис запущен и доступен для получения AST из исходного C кода.

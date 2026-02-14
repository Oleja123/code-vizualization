# 🚀 Быстрый старт - Визуализатор блок-схем

## Что это?

Java-сервис для создания блок-схем по ГОСТ 19.701-90 из AST твоего Go сервиса.

## ⚡ Запуск за 3 шага

### 1. Распаковать архив

```bash
unzip flowchart-visualizer.zip
cd flowchart-visualizer
```

### 2. Запустить сервер

```bash
mvn spring-boot:run
```

Или если Maven не установлен:

```bash
# Сначала установи Maven
sudo apt install maven  # Ubuntu/Debian
brew install maven      # MacOS

# Потом запускай
mvn spring-boot:run
```

### 3. Открыть демо

Открой в браузере: `demo/index.html`

**Готово!** 🎉

---

## 📖 Как использовать

### Вариант 1: Веб-интерфейс (демо)

1. Запусти сервер: `mvn spring-boot:run`
2. Открой `demo/index.html` в браузере
3. Нажми **"Загрузить пример"**
4. Нажми **"Создать блок-схему"**
5. Профит! Блок-схема появится справа

### Вариант 2: REST API (для твоего веб-приложения)

```javascript
// В твоём фронтенде
const ast = { /* JSON от Go сервиса */ };

fetch('http://localhost:8080/api/flowchart/generate', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ ast: ast })
})
.then(res => res.json())
.then(data => {
  // data.svg содержит SVG блок-схему
  document.getElementById('flowchart').innerHTML = data.svg;
});
```

### Вариант 3: Из командной строки

```bash
# Отправить AST через curl
curl -X POST http://localhost:8080/api/flowchart/generate \
  -H "Content-Type: application/json" \
  -d '{
    "ast": {
      "type": "Program",
      "declarations": [...]
    }
  }' | jq -r '.svg' > flowchart.svg

# Открыть в браузере
open flowchart.svg  # MacOS
xdg-open flowchart.svg  # Linux
```

---

## 🔗 Интеграция с Go сервисом

```
Пользователь вводит C код
         ↓
Go сервис: ParseToAST(code) → AST JSON
         ↓
Java сервис: generateSVG(ast) → SVG
         ↓
Браузер отображает SVG
```

Пример workflow:

```go
// В Go сервисе
conv := converter.New()
ast, err := conv.ParseToAST(sourceCode)
astJSON, _ := json.Marshal(ast)

// Отправляем на Java сервис
resp, _ := http.Post("http://localhost:8080/api/flowchart/generate",
    "application/json",
    bytes.NewBuffer(astJSON))
```

---

## 🎯 Блоки по ГОСТ

- **Терминатор** (скруглённый прямоугольник) - начало/конец функции
- **Процесс** (прямоугольник) - присваивание, вызов функции
- **Решение** (ромб) - if/else
- **Цикл** (шестиугольник) - for/while
- **Соединитель** (круг) - break/continue

---

## 🔮 Для трассировки (в будущем)

Каждый блок в SVG имеет уникальный ID:

```javascript
// Подсветить блок при выполнении
function highlightBlock(nodeId) {
  const node = document.getElementById(nodeId);
  node.classList.add('highlight');
}

// CSS для подсветки (уже есть в demo/index.html)
.highlight {
  fill: #ffeb3b !important;
  stroke: #f57c00 !important;
  stroke-width: 3 !important;
}
```

---

## ❓ Проблемы?

### Сервер не запускается

```bash
# Проверь версию Java (нужна 17+)
java -version

# Установи Java 17
sudo apt install openjdk-17-jdk  # Ubuntu
brew install openjdk@17          # MacOS
```

### Порт 8080 занят

Измени порт в `src/main/resources/application.properties`:

```properties
server.port=8081
```

### AST не парсится

Убедись что JSON валидный:

```bash
# Проверь через jq
echo '{"type":"Program"...}' | jq .
```

---

## 📚 Больше информации

Смотри полный **README.md** для деталей.

---

Удачи! 🚀

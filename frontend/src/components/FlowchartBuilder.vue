<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { generateFromCode } from '../api/flowchart.js'

const codeInput = ref(`int main() {
    int x = 10;
    if (x > 5) {
        x = x - 1;
    }
    return 0;
}`)
const loading = ref(false)
const error = ref('')
const success = ref(false)
const svgResult = ref('')
const currentZoom = ref(1)
const lineNumbers = ref('')

// Примеры кода
// Примеры кода
const EXAMPLES = {
  dowhile: `void main() {
    int year = 2014;
    int population = 650;
    do {
        population = (population * 103) / 100;
        year = year + 1;
    } while (year <= 2040);
}`,

  minmax: `void main() {
    int a, b, min, max;

    if (a < b) {
        min = a;
        max = b;
    } else {
        min = b;
        max = a;
    }
}`,

  whilecontinue: `void main() {
    int a = 1999;
    while (a < 2030) {
        a = a + 1;
        if (a % 4 == 0)
            continue;
    }
}`,

  arrayfor: `int a1[5] = {1, 2, 3, 7, 8};

void main() {
    int i, s;

    for (i = 0; i < 5; i++)
        if (a1[i] % 2 == 1)
            a1[i] = 1;

    s = 1;
    for (i = 1; i < 5; i++)
        s += a1[i];
}`,

  prime: `int isPrime(int num) {
    int del = 2;
    while (del < num) {
        if (num % del == 0) {
            return 0;
        }
        del++;
    }
    return 1;
}

void main() {
    int num = 20;

    while (1) {
        if (isPrime(num)) {
            break;
        }
        num++;
    }
}`
}

const showExamples = ref(false)

// Обновление номеров строк
function updateLineNumbers() {
  const lines = codeInput.value.split('\n').length
  lineNumbers.value = Array.from({length: lines}, (_, i) => i + 1).join('\n')
}

watch(codeInput, updateLineNumbers)
onMounted(updateLineNumbers)

// Загрузка примера
function loadExample(key) {
  codeInput.value = EXAMPLES[key]
  showExamples.value = false
  error.value = ''
  success.value = false
}

// Генерация блок-схемы
async function generate() {
  if (!codeInput.value.trim()) {
    error.value = 'Введите C код'
    return
  }

  error.value = ''
  success.value = false
  svgResult.value = ''
  loading.value = true

  try {
    const data = await generateFromCode(codeInput.value)
    svgResult.value = data.svg
    success.value = true
    currentZoom.value = 1
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

// Zoom
function zoom(delta) {
  if (delta === 0) {
    currentZoom.value = 1
  } else {
    currentZoom.value = Math.max(0.3, Math.min(4, currentZoom.value + delta))
  }
}

const zoomLabel = computed(() => Math.round(currentZoom.value * 100) + '%')

// Скачать SVG
function downloadSVG() {
  if (!svgResult.value) return
  const blob = new Blob([svgResult.value], { type: 'image/svg+xml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = 'flowchart.svg'
  a.click()
  URL.revokeObjectURL(url)
}

// Обработка Tab в textarea
function handleTab(e) {
  if (e.key === 'Tab') {
    e.preventDefault()
    const start = e.target.selectionStart
    const end = e.target.selectionEnd
    codeInput.value = codeInput.value.substring(0, start) + '    ' + codeInput.value.substring(end)
    e.target.selectionStart = e.target.selectionEnd = start + 4
  }
}
</script>

<template>
  <div class="workspace-grid">
    <!-- ═══ ЛЕВАЯ ПАНЕЛЬ - РЕДАКТОР ═══ -->
    <div class="editor-panel">
      <!-- Заголовок -->
      <div class="panel-header">
        <div class="panel-label">
          <span class="panel-dot"></span>
          C Code Editor
        </div>
        <div class="examples-wrapper">
          <button class="examples-btn" @click="showExamples = !showExamples">
            📚 Примеры
          </button>
          <div v-if="showExamples" class="examples-dropdown">
  <div class="example-item" @click="loadExample('dowhile')">Do-While population</div>
<div class="example-item" @click="loadExample('minmax')">Min / Max</div>
<div class="example-item" @click="loadExample('whilecontinue')">While + Continue</div>
<div class="example-item" @click="loadExample('arrayfor')">Array + For</div>
<div class="example-item" @click="loadExample('prime')">Prime finder</div>
          </div>
        </div>
      </div>

      <!-- Редактор с номерами строк -->
      <div class="editor-wrapper">
        <div class="line-numbers">{{ lineNumbers }}</div>
        <textarea
          v-model="codeInput"
          class="code-editor"
          spellcheck="false"
          @keydown="handleTab"
        ></textarea>
      </div>

      <!-- Футер с кнопкой -->
      <div class="editor-footer">
        <div class="error-msg" v-if="error">✗ {{ error }}</div>
        <div class="success-msg" v-if="success">✓ Блок-схема успешно сгенерирована</div>
        <button class="btn-run" :disabled="loading" @click="generate">
          <span v-if="loading">⟳</span>
          <span v-else>▶</span>
          {{ loading ? 'Генерация…' : 'Сгенерировать' }}
        </button>
      </div>
    </div>

    <!-- ═══ ПРАВАЯ ПАНЕЛЬ - РЕЗУЛЬТАТ ═══ -->
    <div class="output-panel">
      <!-- Заголовок -->
      <div class="panel-header">
        <div class="panel-label">
          <span class="panel-dot"></span>
          Блок-схема (ГОСТ 19.701-90)
        </div>
        <div class="controls-right" v-if="svgResult">
          <button class="icon-btn" @click="zoom(-0.1)" title="Уменьшить">−</button>
          <span class="zoom-label">{{ zoomLabel }}</span>
          <button class="icon-btn" @click="zoom(0.1)" title="Увеличить">+</button>
          <button class="icon-btn" @click="zoom(0)" title="Сбросить">↻</button>
          <button class="icon-btn download" @click="downloadSVG" title="Скачать SVG">⬇</button>
        </div>
      </div>

      <!-- Контент -->
      <div class="output-content">
        <!-- Placeholder -->
        <div v-if="!svgResult && !loading" class="placeholder">
          <div class="placeholder-icon">📊</div>
          <div class="placeholder-text">Введите код и нажмите «Сгенерировать»</div>
          <div class="placeholder-hint">Блок-схема появится здесь</div>
        </div>

        <!-- Spinner -->
        <div v-if="loading" class="spinner-wrapper">
          <div class="spinner"></div>
          <div class="spinner-text">Генерация блок-схемы…</div>
        </div>

        <!-- SVG -->
        <div v-if="svgResult && !loading" class="svg-container">
          <div 
            class="svg-wrapper" 
            :style="{ transform: `scale(${currentZoom})` }"
            v-html="svgResult"
          ></div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.workspace-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  height: 100%;
  overflow: hidden;
}

/* ═══ ОБЩИЕ СТИЛИ ПАНЕЛЕЙ ═══ */
.editor-panel,
.output-panel {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: white;
}

.editor-panel {
  border-right: 1px solid #e2e8f0;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  height: 40px;
  border-bottom: 1px solid #e2e8f0;
  background: #fafafa;
  flex-shrink: 0;
}

.panel-label {
  font-family: 'Courier New', monospace;
  font-size: 11px;
  color: #475569;
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.panel-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #4f46e5;
}

/* ═══ ПРИМЕРЫ ═══ */
.examples-wrapper {
  position: relative;
}

.examples-btn {
  background: white;
  border: 1px solid #e2e8f0;
  color: #475569;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 11px;
  cursor: pointer;
  transition: all .2s;
}

.examples-btn:hover {
  border-color: #4f46e5;
  color: #4f46e5;
}

.examples-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 4px;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 6px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  min-width: 140px;
  z-index: 100;
}

.example-item {
  padding: 8px 12px;
  font-size: 12px;
  cursor: pointer;
  transition: background .15s;
}

.example-item:hover {
  background: #f5f7fb;
}

.example-item:first-child {
  border-radius: 6px 6px 0 0;
}

.example-item:last-child {
  border-radius: 0 0 6px 6px;
}

/* ═══ РЕДАКТОР ═══ */
.editor-wrapper {
  flex: 1;
  position: relative;
  overflow: hidden;
}

.line-numbers {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 44px;
  padding: 16px 8px 16px 0;
  text-align: right;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  color: #94a3b8;
  background: #fafafa;
  border-right: 1px solid #e2e8f0;
  user-select: none;
  overflow: hidden;
  white-space: pre;
}

.code-editor {
  position: absolute;
  left: 44px;
  top: 0;
  right: 0;
  bottom: 0;
  background: white;
  color: #0f172a;
  border: none;
  outline: none;
  resize: none;
  padding: 16px 16px 16px 12px;
  font-family: 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
  tab-size: 4;
  overflow-y: auto;
}

.code-editor::selection {
  background: rgba(79, 70, 229, 0.2);
}

/* ═══ ФУТЕР РЕДАКТОРА ═══ */
.editor-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  height: 52px;
  border-top: 1px solid #e2e8f0;
  background: #fafafa;
  flex-shrink: 0;
  gap: 12px;
}

.error-msg {
  flex: 1;
  font-size: 12px;
  font-family: 'Courier New', monospace;
  color: #dc2626;
}

.success-msg {
  flex: 1;
  font-size: 12px;
  font-family: 'Courier New', monospace;
  color: #16a34a;
}

.btn-run {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 20px;
  height: 36px;
  background: #4f46e5;
  color: white;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all .2s;
  white-space: nowrap;
}

.btn-run:hover:not(:disabled) {
  background: #4338ca;
  transform: translateY(-1px);
}

.btn-run:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-run span:first-child {
  font-size: 16px;
}

/* ═══ ПРАВАЯ ПАНЕЛЬ ═══ */
.controls-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.icon-btn {
  width: 28px;
  height: 28px;
  border: 1px solid #e2e8f0;
  background: white;
  color: #475569;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all .2s;
}

.icon-btn:hover {
  border-color: #4f46e5;
  color: #4f46e5;
  background: #f5f7fb;
}

.icon-btn.download {
  margin-left: 8px;
}

.zoom-label {
  font-family: 'Courier New', monospace;
  font-size: 11px;
  color: #475569;
  min-width: 42px;
  text-align: center;
}

/* ═══ КОНТЕНТ ПРАВОЙ ПАНЕЛИ ═══ */
.output-content {
  flex: 1;
  overflow: auto;
  position: relative;
  background: #fafafa;
}

.placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  color: #94a3b8;
}

.placeholder-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.placeholder-text {
  font-size: 14px;
  font-weight: 500;
  color: #64748b;
  margin-bottom: 4px;
}

.placeholder-hint {
  font-size: 12px;
  color: #94a3b8;
}

/* ═══ SPINNER ═══ */
.spinner-wrapper {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 4px solid #e2e8f0;
  border-top-color: #4f46e5;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 12px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.spinner-text {
  font-size: 13px;
  color: #64748b;
}

/* ═══ SVG КОНТЕЙНЕР ═══ */
.svg-container {
  padding: 24px;
  overflow: auto;
}

.svg-wrapper {
  transform-origin: top left;
  transition: transform 0.2s;
}

.svg-wrapper :deep(svg) {
  max-width: 100%;
  height: auto;
  display: block;
}
</style>
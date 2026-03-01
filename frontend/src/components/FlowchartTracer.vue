<script setup>
import { ref, watch, computed, nextTick } from 'vue'
import { generateAllFunctions } from '../api/flowchart.js'
import { getSnapshot } from '../api/interpreter.js'
import RuntimeVisualization from './RuntimeVisualization.vue'

// ──────────── Константы ────────────
const LINE_H    = 22   // высота строки в px (должна совпадать с line-height в CSS)
const EDITOR_PT = 14   // padding-top редактора в px

// ──────────── Состояние ────────────
const codeInput = ref(`int factorial(int n) {
    int result = 1;
    int i = 1;
    while (i <= n) {
        result = result * i;
        i = i + 1;
    }
    return result;
}

int main() {
    int x = factorial(5);
    return x;
}`)

const loading      = ref(false)
const error        = ref('')
const phase        = ref('idle') // 'idle' | 'generating' | 'ready'

// Блок-схемы: { funcName -> svgString }
const functionSvgs = ref({})
const functionTabs = ref([])   // [{ name, svg }]
const activeTab    = ref('')

// Трассировка
const snapshot     = ref(null)
const currentStep  = ref(0)
const stepsCount   = ref(0)
const tracing      = ref(false)

// Ссылки на DOM
const svgContainers = ref({}) // { funcName -> el }

// ──────────── Примеры ────────────
const EXAMPLES = {
  simple:    { label: 'Простой',     code: `int main() {\n    int x = 5;\n    int y = 10;\n    int sum = x + y;\n    return sum;\n}` },
  if:        { label: 'If',          code: `int main() {\n    int x = 10;\n    if (x > 5) {\n        x = x - 1;\n    }\n    return 0;\n}` },
  ifelse:    { label: 'If-Else',     code: `int main() {\n    int x = 10;\n    if (x > 5) {\n        x = x - 1;\n    } else {\n        x = x + 1;\n    }\n    return x;\n}` },
  while:     { label: 'While',       code: `int main() {\n    int i = 0;\n    int sum = 0;\n    while (i < 5) {\n        sum = sum + i;\n        i = i + 1;\n    }\n    return sum;\n}` },
  for:       { label: 'For',         code: `int main() {\n    int sum = 0;\n    int i;\n    for (i = 0; i < 5; i = i + 1) {\n        sum = sum + i;\n    }\n    return 0;\n}` },
  nested:    { label: 'Вложенный',   code: `int main() {\n    int x = 15;\n    int result = 0;\n    if (x > 10) {\n        int i = 0;\n        while (i < x) {\n            result = result + 1;\n            i = i + 1;\n        }\n    } else {\n        result = x;\n    }\n    return 0;\n}` },
  multifunc: { label: 'Две функции', code: `int isPrime(int num) {\n    int del = 2;\n    while (del < num) {\n        if (num % del == 0) {\n            return 0;\n        }\n        del++;\n    }\n    return 1;\n}\n\nint main() {\n    int result = isPrime(7);\n    return result;\n}` },
  factorial: { label: 'Факториал',   code: `int factorial(int n) {\n    int result = 1;\n    int i = 1;\n    while (i <= n) {\n        result = result * i;\n        i = i + 1;\n    }\n    return result;\n}\n\nint main() {\n    int x = factorial(5);\n    return x;\n}` },
}
const showExamples = ref(false)

function loadExample(key) {
  codeInput.value = EXAMPLES[key].code
  showExamples.value = false
  resetAll()
}

// ──────────── Линейные номера ────────────
const lineNumbers = computed(() => {
  const count = codeInput.value.split('\n').length
  return Array.from({ length: count }, (_, i) => i + 1)
})

const currentLine = computed(() => snapshot.value?.line ?? null)

// ──────────── Подсветка строки в редакторе ────────────
const editorEl = ref(null)
const lineNumbersEl = ref(null)

watch(currentLine, async (line) => {
  if (!line || !editorEl.value) return
  await nextTick()
  const lineH = LINE_H
  const scrollTo = Math.max(0, (line - 1) * lineH - editorEl.value.clientHeight / 2)
  editorEl.value.scrollTop = scrollTo
  if (lineNumbersEl.value) lineNumbersEl.value.scrollTop = scrollTo
})

function syncScroll(e) {
  if (lineNumbersEl.value) lineNumbersEl.value.scrollTop = e.target.scrollTop
}

function handleTab(e) {
  if (e.key === 'Tab') {
    e.preventDefault()
    const s = e.target.selectionStart
    const end = e.target.selectionEnd
    codeInput.value = codeInput.value.substring(0, s) + '    ' + codeInput.value.substring(end)
    nextTick(() => { e.target.selectionStart = e.target.selectionEnd = s + 4 })
  }
}

// ──────────── Шаг 1: Генерация блок-схемы ────────────
async function generate() {
  if (!codeInput.value.trim()) { error.value = 'Введите C-код'; return }
  resetAll()
  loading.value = true
  error.value = ''
  phase.value = 'generating'

  try {
    const data = await generateAllFunctions(codeInput.value)
    if (!data?.functions || Object.keys(data.functions).length === 0) {
      throw new Error('Блок-схема не сгенерирована')
    }

    const tabs = Object.entries(data.functions).map(([name, svg]) => ({ name, svg }))
    tabs.sort((a, b) => {
      if (a.name === 'main') return -1
      if (b.name === 'main') return 1
      return a.name.localeCompare(b.name)
    })
    functionTabs.value = tabs
    functionSvgs.value = Object.fromEntries(tabs.map(t => [t.name, t.svg]))
    activeTab.value = tabs[0].name
    phase.value = 'ready'
  } catch (e) {
    error.value = e.message
    phase.value = 'idle'
  } finally {
    loading.value = false
  }
}

// ──────────── Шаг 2: Запуск трассировки ────────────
async function startTracing() {
  error.value = ''
  loading.value = true
  currentStep.value = 0
  snapshot.value = null

  try {
    const data = await getSnapshot(codeInput.value, 0)
    snapshot.value = data.snapshot
    currentStep.value = data.current_step ?? 0
    stepsCount.value  = data.steps_count  ?? 0
    tracing.value = true
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function stopTracing() {
  tracing.value = false
  snapshot.value = null
  currentStep.value = 0
  clearHighlight()
}

async function stepForward() {
  if (currentStep.value >= stepsCount.value - 1 || loading.value) return
  await loadStep(currentStep.value + 1)
}

async function stepBackward() {
  if (currentStep.value <= 0 || loading.value) return
  await loadStep(currentStep.value - 1)
}

async function loadStep(step) {
  loading.value = true
  try {
    const data = await getSnapshot(codeInput.value, step)
    snapshot.value = data.snapshot
    currentStep.value = data.current_step ?? step
    stepsCount.value  = data.steps_count  ?? stepsCount.value
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

// ──────────── Сброс ────────────
function resetAll() {
  tracing.value = false
  snapshot.value = null
  currentStep.value = 0
  stepsCount.value = 0
  functionTabs.value = []
  functionSvgs.value = {}
  activeTab.value = ''
  phase.value = 'idle'
  clearHighlight()
}

function editCode() {
  resetAll()
}

// ──────────── Подсветка блока в SVG ────────────

function clearHighlight() {
  for (const el of Object.values(svgContainers.value)) {
    if (!el) continue
    el.querySelectorAll('.node-active').forEach(n => n.classList.remove('node-active'))
  }
}

function highlightNodeInSvg(containerEl, line) {
  if (!containerEl || !line) return

  // Снять старую подсветку
  containerEl.querySelectorAll('.node-active').forEach(n => n.classList.remove('node-active'))

  // Найти узел с наиболее точным совпадением строки
  const nodes = containerEl.querySelectorAll('[data-line]')
  let best = null
  let bestSpan = Infinity

  for (const node of nodes) {
    const start = parseInt(node.dataset.line)
    const end   = parseInt(node.dataset.lineEnd)
    if (line >= start && line <= end) {
      const span = end - start
      if (span < bestSpan) {
        bestSpan = span
        best = node
      }
    }
  }

  if (best) {
    best.classList.add('node-active')
    best.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
  }
}

// Следим за снапшотом: подсвечиваем блок и переключаем вкладку
watch(snapshot, async (snap) => {
  if (!snap) return
  await nextTick()

  const line = snap.line

  // Определяем активную функцию из стека вызовов
  const frames = snap.call_stack?.frames ?? []
  // Последний не-global фрейм = текущая функция
  const activeFuncFrame = [...frames].reverse().find(f => f.func_name !== 'global')
  const activeFunc = activeFuncFrame?.func_name ?? null

  // Переключаем вкладку если функция есть в наших табах
  if (activeFunc && functionSvgs.value[activeFunc] && activeTab.value !== activeFunc) {
    activeTab.value = activeFunc
  }

  // Подсвечиваем блок в активной вкладке
  await nextTick()
  const containerEl = svgContainers.value[activeTab.value]
  if (containerEl && line) {
    highlightNodeInSvg(containerEl, line)
  }
})

// ──────────── Прогресс ────────────
const progressPct = computed(() => {
  if (!stepsCount.value) return 0
  return Math.round((currentStep.value / (stepsCount.value - 1)) * 100)
})

// ──────────── Зум ────────────
const zoom = ref({})
function getZoom(name) { return zoom.value[name] ?? 1 }
function setZoom(name, delta) {
  const cur = zoom.value[name] ?? 1
  zoom.value = { ...zoom.value, [name]: delta === 0 ? 1 : Math.max(0.3, Math.min(4, cur + delta)) }
}
const zoomLabel = computed(() => name => Math.round(getZoom(name) * 100) + '%')

// ──────────── Клавиатура ────────────
function onKeydown(e) {
  if (!tracing.value) return
  if (e.key === 'ArrowRight' || e.key === 'ArrowDown') { e.preventDefault(); stepForward() }
  if (e.key === 'ArrowLeft'  || e.key === 'ArrowUp')   { e.preventDefault(); stepBackward() }
}
</script>

<template>
  <div class="tracer-root" @keydown="onKeydown" tabindex="0">

    <!-- ══════ КОЛОНКА 1: Редактор ══════ -->
    <div class="col col-editor">
      <div class="col-header">
        <span class="panel-label"><span class="dot dot-blue"></span>C Code Editor</span>
        <div class="examples-wrap">
          <button class="btn-sm" @click="showExamples = !showExamples">📚 Примеры</button>
          <div v-if="showExamples" class="dropdown">
            <div v-for="(ex, key) in EXAMPLES" :key="key" class="dropdown-item" @click="loadExample(key)">{{ ex.label }}</div>
          </div>
        </div>
      </div>

      <div class="editor-area">
        <div class="line-numbers" ref="lineNumbersEl">
          <div
            v-for="n in lineNumbers"
            :key="n"
            class="ln"
            :class="{ 'ln-active': n === currentLine }"
          >{{ n }}</div>
        </div>
        <textarea
          ref="editorEl"
          class="code-editor"
          :class="{ 'code-readonly': phase !== 'idle' }"
          v-model="codeInput"
          spellcheck="false"
          :readonly="phase !== 'idle'"
          @scroll="syncScroll"
          @keydown="handleTab"
        ></textarea>
        <div v-if="currentLine" class="line-highlight" :style="{ top: (EDITOR_PT + (currentLine - 1) * LINE_H) + 'px' }"></div>
      </div>

      <div class="col-footer">
        <div class="msg error-msg" v-if="error">✗ {{ error }}</div>

        <template v-if="phase === 'idle'">
          <button class="btn btn-generate" :disabled="loading" @click="generate">
            <span v-if="loading">⟳</span><span v-else>▶</span>
            {{ loading ? 'Генерация…' : 'Сгенерировать схему' }}
          </button>
        </template>

        <template v-else-if="phase === 'ready' && !tracing">
          <button class="btn btn-trace" :disabled="loading" @click="startTracing">
            {{ loading ? '⟳' : '⚡' }} Начать трассировку
          </button>
          <button class="btn btn-secondary" @click="editCode">✏️ Редактировать</button>
        </template>

        <template v-else-if="tracing">
          <div class="trace-controls">
            <div class="progress-row">
              <div class="progress-wrap"><div class="progress-bar" :style="{ width: progressPct + '%' }"></div></div>
              <span class="step-label">{{ currentStep }} / {{ stepsCount - 1 }}</span>
            </div>
            <div class="btn-row">
              <button class="btn btn-step" :disabled="loading || currentStep <= 0" @click="stepBackward">← Назад</button>
              <button class="btn btn-step" :disabled="loading || currentStep >= stepsCount - 1" @click="stepForward">Вперёд →</button>
              <button class="btn btn-secondary icon-btn" @click="() => { currentStep = 0; loadStep(0) }" :disabled="loading" title="В начало">↺</button>
              <button class="btn btn-stop" @click="stopTracing">■ Стоп</button>
            </div>
          </div>
        </template>
      </div>
    </div>

    <!-- ══════ КОЛОНКА 2: Блок-схема ══════ -->
    <div class="col col-flowchart">
      <div v-if="functionTabs.length === 0 && !loading" class="placeholder">
        <div class="ph-icon">📊</div>
        <div class="ph-text">Блок-схема появится здесь</div>
        <div class="ph-hint">Введите C-код и нажмите «Сгенерировать схему»</div>
        <div class="ph-hint" style="margin-top:4px;color:#94a3b8">Затем запустите трассировку — активный блок будет подсвечен</div>
      </div>

      <div v-if="loading && functionTabs.length === 0" class="spinner-wrap">
        <div class="spinner"></div>
        <div>Генерация блок-схемы…</div>
      </div>

      <template v-if="functionTabs.length > 0">
        <div class="col-header">
          <div class="func-tabs">
            <button
              v-for="tab in functionTabs"
              :key="tab.name"
              class="func-tab"
              :class="{ active: activeTab === tab.name }"
              @click="activeTab = tab.name"
            >
              ƒ {{ tab.name }}
            </button>
          </div>
          <div class="zoom-controls">
            <button class="zoom-btn" @click="setZoom(activeTab, -0.15)" title="Уменьшить">−</button>
            <span class="zoom-label">{{ Math.round(getZoom(activeTab) * 100) }}%</span>
            <button class="zoom-btn" @click="setZoom(activeTab, 0.15)" title="Увеличить">+</button>
            <button class="zoom-btn" @click="setZoom(activeTab, 0)" title="Сбросить">↻</button>
          </div>
        </div>

        <div class="svg-scroll">
          <template v-for="tab in functionTabs" :key="tab.name">
            <div
              v-show="activeTab === tab.name"
              :ref="el => { if (el) svgContainers[tab.name] = el }"
              class="svg-container"
            >
              <div
                class="svg-inner"
                :style="{ transform: `scale(${getZoom(tab.name)})` }"
                v-html="tab.svg"
              ></div>
            </div>
          </template>
        </div>
      </template>
    </div>

    <!-- ══════ КОЛОНКА 3: Переменные / стек ══════ -->
    <div class="col col-vars">
      <div class="col-header">
        <span class="panel-label"><span class="dot dot-green"></span>Состояние программы</span>
        <span v-if="tracing" class="step-badge">Шаг {{ currentStep }}</span>
      </div>
      <div class="vars-body">
        <div v-if="!tracing" class="placeholder small">
          <div class="ph-icon" style="font-size:32px">🖥️</div>
          <div class="ph-text" style="font-size:12px">Состояние переменных и стек вызовов<br>отобразятся во время трассировки</div>
        </div>
        <RuntimeVisualization v-else-if="snapshot" :snapshot="snapshot" :current-step="currentStep" />
      </div>
    </div>

  </div>
</template>

<style scoped>
/* ══ Корень ══ */
.tracer-root {
  display: grid;
  grid-template-columns: minmax(320px, 380px) 1fr minmax(280px, 340px);
  height: 100%;
  overflow: hidden;
  outline: none;
  background: #f5f7fb;
}

/* ══ Общие колонки ══ */
.col {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-right: 1px solid #e2e8f0;
  background: white;
}
.col:last-child { border-right: none; }

.col-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 42px;
  padding: 0 14px;
  border-bottom: 1px solid #e2e8f0;
  background: #fafafa;
  flex-shrink: 0;
  gap: 8px;
}

.panel-label {
  font-family: 'Courier New', monospace;
  font-size: 11px;
  font-weight: 700;
  color: #475569;
  display: flex;
  align-items: center;
  gap: 7px;
  white-space: nowrap;
}

.dot {
  width: 8px; height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.dot-blue  { background: #4f46e5; }
.dot-green { background: #16a34a; }

.step-badge {
  font-size: 11px;
  font-family: monospace;
  background: #4f46e5;
  color: white;
  padding: 2px 8px;
  border-radius: 10px;
  white-space: nowrap;
}

/* ══ Редактор ══ */
.editor-area {
  flex: 1;
  position: relative;
  overflow: hidden;
  display: flex;
  min-height: 0;
}

.line-numbers {
  width: 44px;
  flex-shrink: 0;
  padding: 14px 0;
  background: #fafafa;
  border-right: 1px solid #e2e8f0;
  overflow: hidden;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 22px;
  color: #94a3b8;
  user-select: none;
}

.ln {
  text-align: right;
  padding-right: 8px;
  height: 22px;
  line-height: 22px;
}

.ln-active {
  background: #fff3cd;
  color: #856404;
  font-weight: bold;
}

.code-editor {
  flex: 1;
  padding: 14px 10px;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 22px;
  background: white;
  border: none;
  outline: none;
  resize: none;
  overflow-y: auto;
  tab-size: 4;
}

.code-editor.code-readonly {
  background: #f8f9fa;
  color: #495057;
  cursor: default;
}

.line-highlight {
  position: absolute;
  left: 44px;
  right: 0;
  height: 22px;
  background: rgba(255, 200, 0, 0.18);
  border-left: 3px solid #f59e0b;
  pointer-events: none;
  z-index: 1;
}

/* ══ Футер редактора ══ */
.col-footer {
  padding: 8px 12px;
  border-top: 1px solid #e2e8f0;
  background: #fafafa;
  flex-shrink: 0;
}

.trace-controls {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.progress-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn-row {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.msg { font-size: 11px; font-family: monospace; margin-bottom: 6px; }
.error-msg { color: #dc2626; }

/* ══ Кнопки ══ */
.btn {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 0 14px;
  height: 34px;
  border: none;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  white-space: nowrap;
  transition: all .15s;
}
.btn:disabled { opacity: .5; cursor: not-allowed; }

.btn-generate { background: #4f46e5; color: white; width: 100%; justify-content: center; }
.btn-generate:hover:not(:disabled) { background: #4338ca; }

.btn-trace { background: #16a34a; color: white; flex: 1; justify-content: center; }
.btn-trace:hover:not(:disabled) { background: #15803d; }

.btn-step { background: #3b82f6; color: white; padding: 0 10px; }
.btn-step:hover:not(:disabled) { background: #2563eb; }

.btn-secondary { background: #e2e8f0; color: #475569; padding: 0 10px; }
.btn-secondary:hover:not(:disabled) { background: #cbd5e1; }

.icon-btn { padding: 0 8px; }

.btn-stop { background: #ef4444; color: white; padding: 0 10px; }
.btn-stop:hover:not(:disabled) { background: #dc2626; }

.progress-wrap {
  flex: 1;
  height: 5px;
  background: #e2e8f0;
  border-radius: 3px;
  overflow: hidden;
}
.progress-bar {
  height: 100%;
  background: linear-gradient(90deg, #4f46e5, #7c3aed);
  transition: width .2s;
}
.step-label {
  font-size: 11px;
  color: #475569;
  white-space: nowrap;
  font-family: monospace;
}

/* ══ Примеры ══ */
.examples-wrap { position: relative; }
.btn-sm {
  background: white;
  border: 1px solid #e2e8f0;
  color: #475569;
  padding: 3px 9px;
  border-radius: 4px;
  font-size: 11px;
  cursor: pointer;
  white-space: nowrap;
}
.btn-sm:hover { border-color: #4f46e5; color: #4f46e5; }
.dropdown {
  position: absolute; top: 100%; right: 0; margin-top: 4px;
  background: white; border: 1px solid #e2e8f0;
  border-radius: 6px; box-shadow: 0 4px 12px rgba(0,0,0,.12);
  min-width: 130px; z-index: 200;
}
.dropdown-item {
  padding: 7px 12px; font-size: 12px; cursor: pointer; white-space: nowrap;
}
.dropdown-item:hover { background: #f5f7fb; color: #4f46e5; }

/* ══ Блок-схема — колонка ══ */
.col-flowchart { background: #f8fafc; }

.placeholder {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px 24px;
  text-align: center;
}
.placeholder.small { padding: 24px 16px; }
.ph-icon { font-size: 40px; margin-bottom: 12px; opacity: .5; }
.ph-text  { font-size: 13px; font-weight: 600; color: #64748b; margin-bottom: 4px; }
.ph-hint  { font-size: 11px; color: #94a3b8; line-height: 1.5; }

.spinner-wrap {
  flex: 1;
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  gap: 12px; font-size: 13px; color: #64748b;
}
.spinner {
  width: 32px; height: 32px;
  border: 3px solid #e2e8f0;
  border-top-color: #4f46e5;
  border-radius: 50%;
  animation: spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

/* Табы функций */
.func-tabs {
  display: flex;
  align-items: stretch;
  height: 42px;
  flex: 1;
  overflow-x: auto;
  scrollbar-width: none;
}
.func-tabs::-webkit-scrollbar { display: none; }

.func-tab {
  display: flex; align-items: center;
  padding: 0 14px; height: 100%;
  border: none; border-right: 1px solid #e2e8f0;
  background: transparent;
  color: #64748b;
  font-family: 'Courier New', monospace;
  font-size: 11px; font-weight: 500;
  cursor: pointer;
  transition: all .15s;
  position: relative;
  white-space: nowrap;
}
.func-tab:hover { background: #f1f5f9; color: #4f46e5; }
.func-tab.active { background: white; color: #4f46e5; font-weight: 700; }
.func-tab.active::after {
  content: '';
  position: absolute; bottom: 0; left: 0; right: 0;
  height: 2px; background: #4f46e5;
}

.zoom-controls {
  display: flex; align-items: center; gap: 4px; flex-shrink: 0;
}
.zoom-btn {
  width: 24px; height: 24px;
  border: 1px solid #e2e8f0; background: white; color: #475569;
  border-radius: 4px; cursor: pointer; font-size: 13px;
  display: flex; align-items: center; justify-content: center;
  transition: all .15s;
}
.zoom-btn:hover { border-color: #4f46e5; color: #4f46e5; }
.zoom-label {
  font-family: monospace; font-size: 10px; color: #475569;
  min-width: 34px; text-align: center;
}

.svg-scroll {
  flex: 1;
  overflow: auto;
  padding: 20px;
}
.svg-container { width: 100%; min-height: 100%; }
.svg-inner {
  transform-origin: top left;
  transition: transform .2s;
  display: inline-block;
}
.svg-inner :deep(svg) {
  max-width: 100%;
  height: auto;
  display: block;
}

/* ══ Подсветка активного блока ══ */
.svg-inner :deep(.node-active > .shape),
.svg-inner :deep(.node-active > polygon),
.svg-inner :deep(.node-active > ellipse) {
  fill: #fff9c4 !important;
  stroke: #f59e0b !important;
  stroke-width: 3px !important;
}
.svg-inner :deep(.node-active > text) { font-weight: bold; }

/* ══ Колонка переменных ══ */
.col-vars { background: #fafafa; }

.vars-body {
  flex: 1;
  overflow-y: auto;
  overflow-x: hidden;
}
</style>
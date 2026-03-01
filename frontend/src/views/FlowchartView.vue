<script setup>
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { generateFromCode } from '../api/flowchart.js'
import { getSnapshot } from '../api/interpreter.js'

// ─── КОД ───────────────────────────────────────────────────────────
const EXAMPLES = {
  simple_if: {
    label: 'if / else',
    code: `int main() {
    int x = 10;
    if (x > 5) {
        x = x - 1;
    } else {
        x = x + 1;
    }
    return 0;
}`
  },
  dowhile: {
    label: 'Do-While',
    code: `void main() {
    int year = 2014;
    int population = 650;
    do {
        population = (population * 103) / 100;
        year = year + 1;
    } while (year <= 2040);
}`
  },
  minmax: {
    label: 'Min / Max',
    code: `void main() {
    int a, b, min, max;

    if (a < b) {
        min = a;
        max = b;
    } else {
        min = b;
        max = a;
    }
}`
  },
  whilecontinue: {
    label: 'While + Continue',
    code: `void main() {
    int a = 1999;
    while (a < 2030) {
        a = a + 1;
        if (a % 4 == 0)
            continue;
    }
}`
  },
  arrayfor: {
    label: 'Array + For',
    code: `int a1[5] = {1, 2, 3, 7, 8};

void main() {
    int i, s;

    for (i = 0; i < 5; i++)
        if (a1[i] % 2 == 1)
            a1[i] = 1;

    s = 1;
    for (i = 1; i < 5; i++)
        s += a1[i];
}`
  },
  prime: {
    label: 'Простые числа',
    code: `int isPrime(int num) {
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
  },
  factorial: {
    label: 'Факториал',
    code: `int factorial(int n) {
  if(n <= 1) {
    return 1;
  }
  return factorial(n - 1) * n;
}

int main() {
  int res = factorial(4);
  return 0;
}`
  },
  bubble: {
    label: 'Пузырьковая сортировка',
    code: `int main() {
  int arr[5] = {5, 1, 4, 2, 8};
  int i = 0;
  while (i < 4) {
    int j = 0;
    while (j < 4 - i) {
      if (arr[j] > arr[j + 1]) {
        int temp = arr[j];
        arr[j] = arr[j + 1];
        arr[j + 1] = temp;
      }
      j++;
    }
    i++;
  }
  return arr[4];
}`
  }
}

const codeInput = ref(EXAMPLES.simple_if.code)
const selectedExample = ref('simple_if')
const lineNumbers = ref('')
const textareaRef = ref(null)

function updateLineNumbers() {
  const lines = codeInput.value.split('\n').length
  lineNumbers.value = Array.from({ length: lines }, (_, i) => i + 1).join('\n')
}
watch(codeInput, updateLineNumbers)
onMounted(updateLineNumbers)

function loadExample(key) {
  selectedExample.value = key
  codeInput.value = EXAMPLES[key].code
  resetAll()
}

function handleTab(e) {
  if (e.key === 'Tab') {
    e.preventDefault()
    const start = e.target.selectionStart
    const end = e.target.selectionEnd
    codeInput.value = codeInput.value.substring(0, start) + '    ' + codeInput.value.substring(end)
    nextTick(() => {
      e.target.selectionStart = e.target.selectionEnd = start + 4
    })
  }
}

function handleScroll(e) {
  const lineEl = document.querySelector('.line-nums')
  if (lineEl) lineEl.scrollTop = e.target.scrollTop
}

// ─── БЛОК-СХЕМА ────────────────────────────────────────────────────
const svgResult = ref('')
const zoom = ref(1)
const flowLoading = ref(false)
const flowError = ref('')
const flowSuccess = ref(false)

const zoomLabel = computed(() => Math.round(zoom.value * 100) + '%')

async function generateFlowchart() {
  if (!codeInput.value.trim()) return
  flowError.value = ''
  flowSuccess.value = false
  svgResult.value = ''
  flowLoading.value = true
  try {
    const data = await generateFromCode(codeInput.value)
    svgResult.value = data.svg
    flowSuccess.value = true
    zoom.value = 1
  } catch (e) {
    flowError.value = e.message
  } finally {
    flowLoading.value = false
  }
}

function zoomBy(delta) {
  if (delta === 0) { zoom.value = 1; return }
  zoom.value = Math.max(0.25, Math.min(4, zoom.value + delta))
}

function downloadSVG() {
  if (!svgResult.value) return
  const blob = new Blob([svgResult.value], { type: 'image/svg+xml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url; a.download = 'flowchart.svg'; a.click()
  URL.revokeObjectURL(url)
}

// ─── ТРАССИРОВКА ───────────────────────────────────────────────────
const snapshot = ref(null)
const currentStep = ref(0)
const stepsCount = ref(0)
const traceLoading = ref(false)
const traceError = ref(null)
const isExecuted = ref(false)

async function loadSnapshot(step) {
  traceLoading.value = true
  traceError.value = null
  try {
    const data = await getSnapshot(codeInput.value, step)
    snapshot.value = data.snapshot
    currentStep.value = data.current_step ?? step
    stepsCount.value = data.steps_count ?? 0
  } catch (err) {
    traceError.value = err.message
    snapshot.value = null
    isExecuted.value = false
  } finally {
    traceLoading.value = false
  }
}

async function executeTrace() {
  isExecuted.value = false
  currentStep.value = 0
  await loadSnapshot(0)
  if (!traceError.value) isExecuted.value = true
}

function editTrace() {
  isExecuted.value = false
  currentStep.value = 0
  stepsCount.value = 0
  snapshot.value = null
  traceError.value = null
}

async function stepForward() {
  if (currentStep.value < stepsCount.value - 1)
    await loadSnapshot(currentStep.value + 1)
}
async function stepBackward() {
  if (currentStep.value > 0) await loadSnapshot(currentStep.value - 1)
}
async function stepFirst() {
  if (currentStep.value > 0) await loadSnapshot(0)
}
async function stepLast() {
  const last = stepsCount.value - 1
  if (last >= 0 && currentStep.value < last) await loadSnapshot(last)
}

function resetAll() {
  svgResult.value = ''
  flowError.value = ''
  flowSuccess.value = false
  editTrace()
}

// Авто-подсветка текущей строки в редакторе
watch(() => snapshot.value?.line, async (line) => {
  if (line && textareaRef.value) {
    await nextTick()
    const lineHeight = 21
    const scrollTop = (line - 1) * lineHeight - textareaRef.value.clientHeight / 2
    textareaRef.value.scrollTop = Math.max(0, scrollTop)
  }
})

const displayStep = computed(() => stepsCount.value <= 0 ? 0 : Math.min(currentStep.value + 1, stepsCount.value))
const currentLine = computed(() => snapshot.value?.line ?? null)

// Сгруппировать переменные из снимка
const globalVars = computed(() => {
  if (!snapshot.value?.call_stack?.frames) return []
  const g = snapshot.value.call_stack.frames.find(f => f.func_name === 'global')
  return g?.variables ?? []
})
const localFrames = computed(() => {
  if (!snapshot.value?.call_stack?.frames) return []
  return snapshot.value.call_stack.frames.filter(f => f.func_name !== 'global')
})
</script>

<template>
  <div class="fv-root">
    <!-- ══════════════════ ЛЕВАЯ КОЛОНКА: РЕДАКТОР ══════════════════ -->
    <div class="col col-editor">
      <div class="panel-head">
        <span class="head-title">
          <span class="dot dot-blue"></span>
          Редактор C
        </span>
        <div class="examples-wrap">
          <select class="ex-select" :value="selectedExample" @change="e => loadExample(e.target.value)">
            <option v-for="(ex, key) in EXAMPLES" :key="key" :value="key">{{ ex.label }}</option>
          </select>
        </div>
      </div>

      <div class="editor-area">
        <div class="line-nums">{{ lineNumbers }}</div>
        <textarea
          ref="textareaRef"
          class="code-ta"
          :class="{ 'ta-readonly': isExecuted }"
          v-model="codeInput"
          spellcheck="false"
          :readonly="isExecuted"
          @keydown="handleTab"
          @scroll="handleScroll"
        ></textarea>
        <div
          v-if="isExecuted && currentLine"
          class="line-highlight"
          :style="{ top: (currentLine - 1) * 21 + 16 + 'px' }"
        ></div>
      </div>

      <div class="panel-foot">
        <div class="msg err" v-if="flowError || traceError">✗ {{ flowError || traceError }}</div>
        <div class="msg ok" v-else-if="flowSuccess">✓ Блок-схема построена</div>
        <div class="foot-btns">
          <button class="btn btn-schema" :disabled="flowLoading" @click="generateFlowchart">
            <span>{{ flowLoading ? '⟳' : '▶' }}</span>
            {{ flowLoading ? 'Строю…' : 'Блок-схема' }}
          </button>
          <button
            v-if="!isExecuted"
            class="btn btn-trace"
            :disabled="traceLoading"
            @click="executeTrace"
          >▶ Трассировка</button>
          <template v-else>
            <button class="btn btn-edit" @click="editTrace">✏ Редактировать</button>
          </template>
        </div>
      </div>
    </div>

    <!-- ══════════════════ ЦЕНТРАЛЬНАЯ КОЛОНКА: БЛОК-СХЕМА ══════════════════ -->
    <div class="col col-schema">
      <div class="panel-head">
        <span class="head-title">
          <span class="dot dot-violet"></span>
          Блок-схема (ГОСТ 19.701-90)
        </span>
        <div v-if="svgResult" class="zoom-row">
          <button class="zbtn" @click="zoomBy(-0.15)">−</button>
          <span class="zlabel">{{ zoomLabel }}</span>
          <button class="zbtn" @click="zoomBy(0.15)">+</button>
          <button class="zbtn" @click="zoomBy(0)" title="Сбросить">↻</button>
          <button class="zbtn dl" @click="downloadSVG" title="Скачать SVG">⬇</button>
        </div>
      </div>

      <div class="schema-body">
        <div v-if="!svgResult && !flowLoading" class="placeholder">
          <div class="ph-icon">📊</div>
          <div class="ph-text">Нажмите «Блок-схема» чтобы построить</div>
        </div>
        <div v-if="flowLoading" class="spinner-wrap">
          <div class="spinner"></div>
          <div class="spin-text">Генерация…</div>
        </div>
        <div v-if="svgResult && !flowLoading" class="svg-scroll">
          <div class="svg-inner" :style="{ transform: `scale(${zoom})` }" v-html="svgResult"></div>
        </div>
      </div>
    </div>

    <!-- ══════════════════ ПРАВАЯ КОЛОНКА: ТРАССИРОВКА ══════════════════ -->
    <div class="col col-trace">
      <div class="panel-head">
        <span class="head-title">
          <span class="dot dot-green"></span>
          Трассировка
        </span>
        <span v-if="isExecuted" class="step-badge">
          {{ displayStep }} / {{ stepsCount }}
        </span>
      </div>

      <!-- Шаги управления -->
      <div v-if="isExecuted" class="step-controls">
        <button class="sbtn" :disabled="traceLoading || currentStep === 0" @click="stepFirst" title="В начало">⏮</button>
        <button class="sbtn" :disabled="traceLoading || currentStep === 0" @click="stepBackward" title="Назад">‹</button>
        <div class="step-bar">
          <div class="step-fill" :style="{ width: stepsCount > 1 ? (currentStep / (stepsCount-1) * 100) + '%' : '0%' }"></div>
        </div>
        <button class="sbtn" :disabled="traceLoading || currentStep >= stepsCount - 1" @click="stepForward" title="Вперёд">›</button>
        <button class="sbtn" :disabled="traceLoading || currentStep >= stepsCount - 1" @click="stepLast" title="В конец">⏭</button>
      </div>

      <div class="trace-body">
        <!-- Пустое состояние -->
        <div v-if="!isExecuted && !traceLoading && !traceError" class="placeholder">
          <div class="ph-icon">🔍</div>
          <div class="ph-text">Нажмите «Трассировка» чтобы запустить пошаговое выполнение</div>
        </div>

        <div v-if="traceLoading" class="spinner-wrap">
          <div class="spinner"></div>
        </div>

        <div v-if="traceError" class="trace-err">
          <div class="err-title">Ошибка</div>
          <div class="err-body">{{ traceError }}</div>
        </div>

        <template v-if="isExecuted && snapshot">
          <!-- Текущая строка -->
          <div class="info-row" v-if="snapshot.line">
            <span class="info-label">Строка</span>
            <span class="info-val line-val">{{ snapshot.line }}</span>
          </div>

          <!-- Возврат из функции -->
          <div class="ret-banner" v-if="snapshot.function_name && snapshot.return_value !== undefined">
            ↩ return из <strong>{{ snapshot.function_name }}</strong>: {{ snapshot.return_value }}
          </div>

          <!-- Стек вызовов -->
          <div v-if="localFrames.length" class="section">
            <div class="sec-title">Стек вызовов</div>
            <div class="frames">
              <div v-for="(frame, fi) in localFrames" :key="fi" class="frame" :class="{ 'frame-top': fi === localFrames.length - 1 }">
                <div class="frame-name">{{ frame.func_name }}()</div>
                <div v-if="frame.variables && frame.variables.length" class="vars">
                  <div v-for="v in frame.variables" :key="v.name" class="var-row">
                    <span class="vname">{{ v.name }}</span>
                    <span class="vtype">{{ v.type }}</span>
                    <span class="vval" :class="{ uninit: v.value === null || v.value === undefined }">
                      {{ v.value !== null && v.value !== undefined ? v.value : '?' }}
                    </span>
                  </div>
                </div>
                <div v-else class="no-vars">нет локальных переменных</div>
              </div>
            </div>
          </div>

          <!-- Глобальные переменные -->
          <div v-if="globalVars.length" class="section">
            <div class="sec-title">Глобальные</div>
            <div class="vars">
              <div v-for="v in globalVars" :key="v.name" class="var-row">
                <span class="vname">{{ v.name }}</span>
                <span class="vtype">{{ v.type }}</span>
                <span class="vval" :class="{ uninit: v.value === null || v.value === undefined }">
                  {{ v.value !== null && v.value !== undefined ? v.value : '?' }}
                </span>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ─── КОРЕНЬ ─────────────────────────────────────────────────────── */
.fv-root {
  display: grid;
  grid-template-columns: 380px 1fr 280px;
  height: 100%;
  overflow: hidden;
  background: #f1f4f9;
  gap: 0;
}

/* ─── КОЛОНКИ ────────────────────────────────────────────────────── */
.col {
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: #fff;
}
.col-editor  { border-right: 1px solid #e2e8f0; }
.col-schema  { border-right: 1px solid #e2e8f0; background: #fafbfc; }
.col-trace   {}

/* ─── ЗАГОЛОВОК ПАНЕЛИ ───────────────────────────────────────────── */
.panel-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 14px;
  height: 42px;
  border-bottom: 1px solid #e2e8f0;
  background: #fff;
  flex-shrink: 0;
}
.head-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: .06em;
  text-transform: uppercase;
  color: #475569;
  font-family: 'Courier New', monospace;
}
.dot {
  width: 8px; height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.dot-blue   { background: #3b82f6; }
.dot-violet { background: #7c3aed; }
.dot-green  { background: #10b981; }

/* ─── ВЫБОР ПРИМЕРА ──────────────────────────────────────────────── */
.ex-select {
  border: 1px solid #e2e8f0;
  border-radius: 5px;
  padding: 4px 8px;
  font-size: 11px;
  color: #475569;
  background: #fff;
  cursor: pointer;
  outline: none;
  max-width: 160px;
}
.ex-select:focus { border-color: #3b82f6; }

/* ─── РЕДАКТОР ───────────────────────────────────────────────────── */
.editor-area {
  flex: 1;
  position: relative;
  overflow: hidden;
  display: flex;
}
.line-nums {
  width: 42px;
  padding: 16px 8px 16px 0;
  text-align: right;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.75;
  color: #94a3b8;
  background: #f8fafc;
  border-right: 1px solid #e2e8f0;
  user-select: none;
  overflow: hidden;
  white-space: pre;
  flex-shrink: 0;
}
.code-ta {
  flex: 1;
  padding: 16px 12px;
  font-family: 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.75;
  border: none;
  outline: none;
  resize: none;
  background: #fff;
  color: #0f172a;
  overflow-y: auto;
}
.code-ta.ta-readonly {
  background: #f8fafc;
  color: #475569;
  cursor: default;
}
.line-highlight {
  position: absolute;
  left: 42px;
  right: 0;
  height: 21px;
  background: rgba(251, 191, 36, 0.25);
  pointer-events: none;
  border-left: 3px solid #f59e0b;
}

/* ─── ФУТЕР РЕДАКТОРА ────────────────────────────────────────────── */
.panel-foot {
  border-top: 1px solid #e2e8f0;
  padding: 10px 14px;
  background: #fff;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.msg { font-size: 11px; font-family: 'Courier New', monospace; }
.msg.err { color: #dc2626; }
.msg.ok  { color: #16a34a; }
.foot-btns {
  display: flex;
  gap: 8px;
}
.btn {
  flex: 1;
  height: 34px;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: filter .15s, transform .1s;
}
.btn:hover:not(:disabled) { filter: brightness(.92); transform: translateY(-1px); }
.btn:disabled { opacity: .5; cursor: not-allowed; }
.btn-schema { background: #7c3aed; color: #fff; }
.btn-trace  { background: #059669; color: #fff; }
.btn-edit   { background: #f59e0b; color: #fff; }

/* ─── БЛОК-СХЕМА ─────────────────────────────────────────────────── */
.schema-body {
  flex: 1;
  overflow: hidden;
  position: relative;
}
.svg-scroll {
  width: 100%;
  height: 100%;
  overflow: auto;
  padding: 20px;
}
.svg-inner {
  transform-origin: top left;
  transition: transform .2s;
  display: inline-block;
}
.svg-inner :deep(svg) { display: block; }

/* Zoom */
.zoom-row { display: flex; align-items: center; gap: 6px; }
.zbtn {
  width: 26px; height: 26px;
  border: 1px solid #e2e8f0;
  background: #fff;
  border-radius: 4px;
  font-size: 13px;
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  color: #475569;
  transition: all .15s;
}
.zbtn:hover { border-color: #7c3aed; color: #7c3aed; }
.zbtn.dl { margin-left: 4px; }
.zlabel {
  font-family: 'Courier New', monospace;
  font-size: 11px;
  color: #64748b;
  min-width: 38px;
  text-align: center;
}

/* ─── ТРАССИРОВКА ────────────────────────────────────────────────── */
.step-badge {
  font-family: 'Courier New', monospace;
  font-size: 11px;
  background: #ecfdf5;
  color: #059669;
  padding: 3px 8px;
  border-radius: 20px;
  font-weight: 700;
  border: 1px solid #a7f3d0;
}

.step-controls {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  border-bottom: 1px solid #e2e8f0;
  background: #f8fafc;
  flex-shrink: 0;
}
.sbtn {
  width: 28px; height: 28px;
  border: 1px solid #e2e8f0;
  background: #fff;
  border-radius: 5px;
  font-size: 14px;
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  color: #475569;
  transition: all .15s;
  flex-shrink: 0;
}
.sbtn:hover:not(:disabled) { border-color: #10b981; color: #10b981; background: #f0fdf4; }
.sbtn:disabled { opacity: .35; cursor: not-allowed; }

.step-bar {
  flex: 1;
  height: 4px;
  background: #e2e8f0;
  border-radius: 2px;
  overflow: hidden;
}
.step-fill {
  height: 100%;
  background: #10b981;
  transition: width .2s;
  border-radius: 2px;
}

.trace-body {
  flex: 1;
  overflow-y: auto;
  padding: 12px;
}

/* Секции */
.section { margin-bottom: 12px; }
.sec-title {
  font-size: 10px;
  font-weight: 700;
  letter-spacing: .08em;
  text-transform: uppercase;
  color: #94a3b8;
  margin-bottom: 6px;
  padding-bottom: 4px;
  border-bottom: 1px solid #f1f5f9;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
  padding: 6px 10px;
  background: #fffbeb;
  border: 1px solid #fde68a;
  border-radius: 6px;
}
.info-label { font-size: 11px; color: #92400e; font-weight: 600; }
.info-val { font-family: 'Courier New', monospace; font-size: 13px; font-weight: 700; }
.line-val { color: #d97706; }

.ret-banner {
  padding: 7px 10px;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
  border-radius: 6px;
  font-size: 12px;
  color: #166534;
  margin-bottom: 10px;
}

/* Стек */
.frames { display: flex; flex-direction: column; gap: 6px; }
.frame {
  border: 1px solid #e2e8f0;
  border-radius: 7px;
  overflow: hidden;
  background: #fff;
}
.frame-top { border-color: #bfdbfe; }
.frame-name {
  padding: 5px 10px;
  font-size: 11px;
  font-weight: 700;
  font-family: 'Courier New', monospace;
  background: #f8fafc;
  color: #3b82f6;
  border-bottom: 1px solid #e2e8f0;
}
.frame-top .frame-name { background: #eff6ff; border-color: #bfdbfe; }

.vars { padding: 6px 4px; }
.var-row {
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 4px;
  align-items: center;
  padding: 3px 6px;
  border-radius: 4px;
  font-size: 11px;
  transition: background .1s;
}
.var-row:hover { background: #f8fafc; }
.vname {
  font-family: 'Courier New', monospace;
  font-weight: 600;
  color: #1e293b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.vtype {
  font-size: 10px;
  color: #94a3b8;
  background: #f1f5f9;
  padding: 1px 5px;
  border-radius: 3px;
  font-family: 'Courier New', monospace;
}
.vval {
  font-family: 'Courier New', monospace;
  font-weight: 700;
  color: #0f172a;
  text-align: right;
  min-width: 30px;
}
.vval.uninit { color: #94a3b8; font-style: italic; }

.no-vars { padding: 4px 8px; font-size: 11px; color: #94a3b8; }

/* ─── ОБЩИЕ ──────────────────────────────────────────────────────── */
.placeholder {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 24px;
  text-align: center;
}
.ph-icon { font-size: 40px; opacity: .4; }
.ph-text { font-size: 13px; color: #64748b; max-width: 200px; line-height: 1.5; }

.spinner-wrap {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
}
.spinner {
  width: 32px; height: 32px;
  border: 3px solid #e2e8f0;
  border-top-color: #7c3aed;
  border-radius: 50%;
  animation: spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.spin-text { font-size: 12px; color: #94a3b8; }

.trace-err {
  padding: 12px;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 7px;
  margin: 4px 0;
}
.err-title { font-size: 12px; font-weight: 700; color: #dc2626; margin-bottom: 4px; }
.err-body  { font-size: 11px; color: #b91c1c; font-family: 'Courier New', monospace; white-space: pre-wrap; word-break: break-all; }
</style>

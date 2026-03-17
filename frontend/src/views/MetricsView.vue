<template>
  <div class="metrics-root">

    <!-- Левая панель: редактор -->
    <div class="panel panel-editor">
      <div class="panel-head">
        <span>● C Code Editor</span>
        <div class="examples-wrap">
          <span class="label">Примеры</span>
          <select @change="e => loadExample(e.target.value)">
            <option value="">— выбрать —</option>
            <option v-for="(ex, key) in EXAMPLES" :key="key" :value="key">{{ ex.label }}</option>
          </select>
        </div>
      </div>
      <div class="editor-area">
        <div class="line-nums">{{ lineNumbers }}</div>
        <textarea class="code-ta" v-model="code" spellcheck="false"
          @keydown="handleTab" @scroll="syncScroll" @input="updateLineNums"></textarea>
      </div>
      <div class="panel-foot">
        <div class="msg err" v-if="error">✗ {{ error }}</div>
        <div class="msg limit-warn" v-else-if="limitWarning">
          ⚠ {{ limitWarning }}
        </div>
        <div class="msg ok" v-else-if="result">✓ Метрики подсчитаны</div>
        <button class="btn-calc" :disabled="loading" @click="calculate">
          {{ loading ? '⟳ Считаем…' : '▶ Подсчитать метрики' }}
        </button>
      </div>
    </div>

    <!-- Правая панель -->
    <div class="panel panel-results">
      <div class="tabs-head">
        <button class="tab-btn" :class="{ active: activeTab === 'result' }" @click="activeTab = 'result'">
          📈 Результат
        </button>
        <button class="tab-btn" :class="{ active: activeTab === 'history' }" @click="switchHistory">
          🕓 История
        </button>
        <div v-if="activeTab === 'result' && result" class="summary-badges">
          <span class="badge">Функций: {{ result.functionCount }}</span>
          <span class="badge">Глоб. переменных: {{ result.globalVarCount }}</span>
        </div>
      </div>

      <!-- Вкладка Результат -->
      <template v-if="activeTab === 'result'">
        <div v-if="!result && !loading" class="empty-state">
          <div class="empty-icon">📊</div>
          <div>Введите код и нажмите «Подсчитать метрики»</div>
        </div>
        <div v-if="loading" class="empty-state">
          <div class="spinner"></div>
          <div>Анализируем код…</div>
        </div>
        <div v-if="result && !loading" class="results-body">
          <div v-if="result.functions.length > 1" class="chart-section">
            <div class="section-title">Цикломатическая сложность</div>
            <div class="bar-chart">
              <div v-for="fn in result.functions" :key="fn.functionName + '_bar'" class="bar-row">
                <div class="bar-name">{{ fn.functionName }}</div>
                <div class="bar-track">
                  <div class="bar-fill" :class="ccClass(fn.cyclomaticComplexity)"
                       :style="{ width: ccWidth(fn.cyclomaticComplexity) }"></div>
                </div>
                <div class="bar-val" :class="ccClass(fn.cyclomaticComplexity)">{{ fn.cyclomaticComplexity }}</div>
              </div>
            </div>
            <div class="legend">
              <span class="leg cc-low">1–5 низкая</span>
              <span class="leg cc-medium">6–10 умеренная</span>
              <span class="leg cc-high">11–20 высокая</span>
              <span class="leg cc-critical">20+ критическая</span>
            </div>
          </div>
          <div class="section-title" style="margin-top: 16px">Детали по функциям</div>
          <div class="cards-grid">
            <div v-for="fn in result.functions" :key="fn.functionName" class="fn-card">
              <div class="fn-header" :class="ccClass(fn.cyclomaticComplexity)">
                <span class="fn-name">{{ fn.functionName }}</span>
                <span class="fn-cc">CC: {{ fn.cyclomaticComplexity }}</span>
              </div>
              <div class="fn-body">
                <div class="metric-row"><span class="m-label">LOC</span><span class="m-val">{{ fn.loc }}</span></div>
                <div class="metric-row"><span class="m-label">Параметры</span><span class="m-val">{{ fn.parameterCount }}</span></div>
                <div class="metric-row">
                  <span class="m-label">Макс. вложенность</span>
                  <span class="m-val" :class="fn.maxNestingDepth > 4 ? 'warn' : ''">{{ fn.maxNestingDepth }}</span>
                </div>
                <div class="metric-row"><span class="m-label">Вызовов функций</span><span class="m-val">{{ fn.callCount }}</span></div>
                <div class="metric-row"><span class="m-label">Return</span><span class="m-val">{{ fn.returnCount }}</span></div>
                <div v-if="fn.gotoCount > 0" class="metric-row">
                  <span class="m-label" style="color:#b45309">goto (антипаттерн)</span>
                  <span class="m-val" style="color:#dc2626">{{ fn.gotoCount }}</span>
                </div>
                <div class="cc-bar-wrap">
                  <div class="cc-bar-fill" :class="ccClass(fn.cyclomaticComplexity)"
                       :style="{ width: ccWidth(fn.cyclomaticComplexity) }"></div>
                </div>
                <div class="cc-label">{{ ccLabel(fn.cyclomaticComplexity) }}</div>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- Вкладка История -->
      <template v-if="activeTab === 'history'">
        <div v-if="historyLoading" class="empty-state">
          <div class="spinner"></div>
          <div>Загружаем историю…</div>
        </div>
        <div v-else-if="history.length === 0" class="empty-state">
          <div class="empty-icon">🕓</div>
          <div>История пуста — подсчитайте метрики хотя бы раз</div>
        </div>
        <div v-else class="results-body">

          <!-- Шапка с кол-вом и кнопками массового удаления -->
          <div class="history-toolbar">
            <div class="history-count-wrap">
              <span class="history-count">{{ history.length }} / {{ MAX_RECORDS }} записей</span>
              <div class="count-bar-track">
                <div class="count-bar-fill" :class="countBarClass"
                     :style="{ width: Math.min(100, history.length / MAX_RECORDS * 100) + '%' }"></div>
              </div>
            </div>
            <div class="toolbar-actions" v-if="selectedIds.size > 0">
              <span class="selected-label">Выбрано: {{ selectedIds.size }}</span>
              <button class="btn-del-selected" @click="deleteSelected">🗑 Удалить выбранные</button>
              <button class="btn-cancel-sel" @click="selectedIds.clear()">Отмена</button>
            </div>
          </div>

          <!-- Предупреждение о лимите -->
          <div v-if="history.length >= MAX_RECORDS" class="limit-banner">
            ⚠ Достигнут лимит {{ MAX_RECORDS }} записей. Удалите часть истории, чтобы можно было сохранять новые метрики.
          </div>

          <div v-for="(group, date) in groupedHistory" :key="date" class="history-group">
            <!-- Заголовок дня с кнопкой удаления дня -->
            <div class="history-date">
              <div class="date-left">
                <input type="checkbox" class="day-checkbox"
                       :checked="isDaySelected(group)"
                       @change="toggleDaySelection(group)" />
                <span>📅 {{ date }}</span>
                <span class="day-count">{{ group.length }} записей</span>
              </div>
              <button class="btn-del-day" @click="deleteDayGroup(group, date)" title="Удалить все за этот день">
                🗑 Удалить день
              </button>
            </div>

            <div class="history-rows">
              <div v-for="entry in group" :key="entry.id"
                   class="history-row" :class="{ expanded: expandedId === entry.id, selected: selectedIds.has(entry.id) }">
                <div class="history-row-main">
                  <div class="history-left">
                    <input type="checkbox" class="row-checkbox"
                           :checked="selectedIds.has(entry.id)"
                           @change="toggleSelect(entry.id)"
                           @click.stop />
                    <span class="history-time">{{ formatTime(entry.createdAt) }}</span>
                    <span class="fn-pill" :class="'pill-' + ccClass(entry.cyclomaticComplexity)"
                          @click="toggleExpand(entry.id)">
                      ƒ {{ entry.functionName }}
                    </span>
                  </div>
                  <div class="history-right" @click="toggleExpand(entry.id)">
                    <span class="metric-pill">CC <strong>{{ entry.cyclomaticComplexity }}</strong></span>
                    <span class="metric-pill">LOC <strong>{{ entry.loc }}</strong></span>
                    <span class="metric-pill" :class="entry.maxNestingDepth > 4 ? 'pill-warn' : ''">
                      Вложенность <strong>{{ entry.maxNestingDepth }}</strong>
                    </span>
                    <span class="expand-arrow">{{ expandedId === entry.id ? '▲' : '▼' }}</span>
                    <button class="delete-btn" @click.stop="deleteEntry(entry.id)" title="Удалить">✕</button>
                  </div>
                </div>

                <!-- Раскрывающаяся карточка -->
                <div v-if="expandedId === entry.id" class="history-detail">
                  <div class="detail-header" :class="ccClass(entry.cyclomaticComplexity)">
                    <span>ƒ {{ entry.functionName }}</span>
                    <span>{{ ccLabel(entry.cyclomaticComplexity) }}</span>
                  </div>
                  <div class="detail-grid">
                    <div class="detail-item">
                      <div class="detail-label">Цикломатич. сложность</div>
                      <div class="detail-value" :class="'cc-text-' + ccClass(entry.cyclomaticComplexity)">{{ entry.cyclomaticComplexity }}</div>
                    </div>
                    <div class="detail-item">
                      <div class="detail-label">Строк кода (LOC)</div>
                      <div class="detail-value">{{ entry.loc }}</div>
                    </div>
                    <div class="detail-item">
                      <div class="detail-label">Параметры функции</div>
                      <div class="detail-value">{{ entry.parameterCount }}</div>
                    </div>
                    <div class="detail-item">
                      <div class="detail-label">Макс. вложенность</div>
                      <div class="detail-value" :class="entry.maxNestingDepth > 4 ? 'val-warn' : ''">{{ entry.maxNestingDepth }}</div>
                    </div>
                    <div class="detail-item">
                      <div class="detail-label">Вызовов функций</div>
                      <div class="detail-value">{{ entry.callCount }}</div>
                    </div>
                    <div class="detail-item">
                      <div class="detail-label">Return-операторов</div>
                      <div class="detail-value">{{ entry.returnCount }}</div>
                    </div>
                    <div v-if="entry.gotoCount > 0" class="detail-item detail-item-warn">
                      <div class="detail-label" style="color:#b45309">goto (антипаттерн)</div>
                      <div class="detail-value" style="color:#dc2626">{{ entry.gotoCount }}</div>
                    </div>
                  </div>
                  <div class="detail-cc-bar-wrap">
                    <div class="detail-cc-bar" :class="ccClass(entry.cyclomaticComplexity)"
                         :style="{ width: ccWidth(entry.cyclomaticComplexity) }"></div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { calculateMetrics } from '../api/metrics.js'

const METRICS_URL = import.meta.env.VITE_METRICS_SERVICE_URL || 'http://localhost:8085'
const MAX_RECORDS = 50

async function fetchHistory() {
  const res = await fetch(`${METRICS_URL}/api/metrics/history`, { credentials: 'include' })
  if (!res.ok) return []
  return res.json()
}

const EXAMPLES = {
  factorial: {
    label: 'Факториал',
    code: `int factorial(int n) {\n  if (n <= 1) {\n    return 1;\n  }\n  return factorial(n - 1) * n;\n}\n\nint main() {\n  int res = factorial(4);\n  return 0;\n}`
  },
  bubble: {
    label: 'Пузырьковая сортировка',
    code: `int main() {\n  int arr[5] = {5, 1, 4, 2, 8};\n  int i = 0;\n  while (i < 4) {\n    int j = 0;\n    while (j < 4 - i) {\n      if (arr[j] > arr[j + 1]) {\n        int temp = arr[j];\n        arr[j] = arr[j + 1];\n        arr[j + 1] = temp;\n      }\n      j++;\n    }\n    i++;\n  }\n  return 0;\n}`
  },
  prime: {
    label: 'Простые числа',
    code: `int isPrime(int num) {\n  int del = 2;\n  while (del < num) {\n    if (num % del == 0) {\n      return 0;\n    }\n    del++;\n  }\n  return 1;\n}\n\nvoid main() {\n  int num = 20;\n  while (1) {\n    if (isPrime(num)) {\n      break;\n    }\n    num++;\n  }\n}`
  }
}

const code           = ref(EXAMPLES.factorial.code)
const loading        = ref(false)
const error          = ref('')
const limitWarning   = ref('')
const result         = ref(null)
const lineNumbers    = ref('')
const activeTab      = ref('result')
const history        = ref([])
const historyLoading = ref(false)
const expandedId     = ref(null)
const selectedIds    = ref(new Set())

function updateLineNums() {
  const n = code.value.split('\n').length
  lineNumbers.value = Array.from({ length: n }, (_, i) => i + 1).join('\n')
}
onMounted(updateLineNums)

function loadExample(key) {
  if (!key) return
  code.value = EXAMPLES[key].code
  result.value = null
  error.value = ''
  limitWarning.value = ''
  updateLineNums()
}

function handleTab(e) {
  if (e.key === 'Tab') {
    e.preventDefault()
    const s = e.target.selectionStart
    code.value = code.value.substring(0, s) + '    ' + code.value.substring(e.target.selectionEnd)
    setTimeout(() => { e.target.selectionStart = e.target.selectionEnd = s + 4 }, 0)
  }
}

function syncScroll(e) {
  const ln = e.target.closest('.editor-area')?.querySelector('.line-nums')
  if (ln) ln.scrollTop = e.target.scrollTop
}

async function calculate() {
  if (!code.value.trim()) return
  loading.value = true
  error.value = ''
  limitWarning.value = ''
  result.value = null
  try {
    const res = await fetch(`${METRICS_URL}/api/metrics/calculate`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code: code.value })
    })
    const data = await res.json()
    if (!res.ok) {
      if (data.limitExceeded) {
        limitWarning.value = data.error
      } else {
        error.value = data.error || `HTTP ${res.status}`
      }
    } else {
      result.value = data
    }
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function switchHistory() {
  activeTab.value = 'history'
  expandedId.value = null
  selectedIds.value = new Set()
  historyLoading.value = true
  try {
    history.value = await fetchHistory()
  } catch {
    history.value = []
  } finally {
    historyLoading.value = false
  }
}

function toggleExpand(id) {
  expandedId.value = expandedId.value === id ? null : id
}

function toggleSelect(id) {
  const s = new Set(selectedIds.value)
  s.has(id) ? s.delete(id) : s.add(id)
  selectedIds.value = s
}

function isDaySelected(group) {
  return group.every(e => selectedIds.value.has(e.id))
}

function toggleDaySelection(group) {
  const s = new Set(selectedIds.value)
  if (isDaySelected(group)) {
    group.forEach(e => s.delete(e.id))
  } else {
    group.forEach(e => s.add(e.id))
  }
  selectedIds.value = s
}

async function deleteEntry(id) {
  try {
    const res = await fetch(`${METRICS_URL}/api/metrics/${id}`, { method: 'DELETE', credentials: 'include' })
    if (res.ok) {
      history.value = history.value.filter(e => e.id !== id)
      if (expandedId.value === id) expandedId.value = null
      const s = new Set(selectedIds.value); s.delete(id); selectedIds.value = s
    }
  } catch (e) { console.error('Delete failed', e) }
}

async function deleteSelected() {
  const ids = [...selectedIds.value]
  if (!ids.length) return
  try {
    const res = await fetch(`${METRICS_URL}/api/metrics/batch`, {
      method: 'DELETE',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ids })
    })
    if (res.ok) {
      const idSet = new Set(ids)
      history.value = history.value.filter(e => !idSet.has(e.id))
      selectedIds.value = new Set()
      expandedId.value = null
    }
  } catch (e) { console.error('Batch delete failed', e) }
}

async function deleteDayGroup(group, dateLabel) {
  if (!confirm(`Удалить все ${group.length} записей за ${dateLabel}?`)) return
  const ids = group.map(e => e.id)
  try {
    const res = await fetch(`${METRICS_URL}/api/metrics/batch`, {
      method: 'DELETE',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ ids })
    })
    if (res.ok) {
      const idSet = new Set(ids)
      history.value = history.value.filter(e => !idSet.has(e.id))
      const s = new Set(selectedIds.value)
      ids.forEach(id => s.delete(id))
      selectedIds.value = s
    }
  } catch (e) { console.error('Day delete failed', e) }
}

function parseDate(raw) {
  if (!raw) return null
  if (Array.isArray(raw)) return new Date(raw[0], raw[1] - 1, raw[2], raw[3] ?? 0, raw[4] ?? 0, raw[5] ?? 0)
  return new Date(raw)
}

function formatDate(raw) {
  const d = parseDate(raw)
  if (!d) return '—'
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' })
}

function formatTime(raw) {
  const d = parseDate(raw)
  if (!d) return '—'
  return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}

const groupedHistory = computed(() => {
  const groups = {}
  for (const entry of history.value) {
    const date = formatDate(entry.createdAt)
    if (!groups[date]) groups[date] = []
    groups[date].push(entry)
  }
  return groups
})

const countBarClass = computed(() => {
  const ratio = history.value.length / MAX_RECORDS
  if (ratio >= 1) return 'bar-full'
  if (ratio >= 0.8) return 'bar-warn'
  return 'bar-ok'
})

function ccClass(cc) {
  if (cc <= 5)  return 'cc-low'
  if (cc <= 10) return 'cc-medium'
  if (cc <= 20) return 'cc-high'
  return 'cc-critical'
}
function ccLabel(cc) {
  if (cc <= 5)  return 'Низкая сложность'
  if (cc <= 10) return 'Умеренная сложность'
  if (cc <= 20) return 'Высокая сложность'
  return 'Критическая сложность'
}
function ccWidth(cc) {
  return Math.min(100, (cc / 20) * 100) + '%'
}
</script>

<style scoped>
.metrics-root {
  display: grid; grid-template-columns: 420px 1fr;
  height: 100%; overflow: hidden; background: #f5f7fb;
}
.panel { display: flex; flex-direction: column; overflow: hidden; background: #fff; }
.panel-editor { border-right: 1px solid #e2e8f0; }

.panel-head {
  display: flex; align-items: center; justify-content: space-between;
  padding: 0 14px; height: 40px; border-bottom: 1px solid #e2e8f0;
  background: #f8fafc; font-size: 12px; font-weight: 600; color: #475569; flex-shrink: 0;
}
.examples-wrap { display: flex; align-items: center; gap: 6px; }
.examples-wrap .label { font-size: 11px; color: #94a3b8; }
.examples-wrap select {
  border: 1px solid #e2e8f0; border-radius: 4px; padding: 3px 6px;
  font-size: 11px; color: #475569; background: #fff; cursor: pointer; outline: none;
}
.editor-area { flex: 1; display: flex; overflow: hidden; }
.line-nums {
  width: 40px; padding: 14px 6px 14px 0; text-align: right;
  font-family: 'Courier New', monospace; font-size: 12px; line-height: 1.75;
  color: #94a3b8; background: #f8fafc; border-right: 1px solid #e2e8f0;
  user-select: none; overflow: hidden; white-space: pre; flex-shrink: 0;
}
.code-ta {
  flex: 1; padding: 14px 12px; font-family: 'Courier New', monospace;
  font-size: 12px; line-height: 1.75; border: none; outline: none;
  resize: none; background: #fff; color: #0f172a; overflow-y: auto;
}
.panel-foot {
  border-top: 1px solid #e2e8f0; padding: 10px 14px;
  display: flex; flex-direction: column; gap: 8px; flex-shrink: 0;
}
.msg { font-size: 11px; font-family: 'Courier New', monospace; }
.msg.err { color: #dc2626; }
.msg.ok  { color: #16a34a; }
.msg.limit-warn { color: #d97706; font-size: 11px; line-height: 1.4; }
.btn-calc {
  width: 100%; height: 36px; background: #6366f1; color: #fff;
  border: none; border-radius: 6px; font-size: 13px; font-weight: 600;
  cursor: pointer; transition: filter .15s;
}
.btn-calc:hover:not(:disabled) { filter: brightness(.9); }
.btn-calc:disabled { opacity: .5; cursor: not-allowed; }

/* Вкладки */
.tabs-head {
  display: flex; align-items: center; gap: 2px; padding: 0 10px;
  height: 40px; border-bottom: 1px solid #e2e8f0; background: #f8fafc; flex-shrink: 0;
}
.tab-btn {
  height: 28px; padding: 0 12px; border: 1px solid transparent; border-radius: 5px;
  background: transparent; color: #64748b; font-size: 12px; font-weight: 600;
  cursor: pointer; transition: all .15s;
}
.tab-btn:hover { background: #f1f5f9; color: #6366f1; }
.tab-btn.active { background: #6366f1; color: #fff; border-color: #6366f1; }
.summary-badges { display: flex; gap: 8px; margin-left: auto; }
.badge {
  background: #f1f5f9; border: 1px solid #e2e8f0; border-radius: 20px;
  padding: 2px 10px; font-size: 11px; color: #475569; font-weight: 600;
}

/* Результат */
.results-body { flex: 1; overflow-y: auto; padding: 16px; }
.empty-state {
  flex: 1; display: flex; flex-direction: column; align-items: center;
  justify-content: center; gap: 12px; color: #94a3b8; font-size: 13px;
}
.empty-icon { font-size: 40px; opacity: .5; }
.spinner {
  width: 32px; height: 32px; border: 3px solid #e2e8f0;
  border-top-color: #6366f1; border-radius: 50%; animation: spin .7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
.section-title {
  font-size: 11px; font-weight: 700; letter-spacing: .07em;
  text-transform: uppercase; color: #94a3b8; margin-bottom: 10px;
}
.chart-section { margin-bottom: 8px; }
.bar-chart { display: flex; flex-direction: column; gap: 7px; }
.bar-row { display: flex; align-items: center; gap: 10px; }
.bar-name {
  width: 130px; font-size: 12px; color: #334155; text-align: right;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis; flex-shrink: 0;
}
.bar-track { flex: 1; height: 18px; background: #f1f5f9; border-radius: 4px; overflow: hidden; }
.bar-fill  { height: 100%; border-radius: 4px; transition: width .4s; }
.bar-val   { width: 24px; font-size: 12px; font-weight: 700; text-align: right; flex-shrink: 0; }
.legend { display: flex; gap: 10px; margin-top: 8px; flex-wrap: wrap; }
.leg { font-size: 10px; padding: 2px 8px; border-radius: 10px; font-weight: 600; }
.cards-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 12px; }
.fn-card { border: 1px solid #e2e8f0; border-radius: 10px; overflow: hidden; background: #fff; box-shadow: 0 1px 3px rgba(0,0,0,.05); }
.fn-header { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; }
.fn-name { font-family: 'Courier New', monospace; font-size: 13px; font-weight: 700; color: #fff; }
.fn-cc   { font-size: 12px; font-weight: 700; color: rgba(255,255,255,.9); }
.fn-body { padding: 10px 12px; }
.metric-row { display: flex; justify-content: space-between; font-size: 12px; padding: 3px 0; border-bottom: 1px solid #f8fafc; }
.m-label { color: #64748b; }
.m-val   { font-weight: 600; color: #1e293b; }
.m-val.warn { color: #d97706; }
.cc-bar-wrap { height: 5px; background: #f1f5f9; border-radius: 3px; margin-top: 10px; overflow: hidden; }
.cc-bar-fill { height: 100%; border-radius: 3px; transition: width .3s; }
.cc-label    { font-size: 10px; color: #94a3b8; margin-top: 4px; }

/* Тулбар истории */
.history-toolbar {
  display: flex; align-items: center; justify-content: space-between;
  margin-bottom: 12px; gap: 12px; flex-wrap: wrap;
}
.history-count-wrap { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 160px; }
.history-count { font-size: 11px; color: #475569; font-weight: 600; white-space: nowrap; }
.count-bar-track { flex: 1; height: 6px; background: #e2e8f0; border-radius: 3px; overflow: hidden; }
.count-bar-fill  { height: 100%; border-radius: 3px; transition: width .4s; }
.bar-ok   { background: #22c55e; }
.bar-warn { background: #f59e0b; }
.bar-full { background: #ef4444; }

.toolbar-actions { display: flex; align-items: center; gap: 8px; }
.selected-label { font-size: 12px; color: #6366f1; font-weight: 600; }
.btn-del-selected {
  height: 28px; padding: 0 10px; background: #ef4444; color: #fff;
  border: none; border-radius: 5px; font-size: 11px; font-weight: 600; cursor: pointer;
}
.btn-del-selected:hover { background: #dc2626; }
.btn-cancel-sel {
  height: 28px; padding: 0 10px; background: #f1f5f9; color: #475569;
  border: 1px solid #e2e8f0; border-radius: 5px; font-size: 11px; cursor: pointer;
}

/* Баннер лимита */
.limit-banner {
  background: #fff7ed; border: 1px solid #fed7aa; border-radius: 8px;
  padding: 10px 14px; font-size: 12px; color: #c2410c; margin-bottom: 12px;
  font-weight: 500;
}

/* История */
.history-group { margin-bottom: 20px; }
.history-date {
  display: flex; align-items: center; justify-content: space-between;
  font-size: 11px; font-weight: 700; letter-spacing: .06em;
  color: #475569; margin-bottom: 8px; padding: 6px 10px;
  background: #f1f5f9; border-radius: 6px;
}
.date-left { display: flex; align-items: center; gap: 8px; }
.day-count { font-size: 10px; color: #94a3b8; font-weight: 400; }
.day-checkbox { cursor: pointer; width: 14px; height: 14px; }
.btn-del-day {
  height: 24px; padding: 0 8px; background: #fff; color: #ef4444;
  border: 1px solid #fecaca; border-radius: 4px; font-size: 11px;
  cursor: pointer; transition: all .15s; white-space: nowrap;
}
.btn-del-day:hover { background: #fee2e2; }

.history-rows { display: flex; flex-direction: column; gap: 6px; }
.history-row {
  border: 1px solid #e2e8f0; border-radius: 10px; overflow: hidden;
  background: #fff; transition: box-shadow .15s, border-color .15s;
}
.history-row:hover { border-color: #a5b4fc; box-shadow: 0 2px 8px rgba(99,102,241,.1); }
.history-row.expanded { border-color: #6366f1; box-shadow: 0 2px 12px rgba(99,102,241,.15); }
.history-row.selected { background: #f5f3ff; border-color: #a5b4fc; }

.history-row-main {
  display: flex; align-items: center; justify-content: space-between;
  padding: 10px 14px; gap: 12px;
}
.history-left  { display: flex; align-items: center; gap: 8px; }
.history-right { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; cursor: pointer; }

.row-checkbox { cursor: pointer; width: 14px; height: 14px; flex-shrink: 0; }

.history-time {
  font-family: monospace; font-size: 11px; color: #94a3b8;
  background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 4px;
  padding: 2px 6px; flex-shrink: 0;
}

/* fn-pill с белым текстом — отдельные классы */
.fn-pill {
  display: inline-flex; align-items: center; padding: 3px 10px; border-radius: 20px;
  font-family: 'Courier New', monospace; font-size: 12px; font-weight: 700;
  color: #fff; cursor: pointer; user-select: none;
}
.pill-cc-low      { background: #16a34a; }
.pill-cc-medium   { background: #d97706; }
.pill-cc-high     { background: #ea580c; }
.pill-cc-critical { background: #dc2626; }

.metric-pill {
  font-size: 11px; color: #475569; background: #f8fafc;
  border: 1px solid #e2e8f0; border-radius: 20px; padding: 2px 8px;
}
.metric-pill strong { color: #1e293b; }
.metric-pill.pill-warn { background: #fff7ed; border-color: #fed7aa; }
.metric-pill.pill-warn strong { color: #d97706; }
.expand-arrow { font-size: 10px; color: #94a3b8; }
.delete-btn {
  width: 22px; height: 22px; border: 1px solid #fecaca; border-radius: 4px;
  background: #fff; color: #f87171; font-size: 11px; cursor: pointer;
  display: flex; align-items: center; justify-content: center; transition: all .15s; padding: 0;
}
.delete-btn:hover { background: #fee2e2; border-color: #f87171; color: #dc2626; }

/* Раскрывающаяся карточка */
.history-detail { border-top: 1px solid #e2e8f0; animation: slideDown .15s ease; }
@keyframes slideDown { from { opacity: 0; transform: translateY(-4px); } to { opacity: 1; transform: translateY(0); } }
.detail-header {
  display: flex; justify-content: space-between; align-items: center;
  padding: 8px 14px; font-size: 12px; font-weight: 700; color: #fff;
}
.detail-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 1px; background: #e2e8f0; }
.detail-item { background: #fff; padding: 10px 14px; display: flex; flex-direction: column; gap: 4px; }
.detail-item-warn { background: #fff7ed; }
.detail-label { font-size: 10px; color: #94a3b8; text-transform: uppercase; letter-spacing: .04em; }
.detail-value { font-size: 18px; font-weight: 700; color: #1e293b; }
.detail-value.val-warn { color: #d97706; }
.cc-text-cc-low      { color: #16a34a; }
.cc-text-cc-medium   { color: #b45309; }
.cc-text-cc-high     { color: #c2410c; }
.cc-text-cc-critical { color: #b91c1c; }
.detail-cc-bar-wrap { height: 6px; background: #f1f5f9; }
.detail-cc-bar { height: 100%; transition: width .4s; }

/* Цвета */
.cc-low      { background: #22c55e; color: #fff; }
.cc-medium   { background: #f59e0b; color: #fff; }
.cc-high     { background: #f97316; color: #fff; }
.cc-critical { background: #ef4444; color: #fff; }
.fn-header.cc-low      { background: #22c55e; }
.fn-header.cc-medium   { background: #f59e0b; }
.fn-header.cc-high     { background: #f97316; }
.fn-header.cc-critical { background: #ef4444; }
.bar-val.cc-low    { color: #16a34a; }
.bar-val.cc-medium { color: #b45309; }
.bar-val.cc-high   { color: #c2410c; }
.bar-val.cc-critical { color: #b91c1c; }
.leg.cc-low      { background: #dcfce7; color: #16a34a; }
.leg.cc-medium   { background: #fef3c7; color: #b45309; }
.leg.cc-high     { background: #ffedd5; color: #c2410c; }
.leg.cc-critical { background: #fee2e2; color: #b91c1c; }
</style>
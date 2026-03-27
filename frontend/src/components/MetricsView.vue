<template>
  <div class="metrics-view">

    <!-- Header -->
    <div class="metrics-header">
      <h2>Метрики кода</h2>
      <button class="btn-calculate" :disabled="loading || !code" @click="calculate">
        {{ loading ? 'Считаем...' : 'Подсчитать метрики' }}
      </button>
    </div>

    <!-- Error -->
    <div v-if="error" class="metrics-error">{{ error }}</div>

    <!-- Program-level summary -->
    <div v-if="metrics" class="program-summary">
      <div class="summary-card">
        <span class="summary-label">Функций</span>
        <span class="summary-value">{{ metrics.functionCount }}</span>
      </div>
      <div class="summary-card">
        <span class="summary-label">Глобальных переменных</span>
        <span class="summary-value">{{ metrics.globalVarCount }}</span>
      </div>
    </div>

    <!-- Function cards -->
    <div v-if="metrics && metrics.functions.length" class="functions-grid">
      <div
        v-for="fn in metrics.functions"
        :key="fn.functionName"
        class="fn-card"
        :class="complexityClass(fn.cyclomaticComplexity)"
      >
        <div class="fn-name">{{ fn.functionName }}</div>

        <div class="fn-metrics">
          <div class="metric-row">
            <span class="metric-label">Цикл. сложность</span>
            <span class="metric-value" :class="complexityClass(fn.cyclomaticComplexity)">
              {{ fn.cyclomaticComplexity }}
            </span>
          </div>
          <div class="metric-row">
            <span class="metric-label">LOC</span>
            <span class="metric-value">{{ fn.loc }}</span>
          </div>
          <div class="metric-row">
            <span class="metric-label">Параметры</span>
            <span class="metric-value">{{ fn.parameterCount }}</span>
          </div>
          <div class="metric-row">
            <span class="metric-label">Макс. вложенность</span>
            <span class="metric-value" :class="nestingClass(fn.maxNestingDepth)">
              {{ fn.maxNestingDepth }}
            </span>
          </div>
          <div class="metric-row">
            <span class="metric-label">Вызовов функций</span>
            <span class="metric-value">{{ fn.callCount }}</span>
          </div>
          <div class="metric-row">
            <span class="metric-label">Return</span>
            <span class="metric-value">{{ fn.returnCount }}</span>
          </div>
          <div v-if="fn.gotoCount > 0" class="metric-row warn">
            <span class="metric-label">goto (антипаттерн)</span>
            <span class="metric-value warn">{{ fn.gotoCount }}</span>
          </div>
        </div>

        <!-- CC complexity bar -->
        <div class="cc-bar-wrap">
          <div class="cc-bar" :style="{ width: ccBarWidth(fn.cyclomaticComplexity) }"
               :class="complexityClass(fn.cyclomaticComplexity)"></div>
        </div>
        <div class="cc-legend">{{ complexityLabel(fn.cyclomaticComplexity) }}</div>
      </div>
    </div>

    <!-- Bar chart: CC per function -->
    <div v-if="metrics && metrics.functions.length > 1" class="chart-section">
      <h3>Цикломатическая сложность по функциям</h3>
      <div class="bar-chart">
        <div
          v-for="fn in metrics.functions"
          :key="fn.functionName + '_bar'"
          class="bar-item"
        >
          <div class="bar-label">{{ fn.functionName }}</div>
          <div class="bar-track">
            <div
              class="bar-fill"
              :class="complexityClass(fn.cyclomaticComplexity)"
              :style="{ width: ccBarWidth(fn.cyclomaticComplexity) }"
            ></div>
            <span class="bar-num">{{ fn.cyclomaticComplexity }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-if="!metrics && !loading" class="empty-state">
      Введите код и нажмите «Подсчитать метрики»
    </div>

  </div>
</template>

<script setup>
import { ref } from 'vue'
import { calculateMetrics } from '../api/metrics.js'

const props = defineProps({
  code: { type: String, default: '' },
})

const metrics = ref(null)
const loading = ref(false)
const error   = ref(null)

async function calculate() {
  if (!props.code) return
  loading.value = true
  error.value   = null
  metrics.value = null
  try {
    metrics.value = await calculateMetrics(props.code)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

// CC: 1-5 low, 6-10 medium, 11-20 high, 20+ critical
function complexityClass(cc) {
  if (cc <= 5)  return 'cc-low'
  if (cc <= 10) return 'cc-medium'
  if (cc <= 20) return 'cc-high'
  return 'cc-critical'
}

function complexityLabel(cc) {
  if (cc <= 5)  return 'Низкая сложность'
  if (cc <= 10) return 'Умеренная сложность'
  if (cc <= 20) return 'Высокая сложность'
  return 'Критическая сложность'
}

function nestingClass(depth) {
  if (depth <= 3) return ''
  if (depth <= 5) return 'cc-medium'
  return 'cc-high'
}

function ccBarWidth(cc) {
  // Шкала до 20, cap at 100%
  return Math.min(100, (cc / 20) * 100) + '%'
}
</script>

<style scoped>
.metrics-view {
  padding: 16px;
  font-family: inherit;
}

.metrics-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.metrics-header h2 {
  margin: 0;
  font-size: 1.2rem;
}

.btn-calculate {
  padding: 8px 18px;
  background: #2563eb;
  color: #fff;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
}
.btn-calculate:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-calculate:hover:not(:disabled) { background: #1d4ed8; }

.metrics-error {
  background: #fee2e2;
  color: #b91c1c;
  padding: 10px 14px;
  border-radius: 6px;
  margin-bottom: 12px;
  font-size: 0.9rem;
}

/* Program summary */
.program-summary {
  display: flex;
  gap: 12px;
  margin-bottom: 20px;
}
.summary-card {
  background: #f1f5f9;
  border-radius: 8px;
  padding: 12px 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
}
.summary-label { font-size: 0.75rem; color: #64748b; }
.summary-value { font-size: 1.6rem; font-weight: 700; color: #1e293b; }

/* Function cards grid */
.functions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 14px;
  margin-bottom: 24px;
}

.fn-card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 14px;
  box-shadow: 0 1px 4px rgba(0,0,0,.06);
}

.fn-name {
  font-weight: 600;
  font-size: 1rem;
  margin-bottom: 10px;
  color: #1e293b;
  border-bottom: 1px solid #f1f5f9;
  padding-bottom: 6px;
}

.fn-metrics { display: flex; flex-direction: column; gap: 5px; }

.metric-row {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
}
.metric-label { color: #64748b; }
.metric-value { font-weight: 600; color: #1e293b; }

.metric-row.warn .metric-label { color: #b45309; }
.metric-value.warn { color: #dc2626; }

/* CC bar */
.cc-bar-wrap {
  height: 6px;
  background: #f1f5f9;
  border-radius: 3px;
  margin-top: 12px;
  overflow: hidden;
}
.cc-bar { height: 100%; border-radius: 3px; transition: width .3s; }

.cc-legend {
  font-size: 0.72rem;
  margin-top: 4px;
  color: #64748b;
}

/* Complexity colors */
.cc-low      { background: #22c55e; color: #16a34a; }
.cc-medium   { background: #f59e0b; color: #b45309; }
.cc-high     { background: #f97316; color: #c2410c; }
.cc-critical { background: #ef4444; color: #b91c1c; }

/* For metric-value spans (text color only) */
.metric-value.cc-low      { color: #16a34a; }
.metric-value.cc-medium   { color: #b45309; }
.metric-value.cc-high     { color: #c2410c; }
.metric-value.cc-critical { color: #b91c1c; }

/* Bar chart */
.chart-section { margin-top: 8px; }
.chart-section h3 { font-size: 1rem; margin-bottom: 12px; color: #1e293b; }

.bar-chart { display: flex; flex-direction: column; gap: 8px; }
.bar-item  { display: flex; align-items: center; gap: 10px; }

.bar-label {
  width: 120px;
  font-size: 0.85rem;
  color: #334155;
  text-align: right;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.bar-track {
  flex: 1;
  height: 20px;
  background: #f1f5f9;
  border-radius: 4px;
  overflow: hidden;
  position: relative;
  display: flex;
  align-items: center;
}
.bar-fill {
  height: 100%;
  border-radius: 4px;
  transition: width .4s;
}
.bar-num {
  position: absolute;
  right: 6px;
  font-size: 0.75rem;
  font-weight: 600;
  color: #1e293b;
}

.empty-state {
  text-align: center;
  color: #94a3b8;
  padding: 40px 0;
  font-size: 0.9rem;
}
</style>

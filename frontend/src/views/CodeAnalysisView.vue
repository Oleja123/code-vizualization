<template>
  <div class="analysis-view">
    <section class="editor-panel">
      <div class="panel-header">
        <h2>Анализ C-кода</h2>
        <button class="analyze-btn" :disabled="loading" @click="runAnalysis">
          {{ loading ? 'Анализ...' : 'Проверить' }}
        </button>
      </div>

      <div class="examples-row">
        <label class="examples-label" for="analysis-example">Пример:</label>
        <select
          id="analysis-example"
          v-model="selectedExampleId"
          class="example-select"
          @change="applyExample"
        >
          <option v-for="example in examples" :key="example.id" :value="example.id">
            {{ example.name }}
          </option>
        </select>
      </div>

      <p v-if="error" class="error-message">{{ error }}</p>

      <div class="code-container">
        <div class="line-numbers" ref="lineNumbersRef">
          <div
            v-for="lineNumber in lineCount"
            :key="lineNumber"
            class="line-number"
            :class="{
              'issue-line': issueLines.has(lineNumber),
              'active-issue-line': selectedIssueLine === lineNumber
            }"
          >
            {{ lineNumber }}
          </div>
        </div>
        <textarea
          ref="textareaRef"
          v-model="code"
          class="code-input"
          spellcheck="false"
          @keydown="handleKeydown"
          @scroll="syncScroll"
        ></textarea>
      </div>
    </section>

    <section class="issues-panel">
      <div class="panel-header">
        <h2>Результат</h2>
        <span class="issues-count">{{ issues.length }}</span>
      </div>

      <div v-if="loading" class="placeholder">Выполняем анализ...</div>
      <div v-else-if="!analyzed" class="placeholder">Нажмите «Проверить», чтобы получить замечания.</div>
      <div v-else-if="issues.length === 0" class="placeholder success">Проблем в коде не найдено.</div>

      <div v-else class="issues-list">
        <button
          v-for="issue in issues"
          :key="issue.id + '-' + issue.line + '-' + issue.message"
          class="issue-item"
          :class="{ active: selectedIssueLine === issue.line }"
          @click="focusIssue(issue)"
        >
          <div class="issue-top-row">
            <span class="issue-severity">{{ issue.severity }}</span>
            <span class="issue-line-label">Строка {{ issue.line }}</span>
          </div>
          <div class="issue-id">{{ issue.id }}</div>
          <div class="issue-message">{{ issue.message }}</div>
        </button>
      </div>
    </section>
  </div>
</template>

<script>
import { computed, ref } from 'vue'
import { analyzeCode } from '../api/cppcheck.js'

export default {
  name: 'CodeAnalysisView',
  setup() {
    const examples = [
      {
        id: 'uninitialized-variable',
        name: 'Неинициализированная переменная',
        code: 'int main() {\n  int x;\n  int y = x + 1;\n  return y;\n}'
      },
      {
        id: 'possible-null-pointer',
        name: 'Потенциальный null-указатель',
        code: '#include <stdlib.h>\n\nint main() {\n  int *arr = NULL;\n  if (rand() % 2) {\n    arr = (int*)malloc(sizeof(int) * 4);\n  }\n  arr[0] = 10;\n  free(arr);\n  return 0;\n}'
      },
      {
        id: 'out-of-bounds',
        name: 'Выход за границы массива',
        code: 'int main() {\n  int arr[3] = {1, 2, 3};\n  int i = 3;\n  return arr[i];\n}'
      },
      {
        id: 'many-errors',
        name: 'Большой пример с множеством ошибок',
        code: '#include <stdlib.h>\n\nint main() {\n  int uninit;\n  int arr[2] = {1, 2};\n  int *p = NULL;\n  int *buf = (int*)malloc(sizeof(int) * 2);\n\n  arr[3] = 10;\n  p[0] = 1;\n\n  int z = 10 / (arr[0] - 1);\n  int use = uninit + z;\n\n  if (use > 0) {\n    return buf[5];\n  }\n\n  return arr[5];\n}'
      },
      {
        id: 'clean-code',
        name: 'Корректный код без замечаний',
        code: 'int main() {\n  int sum = 0;\n  for (int i = 1; i <= 5; i++) {\n    sum += i;\n  }\n  return sum;\n}'
      }
    ]

    const selectedExampleId = ref(examples[0].id)
    const code = ref(examples[0].code)
    const issues = ref([])
    const error = ref('')
    const loading = ref(false)
    const analyzed = ref(false)
    const selectedIssueLine = ref(null)
    const textareaRef = ref(null)
    const lineNumbersRef = ref(null)

    const lineCount = computed(() => code.value.split('\n').length)

    const issueLines = computed(() => {
      const lines = new Set()
      for (const issue of issues.value) {
        if (issue.line > 0) {
          lines.add(issue.line)
        }
      }
      return lines
    })

    const syncScroll = (event) => {
      if (lineNumbersRef.value) {
        lineNumbersRef.value.scrollTop = event.target.scrollTop
      }
    }

    const handleKeydown = (event) => {
      if (event.key === 'Tab') {
        event.preventDefault()
        const textarea = event.target
        const start = textarea.selectionStart
        const end = textarea.selectionEnd
        const tab = '  '

        code.value = code.value.substring(0, start) + tab + code.value.substring(end)

        setTimeout(() => {
          textarea.selectionStart = textarea.selectionEnd = start + tab.length
        }, 0)
      }
    }

    const applyExample = () => {
      const example = examples.find((item) => item.id === selectedExampleId.value)
      if (!example) {
        return
      }

      code.value = example.code
      issues.value = []
      error.value = ''
      analyzed.value = false
      selectedIssueLine.value = null
      if (textareaRef.value) {
        textareaRef.value.scrollTop = 0
      }
      if (lineNumbersRef.value) {
        lineNumbersRef.value.scrollTop = 0
      }
    }

    const scrollToLine = (line) => {
      if (!textareaRef.value || !line || line <= 0) {
        return
      }

      const lineHeight = 21
      const targetScroll = Math.max(0, (line - 1) * lineHeight - textareaRef.value.clientHeight / 2)
      textareaRef.value.scrollTop = targetScroll
      if (lineNumbersRef.value) {
        lineNumbersRef.value.scrollTop = targetScroll
      }
    }

    const focusIssue = (issue) => {
      selectedIssueLine.value = issue.line
      scrollToLine(issue.line)
    }

    const runAnalysis = async () => {
      if (!code.value.trim()) {
        error.value = 'Введите C-код для анализа'
        return
      }

      loading.value = true
      error.value = ''
      issues.value = []
      selectedIssueLine.value = null

      try {
        const result = await analyzeCode(code.value)
        issues.value = Array.isArray(result.issues) ? result.issues : []
        analyzed.value = true

        if (issues.value.length > 0) {
          focusIssue(issues.value[0])
        }
      } catch (err) {
        analyzed.value = false
        error.value = err.message || 'Не удалось выполнить анализ'
      } finally {
        loading.value = false
      }
    }

    return {
      examples,
      selectedExampleId,
      code,
      issues,
      error,
      loading,
      analyzed,
      selectedIssueLine,
      textareaRef,
      lineNumbersRef,
      lineCount,
      issueLines,
      handleKeydown,
      applyExample,
      syncScroll,
      focusIssue,
      runAnalysis,
    }
  },
}
</script>

<style scoped>
.analysis-view {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 1rem;
  height: 100%;
  padding: 1rem;
}

.editor-panel,
.issues-panel {
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.panel-header {
  padding: 0.9rem 1rem;
  border-bottom: 1px solid #e9ecef;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.panel-header h2 {
  font-size: 1rem;
  font-weight: 600;
}

.examples-row {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.65rem 1rem;
  border-bottom: 1px solid #e9ecef;
}

.examples-label {
  font-size: 0.86rem;
  color: #475467;
  font-weight: 600;
}

.example-select {
  flex: 1;
  min-width: 0;
  border: 1px solid #d0d5dd;
  border-radius: 6px;
  padding: 0.38rem 0.5rem;
  background: #fff;
  color: #101828;
  font-size: 0.88rem;
}

.issues-count {
  background: #f0f4ff;
  color: #284b9b;
  border-radius: 999px;
  font-size: 0.8rem;
  padding: 0.2rem 0.6rem;
}

.analyze-btn {
  border: none;
  background: #1f6feb;
  color: white;
  border-radius: 6px;
  padding: 0.45rem 0.9rem;
  cursor: pointer;
  font-weight: 600;
}

.analyze-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-message {
  margin: 0.75rem 1rem 0;
  color: #b42318;
  background: #fef3f2;
  border: 1px solid #fecdca;
  border-radius: 6px;
  padding: 0.6rem 0.75rem;
  font-size: 0.9rem;
}

.code-container {
  display: flex;
  overflow: hidden;
  flex: 1;
  min-height: 0;
}

.line-numbers {
  width: 52px;
  background: #f8fafc;
  border-right: 1px solid #e2e8f0;
  color: #64748b;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 21px;
  overflow: hidden;
  padding: 0.8rem 0;
}

.line-number {
  text-align: right;
  padding-right: 0.6rem;
}

.line-number.issue-line {
  background: #fff2cc;
  color: #7a4f00;
  font-weight: 600;
}

.line-number.active-issue-line {
  background: #ffd99a;
  color: #4d2d00;
}

.code-input {
  flex: 1;
  min-height: 0;
  border: none;
  resize: none;
  outline: none;
  padding: 0.8rem;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 21px;
  background: #fff;
}

.placeholder {
  margin: auto;
  color: #64748b;
  text-align: center;
  padding: 1rem;
}

.placeholder.success {
  color: #067647;
}

.issues-list {
  overflow: auto;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.55rem;
}

.issue-item {
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 0.65rem;
  background: #fff;
  text-align: left;
  cursor: pointer;
}

.issue-item.active {
  border-color: #d97706;
  box-shadow: 0 0 0 2px rgba(217, 119, 6, 0.15);
}

.issue-top-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.35rem;
}

.issue-severity {
  text-transform: uppercase;
  font-size: 0.72rem;
  color: #7a4f00;
  background: #ffedd5;
  border-radius: 999px;
  padding: 0.15rem 0.5rem;
  font-weight: 700;
}

.issue-line-label {
  font-size: 0.8rem;
  color: #475467;
}

.issue-id {
  font-size: 0.82rem;
  color: #1f2937;
  font-weight: 600;
  margin-bottom: 0.2rem;
}

.issue-message {
  font-size: 0.86rem;
  color: #374151;
  line-height: 1.35;
}

@media (max-width: 1100px) {
  .analysis-view {
    grid-template-columns: 1fr;
  }
}
</style>

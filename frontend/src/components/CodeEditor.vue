<template>
  <div class="code-editor">
    <div class="editor-header">
      <div class="editor-title-row">
        <h2>Код программы</h2>
        <label v-if="!isExecuted && examples.length" class="examples-select-label">
          Пример:
          <select
            class="examples-select"
            :value="selectedExample"
            @change="handleExampleChange"
          >
            <option value="">Выберите пример</option>
            <option
              v-for="example in examples"
              :key="example.id"
              :value="example.id"
            >
              {{ example.name }}
            </option>
          </select>
        </label>
      </div>
      <div v-if="error" class="error-message">
        {{ error }}
      </div>
    </div>
    <div class="code-container" :class="{ 'container-readonly': isExecuted }">
      <div class="line-numbers" ref="lineNumbers">
        <div 
          v-for="n in lineCount" 
          :key="n"
          class="line-number"
          :class="{ 'current-line': n === currentLine }"
        >
          {{ n }}
        </div>
      </div>
      <textarea
        class="code-input"
        :class="{ 'code-readonly': isExecuted }"
        :value="code"
        @input="handleInput"
        @scroll="handleScroll"
        @keydown="handleKeydown"
        placeholder="Введите C код..."
        spellcheck="false"
        :readonly="isExecuted"
        ref="textarea"
      ></textarea>
    </div>
    <div class="controls">
      <button
        v-if="!isExecuted"
        class="control-button execute-button"
        @click="handleExecute"
        :disabled="loading || !code"
      >
        ▶ Выполнить
      </button>
      <template v-else>
        <button
          class="control-button"
          @click="handleStepBackward"
          :disabled="loading || currentStep === 0"
        >
          ← Шаг назад
        </button>
        <div class="step-info">
          Шаг {{ displayCurrentStep }} из {{ displayStepsCount }}
        </div>
        <button
          class="control-button"
          @click="handleStepForward"
          :disabled="loading || currentStep >= stepsCount - 1"
        >
          Шаг вперед →
        </button>
        <button
          class="control-button reset-button"
          @click="handleExecute"
          :disabled="loading"
        >
          🔄 Заново
        </button>
        <button
          class="control-button edit-button"
          @click="handleEdit"
          :disabled="loading"
        >
          ✏️ Редактировать код
        </button>
      </template>
    </div>
  </div>
</template>

<script>
import { ref, computed, watch, nextTick } from 'vue'

export default {
  name: 'CodeEditor',
  props: {
    code: {
      type: String,
      required: true
    },
    currentStep: {
      type: Number,
      default: 0
    },
    stepsCount: {
      type: Number,
      default: 0
    },
    loading: {
      type: Boolean,
      default: false
    },
    error: {
      type: String,
      default: null
    },
    isExecuted: {
      type: Boolean,
      default: false
    },
    currentLine: {
      type: Number,
      default: null
    },
    examples: {
      type: Array,
      default: () => []
    },
    selectedExample: {
      type: String,
      default: ''
    }
  },
  emits: ['update:code', 'update:selectedExample', 'execute', 'edit', 'step-forward', 'step-backward'],
  setup(props, { emit }) {
    const textarea = ref(null)
    const lineNumbers = ref(null)

    const lineCount = computed(() => {
      return props.code.split('\n').length
    })

    const displayCurrentStep = computed(() => {
      if (props.stepsCount <= 0) {
        return 0
      }

      return Math.min(props.currentStep + 1, props.stepsCount)
    })

    const displayStepsCount = computed(() => {
      return Math.max(props.stepsCount, 0)
    })

    const handleInput = (event) => {
      emit('update:code', event.target.value)
    }

    const handleScroll = (event) => {
      if (lineNumbers.value) {
        lineNumbers.value.scrollTop = event.target.scrollTop
      }
    }

    const handleKeydown = (event) => {
      if (event.key === 'Tab') {
        event.preventDefault()
        const textarea = event.target
        const start = textarea.selectionStart
        const end = textarea.selectionEnd
        const tab = '  ' // 2 пробела для отступа
        
        const newCode = props.code.substring(0, start) + tab + props.code.substring(end)
        emit('update:code', newCode)
        
        // Перемещаем курсор после вставленного отступа
        setTimeout(() => {
          textarea.selectionStart = textarea.selectionEnd = start + tab.length
        }, 0)
      }
    }

    const handleExecute = () => {
      console.log('handleExecute called')
      emit('execute')
    }

    const handleStepForward = () => {
      console.log('handleStepForward called')
      emit('step-forward')
    }

    const handleStepBackward = () => {
      console.log('handleStepBackward called')
      emit('step-backward')
    }

    const handleEdit = () => {
      console.log('handleEdit called')
      emit('edit')
    }

    const handleExampleChange = (event) => {
      emit('update:selectedExample', event.target.value)
    }

    // Автоматическая прокрутка к текущей строке
    watch(() => props.currentLine, async (newLine) => {
      if (newLine && textarea.value) {
        await nextTick()
        const lineHeight = 22 // приблизительная высота строки
        const scrollTop = (newLine - 1) * lineHeight - textarea.value.clientHeight / 2
        textarea.value.scrollTop = Math.max(0, scrollTop)
      }
    })

    return {
      textarea,
      lineNumbers,
      lineCount,
      displayCurrentStep,
      displayStepsCount,
      handleInput,
      handleScroll,
      handleKeydown,
      handleExecute,
      handleStepForward,
      handleStepBackward,
      handleEdit,
      handleExampleChange
    }
  }
}
</script>

<style scoped>
.code-editor {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.editor-header {
  padding: 1rem;
  border-bottom: 1px solid #e0e0e0;
}

.editor-header h2 {
  font-size: 1.2rem;
  font-weight: 600;
}

.editor-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.examples-select-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.9rem;
  color: #555;
}

.examples-select {
  padding: 0.35rem 0.5rem;
  border: 1px solid #d0d7de;
  border-radius: 4px;
  background-color: #fff;
  font-size: 0.9rem;
}

.error-message {
  color: #e74c3c;
  font-size: 0.9rem;
  padding: 0.5rem;
  background-color: #fadbd8;
  border-radius: 4px;
  margin-top: 0.5rem;
}

.code-container {
  flex: 1;
  display: flex;
  overflow: hidden;
  background-color: #f8f9fa;
  transition: background-color 0.3s;
}

.code-container.container-readonly {
  background-color: #e9ecef;
}

.line-numbers {
  padding: 1rem 0.5rem;
  background-color: #e9ecef;
  border-right: 1px solid #dee2e6;
  text-align: right;
  user-select: none;
  overflow: hidden;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  color: #6c757d;
  min-width: 40px;
  transition: background-color 0.3s;
}

.container-readonly .line-numbers {
  background-color: #d6d8db;
}

.line-number {
  height: 21px;
  padding-right: 0.5rem;
  transition: background-color 0.2s, color 0.2s;
}

.line-number.current-line {
  background-color: #fff3cd;
  color: #856404;
  font-weight: bold;
}

.code-input {
  flex: 1;
  padding: 1rem;
  font-family: 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  border: none;
  resize: none;
  outline: none;
  background-color: #f8f9fa;
  transition: background-color 0.3s;
}

.code-input.code-readonly {
  background-color: #e9ecef;
  cursor: not-allowed;
  color: #495057;
}

.controls {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  border-top: 1px solid #e0e0e0;
  background-color: white;
}

.control-button {
  padding: 0.6rem 1.2rem;
  background-color: #3498db;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
  transition: background-color 0.2s;
}

.control-button:hover:not(:disabled) {
  background-color: #2980b9;
}

.control-button:disabled {
  background-color: #bdc3c7;
  cursor: not-allowed;
}

.execute-button {
  background-color: #27ae60;
  font-weight: 600;
}

.execute-button:hover:not(:disabled) {
  background-color: #229954;
}

.reset-button {
  background-color: #95a5a6;
  padding: 0.6rem 1rem;
}

.reset-button:hover:not(:disabled) {
  background-color: #7f8c8d;
}

.edit-button {
  background-color: #e67e22;
  padding: 0.6rem 1rem;
}

.edit-button:hover:not(:disabled) {
  background-color: #d35400;
}

.step-info {
  flex: 1;
  text-align: center;
  font-weight: 500;
  color: #555;
}
</style>

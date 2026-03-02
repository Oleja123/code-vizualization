<template>
  <div class="runtime-visualization">
    <div class="visualization-header">
      <h2>Состояние выполнения</h2>
      <div v-if="snapshot && snapshot.line" class="current-line">
        Строка: {{ snapshot.line }}
      </div>
    </div>
    <div v-if="hasFunctionReturn" class="function-return">
      возврат из функции {{ snapshot.function_name }}: {{ snapshot.return_value }}
    </div>
    <div class="call-stack" v-if="snapshot && snapshot.call_stack">
      <StackFrame
        v-for="(frame, index) in snapshot.call_stack.frames"
        :key="index"
        :frame="frame"
        :current-step="currentStep"
        :is-global="frame.func_name === 'global'"
      />
    </div>
    <div v-else class="empty-state">
      Нет данных для визуализации
    </div>

    <!-- Модальное окно ошибки -->
    <div v-if="showErrorModal" class="error-modal-overlay" @click="closeErrorModal">
      <div class="error-modal" @click.stop>
        <div class="error-modal-header">
          <h3>Ошибка выполнения</h3>
          <button class="close-button" @click="closeErrorModal">×</button>
        </div>
        <div class="error-modal-body">
          <p>{{ currentError }}</p>
        </div>
        <div class="error-modal-footer">
          <button class="error-button" @click="closeErrorModal">Закрыть</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, watch } from 'vue'
import StackFrame from './StackFrame.vue'

export default {
  name: 'RuntimeVisualization',
  components: {
    StackFrame
  },
  props: {
    snapshot: {
      type: Object,
      default: null
    },
    currentStep: {
      type: Number,
      default: 0
    }
  },
  computed: {
    hasFunctionReturn() {
      return this.snapshot && this.snapshot.return_value !== null && this.snapshot.return_value !== undefined
    }
  },
  setup(props) {
    const showErrorModal = ref(false)
    const currentError = ref('')

    // Отслеживание изменений в snapshot.error
    watch(() => props.snapshot?.error, (newError) => {
      if (newError && newError.trim() !== '') {
        currentError.value = newError
        showErrorModal.value = true
      }
    })

    const closeErrorModal = () => {
      showErrorModal.value = false
    }

    return {
      showErrorModal,
      currentError,
      closeErrorModal
    }
  }
}
</script>

<style scoped>
.runtime-visualization {
  height: 100%;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
}

.visualization-header {
  padding: 1rem;
  border-bottom: 1px solid #e0e0e0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.visualization-header h2 {
  font-size: 1.2rem;
  font-weight: 600;
}

.current-line {
  color: #3498db;
  font-weight: 500;
  padding: 0.25rem 0.75rem;
  background-color: #e3f2fd;
  border-radius: 4px;
}

.call-stack {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.function-return {
  margin: 1rem 1rem 0;
  padding: 0.5rem 0.75rem;
  border-radius: 4px;
  background-color: #e8f5e9;
  color: #2e7d32;
  font-weight: 500;
}

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 1.1rem;
}

/* Модальное окно ошибки */
.error-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.error-modal {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
  max-width: 500px;
  width: 90%;
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    transform: translateY(-50px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.error-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e0e0e0;
  background-color: #e74c3c;
  color: white;
  border-radius: 8px 8px 0 0;
}

.error-modal-header h3 {
  margin: 0;
  font-size: 1.3rem;
  font-weight: 600;
}

.close-button {
  background: none;
  border: none;
  color: white;
  font-size: 2rem;
  cursor: pointer;
  padding: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.close-button:hover {
  background-color: rgba(255, 255, 255, 0.2);
}

.error-modal-body {
  padding: 1.5rem;
  font-size: 1rem;
  line-height: 1.6;
  color: #333;
}

.error-modal-body p {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
}

.error-modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid #e0e0e0;
  display: flex;
  justify-content: flex-end;
}

.error-button {
  background-color: #e74c3c;
  color: white;
  border: none;
  padding: 0.6rem 1.5rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
  font-weight: 500;
  transition: background-color 0.2s;
}

.error-button:hover {
  background-color: #c0392b;
}
</style>

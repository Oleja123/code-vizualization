<template>
  <div class="runtime-visualization">
    <div class="visualization-header">
      <h2>Состояние выполнения</h2>
      <div v-if="snapshot && snapshot.line" class="current-line">
        Строка: {{ snapshot.line }}
      </div>
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
  </div>
</template>

<script>
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

.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #999;
  font-size: 1.1rem;
}
</style>

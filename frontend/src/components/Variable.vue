<template>
  <div class="variable" :class="{ highlighted: isHighlighted }">
    <span class="variable-name">{{ variable.name }}:</span>
    <span class="variable-value">{{ displayValue }}</span>
  </div>
</template>

<script>
export default {
  name: 'Variable',
  props: {
    variable: {
      type: Object,
      required: true
    },
    currentStep: {
      type: Number,
      required: true
    }
  },
  computed: {
    displayValue() {
      return this.variable.value !== null && this.variable.value !== undefined
        ? this.variable.value
        : '?'
    },
    isHighlighted() {
      return this.variable.step_changed === this.currentStep
    }
  }
}
</script>

<style scoped>
.variable {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  background-color: #ecf0f1;
  border-radius: 4px;
  font-family: 'Courier New', monospace;
  transition: all 0.3s ease;
}

.variable.highlighted {
  background-color: #fff9c4;
  box-shadow: 0 0 8px rgba(255, 235, 59, 0.6);
  animation: pulse 0.5s ease-in-out;
}

@keyframes pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.05);
  }
}

.variable-name {
  font-weight: 600;
  color: #2c3e50;
}

.variable-value {
  color: #27ae60;
  font-weight: 500;
}
</style>

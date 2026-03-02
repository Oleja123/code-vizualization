<template>
  <div class="array">
    <div class="array-name">{{ array.name }}[{{ array.size }}]:</div>
    <div class="array-elements">
      <div
        v-for="(element, index) in array.values"
        :key="index"
        class="array-element"
        :class="{ highlighted: isElementHighlighted(element) }"
      >
        <div class="element-index">[{{ index }}]</div>
        <div class="element-value">{{ getElementValue(element) }}</div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Array',
  props: {
    array: {
      type: Object,
      required: true
    },
    currentStep: {
      type: Number,
      required: true
    }
  },
  methods: {
    getElementValue(element) {
      return element.value !== null && element.value !== undefined
        ? element.value
        : '?'
    },
    isElementHighlighted(element) {
      return element.step_changed === this.currentStep
    }
  }
}
</script>

<style scoped>
.array {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.array-name {
  font-weight: 600;
  color: #2c3e50;
  font-family: 'Courier New', monospace;
}

.array-elements {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.array-element {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.5rem;
  background-color: #ecf0f1;
  border-radius: 4px;
  min-width: 60px;
  transition: all 0.3s ease;
}

.array-element.highlighted {
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

.element-index {
  font-size: 0.8rem;
  color: #7f8c8d;
  font-family: 'Courier New', monospace;
}

.element-value {
  font-weight: 600;
  color: #27ae60;
  font-family: 'Courier New', monospace;
}
</style>

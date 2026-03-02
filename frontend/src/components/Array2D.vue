<template>
  <div class="array2d">
    <div class="array2d-name">{{ array2d.name }}[{{ array2d.size1 }}][{{ array2d.size2 }}]:</div>
    <div class="array2d-matrix">
      <div
        v-for="(row, rowIndex) in array2d.values"
        :key="rowIndex"
        class="array2d-row"
      >
        <div
          v-for="(element, colIndex) in row.values"
          :key="colIndex"
          class="array2d-element"
          :class="{ highlighted: isElementHighlighted(element) }"
        >
          <div class="element-index">[{{ rowIndex }}][{{ colIndex }}]</div>
          <div class="element-value">{{ getElementValue(element) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Array2D',
  props: {
    array2d: {
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
.array2d {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.array2d-name {
  font-weight: 600;
  color: #2c3e50;
  font-family: 'Courier New', monospace;
}

.array2d-matrix {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.array2d-row {
  display: flex;
  gap: 0.5rem;
}

.array2d-element {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.5rem;
  background-color: #ecf0f1;
  border-radius: 4px;
  min-width: 70px;
  transition: all 0.3s ease;
}

.array2d-element.highlighted {
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
  font-size: 0.75rem;
  color: #7f8c8d;
  font-family: 'Courier New', monospace;
}

.element-value {
  font-weight: 600;
  color: #27ae60;
  font-family: 'Courier New', monospace;
  margin-top: 0.25rem;
}
</style>

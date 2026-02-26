<template>
  <div class="visualization-view">
    <div class="left-panel">
      <CodeEditor
        :code="code"
        :examples="examples"
        :selected-example="selectedExample"
        :current-step="currentStep"
        :steps-count="stepsCount"
        :loading="loading"
        :error="error"
        :is-executed="isExecuted"
        :current-line="snapshot?.line"
        @update:code="code = $event"
        @update:selected-example="selectedExample = $event"
        @execute="executeCode"
        @edit="editCode"
        @step-forward="stepForward"
        @step-backward="stepBackward"
      />
    </div>
    <div class="right-panel">
      <RuntimeVisualization
        :snapshot="snapshot"
        :current-step="currentStep"
      />
    </div>
  </div>
</template>

<script>
import { ref } from 'vue'
import CodeEditor from '../components/CodeEditor.vue'
import RuntimeVisualization from '../components/RuntimeVisualization.vue'
import { getSnapshot } from '../api/interpreter.js'

export default {
  name: 'VisualizationView',
  components: {
    CodeEditor,
    RuntimeVisualization
  },
  setup() {
    const examples = [
      {
        id: 'sum',
        name: 'Сложение двух чисел',
        code: 'int main() {\n  int x = 5;\n  int y = 10;\n  int sum = x + y;\n  return sum;\n}'
      },
      {
        id: 'factorial',
        name: 'Рекурсивный факториал',
        code: 'int factorial(int n) {\n  if(n <= 1) {\n    return 1;\n  }\n  return factorial(n - 1) * n;\n}\n\nint main() {\n  int res = factorial(4);\n  return 0;\n}'
      },
      {
        id: 'while-loop',
        name: 'Сумма в цикле while',
        code: 'int main() {\n  int i = 1;\n  int sum = 0;\n  while (i <= 5) {\n    sum += i;\n    i++;\n  }\n  return sum;\n}'
      },
      {
        id: 'array-max',
        name: 'Максимум в массиве',
        code: 'int main() {\n  int arr[5] = {3, 7, 2, 9, 5};\n  int max = arr[0];\n  int i = 1;\n  while (i < 5) {\n    if (arr[i] > max) {\n      max = arr[i];\n    }\n    i++;\n  }\n  return max;\n}'
      },
      {
        id: 'global-for',
        name: 'for + глобальные переменные и массив',
        code: 'int gBase = 2;\nint gSum = 0;\nint gArr[5] = {1, 3, 5, 7, 9};\n\nint main() {\n  for (int i = 0; i < 5; i++) {\n    gSum += gArr[i] * gBase;\n  }\n  return gSum;\n}'
      },
      {
        id: 'nested-loops',
        name: 'Вложенные циклы',
        code: 'int main() {\n  int i = 1;\n  int total = 0;\n  while (i <= 3) {\n    int j = 1;\n    while (j <= 2) {\n      total += i * j;\n      j++;\n    }\n    i++;\n  }\n  return total;\n}'
      }
    ]

    const selectedExample = ref('sum')
    const code = ref(examples[0].code)
    const currentStep = ref(0)
    const stepsCount = ref(0)
    const snapshot = ref(null)
    const loading = ref(false)
    const error = ref(null)
    const isExecuted = ref(false)

    const loadSnapshot = async (step) => {
      console.log('loadSnapshot called with step:', step)
      loading.value = true
      error.value = null
      
      try {
        const data = await getSnapshot(code.value, step)
        console.log('Received snapshot data:', data)
        snapshot.value = data.snapshot
        currentStep.value = data.current_step ?? step
        stepsCount.value = data.steps_count ?? 0
        console.log('Updated state:', { currentStep: currentStep.value, stepsCount: stepsCount.value })
      } catch (err) {
        console.error('Error loading snapshot:', err)
        error.value = err.message
        snapshot.value = null
        isExecuted.value = false
      } finally {
        loading.value = false
      }
    }

    const executeCode = async () => {
      console.log('executeCode called')
      isExecuted.value = false
      currentStep.value = 0
      await loadSnapshot(0)
      if (!error.value) {
        isExecuted.value = true
        console.log('Code executed successfully, isExecuted:', isExecuted.value)
      } else {
        console.log('Code execution failed with error:', error.value)
      }
    }

    const editCode = () => {
      console.log('editCode called')
      isExecuted.value = false
      currentStep.value = 0
      stepsCount.value = 0
      snapshot.value = null
      error.value = null
    }

    const stepForward = async () => {
      console.log('stepForward called', { currentStep: currentStep.value, stepsCount: stepsCount.value })
      if (currentStep.value < stepsCount.value - 1) {
        await loadSnapshot(currentStep.value + 1)
      } else {
        console.log('Cannot step forward: already at last step')
      }
    }

    const applySelectedExample = (exampleId) => {
      const selected = examples.find((example) => example.id === exampleId)
      if (!selected) {
        return
      }

      code.value = selected.code
      editCode()
    }

    const setSelectedExample = (exampleId) => {
      selectedExample.value = exampleId
      applySelectedExample(exampleId)
    }

    const stepBackward = async () => {
      console.log('stepBackward called', { currentStep: currentStep.value, stepsCount: stepsCount.value })
      if (currentStep.value > 0) {
        await loadSnapshot(currentStep.value - 1)
      } else {
        console.log('Cannot step backward: already at first step')
      }
    }

    return {
      code,
      examples,
      selectedExample,
      currentStep,
      stepsCount,
      snapshot,
      loading,
      error,
      isExecuted,
      executeCode,
      editCode,
      setSelectedExample,
      stepForward,
      stepBackward
    }
  },
  watch: {
    selectedExample(newValue) {
      this.setSelectedExample(newValue)
    }
  }
}
</script>

<style scoped>
.visualization-view {
  display: flex;
  height: 100%;
  gap: 1rem;
  padding: 1rem;
}

.left-panel {
  flex: 1;
  min-width: 0;
}

.right-panel {
  flex: 1;
  min-width: 0;
  overflow: auto;
}
</style>

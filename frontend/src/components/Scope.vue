<template>
  <div class="scope">
    <div class="scope-label">{{ scopeLabel }}</div>
    <div class="scope-content">
      <template v-for="(declaration, index) in declarations" :key="index">
        <Variable
          v-if="isVariable(declaration)"
          :variable="declaration"
          :current-step="currentStep"
        />
        <Array
          v-else-if="isArray(declaration)"
          :array="declaration"
          :current-step="currentStep"
        />
        <Array2D
          v-else-if="isArray2D(declaration)"
          :array2d="declaration"
          :current-step="currentStep"
        />
      </template>
      <div v-if="declarations.length === 0" class="empty-scope">
        Нет переменных
      </div>
    </div>
  </div>
</template>

<script>
import Variable from './Variable.vue'
import Array from './Array.vue'
import Array2D from './Array2D.vue'

export default {
  name: 'Scope',
  components: {
    Variable,
    Array,
    Array2D
  },
  props: {
    scope: {
      type: Object,
      required: true
    },
    currentStep: {
      type: Number,
      required: true
    },
    scopeLabel: {
      type: String,
      default: 'Область'
    }
  },
  computed: {
    declarations() {
      return this.scope?.declarations?.declarations || []
    }
  },
  methods: {
    isVariable(declaration) {
      // Это переменная, если у неё есть имя и это не массив
      return declaration.name && !declaration.size && !declaration.size1
    },
    isArray(declaration) {
      return declaration.size !== undefined && declaration.values !== undefined
    },
    isArray2D(declaration) {
      return declaration.size1 !== undefined && declaration.size2 !== undefined
    }
  }
}
</script>

<style scoped>
.scope {
  background-color: white;
  border: 1px solid #ddd;
  border-radius: 6px;
  padding: 0.75rem;
}

.scope-label {
  font-size: 0.9rem;
  font-weight: 600;
  color: #555;
  margin-bottom: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.scope-content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.empty-scope {
  color: #999;
  font-style: italic;
  font-size: 0.9rem;
}
</style>

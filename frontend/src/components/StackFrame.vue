<template>
  <div class="stack-frame">
    <div class="frame-header">
      <span class="frame-name">{{ frame.func_name }}</span>
    </div>
    <div class="frame-scopes">
      <Scope
        v-for="(scope, index) in visibleScopes"
        :key="index"
        :scope="scope"
        :current-step="currentStep"
        :scope-label="getScopeLabel(index)"
      />
    </div>
  </div>
</template>

<script>
import Scope from './Scope.vue'

export default {
  name: 'StackFrame',
  components: {
    Scope
  },
  props: {
    frame: {
      type: Object,
      required: true
    },
    currentStep: {
      type: Number,
      required: true
    },
    isGlobal: {
      type: Boolean,
      default: false
    }
  },
  computed: {
    visibleScopes() {
      // Для global фрейма показываем все scope'ы
      // Для остальных пропускаем первый scope
      if (this.isGlobal) {
        return this.frame.scopes || []
      }
      return (this.frame.scopes || []).slice(1)
    }
  },
  methods: {
    getScopeLabel(index) {
      if (this.isGlobal) {
        return index === 0 ? 'Глобальная область' : `Область ${index}`
      }
      // Для не-global фреймов первый видимый scope - это параметры (индекс 1 в исходном массиве)
      return index === 0 ? 'Параметры' : `Область ${index}`
    }
  }
}
</script>

<style scoped>
.stack-frame {
  border: 2px solid #3498db;
  border-radius: 8px;
  padding: 1rem;
  background-color: #f8f9fa;
}

.frame-header {
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid #3498db;
}

.frame-name {
  font-size: 1.1rem;
  font-weight: 600;
  color: #2c3e50;
}

.frame-scopes {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}
</style>

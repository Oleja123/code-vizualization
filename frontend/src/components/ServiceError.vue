<template>
  <div class="svc-error-wrap">
    <div class="svc-error-box" :class="type">
      <div class="svc-error-icon">{{ icon }}</div>
      <div class="svc-error-title">{{ title }}</div>
      <div class="svc-error-desc">{{ desc }}</div>
      <div v-if="code" class="svc-error-code">{{ code }}</div>
      <button v-if="retryable" class="svc-retry-btn" @click="$emit('retry')">
        ↻ Попробовать снова
      </button>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  message: { type: String, default: '' },
})
defineEmits(['retry'])

const type = computed(() => {
  if (props.message === 'AUTH_REQUIRED') return 'auth'
  if (props.message === 'SERVICE_UNAVAILABLE') return 'unavailable'
  return 'generic'
})

const icon  = computed(() => ({ auth: '🔒', unavailable: '🔌', generic: '⚠️' }[type.value]))
const title = computed(() => ({
  auth:        'Требуется авторизация',
  unavailable: 'Сервис недоступен',
  generic:     'Ошибка выполнения',
}[type.value]))

const desc = computed(() => ({
  auth:        'Ваша сессия истекла или сервис авторизации не отвечает. Попробуйте выйти и войти снова.',
  unavailable: 'Сервис временно не отвечает. Убедитесь что все контейнеры запущены.',
  generic:     props.message,
}[type.value]))

const code = computed(() => {
  if (type.value === 'auth') return '401 Unauthorized'
  if (type.value === 'unavailable') return '502 / 503 Service Unavailable'
  return ''
})

const retryable = computed(() => type.value !== 'auth')
</script>

<style scoped>
.svc-error-wrap {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  padding: 24px;
}

.svc-error-box {
  width: 100%;
  max-width: 420px;
  border-radius: 14px;
  padding: 28px 28px 24px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  text-align: center;
  box-shadow: 0 4px 24px rgba(0,0,0,.08);
}

.svc-error-box.auth        { background: #fafafa; border: 1.5px solid #c7d2fe; }
.svc-error-box.unavailable { background: #fafafa; border: 1.5px solid #e2e8f0; }
.svc-error-box.generic     { background: #fff9f9; border: 1.5px solid #fecaca; }

.svc-error-icon  { font-size: 38px; line-height: 1; }

.svc-error-title {
  font-size: 16px;
  font-weight: 700;
  color: #1e293b;
}

.svc-error-box.auth        .svc-error-title { color: #3730a3; }
.svc-error-box.unavailable .svc-error-title { color: #334155; }
.svc-error-box.generic     .svc-error-title { color: #991b1b; }

.svc-error-desc {
  font-size: 13px;
  color: #475569;
  line-height: 1.6;
  max-width: 340px;
}

.svc-error-code {
  font-family: 'Courier New', monospace;
  font-size: 11px;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 20px;
  background: #f1f5f9;
  color: #64748b;
  letter-spacing: .04em;
}
.svc-error-box.auth        .svc-error-code { background: #e0e7ff; color: #4338ca; }
.svc-error-box.generic     .svc-error-code { background: #fee2e2; color: #b91c1c; }

.svc-retry-btn {
  margin-top: 6px;
  height: 36px;
  padding: 0 20px;
  border: 1.5px solid #e2e8f0;
  border-radius: 8px;
  background: #fff;
  color: #334155;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all .15s;
}
.svc-retry-btn:hover { background: #f8fafc; border-color: #94a3b8; }
</style>
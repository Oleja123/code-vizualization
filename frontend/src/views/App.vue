<template>
  <div id="app">
    <!-- Авторизация -->
    <div v-if="authLoading" class="auth-overlay">
      <div class="auth-card">
        <div class="spinner"></div>
        <p>Проверка авторизации…</p>
      </div>
    </div>

    <div v-else-if="!user" class="auth-overlay">
      <div class="auth-card">
        <div class="auth-header">
          <div class="logo">🔒</div>
          <div>
            <h2>{{ showRegister ? 'Регистрация' : 'Вход в систему' }}</h2>
            <p class="subtitle">
              {{ showRegister ? 'Создайте новый аккаунт' : 'Введите учётные данные' }}
            </p>
          </div>
        </div>

        <div class="auth-body">
          <label>Логин</label>
          <input
            v-model="usernameInput"
            type="text"
            placeholder="Имя пользователя"
            @keyup.enter="showRegister ? submitRegister() : submitLogin()"
            autocomplete="username"
          />

          <label style="margin-top:12px">Пароль</label>
          <input
            v-model="passwordInput"
            type="password"
            :placeholder="showRegister ? 'Минимум 6 символов' : 'Пароль'"
            @keyup.enter="showRegister ? submitRegister() : submitLogin()"
            :autocomplete="showRegister ? 'new-password' : 'current-password'"
          />

          <div class="auth-actions">
            <button
              class="btn-primary"
              :disabled="loginLoading"
              @click="showRegister ? submitRegister() : submitLogin()"
            >
              {{ loginLoading ? (showRegister ? 'Регистрация…' : 'Вход…') : (showRegister ? 'Зарегистрироваться' : 'Войти') }}
            </button>
          </div>

          <p class="error" v-if="authError">{{ authError }}</p>

          <div class="auth-toggle">
            <button class="link-btn" @click="toggleAuthMode">
              {{ showRegister ? 'Уже есть аккаунт? Войти' : 'Нет аккаунта? Зарегистрироваться' }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Основное приложение -->
    <template v-if="user">
      <header class="app-header">
        <h1>Визуализация кода</h1>
        <nav>
          <button
            :class="['nav-button', { active: activeView === 'tracer' }]"
            @click="activeView = 'tracer'"
          >
            🔍 Трассировка схемы
          </button>
          <button
            :class="['nav-button', { active: activeView === 'visualization' }]"
            @click="activeView = 'visualization'"
          >
            ⚡ Трассировка кода
          </button>
          <button
            :class="['nav-button', { active: activeView === 'flowchart' }]"
            @click="activeView = 'flowchart'"
          >
            📊 Блок-схема
          </button>
        </nav>
        <div class="user-info">
          <span class="username">👤 {{ user.username }}</span>
          <button class="logout-btn" @click="handleLogout">Выйти</button>
        </div>
      </header>
      <main class="app-main">
        <FlowchartTracer    v-if="activeView === 'tracer'" />
        <VisualizationView  v-if="activeView === 'visualization'" />
        <FlowchartBuilder   v-if="activeView === 'flowchart'" />
      </main>
    </template>
  </div>
</template>

<script>
import { ref, onMounted } from 'vue'
import VisualizationView from './views/VisualizationView.vue'
import FlowchartBuilder from './components/FlowchartBuilder.vue'
import FlowchartTracer from './components/FlowchartTracer.vue'
import { checkSession, login, register, logout } from '../api/auth.js'

export default {
  name: 'App',
  components: {
    VisualizationView,
    FlowchartBuilder,
    FlowchartTracer
  },
  setup() {
    const activeView = ref('tracer')

    // Auth
    const user = ref(null)
    const authLoading = ref(true)
    const loginLoading = ref(false)
    const usernameInput = ref('')
    const passwordInput = ref('')
    const authError = ref('')
    const showRegister = ref(false)

    onMounted(async () => {
      user.value = await checkSession()
      authLoading.value = false
    })

    async function submitLogin() {
      authError.value = ''
      loginLoading.value = true
      const result = await login(usernameInput.value, passwordInput.value)
      loginLoading.value = false
      if (result.ok) {
        user.value = await checkSession()
      } else {
        authError.value = result.message
      }
    }

    async function submitRegister() {
      authError.value = ''
      if (!usernameInput.value || !passwordInput.value) {
        authError.value = 'Заполните все поля'
        return
      }
      if (passwordInput.value.length < 6) {
        authError.value = 'Пароль должен содержать минимум 6 символов'
        return
      }
      loginLoading.value = true
      const result = await register(usernameInput.value, passwordInput.value)
      loginLoading.value = false
      if (result.ok) {
        const loginResult = await login(usernameInput.value, passwordInput.value)
        if (loginResult.ok) {
          user.value = await checkSession()
        }
      } else {
        authError.value = result.message
      }
    }

    async function handleLogout() {
      await logout()
      user.value = null
      usernameInput.value = ''
      passwordInput.value = ''
      authError.value = ''
      showRegister.value = false
    }

    function toggleAuthMode() {
      showRegister.value = !showRegister.value
      authError.value = ''
      usernameInput.value = ''
      passwordInput.value = ''
    }

    return {
      activeView,
      user,
      authLoading,
      loginLoading,
      usernameInput,
      passwordInput,
      authError,
      showRegister,
      submitLogin,
      submitRegister,
      handleLogout,
      toggleAuthMode
    }
  }
}
</script>

<style>
* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
  background: #f5f7fb;
  height: 100vh;
  overflow: hidden;
}

#app {
  height: 100vh;
  display: flex;
  flex-direction: column;
}

/* Header */
.app-header {
  background-color: #2c3e50;
  color: white;
  padding: 0.75rem 2rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  flex-shrink: 0;
}

.app-header h1 {
  font-size: 1.3rem;
  font-weight: 600;
}

.app-header nav {
  display: flex;
  gap: 0.5rem;
}

.nav-button {
  background-color: #34495e;
  color: white;
  border: none;
  padding: 0.4rem 1rem;
  border-radius: 4px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: background-color 0.2s;
}

.nav-button:hover { background-color: #415a77; }
.nav-button.active { background-color: #3498db; }

.user-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.username {
  color: white;
  font-size: 14px;
  font-weight: 500;
}

.logout-btn {
  background: #34495e;
  border: 1px solid #415a77;
  color: white;
  padding: 6px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  transition: all .2s;
  font-weight: 500;
}
.logout-btn:hover { background: #415a77; }

.app-main {
  flex: 1;
  overflow: hidden;
}

/* Auth overlay */
.auth-overlay {
  position: fixed;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0,0,0,.5);
  z-index: 9999;
}

.auth-card {
  background: linear-gradient(180deg,#fff 0%,#f7f9fb 100%);
  padding: 28px;
  border-radius: 12px;
  width: 440px;
  max-width: calc(100% - 40px);
  box-shadow: 0 12px 30px rgba(16,24,40,.35);
  border: 1px solid rgba(99,102,241,.08);
  text-align: center;
}

.auth-header {
  display: flex;
  gap: 12px;
  align-items: center;
  margin-bottom: 20px;
  text-align: left;
}

.logo { font-size: 32px; }

.auth-card h2 { margin: 0; font-size: 18px; color: #0f172a; }
.auth-card .subtitle { color: #6b7280; font-size: 13px; }

.auth-body { text-align: left; }

.auth-body label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 4px;
}

.auth-body input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  font-size: 14px;
  outline: none;
  transition: border-color .15s;
}

.auth-body input:focus { border-color: #6366f1; }

.auth-actions {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.btn-primary {
  background: #6366f1;
  color: white;
  padding: 10px 20px;
  border-radius: 8px;
  border: none;
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  transition: filter .15s;
}

.btn-primary:hover:not(:disabled) { filter: brightness(.93); }
.btn-primary:disabled { opacity: .6; cursor: not-allowed; }

.error { color: #b91c1c; margin-top: 10px; font-size: 13px; }

.auth-toggle {
  margin-top: 16px;
  text-align: center;
}

.link-btn {
  background: none;
  border: none;
  color: #6366f1;
  font-size: 13px;
  cursor: pointer;
  text-decoration: underline;
  padding: 4px;
  transition: opacity .15s;
}

.link-btn:hover { opacity: .8; }

/* Spinner */
.spinner {
  width: 36px;
  height: 36px;
  border: 4px solid #e5e7eb;
  border-top-color: #6366f1;
  border-radius: 50%;
  animation: spin .7s linear infinite;
  margin: 0 auto 14px;
}

@keyframes spin { to { transform: rotate(360deg); } }
</style>
// Сервис авторизации
// Поведение управляется переменной VITE_AUTH_ENABLED в .env:
//   true  — реальная авторизация через HTTP-сессию (порт 8083)
//   false — авторизация пропускается (удобно при тестировании)

const AUTH_ENABLED = import.meta.env.VITE_AUTH_ENABLED !== 'false'
const AUTH_URL = import.meta.env.VITE_AUTH_SERVICE_URL || 'http://localhost:8083'

/**
 * Проверяет текущую сессию пользователя.
 * Возвращает объект пользователя или null если не авторизован.
 */
export async function checkSession() {
  if (!AUTH_ENABLED) {
    return { username: 'dev', role: 'DEVELOPER' }
  }
  try {
    const res = await fetch(`${AUTH_URL}/api/auth/me`, {
      method: 'GET',
      credentials: 'include',
    })
    if (res.ok) {
      const username = await res.text()
      return { username }
    }
    return null
  } catch {
    return null
  }
}

/**
 * Регистрация нового пользователя.
 * Возвращает { ok: true } или { ok: false, message: '...' }
 */
export async function register(username, password) {
  if (!AUTH_ENABLED) {
    return { ok: true }
  }
  try {
    const res = await fetch(`${AUTH_URL}/api/auth/register`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        username, 
        rawPassword: password 
      }),
    })
    if (res.ok) {
      return { ok: true }
    }
    const message = await res.text().catch(() => 'Ошибка регистрации')
    return { ok: false, message }
  } catch {
    return { ok: false, message: 'Сервис авторизации недоступен' }
  }
}

/**
 * Логин через логин/пароль.
 * Возвращает { ok: true } или { ok: false, message: '...' }
 */
export async function login(username, password) {
  if (!AUTH_ENABLED) {
    return { ok: true }
  }
  try {
    const res = await fetch(`${AUTH_URL}/api/auth/login`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ 
        username, 
        rawPassword: password 
      }),
    })
    if (res.ok) {
      return { ok: true }
    }
    const message = await res.text().catch(() => 'Неверный логин или пароль')
    return { ok: false, message }
  } catch {
    return { ok: false, message: 'Сервис авторизации недоступен' }
  }
}

/**
 * Выход из системы.
 */
export async function logout() {
  if (!AUTH_ENABLED) return
  try {
    await fetch(`${AUTH_URL}/api/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    })
  } catch {
    // ignore
  }
}

export { AUTH_ENABLED }
export default { checkSession, register, login, logout, AUTH_ENABLED }

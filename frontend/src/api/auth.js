// Сервис авторизации
// Поведение управляется переменной VITE_AUTH_ENABLED в .env:
//   true  — реальная авторизация через HTTP-сессию (порт 8083)
//   false — авторизация пропускается (удобно при тестировании)

const AUTH_ENABLED = import.meta.env.VITE_AUTH_ENABLED !== 'false'
const AUTH_URL = import.meta.env.VITE_AUTH_SERVICE_URL || 'http://localhost:8083'

/**
 * Возвращает человекочитаемое сообщение об ошибке по HTTP-статусу.
 */
function httpErrorMessage(status) {
  if (status === 500 || status === 502 || status === 503 || status === 504) {
    return 'Сервис авторизации временно недоступен. Попробуйте позже.'
  }
  if (status === 401) return 'Неверный логин или пароль'
  if (status === 409) return 'Пользователь с таким именем уже существует'
  if (status === 400) return 'Некорректные данные. Проверьте логин и пароль.'
  return `Ошибка сервера (${status})`
}

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
    // При недоступном сервисе (500/502/503/504) не читаем тело — там может быть HTML
    if (res.status === 500 || (res.status >= 502 && res.status <= 504)) {
      return { ok: false, message: httpErrorMessage(res.status) }
    }
    const message = await res.text().catch(() => httpErrorMessage(res.status))
    return { ok: false, message: message || httpErrorMessage(res.status) }
  } catch {
    // fetch кидает исключение только при сетевом обрыве (офлайн, CORS, DNS)
    return { ok: false, message: 'Нет соединения с сервером. Проверьте подключение.' }
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
    // При недоступном сервисе (500/502/503/504) не читаем тело — там может быть HTML
    if (res.status === 500 || (res.status >= 502 && res.status <= 504)) {
      return { ok: false, message: httpErrorMessage(res.status) }
    }
    const message = await res.text().catch(() => httpErrorMessage(res.status))
    return { ok: false, message: message || httpErrorMessage(res.status) }
  } catch {
    // fetch кидает исключение только при сетевом обрыве (офлайн, CORS, DNS)
    return { ok: false, message: 'Нет соединения с сервером. Проверьте подключение.' }
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

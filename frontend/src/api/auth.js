const AUTH_ENABLED = import.meta.env.VITE_AUTH_ENABLED !== 'false'
const AUTH_URL = import.meta.env.VITE_AUTH_SERVICE_URL || 'http://localhost:8083'

function httpErrorMessage(status) {
  if (status === 500 || status === 502 || status === 503 || status === 504) {
    return 'Сервис авторизации временно недоступен. Попробуйте позже.'
  }
  if (status === 401) return 'Неверный логин или пароль'
  if (status === 409) return 'Пользователь с таким именем уже существует'
  if (status === 400) return 'Некорректные данные. Проверьте логин и пароль.'
  return `Ошибка сервера (${status})`
}

// Fetch с жёстким таймаутом — не даём зависнуть при недоступном сервисе
async function fetchWithTimeout(url, options = {}, timeoutMs = 5000) {
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(), timeoutMs)
  try {
    return await fetch(url, { ...options, signal: controller.signal })
  } finally {
    clearTimeout(timer)
  }
}

export async function checkSession() {
  if (!AUTH_ENABLED) {
    return { username: 'dev', role: 'DEVELOPER' }
  }
  try {
    const res = await fetchWithTimeout(`${AUTH_URL}/api/auth/me`, {
      method: 'GET',
      credentials: 'include',
    })
    if (res.ok) {
      const username = await res.text()
      return { username }
    }
    return null
  } catch {
    // Сеть недоступна или таймаут — не залогинен
    return null
  }
}

export async function register(username, password) {
  if (!AUTH_ENABLED) return { ok: true }
  try {
    const res = await fetchWithTimeout(`${AUTH_URL}/api/auth/register`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, rawPassword: password }),
    })
    if (res.ok) return { ok: true }
    if (res.status === 500 || (res.status >= 502 && res.status <= 504)) {
      return { ok: false, message: httpErrorMessage(res.status) }
    }
    const message = await res.text().catch(() => httpErrorMessage(res.status))
    return { ok: false, message: message || httpErrorMessage(res.status) }
  } catch (e) {
    if (e.name === 'AbortError') {
      return { ok: false, message: 'Сервис авторизации не отвечает. Попробуйте позже.' }
    }
    return { ok: false, message: 'Нет соединения с сервером. Проверьте подключение.' }
  }
}

export async function login(username, password) {
  if (!AUTH_ENABLED) return { ok: true }
  try {
    const res = await fetchWithTimeout(`${AUTH_URL}/api/auth/login`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, rawPassword: password }),
    })
    if (res.ok) return { ok: true }
    if (res.status === 500 || (res.status >= 502 && res.status <= 504)) {
      return { ok: false, message: httpErrorMessage(res.status) }
    }
    const message = await res.text().catch(() => httpErrorMessage(res.status))
    return { ok: false, message: message || httpErrorMessage(res.status) }
  } catch (e) {
    if (e.name === 'AbortError') {
      return { ok: false, message: 'Сервис авторизации не отвечает. Попробуйте позже.' }
    }
    return { ok: false, message: 'Нет соединения с сервером. Проверьте подключение.' }
  }
}

export async function logout() {
  if (!AUTH_ENABLED) return
  try {
    await fetchWithTimeout(`${AUTH_URL}/api/auth/logout`, {
      method: 'POST',
      credentials: 'include',
    }, 3000)
  } catch { /* ignore */ }
}

export { AUTH_ENABLED }
export default { checkSession, register, login, logout, AUTH_ENABLED }
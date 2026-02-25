// Сервис для генерации блок-схем через flowchart-visualizer (порт 8081)

const FLOWCHART_URL = import.meta.env.VITE_FLOWCHART_SERVICE_URL || 'http://localhost:8081'

/**
 * Генерирует SVG блок-схему из C-кода.
 * @param {string} code - исходный C-код
 * @returns {{ svg: string, ast: object, metadata: object } | null}
 */
export async function generateFromCode(code) {
  const res = await fetch(`${FLOWCHART_URL}/api/flowchart/generate-from-code`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }),
  })

  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err?.metadata?.error || `HTTP ${res.status}`)
  }

  return res.json()
}

/**
 * Проверяет доступность сервиса.
 * @returns {boolean}
 */
export async function isHealthy() {
  try {
    const res = await fetch(`${FLOWCHART_URL}/api/flowchart/health`)
    return res.ok
  } catch {
    return false
  }
}

export default { generateFromCode, isHealthy }

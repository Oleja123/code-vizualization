// Сервис для генерации блок-схем через flowchart-visualizer (порт 8081)

const FLOWCHART_URL = import.meta.env.VITE_FLOWCHART_SERVICE_URL || 'http://localhost:8081'

/**
 * Генерирует SVG блок-схему из C-кода.
 * Возвращает { svg, functions, ast, metadata }
 * где functions — объект { имяФункции: svgСтрока, ... } (если бэкенд поддерживает)
 * @param {string} code - исходный C-код
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
 * Генерирует SVG блок-схемы для каждой функции отдельно.
 * Возвращает { functions: { имяФункции: svgСтрока } }
 * @param {string} code - исходный C-код
 */
export async function generateAllFunctions(code) {
  const res = await fetch(`${FLOWCHART_URL}/api/flowchart/generate-all-functions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }),
  })

  if (!res.ok) {
    return null
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

export default { generateFromCode, generateAllFunctions, isHealthy }
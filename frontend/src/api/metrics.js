const METRICS_URL = import.meta.env.VITE_METRICS_SERVICE_URL || 'http://localhost:8085'

export async function calculateMetrics(code) {
  const res = await fetch(`${METRICS_URL}/api/metrics/calculate`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({}))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export async function getLatestMetrics() {
  const res = await fetch(`${METRICS_URL}/api/metrics/latest`, { credentials: 'include' })
  if (!res.ok) return []
  return res.json()
}

export default { calculateMetrics, getLatestMetrics }

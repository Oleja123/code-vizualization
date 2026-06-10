const FLOWCHART_URL = import.meta.env.VITE_FLOWCHART_SERVICE_URL || 'http://localhost:8081'

function handleError(res) {
  if (res.status === 401) throw new Error('AUTH_REQUIRED')
  if (res.status === 502 || res.status === 503) throw new Error('SERVICE_UNAVAILABLE')
}

async function safeJson(res) {
  const ct = res.headers.get('content-type') || ''
  if (!ct.includes('application/json')) throw new Error('SERVICE_UNAVAILABLE')
  return res.json().catch(() => ({}))
}

export async function generateFromCode(code) {
  const res = await fetch(`${FLOWCHART_URL}/api/flowchart/generate-from-code`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }),
  })
  if (!res.ok) {
    handleError(res)
    const err = await safeJson(res)
    throw new Error(err?.metadata?.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export async function generateAllFunctions(code) {
  const res = await fetch(`${FLOWCHART_URL}/api/flowchart/generate-all-functions`, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }),
  })
  if (!res.ok) {
    handleError(res)
    const err = await safeJson(res)
    throw new Error(err?.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export async function isHealthy() {
  try {
    const res = await fetch(`${FLOWCHART_URL}/api/flowchart/health`)
    return res.ok
  } catch {
    return false
  }
}

export default { generateFromCode, generateAllFunctions, isHealthy }
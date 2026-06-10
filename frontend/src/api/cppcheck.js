export async function analyzeCode(code) {
  const res = await fetch('/api/analyze', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }),
  })

  if (res.status === 401) throw new Error('AUTH_REQUIRED')
  if (res.status === 502 || res.status === 503) throw new Error('SERVICE_UNAVAILABLE')

  const contentType = res.headers.get('content-type') || ''
  if (!contentType.includes('application/json')) {
    throw new Error('SERVICE_UNAVAILABLE')
  }

  let data = null
  try { data = await res.json() } catch { throw new Error(`HTTP ${res.status}`) }

  if (!res.ok || !data.success) {
    throw new Error(data?.error || `HTTP ${res.status}`)
  }
  return data
}

export default { analyzeCode }
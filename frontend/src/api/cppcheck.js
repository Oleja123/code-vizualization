export async function analyzeCode(code) {
  const res = await fetch('/api/analyze', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }),
  })

  let data = null
  try {
    data = await res.json()
  } catch {
    throw new Error(`HTTP ${res.status}`)
  }

  if (!res.ok || !data.success) {
    throw new Error(data?.error || `HTTP ${res.status}`)
  }

  return data
}

export default { analyzeCode }

export async function getSnapshot(code, step) {
  const response = await fetch('/api/snapshot', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code, step }),
  })

  if (response.status === 401) throw new Error('AUTH_REQUIRED')
  if (response.status === 502 || response.status === 503) throw new Error('SERVICE_UNAVAILABLE')

  // Защита от HTML-ответов (nginx default error pages)
  const contentType = response.headers.get('content-type') || ''
  if (!contentType.includes('application/json')) {
    throw new Error('SERVICE_UNAVAILABLE')
  }

  const data = await response.json()
  if (!data.success) {
    throw new Error(data.error || 'Неизвестная ошибка')
  }
  return data
}
export async function getSnapshot(code, step) {
  console.log('API call: getSnapshot', { code: code.substring(0, 50) + '...', step })
  const response = await fetch('/api/snapshot', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ code, step }),
  })

  const data = await response.json()
  console.log('API response:', data)
  
  if (!data.success) {
    throw new Error(data.error || 'Неизвестная ошибка')
  }

  return data
}

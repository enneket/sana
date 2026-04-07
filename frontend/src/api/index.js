const API_BASE = '/api'

function getToken() {
  return localStorage.getItem('token')
}

function fetchWithAuth(url, options = {}) {
  const token = getToken()
  const headers = options.headers || {}
  if (token) {
    headers['Authorization'] = `Bearer ${token}`
  }
  return fetch(url, { ...options, headers })
}

async function handleResponse(r) {
  if (r.status === 401) {
    localStorage.removeItem('token')
    throw new Error('unauthorized')
  }
  if (r.status === 204) return null
  if (!r.ok) {
    const err = await r.json().catch(() => ({ error: 'request failed' }))
    throw new Error(err.error || `HTTP ${r.status}`)
  }
  return r.json()
}

const api = {
  // Auth
  login: (password) => fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    body: JSON.stringify({ password }),
    headers: { 'Content-Type': 'application/json' },
  }).then(r => r.json()),

  logout: () => {
    localStorage.removeItem('token')
    return fetch(`${API_BASE}/auth/logout`, { method: 'POST' })
  },

  me: () => fetchWithAuth(`${API_BASE}/auth/me`).then(handleResponse),

  // Sanas (Timeline)
  listMemos: (cursor) => {
    let url = `${API_BASE}/sanas?limit=20`
    if (cursor) url += `&cursor=${cursor}`
    return fetchWithAuth(url).then(handleResponse)
  },

  createMemo: (content) => fetchWithAuth(`${API_BASE}/sanas`, {
    method: 'POST',
    body: JSON.stringify({ content }),
    headers: { 'Content-Type': 'application/json' },
  }).then(handleResponse),

  getMemo: (id) => fetchWithAuth(`${API_BASE}/sanas/${id}`).then(handleResponse),

  updateMemo: (id, content) => fetchWithAuth(`${API_BASE}/sanas/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ content }),
    headers: { 'Content-Type': 'application/json' },
  }).then(handleResponse),

  deleteMemo: (id) => fetchWithAuth(`${API_BASE}/sanas/${id}`, { method: 'DELETE' }).then(handleResponse),

  searchMemos: (q) => fetchWithAuth(`${API_BASE}/sanas/search?q=${encodeURIComponent(q)}`)
    .then(handleResponse),

  getStats: () => fetchWithAuth(`${API_BASE}/sanas/stats`).then(handleResponse),

  exportMemos: () => fetchWithAuth(`${API_BASE}/export/sanas`).then(r => r.blob()),

  importMemos: (formData) => fetchWithAuth(`${API_BASE}/import/sanas`, {
    method: 'POST',
    body: formData,
  }).then(handleResponse),
}

export default api

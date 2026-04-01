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

function handleResponse(r) {
  if (r.status === 401) {
    localStorage.removeItem('token')
    window.location.href = '/login'
    throw new Error('unauthorized')
  }
  if (r.status === 204) return null
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

  // Memos (Timeline)
  listMemos: (cursor) => {
    let url = `${API_BASE}/memos?limit=20`
    if (cursor) url += `&cursor=${cursor}`
    return fetchWithAuth(url).then(handleResponse)
  },

  createMemo: (content) => fetchWithAuth(`${API_BASE}/memos`, {
    method: 'POST',
    body: JSON.stringify({ content }),
    headers: { 'Content-Type': 'application/json' },
  }).then(handleResponse),

  getMemo: (id) => fetchWithAuth(`${API_BASE}/memos/${id}`).then(handleResponse),

  updateMemo: (id, content) => fetchWithAuth(`${API_BASE}/memos/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ content }),
    headers: { 'Content-Type': 'application/json' },
  }).then(handleResponse),

  deleteMemo: (id) => fetchWithAuth(`${API_BASE}/memos/${id}`, { method: 'DELETE' }).then(handleResponse),

  searchMemos: (q) => fetchWithAuth(`${API_BASE}/memos/search?q=${encodeURIComponent(q)}`)
    .then(handleResponse),

  exportMemos: () => fetchWithAuth(`${API_BASE}/export/memos`).then(r => r.blob()),

  importMemos: (formData) => fetchWithAuth(`${API_BASE}/import/memos`, {
    method: 'POST',
    body: formData,
  }).then(handleResponse),
}

export default api

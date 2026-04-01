const API_BASE = '/api'

const api = {
  // Auth
  login: (password) => fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    body: JSON.stringify({ password }),
    headers: { 'Content-Type': 'application/json' },
  }).then(r => r.json()),

  logout: () => fetch(`${API_BASE}/auth/logout`, { method: 'POST' }),

  me: () => fetch(`${API_BASE}/auth/me`).then(r => r.json()),

  // Memos (Timeline)
  listMemos: (cursor) => {
    let url = `${API_BASE}/memos?limit=20`
    if (cursor) url += `&cursor=${cursor}`
    return fetch(url).then(r => r.json())
  },

  createMemo: (content) => fetch(`${API_BASE}/memos`, {
    method: 'POST',
    body: JSON.stringify({ content }),
    headers: { 'Content-Type': 'application/json' },
  }).then(r => r.json()),

  getMemo: (id) => fetch(`${API_BASE}/memos/${id}`).then(r => r.json()),

  updateMemo: (id, content) => fetch(`${API_BASE}/memos/${id}`, {
    method: 'PUT',
    body: JSON.stringify({ content }),
    headers: { 'Content-Type': 'application/json' },
  }).then(r => r.json()),

  deleteMemo: (id) => fetch(`${API_BASE}/memos/${id}`, { method: 'DELETE' }),

  searchMemos: (q) => fetch(`${API_BASE}/memos/search?q=${encodeURIComponent(q)}`)
    .then(r => r.json()),

  exportMemos: () => fetch(`${API_BASE}/export/memos`).then(r => r.blob()),

  importMemos: (formData) => fetch(`${API_BASE}/import/memos`, {
    method: 'POST',
    body: formData,
  }).then(r => r.json()),
}

export default api

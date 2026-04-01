import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import './style.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('./views/Login.vue') },
    {
      path: '/',
      component: () => import('./views/TimelineView.vue'),
      meta: { requiresAuth: true },
    }
  ]
})

const API = import.meta.env.VITE_API_URL || '/api'

let _token = localStorage.getItem('token')

export async function apiFetch(path, options = {}) {
  const headers = { 'Content-Type': 'application/json' }
  if (_token) headers['Authorization'] = `Bearer ${_token}`
  const res = await fetch(`${API}${path}`, { ...options, headers })
  if (res.status === 401) {
    _token = null
    localStorage.removeItem('token')
    router.push('/login')
    throw new Error('unauthorized')
  }
  if (res.status === 204) return null
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: 'request failed' }))
    throw new Error(err.error || `HTTP ${res.status}`)
  }
  return res.json()
}

export function setToken(t) {
  _token = t
  localStorage.setItem('token', t)
}

export function clearToken() {
  _token = null
  localStorage.removeItem('token')
}

export function isAuthenticated() {
  _token = localStorage.getItem('token')
  return !!_token
}

const app = createApp(App)
app.provide('api', API)
app.use(router)
app.mount('#app')

router.beforeEach((to) => {
  if (to.meta.requiresAuth && !isAuthenticated()) {
    return '/login'
  }
  if (!to.meta.requiresAuth && isAuthenticated() && to.path === '/login') {
    return '/'
  }
})

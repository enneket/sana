<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch, clearToken } from '../main.js'
import FolderSidebar from '../components/FolderSidebar.vue'

const router = useRouter()
const user = ref(null)

onMounted(async () => {
  try {
    user.value = await apiFetch('/auth/me')
  } catch {
    router.push('/login')
  }
})

async function logout() {
  await apiFetch('/auth/logout', { method: 'POST' })
  clearToken()
  router.push('/login')
}
</script>

<template>
  <div class="layout">
    <header class="header">
      <span class="logo">Sana</span>
      <div class="user-menu">
        <span v-if="user">{{ user.username }}</span>
        <button @click="logout">Logout</button>
      </div>
    </header>
    <div class="body">
      <FolderSidebar class="sidebar" />
      <main class="main">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style scoped>
.layout { min-height: 100vh; display: flex; flex-direction: column; background: #1a1a2e; }
.header {
  height: 48px;
  padding: 0 20px;
  background: #16162a;
  border-bottom: 1px solid #2a2a4a;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}
.logo { color: #e8e8e8; font-weight: 600; font-size: 16px; }
.user-menu { display: flex; align-items: center; gap: 12px; color: #888; font-size: 13px; }
.user-menu button {
  background: none;
  border: 1px solid #2a2a4a;
  color: #888;
  padding: 4px 12px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
}
.user-menu button:hover { border-color: #4a9eff; color: #4a9eff; }
.body { display: flex; flex: 1; overflow: hidden; }
.sidebar { width: 220px; flex-shrink: 0; border-right: 1px solid #2a2a4a; overflow-y: auto; }
.main { flex: 1; overflow-y: auto; }
</style>

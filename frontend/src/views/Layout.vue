<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch, clearToken } from '../main.js'
import TreeView from '../components/TreeView.vue'
import NoteView from './NoteView.vue'

const router = useRouter()
const selectedNote = ref(null)
const treeRefreshKey = ref(0)

function onSelectNote(note) {
  selectedNote.value = note
}

async function onNoteDeleted() {
  selectedNote.value = null
}

function onNoteSaved() {
  treeRefreshKey.value++
}

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
        <button @click="logout">退出</button>
      </div>
    </header>
    <div class="body">
      <TreeView
        class="sidebar"
        :selected-note-id="selectedNote?.id"
        :refresh="treeRefreshKey"
        @select-note="onSelectNote"
      />
      <main class="main">
        <NoteView
          v-if="selectedNote"
          :note="selectedNote"
          @deleted="onNoteDeleted"
          @saved="onNoteSaved"
        />
        <div v-else class="empty-state">
          <p>选择或新建一个笔记</p>
        </div>
      </main>
    </div>
  </div>
</template>

<style scoped>
.layout { min-height: 100vh; display: flex; flex-direction: column; background: #F5F0E8; }
.header {
  height: 40px;
  padding: 0 16px;
  background: #EDE8DF;
  border-bottom: 1px solid #DDD8CC;
  display: flex;
  align-items: center;
  justify-content: space-between;
  flex-shrink: 0;
}
.logo { color: #3D3830; font-weight: 600; font-size: 14px; letter-spacing: -0.3px; }
.user-menu { display: flex; align-items: center; gap: 12px; }
.user-menu button {
  background: none;
  border: 1px solid #D4CCBA;
  color: #8A8478;
  padding: 3px 10px;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  transition: border-color 0.15s, color 0.15s;
}
.user-menu button:hover { border-color: #6B8FCC; color: #6B8FCC; }
.body { display: flex; flex: 1; overflow: hidden; }
.sidebar { width: 240px; flex-shrink: 0; border-right: 1px solid #DDD8CC; overflow-y: auto; }
.main { flex: 1; overflow-y: auto; background: #F5F0E8; }
.empty-state {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ADA99F;
  font-size: 14px;
}
</style>

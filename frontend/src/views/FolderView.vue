<script setup>
import { ref, watch, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiFetch } from '../main.js'

const route = useRoute()
const router = useRouter()
const notes = ref([])

const folderId = () => route.params.folderId

async function loadNotes() {
  if (!folderId()) {
    notes.value = []
    return
  }
  notes.value = await apiFetch(`/notes?folder_id=${folderId()}`)
}

onMounted(loadNotes)
watch(() => route.params.folderId, loadNotes)

async function createNote() {
  if (!folderId()) return
  const data = await apiFetch('/notes', {
    method: 'POST',
    body: JSON.stringify({ title: '无标题', content: '', folder_id: folderId() })
  })
  router.push(`/note/${data.id}`)
}

function openNote(id) {
  router.push(`/note/${id}`)
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
}
</script>

<template>
  <div class="folder-view">
    <div class="topbar">
      <h2>笔记</h2>
      <button v-if="folderId()" @click="createNote">+ 新建笔记</button>
    </div>

    <div v-if="!folderId()" class="empty">
      <p>选择一个文件夹查看笔记</p>
    </div>

    <div v-if="!notes || notes.length === 0 && folderId()" class="empty">
      <p>暂无笔记，创建第一篇吧</p>
    </div>

    <div class="note-list">
      <div
        v-for="note in (notes || [])"
        :key="note.id"
        class="note-item"
        @click="openNote(note.id)"
      >
        <div class="note-title">{{ note.title }}</div>
        <div class="note-date">{{ formatDate(note.updated_at) }}</div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.folder-view { padding: 24px; max-width: 600px; }
.topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
h2 { color: #e8e8e8; margin: 0; font-size: 18px; font-weight: 500; }
.topbar button {
  background: #4a9eff;
  border: none;
  color: #fff;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
}
.topbar button:hover { background: #3a8eef; }
.new-note-form {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
}
.new-note-form input {
  flex: 1;
  background: #16162a;
  border: 1px solid #2a2a4a;
  border-radius: 6px;
  color: #e8e8e8;
  padding: 10px 12px;
  font-size: 14px;
}
.new-note-form input:focus { outline: none; border-color: #4a9eff; }
.new-note-form button {
  background: #4a9eff;
  border: none;
  color: #fff;
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 13px;
  cursor: pointer;
}
.empty { color: #666; text-align: center; padding: 40px 0; }
.empty p { margin: 0; font-size: 14px; }
.note-list { display: flex; flex-direction: column; gap: 4px; }
.note-item {
  padding: 14px 16px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.15s;
  background: #16162a;
  border: 1px solid #2a2a4a;
}
.note-item:hover { background: #1e1e38; border-color: #3a3a5a; }
.note-title { color: #e8e8e8; font-size: 14px; font-weight: 500; margin-bottom: 4px; }
.note-date { color: #666; font-size: 12px; }
</style>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch } from '../main.js'

const router = useRouter()
const folders = ref([])
const newFolderName = ref('')
const showNewFolder = ref(false)
const expandedFolders = ref(new Set())
const importInput = ref(null)
const importMessage = ref('')

async function refreshTree() {
  await loadFolders()
}

async function exportNotes() {
  const API = import.meta.env.VITE_API_URL || '/api'
  const token = localStorage.getItem('token')
  if (!token) return
  const res = await fetch(`${API}/export`, {
    headers: { 'Authorization': `Bearer ${token}` }
  })
  if (!res.ok) {
    alert('导出失败')
    return
  }
  const blob = await res.blob()
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  const date = new Date().toISOString().slice(0, 10).replace(/-/g, '')
  a.download = `sana_export_${date}.zip`
  a.click()
  URL.revokeObjectURL(url)
}

function triggerImport() {
  importInput.value.click()
}

async function onImportFile(event) {
  const file = event.target.files[0]
  if (!file) return
  const API = import.meta.env.VITE_API_URL || '/api'
  const token = localStorage.getItem('token')
  const formData = new FormData()
  formData.append('file', file)
  let message = ''
  try {
    const res = await fetch(`${API}/import`, {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${token}` },
      body: formData
    })
    const data = await res.json()
    if (res.ok) {
      const skipped = data.skipped && data.skipped.length > 0
        ? `，${data.skipped.length} 个文件跳过`
        : ''
      message = `导入完成：${data.folders_imported} 个文件夹、${data.notes_imported} 个笔记${skipped}`
      await refreshTree()
    } else {
      message = `导入失败：${data.error || res.status}`
    }
  } catch (e) {
    message = `导入失败：${e.message}`
  }
  importMessage.value = message
  setTimeout(() => { importMessage.value = '' }, 5000)
  event.target.value = '' // reset so same file can be re-selected
}

onMounted(loadFolders)

async function loadFolders() {
  folders.value = await apiFetch('/folders')
}

const rootFolders = computed(() => (folders.value || []).filter(f => !f.parent_id))

function selectFolder(id) {
  router.push(`/folder/${id}`)
}

async function createFolder() {
  if (!newFolderName.value.trim()) return
  await apiFetch('/folders', {
    method: 'POST',
    body: JSON.stringify({ name: newFolderName.value.trim() })
  })
  newFolderName.value = ''
  showNewFolder.value = false
  await loadFolders()
}

async function deleteFolder(id, e) {
  e.stopPropagation()
  if (!confirm('删除文件夹及其所有笔记？')) return
  await apiFetch(`/folders/${id}`, { method: 'DELETE' })
  router.push('/folders')
  await loadFolders()
}

function getChildren(parentId) {
  return folders.value.filter(f => f.parent_id === parentId)
}

defineExpose({ refreshTree })
</script>

<template>
  <div class="sidebar">
    <div class="sidebar-header">
      <span>文件夹</span>
      <div class="sidebar-actions">
        <button @click="exportNotes" title="导出所有笔记">↓</button>
        <button @click="triggerImport" title="导入笔记">↑</button>
        <button @click="showNewFolder = !showNewFolder" title="新建文件夹">+</button>
      </div>
    </div>
    <div v-if="importMessage" class="import-message">{{ importMessage }}</div>
    <input ref="importInput" type="file" accept=".zip" style="display:none" @change="onImportFile" />

    <div v-if="showNewFolder" class="new-folder">
      <input v-model="newFolderName" placeholder="文件夹名称" @keyup.enter="createFolder" />
      <button @click="createFolder">添加</button>
    </div>

    <div class="folder-list">
      <div
        v-for="folder in rootFolders"
        :key="folder.id"
        class="folder-item"
        @click="selectFolder(folder.id)"
      >
        <span class="folder-icon">📁</span>
        <span class="folder-name">{{ folder.name }}</span>
        <button class="delete-btn" @click="deleteFolder(folder.id, $event)">×</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.sidebar {
  background: #16162a;
  height: 100%;
  padding: 12px;
}
.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  color: #888;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.sidebar-actions { display: flex; gap: 4px; }
.sidebar-actions button {
  background: none;
  border: none;
  color: #4a9eff;
  cursor: pointer;
  font-size: 14px;
  padding: 0 2px;
}
.import-message {
  font-size: 11px;
  color: #888;
  padding: 4px 8px;
  margin-bottom: 8px;
  background: #1a1a2e;
  border-radius: 4px;
}
.sidebar-header button {
  background: none;
  border: none;
  color: #4a9eff;
  cursor: pointer;
  font-size: 18px;
  padding: 0 4px;
}
.new-folder {
  display: flex;
  gap: 6px;
  margin-bottom: 12px;
}
.new-folder input {
  flex: 1;
  background: #1a1a2e;
  border: 1px solid #2a2a4a;
  border-radius: 4px;
  color: #e8e8e8;
  padding: 6px 8px;
  font-size: 13px;
}
.new-folder input:focus { outline: none; border-color: #4a9eff; }
.new-folder button {
  background: #4a9eff;
  border: none;
  border-radius: 4px;
  color: #fff;
  padding: 4px 10px;
  font-size: 12px;
  cursor: pointer;
}
.folder-list { display: flex; flex-direction: column; gap: 2px; }
.folder-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  color: #c8c8c8;
  font-size: 13px;
  transition: background 0.15s;
}
.folder-item:hover { background: #1e1e38; }
.folder-item:hover .delete-btn { opacity: 1; }
.folder-icon { font-size: 14px; }
.folder-name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.delete-btn {
  opacity: 0;
  background: none;
  border: none;
  color: #ff6b6b;
  cursor: pointer;
  font-size: 14px;
  padding: 0 2px;
  transition: opacity 0.15s;
}
</style>

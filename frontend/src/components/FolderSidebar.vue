<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch } from '../main.js'

const router = useRouter()
const folders = ref([])
const newFolderName = ref('')
const showNewFolder = ref(false)
const expandedFolders = ref(new Set())

onMounted(loadFolders)

async function loadFolders() {
  folders.value = await apiFetch('/folders')
}

const rootFolders = computed(() => folders.value.filter(f => !f.parent_id))

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
  if (!confirm('Delete folder and all notes inside?')) return
  await apiFetch(`/folders/${id}`, { method: 'DELETE' })
  router.push('/folders')
  await loadFolders()
}

function getChildren(parentId) {
  return folders.value.filter(f => f.parent_id === parentId)
}
</script>

<template>
  <div class="sidebar">
    <div class="sidebar-header">
      <span>Folders</span>
      <button @click="showNewFolder = !showNewFolder" title="New folder">+</button>
    </div>

    <div v-if="showNewFolder" class="new-folder">
      <input v-model="newFolderName" placeholder="Folder name" @keyup.enter="createFolder" />
      <button @click="createFolder">Add</button>
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

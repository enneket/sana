<script setup>
import { ref, onMounted, watch } from 'vue'
import { apiFetch } from '../main.js'
import TreeNode from './TreeNode.vue'

const props = defineProps({
  selectedNoteId: String,
  refresh: Number,
})

const emit = defineEmits(['select-note'])

const folders = ref([])
const notes = ref([])
const expanded = ref(new Set())
const contextMenu = ref(null)
const newFolderInput = ref('')
const creatingIn = ref(null)

onMounted(loadData)
watch(() => props.refresh, loadData)

async function loadData() {
  folders.value = await apiFetch('/folders')
  const allNotes = []
  for (const f of folders.value) {
    try {
      const n = await apiFetch(`/notes?folder_id=${f.id}`)
      allNotes.push(...(n || []))
    } catch {}
  }
  notes.value = allNotes
}

function toggleFolder(id) {
  if (expanded.value.has(id)) {
    expanded.value.delete(id)
  } else {
    expanded.value.add(id)
  }
}

function selectNote(note) {
  emit('select-note', note)
}

function getNotesForFolder(folderId) {
  return notes.value.filter(n => n.folder_id === folderId)
}

function getRootFolders() {
  return folders.value.filter(f => !f.parent_id)
}

function getChildrenFolders(parentId) {
  return folders.value.filter(f => f.parent_id === parentId)
}

function onContextMenuFolder(e, folder) {
  e.preventDefault()
  e.stopPropagation()
  contextMenu.value = { type: 'folder', x: e.clientX, y: e.clientY, folder }
}

function onContextMenuNote(e, note) {
  e.preventDefault()
  e.stopPropagation()
  contextMenu.value = { type: 'note', x: e.clientX, y: e.clientY, note }
}

function closeContextMenu() {
  contextMenu.value = null
}

async function newNoteInFolder(folder) {
  closeContextMenu()
  const data = await apiFetch('/notes', {
    method: 'POST',
    body: JSON.stringify({ title: '无标题', content: '', folder_id: folder.id })
  })
  await loadData()
  emit('select-note', { id: data.id, title: data.title })
}

async function newFolderUnder(parentId) {
  closeContextMenu()
  expanded.value.add(parentId)
  creatingIn.value = parentId
  newFolderInput.value = ''
}

async function createNewFolder() {
  if (!newFolderInput.value.trim()) {
    creatingIn.value = null
    return
  }
  await apiFetch('/folders', {
    method: 'POST',
    body: JSON.stringify({ name: newFolderInput.value.trim(), parent_id: creatingIn.value === 'root' ? null : creatingIn.value })
  })
  creatingIn.value = null
  newFolderInput.value = ''
  await loadData()
}

async function deleteFolder(folder, e) {
  e.stopPropagation()
  closeContextMenu()
  if (!confirm('删除文件夹及其所有笔记？')) return
  await apiFetch(`/folders/${folder.id}`, { method: 'DELETE' })
  await loadData()
}

async function deleteNote(note, e) {
  e.stopPropagation()
  closeContextMenu()
  if (!confirm('删除此笔记？')) return
  await apiFetch(`/notes/${note.id}`, { method: 'DELETE' })
  await loadData()
}

function onBodyClick() {
  closeContextMenu()
}
</script>

<template>
  <div class="tree" @click="closeContextMenu">
    <div class="tree-header">
      <span class="tree-title">Sana</span>
      <button class="add-btn" @click="creatingIn = 'root'" title="新建文件夹">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round">
          <path d="M12 5v14M5 12h14"/>
        </svg>
      </button>
    </div>

    <div v-if="creatingIn === 'root'" class="new-folder-row">
      <input
        v-model="newFolderInput"
        placeholder="文件夹名称"
        class="new-folder-input"
        @keyup.enter="createNewFolder"
        @keyup.escape="creatingIn = null"
        autofocus
      />
    </div>

    <div class="tree-content">
      <div v-for="folder in getRootFolders()" :key="folder.id">
        <!-- Root folder -->
        <div class="tree-row is-folder" @click="toggleFolder(folder.id)" @contextmenu="onContextMenuFolder($event, folder)">
          <svg class="expand-icon" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <path v-if="expanded.has(folder.id)" d="M6 9l6 6 6-6"/>
            <path v-else d="M9 18l6-6-6-6"/>
          </svg>
          <svg class="folder-icon" width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
            <path d="M2 6c0-1.1.9-2 2-2h5l2 2h9c1.1 0 2 .9 2 2v10c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6z"/>
          </svg>
          <span class="label">{{ folder.name }}</span>
        </div>

        <!-- Root folder expanded content -->
        <div v-if="expanded.has(folder.id)">
          <!-- Child folders -->
          <div v-for="child in getChildrenFolders(folder.id)" :key="child.id">
            <TreeNode
              :folder="child"
              :selected-note-id="selectedNoteId"
              :depth="1"
              :expanded="expanded"
              :creating-in="creatingIn"
              :new-folder-input="newFolderInput"
              :folders="folders"
              :notes="notes"
              @toggle="toggleFolder"
              @select-note="selectNote"
              @ctx-note="onContextMenuNote"
              @ctx-folder="onContextMenuFolder"
              @new-folder-under="newFolderUnder"
              @create-folder="createNewFolder"
            />
          </div>

          <!-- Notes in root folder -->
          <div
            v-for="note in getNotesForFolder(folder.id)"
            :key="note.id"
            class="tree-row"
            :class="{ active: note.id === selectedNoteId }"
            style="padding-left: 28px"
            @click="selectNote(note)"
            @contextmenu="onContextMenuNote($event, note)"
          >
            <svg class="file-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
            </svg>
            <span class="label">{{ note.title }}</span>
          </div>

          <div v-if="creatingIn === folder.id" class="new-folder-row" style="padding-left: 28px">
            <input
              v-model="newFolderInput"
              placeholder="文件夹名称"
              class="new-folder-input"
              @keyup.enter="createNewFolder"
              @keyup.escape="creatingIn = null"
              autofocus
            />
          </div>
        </div>
      </div>
    </div>

    <Teleport to="body">
      <div
        v-if="contextMenu"
        class="ctx-menu"
        :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
        @click.stop
      >
        <template v-if="contextMenu.type === 'folder'">
          <div class="ctx-item" @click="newNoteInFolder(contextMenu.folder)">新建笔记</div>
          <template v-if="getRootFolders().some(f => f.id === contextMenu.folder?.id)">
            <div class="ctx-item" @click="newFolderUnder(contextMenu.folder.id)">新建子文件夹</div>
            <div class="ctx-sep"></div>
          </template>
          <div class="ctx-item danger" @click="deleteFolder(contextMenu.folder, $event)">删除</div>
        </template>
        <template v-else-if="contextMenu.type === 'note'">
          <div class="ctx-item danger" @click="deleteNote(contextMenu.note, $event)">删除</div>
        </template>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.tree {
  background: #EAE5DC;
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', sans-serif;
}
.tree-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px 10px 16px;
  flex-shrink: 0;
  border-bottom: 1px solid #DDD8CC;
}
.tree-title {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.08em;
  color: #A09888;
}
.add-btn {
  background: none;
  border: none;
  color: #A09888;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  transition: color 0.15s, background 0.15s;
}
.add-btn:hover { color: #6B8FCC; background: #E0DBD0; }

.tree-content {
  flex: 1;
  overflow-y: auto;
  padding: 4px 0 8px;
}
.tree-content::-webkit-scrollbar { width: 6px; }
.tree-content::-webkit-scrollbar-track { background: transparent; }
.tree-content::-webkit-scrollbar-thumb { background: #D4CCBA; border-radius: 3px; }
.tree-content::-webkit-scrollbar-thumb:hover { background: #C4BA9E; }

.tree-row {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 5px 12px 5px 14px;
  cursor: pointer;
  color: #6B6458;
  font-size: 13px;
  user-select: none;
  transition: background 0.1s, color 0.1s;
  position: relative;
}
.tree-row.is-folder { color: #4A453E; }
.tree-row:hover { background: #E0DBD0; color: #3D3830; }
.tree-row.active { background: #DDD8CC; color: #3D3830; }
.tree-row.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 2px;
  background: #6B8FCC;
  border-radius: 0 2px 2px 0;
}
.expand-icon { flex-shrink: 0; color: #B8AFA0; transition: transform 0.1s; }
.folder-icon { flex-shrink: 0; color: #9A8E7E; }
.file-icon { flex-shrink: 0; color: #9A9080; }
.label { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.new-folder-row {
  padding: 3px 12px 3px 14px;
}
.new-folder-input {
  width: 100%;
  background: #F5F0E8;
  border: 1px solid #C4BA9E;
  border-radius: 4px;
  color: #3D3830;
  padding: 4px 8px;
  font-size: 13px;
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.15s;
}
.new-folder-input:focus { border-color: #6B8FCC; }
.new-folder-input::placeholder { color: #B8AFA0; }

/* Context menu */
.ctx-menu {
  position: fixed;
  background: #FAFAF7;
  border: 1px solid #D4CCBA;
  border-radius: 6px;
  padding: 4px 0;
  min-width: 150px;
  box-shadow: 0 4px 16px rgba(60,50,40,0.12);
  z-index: 9999;
}
.ctx-item {
  padding: 7px 14px;
  font-size: 13px;
  color: #6B6458;
  cursor: pointer;
  transition: background 0.1s, color 0.1s;
}
.ctx-item:hover { background: #F0EDE6; color: #3D3830; }
.ctx-item.danger { color: #C06050; }
.ctx-item.danger:hover { background: #F8EDE8; }
.ctx-sep { height: 1px; background: #E8E2D8; margin: 4px 0; }
</style>

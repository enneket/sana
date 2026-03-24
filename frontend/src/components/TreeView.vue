<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { apiFetch } from '../main.js'

const props = defineProps({
  selectedNoteId: String,
  refresh: Number,
})

const emit = defineEmits(['select-note'])

const folders = ref([])
const notes = ref([])
const expanded = ref(new Set())
const contextMenu = ref(null) // { x, y, folderId } or { x, y, noteId }
const newFolderInput = ref('')
const creatingIn = ref(null) // folderId or 'root'

onMounted(loadData)
watch(() => props.refresh, loadData)

async function loadData() {
  folders.value = await apiFetch('/folders')
  // Load notes for all folders at once
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
  contextMenu.value = { type: 'folder', x: e.clientX, y: e.clientY, folder }
}

function onContextMenuNote(e, note) {
  e.preventDefault()
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

// Close context menu on outside click
function onBodyClick() {
  closeContextMenu()
}
</script>

<template>
  <div class="tree" @click="closeContextMenu">
    <!-- Header -->
    <div class="tree-header">
      <span>Sana</span>
      <button @click="creatingIn = 'root'" title="新建文件夹">+</button>
    </div>

    <!-- New folder inline input -->
    <div v-if="creatingIn === 'root'" class="new-folder-inline">
      <input
        v-model="newFolderInput"
        placeholder="文件夹名称"
        @keyup.enter="createNewFolder"
        @keyup.escape="creatingIn = null"
        autofocus
      />
      <button @click="createNewFolder">✓</button>
      <button @click="creatingIn = null">×</button>
    </div>

    <!-- Tree items -->
    <div class="tree-content">
      <div v-for="folder in getRootFolders()" :key="folder.id" class="tree-node">
        <!-- Folder row -->
        <div
          class="tree-row folder-row"
          @click="toggleFolder(folder.id)"
          @contextmenu="onContextMenuFolder($event, folder)"
        >
          <span class="expand-icon">{{ expanded.has(folder.id) ? '▼' : '▶' }}</span>
          <span class="icon">📁</span>
          <span class="label">{{ folder.name }}</span>
        </div>

        <!-- Expanded: children folders -->
        <div v-if="expanded.has(folder.id)" class="tree-children">
          <div v-for="child in getChildrenFolders(folder.id)" :key="child.id" class="tree-node">
            <div
              class="tree-row folder-row indent1"
              @click="toggleFolder(child.id)"
              @contextmenu="onContextMenuFolder($event, child)"
            >
              <span class="expand-icon">{{ expanded.has(child.id) ? '▼' : '▶' }}</span>
              <span class="icon">📁</span>
              <span class="label">{{ child.name }}</span>
            </div>

            <!-- Child's notes -->
            <div v-if="expanded.has(child.id)" class="tree-children">
              <div
                v-for="note in getNotesForFolder(child.id)"
                :key="note.id"
                class="tree-row note-row indent2"
                :class="{ active: note.id === selectedNoteId }"
                @click="selectNote(note)"
                @contextmenu="onContextMenuNote($event, note)"
              >
                <span class="icon">📄</span>
                <span class="label">{{ note.title }}</span>
              </div>
            </div>
          </div>

          <!-- Inline new folder under this folder -->
          <div v-if="creatingIn === folder.id" class="new-folder-inline indent1">
            <input
              v-model="newFolderInput"
              placeholder="文件夹名称"
              @keyup.enter="createNewFolder"
              @keyup.escape="creatingIn = null"
              autofocus
            />
            <button @click="createNewFolder">✓</button>
            <button @click="creatingIn = null">×</button>
          </div>

          <!-- Notes in this folder -->
          <div
            v-for="note in getNotesForFolder(folder.id)"
            :key="note.id"
            class="tree-row note-row indent1"
            :class="{ active: note.id === selectedNoteId }"
            @click="selectNote(note)"
            @contextmenu="onContextMenuNote($event, note)"
          >
            <span class="icon">📄</span>
            <span class="label">{{ note.title }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Context menu -->
    <div
      v-if="contextMenu"
      class="context-menu"
      :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
      @click.stop
    >
      <template v-if="contextMenu.type === 'folder'">
        <div class="ctx-item" @click="newNoteInFolder(contextMenu.folder)">📄 新建笔记</div>
        <div class="ctx-item" @click="newFolderUnder(contextMenu.folder.id)">📁 新建子文件夹</div>
        <div class="ctx-sep"></div>
        <div class="ctx-item danger" @click="deleteFolder(contextMenu.folder, $event)">🗑 删除</div>
      </template>
      <template v-else-if="contextMenu.type === 'note'">
        <div class="ctx-item danger" @click="deleteNote(contextMenu.note, $event)">🗑 删除</div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.tree {
  background: #16162a;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
}
.tree-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 14px;
  color: #888;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  flex-shrink: 0;
  border-bottom: 1px solid #2a2a4a;
}
.tree-header button {
  background: none;
  border: none;
  color: #4a9eff;
  cursor: pointer;
  font-size: 18px;
  padding: 0 4px;
}
.tree-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}
.tree-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  cursor: pointer;
  color: #c8c8c8;
  font-size: 13px;
  user-select: none;
  transition: background 0.1s;
}
.tree-row:hover { background: #1e1e38; }
.tree-row.active { background: #1e2e4a; color: #4a9eff; }
.expand-icon { font-size: 8px; color: #555; width: 12px; flex-shrink: 0; }
.icon { font-size: 13px; flex-shrink: 0; }
.label { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.indent1 { padding-left: 28px; }
.indent2 { padding-left: 44px; }
.tree-children { }
.new-folder-inline {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 12px;
}
.new-folder-inline input {
  flex: 1;
  background: #1a1a2e;
  border: 1px solid #4a9eff;
  border-radius: 4px;
  color: #e8e8e8;
  padding: 4px 8px;
  font-size: 12px;
}
.new-folder-inline input:focus { outline: none; }
.new-folder-inline button {
  background: none;
  border: none;
  color: #888;
  cursor: pointer;
  font-size: 14px;
  padding: 2px 4px;
}
.new-folder-inline button:hover { color: #e8e8e8; }
.indent1.new-folder-inline { padding-left: 28px; }

/* Context menu */
.context-menu {
  position: fixed;
  background: #1e1e38;
  border: 1px solid #3a3a5a;
  border-radius: 8px;
  padding: 4px 0;
  min-width: 160px;
  box-shadow: 0 8px 24px rgba(0,0,0,0.4);
  z-index: 1000;
}
.ctx-item {
  padding: 8px 14px;
  font-size: 13px;
  color: #c8c8c8;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
}
.ctx-item:hover { background: #2a2a4a; }
.ctx-item.danger { color: #ff6b6b; }
.ctx-item.danger:hover { background: #2a1a1a; }
.ctx-sep { height: 1px; background: #2a2a4a; margin: 4px 0; }
</style>

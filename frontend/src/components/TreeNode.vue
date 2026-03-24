<script setup>
import { computed } from 'vue'

defineOptions({ name: 'TreeNode' })

const props = defineProps({
  folder: Object,
  selectedNoteId: String,
  depth: { type: Number, default: 1 },
  expanded: Object,
  creatingIn: String,
  newFolderInput: String,
  folders: Array,
  notes: Array,
})
const emit = defineEmits(['toggle', 'select-note', 'ctx-note', 'ctx-folder', 'new-folder-under', 'create-folder'])

const folderNotes = computed(() =>
  (props.notes || []).filter(n => n.folder_id === props.folder.id)
)

function onCtxFolder(e, folder) { emit('ctx-folder', e, folder) }
function onCtxNote(e, note) { emit('ctx-note', e, note) }
</script>

<template>
  <div>
    <!-- Folder row -->
    <div class="tree-row is-folder" :style="{ paddingLeft: (14 + depth * 14) + 'px' }" @click="$emit('toggle', folder.id)" @contextmenu="$emit('ctx-folder', $event, folder)">
      <svg class="expand-icon" width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
        <path v-if="expanded.has(folder.id)" d="M6 9l6 6 6-6"/>
        <path v-else d="M9 18l6-6-6-6"/>
      </svg>
      <svg class="folder-icon" width="14" height="14" viewBox="0 0 24 24" fill="currentColor">
        <path d="M2 6c0-1.1.9-2 2-2h5l2 2h9c1.1 0 2 .9 2 2v10c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6z"/>
      </svg>
      <span class="label">{{ folder.name }}</span>
    </div>

    <!-- Expanded content -->
    <div v-if="expanded.has(folder.id)">
      <!-- Notes in this folder -->
      <div
        v-for="note in folderNotes"
        :key="note.id"
        class="tree-row"
        :class="{ active: note.id === selectedNoteId }"
        :style="{ paddingLeft: (28 + depth * 14) + 'px' }"
        @click="$emit('select-note', note)"
        @contextmenu="$emit('ctx-note', $event, note)"
      >
        <svg class="file-icon" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
        </svg>
        <span class="label">{{ note.title }}</span>
      </div>

      <!-- New folder input -->
      <div v-if="creatingIn === folder.id" class="new-folder-row" :style="{ paddingLeft: (14 + depth * 14) + 'px' }">
        <input
          :value="newFolderInput"
          placeholder="文件夹名称"
          class="new-folder-input"
          @keyup.enter="$emit('create-folder')"
          @keyup.escape="$emit('create-folder')"
          autofocus
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
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
.expand-icon { flex-shrink: 0; color: #B8AFA0; }
.folder-icon { flex-shrink: 0; color: #B8A898; }
.file-icon { flex-shrink: 0; color: #A8A090; }
.label { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.new-folder-row { padding: 3px 12px 3px 14px; }
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
}
.new-folder-input:focus { border-color: #6B8FCC; }
.new-folder-input::placeholder { color: #B8AFA0; }
</style>

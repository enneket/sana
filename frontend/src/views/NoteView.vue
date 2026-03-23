<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiFetch } from '../main.js'

const route = useRoute()
const router = useRouter()
const note = ref(null)
const title = ref('')
const content = ref('')
const saveTimer = ref(null)
const saved = ref(false)

async function loadNote() {
  const id = route.params.noteId
  if (!id) return
  note.value = await apiFetch(`/notes/${id}`)
  title.value = note.value.title
  content.value = note.value.content
}

onMounted(loadNote)
watch(() => route.params.noteId, loadNote)

function scheduleSave() {
  saved.value = false
  clearTimeout(saveTimer.value)
  saveTimer.value = setTimeout(saveNote, 1000)
}

async function saveNote() {
  if (!note.value) return
  await apiFetch(`/notes/${note.value.id}`, {
    method: 'PUT',
    body: JSON.stringify({ title: title.value, content: content.value })
  })
  saved.value = true
}

onUnmounted(() => clearTimeout(saveTimer.value))

async function deleteNote() {
  if (!note.value) return
  if (!confirm('Delete this note?')) return
  await apiFetch(`/notes/${note.value.id}`, { method: 'DELETE' })
  router.back()
}
</script>

<template>
  <div class="note-view">
    <div class="toolbar">
      <div class="left">
        <button class="back-btn" @click="router.back()">←</button>
        <input
          v-model="title"
          class="title-input"
          placeholder="Note title"
          @input="scheduleSave"
        />
      </div>
      <div class="right">
        <span v-if="saved" class="saved-indicator">Saved</span>
        <button class="delete-btn" @click="deleteNote">Delete</button>
      </div>
    </div>
    <textarea
      v-model="content"
      class="editor"
      placeholder="Start writing..."
      @input="scheduleSave"
    ></textarea>
  </div>
</template>

<style scoped>
.note-view {
  display: flex;
  flex-direction: column;
  height: 100%;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #2a2a4a;
  background: #16162a;
  gap: 12px;
}
.left { display: flex; align-items: center; gap: 10px; flex: 1; }
.right { display: flex; align-items: center; gap: 12px; }
.back-btn {
  background: none;
  border: none;
  color: #888;
  cursor: pointer;
  font-size: 18px;
  padding: 4px;
}
.back-btn:hover { color: #e8e8e8; }
.title-input {
  flex: 1;
  background: none;
  border: none;
  color: #e8e8e8;
  font-size: 16px;
  font-weight: 500;
  padding: 4px 0;
}
.title-input:focus { outline: none; }
.title-input::placeholder { color: #555; }
.saved-indicator { color: #4a9eff; font-size: 12px; }
.delete-btn {
  background: none;
  border: 1px solid #2a2a4a;
  color: #888;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
}
.delete-btn:hover { border-color: #ff6b6b; color: #ff6b6b; }
.editor {
  flex: 1;
  background: #1a1a2e;
  border: none;
  color: #e8e8e8;
  padding: 24px;
  font-size: 15px;
  line-height: 1.7;
  resize: none;
  font-family: inherit;
}
.editor:focus { outline: none; }
.editor::placeholder { color: #444; }
</style>

<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiFetch } from '../main.js'
import { marked } from 'marked'

const props = defineProps({
  note: Object, // inline mode: pass full note object directly
})
const emit = defineEmits(['deleted', 'saved'])

const route = useRoute()
const router = useRouter()
const noteData = ref(null)
const title = ref('')
const content = ref('')
const saveTimer = ref(null)
const saved = ref(false)
const preview = ref(false)

async function loadNote() {
  if (props.note) {
    // Inline mode: note object passed as prop
    noteData.value = await apiFetch(`/notes/${props.note.id}`)
    title.value = noteData.value.title
    content.value = noteData.value.content
  } else {
    // Route mode: load from URL param
    const id = route.params.noteId
    if (!id) return
    noteData.value = await apiFetch(`/notes/${id}`)
    title.value = noteData.value.title
    content.value = noteData.value.content
  }
}

onMounted(loadNote)
watch(() => props.note?.id, loadNote)
watch(() => route.params.noteId, loadNote)

function scheduleSave() {
  saved.value = false
  clearTimeout(saveTimer.value)
  saveTimer.value = setTimeout(saveNote, 1000)
}

async function saveNote() {
  if (!noteData.value) return
  try {
    await apiFetch(`/notes/${noteData.value.id}`, {
      method: 'PUT',
      body: JSON.stringify({ title: title.value, content: content.value })
    })
    saved.value = true
    emit('saved')
  } catch {
    saved.value = false
  }
}

onUnmounted(() => clearTimeout(saveTimer.value))

async function deleteNote() {
  if (!noteData.value) return
  if (!confirm('删除此笔记？')) return
  await apiFetch(`/notes/${noteData.value.id}`, { method: 'DELETE' })
  emit('deleted')
  if (!props.note) {
    router.back()
  }
}
</script>

<template>
  <div class="note-view">
    <div class="toolbar">
      <div class="left">
        <button v-if="!note" class="back-btn" @click="router.back()">←</button>
        <input
          v-model="title"
          class="title-input"
          placeholder="无标题"
          @input="scheduleSave"
        />
      </div>
      <div class="right">
        <button
          class="toggle-btn"
          :class="{ active: preview }"
          @click="preview = !preview"
        >
          {{ preview ? '编辑' : '预览' }}
        </button>
        <span v-if="saved && !preview" class="saved-indicator">已保存</span>
        <button class="delete-btn" @click="deleteNote">删除</button>
      </div>
    </div>
    <textarea
      v-if="!preview"
      v-model="content"
      class="editor"
      placeholder="开始写作..."
      @input="scheduleSave"
    ></textarea>
    <div
      v-else
      class="preview markdown-body"
      v-html="marked(content || '')"
    ></div>
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
.toggle-btn {
  background: none;
  border: 1px solid #2a2a4a;
  color: #888;
  padding: 6px 12px;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
}
.toggle-btn:hover { border-color: #4a9eff; color: #4a9eff; }
.toggle-btn.active { border-color: #4a9eff; color: #4a9eff; background: #1e2e4a; }
.preview {
  flex: 1;
  background: #1a1a2e;
  padding: 24px;
  overflow-y: auto;
  font-size: 15px;
  line-height: 1.7;
  color: #c8c8c8;
}
.markdown-body h1, .markdown-body h2, .markdown-body h3,
.markdown-body h4, .markdown-body h5, .markdown-body h6 {
  color: #e8e8e8;
  margin: 16px 0 8px;
}
.markdown-body h1 { font-size: 1.8em; border-bottom: 1px solid #2a2a4a; padding-bottom: 8px; }
.markdown-body h2 { font-size: 1.4em; }
.markdown-body h3 { font-size: 1.2em; }
.markdown-body p { margin: 8px 0; }
.markdown-body code {
  background: #16162a;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 0.9em;
}
.markdown-body pre {
  background: #16162a;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
}
.markdown-body pre code {
  background: none;
  padding: 0;
}
.markdown-body blockquote {
  border-left: 3px solid #4a9eff;
  margin: 8px 0;
  padding-left: 16px;
  color: #888;
}
.markdown-body a { color: #4a9eff; }
.markdown-body ul, .markdown-body ol { padding-left: 24px; }
.markdown-body li { margin: 4px 0; }
.markdown-body table { border-collapse: collapse; width: 100%; }
.markdown-body th, .markdown-body td {
  border: 1px solid #2a2a4a;
  padding: 8px 12px;
  text-align: left;
}
.markdown-body th { background: #16162a; }
.markdown-body hr { border: none; border-top: 1px solid #2a2a4a; margin: 16px 0; }
.markdown-body img { max-width: 100%; border-radius: 4px; }
</style>

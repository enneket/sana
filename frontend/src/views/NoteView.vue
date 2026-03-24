<script setup>
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiFetch } from '../main.js'
import { marked } from 'marked'

const props = defineProps({
  note: Object,
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
    noteData.value = await apiFetch(`/notes/${props.note.id}`)
    title.value = noteData.value.title
    content.value = noteData.value.content
  } else {
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
        <button v-if="!note" class="icon-btn" @click="router.back()" title="返回">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="15 18 9 12 15 6"/>
          </svg>
        </button>
        <input
          v-model="title"
          class="title-input"
          placeholder="无标题"
          @input="scheduleSave"
        />
      </div>
      <div class="right">
        <span v-if="saved && !preview" class="saved-indicator">已保存</span>
        <button
          class="mode-btn"
          :class="{ active: preview }"
          @click="preview = !preview"
          title="预览"
        >
          <svg v-if="!preview" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
            <circle cx="12" cy="12" r="3"/>
          </svg>
          <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 20h9"/>
            <path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/>
          </svg>
        </button>
        <button class="icon-btn danger" @click="deleteNote" title="删除">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="3 6 5 6 21 6"/>
            <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a1 1 0 0 1 1-1h4a1 1 0 0 1 1 1v2"/>
          </svg>
        </button>
      </div>
    </div>

    <div class="editor-wrap">
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
  </div>
</template>

<style scoped>
.note-view {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #F5F0E8;
  overflow: hidden;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 16px;
  height: 44px;
  flex-shrink: 0;
  border-bottom: 1px solid #DDD8CC;
  gap: 12px;
}
.left { display: flex; align-items: center; gap: 8px; flex: 1; min-width: 0; }
.right { display: flex; align-items: center; gap: 4px; flex-shrink: 0; }

.icon-btn {
  background: none;
  border: none;
  color: #A09888;
  cursor: pointer;
  padding: 5px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.15s, background 0.15s;
  flex-shrink: 0;
}
.icon-btn:hover { color: #3D3830; background: #E8E2D6; }
.icon-btn.danger:hover { color: #C06050; background: #F5EDE8; }

.title-input {
  flex: 1;
  min-width: 0;
  background: none;
  border: none;
  color: #3D3830;
  font-size: 15px;
  font-weight: 500;
  padding: 0;
  outline: none;
  font-family: inherit;
  text-align: center;
}
.title-input::placeholder { color: #B8AFA0; }

.mode-btn {
  background: none;
  border: 1px solid #D4CCBA;
  color: #A09888;
  padding: 4px 10px;
  border-radius: 4px;
  font-size: 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 5px;
  transition: border-color 0.15s, color 0.15s, background 0.15s;
}
.mode-btn:hover { border-color: #9BB0D4; color: #6B8FCC; background: #F0EDE6; }
.mode-btn.active { border-color: #9BB0D4; color: #6B8FCC; background: #EEF0F5; }

.saved-indicator {
  color: #6B8FCC;
  font-size: 12px;
  min-width: 48px;
}

.editor-wrap {
  height: calc(100vh - 84px);
  overflow: hidden;
}
.editor {
  height: 100%;
  background: #F5F0E8;
  border: none;
  color: #3D3830;
  padding: 32px 48px;
  font-size: 14px;
  line-height: 1.8;
  resize: none;
  font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', 'Cascadia Code', monospace;
  outline: none;
  max-width: 720px;
  width: 100%;
  margin: 0 auto;
  display: block;
  box-sizing: border-box;
}
.editor:focus { outline: none; }
.editor::placeholder { color: #C4BA9E; }

.preview {
  height: calc(100vh - 84px);
  background: #F5F0E8;
  padding: 32px 48px;
  overflow-y: auto;
  font-size: 14px;
  line-height: 1.8;
  color: #4A453E;
  max-width: 720px;
  margin: 0 auto;
  box-sizing: border-box;
}
.preview::-webkit-scrollbar { width: 6px; }
.preview::-webkit-scrollbar-track { background: transparent; }
.preview::-webkit-scrollbar-thumb { background: #D4CCBA; border-radius: 3px; }

/* Markdown styles */
.markdown-body h1, .markdown-body h2, .markdown-body h3,
.markdown-body h4, .markdown-body h5, .markdown-body h6 {
  color: #3D3830;
  font-weight: 600;
  margin: 24px 0 10px;
  line-height: 1.3;
}
.markdown-body h1 { font-size: 1.75em; border-bottom: 1px solid #DDD8CC; padding-bottom: 8px; margin-top: 0; }
.markdown-body h2 { font-size: 1.4em; }
.markdown-body h3 { font-size: 1.15em; }
.markdown-body h4 { font-size: 1em; }
.markdown-body p { margin: 8px 0 12px; }
.markdown-body a { color: #6B8FCC; text-decoration: none; }
.markdown-body a:hover { text-decoration: underline; }
.markdown-body code {
  background: #EAE5DC;
  padding: 2px 5px;
  border-radius: 3px;
  font-size: 0.9em;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  color: #4A453E;
}
.markdown-body pre {
  background: #EAE5DC;
  border: 1px solid #DDD8CC;
  border-radius: 6px;
  padding: 14px 16px;
  overflow-x: auto;
  margin: 12px 0;
}
.markdown-body pre code {
  background: none;
  padding: 0;
  font-size: 0.88em;
  color: #4A453E;
}
.markdown-body blockquote {
  border-left: 3px solid #C4BA9E;
  margin: 10px 0;
  padding: 2px 0 2px 16px;
  color: #8A8478;
}
.markdown-body ul, .markdown-body ol { padding-left: 22px; margin: 8px 0; }
.markdown-body li { margin: 4px 0; }
.markdown-body li::marker { color: #B8AFA0; }
.markdown-body table { border-collapse: collapse; width: 100%; margin: 12px 0; }
.markdown-body th, .markdown-body td {
  border: 1px solid #DDD8CC;
  padding: 7px 12px;
  text-align: left;
}
.markdown-body th { background: #EAE5DC; color: #4A453E; font-weight: 500; }
.markdown-body tr:hover td { background: #EEEBE3; }
.markdown-body hr { border: none; border-top: 1px solid #DDD8CC; margin: 24px 0; }
.markdown-body img { max-width: 100%; border-radius: 4px; }
.markdown-body strong { color: #3D3830; font-weight: 600; }
.markdown-body em { color: #6B6458; }
</style>

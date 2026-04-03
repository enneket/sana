<template>
  <div class="editor-overlay" @click.self="$emit('close')">
    <div class="editor-modal">
      <div class="editor-header">
        <span>编辑笔记</span>
        <button class="close-btn" @click="$emit('close')">✕</button>
      </div>
      <textarea
        v-model="editContent"
        class="editor-textarea"
        rows="10"
      ></textarea>
      <div class="editor-footer">
        <button class="cancel-btn" @click="$emit('close')">取消</button>
        <button class="save-btn" @click="save">保存</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps(['memo'])
const emit = defineEmits(['close', 'save'])

const editContent = ref(props.memo.content)

watch(() => props.memo, (m) => {
  editContent.value = m.content
})

function save() {
  const c = editContent.value.trim()
  if (!c) return
  emit('save', { id: props.memo.id, content: c })
}
</script>

<style scoped>
.editor-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0,0,0,0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.editor-modal {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 600px;
  overflow: hidden;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
}

.editor-textarea {
  width: 100%;
  padding: 16px;
  border: none;
  font-family: inherit;
  font-size: 14px;
  resize: vertical;
  min-height: 200px;
}

.editor-textarea:focus {
  outline: none;
}

.editor-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  padding: 12px 16px;
  border-top: 1px solid #eee;
}

.cancel-btn, .save-btn {
  padding: 8px 16px;
  border-radius: 6px;
  font-size: 14px;
  cursor: pointer;
}

.cancel-btn {
  background: #f0f0f0;
  border: none;
}

.save-btn {
  background: #007AFF;
  color: white;
  border: none;
}
</style>

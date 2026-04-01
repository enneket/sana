<template>
  <div class="memo-composer">
    <textarea
      v-model="content"
      class="composer-input"
      placeholder="写下此刻的想法..."
      rows="2"
      @keydown.enter.ctrl="submit"
    ></textarea>
    <button class="composer-btn" @click="submit" :disabled="!content.trim()">
      创建
    </button>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import api from '../api/index.js'

const emit = defineEmits(['created'])
const content = ref('')

async function submit() {
  const c = content.value.trim()
  if (!c) return
  const memo = await api.createMemo(c)
  content.value = ''
  emit('created', memo)
}
</script>

<style scoped>
.memo-composer {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
}

.composer-input {
  flex: 1;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 8px;
  font-size: 14px;
  font-family: inherit;
  resize: none;
}

.composer-input:focus {
  outline: none;
  border-color: #007AFF;
}

.composer-btn {
  padding: 8px 16px;
  background: #007AFF;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
}

.composer-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
}
</style>

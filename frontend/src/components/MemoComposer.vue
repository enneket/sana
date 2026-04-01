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
  gap: 12px;
  margin-bottom: 24px;
  align-items: flex-end;
}

.composer-input {
  flex: 1;
  padding: 12px 14px;
  border: 1px solid #e8e8e8;
  border-radius: 10px;
  font-size: 14px;
  font-family: inherit;
  resize: none;
  background: #fff;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}

.composer-input:focus {
  outline: none;
  border-color: #2ecc71;
}

.composer-btn {
  padding: 10px 20px;
  background: #2ecc71;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: background 0.15s;
}

.composer-btn:hover {
  background: #27ae60;
}

.composer-btn:disabled {
  background: #ccc;
  cursor: not-allowed;
}
</style>

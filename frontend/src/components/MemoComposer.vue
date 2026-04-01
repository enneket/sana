<template>
  <div class="memo-composer">
    <textarea
      v-model="content"
      class="composer-input"
      placeholder="写下此刻的想法..."
      rows="3"
      @keydown.enter.ctrl="submit"
    ></textarea>
    <div class="composer-footer">
      <button class="composer-btn" @click="submit" :disabled="!content.trim()">
        发送
      </button>
    </div>
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
  background: #fff;
  border-radius: 10px;
  border: 1px solid #e8e8e8;
  overflow: hidden;
  margin-bottom: 20px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.04);
}

.composer-input {
  width: 100%;
  padding: 14px 16px;
  border: none;
  font-size: 14px;
  font-family: inherit;
  resize: none;
  box-sizing: border-box;
  line-height: 1.5;
}

.composer-input:focus {
  outline: none;
}

.composer-input::placeholder {
  color: #ccc;
}

.composer-footer {
  display: flex;
  justify-content: flex-end;
  padding: 8px 12px;
  border-top: 1px solid #f0f0f0;
}

.composer-btn {
  padding: 8px 20px;
  background: #2ecc71;
  color: white;
  border: none;
  border-radius: 6px;
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

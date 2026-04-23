<template>
  <div class="memo-composer">
    <textarea
      v-model="content"
      class="composer-input"
      placeholder="写下此刻的想法..."
      rows="4"
      @keydown.enter.ctrl="submit"
    ></textarea>
    <div class="composer-footer">
      <button class="composer-btn" :class="{ active: hasContent }" @click="submit" :disabled="!hasContent">
        发送
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import api from '../api/index.js'

const emit = defineEmits(['created'])
const content = ref('')
const hasContent = computed(() => content.value.trim().length > 0)

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
  border-radius: 12px;
  border: 1px solid #e5e5e5;
  overflow: hidden;
  margin-bottom: 24px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.02);
}

.composer-input {
  width: 100%;
  min-height: 120px;
  padding: 0;
  border: none;
  font-size: 16px;
  font-family: inherit;
  resize: none;
  line-height: 1.6;
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
  margin-top: 16px;
}

.composer-btn {
  padding: 10px 24px;
  background: #f0f0f0;
  color: #ccc;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.15s;
}

.composer-btn:hover:not(:disabled) {
  background: #2ecc71;
  color: white;
}

.composer-btn:disabled {
  background: #f0f0f0;
  color: #ccc;
  cursor: not-allowed;
}

.composer-btn.active {
  background: #2ecc71;
  color: white;
}
</style>

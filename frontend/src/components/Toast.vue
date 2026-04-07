<template>
  <Teleport to="body">
    <div v-if="visible" class="toast-overlay" @click.self="dismiss">
      <div :class="['toast', type]" role="alert">
        <span class="toast-message">{{ message }}</span>
        <button v-if="type === 'error'" class="toast-close" @click="dismiss">✕</button>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref } from 'vue'

const message = ref('')
const type = ref('info') // 'success' | 'error' | 'info'
const duration = ref(3000)
const visible = ref(false)
let timer = null

function show(msg, t = 'info', dur = 3000) {
  message.value = msg
  type.value = t
  duration.value = dur
  visible.value = true
  clearTimeout(timer)
  if (t !== 'error') {
    timer = setTimeout(dismiss, dur)
  }
}

function dismiss() {
  visible.value = false
  clearTimeout(timer)
}

defineExpose({ show, dismiss })
</script>

<style scoped>
.toast-overlay {
  position: fixed;
  bottom: 32px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
  animation: toast-in 200ms ease-out;
}
.toast {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  border-radius: 12px;
  font-size: 14px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  white-space: nowrap;
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
}
.toast.success { background: #D4EDDA; color: #155724; }
.toast.error { background: #F8D7DA; color: #721C24; }
.toast.info { background: #EAE5DC; color: #3D3830; }
.toast-close {
  background: none; border: none; cursor: pointer; font-size: 14px; padding: 0 4px;
  color: inherit; opacity: 0.7;
}
.toast-close:hover { opacity: 1; }
@keyframes toast-in {
  from { opacity: 0; transform: translateX(-50%) translateY(16px); }
  to   { opacity: 1; transform: translateX(-50%) translateY(0); }
}
</style>

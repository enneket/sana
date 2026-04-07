<template>
  <Teleport to="body">
    <div v-if="visible" class="dialog-overlay" @click.self="resolve(false)">
      <div class="dialog-card">
        <div class="dialog-header">{{ title }}</div>
        <div class="dialog-body">{{ message }}</div>
        <div class="dialog-footer">
          <button class="btn-cancel" @click="resolve(false)">取消</button>
          <button :class="['btn-confirm', danger ? 'btn-danger' : 'btn-primary']"
                  @click="resolve(true)">{{ danger ? '删除' : '确定' }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref } from 'vue'

const visible = ref(false)
const title = ref('')
const message = ref('')
const danger = ref(false)
let resolver = null

function show({ title: t, message: m, danger: d = false }) {
  title.value = t
  message.value = m
  danger.value = d
  visible.value = true
  return new Promise((resolve) => {
    resolver = resolve
  })
}

function resolve(value) {
  visible.value = false
  if (resolver) resolver(value)
}

defineExpose({ show })
</script>

<style scoped>
.dialog-overlay {
  position: fixed; inset: 0;
  background: rgba(61,56,48,0.3);
  backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center;
  z-index: 3000;
  animation: fade-in 200ms ease-out;
}
.dialog-card {
  background: #EAE5DC; border: 1px solid #D4CCBA;
  border-radius: 20px;
  width: 90%; max-width: 360px;
  box-shadow: 0 8px 32px rgba(61,56,48,0.12);
  animation: scale-in 200ms ease-out;
  overflow: hidden;
}
.dialog-header {
  padding: 20px 24px 12px;
  font-weight: 600; font-size: 16px; color: #3D3830;
}
.dialog-body {
  padding: 0 24px 16px;
  font-size: 14px; color: #6B6358; line-height: 1.5;
}
.dialog-footer {
  display: flex; justify-content: flex-end; gap: 8px;
  padding: 12px 24px 20px;
}
.btn-cancel, .btn-confirm {
  padding: 8px 20px; border-radius: 8px;
  font-size: 14px; cursor: pointer; border: none;
}
.btn-cancel { background: #E0DBD3; color: #6B6358; }
.btn-primary { background: #6B8FCC; color: #fff; }
.btn-danger { background: #C06050; color: #fff; }
@keyframes fade-in { from { opacity: 0 } to { opacity: 1 } }
@keyframes scale-in {
  from { opacity: 0; transform: scale(0.95); }
  to   { opacity: 1; transform: scale(1); }
}
</style>

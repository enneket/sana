<template>
  <aside class="sidebar">
    <div class="sidebar-actions">
      <button class="action-btn" @click="$emit('openSearch')" title="搜索">🔍</button>
      <button class="action-btn" @click="$emit('export')" title="导出">📤</button>
      <button class="action-btn" @click="$emit('import')" title="导入">📥</button>
    </div>
    <div class="heatmap-section">
      <Heatmap :heatmap="stats.heatmap || {}" />
    </div>
  </aside>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import Heatmap from './Heatmap.vue'
import api from '../api/index.js'

defineEmits(['openSearch', 'export', 'import'])

const stats = ref({})

onMounted(async () => {
  try {
    stats.value = await api.getStats()
  } catch {
    // silent fail
  }
})
</script>

<style scoped>
.sidebar {
  width: 300px;
  min-width: 300px;
  padding: 24px 16px;
  background: #f7f7f7;
  border-right: 1px solid #e5e5e5;
  height: 100vh;
  overflow-y: auto;
}

.sidebar-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.action-btn {
  background: #fff;
  border: 1px solid #eee;
  border-radius: 8px;
  padding: 8px 12px;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.15s;
}

.action-btn:hover {
  background: #f0f0f0;
}

.heatmap-section {
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  border: 1px solid #eee;
}
</style>

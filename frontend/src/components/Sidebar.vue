<template>
  <aside class="sidebar">
    <div class="stats">
      <div class="stat-item">
        <span class="stat-value">{{ stats.memo_count || 0 }}</span>
        <span class="stat-label">笔记</span>
      </div>
      <div class="stat-item">
        <span class="stat-value">{{ stats.active_days || 0 }}</span>
        <span class="stat-label">天</span>
      </div>
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
  width: 240px;
  min-width: 240px;
  padding: 16px;
  background: #fafafa;
  border-right: 1px solid #eee;
  height: 100vh;
  overflow-y: auto;
}

.stats {
  display: flex;
  gap: 16px;
  margin-bottom: 24px;
}

.stat-item {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 24px;
  font-weight: 600;
  color: #3D3830;
}

.stat-label {
  font-size: 12px;
  color: #999;
}

.heatmap-section {
  margin-top: 16px;
}
</style>

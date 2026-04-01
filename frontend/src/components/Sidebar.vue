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
      <div class="heatmap-title">活动热力图</div>
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
  width: 220px;
  min-width: 220px;
  padding: 20px 14px;
  background: #f7f7f7;
  border-right: 1px solid #e8e8e8;
  height: 100vh;
  overflow-y: auto;
}

.stats {
  display: flex;
  gap: 24px;
  margin-bottom: 20px;
}

.stat-item {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 32px;
  font-weight: 400;
  color: #333;
  line-height: 1.1;
}

.stat-label {
  font-size: 12px;
  color: #999;
  margin-top: 2px;
}

.heatmap-section {
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  border: 1px solid #eee;
}

.heatmap-title {
  font-size: 12px;
  color: #999;
  margin-bottom: 8px;
}
</style>

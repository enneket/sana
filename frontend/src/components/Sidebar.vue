<template>
  <aside class="sidebar">
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
  width: 220px;
  min-width: 220px;
  padding: 20px 14px;
  background: #f7f7f7;
  border-right: 1px solid #e8e8e8;
  height: 100vh;
  overflow-y: auto;
}

.heatmap-section {
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  border: 1px solid #eee;
}
</style>

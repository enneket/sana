<template>
  <div class="heatmap">
    <div class="heatmap-months">
      <span v-for="m in months" :key="m" class="month-label">{{ m }}</span>
    </div>
    <div class="heatmap-grid">
      <div
        v-for="(count, date) in heatmapData"
        :key="date"
        class="heat-cell"
        :class="getLevel(count)"
        :title="`${date}: ${count} 条`"
      />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  heatmap: {
    type: Object,
    default: () => ({})
  }
})

const months = computed(() => {
  const now = new Date()
  const result = []
  for (let i = 2; i >= 0; i--) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1)
    result.push(d.toLocaleDateString('zh-CN', { month: 'short' }))
  }
  return result
})

const heatmapData = computed(() => props.heatmap)

function getLevel(count) {
  if (count === 0) return 'level-0'
  if (count === 1) return 'level-1'
  if (count === 2) return 'level-2'
  return 'level-3'
}
</script>

<style scoped>
.heatmap {
  font-size: 10px;
}

.heatmap-months {
  display: flex;
  gap: 4px;
  margin-bottom: 6px;
  color: #999;
}

.month-label {
  flex: 1;
}

.heatmap-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 3px;
}

.heat-cell {
  aspect-ratio: 1;
  border-radius: 2px;
  background: #ebebeb;
}

.level-1 { background: #c3e8d1; }
.level-2 { background: #7cd69e; }
.level-3 { background: #2ecc71; }
</style>

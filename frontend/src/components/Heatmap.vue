<template>
  <div class="heatmap">
    <div class="heatmap-grid">
      <div
        v-for="(count, index) in grid"
        :key="index"
        class="heat-cell"
        :class="getLevel(count)"
        :title="getTitle(index, count)"
      />
    </div>
    <div class="heatmap-months">
      <span v-for="m in months" :key="m" class="month-label">{{ m }}</span>
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

// 7 columns (days) x 12 rows (weeks), matching CSS grid row-major order
const grid = computed(() => {
  const result = []
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())

  // Calculate start: go back 12 weeks from today
  const start = new Date(today)
  start.setDate(today.getDate() - 11 * 7 - today.getDay())

  // Generate: row (weeks) outer, col (days) inner -> matches CSS grid row-major
  for (let week = 0; week < 12; week++) {
    for (let day = 0; day < 7; day++) {
      const current = new Date(start)
      current.setDate(start.getDate() + week * 7 + day)
      const dateStr = current.toISOString().split('T')[0]
      result.push(props.heatmap[dateStr] || 0)
    }
  }
  return result
})

// Months at bottom
const months = computed(() => {
  const now = new Date()
  const result = []
  for (let i = 2; i >= 0; i--) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1)
    result.push(d.toLocaleDateString('zh-CN', { month: 'short' }))
  }
  return result
})

function getLevel(count) {
  if (count === 0) return 'level-0'
  if (count === 1) return 'level-1'
  if (count === 2) return 'level-2'
  return 'level-3'
}

function getTitle(index, count) {
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const start = new Date(today)
  start.setDate(today.getDate() - 11 * 7 - today.getDay())

  // grid is 7 columns (days) x 12 rows (weeks), row-major order
  const col = index % 7
  const row = Math.floor(index / 7)
  const current = new Date(start)
  current.setDate(start.getDate() + row * 7 + col)
  return `${current.toLocaleDateString('zh-CN')} ${count} 条`
}
</script>

<style scoped>
.heatmap {
  font-size: 10px;
}

.heatmap-grid {
  display: grid;
  grid-template-columns: repeat(7, 14px);
  grid-template-rows: repeat(12, 14px);
  gap: 3px;
}

.heat-cell {
  border-radius: 2px;
  background: #ebebeb;
}

.level-0 { background: #ebebeb; }
.level-1 { background: #c3e8d1; }
.level-2 { background: #7cd69e; }
.level-3 { background: #2ecc71; }

.heatmap-months {
  display: flex;
  gap: 4px;
  margin-top: 6px;
  justify-content: space-between;
}

.month-label {
  color: #999;
  font-size: 10px;
}
</style>

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

// 7 columns (days) x 12 rows (weeks)
const grid = computed(() => {
  const result = []
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())

  // Calculate start: go back 12 weeks from today
  const start = new Date(today)
  start.setDate(today.getDate() - 11 * 7 - today.getDay())

  // Generate 7 columns x 12 rows
  for (let row = 0; row < 7; row++) {
    for (let col = 0; col < 12; col++) {
      const current = new Date(start)
      current.setDate(start.getDate() + col * 7 + row)
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

  const row = index % 7
  const col = Math.floor(index / 7)
  const current = new Date(start)
  current.setDate(start.getDate() + col * 7 + row)
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

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

// GitHub style: 7 rows (days) x 12 columns (weeks)
const grid = computed(() => {
  const result = []
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())

  // Calculate start: go back 12 weeks from today
  const start = new Date(today)
  start.setDate(today.getDate() - 11 * 7 - today.getDay())

  // Generate 12 columns x 7 rows
  for (let col = 0; col < 12; col++) {
    for (let row = 0; row < 7; row++) {
      const current = new Date(start)
      current.setDate(start.getDate() + col * 7 + row)
      const dateStr = current.toISOString().split('T')[0]
      result.push(props.heatmap[dateStr] || 0)
    }
  }
  return result
})

const months = computed(() => {
  const now = new Date()
  const result = []
  for (let i = 11; i >= 0; i--) {
    const d = new Date(now.getFullYear(), now.getMonth(), 1)
    d.setMonth(d.getMonth() - Math.floor(i * 12 / 12))
    const idx = Math.floor(i * 12 / 12)
    if (result.length === 0 || result[result.length - 1].idx !== idx) {
      result.push({ label: d.toLocaleDateString('zh-CN', { month: 'short' }), idx })
    }
  }
  return result.map(r => r.label).slice(-3)
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

  const col = Math.floor(index / 7)
  const row = index % 7
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
  grid-template-columns: repeat(13, 1fr);
  gap: 4px;
}

.heat-cell {
  aspect-ratio: 1;
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
  color: #999;
}

.month-label {
  flex: 1;
}
</style>

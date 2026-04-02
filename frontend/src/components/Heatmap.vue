<template>
  <div class="heatmap">
    <!-- 左侧星期标签 -->
    <div class="heatmap-left">
      <div class="day-label"></div>
      <div v-for="d in dayLabels" :key="d" class="day-label">{{ d }}</div>
    </div>
    <!-- 网格和月份 -->
    <div class="heatmap-main">
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
        <span
          v-for="(m, i) in months"
          :key="i"
          class="month-label"
          :style="{ gridColumn: m.col + 1 }"
        >{{ m.label }}</span>
      </div>
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

const dayLabels = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']

// 7 columns (days) x 12 rows (weeks)
// Column-major order: each column is one week, dates increase top to bottom
const grid = computed(() => {
  const result = []
  const today = new Date()
  today.setHours(0, 0, 0, 0)

  // Start from 11 weeks ago, aligned to Sunday
  const start = new Date(today)
  start.setDate(today.getDate() - 11 * 7 - today.getDay())

  // column-major: day outer, week inner -> 7 cols x 12 rows
  for (let day = 0; day < 7; day++) {
    for (let week = 0; week < 12; week++) {
      const current = new Date(start)
      current.setDate(start.getDate() + week * 7 + day)
      const dateStr = current.toISOString().split('T')[0]
      result.push(props.heatmap[dateStr] || 0)
    }
  }
  return result
})

// 月份标签
const months = computed(() => {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const start = new Date(today)
  start.setDate(today.getDate() - 11 * 7 - today.getDay())

  const result = []
  let lastMonth = ''

  for (let week = 0; week < 12; week++) {
    const current = new Date(start)
    current.setDate(start.getDate() + week * 7)
    const month = current.toLocaleDateString('zh-CN', { month: 'short' })

    if (month !== lastMonth) {
      result.push({ label: month, col: week })
      lastMonth = month
    }
  }
  return result
})

function getLevel(count) {
  if (count === 0) return 'level-0'
  if (count === 1) return 'level-1'
  if (count === 2) return 'level-2'
  return 'level-3'
}

// grid is 7 columns (days) x 12 rows (weeks), column-major
function getTitle(index, count) {
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  const start = new Date(today)
  start.setDate(today.getDate() - 11 * 7 - today.getDay())

  // column-major: col = day index, row = week index
  const col = Math.floor(index / 12)
  const row = index % 12
  const current = new Date(start)
  current.setDate(start.getDate() + row * 7 + col)
  return `${current.toLocaleDateString('zh-CN')} ${count} 条`
}
</script>

<style scoped>
.heatmap {
  display: flex;
  gap: 4px;
  font-size: 10px;
}

.heatmap-left {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.day-label {
  height: 14px;
  line-height: 14px;
  color: #999;
  font-size: 10px;
  width: 24px;
  text-align: right;
}

.heatmap-main {
  display: flex;
  flex-direction: column;
  gap: 4px;
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
  display: grid;
  grid-template-columns: repeat(7, 14px);
  gap: 3px;
}

.month-label {
  color: #999;
  font-size: 10px;
  text-align: left;
}
</style>

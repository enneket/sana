<template>
  <div class="heatmap">
    <!-- 左侧：空 + 星期 -->
    <div class="day-col">
      <div class="header-corner"></div>
      <div v-for="(label, i) in dayLabels" :key="i" class="day-label">{{ label }}</div>
    </div>

    <!-- 右侧：月份 + 网格 -->
    <div class="grid-area">
      <!-- 月份标签 -->
      <div class="months-row">
        <span
          v-for="(m, i) in monthPositions"
          :key="i"
          class="month-label"
          :style="{ left: m.x + 'px' }"
        >{{ m.label }}</span>
      </div>

      <!-- 网格 -->
      <div class="grid">
        <div
          v-for="(cell, idx) in cells"
          :key="idx"
          class="cell"
          :class="getLevel(cell.count)"
          :title="cell.dateStr + ' ' + cell.count + '条'"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  heatmap: { type: Object, default: () => ({}) }
})

const dayLabels = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
const CELL = 14
const GAP = 3

// 计算起始日期（12周前的周日）
const startDate = computed(() => {
  const d = new Date()
  d.setHours(0, 0, 0, 0)
  d.setDate(d.getDate() - 11 * 7 - d.getDay())
  return d
})

// 12列(周) x 7行(天)，row-major CSS grid
const cells = computed(() => {
  const result = []
  const s = startDate.value
  for (let week = 0; week < 12; week++) {
    for (let day = 0; day < 7; day++) {
      const d = new Date(s)
      d.setDate(s.getDate() + week * 7 + day)
      const dateStr = d.toISOString().split('T')[0]
      result.push({ count: props.heatmap[dateStr] || 0, dateStr })
    }
  }
  return result
})

// 月份标签位置
const monthPositions = computed(() => {
  const s = startDate.value
  const result = []
  let lastMonth = ''

  for (let week = 0; week < 12; week++) {
    const d = new Date(s)
    d.setDate(s.getDate() + week * 7)
    const month = d.toLocaleDateString('zh-CN', { month: 'short' })
    if (month !== lastMonth) {
      result.push({ label: month, x: week * (CELL + GAP) })
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
</script>

<style scoped>
.heatmap {
  display: flex;
  gap: 4px;
  font-size: 10px;
  position: relative;
}

.day-col {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.header-corner {
  height: 14px;
}

.day-label {
  height: 14px;
  line-height: 14px;
  color: #999;
  width: 24px;
  text-align: right;
}

.grid-area {
  position: relative;
}

.months-row {
  position: relative;
  height: 14px;
  margin-bottom: 4px;
}

.month-label {
  position: absolute;
  color: #999;
  font-size: 10px;
  white-space: nowrap;
}

.grid {
  display: grid;
  grid-template-columns: repeat(12, 14px);
  grid-template-rows: repeat(7, 14px);
  gap: 3px;
}

.cell {
  width: 14px;
  height: 14px;
  border-radius: 2px;
  background: #ebebeb;
}

.level-0 { background: #ebebeb; }
.level-1 { background: #c3e8d1; }
.level-2 { background: #7cd69e; }
.level-3 { background: #2ecc71; }
</style>

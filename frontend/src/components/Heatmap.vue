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
      <div class="months-row" :style="{ width: totalWeeks * (CELL + GAP) - GAP + 'px' }">
        <span
          v-for="(m, i) in monthPositions"
          :key="i"
          class="month-label"
          :style="{ left: m.x + 'px' }"
        >{{ m.label }}</span>
      </div>

      <!-- 网格 -->
      <div
        class="grid"
        :style="{
          gridTemplateColumns: `repeat(${totalWeeks}, 14px)`,
          gridAutoFlow: 'column'
        }"
      >
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

// 今天的 UTC 日期
const today = computed(() => {
  const d = new Date()
  d.setUTCHours(0, 0, 0, 0)
  return d
})

// 计算起始日期（今天之前的最近周日）
const startDate = computed(() => {
  const d = new Date(today.value)
  d.setUTCDate(d.getUTCDate() - d.getUTCDay())
  return d
})

// 总周数（从起始周到今天）
const totalWeeks = computed(() => {
  const days = Math.ceil((today.value.getTime() - startDate.value.getTime()) / (7 * 24 * 60 * 60 * 1000))
  return days
})

// 生成单元格数据
const cells = computed(() => {
  const result = []
  const s = startDate.value
  for (let week = 0; week < totalWeeks.value; week++) {
    for (let day = 0; day < 7; day++) {
      const d = new Date(s)
      d.setUTCDate(s.getUTCDate() + week * 7 + day)
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

  for (let week = 0; week < totalWeeks.value; week++) {
    const d = new Date(s)
    d.setUTCDate(s.getUTCDate() + week * 7)
    const month = d.toLocaleDateString('zh-CN', { month: 'short', timeZone: 'UTC' })
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

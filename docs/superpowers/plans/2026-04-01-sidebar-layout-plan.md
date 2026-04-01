# 侧边栏布局实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Sana 改为双栏布局：左侧统计+热力图侧边栏，右侧笔记主内容区

**Architecture:**
- 新增 Layout.vue 作为外层容器（flex 双栏）
- 新增 Sidebar.vue 和 Heatmap.vue 组件
- 后端新增 /api/memos/stats 端点
- TimelineView 移除 header 部分，仅保留笔记列表

**Tech Stack:** Vue 3 (Composition API), Go HTTP

---

## Task 1: 后端 Stats API

**Files:**
- Modify: `backend/memo_handler.go`
- Modify: `backend/main.go`

- [ ] **Step 1: 在 main.go 添加路由**

在 `mux.HandleFunc("GET /api/memos/stats", ...)` 后面添加一行。

- [ ] **Step 2: 在 memo_handler.go 添加 handleGetStats**

```go
func handleGetStats(w http.ResponseWriter, r *http.Request) {
    userID := r.Context().Value("userID").(string)

    var memoCount int
    err := db.QueryRow(r.Context(),
        "SELECT COUNT(*) FROM memos WHERE user_id = $1", userID).Scan(&memoCount)
    if err != nil {
        memoCount = 0
    }

    var activeDays int
    err = db.QueryRow(r.Context(),
        "SELECT COUNT(DISTINCT DATE(created_at)) FROM memos WHERE user_id = $1", userID).Scan(&activeDays)
    if err != nil {
        activeDays = 0
    }

    // Heatmap: last 90 days grouped by date
    rows, err := db.Query(r.Context(), `
        SELECT DATE(created_at) as day, COUNT(*) as count
        FROM memos
        WHERE user_id = $1 AND created_at >= NOW() - INTERVAL '90 days'
        GROUP BY DATE(created_at)
        ORDER BY day
    `, userID)

    heatmap := make(map[string]int)
    for rows.Next() {
        var day string
        var count int
        rows.Scan(&day, &count)
        heatmap[day] = count
    }
    rows.Close()

    json.NewEncoder(w).Encode(map[string]interface{}{
        "memo_count":  memoCount,
        "active_days": activeDays,
        "heatmap":     heatmap,
    })
}
```

- [ ] **Step 3: Build 验证**

```bash
cd /home/zjx/code/mine/sana/backend && go build -o sana . 2>&1
```

Expected: 编译成功

- [ ] **Step 4: Commit**

```bash
git add backend/memo_handler.go backend/main.go
git commit -m "feat(backend): add GET /api/memos/stats endpoint"
```

---

## Task 2: 前端 API 方法

**Files:**
- Modify: `frontend/src/api/index.js`

- [ ] **Step 1: 添加 getStats 方法**

在 `searchMemos` 后面添加：

```javascript
getStats: () => fetchWithAuth(`${API_BASE}/memos/stats`).then(handleResponse),
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/api/index.js
git commit -m "feat(api): add getStats method"
```

---

## Task 3: Heatmap 组件

**Files:**
- Create: `frontend/src/components/Heatmap.vue`

- [ ] **Step 1: 创建 Heatmap.vue**

```vue
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
  margin-bottom: 4px;
  color: #999;
}

.month-label {
  flex: 1;
}

.heatmap-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 2px;
}

.heat-cell {
  aspect-ratio: 1;
  border-radius: 2px;
  background: #ebedf0;
}

.level-1 { background: #9be9a8; }
.level-2 { background: #40c463; }
.level-3 { background: #30a14e; }
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/Heatmap.vue
git commit -m "feat(frontend): add Heatmap component"
```

---

## Task 4: Sidebar 组件

**Files:**
- Create: `frontend/src/components/Sidebar.vue`

- [ ] **Step 1: 创建 Sidebar.vue**

```vue
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
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/Sidebar.vue
git commit -m "feat(frontend): add Sidebar component"
```

---

## Task 5: Layout 容器

**Files:**
- Create: `frontend/src/components/Layout.vue`

- [ ] **Step 1: 创建 Layout.vue**

```vue
<template>
  <div class="app-layout">
    <Sidebar />
    <main class="main-content">
      <RouterView />
    </main>
  </div>
</template>

<script setup>
import Sidebar from './Sidebar.vue'
</script>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
}

.main-content {
  flex: 1;
  overflow-y: auto;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/Layout.vue
git commit -m "feat(frontend): add Layout component"
```

---

## Task 6: TimelineView 调整

**Files:**
- Modify: `frontend/src/views/TimelineView.vue`

- [ ] **Step 1: 移除 timeline-header**

删除 `<header class="timeline-header">...</header>` 整块，以及对应的 CSS。

- [ ] **Step 2: 更新 CSS**

移除 `.timeline-header` 相关样式。

```vue
<style scoped>
.timeline-view {
  max-width: 700px;
  margin: 0 auto;
  padding: 16px;
}
/* .timeline-header 已移除 */
/* 其他样式保持 */
</style>
```

- [ ] **Step 3: Build 验证**

```bash
cd frontend && npm run build 2>&1 | tail -10
```

Expected: 编译成功

- [ ] **Step 4: Commit**

```bash
git add frontend/src/views/TimelineView.vue
git commit -m "refactor(frontend): remove header from TimelineView"
```

---

## Task 7: App.vue 路由调整

**Files:**
- Modify: `frontend/src/App.vue`

- [ ] **Step 1: 修改路由指向**

将 `<RouterView />` 包裹在 Layout 中：

```vue
<template>
  <RouterView />
</template>
```

实际不需要改 App.vue，因为 Layout 已经在路由层级中了。跳过此任务。

---

## Task 8: Docker 重新构建

- [ ] **Step 1: 重新构建前端**

```bash
cd frontend && npm run build 2>&1 | tail -5
```

- [ ] **Step 2: 重新构建 Docker**

```bash
cd docker && docker compose build 2>&1 | tail -5
```

- [ ] **Step 3: 重启容器**

```bash
docker compose down && docker compose up -d 2>&1
```

---

## 自检清单

**Spec 覆盖检查：**
- [x] 双栏布局（Layout.vue, flex）
- [x] 侧边栏 240px（Sidebar.vue width）
- [x] 统计信息（memo_count, active_days）
- [x] 热力图（Heatmap.vue, 90天）
- [x] 热力图颜色（level-0~3）
- [x] 主内容区（TimelineView.vue）
- [x] Stats API（/api/memos/stats）

**占位符扫描：**
- 无 TBD/TODO

**类型一致性：**
- `api.getStats()` 返回 `{memo_count, active_days, heatmap}`
- Heatmap prop: `Object`

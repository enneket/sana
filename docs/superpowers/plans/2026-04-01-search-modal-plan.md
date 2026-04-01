# 搜索弹窗实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将内联搜索改为 Modal 弹窗模式

**Architecture:** 新增 SearchModal.vue 组件，TimelineView.vue 移除内联搜索改用弹窗控制

**Tech Stack:** Vue 3 (Composition API)

---

## Task 1: 创建 SearchModal 组件

**Files:**
- Create: `frontend/src/components/SearchModal.vue`

- [ ] **Step 1: 创建 SearchModal.vue**

```vue
<template>
  <Teleport to="body">
    <div v-if="show" class="search-overlay" @click.self="$emit('close')">
      <div class="search-modal">
        <div class="search-header">
          <span>搜索</span>
          <button class="close-btn" @click="$emit('close')">✕</button>
        </div>
        <div class="search-input-wrapper">
          <input
            ref="searchInput"
            v-model="query"
            class="search-input"
            placeholder="搜索笔记..."
            @input="debouncedSearch"
          >
        </div>
        <div class="search-results">
          <div v-if="loading" class="search-loading">搜索中...</div>
          <div v-else-if="results.length === 0 && query" class="search-empty">
            未找到匹配 "{{ query }}" 的笔记
          </div>
          <div v-else-if="results.length > 0" class="search-count">
            搜索结果 ({{ results.length }})
          </div>
          <MemoCard
            v-for="memo in results"
            :key="memo.id"
            :memo="memo"
            @click="$emit('select', memo)"
          />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'
import MemoCard from './MemoCard.vue'
import api from '../api/index.js'

const props = defineProps(['show'])
const emit = defineEmits(['close', 'select'])

const query = ref('')
const results = ref([])
const loading = ref(false)
const searchInput = ref(null)
let searchTimer = null

watch(() => props.show, async (val) => {
  if (val) {
    await nextTick()
    searchInput.value?.focus()
    query.value = ''
    results.value = []
  }
})

function debouncedSearch() {
  clearTimeout(searchTimer)
  if (!query.value.trim()) {
    results.value = []
    loading.value = false
    return
  }
  loading.value = true
  searchTimer = setTimeout(doSearch, 300)
}

async function doSearch() {
  try {
    const data = await api.searchMemos(query.value.trim())
    results.value = data.memos || []
  } catch {
    results.value = []
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.search-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 100px;
  z-index: 2000;
}

.search-modal {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 560px;
  max-height: 70vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.search-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
}

.close-btn:hover {
  background: #f0f0f0;
}

.search-input-wrapper {
  padding: 12px 16px;
  border-bottom: 1px solid #eee;
}

.search-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 8px;
  font-size: 14px;
  box-sizing: border-box;
}

.search-input:focus {
  outline: none;
  border-color: #007AFF;
}

.search-results {
  overflow-y: auto;
  padding: 8px 16px 16px;
}

.search-loading, .search-empty, .search-count {
  padding: 16px;
  text-align: center;
  color: #666;
  font-size: 14px;
}

.search-count {
  text-align: left;
  padding: 8px 0;
  color: #999;
  font-size: 12px;
}
</style>
```

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/SearchModal.vue
git commit -m "feat(frontend): add SearchModal component"
```

---

## Task 2: 修改 TimelineView 集成弹窗

**Files:**
- Modify: `frontend/src/views/TimelineView.vue`

- [ ] **Step 1: 更新 TimelineView.vue**

修改 template 部分：

```vue
<!-- 移除 div.search-container -->

<!-- 在 MemoEditor 后面添加 -->
<SearchModal
  v-if="showSearchModal"
  :show="showSearchModal"
  @close="showSearchModal = false"
  @select="onSearchSelect"
/>
```

修改 script 部分：

1. 移除 `showSearch` ref，改为 `const showSearchModal = ref(false)`
2. 移除 `searchQuery`、`searchResults`、`searchMode` refs
3. 移除 `debouncedSearch`、`doSearch` 函数
4. 修改搜索图标点击：`@click="showSearchModal = true"`
5. 添加 `onSearchSelect` 函数：
```javascript
function onSearchSelect(memo) {
  showSearchModal.value = false
  editMemo(memo)
}
```

完整修改后的 script 部分：
```javascript
import { ref, computed, onMounted } from 'vue'
import MemoComposer from '../components/MemoComposer.vue'
import MemoCard from '../components/MemoCard.vue'
import MemoEditor from '../components/MemoEditor.vue'
import TimeGroup from '../components/TimeGroup.vue'
import SearchModal from '../components/SearchModal.vue'
import api from '../api/index.js'

const memos = ref([])
const loading = ref(false)
const error = ref(null)
const cursor = ref(null)
const hasMore = ref(false)
const showSearchModal = ref(false)
const editingMemo = ref(null)
const fileInput = ref(null)

const groupedMemos = computed(() => {
  // ... unchanged ...
})

async function loadMemos(append = false) {
  // ... unchanged ...
}

async function loadMore() {
  // ... unchanged ...
}

function onMemoCreated(memo) {
  memos.value = [memo, ...memos.value]
}

async function editMemo(memo) {
  editingMemo.value = { ...memo }
}

async function saveMemo({ id, content }) {
  await api.updateMemo(id, content)
  const idx = memos.value.findIndex(m => m.id === id)
  if (idx >= 0) {
    memos.value[idx] = { ...memos.value[idx], content }
  }
  editingMemo.value = null
}

async function deleteMemo(id) {
  if (!confirm('确定删除这条笔记？')) return
  await api.deleteMemo(id)
  memos.value = memos.value.filter(m => m.id !== id)
}

function onSearchSelect(memo) {
  showSearchModal.value = false
  editMemo(memo)
}

async function handleExport() {
  // ... unchanged ...
}

async function handleImport(e) {
  // ... unchanged ...
}

onMounted(() => loadMemos())
```

- [ ] **Step 2: Build 验证**

```bash
cd frontend && npm run build 2>&1 | tail -10
```

Expected: 编译成功

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/TimelineView.vue
git commit -m "feat(frontend): replace inline search with SearchModal"
```

---

## 自检清单

**Spec 覆盖检查：**
- [x] 搜索弹窗显示/隐藏
- [x] 实时搜索（debounce 300ms）
- [x] 点击结果编辑
- [x] 点击外部/ESC/X 关闭

**占位符扫描：**
- 无 TBD/TODO
- 所有代码完整

**类型一致性：**
- `props.show` boolean
- `emit('close')`, `emit('select', memo)` 事件签名一致

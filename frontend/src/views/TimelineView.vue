<template>
  <div class="timeline-view">
    <header class="timeline-header">
      <h1>Sana</h1>
      <div class="header-actions">
        <button class="icon-btn" @click="showSearchModal = true" title="搜索">
          🔍
        </button>
        <button class="icon-btn" @click="handleExport" title="导出">📤</button>
        <button class="icon-btn" @click="fileInput.click()" title="导入">📥</button>
        <input ref="fileInput" type="file" accept=".zip" style="display:none" @change="handleImport">
      </div>
    </header>

    <MemoComposer @created="onMemoCreated" />

    <div v-if="loading" class="loading">加载中...</div>

    <div v-else-if="error" class="error">{{ error }}</div>

    <div v-else class="memo-list">
      <div v-if="groupedMemos.length === 0" class="empty">
        还没有笔记，写下第一条吧 ✨
      </div>
      <TimeGroup
        v-for="group in groupedMemos"
        :key="group.label"
        :label="group.label"
        :memos="group.memos"
        @edit="editMemo"
        @delete="deleteMemo"
      />
      <button v-if="hasMore" class="load-more" @click="loadMore">
        加载更多
      </button>
    </div>

    <SearchModal
      v-if="showSearchModal"
      :show="showSearchModal"
      @close="showSearchModal = false"
      @select="onSearchSelect"
    />

    <MemoEditor
      v-if="editingMemo"
      :memo="editingMemo"
      @close="editingMemo = null"
      @save="saveMemo"
    />
  </div>
</template>

<script setup>
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
  const now = new Date()
  const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const yesterday = new Date(today - 86400000)
  const thisWeek = new Date(today - 7 * 86400000)

  const groups = { today: [], yesterday: [], thisWeek: [], older: [] }

  for (const m of memos.value) {
    const d = new Date(m.updated_ts * 1000)
    if (d >= today) {
      groups.today.push(m)
    } else if (d >= yesterday) {
      groups.yesterday.push(m)
    } else if (d >= thisWeek) {
      groups.thisWeek.push(m)
    } else {
      groups.older.push(m)
    }
  }

  const result = []
  if (groups.today.length) result.push({ label: '今天', memos: groups.today })
  if (groups.yesterday.length) result.push({ label: '昨天', memos: groups.yesterday })
  if (groups.thisWeek.length) result.push({ label: '本周', memos: groups.thisWeek })
  if (groups.older.length) result.push({ label: '更早', memos: groups.older })
  return result
})

async function loadMemos(append = false) {
  loading.value = true
  error.value = null
  try {
    const data = await api.listMemos(append ? cursor.value : null)
    if (append) {
      memos.value = [...memos.value, ...(data.memos || [])]
    } else {
      memos.value = data.memos || []
    }
    cursor.value = data.next_cursor
    hasMore.value = data.has_more
  } catch (e) {
    error.value = '加载失败'
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (cursor.value) await loadMemos(true)
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
  try {
    const blob = await api.exportMemos()
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `sana_export_${Date.now()}.zip`
    a.click()
    URL.revokeObjectURL(url)
  } catch (e) {
    alert('导出失败')
  }
}

async function handleImport(e) {
  const file = e.target.files[0]
  if (!file) return
  const formData = new FormData()
  formData.append('file', file)
  try {
    const result = await api.importMemos(formData)
    alert(`导入完成：${result.memos_imported} 条笔记`)
    await loadMemos()
  } catch (e) {
    alert('导入失败')
  }
  e.target.value = ''
}

onMounted(() => loadMemos())
</script>

<style scoped>
.timeline-view {
  max-width: 700px;
  margin: 0 auto;
  padding: 16px;
}

.timeline-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.timeline-header h1 {
  font-size: 24px;
  font-weight: 600;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.icon-btn {
  background: none;
  border: none;
  font-size: 18px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
}

.icon-btn:hover {
  background: #f0f0f0;
}

.loading, .error, .empty {
  text-align: center;
  padding: 32px;
  color: #666;
}

.error {
  color: #d00;
}

.load-more {
  display: block;
  margin: 16px auto;
  padding: 8px 24px;
  background: #f0f0f0;
  border: none;
  border-radius: 8px;
  cursor: pointer;
}

.load-more:hover {
  background: #e0e0e0;
}
</style>

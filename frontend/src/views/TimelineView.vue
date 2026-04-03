<template>
  <div class="timeline-view">
    <SanaComposer @created="onMemoCreated" />

    <div v-if="loading && memos.length === 0" class="loading">加载中...</div>

    <div v-else-if="error && memos.length === 0" class="error">{{ error }}</div>

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
      <div ref="sentinel" class="sentinel">
        <div v-if="loading && memos.length > 0" class="loading-more">加载中...</div>
        <div v-else-if="!hasMore && memos.length > 0" class="no-more">没有更多了</div>
      </div>
    </div>

    <SearchModal
      v-if="showSearchModal"
      :show="showSearchModal"
      @close="showSearchModal = false"
      @select="onSearchSelect"
    />

    <SanaEditor
      v-if="editingMemo"
      :memo="editingMemo"
      @close="editingMemo = null"
      @save="saveMemo"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import SanaComposer from '../components/SanaComposer.vue'
import SanaCard from '../components/SanaCard.vue'
import SanaEditor from '../components/SanaEditor.vue'
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
const sentinel = ref(null)
const emit = defineEmits(['created'])

let observer = null

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
  if (cursor.value && !loading.value) await loadMemos(true)
}

function setupObserver() {
  if (!sentinel.value) return
  observer = new IntersectionObserver((entries) => {
    if (entries[0].isIntersecting && hasMore.value && !loading.value) {
      loadMore()
    }
  }, { threshold: 0.1 })
  observer.observe(sentinel.value)
}

function onMemoCreated(memo) {
  memos.value = [memo, ...memos.value]
  emit('created', memo)
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

onMounted(async () => {
  await loadMemos()
  setupObserver()
})

onUnmounted(() => {
  if (observer) observer.disconnect()
})

function triggerImport() {
  fileInput.value?.click()
}

function openSearch() {
  showSearchModal.value = true
}

defineExpose({
  openSearch,
  handleExport,
  triggerImport
})
</script>

<style scoped>
.timeline-view {
  padding-top: 24px;
}

.loading, .error, .empty {
  text-align: center;
  padding: 40px 0;
  color: #bbb;
  font-size: 14px;
}

.error {
  color: #e74c3c;
}

.load-more {
  display: block;
  margin: 20px auto;
  padding: 10px 24px;
  background: #fff;
  border: 1px solid #eee;
  border-radius: 8px;
  cursor: pointer;
  font-size: 14px;
  color: #999;
  transition: all 0.15s;
}

.load-more:hover {
  background: #f7f7f7;
  border-color: #ddd;
  color: #666;
}

.sentinel {
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.loading-more, .no-more {
  font-size: 12px;
  color: #bbb;
}
</style>

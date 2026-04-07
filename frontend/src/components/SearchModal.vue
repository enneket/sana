<template>
  <Teleport to="body">
    <div v-if="show" class="search-overlay" @click.self="$emit('close')" @keydown.esc="$emit('close')" tabindex="-1">
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
          <SanaCard
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
import SanaCard from './SanaCard.vue'
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
    results.value = data.sanas || []
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
  background: rgba(61, 56, 48, 0.3);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 100px;
  z-index: 2000;
  border-radius: 20px;
  animation: overlay-in 200ms ease-out;
}

@keyframes overlay-in {
  from { opacity: 0; }
  to { opacity: 1; }
}

@keyframes modal-scale-in {
  from { opacity: 0; transform: scale(0.95); }
  to { opacity: 1; transform: scale(1); }
}

.search-modal {
  background: white;
  border-radius: 20px;
  width: 90%;
  max-width: 560px;
  max-height: 70vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 8px 32px rgba(61,56,48,0.12);
  animation: modal-scale-in 200ms ease-out;
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

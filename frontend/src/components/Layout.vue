<template>
  <div class="app-layout">
    <Sidebar ref="sidebar" @openSearch="showSearch" @export="handleExport" @import="triggerImport" />
    <main class="main-content">
      <TimelineView ref="timeline" @created="onMemoCreated" />
    </main>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import Sidebar from './Sidebar.vue'
import TimelineView from '../views/TimelineView.vue'

const timeline = ref(null)
const sidebar = ref(null)
const showSearchModal = ref(false)

function showSearch() {
  timeline.value?.openSearch()
}

function handleExport() {
  timeline.value?.handleExport()
}

function triggerImport() {
  timeline.value?.triggerImport()
}

function onMemoCreated() {
  sidebar.value?.refreshStats()
}
</script>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  background: #f7f7f7;
}

.main-content {
  flex: 1;
  overflow-y: auto;
  padding: 0 40px;
}
</style>

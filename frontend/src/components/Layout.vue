<template>
  <div class="app-layout">
    <Sidebar />
    <main class="main-content">
      <TimelineView />
    </main>
  </div>
</template>

<script setup>
import Sidebar from './Sidebar.vue'
import TimelineView from '../views/TimelineView.vue'
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
  background: #f7f7f7;
}
</style>

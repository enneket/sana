<template>
  <div class="time-group">
    <div class="group-label">{{ label }}</div>
    <SanaCard
      v-for="memo in memos"
      :key="memo.id"
      :memo="memo"
      @edit="$emit('edit', $event)"
      @delete="$emit('delete', $event)"
    />
  </div>
</template>

<script setup>
import SanaCard from './SanaCard.vue'
defineProps(['label', 'memos'])
defineEmits(['edit', 'delete'])
</script>

<style scoped>
.time-group {
  margin-bottom: 24px;
}

.group-label {
  font-size: 12px;
  color: #999;
  font-weight: 500;
  margin-bottom: 8px;
  padding-left: 4px;
}
</style>

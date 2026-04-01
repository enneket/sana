<template>
  <div class="memo-card">
    <div class="memo-content">{{ memo.content }}</div>
    <div class="memo-meta">
      <span class="memo-time">{{ formatTime(memo.updated_ts) }}</span>
      <div class="memo-actions">
        <button class="action-btn" @click="$emit('edit', memo)">✎</button>
        <button class="action-btn delete" @click="$emit('delete', memo.id)">🗑</button>
      </div>
    </div>
  </div>
</template>

<script setup>
defineProps(['memo'])
defineEmits(['edit', 'delete'])

function formatTime(ts) {
  if (!ts) return ''
  const d = new Date(ts * 1000)
  const now = new Date()
  const diff = now - d
  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  return d.toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.memo-card {
  background: #fff;
  border: 1px solid #eee;
  border-radius: 10px;
  padding: 14px 16px;
  margin-bottom: 12px;
  transition: box-shadow 0.15s;
}

.memo-card:hover {
  box-shadow: 0 2px 8px rgba(0,0,0,0.08);
}

.memo-content {
  font-size: 14px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
  margin-bottom: 10px;
  color: #333;
}

.memo-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.memo-time {
  font-size: 12px;
  color: #999;
}

.memo-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.15s;
}

.memo-card:hover .memo-actions {
  opacity: 1;
}

.action-btn {
  background: none;
  border: none;
  font-size: 14px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  color: #666;
}

.action-btn:hover {
  background: #f0f0f0;
}

.action-btn.delete:hover {
  background: #fee;
  color: #e74c3c;
}
</style>

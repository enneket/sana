<template>
  <div class="memo-card">
    <div class="memo-content" v-html="renderedContent"></div>
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
import { computed } from 'vue'
import { marked } from 'marked'

const props = defineProps(['memo'])
defineEmits(['edit', 'delete'])

marked.setOptions({
  breaks: true,
  gfm: true,
})

const renderedContent = computed(() => {
  return marked.parse(props.memo.content || '')
})

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
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 20px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.03);
}

.memo-content {
  font-size: 15px;
  line-height: 1.7;
  margin-bottom: 16px;
  color: #333;
  word-break: break-word;
}

.memo-content h1,
.memo-content h2,
.memo-content h3,
.memo-content h4,
.memo-content h5,
.memo-content h6 {
  margin: 16px 0 8px 0;
  font-weight: 600;
  line-height: 1.4;
}

.memo-content h1 { font-size: 1.5em; }
.memo-content h2 { font-size: 1.3em; }
.memo-content h3 { font-size: 1.1em; }

.memo-content p {
  margin: 8px 0;
}

.memo-content a {
  color: #b8860b;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.memo-content a::before {
  content: '↗';
  font-size: 0.85em;
  opacity: 0.7;
}

.memo-content a:hover {
  text-decoration: underline;
  color: #d4a017;
}

.memo-content code {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: monospace;
  font-size: 0.9em;
}

.memo-content pre {
  background: #f5f5f5;
  padding: 12px;
  border-radius: 8px;
  overflow-x: auto;
  margin: 12px 0;
}

.memo-content pre code {
  background: none;
  padding: 0;
}

.memo-content blockquote {
  border-left: 3px solid #2ecc71;
  margin: 12px 0;
  padding: 4px 12px;
  color: #666;
  background: #f9f9f9;
}

.memo-content ul,
.memo-content ol {
  margin: 8px 0;
  padding-left: 24px;
}

.memo-content li {
  margin: 4px 0;
}

.memo-content img {
  max-width: 100%;
  border-radius: 8px;
  margin: 8px 0;
}

.memo-content hr {
  border: none;
  border-top: 1px solid #eee;
  margin: 16px 0;
}

.memo-content strong {
  font-weight: 600;
}

.memo-content em {
  font-style: italic;
}

.memo-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.memo-time {
  font-size: 13px;
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

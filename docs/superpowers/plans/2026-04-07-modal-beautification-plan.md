# Modal Beautification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Beautify SearchModal and SanaEditor with blur overlay + scale animation, replace native alert()/confirm() with custom Toast and ConfirmDialog components.

**Architecture:** A global toast/confirm system using Vue 3 provide/inject at App.vue level. Toast renders via Teleport to body with fixed positioning. ConfirmDialog is a modal with promise-based API (opens with `confirmDialog.show({ title, message, danger })` and resolves true/false). Both modals get backdrop-filter blur + CSS scale animation.

**Tech Stack:** Vue 3, CSS transitions, Teleport, provide/inject

---

## Architecture

- `App.vue` acts as the global provider for `toast()` and `confirm()` functions via `provide()`
- `Toast.vue` — single instance rendered in `App.vue` template, triggered via ref/composable
- `ConfirmDialog.vue` — single instance rendered in `App.vue`, shown/hidden via provide-injected ref
- `TimelineView.vue` injects `toast` and `confirm`, replaces `alert()` and `confirm()` calls

---

## Task 1: Create Toast.vue

**File:** Create `frontend/src/components/Toast.vue`

- [ ] **Step 1: Write the Toast component**

```vue
<template>
  <Teleport to="body">
    <div v-if="visible" class="toast-overlay" @click.self="dismiss">
      <div :class="['toast', type]" role="alert">
        <span class="toast-message">{{ message }}</span>
        <button v-if="type === 'error'" class="toast-close" @click="dismiss">✕</button>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  message: { type: String, default: '' },
  type: { type: String, default: 'info' }, // 'success' | 'error' | 'info'
  duration: { type: Number, default: 3000 }
})

const visible = ref(false)
let timer = null

function show() {
  visible.value = true
  if (props.type !== 'error') {
    timer = setTimeout(dismiss, props.duration)
  }
}

function dismiss() {
  visible.value = false
  clearTimeout(timer)
}

defineExpose({ show, dismiss })
</script>

<style scoped>
.toast-overlay {
  position: fixed;
  bottom: 32px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
  animation: toast-in 200ms ease-out;
}
.toast {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  border-radius: 12px;
  font-size: 14px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  white-space: nowrap;
  max-width: 400px;
  overflow: hidden;
  text-overflow: ellipsis;
}
.toast.success { background: #D4EDDA; color: #155724; }
.toast.error { background: #F8D7DA; color: #721C24; }
.toast.info { background: #EAE5DC; color: #3D3830; }
.toast-close {
  background: none; border: none; cursor: pointer; font-size: 14px; padding: 0 4px;
  color: inherit; opacity: 0.7;
}
.toast-close:hover { opacity: 1; }
@keyframes toast-in {
  from { opacity: 0; transform: translateX(-50%) translateY(16px); }
  to   { opacity: 1; transform: translateX(-50%) translateY(0); }
}
</style>
```

---

## Task 2: Create ConfirmDialog.vue

**File:** Create `frontend/src/components/ConfirmDialog.vue`

- [ ] **Step 1: Write the ConfirmDialog component**

```vue
<template>
  <Teleport to="body">
    <div v-if="visible" class="dialog-overlay" @click.self="resolve(false)">
      <div class="dialog-card">
        <div class="dialog-header">{{ title }}</div>
        <div class="dialog-body">{{ message }}</div>
        <div class="dialog-footer">
          <button class="btn-cancel" @click="resolve(false)">取消</button>
          <button :class="['btn-confirm', danger ? 'btn-danger' : 'btn-primary']"
                  @click="resolve(true)">{{ danger ? '删除' : '确定' }}</button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref } from 'vue'

const visible = ref(false)
const title = ref('')
const message = ref('')
const danger = ref(false)
let resolver = null

function show({ title: t, message: m, danger: d = false }) {
  title.value = t
  message.value = m
  danger.value = d
  visible.value = true
  return new Promise((resolve) => {
    resolver = resolve
  })
}

function resolve(value) {
  visible.value = false
  if (resolver) resolver(value)
}

defineExpose({ show })
</script>

<style scoped>
.dialog-overlay {
  position: fixed; inset: 0;
  background: rgba(61,56,48,0.3);
  backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center;
  z-index: 3000;
  animation: fade-in 200ms ease-out;
}
.dialog-card {
  background: #EAE5DC; border: 1px solid #D4CCBA;
  border-radius: 20px;
  width: 90%; max-width: 360px;
  box-shadow: 0 8px 32px rgba(61,56,48,0.12);
  animation: scale-in 200ms ease-out;
  overflow: hidden;
}
.dialog-header {
  padding: 20px 24px 12px;
  font-weight: 600; font-size: 16px; color: #3D3830;
}
.dialog-body {
  padding: 0 24px 16px;
  font-size: 14px; color: #6B6358; line-height: 1.5;
}
.dialog-footer {
  display: flex; justify-content: flex-end; gap: 8px;
  padding: 12px 24px 20px;
}
.btn-cancel, .btn-confirm {
  padding: 8px 20px; border-radius: 8px;
  font-size: 14px; cursor: pointer; border: none;
}
.btn-cancel { background: #E0DBD3; color: #6B6358; }
.btn-primary { background: #6B8FCC; color: #fff; }
.btn-danger { background: #C06050; color: #fff; }
@keyframes fade-in { from { opacity: 0 } to { opacity: 1 } }
@keyframes scale-in {
  from { opacity: 0; transform: scale(0.95); }
  to   { opacity: 1; transform: scale(1); }
}
</style>
```

---

## Task 3: Wire toast and confirm into App.vue

**File:** Modify `frontend/src/App.vue`

- [ ] **Step 1: Add Toast and ConfirmDialog to App.vue and provide global functions**

```vue
<template>
  <RouterView />
  <Toast ref="toastRef" />
  <ConfirmDialog ref="confirmRef" />
</template>

<script setup>
import { ref } from 'vue'
import { provide } from 'vue'
import Toast from './components/Toast.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'

const toastRef = ref(null)
const confirmRef = ref(null)

function toast(message, type = 'info', duration = 3000) {
  toastRef.value?.show()
  // Update toast message/type reactively via a simple event approach
}

provide('toast', (message, type = 'info', duration = 3000) => {
  // Will be handled by a global toast store approach below
})
</script>
```

- [ ] **Step 2: Use a simple reactive store approach — update App.vue fully**

```vue
<template>
  <RouterView />
  <Toast ref="toastRef" />
  <ConfirmDialog ref="confirmRef" />
</template>

<script setup>
import { ref, provide } from 'vue'
import { onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import api from './api/index.js'
import Toast from './components/Toast.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'

const router = useRouter()
const route = useRoute()
const toastRef = ref(null)
const confirmRef = ref(null)

// Global toast function
function toast(message, type = 'info', duration = 3000) {
  toastRef.value?.show(message, type, duration)
}
provide('toast', toast)

// Global confirm function
function confirm({ title, message, danger = false }) {
  return confirmRef.value?.show({ title, message, danger })
}
provide('confirm', confirm)

onMounted(async () => {
  try {
    await api.me()
  } catch (e) {
    if (e.message === 'unauthorized' && route.path !== '/login') {
      router.push('/login')
    }
  }
})
</script>
```

- [ ] **Step 3: Update Toast.vue to accept show(message, type, duration) signature**

Modify the `show` function in `Toast.vue` to accept parameters:

```javascript
const message = ref('')
const type = ref('info')
const duration = ref(3000)
let timer = null

function show(msg, t = 'info', dur = 3000) {
  message.value = msg
  type.value = t
  duration.value = dur
  visible.value = true
  clearTimeout(timer)
  if (t !== 'error') {
    timer = setTimeout(dismiss, dur)
  }
}
```

Remove `props.message`, `props.type`, `props.duration` from `<script setup>` and replace with the ref-based approach above.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/App.vue frontend/src/components/Toast.vue frontend/src/components/ConfirmDialog.vue
git commit -m "feat: add Toast and ConfirmDialog components with global provide/inject"
```

---

## Task 4: Update SearchModal with blur overlay and animation

**File:** Modify `frontend/src/components/SearchModal.vue`

Changes:
1. Add `@keydown.esc` to the overlay div
2. Change overlay styles: `backdrop-filter: blur(4px)`, `background: rgba(61,56,48,0.3)`
3. Add `.search-modal` animation classes
4. Update border-radius to 20px
5. Update shadow to `0 8px 32px rgba(61,56,48,0.12)`

Key CSS changes to overlay div:
```css
.search-overlay {
  background: rgba(61, 56, 48, 0.3);
  backdrop-filter: blur(4px);
  animation: overlay-in 200ms ease-out;
}
```

Add `.search-modal` animation:
```css
.search-modal {
  animation: modal-scale-in 200ms ease-out;
}
@keyframes overlay-in { from { opacity: 0 } to { opacity: 1 } }
@keyframes modal-scale-in {
  from { opacity: 0; transform: scale(0.95); }
  to   { opacity: 1; transform: scale(1); }
}
```

Update border-radius: `20px` and shadow: `0 8px 32px rgba(61,56,48,0.12)`.

Add ESC key handler to overlay:
```html
<div v-if="show" class="search-overlay" @click.self="$emit('close')" @keydown.esc="$emit('close')" tabindex="-1">
```

- [ ] **Step 1: Update SearchModal.vue styles and add animation**

Apply the overlay and modal CSS changes above. Also update `.close-btn:hover` background on hover state.

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/SearchModal.vue
git commit -m "feat(SearchModal): add blur overlay, scale animation, warm styling"
```

---

## Task 5: Update SanaEditor with blur overlay and animation

**File:** Modify `frontend/src/components/SanaEditor.vue`

Same pattern as SearchModal:
1. Overlay: `backdrop-filter: blur(4px)`, `background: rgba(61,56,48,0.3)`
2. `.editor-modal` animation: `modal-scale-in` with `transform: scale(0.95→1)`
3. Overlay: `animation: overlay-in`
4. Border-radius: `20px`, shadow: `0 8px 32px rgba(61,56,48,0.12)`
5. No ESC on SanaEditor (intentional — prevents data loss)
6. `.editor-overlay` keeps click-outside-to-close behavior (for consistency)

- [ ] **Step 1: Update SanaEditor.vue styles**

Apply same CSS pattern as SearchModal.

- [ ] **Step 2: Commit**

```bash
git add frontend/src/components/SanaEditor.vue
git commit -m "feat(SanaEditor): add blur overlay, scale animation, warm styling"
```

---

## Task 6: Update TimelineView to use toast and confirm

**File:** Modify `frontend/src/views/TimelineView.vue`

- [ ] **Step 1: Inject toast and confirm, replace alert/confirm calls**

Add to `<script setup>`:
```javascript
import { inject } from 'vue'
const toast = inject('toast')
const confirmDialog = inject('confirm')
```

Replace all calls:

**Delete confirmation** (line ~154):
```javascript
// Before:
if (!confirm('确定删除这条笔记？')) return

// After:
const ok = await confirmDialog({ title: '删除笔记', message: '确定删除这条笔记？', danger: true })
if (!ok) return
```

**Export failure** (line ~174):
```javascript
// Before:
alert('导出失败')

// After:
toast('导出失败', 'error')
```

**Import complete** (line ~185):
```javascript
// Before:
alert(`导入完成：${result.sanas_imported ?? result.memos_imported ?? 0} 条笔记`)

// After:
toast(`导入完成：${result.sanas_imported ?? result.memos_imported ?? 0} 条笔记`, 'success')
```

**Import failure** (line ~188):
```javascript
// Before:
alert('导入失败')

// After:
toast('导入失败', 'error')
```

- [ ] **Step 2: Build to verify no errors**

```bash
cd frontend && bun run build
```

Expected: build succeeds

- [ ] **Step 3: Commit**

```bash
git add frontend/src/views/TimelineView.vue
git commit -m "feat(TimelineView): replace alert/confirm with toast and confirm dialog"
```

---

## Task 7: Verify and final check

- [ ] **Step 1: Verify build passes with no errors**

```bash
cd frontend && bun run build
```

- [ ] **Step 2: Verify all acceptance criteria from spec**

- SearchModal has blur overlay + scale animation on open ✓
- SanaEditor has blur overlay + scale animation on open ✓
- ESC closes SearchModal ✓
- Native `alert('导出失败')` → Toast (error type) ✓
- Native `alert('导入完成 N 条')` → Toast (success type) ✓
- Native `confirm('确定删除')` → ConfirmDialog ✓
- All toast auto-dismiss in 3s (except error) ✓

- [ ] **Step 3: Push**

```bash
git push
```

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

function toast(message, type = 'info', duration = 3000) {
  toastRef.value?.show(message, type, duration)
}
provide('toast', toast)

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

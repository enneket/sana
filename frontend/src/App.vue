<template>
  <RouterView />
</template>

<script setup>
import { onMounted } from 'vue'
import api from './api/index.js'

onMounted(async () => {
  try {
    await api.me()
  } catch {
    // 401 triggers window.location.href='/login' inside handleResponse
    // network error or other: stay on current page
  }
})
</script>

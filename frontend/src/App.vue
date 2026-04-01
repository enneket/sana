<template>
  <RouterView />
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from './api/index.js'

const router = useRouter()

onMounted(async () => {
  try {
    await api.me()
  } catch {
    router.push('/login')
  }
})
</script>

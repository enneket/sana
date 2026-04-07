<template>
  <RouterView />
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import api from './api/index.js'

const router = useRouter()
const route = useRoute()

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

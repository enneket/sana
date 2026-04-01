<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api/index.js'

const router = useRouter()
const password = ref('')
const error = ref('')

async function login() {
  error.value = ''
  try {
    const data = await api.login(password.value)
    localStorage.setItem('token', data.token)
    router.push('/')
  } catch (e) {
    error.value = '密码错误'
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-card">
      <h1>Sana</h1>
      <p class="subtitle">输入密码登录</p>
      <form @submit.prevent="login">
        <input v-model="password" type="password" placeholder="密码" required autofocus />
        <button type="submit">登录</button>
        <p v-if="error" class="error">{{ error }}</p>
      </form>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #F5F0E8;
}
.auth-card {
  background: #EAE5DC;
  border: 1px solid #DDD8CC;
  border-radius: 8px;
  padding: 40px;
  width: 320px;
}
h1 { color: #3D3830; margin: 0 0 8px; text-align: center; font-weight: 600; font-size: 24px; letter-spacing: -0.5px; }
.subtitle { color: #A09888; text-align: center; margin: 0 0 32px; font-size: 13px; }
input {
  display: block;
  width: 100%;
  padding: 10px 12px;
  margin-bottom: 12px;
  background: #F5F0E8;
  border: 1px solid #D4CCBA;
  border-radius: 6px;
  color: #3D3830;
  font-size: 14px;
  box-sizing: border-box;
  transition: border-color 0.15s;
}
input:focus { outline: none; border-color: #6B8FCC; }
input::placeholder { color: #B8AFA0; }
button {
  width: 100%;
  padding: 10px;
  background: #6B8FCC;
  border: none;
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
  margin-top: 8px;
  font-weight: 500;
  transition: background 0.15s;
}
button:hover { background: #5A7EBB; }
.error { color: #C06050; font-size: 13px; text-align: center; margin-top: 12px; }
</style>

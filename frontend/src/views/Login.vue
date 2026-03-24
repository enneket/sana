<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { apiFetch, setToken } from '../main.js'

const router = useRouter()
const username = ref('')
const password = ref('')
const error = ref('')

async function login() {
  error.value = ''
  try {
    const data = await apiFetch('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username: username.value, password: password.value })
    })
    setToken(data.token)
    router.push('/')
  } catch (e) {
    error.value = '用户名或密码错误'
  }
}
</script>

<template>
  <div class="auth-page">
    <div class="auth-card">
      <h1>Sana</h1>
      <p class="subtitle">登录到 Sana</p>
      <form @submit.prevent="login">
        <input v-model="username" placeholder="用户名" required />
        <input v-model="password" type="password" placeholder="密码" required />
        <button type="submit">登录</button>
        <p v-if="error" class="error">{{ error }}</p>
      </form>
      <p class="switch">
        没有账号？<router-link to="/register">注册</router-link>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #1a1a2e;
}
.auth-card {
  background: #16162a;
  border: 1px solid #2a2a4a;
  border-radius: 12px;
  padding: 40px;
  width: 320px;
}
h1 { color: #e8e8e8; margin: 0 0 8px; text-align: center; }
.subtitle { color: #888; text-align: center; margin: 0 0 32px; font-size: 14px; }
input {
  display: block;
  width: 100%;
  padding: 12px;
  margin-bottom: 12px;
  background: #1a1a2e;
  border: 1px solid #2a2a4a;
  border-radius: 6px;
  color: #e8e8e8;
  font-size: 14px;
  box-sizing: border-box;
}
input:focus { outline: none; border-color: #4a9eff; }
button {
  width: 100%;
  padding: 12px;
  background: #4a9eff;
  border: none;
  border-radius: 6px;
  color: #fff;
  font-size: 14px;
  cursor: pointer;
  margin-top: 8px;
}
button:hover { background: #3a8eef; }
.error { color: #ff6b6b; font-size: 13px; text-align: center; margin-top: 12px; }
.switch { color: #888; font-size: 13px; text-align: center; margin-top: 20px; }
.switch a { color: #4a9eff; }
</style>

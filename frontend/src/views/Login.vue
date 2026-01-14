<template>
  <div class="login-container">
    <div class="login-box">
      <div class="logo">
        <div class="logo-icon">G</div>
      </div>
      <h1 class="title">GPanel</h1>
      <p class="subtitle">服务器管理控制台</p>
      <form @submit.prevent="handleLogin" class="login-form">
        <div class="form-group">
          <label for="username">用户名</label>
          <input
            id="username"
            v-model="username"
            type="text"
            placeholder="请输入用户名"
            required
          />
        </div>
        <div class="form-group">
          <label for="password">密码</label>
          <input
            id="password"
            v-model="password"
            type="password"
            placeholder="请输入密码"
            required
          />
        </div>
        <div v-if="errorMessage" class="error-message">
          {{ errorMessage }}
        </div>
        <button type="submit" class="login-button" :disabled="loading">
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'

const router = useRouter()
const username = ref('')
const password = ref('')
const loading = ref(false)
const errorMessage = ref('')

const handleLogin = async () => {
  loading.value = true
  errorMessage.value = ''

  try {
    const response = await axios.post('/api/v1/auth/login', {
      username: username.value,
      password: password.value,
    })

    const token = response.data.token
    localStorage.setItem('token', token)
    router.push('/dashboard')
  } catch (error: any) {
    if (error.response && error.response.data) {
      errorMessage.value = error.response.data.error || '登录失败'
    } else {
      errorMessage.value = '登录失败，请检查网络连接'
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--primary-light) 0%, var(--primary) 50%, var(--primary-dark) 100%);
}

.login-box {
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  padding: 2.5rem;
  box-shadow: var(--shadow-lg);
  width: 100%;
  max-width: 340px;
}

.logo {
  display: flex;
  justify-content: center;
  margin-bottom: 1.5rem;
}

.logo-icon {
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: 700;
  color: white;
  box-shadow: var(--shadow-md);
}

.title {
  font-size: 1.75rem;
  font-weight: 600;
  text-align: center;
  margin: 0 0 0.5rem 0;
  color: var(--text-primary);
  letter-spacing: -0.5px;
}

.subtitle {
  text-align: center;
  color: var(--text-secondary);
  margin: 0 0 2rem 0;
  font-size: 0.85rem;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-group label {
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--text-primary);
}

.form-group input {
  padding: 0.6rem 0.75rem;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  transition: all 0.2s;
  background: var(--bg-color);
}

.form-group input:focus {
  outline: none;
  border-color: var(--primary);
  background: var(--card-bg);
  box-shadow: 0 0 0 2px var(--primary-light);
}

.form-group input::placeholder {
  color: var(--text-secondary);
  opacity: 0.7;
}

.error-message {
  color: #e74c3c;
  font-size: 0.8rem;
  text-align: center;
  padding: 0.5rem;
  background: #fdf2f2;
  border-radius: var(--radius-sm);
  border: 1px solid #fee2e2;
}

.login-button {
  padding: 0.65rem;
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  margin-top: 0.5rem;
}

.login-button:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 2px 8px rgba(122, 201, 224, 0.4);
}

.login-button:active:not(:disabled) {
  transform: translateY(0);
}

.login-button:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
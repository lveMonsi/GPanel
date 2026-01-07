<template>
  <div class="dashboard">
    <header class="header">
      <h1>GPanel 仪表盘</h1>
    </header>
    <main class="content">
      <div class="card">
        <h2>系统信息</h2>
        <div v-if="systemInfo">
          <p><strong>主机名:</strong> {{ systemInfo.hostname }}</p>
          <p><strong>操作系统:</strong> {{ systemInfo.os }}</p>
          <p><strong>版本:</strong> {{ systemInfo.version }}</p>
        </div>
        <div v-else>
          <p>加载中...</p>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'

interface SystemInfo {
  hostname: string
  os: string
  version: string
}

const systemInfo = ref<SystemInfo | null>(null)

const fetchSystemInfo = async () => {
  try {
    const response = await axios.get('/api/v1/system/info')
    systemInfo.value = response.data
  } catch (error) {
    console.error('获取系统信息失败:', error)
  }
}

onMounted(() => {
  fetchSystemInfo()
})
</script>

<style scoped>
.dashboard {
  min-height: 100vh;
  background-color: #f5f5f5;
}

.header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  padding: 1rem 2rem;
}

.header h1 {
  font-size: 1.5rem;
}

.content {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.card {
  background: white;
  border-radius: 8px;
  padding: 2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.card h2 {
  margin-bottom: 1rem;
  color: #333;
}

.card p {
  margin: 0.5rem 0;
  color: #666;
}
</style>
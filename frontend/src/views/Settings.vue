<template>
  <div class="settings">
    <header class="header">
      <h1>面板设置</h1>
    </header>

    <main class="content">
      <div class="settings-card">
        <h2>服务器配置</h2>
        <form @submit.prevent="handleSave">
          <div class="form-section">
            <h3>服务器设置</h3>
            <div class="form-group">
              <label for="server-port">端口</label>
              <input
                id="server-port"
                v-model="config.server.port"
                type="text"
                placeholder="8080"
              />
            </div>
            <div class="form-group">
              <label for="server-mode">运行模式</label>
              <select id="server-mode" v-model="config.server.mode">
                <option value="debug">Debug</option>
                <option value="release">Release</option>
              </select>
            </div>
            <div class="form-group">
              <label for="security-entrance">安全入口</label>
              <input
                id="security-entrance"
                v-model="config.securityEntrance"
                type="text"
                placeholder="/"
                @blur="ensureLeadingSlash"
              />
              <small class="hint">访问面板的安全路径，例如：/abc123</small>
            </div>
          </div>

          <div class="form-section">
            <h3>数据库配置</h3>
            <div class="form-group">
              <label for="db-host">主机地址</label>
              <input
                id="db-host"
                v-model="config.database.host"
                type="text"
                placeholder="localhost"
              />
            </div>
            <div class="form-group">
              <label for="db-port">端口</label>
              <input
                id="db-port"
                v-model.number="config.database.port"
                type="number"
                placeholder="3306"
              />
            </div>
            <div class="form-group">
              <label for="db-user">用户名</label>
              <input
                id="db-user"
                v-model="config.database.user"
                type="text"
                placeholder="root"
              />
            </div>
            <div class="form-group">
              <label for="db-password">密码</label>
              <input
                id="db-password"
                v-model="config.database.password"
                type="password"
                placeholder="请输入密码"
              />
            </div>
            <div class="form-group">
              <label for="db-name">数据库名</label>
              <input
                id="db-name"
                v-model="config.database.name"
                type="text"
                placeholder="gpanel"
              />
            </div>
          </div>

          <div class="form-actions">
            <button type="submit" class="btn btn-primary" :disabled="loading">
              {{ loading ? '保存中...' : '保存配置' }}
            </button>
            <button type="button" class="btn btn-secondary" @click="handleCancel">
              取消
            </button>
          </div>
        </form>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'

interface Config {
  server: {
    port: string
    mode: string
  }
  database: {
    host: string
    port: number
    user: string
    password: string
    name: string
  }
  securityEntrance: string
  initialized: boolean
}

const router = useRouter()
const loading = ref(false)
const config = reactive<Config>({
  server: {
    port: '8080',
    mode: 'debug'
  },
  database: {
    host: 'localhost',
    port: 3306,
    user: 'root',
    password: '',
    name: 'gpanel'
  },
  securityEntrance: '/',
  initialized: false
})

const fetchConfig = async () => {
  try {
    const response = await axios.get('/api/v1/config')
    const data = response.data

    // 将扁平的键值对转换为嵌套的配置对象
    config.server.port = data.ServerPort || '8080'
    config.server.mode = data.ServerMode || 'debug'
    config.securityEntrance = data.SecurityEntrance || '/'
    config.initialized = data.Initialized === 'true'
    config.database.host = data.DatabaseHost || 'localhost'
    config.database.port = parseInt(data.DatabasePort) || 3306
    config.database.user = data.DatabaseUser || 'root'
    config.database.password = data.DatabasePassword || ''
    config.database.name = data.DatabaseName || 'gpanel'
  } catch (error) {
    console.error('获取配置失败:', error)
  }
}

const handleSave = async () => {
  loading.value = true
  try {
    // 将嵌套的配置对象转换为扁平的键值对
    const flatConfig: Record<string, string> = {
      ServerPort: config.server.port,
      ServerMode: config.server.mode,
      SecurityEntrance: config.securityEntrance,
      Initialized: config.initialized ? 'true' : 'false',
      DatabaseHost: config.database.host,
      DatabasePort: String(config.database.port),
      DatabaseUser: config.database.user,
      DatabasePassword: config.database.password,
      DatabaseName: config.database.name
    }

    await axios.post('/api/v1/config', flatConfig)
    alert('配置保存成功！\n\n系统将在2秒后自动重启以加载新配置...')
    await axios.post('/api/v1/server/restart')
  } catch (error) {
    console.error('保存配置失败:', error)
    alert('保存配置失败，请重试')
  } finally {
    loading.value = false
  }
}

const handleCancel = () => {
  router.push('/dashboard')
}

const ensureLeadingSlash = () => {
  if (config.securityEntrance && !config.securityEntrance.startsWith('/')) {
    config.securityEntrance = '/' + config.securityEntrance
  }
}

onMounted(() => {
  fetchConfig()
})
</script>

<style scoped>
.settings {
}

.header {
  background: white;
  padding: 1.5rem 2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header h1 {
  margin: 0;
  font-size: 1.5rem;
  color: #333;
}

.content {
  padding: 2rem;
  max-width: 800px;
  margin: 0 auto;
}

.settings-card {
  background: white;
  border-radius: 12px;
  padding: 2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.settings-card h2 {
  margin: 0 0 1.5rem 0;
  font-size: 1.5rem;
  color: #333;
}

.form-section {
  margin-bottom: 2rem;
}

.form-section h3 {
  margin: 0 0 1rem 0;
  font-size: 1.1rem;
  color: #666;
  padding-bottom: 0.5rem;
  border-bottom: 2px solid #667eea;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #333;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #ddd;
  border-radius: 6px;
  font-size: 1rem;
  transition: border-color 0.3s;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid #eee;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: #f5f5f5;
  color: #666;
}

.btn-secondary:hover {
  background: #e0e0e0;
}
</style>
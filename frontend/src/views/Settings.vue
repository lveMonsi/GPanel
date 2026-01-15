<template>
  <div class="settings">
    <header class="header">
      <h1>面板设置</h1>
    </header>

    <main class="content">
      <Modal
        v-model:visible="modalVisible"
        :title="modalTitle"
        :message="modalMessage"
        :show-confirm="modalTitle === '即将重启面板' || modalTitle === '安全提醒'"
        @confirm="handleSave"
      />
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
              <label for="security-enabled">安全入口</label>
              <div class="switch-wrapper">
                <el-switch
                  id="security-enabled"
                  v-model="config.securityEnabled"
                  active-color="var(--primary)"
                  inactive-color="var(--border-color)"
                  @change="handleSecurityChange"
                />
                <span class="switch-label">{{ config.securityEnabled ? '已开启' : '已关闭' }}</span>
              </div>
              <small class="hint">关闭安全入口后，任何人都可以访问面板，存在安全风险</small>
            </div>
            <div class="form-group">
              <label for="security-entrance">安全入口</label>
              <div class="input-with-button">
                <input
                  id="security-entrance"
                  v-model="config.securityEntrance"
                  type="text"
                  placeholder="/"
                  @blur="ensureLeadingSlash"
                />
                <button type="button" class="btn-generate" @click="generateRandomEntrance">
                  <el-icon><Refresh /></el-icon>
                </button>
              </div>
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
            <button type="button" class="btn btn-primary" :disabled="loading" @click="configChanged ? showRestartConfirm() : handleSave()">
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
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'
import Modal from '@/components/Modal.vue'
import { Refresh } from '@element-plus/icons-vue'

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
  securityEnabled: boolean
  initialized: boolean
}

const router = useRouter()
const loading = ref(false)
const modalVisible = ref(false)
const modalMessage = ref('')
const modalTitle = ref('提示')
const configChanged = ref(false)
const originalConfig = ref<Config | null>(null)
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
  securityEnabled: true,
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
    config.securityEnabled = data.SecurityEnabled !== 'false'
    config.initialized = data.Initialized === 'true'
    config.database.host = data.DatabaseHost || 'localhost'
    config.database.port = parseInt(data.DatabasePort) || 3306
    config.database.user = data.DatabaseUser || 'root'
    config.database.password = data.DatabasePassword || ''
    config.database.name = data.DatabaseName || 'gpanel'

    // 保存原始配置用于对比
    originalConfig.value = JSON.parse(JSON.stringify(config))
    configChanged.value = false
  } catch (error) {
    console.error('获取配置失败:', error)
  }
}

const showModal = (title: string, message: string, showConfirm: boolean = false) => {
  modalTitle.value = title
  modalMessage.value = message
  modalVisible.value = true
  // 存储是否显示确认按钮的状态
  ;(modalVisible as any).showConfirm = showConfirm
}

const handleCancel = () => {
  router.push('/dashboard')
}

const ensureLeadingSlash = () => {
  if (config.securityEntrance && !config.securityEntrance.startsWith('/')) {
    config.securityEntrance = '/' + config.securityEntrance
  }
}

const generateRandomEntrance = () => {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789'
  let result = '/'
  for (let i = 0; i < 12; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  config.securityEntrance = result
  checkConfigChanged()
}

const handleSecurityChange = (value: boolean) => {
  if (!value) {
    // 用户尝试关闭安全入口，先弹窗确认
    // 暂时恢复开关状态
    config.securityEnabled = true
    showModal('安全提醒', '关闭安全入口后，任何人都可以访问面板，存在严重安全风险！\n\n建议仅在测试环境或内网环境中关闭。\n\n是否确认关闭？', true)
  } else {
    checkConfigChanged()
  }
}

const checkConfigChanged = () => {
  if (!originalConfig.value) return
  configChanged.value = JSON.stringify(config) !== JSON.stringify(originalConfig.value)
}

const showRestartConfirm = () => {
  modalTitle.value = '即将重启面板'
  modalMessage.value = '配置修改后需要重启面板才能生效。\n\n请确保已妥善保存所有数据，避免造成数据损失。\n\n是否继续保存并重启？'
  modalVisible.value = true
}

const handleSave = async () => {
  // 如果是安全提醒弹窗，确认后关闭安全入口
  if (modalTitle.value === '安全提醒') {
    config.securityEnabled = false
    checkConfigChanged()
    modalVisible.value = false
    return
  }

  loading.value = true
  try {
    // 将嵌套的配置对象转换为扁平的键值对
    const flatConfig: Record<string, string> = {
      ServerPort: config.server.port,
      ServerMode: config.server.mode,
      SecurityEntrance: config.securityEntrance,
      SecurityEnabled: config.securityEnabled ? 'true' : 'false',
      Initialized: config.initialized ? 'true' : 'false',
      DatabaseHost: config.database.host,
      DatabasePort: String(config.database.port),
      DatabaseUser: config.database.user,
      DatabasePassword: config.database.password,
      DatabaseName: config.database.name
    }

    await axios.post('/api/v1/config', flatConfig)
    await axios.post('/api/v1/server/restart')
    showModal('保存成功', '配置保存成功！\n\n系统将在2秒后自动重启以加载新配置...')

    // 2秒后自动刷新页面
    setTimeout(() => {
      window.location.reload()
    }, 2000)
  } catch (error) {
    console.error('保存配置失败:', error)
    showModal('保存失败', '保存配置失败，请重试')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchConfig()
})

// 监听配置变更
watch(() => config, () => {
  checkConfigChanged()
}, { deep: true })
</script>

<style scoped>
.settings {
}

.header {
  background: var(--card-bg);
  padding: 1rem 1.25rem;
  box-shadow: var(--shadow-sm);
  border-bottom: 1px solid var(--border-color);
}

.header h1 {
  margin: 0;
  font-size: 1.1rem;
  color: var(--text-primary);
  font-weight: 600;
  letter-spacing: -0.3px;
}

.content {
  padding: 1rem;
  max-width: 600px;
  margin: 0 auto;
}

.settings-card {
  background: var(--card-bg);
  border-radius: var(--radius-md);
  padding: 1.25rem;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
}

.settings-card h2 {
  margin: 0 0 1.25rem 0;
  font-size: 1rem;
  color: var(--text-primary);
  font-weight: 600;
}

.form-section {
  margin-bottom: 1.5rem;
}

.form-section h3 {
  margin: 0 0 0.75rem 0;
  font-size: 0.85rem;
  color: var(--text-secondary);
  padding-bottom: 0.4rem;
  border-bottom: 2px solid var(--primary);
  font-weight: 500;
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.4rem;
  font-weight: 500;
  color: var(--text-primary);
  font-size: 0.8rem;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 0.5rem 0.6rem;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  transition: all 0.2s;
  background: var(--bg-color);
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: var(--primary);
  background: var(--card-bg);
  box-shadow: 0 0 0 2px var(--primary-light);
}

.form-group input::placeholder {
  color: var(--text-secondary);
  opacity: 0.7;
}

.form-group .hint {
  display: block;
  margin-top: 0.3rem;
  font-size: 0.7rem;
  color: var(--text-secondary);
}

.input-with-button {
  display: flex;
  gap: 0.4rem;
}

.input-with-button input {
  flex: 1;
}

.btn-generate {
  padding: 0.5rem 0.6rem;
  background: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--primary-dark);
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-generate:hover {
  background: var(--primary-light);
  border-color: var(--primary);
}

.btn-generate:active {
  transform: scale(0.95);
}

.form-actions {
  display: flex;
  gap: 0.5rem;
  margin-top: 1.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border-color);
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
  color: white;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(122, 201, 224, 0.3);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-secondary {
  background: var(--bg-color);
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
}

.btn-secondary:hover {
  background: var(--border-color);
  color: var(--text-primary);
}

.switch-wrapper {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.switch-label {
  font-size: 0.8rem;
  color: var(--text-secondary);
  font-weight: 500;
}
</style>
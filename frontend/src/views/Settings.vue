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
      <EditModal
        v-model:visible="editModalVisible"
        :title="editModalTitle"
        :field-name="editFieldName"
        :field-value="editFieldValue"
        :field-type="editFieldType"
        @save="handleEditSave"
      />
      <div class="settings-container">
        <div class="tabs">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            :class="['tab-button', { active: activeTab === tab.key }]"
            @click="activeTab = tab.key"
          >
            {{ tab.label }}
          </button>
        </div>

        <div class="settings-content">
          <!-- 面板标签页 -->
          <div v-show="activeTab === 'panel'" class="tab-content">
            <div class="settings-card">
              <h2>面板设置</h2>
              <form @submit.prevent="handleSave">
                <div class="form-section">
                  <div class="form-group">
                    <label for="panel-user">面板用户</label>
                    <div class="input-with-edit">
                      <input
                        id="panel-user"
                        v-model="config.panelUser"
                        type="text"
                        disabled
                        placeholder="admin"
                      />
                      <button type="button" class="btn-edit" @click="openEditModal('panelUser', '面板用户', config.panelUser, 'text')">
                        <el-icon><Edit /></el-icon>
                      </button>
                    </div>
                  </div>
                  <div class="form-group">
                    <label for="panel-password">面板密码</label>
                    <div class="input-with-edit">
                      <input
                        id="panel-password"
                        v-model="config.panelPassword"
                        type="password"
                        disabled
                        placeholder="••••••••"
                      />
                      <button type="button" class="btn-edit" @click="openEditModal('panelPassword', '面板密码', config.panelPassword, 'password')">
                        <el-icon><Edit /></el-icon>
                      </button>
                    </div>
                  </div>
                  <div class="form-group">
                    <label for="session-timeout">超时时间（秒）</label>
                    <div class="input-with-edit">
                      <input
                        id="session-timeout"
                        v-model="config.sessionTimeout"
                        type="number"
                        disabled
                        placeholder="86400"
                      />
                      <button type="button" class="btn-edit" @click="openEditModal('sessionTimeout', '超时时间', config.sessionTimeout, 'number')">
                        <el-icon><Edit /></el-icon>
                      </button>
                    </div>
                    <small class="hint">如果用户超过指定秒数未操作面板，面板将自动退出登录</small>
                  </div>
                  <div class="form-group">
                    <label for="server-address">服务器地址</label>
                    <div class="input-with-edit">
                      <input
                        id="server-address"
                        v-model="config.serverAddress"
                        type="text"
                        disabled
                        placeholder="http://localhost:8080"
                      />
                      <button type="button" class="btn-edit" @click="openEditModal('serverAddress', '服务器地址', config.serverAddress, 'text')">
                        <el-icon><Edit /></el-icon>
                      </button>
                    </div>
                    <small class="hint">支持输入 ip 或者域名</small>
                  </div>
                </div>

                <div class="form-actions">
                  <button type="button" class="btn btn-primary" :disabled="loading" @click="configChanged ? showRestartConfirm() : handleSave()">
                    {{ loading ? '保存中...' : '保存配置' }}
                  </button>
                  <button type="button" class="btn btn-reset" @click="handleResetToDefault">
                    重置为默认
                  </button>
                </div>
              </form>
            </div>
          </div>

          <!-- 安全标签页 -->
          <div v-show="activeTab === 'security'" class="tab-content">
            <div class="settings-card">
              <h2>安全设置</h2>
              <form @submit.prevent="handleSave">
                <div class="form-section">
                  <div class="form-group">
                    <label for="server-port">面板端口</label>
                    <div class="input-with-edit">
                      <input
                        id="server-port"
                        v-model="config.serverPort"
                        type="text"
                        disabled
                        placeholder="8080"
                      />
                      <button type="button" class="btn-edit" @click="openEditModal('serverPort', '面板端口', config.serverPort, 'text')">
                        <el-icon><Edit /></el-icon>
                      </button>
                    </div>
                  </div>
                  <div class="form-group">
                    <label for="listen-address">监听地址</label>
                    <div class="input-with-edit">
                      <input
                        id="listen-address"
                        v-model="config.listenAddress"
                        type="text"
                        disabled
                        placeholder="0.0.0.0"
                      />
                      <button type="button" class="btn-edit" @click="openEditModal('listenAddress', '监听地址', config.listenAddress, 'text')">
                        <el-icon><Edit /></el-icon>
                      </button>
                    </div>
                  </div>
                  <div class="form-group">
                    <label for="security-entrance">安全入口</label>
                    <div class="input-with-edit">
                      <input
                        id="security-entrance"
                        v-model="config.securityEntrance"
                        type="text"
                        disabled
                        placeholder="/"
                      />
                      <button type="button" class="btn-edit" @click="openEditModal('securityEntrance', '安全入口', config.securityEntrance, 'security')">
                        <el-icon><Edit /></el-icon>
                      </button>
                    </div>
                    <small class="hint">开启安全入口后只能通过指定安全入口登录面板</small>
                  </div>
                  <div class="form-group">
                    <label for="password-complexity">密码复杂度验证</label>
                    <div class="switch-wrapper">
                      <el-switch
                        id="password-complexity"
                        v-model="config.passwordComplexityCheck"
                        active-color="var(--primary)"
                        inactive-color="var(--border-color)"
                        @change="handlePasswordComplexityChange"
                      />
                      <span class="switch-label">{{ config.passwordComplexityCheck ? '已开启' : '已关闭' }}</span>
                    </div>
                    <small class="hint">开启后密码必须满足长度为 8-30 位且包含字母、数字、特殊字符至少两项</small>
                  </div>
                </div>

                <div class="form-actions">
                  <button type="button" class="btn btn-primary" :disabled="loading" @click="configChanged ? showRestartConfirm() : handleSave()">
                    {{ loading ? '保存中...' : '保存配置' }}
                  </button>
                  <button type="button" class="btn btn-reset" @click="handleResetToDefault">
                    重置为默认
                  </button>
                </div>
              </form>
            </div>
          </div>

          <!-- 关于标签页 -->
          <div v-show="activeTab === 'about'" class="tab-content">
            <div class="settings-card about-card">
              <h2>关于</h2>
              <div class="about-content">
                <div class="about-item">
                  <div class="about-label">版本</div>
                  <div class="about-value">{{ versionInfo.version || '加载中...' }}</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import axios from '@/utils/axios'
import Modal from '@/components/Modal.vue'
import EditModal from '@/components/EditModal.vue'
import { Refresh, Edit } from '@element-plus/icons-vue'

interface Config {
  panelUser: string
  panelPassword: string
  sessionTimeout: number
  serverAddress: string
  serverPort: string
  listenAddress: string
  securityEntrance: string
  passwordComplexityCheck: boolean
}

interface VersionInfo {
  version: string
}

const router = useRouter()
const loading = ref(false)
const modalVisible = ref(false)
const modalMessage = ref('')
const modalTitle = ref('提示')
const configChanged = ref(false)
const originalConfig = ref<Config | null>(null)

// 编辑弹窗相关状态
const editModalVisible = ref(false)
const editModalTitle = ref('')
const editFieldName = ref('')
const editFieldValue = ref('')
const editFieldType = ref<'text' | 'password' | 'number'>('text')

// 标签页相关
const activeTab = ref('panel')
const tabs = [
  { key: 'panel', label: '面板' },
  { key: 'security', label: '安全' },
  { key: 'about', label: '关于' }
]

// 版本信息
const versionInfo = ref<VersionInfo>({
  version: ''
})

const config = reactive<Config>({
  panelUser: 'admin',
  panelPassword: 'admin123',
  sessionTimeout: 86400,
  serverAddress: '',
  serverPort: '8080',
  listenAddress: '0.0.0.0',
  securityEntrance: '/',
  passwordComplexityCheck: false
})

const fetchConfig = async () => {
  try {
    const response = await axios.get('/api/v1/config')
    const data = response.data

    config.panelUser = data.PanelUser || 'admin'
    config.panelPassword = data.PanelPassword || 'admin123'
    config.sessionTimeout = parseInt(data.SessionTimeout) || 86400
    config.serverAddress = data.ServerAddress || ''
    config.serverPort = data.ServerPort || '8080'
    config.listenAddress = data.ListenAddress || '0.0.0.0'
    config.securityEntrance = data.SecurityEntrance || '/'
    config.passwordComplexityCheck = data.PasswordComplexityCheck === 'true'

    // 保存原始配置用于对比
    originalConfig.value = JSON.parse(JSON.stringify(config))
    configChanged.value = false
  } catch (error) {
    console.error('获取配置失败:', error)
  }
}

const fetchVersion = async () => {
  try {
    const response = await axios.get('/api/v1/system/version')
    versionInfo.value = response.data
  } catch (error) {
    console.error('获取版本信息失败:', error)
  }
}

const showModal = (title: string, message: string, showConfirm: boolean = false) => {
  modalTitle.value = title
  modalMessage.value = message
  modalVisible.value = true
  ;(modalVisible as any).showConfirm = showConfirm
}

const handleCancel = () => {
  router.push('/dashboard')
}

const handleResetToDefault = () => {
  modalTitle.value = '重置为默认配置'
  modalMessage.value = '确定要将所有配置重置为默认值吗？\n\n注意：安全入口将保持不变。'
  modalVisible.value = true
  ;(modalVisible as any).showConfirm = true
  ;(modalVisible as any).resetConfirm = true
}

const handlePasswordComplexityChange = (value: boolean) => {
  checkConfigChanged()
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
  loading.value = true
  try {
    const isReset = (modalVisible as any).resetConfirm === true
    const flatConfig: Record<string, string> = {
      PanelUser: config.panelUser,
      PanelPassword: config.panelPassword,
      SessionTimeout: String(config.sessionTimeout),
      ServerAddress: config.serverAddress,
      ServerPort: config.serverPort,
      ListenAddress: config.listenAddress,
      SecurityEntrance: config.securityEntrance,
      PasswordComplexityCheck: config.passwordComplexityCheck ? 'true' : 'false'
    }

    if (isReset) {
      const currentSecurityEntrance = config.securityEntrance
      flatConfig.PanelUser = 'admin'
      flatConfig.PanelPassword = 'admin123'
      flatConfig.SessionTimeout = '86400'
      flatConfig.ServerAddress = ''
      flatConfig.ServerPort = '8080'
      flatConfig.ListenAddress = '0.0.0.0'
      flatConfig.SecurityEntrance = currentSecurityEntrance
      flatConfig.PasswordComplexityCheck = 'false'
    }

    await axios.post('/api/v1/config', flatConfig)
    await axios.post('/api/v1/server/restart')
    showModal('保存成功', '配置保存成功！\n\n系统将在2秒后自动重启以加载新配置...')

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
  fetchVersion()
})

watch(() => config, () => {
  checkConfigChanged()
}, { deep: true })

const openEditModal = (fieldName: string, title: string, value: string | number, type: 'text' | 'password' | 'number') => {
  editFieldName.value = fieldName
  editModalTitle.value = title
  editFieldValue.value = String(value)
  editFieldType.value = type
  editModalVisible.value = true
}

const handleEditSave = (value: string) => {
  switch (editFieldName.value) {
    case 'panelUser':
      config.panelUser = value
      break
    case 'panelPassword':
      config.panelPassword = value
      break
    case 'sessionTimeout':
      config.sessionTimeout = parseInt(value) || 86400
      break
    case 'serverAddress':
      config.serverAddress = value
      break
    case 'serverPort':
      config.serverPort = value
      break
    case 'listenAddress':
      config.listenAddress = value
      break
    case 'securityEntrance':
      if (value && !value.startsWith('/')) {
        config.securityEntrance = '/' + value
      } else {
        config.securityEntrance = value
      }
      break
  }
  checkConfigChanged()
  editModalVisible.value = false
}
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
}

.settings-container {
  display: flex;
  gap: 1rem;
  height: calc(100vh - 120px);
}

.tabs {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  min-width: 120px;
}

.tab-button {
  padding: 0.75rem 1rem;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.tab-button:hover {
  background: var(--bg-color);
  color: var(--text-primary);
}

.tab-button.active {
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
  color: white;
  border-color: var(--primary);
}

.settings-content {
  flex: 1;
  overflow-y: auto;
}

.tab-content {
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.settings-card {
  background: var(--card-bg);
  border-radius: var(--radius-md);
  padding: 1.25rem;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
  max-width: 600px;
}

.about-card {
  max-width: 400px;
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

.form-group input {
  width: 100%;
  padding: 0.5rem 0.6rem;
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

.btn-reset {
  background: #fff0f0;
  color: #e74c3c;
  border: 1px solid #fadbd8;
}

.btn-reset:hover {
  background: #fadbd8;
  border-color: #e74c3c;
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

.input-with-edit {
  display: flex;
  gap: 0.4rem;
  align-items: center;
}

.input-with-edit input {
  flex: 1;
}

.input-with-edit input:disabled {
  background: var(--bg-color);
  color: var(--text-secondary);
  cursor: not-allowed;
}

.btn-edit {
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
  min-width: 32px;
}

.btn-edit:hover {
  background: var(--primary-light);
  border-color: var(--primary);
}

.btn-edit:active {
  transform: scale(0.95);
}

.about-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.about-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.about-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.about-value {
  font-size: 0.9rem;
  color: var(--text-primary);
  font-weight: 500;
  font-family: 'Consolas', 'Monaco', monospace;
}
</style>
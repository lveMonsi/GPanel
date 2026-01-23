<template>
  <div class="dashboard">
    <main class="content">
      <div class="card-grid">
        <!-- 左上：告警通知卡片 -->
        <div class="card alert-card">
          <div class="card-header">
            <el-icon class="card-icon"><Bell /></el-icon>
            <span class="card-title">告警通知</span>
            <span class="alert-count">{{ alerts.length }}</span>
          </div>
          <div class="card-body">
            <div class="alert-items">
              <div
                v-for="alert in alerts"
                :key="alert.id"
                :class="['alert-item', `alert-${alert.level}`]"
                @click="showAlertDetail(alert)"
              >
                <div class="alert-item-icon">
                  <el-icon>
                    <component :is="getAlertIcon(alert.level)" />
                  </el-icon>
                </div>
                <div class="alert-item-content">
                  <div class="alert-item-title">{{ alert.title }}</div>
                  <div class="alert-item-time">{{ formatAlertTime(alert.time) }}</div>
                </div>
              </div>
              <div v-if="alerts.length === 0" class="alert-empty">
                <el-icon><Check /></el-icon>
                <span>暂无告警</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 右上：系统信息卡片 -->
        <div class="card system-card">
          <div class="card-header">
            <el-icon class="card-icon"><Monitor /></el-icon>
            <span class="card-title">系统信息</span>
          </div>
          <div class="card-body">
            <div v-if="systemInfo" class="system-details">
              <div class="detail-row">
                <span class="detail-label">主机名</span>
                <span class="detail-value">{{ systemInfo.hostname }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">系统</span>
                <span class="detail-value">{{ systemInfo.platform }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">内核</span>
                <span class="detail-value">{{ systemInfo.kernelVersion }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">架构</span>
                <span class="detail-value">{{ systemInfo.kernelArch }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">地址</span>
                <span class="detail-value">{{ systemInfo.hostAddress || '未知' }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">启动</span>
                <span class="detail-value">{{ formatBootTime(systemInfo.bootTime) }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">运行</span>
                <span class="detail-value">{{ formatUptime(systemInfo.uptime) }}</span>
              </div>
            </div>
            <div v-else class="loading">加载中...</div>
          </div>
        </div>

        <!-- 左下：仪表盘卡片 -->
        <div class="card gauges-card">
          <div class="card-header">
            <el-icon class="card-icon"><TrendCharts /></el-icon>
            <span class="card-title">系统监控</span>
          </div>
          <div class="card-body">
            <SystemStatus
              v-if="currentInfo"
              :cpu-info="currentInfo.cpuInfo"
              :memory-info="currentInfo.memoryInfo"
              :load-info="currentInfo.loadInfo"
              :disk-info="currentInfo.diskInfo"
            />
          </div>
        </div>
      </div>
    </main>
  </div>

  <!-- 告警详情弹窗 -->
  <el-dialog
    v-model="alertDetailVisible"
    :title="currentAlert?.title"
    width="500px"
    class="alert-detail-dialog"
  >
    <div v-if="currentAlert" class="alert-detail">
      <div :class="['alert-detail-level', `level-${currentAlert.level}`]">
        <el-icon class="level-icon">
          <component :is="getAlertIcon(currentAlert.level)" />
        </el-icon>
        <span class="level-text">{{ getAlertLevelText(currentAlert.level) }}</span>
      </div>
      <div class="alert-detail-section">
        <div class="section-label">告警时间</div>
        <div class="section-value">{{ formatAlertTime(currentAlert.time) }}</div>
      </div>
      <div class="alert-detail-section">
        <div class="section-label">告警详情</div>
        <div class="section-value">{{ currentAlert.message }}</div>
      </div>
      <div class="alert-detail-section">
        <div class="section-label">建议操作</div>
        <div class="section-value">{{ currentAlert.suggestion }}</div>
      </div>
    </div>
    <template #footer>
      <el-button @click="alertDetailVisible = false">关闭</el-button>
      <el-button type="primary" @click="dismissAlert">忽略</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import axios from '@/utils/axios'
import SystemStatus from '@/components/SystemStatus.vue'
import { Bell, Warning, CircleCheck, CircleClose, Check } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

interface CPUInfo {
  cores: number
  logicalCores: number
  modelName: string
  mhz: number
  usedPercent: number
  perCorePercent: number[]
}

interface MemoryInfo {
  total: number
  used: number
  free: number
  available: number
  usedPercent: number
  cached: number
  buffers: number
}

interface DiskInfo {
  device: string
  mountpoint: string
  fstype: string
  total: number
  used: number
  free: number
  usedPercent: number
}

interface LoadInfo {
  load1: number
  load5: number
  load15: number
}

interface NetworkInfo {
  bytesSent: number
  bytesRecv: number
  packetsSent: number
  packetsRecv: number
}

interface CurrentInfo {
  cpuInfo: CPUInfo
  memoryInfo: MemoryInfo
  diskInfo: DiskInfo[]
  loadInfo: LoadInfo
  networkInfo: NetworkInfo
}

interface SystemInfo {
  hostname: string
  os: string
  platform: string
  platformFamily: string
  platformVersion: string
  kernelArch: string
  kernelVersion: string
  bootTime: number
  uptime: number
  procs: number
  hostAddress: string
  currentInfo: CurrentInfo
}

interface Alert {
  id: string
  title: string
  message: string
  level: 'info' | 'warning' | 'error'
  time: Date
  suggestion: string
}

const systemInfo = ref<SystemInfo | null>(null)
const currentInfo = ref<CurrentInfo | null>(null)
let refreshTimer: number | null = null

// 告警通知相关
const alerts = ref<Alert[]>([
  {
    id: '1',
    title: 'CPU使用率过高',
    message: '当前CPU使用率达到85%，请检查是否有异常进程占用大量资源。',
    level: 'warning',
    time: new Date(Date.now() - 1800000),
    suggestion: '使用 top 或 htop 命令查看进程列表，结束异常进程。'
  },
  {
    id: '2',
    title: '磁盘空间不足',
    message: '根分区磁盘使用率超过90%，建议清理不必要的文件。',
    level: 'error',
    time: new Date(Date.now() - 3600000),
    suggestion: '清理日志文件、临时文件或扩展磁盘容量。'
  },
  {
    id: '3',
    title: '系统更新可用',
    message: '检测到有新的系统更新可用，建议及时更新以获得安全补丁。',
    level: 'info',
    time: new Date(Date.now() - 7200000),
    suggestion: '使用包管理器进行系统更新。'
  }
])

const alertDetailVisible = ref(false)
const currentAlert = ref<Alert | null>(null)

const showAlertDetail = (alert: Alert) => {
  currentAlert.value = alert
  alertDetailVisible.value = true
}

const dismissAlert = () => {
  if (currentAlert.value) {
    alerts.value = alerts.value.filter(a => a.id !== currentAlert.value?.id)
    alertDetailVisible.value = false
    currentAlert.value = null
    ElMessage.success('已忽略该告警')
  }
}

const getAlertIcon = (level: string) => {
  switch (level) {
    case 'info':
      return CircleCheck
    case 'warning':
      return Warning
    case 'error':
      return CircleClose
    default:
      return Bell
  }
}

const getAlertLevelText = (level: string) => {
  switch (level) {
    case 'info':
      return '信息'
    case 'warning':
      return '警告'
    case 'error':
      return '错误'
    default:
      return '未知'
  }
}

const formatAlertTime = (time: Date) => {
  const now = new Date()
  const diff = now.getTime() - time.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (days > 0) {
    return `${days}天前`
  } else if (hours > 0) {
    return `${hours}小时前`
  } else if (minutes > 0) {
    return `${minutes}分钟前`
  } else {
    return '刚刚'
  }
}

const fetchSystemInfo = async () => {
  try {
    const response = await axios.get('/api/v1/system/info')
    systemInfo.value = response.data
    currentInfo.value = response.data.currentInfo
  } catch (error) {
    console.error('获取系统信息失败:', error)
  }
}

const fetchCurrentInfo = async () => {
  try {
    const response = await axios.get('/api/v1/system/current')
    currentInfo.value = response.data
  } catch (error) {
    console.error('获取实时信息失败:', error)
  }
}

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}

const formatBootTime = (timestamp: number): string => {
  const date = new Date(timestamp * 1000)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const formatUptime = (seconds: number): string => {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)

  return `${days}天 ${hours}小时 ${minutes}分钟 ${secs}秒`
}

onMounted(() => {
  fetchSystemInfo()
  refreshTimer = window.setInterval(fetchCurrentInfo, 3000)
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.dashboard {
  background-color: var(--bg-color);
}

.content {
  padding: 1rem;
}

.card-grid {
  display: grid;
  grid-template-columns: 7fr 3fr;
  gap: 1rem;
  align-items: start;
  grid-auto-rows: min-content;
}

.card {
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.875rem 1rem;
  border-bottom: 1px solid var(--border-color);
}

.card-icon {
  font-size: 1.1rem;
  color: var(--primary-dark);
}

.card-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-primary);
}

.card-body {
  padding: 1rem;
}

/* 告警卡片 */
.alert-count {
  margin-left: auto;
  background: var(--primary);
  color: white;
  font-size: 0.75rem;
  font-weight: 600;
  padding: 0.1rem 0.4rem;
  border-radius: 10px;
}

.alert-items {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  max-height: 300px;
  overflow-y: auto;
}

.alert-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 0.75rem;
  background: var(--bg-color);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s;
  border-left: 3px solid transparent;
}

.alert-item:hover {
  background: var(--border-color);
  transform: translateX(2px);
}

.alert-item-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  flex-shrink: 0;
}

.alert-info .alert-item-icon {
  background: #e0f2fe;
  color: #0284c7;
}

.alert-warning .alert-item-icon {
  background: #fef3c7;
  color: #d97706;
}

.alert-error .alert-item-icon {
  background: #fee2e2;
  color: #dc2626;
}

.alert-info {
  border-left-color: #0284c7;
}

.alert-warning {
  border-left-color: #d97706;
}

.alert-error {
  border-left-color: #dc2626;
}

.alert-item-content {
  flex: 1;
  min-width: 0;
}

.alert-item-title {
  font-size: 0.8rem;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 0.15rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.alert-item-time {
  font-size: 0.7rem;
  color: var(--text-secondary);
}

.alert-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1.5rem;
  color: var(--text-secondary);
  font-size: 0.85rem;
}

/* 系统信息卡片 */

.system-details {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.35rem 0;
  border-bottom: 1px solid var(--border-color);
}

.detail-row:last-child {
  border-bottom: none;
}

.detail-label {
  font-size: 0.8rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.detail-value {
  font-size: 0.8rem;
  color: var(--text-primary);
  font-weight: 600;
}

/* 仪表盘卡片 */
.loading {
  text-align: center;
  color: var(--text-secondary);
  padding: 1.5rem;
  font-size: 0.8rem;
}

/* 告警详情弹窗 */
.alert-detail {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.alert-detail-level {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  border-radius: var(--radius-md);
  font-size: 0.9rem;
  font-weight: 600;
}

.level-icon {
  font-size: 1.2rem;
}

.level-info {
  background: #e0f2fe;
  color: #0284c7;
}

.level-warning {
  background: #fef3c7;
  color: #d97706;
}

.level-error {
  background: #fee2e2;
  color: #dc2626;
}

.alert-detail-section {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.section-label {
  font-size: 0.8rem;
  color: var(--text-secondary);
  font-weight: 500;
}

.section-value {
  font-size: 0.9rem;
  color: var(--text-primary);
  line-height: 1.5;
}

@media (max-width: 1024px) {
  .card-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .content {
    padding: 0.75rem;
  }

  .card-header {
    padding: 0.75rem;
  }

  .card-body {
    padding: 0.75rem;
  }

  .alert-items {
    max-height: 200px;
  }

  .system-details {
    padding: 0;
  }

  .detail-label,
  .detail-value {
    font-size: 0.75rem;
  }
}
</style>
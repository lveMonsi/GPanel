<template>
  <div class="dashboard">
    <header class="header">
      <div class="header-content">
        <h1>系统概览</h1>
        <div class="system-info">
          <span class="info-item">🖥️ {{ systemInfo?.hostname || '加载中...' }}</span>
          <span class="info-item">⚙️ {{ systemInfo?.os || '' }} {{ systemInfo?.kernelArch || '' }}</span>
          <span class="info-item">📦 {{ systemInfo?.platformVersion || '' }}</span>
        </div>
      </div>
    </header>

    <main class="content">
      <SystemStatus
        v-if="currentInfo"
        :cpu-info="currentInfo.cpuInfo"
        :memory-info="currentInfo.memoryInfo"
        :load-info="currentInfo.loadInfo"
        :uptime="systemInfo?.uptime || 0"
      />

      <div class="info-grid">
        <div class="info-card disk-card">
          <h3>💽 磁盘信息</h3>
          <div v-if="currentInfo && currentInfo.diskInfo.length > 0" class="disk-list">
            <div v-for="disk in currentInfo.diskInfo" :key="disk.mountpoint" class="disk-item">
              <div class="disk-header">
                <span class="disk-device">{{ disk.device }}</span>
                <span class="disk-percent">{{ disk.usedPercent.toFixed(1) }}%</span>
              </div>
              <div class="disk-mount">{{ disk.mountpoint }}</div>
              <div class="disk-details">
                <span>{{ formatBytes(disk.used) }} / {{ formatBytes(disk.total) }}</span>
              </div>
              <div class="progress-bar">
                <div class="progress-fill" :style="{ width: disk.usedPercent + '%' }"></div>
              </div>
            </div>
          </div>
          <div v-else class="loading">加载中...</div>
        </div>

        <div class="info-card network-card">
          <h3>🌐 网络信息</h3>
          <div v-if="currentInfo && currentInfo.networkInfo" class="network-stats">
            <div class="network-item">
              <div class="network-icon">⬆️</div>
              <div class="network-data">
                <div class="network-label">发送</div>
                <div class="network-value">{{ formatBytes(currentInfo.networkInfo.bytesSent) }}</div>
              </div>
            </div>
            <div class="network-item">
              <div class="network-icon">⬇️</div>
              <div class="network-data">
                <div class="network-label">接收</div>
                <div class="network-value">{{ formatBytes(currentInfo.networkInfo.bytesRecv) }}</div>
              </div>
            </div>
          </div>
          <div v-else class="loading">加载中...</div>
        </div>
      </div>

      <div class="info-card processes-card">
        <h3>📋 系统详情</h3>
        <div v-if="systemInfo" class="processes-details">
          <div class="detail-row">
            <span class="detail-label">进程数:</span>
            <span class="detail-value">{{ systemInfo.procs }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">启动时间:</span>
            <span class="detail-value">{{ formatBootTime(systemInfo.bootTime) }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">内核版本:</span>
            <span class="detail-value">{{ systemInfo.kernelVersion }}</span>
          </div>
          <div class="detail-row">
            <span class="detail-label">平台:</span>
            <span class="detail-value">{{ systemInfo.platform }} ({{ systemInfo.platformFamily }})</span>
          </div>
        </div>
        <div v-else class="loading">加载中...</div>
      </div>
    </main>

    <div v-if="showConfigAlert" class="modal-overlay" @click="closeAlert">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h2>⚠️ 配置提醒</h2>
        </div>
        <div class="modal-body">
          <p>检测到您的面板配置尚未完成初始化，请前往设置页面完善配置信息。</p>
          <p class="hint">配置完成后，此提示将不再显示。</p>
        </div>
        <div class="modal-footer">
          <router-link to="/settings" class="btn btn-primary" @click="closeAlert">
            前往设置
          </router-link>
          <button class="btn btn-secondary" @click="closeAlert">
            稍后设置
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import axios from '@/utils/axios'
import SystemStatus from '@/components/SystemStatus.vue'

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
  currentInfo: CurrentInfo
}

const systemInfo = ref<SystemInfo | null>(null)
const currentInfo = ref<CurrentInfo | null>(null)
const showConfigAlert = ref(false)
let refreshTimer: number | null = null

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

const checkConfigInitialized = async () => {
  try {
    const response = await axios.get('/api/v1/config/initialized')
    if (!response.data.initialized) {
      showConfigAlert.value = true
    }
  } catch (error) {
    console.error('检查配置状态失败:', error)
  }
}

const closeAlert = () => {
  showConfigAlert.value = false
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
    minute: '2-digit'
  })
}

onMounted(() => {
  fetchSystemInfo()
  refreshTimer = window.setInterval(fetchCurrentInfo, 3000)
  checkConfigInitialized()
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.dashboard {
  background-color: #f5f5f5;
}

.header {
  background: white;
  padding: 1.5rem 2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-content {
}

.header h1 {
  font-size: 1.5rem;
  margin: 0 0 1rem 0;
  color: #333;
}

.system-info {
  display: flex;
  gap: 1.5rem;
  flex-wrap: wrap;
}

.info-item {
  font-size: 0.85rem;
  color: #666;
}

.content {
  padding: 1.5rem;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1rem;
  margin-bottom: 1rem;
}

.info-card {
  background: white;
  border-radius: 8px;
  padding: 1rem;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
}

.info-card h3 {
  margin: 0 0 0.75rem 0;
  font-size: 1.1rem;
  color: #333;
}

.disk-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.disk-item {
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
}

.disk-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.disk-device {
  font-weight: 600;
  color: #333;
}

.disk-percent {
  font-weight: 600;
  color: #667eea;
}

.disk-mount {
  font-size: 0.85rem;
  color: #666;
  margin-bottom: 0.5rem;
}

.disk-details {
  font-size: 0.9rem;
  color: #666;
  margin-bottom: 0.75rem;
}

.progress-bar {
  height: 6px;
  background: #e0e0e0;
  border-radius: 3px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  transition: width 0.3s ease;
}

.network-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.network-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: #f8f9fa;
  border-radius: 8px;
}

.network-icon {
  font-size: 2rem;
}

.network-data {
  flex: 1;
}

.network-label {
  font-size: 0.85rem;
  color: #666;
  margin-bottom: 0.25rem;
}

.network-value {
  font-size: 1.25rem;
  font-weight: 600;
  color: #667eea;
}

.processes-details {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 0.75rem;
  background: #f8f9fa;
  border-radius: 6px;
}

.detail-label {
  color: #666;
  font-weight: 500;
}

.detail-value {
  color: #333;
  font-weight: 600;
}

.loading {
  text-align: center;
  color: #999;
  padding: 2rem;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 12px;
  max-width: 500px;
  width: 90%;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  animation: slideIn 0.3s ease;
}

@keyframes slideIn {
  from {
    transform: translateY(-20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.modal-header {
  padding: 1.5rem;
  border-bottom: 1px solid #eee;
}

.modal-header h2 {
  margin: 0;
  font-size: 1.25rem;
  color: #f57c00;
}

.modal-body {
  padding: 1.5rem;
}

.modal-body p {
  margin: 0 0 1rem 0;
  color: #333;
  line-height: 1.6;
}

.modal-body .hint {
  font-size: 0.9rem;
  color: #666;
  font-style: italic;
}

.modal-footer {
  padding: 1rem 1.5rem;
  border-top: 1px solid #eee;
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

.btn {
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 6px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s;
  text-decoration: none;
  display: inline-block;
}

.btn-primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.btn-primary:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
}

.btn-secondary {
  background: #f5f5f5;
  color: #666;
}

.btn-secondary:hover {
  background: #e0e0e0;
}
</style>
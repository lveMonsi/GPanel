<template>
  <div class="dashboard">
    <header class="header">
      <div class="header-content">
        <h1>GPanel 概览</h1>
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
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
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
  padding: 1.5rem 2rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.header-content {
  max-width: 1400px;
  margin: 0 auto;
}

.header h1 {
  font-size: 1.75rem;
  margin: 0 0 1rem 0;
}

.system-info {
  display: flex;
  gap: 2rem;
  flex-wrap: wrap;
}

.info-item {
  font-size: 0.9rem;
  opacity: 0.9;
}

.content {
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 1.5rem;
  margin-bottom: 1.5rem;
}

.info-card {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.info-card h3 {
  margin: 0 0 1rem 0;
  font-size: 1.2rem;
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
</style>
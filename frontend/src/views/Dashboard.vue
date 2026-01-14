<template>
  <div class="dashboard">
    <header class="header">
      <div class="header-content">
        <h1>系统概览</h1>
        <div class="system-info">
          <span class="info-item"><el-icon><Monitor /></el-icon> {{ systemInfo?.hostname || '加载中...' }}</span>
          <span class="info-item"><el-icon><Setting /></el-icon> {{ systemInfo?.os || '' }} {{ systemInfo?.kernelArch || '' }}</span>
          <span class="info-item"><el-icon><Box /></el-icon> {{ systemInfo?.platformVersion || '' }}</span>
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
          <h3><el-icon><Coin /></el-icon> 磁盘信息</h3>
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
          <h3><el-icon><Connection /></el-icon> 网络信息</h3>
          <div v-if="currentInfo && currentInfo.networkInfo" class="network-stats">
            <div class="network-item">
              <el-icon class="network-icon"><Top /></el-icon>
              <div class="network-data">
                <div class="network-label">发送</div>
                <div class="network-value">{{ formatBytes(currentInfo.networkInfo.bytesSent) }}</div>
              </div>
            </div>
            <div class="network-item">
              <el-icon class="network-icon"><Bottom /></el-icon>
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
        <h3><el-icon><Document /></el-icon> 系统详情</h3>
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
import { Monitor, Setting, Box, Coin, Connection, Top, Bottom, Document } from '@element-plus/icons-vue'

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
  background-color: var(--bg-color);
}

.header {
  background: var(--card-bg);
  padding: 1rem 1.25rem;
  box-shadow: var(--shadow-sm);
  border-bottom: 1px solid var(--border-color);
}

.header-content {
}

.header h1 {
  font-size: 1.1rem;
  margin: 0 0 0.6rem 0;
  color: var(--text-primary);
  font-weight: 600;
  letter-spacing: -0.3px;
}

.system-info {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

.info-item {
  font-size: 0.75rem;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.info-item .el-icon {
  font-size: 0.85rem;
  color: var(--primary-dark);
}

.content {
  padding: 1rem;
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 0.75rem;
  margin-bottom: 0.75rem;
}

.info-card {
  background: var(--card-bg);
  border-radius: var(--radius-md);
  padding: 0.875rem;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
}

.info-card h3 {
  margin: 0 0 0.6rem 0;
  font-size: 0.85rem;
  color: var(--text-primary);
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.info-card h3 .el-icon {
  color: var(--primary-dark);
}

.disk-list {
  display: flex;
  flex-direction: column;
  gap: 0.6rem;
}

.disk-item {
  padding: 0.6rem;
  background: var(--bg-color);
  border-radius: var(--radius-sm);
}

.disk-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.35rem;
}

.disk-device {
  font-weight: 500;
  color: var(--text-primary);
  font-size: 0.8rem;
}

.disk-percent {
  font-weight: 600;
  color: var(--primary-dark);
  font-size: 0.8rem;
}

.disk-mount {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-bottom: 0.35rem;
}

.disk-details {
  font-size: 0.8rem;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
}

.progress-bar {
  height: 3px;
  background: var(--border-color);
  border-radius: 1.5px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--primary) 0%, var(--primary-dark) 100%);
  transition: width 0.3s ease;
}

.network-stats {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.5rem;
}

.network-item {
  display: flex;
  align-items: center;
  gap: 0.6rem;
  padding: 0.6rem;
  background: var(--bg-color);
  border-radius: var(--radius-sm);
}

.network-icon {
  font-size: 1.2rem;
  color: var(--primary-dark);
}

.network-data {
  flex: 1;
}

.network-label {
  font-size: 0.7rem;
  color: var(--text-secondary);
  margin-bottom: 0.15rem;
}

.network-value {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--primary-dark);
}

.processes-details {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem;
  background: var(--bg-color);
  border-radius: var(--radius-sm);
}

.detail-label {
  color: var(--text-secondary);
  font-weight: 500;
  font-size: 0.8rem;
}

.detail-value {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 0.8rem;
}

.loading {
  text-align: center;
  color: var(--text-secondary);
  padding: 1.5rem;
  font-size: 0.8rem;
}
</style>
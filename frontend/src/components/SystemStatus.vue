<template>
  <div class="system-status">
    <div class="status-grid">
      <div class="status-card cpu-card">
        <div class="card-header">
          <el-icon class="icon"><Monitor /></el-icon>
          <span class="title">CPU</span>
        </div>
        <div class="card-content">
          <div class="value">{{ cpuInfo.usedPercent.toFixed(1) }}%</div>
          <div class="label">使用率</div>
          <div class="details">
            <div class="detail-item">
              <span class="detail-label">核心数:</span>
              <span class="detail-value">{{ cpuInfo.cores }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">逻辑核心:</span>
              <span class="detail-value">{{ cpuInfo.logicalCores }}</span>
            </div>
          </div>
        </div>
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: cpuInfo.usedPercent + '%' }"></div>
        </div>
      </div>

      <div class="status-card memory-card">
        <div class="card-header">
          <el-icon class="icon"><Cpu /></el-icon>
          <span class="title">内存</span>
        </div>
        <div class="card-content">
          <div class="value">{{ memoryInfo.usedPercent.toFixed(1) }}%</div>
          <div class="label">使用率</div>
          <div class="details">
            <div class="detail-item">
              <span class="detail-label">已用:</span>
              <span class="detail-value">{{ formatBytes(memoryInfo.used) }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">总计:</span>
              <span class="detail-value">{{ formatBytes(memoryInfo.total) }}</span>
            </div>
          </div>
        </div>
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: memoryInfo.usedPercent + '%' }"></div>
        </div>
      </div>

      <div class="status-card load-card">
        <div class="card-header">
          <el-icon class="icon"><TrendCharts /></el-icon>
          <span class="title">负载</span>
        </div>
        <div class="card-content">
          <div class="load-values">
            <div class="load-item">
              <span class="load-label">1分钟</span>
              <span class="load-value">{{ loadInfo.load1.toFixed(2) }}</span>
            </div>
            <div class="load-item">
              <span class="load-label">5分钟</span>
              <span class="load-value">{{ loadInfo.load5.toFixed(2) }}</span>
            </div>
            <div class="load-item">
              <span class="load-label">15分钟</span>
              <span class="load-value">{{ loadInfo.load15.toFixed(2) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="status-card uptime-card">
        <div class="card-header">
          <el-icon class="icon"><Timer /></el-icon>
          <span class="title">运行时间</span>
        </div>
        <div class="card-content">
          <div class="uptime-value">{{ formatUptime(uptime) }}</div>
          <div class="uptime-label">系统运行时间</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Monitor, Cpu, TrendCharts, Timer } from '@element-plus/icons-vue'

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

interface LoadInfo {
  load1: number
  load5: number
  load15: number
}

defineProps<{
  cpuInfo: CPUInfo
  memoryInfo: MemoryInfo
  loadInfo: LoadInfo
  uptime: number
}>()

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}

const formatUptime = (seconds: number): string => {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  
  if (days > 0) {
    return `${days}天 ${hours}小时 ${minutes}分钟`
  } else if (hours > 0) {
    return `${hours}小时 ${minutes}分钟`
  } else {
    return `${minutes}分钟`
  }
}
</script>

<style scoped>
.system-status {
  margin-bottom: 1.25rem;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 0.75rem;
}

.status-card {
  background: var(--card-bg);
  border-radius: var(--radius-md);
  padding: 1rem;
  box-shadow: var(--shadow-sm);
  transition: all 0.2s;
  border: 1px solid var(--border-color);
}

.status-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
}

.card-header .icon {
  font-size: 1.1rem;
  color: var(--primary-dark);
}

.card-header .title {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-primary);
}

.card-content {
  margin-bottom: 0.75rem;
}

.value {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--primary-dark);
  margin-bottom: 0.25rem;
  letter-spacing: -0.5px;
}

.label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
}

.details {
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
}

.detail-label {
  color: var(--text-secondary);
}

.detail-value {
  color: var(--text-primary);
  font-weight: 500;
}

.progress-bar {
  height: 4px;
  background: var(--border-color);
  border-radius: 2px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--primary) 0%, var(--primary-dark) 100%);
  transition: width 0.3s ease;
}

.load-values {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.load-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.load-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.load-value {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--primary-dark);
}

.uptime-value {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--primary-dark);
  margin-bottom: 0.25rem;
}

.uptime-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
}
</style>
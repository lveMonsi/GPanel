<template>
  <div class="system-status">
    <div class="status-grid">
      <div class="status-card cpu-card">
        <div class="card-header">
          <span class="icon">🖥️</span>
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
          <span class="icon">💾</span>
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
          <span class="icon">📊</span>
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
          <span class="icon">⏱️</span>
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
  margin-bottom: 2rem;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
}

.status-card {
  background: white;
  border-radius: 12px;
  padding: 1.5rem;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transition: transform 0.2s, box-shadow 0.2s;
}

.status-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.card-header .icon {
  font-size: 1.5rem;
}

.card-header .title {
  font-size: 1.1rem;
  font-weight: 600;
  color: #333;
}

.card-content {
  margin-bottom: 1rem;
}

.value {
  font-size: 2rem;
  font-weight: bold;
  color: #667eea;
  margin-bottom: 0.25rem;
}

.label {
  font-size: 0.9rem;
  color: #666;
  margin-bottom: 0.75rem;
}

.details {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
}

.detail-label {
  color: #999;
}

.detail-value {
  color: #333;
  font-weight: 500;
}

.progress-bar {
  height: 8px;
  background: #f0f0f0;
  border-radius: 4px;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, #667eea 0%, #764ba2 100%);
  transition: width 0.3s ease;
}

.load-values {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.load-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.load-label {
  font-size: 0.9rem;
  color: #666;
}

.load-value {
  font-size: 1.2rem;
  font-weight: 600;
  color: #667eea;
}

.uptime-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: #667eea;
  margin-bottom: 0.25rem;
}

.uptime-label {
  font-size: 0.9rem;
  color: #666;
}
</style>
<template>
  <div class="dashboard">
    <main class="content">
      <!-- 初始加载状态 -->
      <div v-if="initialLoading" class="loading-container">
        <el-icon class="loading-icon" :size="48"><Loading /></el-icon>
        <p class="loading-text">正在加载系统信息...</p>
      </div>

      <template v-else>
      <div class="card-grid">
        <!-- 左侧列：快捷概览 + 系统监控 -->
        <div class="card-grid-left">
          <!-- 快捷概览卡片 -->
          <div class="card overview-card">
            <div class="card-header">
              <el-icon class="card-icon"><DataAnalysis /></el-icon>
              <span class="card-title">快捷概览</span>
            </div>
            <div class="card-body">
              <div class="stat-grid">
                <div class="stat-item">
                  <div class="stat-icon" style="background: #e0f2fe; color: #0284c7;">
                    <el-icon><Monitor /></el-icon>
                  </div>
                  <div class="stat-content">
                    <div class="stat-value">{{ overviewStats.hostCount }}</div>
                    <div class="stat-label">主机</div>
                  </div>
                </div>
                <div class="stat-item">
                  <div class="stat-icon" style="background: #fef3c7; color: #d97706;">
                    <el-icon><Timer /></el-icon>
                  </div>
                  <div class="stat-content">
                    <div class="stat-value">{{ overviewStats.cronjobCount }}</div>
                    <div class="stat-label">计划任务</div>
                  </div>
                </div>
                <div class="stat-item">
                  <div class="stat-icon" style="background: #dcfce7; color: #16a34a;">
                    <el-icon><Document /></el-icon>
                  </div>
                  <div class="stat-content">
                    <div class="stat-value">{{ overviewStats.todayOps }}</div>
                    <div class="stat-label">今日操作</div>
                  </div>
                </div>
                <div class="stat-item">
                  <div class="stat-icon" style="background: #f3e8ff; color: #9333ea;">
                    <el-icon><Cpu /></el-icon>
                  </div>
                  <div class="stat-content">
                    <div class="stat-value">{{ systemInfo?.procs ?? '-' }}</div>
                    <div class="stat-label">运行进程</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- 系统监控卡片 -->
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
                :swap-info="currentInfo.swapInfo"
                :load-info="currentInfo.loadInfo"
                :disk-info="currentInfo.diskInfo"
              />
            </div>
          </div>
        </div>

        <!-- 右侧列：系统信息 + 最近操作 -->
        <div class="card-grid-right">
          <!-- 系统信息卡片 -->
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

          <!-- 最近操作卡片 -->
          <div class="card recent-card">
            <div class="card-header">
              <el-icon class="card-icon"><List /></el-icon>
              <span class="card-title">最近操作</span>
            </div>
            <div class="card-body">
              <div class="recent-list">
                <div
                  v-for="log in recentLogs"
                  :key="log.id"
                  class="recent-item"
                >
                  <div class="recent-item-dot" :class="log.status === 'success' ? 'dot-success' : 'dot-fail'" />
                  <div class="recent-item-content">
                    <div class="recent-item-title">
                      <span class="recent-resource">{{ log.resource }}</span>
                      <span class="recent-action">{{ log.action }}</span>
                    </div>
                    <div class="recent-item-meta">
                      <span>{{ log.username }}</span>
                      <span>{{ formatLogTime(log.createdAt) }}</span>
                    </div>
                  </div>
                </div>
                <div v-if="recentLogs.length === 0" class="recent-empty">
                  暂无操作记录
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      </template>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import axios from '@/utils/axios'
import SystemStatus from '@/components/SystemStatus.vue'
import { Loading, DataAnalysis, Timer, Document, Cpu, List, Monitor, TrendCharts } from '@element-plus/icons-vue'
import { getOperationLogStats, searchOperationLogs } from '@/api/modules/log'
import { listHosts } from '@/api/modules/host'
import { searchCronjobs } from '@/api/modules/cronjob'

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

interface SwapInfo {
  total: number
  used: number
  free: number
  usedPercent: number
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
  swapInfo: SwapInfo
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

interface RecentLog {
  id: number
  createdAt: string
  username: string
  resource: string
  action: string
  status: string
}

const systemInfo = ref<SystemInfo | null>(null)
const currentInfo = ref<CurrentInfo | null>(null)
const initialLoading = ref(true)
let refreshTimer: number | null = null

const overviewStats = ref({
  hostCount: 0,
  cronjobCount: 0,
  todayOps: 0,
})

const recentLogs = ref<RecentLog[]>([])

const fetchSystemInfo = async () => {
  try {
    const response = await axios.get('/api/v1/system/info')
    systemInfo.value = response.data.data
    currentInfo.value = response.data.data?.currentInfo
  } catch (error) {
    console.error('获取系统信息失败:', error)
  } finally {
    initialLoading.value = false
  }
}

const fetchCurrentInfo = async () => {
  try {
    const response = await axios.get('/api/v1/system/current')
    currentInfo.value = response.data.data
  } catch (error) {
    console.error('获取实时信息失败:', error)
  }
}

const fetchOverviewStats = async () => {
  try {
    const [hostRes, cronjobRes, logStatsRes] = await Promise.all([
      listHosts({ page: 1, pageSize: 1 }),
      searchCronjobs({ page: 1, pageSize: 1 }),
      getOperationLogStats(),
    ])
    overviewStats.value.hostCount = hostRes.data?.total ?? 0
    overviewStats.value.cronjobCount = cronjobRes.data?.data?.total ?? 0
    overviewStats.value.todayOps = logStatsRes.data?.data?.todayCount ?? 0
  } catch (error) {
    console.error('获取概览统计失败:', error)
  }
}

const fetchRecentLogs = async () => {
  try {
    const res = await searchOperationLogs({ page: 1, pageSize: 5 })
    recentLogs.value = res.data?.data?.items ?? []
  } catch (error) {
    console.error('获取最近操作失败:', error)
  }
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

const formatLogTime = (timeStr: string): string => {
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (days > 0) return `${days}天前`
  if (hours > 0) return `${hours}小时前`
  if (minutes > 0) return `${minutes}分钟前`
  return '刚刚'
}

onMounted(() => {
  fetchSystemInfo()
  fetchOverviewStats()
  fetchRecentLogs()
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
  display: flex;
  gap: 1rem;
  align-items: start;
}

.card-grid-left {
  flex: 7;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.card-grid-right {
  flex: 3;
  display: flex;
  flex-direction: column;
  gap: 1rem;
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

/* 快捷概览 */
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.75rem;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--bg-color);
  border-radius: var(--radius-sm);
}

.stat-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  flex-shrink: 0;
  font-size: 1.1rem;
}

.stat-content {
  min-width: 0;
}

.stat-value {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
}

.stat-label {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-top: 0.15rem;
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

/* 最近操作卡片 */
.recent-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.recent-item {
  display: flex;
  align-items: flex-start;
  gap: 0.6rem;
  padding: 0.5rem 0;
  border-bottom: 1px solid var(--border-color);
}

.recent-item:last-child {
  border-bottom: none;
}

.recent-item-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: 0.35rem;
}

.dot-success {
  background: #16a34a;
}

.dot-fail {
  background: #dc2626;
}

.recent-item-content {
  flex: 1;
  min-width: 0;
}

.recent-item-title {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8rem;
  color: var(--text-primary);
}

.recent-resource {
  font-weight: 600;
}

.recent-action {
  color: var(--text-secondary);
}

.recent-item-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.7rem;
  color: var(--text-secondary);
  margin-top: 0.2rem;
}

.recent-empty {
  text-align: center;
  color: var(--text-secondary);
  font-size: 0.8rem;
  padding: 1.5rem;
}

/* 仪表盘卡片 */
.loading {
  text-align: center;
  color: var(--text-secondary);
  padding: 1.5rem;
  font-size: 0.8rem;
}

@media (max-width: 1024px) {
  .card-grid {
    flex-direction: column;
  }

  .card-grid-left,
  .card-grid-right {
    flex: 1;
  }

  .stat-grid {
    grid-template-columns: repeat(2, 1fr);
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

  .stat-grid {
    grid-template-columns: 1fr 1fr;
  }

  .system-details {
    padding: 0;
  }

  .detail-label,
  .detail-value {
    font-size: 0.75rem;
  }
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: calc(100vh - 200px);
  min-height: 300px;
}

.loading-icon {
  animation: spin 1s linear infinite;
  color: #409eff;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.loading-text {
  margin-top: 16px;
  font-size: 14px;
  color: #606266;
}
</style>

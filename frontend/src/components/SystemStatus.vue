<template>
  <div class="system-status">
    <div class="gauge-grid">
      <div class="gauge-card" ref="cpuCard" @mouseenter="(e) => showTooltip('cpu', e)" @mouseleave="hideTooltip">
        <div class="gauge-header">
          <el-icon class="icon"><Monitor /></el-icon>
          <span class="title">CPU</span>
        </div>
        <GaugeChart
          :value="cpuInfo.usedPercent"
          label="使用率"
          :sub-label="`核心: ${cpuInfo.cores}`"
          unit="%"
        />
        <!-- CPU 悬浮详情 -->
        <div v-if="tooltipType === 'cpu'" :class="['tooltip', { 'tooltip-bottom': tooltipDirection === 'bottom' }]">
          <div class="tooltip-header">
            <el-icon><Monitor /></el-icon>
            <span>CPU 详情</span>
          </div>
          <div class="tooltip-content">
            <div class="cpu-columns">
              <div class="tooltip-section">
                <div class="section-title">{{ cpuInfo.modelName }}</div>
                <div class="info-row">
                  <span class="info-label">物理核心</span>
                  <span class="info-value">{{ cpuInfo.cores }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">逻辑核心</span>
                  <span class="info-value">{{ cpuInfo.logicalCores }}</span>
                </div>
              </div>
              <div class="tooltip-section">
                <div class="section-title">核心使用率</div>
                <div v-for="(percent, index) in cpuInfo.perCorePercent" :key="index" class="info-row">
                  <span class="info-label">CPU-{{ index }}</span>
                  <span class="info-value">{{ percent.toFixed(0) }}%</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="gauge-card" ref="memoryCard" @mouseenter="(e) => showTooltip('memory', e)" @mouseleave="hideTooltip">
        <div class="gauge-header">
          <el-icon class="icon"><Cpu /></el-icon>
          <span class="title">内存</span>
        </div>
        <GaugeChart
          :value="memoryInfo.usedPercent"
          label="使用率"
          :sub-label="`${formatBytes(memoryInfo.used)} / ${formatBytes(memoryInfo.total)}`"
          unit="%"
        />
        <!-- 内存悬浮详情 -->
        <div v-if="tooltipType === 'memory'" :class="['tooltip', { 'tooltip-bottom': tooltipDirection === 'bottom' }]">
          <div class="tooltip-header">
            <el-icon><Cpu /></el-icon>
            <span>内存详情</span>
          </div>
          <div class="tooltip-content">
            <div class="memory-columns">
              <div class="tooltip-section">
                <div class="section-title">系统内存</div>
                <div class="info-row">
                  <span class="info-label">总数</span>
                  <span class="info-value">{{ formatBytes(memoryInfo.total) }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">已用</span>
                  <span class="info-value">{{ formatBytes(memoryInfo.used) }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">可用</span>
                  <span class="info-value">{{ formatBytes(memoryInfo.available) }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">使用率</span>
                  <span class="info-value">{{ memoryInfo.usedPercent.toFixed(2) }}%</span>
                </div>
              </div>
              <div class="tooltip-section">
                <div class="section-title">SWAP 分区</div>
                <div class="info-row">
                  <span class="info-label">总数</span>
                  <span class="info-value">{{ formatBytes(swapInfo.total) }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">已用</span>
                  <span class="info-value">{{ formatBytes(swapInfo.used) }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">可用</span>
                  <span class="info-value">{{ formatBytes(swapInfo.free) }}</span>
                </div>
                <div class="info-row">
                  <span class="info-label">使用率</span>
                  <span class="info-value">{{ swapInfo.usedPercent.toFixed(2) }}%</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="gauge-card" ref="loadCard" @mouseenter="(e) => showTooltip('load', e)" @mouseleave="hideTooltip">
        <div class="gauge-header">
          <el-icon class="icon"><TrendCharts /></el-icon>
          <span class="title">负载</span>
        </div>
        <GaugeChart
          :value="loadInfo.load1"
          :max-value="cpuInfo.logicalCores"
          label="1分钟负载"
          :sub-label="`5分钟: ${loadInfo.load5.toFixed(2)}`"
        />
        <!-- 负载悬浮详情 -->
        <div v-if="tooltipType === 'load'" :class="['tooltip', { 'tooltip-bottom': tooltipDirection === 'bottom' }]">
          <div class="tooltip-header">
            <el-icon><TrendCharts /></el-icon>
            <span>负载详情</span>
          </div>
          <div class="tooltip-content">
            <div class="info-row">
              <span class="info-label">最近 1 分钟平均负载</span>
              <span class="info-value">{{ loadInfo.load1.toFixed(2) }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">最近 5 分钟平均负载</span>
              <span class="info-value">{{ loadInfo.load5.toFixed(2) }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">最近 15 分钟平均负载</span>
              <span class="info-value">{{ loadInfo.load15.toFixed(2) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="gauge-card" ref="diskCard" @mouseenter="(e) => showTooltip('disk', e)" @mouseleave="hideTooltip">
        <div class="gauge-header">
          <el-icon class="icon"><Coin /></el-icon>
          <span class="title">存储</span>
        </div>
        <GaugeChart
          :value="diskInfo.length > 0 ? diskInfo[0].usedPercent : 0"
          label="主盘使用率"
          :sub-label="diskInfo.length > 0 ? `${formatBytes(diskInfo[0].used)} / ${formatBytes(diskInfo[0].total)}` : '无数据'"
          unit="%"
        />
        <!-- 存储悬浮详情 -->
        <div v-if="tooltipType === 'disk'" :class="['tooltip', { 'tooltip-bottom': tooltipDirection === 'bottom' }]">
          <div class="tooltip-header">
            <el-icon><Coin /></el-icon>
            <span>存储详情</span>
          </div>
          <div class="tooltip-content">
            <template v-if="diskInfo.length > 0">
              <div v-for="disk in diskInfo" :key="disk.mountpoint" class="disk-section">
                <!-- 基本信息单列 -->
                <div class="tooltip-section disk-base-info">
                  <div class="section-title">基本信息</div>
                  <div class="info-row">
                    <span class="info-label">挂载点</span>
                    <span class="info-value">{{ disk.mountpoint }}</span>
                  </div>
                  <div class="info-row">
                    <span class="info-label">类型</span>
                    <span class="info-value">{{ disk.fstype }}</span>
                  </div>
                  <div class="info-row">
                    <span class="info-label">文件系统</span>
                    <span class="info-value">{{ disk.device }}</span>
                  </div>
                </div>
                <!-- Inode 和磁盘双列 -->
                <div class="disk-columns">
                  <div class="tooltip-section">
                    <div class="section-title">Inode</div>
                    <div class="info-row">
                      <span class="info-label">总数</span>
                      <span class="info-value">{{ formatNumber(disk.inodesTotal) }}</span>
                    </div>
                    <div class="info-row">
                      <span class="info-label">已用</span>
                      <span class="info-value">{{ formatNumber(disk.inodesUsed) }}</span>
                    </div>
                    <div class="info-row">
                      <span class="info-label">可用</span>
                      <span class="info-value">{{ formatNumber(disk.inodesFree) }}</span>
                    </div>
                    <div class="info-row">
                      <span class="info-label">使用率</span>
                      <span class="info-value">{{ disk.inodesUsedPercent.toFixed(2) }}%</span>
                    </div>
                  </div>
                  <div class="tooltip-section">
                    <div class="section-title">磁盘</div>
                    <div class="info-row">
                      <span class="info-label">总数</span>
                      <span class="info-value">{{ formatBytes(disk.total) }}</span>
                    </div>
                    <div class="info-row">
                      <span class="info-label">已用</span>
                      <span class="info-value">{{ formatBytes(disk.used) }}</span>
                    </div>
                    <div class="info-row">
                      <span class="info-label">可用</span>
                      <span class="info-value">{{ formatBytes(disk.free) }}</span>
                    </div>
                    <div class="info-row">
                      <span class="info-label">使用率</span>
                      <span class="info-value">{{ disk.usedPercent.toFixed(2) }}%</span>
                    </div>
                  </div>
                </div>
              </div>
            </template>
            <div v-else class="no-data">无数据</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { Monitor, Cpu, TrendCharts, Coin } from '@element-plus/icons-vue'
import GaugeChart from './GaugeChart.vue'

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

interface LoadInfo {
  load1: number
  load5: number
  load15: number
}

interface DiskInfo {
  device: string
  mountpoint: string
  fstype: string
  total: number
  used: number
  free: number
  usedPercent: number
  inodesTotal: number
  inodesUsed: number
  inodesFree: number
  inodesUsedPercent: number
}

defineProps<{
  cpuInfo: CPUInfo
  memoryInfo: MemoryInfo
  swapInfo: SwapInfo
  loadInfo: LoadInfo
  diskInfo: DiskInfo[]
}>()

const tooltipType = ref<'cpu' | 'memory' | 'load' | 'disk' | null>(null)
const tooltipDirection = ref<'top' | 'bottom'>('top')

const showTooltip = (type: 'cpu' | 'memory' | 'load' | 'disk', event: MouseEvent) => {
  tooltipType.value = type

  // 获取目标元素的位置
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  const viewportHeight = window.innerHeight
  const elementCenter = rect.top + rect.height / 2

  // 如果元素在视口上半部分，向下弹出；否则向上弹出
  tooltipDirection.value = elementCenter < viewportHeight / 2 ? 'bottom' : 'top'
}

const hideTooltip = () => {
  tooltipType.value = null
}

const formatBytes = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i]
}

const formatNumber = (num: number): string => {
  if (num === 0) return '0'
  return num.toLocaleString('zh-CN')
}
</script>

<style scoped>
.system-status {
  position: relative;
}

.gauge-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}

.gauge-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  cursor: pointer;
  position: relative;
  padding: 0.5rem;
}

.gauge-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.gauge-header .icon {
  font-size: 1rem;
  color: var(--primary-dark);
}

.gauge-header .title {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--text-primary);
}

.tooltip {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  width: calc(100% + 2rem);
  min-width: 220px;
  max-width: 280px;
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-lg);
  z-index: 100;
  padding: 1rem;
  animation: fadeIn 0.2s ease;
}

.tooltip:not(.tooltip-bottom) {
  bottom: calc(100% + 10px);
}

.tooltip:not(.tooltip-bottom)::after {
  content: '';
  position: absolute;
  top: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 8px solid transparent;
  border-top-color: var(--card-bg);
}

.tooltip:not(.tooltip-bottom)::before {
  content: '';
  position: absolute;
  top: calc(100% + 1px);
  left: 50%;
  transform: translateX(-50%);
  border: 8px solid transparent;
  border-top-color: var(--border-color);
  z-index: -1;
}

.tooltip.tooltip-bottom {
  top: calc(100% + 10px);
}

.tooltip.tooltip-bottom::after {
  content: '';
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  border: 8px solid transparent;
  border-bottom-color: var(--card-bg);
}

.tooltip.tooltip-bottom::before {
  content: '';
  position: absolute;
  bottom: calc(100% + 1px);
  left: 50%;
  transform: translateX(-50%);
  border: 8px solid transparent;
  border-bottom-color: var(--border-color);
  z-index: -1;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.tooltip-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 0.75rem;
  font-size: 0.95rem;
  font-weight: 600;
  color: var(--text-primary);
}

.tooltip-header .el-icon {
  color: var(--primary-dark);
}

.tooltip-content {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.tooltip-section {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.section-title {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--primary-dark);
  margin-bottom: 0.25rem;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.8rem;
}

.info-label {
  color: var(--text-secondary);
  font-weight: 500;
}

.info-value {
  color: var(--text-primary);
  font-weight: 600;
}

.cpu-columns,
.disk-columns,
.memory-columns {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.disk-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--border-color);
}

.disk-base-info {
  width: 100%;
}

.disk-section:last-child {
  padding-bottom: 0;
  border-bottom: none;
}

.no-data {
  text-align: center;
  color: var(--text-secondary);
  font-size: 0.85rem;
  padding: 1rem;
}

@media (max-width: 1024px) {
  .gauge-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .tooltip {
    width: calc(100% + 1.5rem);
    min-width: 200px;
  }
}

@media (max-width: 640px) {
  .gauge-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.75rem;
  }

  .tooltip {
    position: fixed;
    bottom: auto;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    width: 90%;
    max-width: 90vw;
    max-height: 80vh;
    overflow-y: auto;
  }

  .tooltip::after,
  .tooltip::before {
    display: none;
  }

  .cpu-columns,
  .disk-columns,
  .memory-columns {
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }
}
</style>
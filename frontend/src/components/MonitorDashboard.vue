<template>
  <div class="monitor-dashboard">
    <el-card class="global-time-card">
      <div class="time-controls">
        <el-date-picker
          v-model="globalTimeRange"
          type="datetimerange"
          range-separator="-"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          :shortcuts="timeShortcuts"
          @change="applyGlobalTime"
        />
        <el-button @click="refreshAll" :icon="Refresh">刷新</el-button>
      </div>
    </el-card>

    <el-card class="chart-card full-width">
      <template #header>
        <div class="card-header">
          <span class="chart-title">平均负载</span>
          <el-date-picker
            v-model="loadTimeRange"
            type="datetimerange"
            range-separator="-"
            start-placeholder="开始时间"
            end-placeholder="结束时间"
            :shortcuts="timeShortcuts"
            @change="() => fetchChartData('load')"
          />
        </div>
      </template>
      <div ref="loadChartRef" class="chart-container"></div>
    </el-card>

    <div class="chart-row">
      <el-card class="chart-card">
        <template #header>
          <div class="card-header">
            <span class="chart-title">CPU</span>
            <el-date-picker
              v-model="cpuTimeRange"
              type="datetimerange"
              range-separator="-"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              :shortcuts="timeShortcuts"
              @change="() => fetchChartData('cpu')"
            />
          </div>
        </template>
        <div ref="cpuChartRef" class="chart-container"></div>
      </el-card>

      <el-card class="chart-card">
        <template #header>
          <div class="card-header">
            <span class="chart-title">内存</span>
            <el-date-picker
              v-model="memTimeRange"
              type="datetimerange"
              range-separator="-"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              :shortcuts="timeShortcuts"
              @change="() => fetchChartData('memory')"
            />
          </div>
        </template>
        <div ref="memChartRef" class="chart-container"></div>
      </el-card>
    </div>

    <div class="chart-row">
      <el-card class="chart-card">
        <template #header>
          <div class="card-header">
            <div class="title-with-select">
              <span class="chart-title">磁盘 I/O：</span>
              <el-select v-model="selectedDisk" @change="() => fetchChartData('io')" size="small">
                <el-option label="全部" value="all" />
                <el-option v-for="disk in diskOptions" :key="disk" :label="disk" :value="disk" />
              </el-select>
            </div>
            <el-date-picker
              v-model="ioTimeRange"
              type="datetimerange"
              range-separator="-"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              :shortcuts="timeShortcuts"
              @change="() => fetchChartData('io')"
            />
          </div>
        </template>
        <div ref="ioChartRef" class="chart-container"></div>
      </el-card>

      <el-card class="chart-card">
        <template #header>
          <div class="card-header">
            <div class="title-with-select">
              <span class="chart-title">网络：</span>
              <el-select v-model="selectedNetwork" @change="() => fetchChartData('network')" size="small">
                <el-option label="全部" value="all" />
                <el-option v-for="net in networkOptions" :key="net" :label="net" :value="net" />
              </el-select>
            </div>
            <el-date-picker
              v-model="netTimeRange"
              type="datetimerange"
              range-separator="-"
              start-placeholder="开始时间"
              end-placeholder="结束时间"
              :shortcuts="timeShortcuts"
              @change="() => fetchChartData('network')"
            />
          </div>
        </template>
        <div ref="netChartRef" class="chart-container"></div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import axios from '@/utils/axios'
import * as echarts from 'echarts'
import { Refresh } from '@element-plus/icons-vue'

type MonitorSetting = {
  defaultIO?: string
  defaultNetwork?: string
}

const monitorSettings = ref<MonitorSetting>({
  defaultIO: 'all',
  defaultNetwork: 'all'
})

const getTodayRange = () => [new Date(new Date().setHours(0, 0, 0, 0)), new Date()]

const globalTimeRange = ref<[Date, Date]>(getTodayRange())
const loadTimeRange = ref<[Date, Date]>(getTodayRange())
const cpuTimeRange = ref<[Date, Date]>(getTodayRange())
const memTimeRange = ref<[Date, Date]>(getTodayRange())
const ioTimeRange = ref<[Date, Date]>(getTodayRange())
const netTimeRange = ref<[Date, Date]>(getTodayRange())

const selectedDisk = ref('all')
const selectedNetwork = ref('all')
const diskOptions = ref<string[]>([])
const networkOptions = ref<string[]>([])

const loadChartRef = ref<HTMLElement>()
const cpuChartRef = ref<HTMLElement>()
const memChartRef = ref<HTMLElement>()
const ioChartRef = ref<HTMLElement>()
const netChartRef = ref<HTMLElement>()

let chartInstances: Record<string, echarts.ECharts> = {}

const normalizeDevice = (value?: string) => value && value.trim() ? value : 'all'

const pickDevice = (preferred: string, options: string[]) => {
  if (preferred === 'all') return 'all'
  return options.includes(preferred) ? preferred : 'all'
}

const timeShortcuts = [
  { text: '最近1小时', value: () => [new Date(Date.now() - 3600000), new Date()] },
  { text: '最近6小时', value: () => [new Date(Date.now() - 21600000), new Date()] },
  { text: '今天', value: getTodayRange },
  { text: '最近3天', value: () => [new Date(Date.now() - 259200000), new Date()] },
  { text: '最近7天', value: () => [new Date(Date.now() - 604800000), new Date()] }
]

const applyGlobalTime = () => {
  loadTimeRange.value = globalTimeRange.value
  cpuTimeRange.value = globalTimeRange.value
  memTimeRange.value = globalTimeRange.value
  ioTimeRange.value = globalTimeRange.value
  netTimeRange.value = globalTimeRange.value
  refreshAll()
}

const refreshAll = () => {
  fetchChartData('load')
  fetchChartData('cpu')
  fetchChartData('memory')
  fetchChartData('io')
  fetchChartData('network')
}

const formatTime = (date: string) => {
  const d = new Date(date)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours()}:${String(d.getMinutes()).padStart(2, '0')}`
}

const formatBytes = (kb: number) => {
  if (kb < 1024) return `${kb.toFixed(2)} KB`
  if (kb < 1048576) return `${(kb / 1024).toFixed(2)} MB`
  return `${(kb / 1048576).toFixed(2)} GB`
}

const buildProcessTable = (processes: any[], type: 'cpu' | 'mem') => {
  if (!processes?.length) return ''
  const headers = type === 'cpu'
    ? ['PID', '用户', '进程', '占用率']
    : ['PID', '用户', '进程', '内存', '占用率']

  let html = '<div style="margin-top:10px;border-top:1px dashed #ccc;padding-top:10px">'
  html += '<table style="width:100%;font-size:12px;border-collapse:collapse">'
  html += '<tr>' + headers.map(h => `<th style="padding:4px;text-align:center">${h}</th>`).join('') + '</tr>'

  processes.forEach(p => {
    html += '<tr>'
    html += `<td style="padding:4px;text-align:center">${p.pid}</td>`
    html += `<td style="padding:4px;text-align:center">${p.user}</td>`
    html += `<td style="padding:4px;text-align:center">${p.name}</td>`
    if (type === 'mem') html += `<td style="padding:4px;text-align:center">${formatBytes(p.memory)}</td>`
    html += `<td style="padding:4px;text-align:center">${p.percent.toFixed(2)}%</td>`
    html += '</tr>'
  })
  html += '</table></div>'
  return html
}

const renderEmptyChart = (type: string, message = '暂无监控数据') => {
  const refMap: Record<string, HTMLElement | undefined> = {
    load: loadChartRef.value,
    cpu: cpuChartRef.value,
    memory: memChartRef.value,
    io: ioChartRef.value,
    network: netChartRef.value
  }

  const el = refMap[type]
  if (!el) return

  if (!chartInstances[type]) {
    chartInstances[type] = echarts.init(el)
  }

  chartInstances[type].clear()
  chartInstances[type].setOption({
    animation: false,
    xAxis: { type: 'category', show: false, data: [] },
    yAxis: { type: 'value', show: false },
    series: [],
    graphic: {
      type: 'text',
      left: 'center',
      top: 'middle',
      style: {
        text: message,
        fill: '#909399',
        fontSize: 16,
        fontWeight: 500
      }
    }
  })
}

const renderChart = (type: string, data: any[]) => {
  const refMap: Record<string, any> = {
    load: loadChartRef.value,
    cpu: cpuChartRef.value,
    memory: memChartRef.value,
    io: ioChartRef.value,
    network: netChartRef.value
  }

  const el = refMap[type]
  if (!el) return

  if (!chartInstances[type]) {
    chartInstances[type] = echarts.init(el)
  }

  if (!data.length) {
    renderEmptyChart(type)
    return
  }

  const times = data.map(d => formatTime(d.date))

  if (type === 'load') {
    chartInstances[type].setOption({
      tooltip: {
        trigger: 'axis',
        formatter: (params: any) => {
          let html = `<div style="padding:5px"><b>${params[0].name}</b><br/>`
          params.forEach((p: any) => {
            const val = p.data?.value ?? p.data
            html += `${p.marker} ${p.seriesName}: ${val}<br/>`
          })
          if (params[3]?.data?.topCPU) html += buildProcessTable(params[3].data.topCPU, 'cpu')
          return html + '</div>'
        }
      },
      legend: { data: ['1分钟', '5分钟', '15分钟', '资源使用率'] },
      grid: { left: '10%', right: '10%', bottom: '15%' },
      xAxis: { type: 'category', data: times },
      yAxis: [
        { type: 'value', name: '负载详情' },
        { type: 'value', name: '资源使用率 (%)', position: 'right' }
      ],
      series: [
        { name: '1分钟', type: 'line', data: data.map(d => d.load1?.toFixed(2) || 0) },
        { name: '5分钟', type: 'line', data: data.map(d => d.load5?.toFixed(2) || 0) },
        { name: '15分钟', type: 'line', data: data.map(d => d.load15?.toFixed(2) || 0) },
        {
          name: '资源使用率',
          type: 'line',
          yAxisIndex: 1,
          data: data.map(d => ({ value: d.loadUsage?.toFixed(2) || 0, topCPU: d.topCPU }))
        }
      ]
    })
  } else if (type === 'cpu') {
    chartInstances[type].setOption({
      tooltip: {
        trigger: 'axis',
        formatter: (params: any) => {
          let html = `<div style="padding:5px"><b>${params[0].name}</b><br/>`
          html += `${params[0].marker} CPU: ${params[0].data.value}%<br/>`
          if (params[0].data.topCPU) html += buildProcessTable(params[0].data.topCPU, 'cpu')
          return html + '</div>'
        }
      },
      grid: { left: '10%', right: '5%', bottom: '15%' },
      xAxis: { type: 'category', data: times },
      yAxis: { type: 'value', name: 'CPU (%)', max: 100 },
      series: [{
        name: 'CPU',
        type: 'line',
        smooth: true,
        data: data.map(d => ({ value: d.cpu?.toFixed(2) || 0, topCPU: d.topCPU }))
      }]
    })
  } else if (type === 'memory') {
    chartInstances[type].setOption({
      tooltip: {
        trigger: 'axis',
        formatter: (params: any) => {
          let html = `<div style="padding:5px"><b>${params[0].name}</b><br/>`
          html += `${params[0].marker} 内存: ${params[0].data.value}%<br/>`
          if (params[0].data.topMem) html += buildProcessTable(params[0].data.topMem, 'mem')
          return html + '</div>'
        }
      },
      grid: { left: '10%', right: '5%', bottom: '15%' },
      xAxis: { type: 'category', data: times },
      yAxis: { type: 'value', name: '内存 (%)', max: 100 },
      series: [{
        name: '内存',
        type: 'line',
        smooth: true,
        data: data.map(d => ({ value: d.memory?.toFixed(2) || 0, topMem: d.topMem }))
      }]
    })
  } else if (type === 'io') {
    chartInstances[type].setOption({
      tooltip: {
        trigger: 'axis',
        formatter: (params: any) => {
          let html = `<div style="padding:5px"><b>${params[0].name}</b><br/>`
          params.forEach((p: any) => {
            if (p.seriesName === '读取' || p.seriesName === '写入') {
              html += `${p.marker} ${p.seriesName}: ${formatBytes(p.data)}<br/>`
            } else if (p.seriesName === '读写次数') {
              html += `${p.marker} ${p.seriesName}: ${p.data} 次/s<br/>`
            } else if (p.seriesName === '读写时间') {
              html += `${p.marker} ${p.seriesName}: ${p.data} ms<br/>`
            }
          })
          return html + '</div>'
        }
      },
      legend: { data: ['读取', '写入', '读写次数', '读写时间'] },
      grid: { left: '10%', right: '10%', bottom: '15%' },
      xAxis: { type: 'category', data: times },
      yAxis: [
        { type: 'value', name: '(KB/s)' },
        { type: 'value', position: 'right' }
      ],
      series: [
        { name: '读取', type: 'line', data: data.map(d => (d.read / 1024).toFixed(2)) },
        { name: '写入', type: 'line', data: data.map(d => (d.write / 1024).toFixed(2)) },
        { name: '读写次数', type: 'line', yAxisIndex: 1, data: data.map(d => d.count || 0) },
        { name: '读写时间', type: 'line', yAxisIndex: 1, data: data.map(d => d.time || 0) }
      ]
    })
  } else if (type === 'network') {
    chartInstances[type].setOption({
      tooltip: {
        trigger: 'axis',
        formatter: (params: any) => {
          let html = `<div style="padding:5px"><b>${params[0].name}</b><br/>`
          params.forEach((p: any) => {
            html += `${p.marker} ${p.seriesName}: ${formatBytes(p.data)}<br/>`
          })
          return html + '</div>'
        }
      },
      legend: { data: ['上传', '下载'] },
      grid: { left: '10%', right: '10%', bottom: '15%' },
      xAxis: { type: 'category', data: times },
      yAxis: { type: 'value', name: '(KB/s)' },
      series: [
        { name: '上传', type: 'line', data: data.map(d => d.up?.toFixed(2) || 0) },
        { name: '下载', type: 'line', data: data.map(d => d.down?.toFixed(2) || 0) }
      ]
    })
  }
}

const fetchChartData = async (type: string) => {
  const timeRangeMap: Record<string, any> = {
    load: loadTimeRange.value,
    cpu: cpuTimeRange.value,
    memory: memTimeRange.value,
    io: ioTimeRange.value,
    network: netTimeRange.value
  }

  const range = timeRangeMap[type]
  if (!range) return

  try {
    const payload: Record<string, string> = {
      param: type,
      startTime: range[0].toISOString(),
      endTime: range[1].toISOString()
    }

    if (type === 'io') payload.io = selectedDisk.value
    if (type === 'network') payload.network = selectedNetwork.value

    const res = await axios.post('/api/v1/monitor/data', payload)
    const chartData = Array.isArray(res.data.data?.[0]) ? res.data.data[0] : (res.data.data || [])
    renderChart(type, chartData)
  } catch (error) {
    renderEmptyChart(type, '数据加载失败')
    console.error(`获取${type}数据失败:`, error)
  }
}

const loadMonitorSettings = async () => {
  const { data } = await axios.get('/api/v1/monitor/setting')
  monitorSettings.value = {
    defaultIO: normalizeDevice(data.data?.defaultIO),
    defaultNetwork: normalizeDevice(data.data?.defaultNetwork)
  }
}

const applyDeviceSelection = () => {
  selectedDisk.value = pickDevice(normalizeDevice(monitorSettings.value.defaultIO), diskOptions.value)
  selectedNetwork.value = pickDevice(normalizeDevice(monitorSettings.value.defaultNetwork), networkOptions.value)
}

const loadDeviceOptions = async () => {
  try {
    const [diskRes, netRes] = await Promise.all([
      axios.get('/api/v1/monitor/io-options'),
      axios.get('/api/v1/monitor/network-options')
    ])
    diskOptions.value = diskRes.data.data || []
    networkOptions.value = netRes.data.data || []
    applyDeviceSelection()
  } catch (error) {
    diskOptions.value = []
    networkOptions.value = []
    applyDeviceSelection()
    console.error('获取设备选项失败:', error)
  }
}

const loadMonitorMeta = async () => {
  try {
    await loadMonitorSettings()
  } catch (error) {
    monitorSettings.value = {
      defaultIO: 'all',
      defaultNetwork: 'all'
    }
    console.error('获取监控设置失败:', error)
  }

  await loadDeviceOptions()
}

const handleMonitorSettingsUpdated = async () => {
  await loadMonitorMeta()
  refreshAll()
}

const handleMonitorDataCleared = async () => {
  await loadMonitorMeta()
  refreshAll()
}

onMounted(async () => {
  window.addEventListener('monitor-settings-updated', handleMonitorSettingsUpdated)
  window.addEventListener('monitor-data-cleared', handleMonitorDataCleared)

  await loadMonitorMeta()
  refreshAll()
})

onUnmounted(() => {
  window.removeEventListener('monitor-settings-updated', handleMonitorSettingsUpdated)
  window.removeEventListener('monitor-data-cleared', handleMonitorDataCleared)
  Object.values(chartInstances).forEach(chart => chart?.dispose())
})
</script>

<style scoped>
.monitor-dashboard {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.global-time-card {
  --el-card-padding: 12px;
}

.time-controls {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.chart-card {
  overflow: visible;
}

.full-width {
  width: 100%;
}

.chart-row {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.title-with-select {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.chart-title {
  font-size: 16px;
  font-weight: 500;
}

.chart-container {
  width: 100%;
  height: 400px;
}

@media (max-width: 1024px) {
  .chart-row {
    grid-template-columns: 1fr;
  }

  .time-controls {
    flex-direction: column;
  }

  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>

<template>
  <div class="monitor-dashboard">
    <el-card class="global-time-card">
      <div class="time-controls">
        <el-date-picker
          v-model="globalTimeRange"
          type="datetimerange"
          value-format="YYYY-MM-DD HH:mm:ss"
          range-separator="-"
          start-placeholder="开始时间"
          end-placeholder="结束时间"
          :shortcuts="timeShortcuts"
          @change="applyGlobalTime"
        />
        <div class="time-actions">
          <span class="timezone-text">按 {{ panelTimezone }} 显示</span>
          <el-button @click="handleRefresh" :icon="Refresh">刷新</el-button>
        </div>
      </div>
    </el-card>

    <el-alert v-if="monitorNotice" :title="monitorNotice" type="warning" :closable="false" show-icon />

    <el-card class="chart-card full-width">
      <template #header>
        <div class="card-header">
          <span class="chart-title">平均负载</span>
          <el-date-picker
            v-model="loadTimeRange"
            type="datetimerange"
            value-format="YYYY-MM-DD HH:mm:ss"
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
              value-format="YYYY-MM-DD HH:mm:ss"
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
              value-format="YYYY-MM-DD HH:mm:ss"
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
              value-format="YYYY-MM-DD HH:mm:ss"
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
              value-format="YYYY-MM-DD HH:mm:ss"
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

type TimeRange = [string, string]

type MonitorSetting = {
  enabled?: boolean
  defaultIO?: string
  defaultNetwork?: string
}

type DateParts = {
  year: number
  month: number
  day: number
  hour: number
  minute: number
  second: number
}

const defaultPanelTimezone = 'Asia/Shanghai'
const panelTimezone = ref(defaultPanelTimezone)
const monitorNotice = ref('')

const monitorSettings = ref<MonitorSetting>({
  enabled: false,
  defaultIO: 'all',
  defaultNetwork: 'all'
})

const pad2 = (value: number) => String(value).padStart(2, '0')

const resolveTimeZone = (value?: string) => {
  if (!value) return defaultPanelTimezone

  try {
    new Intl.DateTimeFormat('en-US', { timeZone: value }).format(new Date())
    return value
  } catch {
    return defaultPanelTimezone
  }
}

const getTimeZoneParts = (date: Date, timeZone: string): DateParts => {
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  }).formatToParts(date)

  const getValue = (type: string) => Number(parts.find(part => part.type === type)?.value || 0)

  return {
    year: getValue('year'),
    month: getValue('month'),
    day: getValue('day'),
    hour: getValue('hour'),
    minute: getValue('minute'),
    second: getValue('second')
  }
}

const formatPanelDateTime = (date: Date, timeZone = panelTimezone.value) => {
  const parts = getTimeZoneParts(date, timeZone)
  return `${parts.year}-${pad2(parts.month)}-${pad2(parts.day)} ${pad2(parts.hour)}:${pad2(parts.minute)}:${pad2(parts.second)}`
}

const formatAxisTime = (date: string, timeZone = panelTimezone.value) => {
  const parts = getTimeZoneParts(new Date(date), timeZone)
  return `${parts.month}/${parts.day} ${pad2(parts.hour)}:${pad2(parts.minute)}`
}

const getTimeZoneOffset = (date: Date, timeZone: string) => {
  const parts = getTimeZoneParts(date, timeZone)
  const utcTime = Date.UTC(parts.year, parts.month - 1, parts.day, parts.hour, parts.minute, parts.second)
  return utcTime - date.getTime()
}

const parseDateTimeValue = (value: string) => {
  const match = value.match(/^(\d{4})-(\d{2})-(\d{2}) (\d{2}):(\d{2})(?::(\d{2}))?$/)
  if (!match) return null

  return {
    year: Number(match[1]),
    month: Number(match[2]),
    day: Number(match[3]),
    hour: Number(match[4]),
    minute: Number(match[5]),
    second: Number(match[6] || 0)
  }
}

const toUtcISOString = (value: string, timeZone = panelTimezone.value) => {
  const parsed = parseDateTimeValue(value)
  if (!parsed) {
    const date = new Date(value)
    return Number.isNaN(date.getTime()) ? new Date().toISOString() : date.toISOString()
  }

  const utcGuess = Date.UTC(parsed.year, parsed.month - 1, parsed.day, parsed.hour, parsed.minute, parsed.second)
  let offset = getTimeZoneOffset(new Date(utcGuess), timeZone)
  let timestamp = utcGuess - offset
  const adjustedOffset = getTimeZoneOffset(new Date(timestamp), timeZone)
  if (adjustedOffset !== offset) {
    timestamp = utcGuess - adjustedOffset
  }

  return new Date(timestamp).toISOString()
}

const createRecentRange = (durationMs: number): TimeRange => {
  const end = new Date()
  return [
    formatPanelDateTime(new Date(end.getTime() - durationMs)),
    formatPanelDateTime(end)
  ]
}

const getTodayRange = (): TimeRange => {
  const now = new Date()
  const parts = getTimeZoneParts(now, panelTimezone.value)
  return [
    `${parts.year}-${pad2(parts.month)}-${pad2(parts.day)} 00:00:00`,
    formatPanelDateTime(now)
  ]
}

const cloneRange = (range: TimeRange): TimeRange => [range[0], range[1]]

const syncAllTimeRanges = (range: TimeRange) => {
  globalTimeRange.value = cloneRange(range)
  loadTimeRange.value = cloneRange(range)
  cpuTimeRange.value = cloneRange(range)
  memTimeRange.value = cloneRange(range)
  ioTimeRange.value = cloneRange(range)
  netTimeRange.value = cloneRange(range)
}

const globalTimeRange = ref<TimeRange>(getTodayRange())
const loadTimeRange = ref<TimeRange>(getTodayRange())
const cpuTimeRange = ref<TimeRange>(getTodayRange())
const memTimeRange = ref<TimeRange>(getTodayRange())
const ioTimeRange = ref<TimeRange>(getTodayRange())
const netTimeRange = ref<TimeRange>(getTodayRange())

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

const getErrorMessage = (error: unknown, fallback: string) => {
  if (error instanceof Error && error.message) return error.message
  return fallback
}

const normalizeMonitorError = (error: unknown, fallback: string) => {
  const message = getErrorMessage(error, fallback)
  if (/agent|connectex|refused|timeout|send request failed/i.test(message)) {
    return 'Agent 未启动或不可达，请先确认 Agent 服务正常运行'
  }
  return message
}

const timeShortcuts = [
  { text: '最近1小时', value: () => createRecentRange(3600000) },
  { text: '最近6小时', value: () => createRecentRange(21600000) },
  { text: '今天', value: getTodayRange },
  { text: '最近3天', value: () => createRecentRange(259200000) },
  { text: '最近7天', value: () => createRecentRange(604800000) }
]

const isValidTimeRange = (range?: TimeRange | null): range is TimeRange => {
  return Array.isArray(range) && range.length === 2 && Boolean(range[0]) && Boolean(range[1])
}

const extendTimeRangeToNow = (range: TimeRange): TimeRange => {
  if (!isValidTimeRange(range)) return getTodayRange()

  const end = formatPanelDateTime(new Date())
  return [range[0], end]
}

const syncLiveTimeRanges = () => {
  globalTimeRange.value = extendTimeRangeToNow(globalTimeRange.value)
  loadTimeRange.value = extendTimeRangeToNow(loadTimeRange.value)
  cpuTimeRange.value = extendTimeRangeToNow(cpuTimeRange.value)
  memTimeRange.value = extendTimeRangeToNow(memTimeRange.value)
  ioTimeRange.value = extendTimeRangeToNow(ioTimeRange.value)
  netTimeRange.value = extendTimeRangeToNow(netTimeRange.value)
}

const applyGlobalTime = () => {
  syncAllTimeRanges(globalTimeRange.value)
  refreshAll()
}

const refreshAll = () => {
  syncLiveTimeRanges()
  fetchChartData('load')
  fetchChartData('cpu')
  fetchChartData('memory')
  fetchChartData('io')
  fetchChartData('network')
}

const formatTime = (date: string) => {
  return formatAxisTime(date)
}

const formatBytes = (bytes: number | string) => {
  const value = Number(bytes)
  if (!Number.isFinite(value) || value <= 0) return '0 B'
  if (value < 1024) return `${value.toFixed(2)} B`
  if (value < 1048576) return `${(value / 1024).toFixed(2)} KB`
  if (value < 1073741824) return `${(value / 1048576).toFixed(2)} MB`
  return `${(value / 1073741824).toFixed(2)} GB`
}

const formatKilobytes = (kb: number | string) => {
  const value = Number(kb)
  if (!Number.isFinite(value) || value <= 0) return '0.00 KB'
  if (value < 1024) return `${value.toFixed(2)} KB`
  if (value < 1048576) return `${(value / 1024).toFixed(2)} MB`
  return `${(value / 1048576).toFixed(2)} GB`
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

  chartInstances[type].clear()

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
              html += `${p.marker} ${p.seriesName}: ${formatKilobytes(p.data)}<br/>`
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
        { name: '读取', type: 'line', data: data.map(d => d.read / 1024) },
        { name: '写入', type: 'line', data: data.map(d => d.write / 1024) },
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
            html += `${p.marker} ${p.seriesName}: ${formatKilobytes(p.data)}<br/>`
          })
          return html + '</div>'
        }
      },
      legend: { data: ['上传', '下载'] },
      grid: { left: '10%', right: '10%', bottom: '15%' },
      xAxis: { type: 'category', data: times },
      yAxis: { type: 'value', name: '(KB/s)' },
      series: [
        { name: '上传', type: 'line', data: data.map(d => d.up ?? 0) },
        { name: '下载', type: 'line', data: data.map(d => d.down ?? 0) }
      ]
    })
  }
}

const fetchChartData = async (type: string) => {
  const timeRangeMap: Record<string, TimeRange> = {
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
      startTime: toUtcISOString(range[0]),
      endTime: toUtcISOString(range[1])
    }

    if (type === 'io') payload.io = selectedDisk.value
    if (type === 'network') payload.network = selectedNetwork.value

    const res = await axios.post('/api/v1/monitor/data', payload)
    const chartData = Array.isArray(res.data.data?.[0]) ? res.data.data[0] : (res.data.data || [])
    if (chartData.length === 0 && monitorSettings.value.enabled === false) {
      monitorNotice.value = '监控尚未开启，请先在设置里打开监控开关'
      renderEmptyChart(type, monitorNotice.value)
      return
    }

    monitorNotice.value = ''
    renderChart(type, chartData)
  } catch (error) {
    const message = normalizeMonitorError(error, '数据加载失败')
    monitorNotice.value = message
    renderEmptyChart(type, message)
    console.error(`获取${type}数据失败:`, error)
  }
}

const loadPanelTimezone = async () => {
  try {
    const response = await axios.get('/api/v1/settings/system')
    panelTimezone.value = resolveTimeZone(response.data.settings?.Timezone)
  } catch (error) {
    panelTimezone.value = defaultPanelTimezone
    console.error('获取面板时区失败:', error)
  }

  syncAllTimeRanges(getTodayRange())
}

const loadMonitorSettings = async () => {
  const { data } = await axios.get('/api/v1/monitor/setting')
  monitorSettings.value = {
    enabled: data.data?.enabled ?? false,
    defaultIO: normalizeDevice(data.data?.defaultIO),
    defaultNetwork: normalizeDevice(data.data?.defaultNetwork)
  }
}

const applyDeviceSelection = (preserveCurrent = false) => {
  const nextDisk = preserveCurrent
    ? normalizeDevice(selectedDisk.value)
    : normalizeDevice(monitorSettings.value.defaultIO)
  const nextNetwork = preserveCurrent
    ? normalizeDevice(selectedNetwork.value)
    : normalizeDevice(monitorSettings.value.defaultNetwork)

  selectedDisk.value = pickDevice(nextDisk, diskOptions.value)
  selectedNetwork.value = pickDevice(nextNetwork, networkOptions.value)
}

const loadDeviceOptions = async (preserveSelection = false) => {
  try {
    const [diskRes, netRes] = await Promise.all([
      axios.get('/api/v1/monitor/io-options'),
      axios.get('/api/v1/monitor/network-options')
    ])
    diskOptions.value = diskRes.data.data || []
    networkOptions.value = netRes.data.data || []
    applyDeviceSelection(preserveSelection)
  } catch (error) {
    diskOptions.value = []
    networkOptions.value = []
    applyDeviceSelection(preserveSelection)
    console.error('获取设备选项失败:', error)
  }
}

const handleRefresh = async () => {
  syncLiveTimeRanges()
  await loadDeviceOptions(true)
  fetchChartData('load')
  fetchChartData('cpu')
  fetchChartData('memory')
  fetchChartData('io')
  fetchChartData('network')
}

const loadMonitorMeta = async () => {
  await loadPanelTimezone()

  try {
    await loadMonitorSettings()
    monitorNotice.value = ''
  } catch (error) {
    monitorSettings.value = {
      enabled: false,
      defaultIO: 'all',
      defaultNetwork: 'all'
    }
    monitorNotice.value = normalizeMonitorError(error, '监控设置加载失败')
    console.error('获取监控设置失败:', error)
  }

  if (monitorSettings.value.enabled === false && !monitorNotice.value) {
    monitorNotice.value = '监控尚未开启，请先在设置里打开监控开关'
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

const handleMonitorDashboardActivated = () => {
  refreshAll()
}

onMounted(async () => {
  window.addEventListener('monitor-settings-updated', handleMonitorSettingsUpdated)
  window.addEventListener('monitor-data-cleared', handleMonitorDataCleared)
  window.addEventListener('monitor-dashboard-activated', handleMonitorDashboardActivated)

  await loadMonitorMeta()
  refreshAll()
})

onUnmounted(() => {
  window.removeEventListener('monitor-settings-updated', handleMonitorSettingsUpdated)
  window.removeEventListener('monitor-data-cleared', handleMonitorDataCleared)
  window.removeEventListener('monitor-dashboard-activated', handleMonitorDashboardActivated)
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

.time-actions {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.timezone-text {
  color: var(--text-secondary);
  font-size: 0.85rem;
  white-space: nowrap;
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

  .time-actions {
    width: 100%;
    justify-content: space-between;
  }

  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>

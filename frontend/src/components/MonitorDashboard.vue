<template>
  <div class="monitor-dashboard">
    <div class="time-selector">
      <el-radio-group v-model="timeRange" @change="handleTimeChange">
        <el-radio-button label="1h">最近1小时</el-radio-button>
        <el-radio-button label="6h">最近6小时</el-radio-button>
        <el-radio-button label="24h">最近24小时</el-radio-button>
        <el-radio-button label="7d">最近7天</el-radio-button>
      </el-radio-group>
    </div>

    <div v-if="loading" class="loading">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>加载中...</span>
    </div>

    <div v-else class="charts-grid">
      <div class="chart-card">
        <div class="chart-title">平均负载</div>
        <div ref="loadChart" class="chart"></div>
      </div>

      <div class="chart-card">
        <div class="chart-title">CPU性能监控</div>
        <div ref="cpuChart" class="chart"></div>
      </div>

      <div class="chart-card">
        <div class="chart-title">内存使用监控</div>
        <div ref="memChart" class="chart"></div>
      </div>

      <div class="chart-card">
        <div class="chart-title">磁盘IO监控</div>
        <div ref="diskChart" class="chart"></div>
      </div>

      <div class="chart-card">
        <div class="chart-title">网络IO监控</div>
        <div ref="netChart" class="chart"></div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import axios from '@/utils/axios'
import * as echarts from 'echarts'
import { Loading } from '@element-plus/icons-vue'

const timeRange = ref('1h')
const loading = ref(true)
const loadChart = ref()
const cpuChart = ref()
const memChart = ref()
const diskChart = ref()
const netChart = ref()

let charts: echarts.ECharts[] = []

const getTimeRange = () => {
  const end = new Date()
  const start = new Date()
  switch (timeRange.value) {
    case '1h': start.setHours(start.getHours() - 1); break
    case '6h': start.setHours(start.getHours() - 6); break
    case '24h': start.setHours(start.getHours() - 24); break
    case '7d': start.setDate(start.getDate() - 7); break
  }
  return { startTime: start.toISOString(), endTime: end.toISOString() }
}

const fetchData = async () => {
  loading.value = true
  try {
    const { data } = await axios.post('/api/v1/monitor/data', getTimeRange())
    renderCharts(data.data || [])
  } catch (error) {
    console.error('获取监控数据失败:', error)
  } finally {
    loading.value = false
  }
}

const renderCharts = (data: any[]) => {
  const times = data.map(d => new Date(d.timestamp).toLocaleTimeString())

  charts.forEach(c => c.dispose())
  charts = []

  charts.push(echarts.init(loadChart.value))
  charts[0].setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['1分钟', '5分钟', '15分钟'] },
    xAxis: { type: 'category', data: times },
    yAxis: { type: 'value' },
    series: [
      { name: '1分钟', type: 'line', data: data.map(d => d.load1) },
      { name: '5分钟', type: 'line', data: data.map(d => d.load5) },
      { name: '15分钟', type: 'line', data: data.map(d => d.load15) }
    ]
  })

  charts.push(echarts.init(cpuChart.value))
  charts[1].setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: times },
    yAxis: { type: 'value', max: 100 },
    series: [{ name: 'CPU使用率', type: 'line', data: data.map(d => d.cpuPercent.toFixed(2)), areaStyle: {} }]
  })

  charts.push(echarts.init(memChart.value))
  charts[2].setOption({
    tooltip: { trigger: 'axis' },
    xAxis: { type: 'category', data: times },
    yAxis: { type: 'value', max: 100 },
    series: [{ name: '内存使用率', type: 'line', data: data.map(d => d.memPercent.toFixed(2)), areaStyle: {} }]
  })

  charts.push(echarts.init(diskChart.value))
  charts[3].setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['读取', '写入'] },
    xAxis: { type: 'category', data: times },
    yAxis: { type: 'value' },
    series: [
      { name: '读取', type: 'line', data: data.map(d => (d.diskReadBytes / 1024 / 1024).toFixed(2)) },
      { name: '写入', type: 'line', data: data.map(d => (d.diskWriteBytes / 1024 / 1024).toFixed(2)) }
    ]
  })

  charts.push(echarts.init(netChart.value))
  charts[4].setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['接收', '发送'] },
    xAxis: { type: 'category', data: times },
    yAxis: { type: 'value' },
    series: [
      { name: '接收', type: 'line', data: data.map(d => (d.netRecvBytes / 1024 / 1024).toFixed(2)) },
      { name: '发送', type: 'line', data: data.map(d => (d.netSentBytes / 1024 / 1024).toFixed(2)) }
    ]
  })
}

const handleTimeChange = () => fetchData()

onMounted(() => fetchData())
onUnmounted(() => charts.forEach(c => c.dispose()))
</script>

<style scoped>
.monitor-dashboard {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.time-selector {
  display: flex;
  justify-content: center;
  padding: 1rem 0;
}

.loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 3rem;
  color: var(--text-secondary);
}

.charts-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.chart-card {
  background: var(--card-bg);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 1rem;
}

.chart-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 0.75rem;
}

.chart {
  width: 100%;
  height: 300px;
}

@media (max-width: 1024px) {
  .charts-grid {
    grid-template-columns: 1fr;
  }
}
</style>

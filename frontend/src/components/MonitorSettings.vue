<template>
  <div class="monitor-settings">
    <el-form :model="form" label-width="120px">
      <el-form-item label="监控功能">
        <el-switch v-model="form.enabled" @change="handleSave" />
      </el-form-item>

      <el-form-item label="数据保存时长">
        <el-input-number v-model="form.retentionDays" :min="1" :max="90" @change="handleSave" />
        <span class="unit">天</span>
      </el-form-item>

      <el-form-item label="采集间隔">
        <el-input-number v-model="form.collectInterval" :min="10" :max="3600" @change="handleSave" />
        <span class="unit">秒</span>
      </el-form-item>

      <el-form-item label="默认磁盘">
        <el-select v-model="form.defaultIO" placeholder="请选择默认磁盘" clearable @change="handleSave">
          <el-option label="全部" value="all" />
          <el-option v-for="disk in diskOptions" :key="disk" :label="disk" :value="disk" />
        </el-select>
      </el-form-item>

      <el-form-item label="默认网卡">
        <el-select v-model="form.defaultNetwork" placeholder="请选择默认网卡" clearable @change="handleSave">
          <el-option label="全部" value="all" />
          <el-option v-for="network in networkOptions" :key="network" :label="network" :value="network" />
        </el-select>
      </el-form-item>

      <el-form-item>
        <el-button type="danger" @click="handleClear">清空监控记录</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import axios from '@/utils/axios'
import { ElMessage, ElMessageBox } from 'element-plus'

const form = ref({
  enabled: false,
  retentionDays: 7,
  collectInterval: 300,
  defaultIO: 'all',
  defaultNetwork: 'all'
})

const diskOptions = ref<string[]>([])
const networkOptions = ref<string[]>([])

const normalizeDevice = (value?: string) => value && value.trim() ? value : 'all'

const getErrorMessage = (error: unknown, fallback: string) => {
  if (typeof error === 'string') return error
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

const ensureDeviceValue = (value: string, options: string[]) => {
  if (value === 'all') return 'all'
  return options.includes(value) ? value : 'all'
}

const loadDeviceOptions = async () => {
  const [diskRes, netRes] = await Promise.all([
    axios.get('/api/v1/monitor/io-options'),
    axios.get('/api/v1/monitor/network-options')
  ])

  diskOptions.value = diskRes.data.data || []
  networkOptions.value = netRes.data.data || []

  form.value.defaultIO = ensureDeviceValue(normalizeDevice(form.value.defaultIO), diskOptions.value)
  form.value.defaultNetwork = ensureDeviceValue(normalizeDevice(form.value.defaultNetwork), networkOptions.value)
}

const fetchSettings = async () => {
  try {
    const [{ data }] = await Promise.all([
      axios.get('/api/v1/monitor/setting'),
      loadDeviceOptions()
    ])

    form.value = {
      enabled: data.data?.enabled ?? false,
      retentionDays: data.data?.retentionDays ?? 7,
      collectInterval: data.data?.collectInterval ?? 300,
      defaultIO: ensureDeviceValue(normalizeDevice(data.data?.defaultIO), diskOptions.value),
      defaultNetwork: ensureDeviceValue(normalizeDevice(data.data?.defaultNetwork), networkOptions.value)
    }
  } catch (error) {
    ElMessage.error(normalizeMonitorError(error, '获取设置失败'))
    console.error('获取设置失败:', error)
  }
}

const handleSave = async () => {
  try {
    const payload = {
      ...form.value,
      defaultIO: ensureDeviceValue(normalizeDevice(form.value.defaultIO), diskOptions.value),
      defaultNetwork: ensureDeviceValue(normalizeDevice(form.value.defaultNetwork), networkOptions.value)
    }

    form.value = payload
    await axios.post('/api/v1/monitor/setting', payload)
    window.dispatchEvent(new CustomEvent('monitor-settings-updated'))
    ElMessage.success('设置已保存')
  } catch (error) {
    ElMessage.error(normalizeMonitorError(error, '保存失败'))
  }
}

const handleClear = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有监控记录吗？', '警告', { type: 'warning' })
    await axios.delete('/api/v1/monitor/data')
    window.dispatchEvent(new CustomEvent('monitor-data-cleared'))
    ElMessage.success('监控记录已清空')
  } catch (error) {
    const message = getErrorMessage(error, '清空失败')
    if (message !== 'cancel' && message !== 'close') {
      ElMessage.error(normalizeMonitorError(error, '清空失败'))
    }
  }
}

onMounted(() => fetchSettings())
</script>

<style scoped>
.monitor-settings {
  max-width: 600px;
  padding: 1rem;
}

.unit {
  margin-left: 0.5rem;
  color: var(--text-secondary);
  font-size: 0.85rem;
}
</style>

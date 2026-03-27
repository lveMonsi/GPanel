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
  collectInterval: 60
})

const fetchSettings = async () => {
  try {
    const { data } = await axios.get('/api/v1/monitor/setting')
    form.value = data.data
  } catch (error) {
    console.error('获取设置失败:', error)
  }
}

const handleSave = async () => {
  try {
    await axios.post('/api/v1/monitor/setting', form.value)
    ElMessage.success('设置已保存')
  } catch (error) {
    ElMessage.error('保存失败')
  }
}

const handleClear = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有监控记录吗？', '警告', { type: 'warning' })
    await axios.delete('/api/v1/monitor/data')
    ElMessage.success('监控记录已清空')
  } catch (error) {
    if (error !== 'cancel') ElMessage.error('清空失败')
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

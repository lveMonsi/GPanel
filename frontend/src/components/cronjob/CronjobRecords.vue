<template>
  <el-drawer
    v-model="visible"
    title="执行记录"
    size="75%"
    destroy-on-close
  >
    <template #header>
      <div class="records-header">
        <div class="header-info">
          <el-tag effect="dark" type="success">{{ typeLabel }} - {{ jobName }}</el-tag>
          <el-tag :type="statusTagType(jobStatus)">{{ statusLabel(jobStatus) }}</el-tag>
        </div>
        <div class="header-actions">
          <el-button type="primary" link @click="onHandleOnce" :disabled="jobStatus === 'disabled'">
            手动执行
          </el-button>
          <el-button type="primary" link @click="onToggle">
            {{ jobStatus === 'enabled' ? '禁用' : '启用' }}
          </el-button>
          <el-button type="danger" link @click="onClean" :disabled="records.length === 0">
            清空记录
          </el-button>
        </div>
      </div>
    </template>

    <div class="records-content">
      <!-- 筛选 -->
      <div class="records-filter">
        <el-select v-model="searchStatus" placeholder="状态筛选" clearable style="width: 140px" @change="loadRecords(true)">
          <el-option label="全部" value="" />
          <el-option label="成功" value="success" />
          <el-option label="等待中" value="waiting" />
          <el-option label="失败" value="failed" />
        </el-select>
        <el-button :icon="Refresh" circle @click="loadRecords(false)" />
      </div>

      <div v-if="records.length === 0" class="no-records">
        暂无执行记录
      </div>

      <el-row v-else :gutter="16">
        <!-- 左侧记录列表 -->
        <el-col :span="8">
          <div class="record-list">
            <div
              v-for="record in records"
              :key="record.id"
              class="record-item"
              :class="{ active: currentRecord?.id === record.id }"
              @click="selectRecord(record)"
            >
              <span class="record-status-dot" :class="record.status" />
              <span class="record-time">{{ record.startTime }}</span>
              <el-button
                v-if="record.status === 'waiting'"
                link
                type="warning"
                size="small"
                @click.stop="onStop(record)"
              >
                停止
              </el-button>
            </div>

            <el-pagination
              v-model:current-page="page"
              v-model:page-size="pageSize"
              :total="total"
              :page-sizes="[10, 20, 50, 100]"
              small
              layout="total, sizes, prev, next"
              @size-change="loadRecords(true)"
              @current-change="loadRecords(false)"
              style="margin-top: 12px"
            />
          </div>
        </el-col>

        <!-- 右侧详情 -->
        <el-col :span="16">
          <el-card v-if="currentRecord" shadow="never">
            <div class="record-detail">
              <div class="detail-row">
                <span class="detail-label">开始时间</span>
                <span>{{ currentRecord.startTime }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">执行耗时</span>
                <span>{{ formatDuration(currentRecord.duration) }}</span>
              </div>
              <div class="detail-row">
                <span class="detail-label">执行状态</span>
                <el-tag size="small" :type="statusTagType(currentRecord.status)">
                  {{ statusLabel(currentRecord.status) }}
                </el-tag>
                <span v-if="currentRecord.message" class="detail-message">
                  {{ currentRecord.message }}
                </span>
              </div>
            </div>

            <el-divider />

            <div class="log-section">
              <div class="log-header">
                <span>执行日志</span>
                <el-button link type="primary" @click="loadLog" :loading="logLoading">
                  刷新
                </el-button>
              </div>
              <pre class="log-content" v-if="logContent">{{ logContent }}</pre>
              <div v-else class="no-log">暂无日志</div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import {
  searchRecords,
  getRecordLog,
  cleanRecords,
  handleOnceCronjob,
  toggleCronjob,
  stopCronjob
} from '@/api/modules/cronjob'
import type { Cronjob, JobRecord, CronjobType } from '@/api/interface/cronjob'

const emit = defineEmits<{
  refresh: []
}>()

const visible = ref(false)
const jobId = ref(0)
const jobName = ref('')
const jobType = ref<CronjobType>('shell')
const jobStatus = ref('')
const records = ref<JobRecord[]>([])
const currentRecord = ref<JobRecord | null>(null)
const logContent = ref('')
const logLoading = ref(false)
const searchStatus = ref('')
const page = ref(1)
const pageSize = ref(20)
const total = ref(0)

const typeLabels: Record<string, string> = {
  shell: 'Shell 脚本',
  curl: 'Curl 请求',
  directory: '目录备份',
  clean: '磁盘清理',
  cleanLog: '日志清理'
}

const typeLabel = ref('')

const statusLabel = (status: string) => {
  const map: Record<string, string> = {
    enabled: '已启用',
    disabled: '已禁用',
    success: '成功',
    failed: '失败',
    waiting: '执行中'
  }
  return map[status] || status
}

const statusTagType = (status: string): '' | 'success' | 'warning' | 'danger' | 'info' => {
  const map: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
    enabled: 'success',
    disabled: 'info',
    success: 'success',
    failed: 'danger',
    waiting: 'warning'
  }
  return map[status] || 'info'
}

const formatDuration = (ms: number) => {
  if (ms <= 0) return '-'
  if (ms < 1000) return `${ms} ms`
  return `${(ms / 1000).toFixed(1)} s`
}

const open = (row: Cronjob) => {
  jobId.value = row.id
  jobName.value = row.name
  jobType.value = row.type
  jobStatus.value = row.status
  typeLabel.value = typeLabels[row.type] || row.type
  visible.value = true
  searchStatus.value = ''
  page.value = 1
  loadRecords(true)
}

const loadRecords = async (resetSelection: boolean) => {
  try {
    const res = await searchRecords({
      page: page.value,
      pageSize: pageSize.value,
      cronjobId: jobId.value,
      status: searchStatus.value
    })
    records.value = res.data?.data?.items || []
    total.value = res.data?.data?.total || 0

    if (resetSelection && records.value.length > 0) {
      selectRecord(records.value[0])
    }
  } catch {
    ElMessage.error('加载记录失败')
  }
}

const selectRecord = async (record: JobRecord) => {
  currentRecord.value = record
  await loadLog()
}

const loadLog = async () => {
  if (!currentRecord.value) return
  logLoading.value = true
  try {
    const res = await getRecordLog(currentRecord.value.id)
    logContent.value = res.data?.data || res.data || ''
  } catch {
    logContent.value = ''
  } finally {
    logLoading.value = false
  }
}

const onHandleOnce = async () => {
  try {
    await handleOnceCronjob(jobId.value)
    ElMessage.success('已触发执行')
    setTimeout(() => loadRecords(true), 1000)
  } catch {
    ElMessage.error('执行失败')
  }
}

const onToggle = async () => {
  const newStatus = jobStatus.value === 'enabled' ? 'disabled' : 'enabled'
  const msg = newStatus === 'disabled' ? '确定要禁用该任务吗？' : '确定要启用该任务吗？'

  try {
    await ElMessageBox.confirm(msg, '提示', { type: 'warning' })
    await toggleCronjob({ id: jobId.value, status: newStatus })
    jobStatus.value = newStatus
    ElMessage.success('操作成功')
    emit('refresh')
  } catch {
    // cancelled
  }
}

const onClean = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有执行记录吗？此操作不可恢复。', '清空记录', { type: 'warning' })
    await cleanRecords(jobId.value)
    ElMessage.success('清空成功')
    records.value = []
    currentRecord.value = null
    logContent.value = ''
    total.value = 0
  } catch {
    // cancelled
  }
}

const onStop = async (_record: JobRecord) => {
  try {
    await ElMessageBox.confirm('确定要停止当前执行吗？', '停止任务', { type: 'warning' })
    await stopCronjob(jobId.value)
    ElMessage.success('已发送停止信号')
    setTimeout(() => loadRecords(false), 1000)
  } catch {
    // cancelled
  }
}

defineExpose({ open })
</script>

<style scoped>
.records-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.header-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 4px;
}

.records-filter {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
}

.no-records {
  text-align: center;
  padding: 60px 0;
  color: var(--el-text-color-secondary);
}

.record-list {
  max-height: calc(100vh - 240px);
  overflow-y: auto;
}

.record-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
}

.record-item:hover {
  background: var(--el-fill-color-light);
}

.record-item.active {
  background: var(--el-color-primary-light-9);
  border-left: 3px solid var(--el-color-primary);
}

.record-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.record-status-dot.success {
  background: var(--el-color-success);
}

.record-status-dot.failed {
  background: var(--el-color-danger);
}

.record-status-dot.waiting {
  background: var(--el-color-warning);
  animation: pulse 1.5s infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.record-time {
  font-size: 13px;
  flex: 1;
}

.record-detail {
  display: flex;
  gap: 32px;
  flex-wrap: wrap;
}

.detail-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.detail-label {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}

.detail-message {
  color: var(--el-color-danger);
  font-size: 13px;
  margin-left: 8px;
}

.log-section {
  margin-top: 8px;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  font-weight: 500;
}

.log-content {
  background: #1e1e1e;
  color: #d4d4d4;
  border-radius: 6px;
  padding: 16px;
  font-family: 'Cascadia Code', 'Fira Code', Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
  max-height: calc(100vh - 420px);
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}

.no-log {
  text-align: center;
  padding: 40px 0;
  color: var(--el-text-color-secondary);
}
</style>

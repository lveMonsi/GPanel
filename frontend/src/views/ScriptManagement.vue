<template>
  <div class="cronjob-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-button type="primary" @click="openCreate">
              <el-icon><Plus /></el-icon>
              新建任务
            </el-button>
            <el-button type="danger" plain @click="batchDelete" :disabled="selectedIds.length === 0">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </div>
          <div class="header-right">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索任务名称"
              clearable
              style="width: 200px"
              @clear="loadData"
              @keyup.enter="loadData"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-select v-model="searchType" placeholder="任务类型" clearable style="width: 140px" @change="loadData">
              <el-option label="全部类型" value="" />
              <el-option label="Shell 脚本" value="shell" />
              <el-option label="Curl 请求" value="curl" />
              <el-option label="目录备份" value="directory" />
              <el-option label="磁盘清理" value="clean" />
              <el-option label="日志清理" value="cleanLog" />
            </el-select>
            <el-select v-model="searchStatus" placeholder="状态" clearable style="width: 120px" @change="loadData">
              <el-option label="全部状态" value="" />
              <el-option label="已启用" value="enabled" />
              <el-option label="已禁用" value="disabled" />
            </el-select>
            <el-button :icon="Refresh" circle @click="loadData" />
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="jobs"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="50" />

        <el-table-column prop="name" label="任务名称" min-width="140" show-overflow-tooltip />

        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-switch
              :model-value="row.status === 'enabled'"
              @change="(val: boolean) => onToggle(row, val)"
              size="small"
            />
          </template>
        </el-table-column>

        <el-table-column label="任务类型" width="110">
          <template #default="{ row }">
            <el-tag size="small" :type="typeTagColor(row.type)">
              {{ typeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>

        <el-table-column label="执行周期" min-width="180">
          <template #default="{ row }">
            <span class="spec-text">{{ specToReadable(row.spec) }}</span>
            <span class="spec-raw">({{ row.spec }})</span>
          </template>
        </el-table-column>

        <el-table-column label="上次执行" min-width="180">
          <template #default="{ row }">
            <template v-if="row.lastRecordTime">
              <el-tag size="small" :type="statusTagType(row.lastRecordStatus)">
                {{ statusLabel(row.lastRecordStatus) }}
              </el-tag>
              <span class="last-time">{{ row.lastRecordTime }}</span>
            </template>
            <span v-else class="no-record">暂无记录</span>
          </template>
        </el-table-column>

        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="onHandleOnce(row)">执行</el-button>
            <el-button link type="primary" @click="openRecords(row)">记录</el-button>
            <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="onDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="loadData"
        @current-change="loadData"
        style="margin-top: 16px"
      />
    </el-card>

    <CronjobForm ref="formRef" @success="loadData" />
    <CronjobRecords ref="recordsRef" @refresh="loadData" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Delete, Search, Refresh } from '@element-plus/icons-vue'
import { searchCronjobs, deleteCronjob, toggleCronjob, handleOnceCronjob } from '@/api/modules/cronjob'
import type { Cronjob } from '@/api/interface/cronjob'
import CronjobForm from '@/components/cronjob/CronjobForm.vue'
import CronjobRecords from '@/components/cronjob/CronjobRecords.vue'

const loading = ref(false)
const jobs = ref<Cronjob[]>([])
const selectedIds = ref<number[]>([])
const searchKeyword = ref('')
const searchType = ref('')
const searchStatus = ref('')
const formRef = ref<InstanceType<typeof CronjobForm>>()
const recordsRef = ref<InstanceType<typeof CronjobRecords>>()

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const typeLabels: Record<string, string> = {
  shell: 'Shell 脚本',
  curl: 'Curl 请求',
  directory: '目录备份',
  clean: '磁盘清理',
  cleanLog: '日志清理'
}

const typeLabel = (type: string) => typeLabels[type] || type

const typeTagColor = (type: string): '' | 'success' | 'warning' | 'danger' | 'info' => {
  const map: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
    shell: '',
    curl: 'success',
    directory: 'warning',
    clean: 'danger',
    cleanLog: 'info'
  }
  return map[type] || 'info'
}

const statusTagType = (status: string): '' | 'success' | 'warning' | 'danger' | 'info' => {
  const map: Record<string, '' | 'success' | 'warning' | 'danger' | 'info'> = {
    success: 'success',
    failed: 'danger',
    waiting: 'warning'
  }
  return map[status] || 'info'
}

const statusLabel = (status: string) => {
  const map: Record<string, string> = { success: '成功', failed: '失败', waiting: '执行中' }
  return map[status] || status
}

const weekNames: Record<number, string> = {
  0: '周日', 1: '周一', 2: '周二', 3: '周三', 4: '周四', 5: '周五', 6: '周六'
}

const specToReadable = (spec: string): string => {
  if (!spec) return ''

  // 处理 @every 格式
  if (spec.startsWith('@every')) {
    return spec
  }

  const parts = spec.split(' ')
  if (parts.length !== 5) return spec

  const [minute, hour, dom, _month, dow] = parts
  const pad = (s: string) => s.padStart(2, '0')

  if (hour === '*' && dom === '*' && dow === '*') {
    if (minute.startsWith('*/')) return `每 ${minute.replace('*/', '')} 分钟`
    return `每小时 第${pad(minute)}分`
  }
  if (hour.startsWith('*/')) return `每 ${hour.replace('*/', '')} 小时 第${pad(minute)}分`
  if (dom.startsWith('*/')) return `每 ${dom.replace('*/', '')} 天 ${pad(hour)}:${pad(minute)}`
  if (dom !== '*') return `每月 ${dom} 日 ${pad(hour)}:${pad(minute)}`
  if (dow !== '*') return `每${weekNames[parseInt(dow)] || dow} ${pad(hour)}:${pad(minute)}`
  return `每天 ${pad(hour)}:${pad(minute)}`
}

const loadData = async () => {
  loading.value = true
  try {
    const res = await searchCronjobs({
      page: pagination.page,
      pageSize: pagination.pageSize,
      keyword: searchKeyword.value,
      type: searchType.value,
      status: searchStatus.value
    })
    jobs.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch {
    ElMessage.error('加载任务列表失败')
  } finally {
    loading.value = false
  }
}

const handleSelectionChange = (selection: Cronjob[]) => {
  selectedIds.value = selection.map(item => item.id)
}

const openCreate = () => {
  formRef.value?.open()
}

const openEdit = (row: Cronjob) => {
  formRef.value?.open(row)
}

const openRecords = (row: Cronjob) => {
  recordsRef.value?.open(row)
}

const onToggle = async (row: Cronjob, val: boolean) => {
  const newStatus = val ? 'enabled' : 'disabled'
  try {
    await toggleCronjob({ id: row.id, status: newStatus })
    row.status = newStatus
    ElMessage.success(val ? '已启用' : '已禁用')
  } catch {
    ElMessage.error('操作失败')
  }
}

const onHandleOnce = async (row: Cronjob) => {
  try {
    await handleOnceCronjob(row.id)
    ElMessage.success('已触发执行')
    setTimeout(() => loadData(), 1500)
  } catch {
    ElMessage.error('执行失败')
  }
}

const onDelete = async (row: Cronjob) => {
  try {
    await ElMessageBox.confirm(`确定要删除任务 "${row.name}" 吗？相关的执行记录和日志也将被删除。`, '删除确认', {
      type: 'warning'
    })
    await deleteCronjob({ ids: [row.id] })
    ElMessage.success('删除成功')
    loadData()
  } catch {
    // cancelled
  }
}

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedIds.value.length} 个任务吗？`, '批量删除', {
      type: 'warning'
    })
    await deleteCronjob({ ids: selectedIds.value })
    ElMessage.success('删除成功')
    loadData()
  } catch {
    // cancelled
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped>
.cronjob-page {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.header-left {
  display: flex;
  gap: 10px;
}

.header-right {
  display: flex;
  gap: 10px;
  align-items: center;
}

.spec-text {
  color: var(--el-text-color-primary);
}

.spec-raw {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  margin-left: 6px;
}

.last-time {
  margin-left: 8px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.no-record {
  color: var(--el-text-color-placeholder);
  font-size: 13px;
}
</style>

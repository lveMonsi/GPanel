<template>
  <div class="log-page">
    <el-tabs v-model="activeTab" @tab-change="onTabChange">
      <!-- 操作日志 -->
      <el-tab-pane label="操作日志" name="operation">
        <el-card>
          <template #header>
            <div class="card-header">
              <div class="header-left">
                <el-button type="danger" plain @click="openCleanDialog">
                  <el-icon><Delete /></el-icon>
                  清理日志
                </el-button>
                <div class="stats-info" v-if="opStats">
                  <el-tag type="info" size="small">总计 {{ opStats.total }} 条</el-tag>
                  <el-tag type="success" size="small">今日 {{ opStats.todayCount }} 条</el-tag>
                </div>
              </div>
              <div class="header-right">
                <el-input
                  v-model="opSearch.keyword"
                  placeholder="搜索详情"
                  clearable
                  style="width: 180px"
                  @clear="loadOperationLogs"
                  @keyup.enter="loadOperationLogs"
                >
                  <template #prefix>
                    <el-icon><Search /></el-icon>
                  </template>
                </el-input>
                <el-select v-model="opSearch.resource" placeholder="资源类型" clearable style="width: 130px" @change="loadOperationLogs">
                  <el-option label="全部类型" value="" />
                  <el-option label="认证" value="auth" />
                  <el-option label="主机" value="host" />
                  <el-option label="主机分组" value="host_group" />
                  <el-option label="计划任务" value="cronjob" />
                  <el-option label="防火墙" value="firewall" />
                  <el-option label="进程" value="process" />
                  <el-option label="SSH" value="ssh" />
                  <el-option label="设置" value="setting" />
                  <el-option label="快捷命令" value="quick_command" />
                  <el-option label="系统" value="system" />
                  <el-option label="日志" value="log" />
                </el-select>
                <el-select v-model="opSearch.status" placeholder="状态" clearable style="width: 110px" @change="loadOperationLogs">
                  <el-option label="全部" value="" />
                  <el-option label="成功" value="success" />
                  <el-option label="失败" value="failed" />
                </el-select>
                <el-date-picker
                  v-model="opDateRange"
                  type="daterange"
                  start-placeholder="开始日期"
                  end-placeholder="结束日期"
                  value-format="YYYY-MM-DD"
                  style="width: 240px"
                  @change="onDateRangeChange"
                />
                <el-button :icon="Refresh" circle @click="loadOperationLogs" />
              </div>
            </div>
          </template>

          <el-table v-loading="opLoading" :data="opLogs">
            <el-table-column prop="createdAt" label="时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.createdAt) }}
              </template>
            </el-table-column>
            <el-table-column prop="username" label="用户" width="110" />
            <el-table-column prop="ip" label="IP 地址" width="140" />
            <el-table-column label="资源" width="110">
              <template #default="{ row }">
                <el-tag size="small">{{ resourceLabel(row.resource) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                {{ actionLabel(row.action) }}
              </template>
            </el-table-column>
            <el-table-column prop="detail" label="详情" min-width="200" show-overflow-tooltip />
            <el-table-column label="状态" width="80">
              <template #default="{ row }">
                <el-tag size="small" :type="row.status === 'success' ? 'success' : 'danger'">
                  {{ row.status === 'success' ? '成功' : '失败' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="opPagination.page"
            v-model:page-size="opPagination.pageSize"
            :total="opPagination.total"
            :page-sizes="[20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="loadOperationLogs"
            @current-change="loadOperationLogs"
            style="margin-top: 16px"
          />
        </el-card>
      </el-tab-pane>

      <!-- 系统日志 -->
      <el-tab-pane label="系统日志" name="system">
        <el-card>
          <template #header>
            <div class="card-header">
              <div class="header-left">
                <el-select v-model="sysSearch.logFile" style="width: 160px" @change="loadSystemLogs">
                  <el-option
                    v-for="f in sysFiles"
                    :key="f.name"
                    :label="f.name"
                    :value="f.name"
                  >
                    {{ f.name }} ({{ formatSize(f.size) }})
                  </el-option>
                </el-select>
              </div>
              <div class="header-right">
                <el-input
                  v-model="sysSearch.keyword"
                  placeholder="搜索关键词"
                  clearable
                  style="width: 240px"
                  @clear="loadSystemLogs"
                  @keyup.enter="loadSystemLogs"
                >
                  <template #prefix>
                    <el-icon><Search /></el-icon>
                  </template>
                </el-input>
                <el-switch
                  v-model="autoRefresh"
                  active-text="自动刷新"
                  @change="onAutoRefreshChange"
                />
                <el-button :icon="Refresh" circle @click="loadSystemLogs" />
              </div>
            </div>
          </template>

          <div class="log-viewer" v-loading="sysLoading">
            <div class="log-content">
              <div v-if="sysLines.length === 0" class="log-empty">暂无日志</div>
              <div v-for="(line, idx) in sysLines" :key="idx" class="log-line">
                <span class="line-no">{{ sysPagination.total - ((sysPagination.page - 1) * sysPagination.pageSize) - idx }}</span>
                <span class="line-text">{{ line }}</span>
              </div>
            </div>
          </div>

          <el-pagination
            v-model:current-page="sysPagination.page"
            v-model:page-size="sysPagination.pageSize"
            :total="sysPagination.total"
            :page-sizes="[50, 100, 200, 500]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="loadSystemLogs"
            @current-change="loadSystemLogs"
            style="margin-top: 16px"
          />
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 清理日志对话框 -->
    <el-dialog v-model="cleanDialogVisible" title="清理操作日志" width="420px">
      <el-form :model="cleanForm" label-width="100px">
        <el-form-item label="保留天数">
          <el-input-number v-model="cleanForm.retainDays" :min="1" :max="365" />
          <span style="margin-left: 8px; color: var(--el-text-color-secondary); font-size: 13px">
            将删除 {{ cleanForm.retainDays }} 天前的日志
          </span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="cleanDialogVisible = false">取消</el-button>
        <el-button type="danger" @click="doClean">确认清理</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Refresh, Delete } from '@element-plus/icons-vue'
import {
  searchOperationLogs,
  cleanOperationLogs,
  getOperationLogStats,
  searchSystemLogs,
  getSystemLogInfo
} from '@/api/modules/log'
import type { OperationLog, OperationLogStats, SystemLogFile } from '@/api/interface/log'

const activeTab = ref('operation')

// ========== 操作日志 ==========
const opLoading = ref(false)
const opLogs = ref<OperationLog[]>([])
const opStats = ref<OperationLogStats | null>(null)
const opDateRange = ref<string[]>([])
const opSearch = reactive({
  keyword: '',
  resource: '',
  status: ''
})
const opPagination = reactive({ page: 1, pageSize: 20, total: 0 })

const loadOperationLogs = async () => {
  opLoading.value = true
  try {
    const [logsRes, statsRes] = await Promise.all([
      searchOperationLogs({
        page: opPagination.page,
        pageSize: opPagination.pageSize,
        keyword: opSearch.keyword,
        resource: opSearch.resource,
        status: opSearch.status,
        startTime: opDateRange.value?.[0] || undefined,
        endTime: opDateRange.value?.[1] || undefined
      }),
      getOperationLogStats()
    ])
    opLogs.value = logsRes.data?.data?.items || []
    opPagination.total = logsRes.data?.data?.total || 0
    opStats.value = statsRes.data?.data || null
  } catch {
    ElMessage.error('加载操作日志失败')
  } finally {
    opLoading.value = false
  }
}

const onDateRangeChange = () => {
  opPagination.page = 1
  loadOperationLogs()
}

const resourceLabels: Record<string, string> = {
  auth: '认证', host: '主机', host_group: '主机分组', cronjob: '计划任务',
  firewall: '防火墙', process: '进程', ssh: 'SSH', setting: '设置',
  quick_command: '快捷命令', system: '系统', log: '日志'
}
const resourceLabel = (r: string) => resourceLabels[r] || r

const actionLabels: Record<string, string> = {
  login: '登录', create: '创建', update: '更新', delete: '删除',
  move: '移动', import: '导入', toggle: '切换', execute: '执行',
  operate: '操作', config: '配置', create_key: '创建密钥', delete_key: '删除密钥',
  port_rule: '端口规则', ip_rule: 'IP 规则', forward_rule: '转发规则',
  stop: '停止', clean: '清理', restart: '重启'
}
const actionLabel = (a: string) => actionLabels[a] || a

const formatTime = (t: string) => {
  if (!t) return ''
  return t.replace('T', ' ').replace(/\.\d+.*$/, '')
}

// 清理
const cleanDialogVisible = ref(false)
const cleanForm = reactive({ retainDays: 30 })

const openCleanDialog = () => {
  cleanDialogVisible.value = true
}

const doClean = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除 ${cleanForm.retainDays} 天前的操作日志吗？此操作不可撤销。`,
      '确认清理',
      { type: 'warning' }
    )
    const res = await cleanOperationLogs({ retainDays: cleanForm.retainDays })
    const deleted = res.data?.data?.deleted || 0
    ElMessage.success(`已清理 ${deleted} 条日志`)
    cleanDialogVisible.value = false
    loadOperationLogs()
  } catch {
    // cancelled
  }
}

// ========== 系统日志 ==========
const sysLoading = ref(false)
const sysLines = ref<string[]>([])
const sysFiles = ref<SystemLogFile[]>([])
const autoRefresh = ref(false)
let refreshTimer: ReturnType<typeof setInterval> | null = null
const sysSearch = reactive({
  keyword: '',
  logFile: 'agent.log'
})
const sysPagination = reactive({ page: 1, pageSize: 100, total: 0 })

const loadSystemLogs = async () => {
  sysLoading.value = true
  try {
    const res = await searchSystemLogs({
      page: sysPagination.page,
      pageSize: sysPagination.pageSize,
      keyword: sysSearch.keyword,
      logFile: sysSearch.logFile
    })
    sysLines.value = res.data?.data?.lines || []
    sysPagination.total = res.data?.data?.total || 0
  } catch {
    ElMessage.error('加载系统日志失败')
  } finally {
    sysLoading.value = false
  }
}

const loadSystemLogInfo = async () => {
  try {
    const res = await getSystemLogInfo()
    sysFiles.value = res.data?.data?.files || []
    if (sysFiles.value.length > 0 && !sysFiles.value.find(f => f.name === sysSearch.logFile)) {
      sysSearch.logFile = sysFiles.value[0].name
    }
  } catch {
    // ignore
  }
}

const formatSize = (bytes: number): string => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

const onAutoRefreshChange = (val: boolean) => {
  if (val) {
    refreshTimer = setInterval(() => {
      loadSystemLogs()
      loadSystemLogInfo()
    }, 5000)
  } else if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

const onTabChange = (tab: string) => {
  if (tab === 'operation') {
    loadOperationLogs()
  } else {
    loadSystemLogInfo()
    loadSystemLogs()
  }
}

onMounted(() => {
  loadOperationLogs()
})

onBeforeUnmount(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.log-page {
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
  align-items: center;
}

.header-right {
  display: flex;
  gap: 10px;
  align-items: center;
}

.stats-info {
  display: flex;
  gap: 6px;
}

.log-viewer {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 6px;
  background: #1e1e1e;
  min-height: 400px;
  max-height: 600px;
  overflow: auto;
}

.log-content {
  padding: 12px;
  font-family: 'Cascadia Code', 'Fira Code', 'Consolas', 'Monaco', monospace;
  font-size: 12.5px;
  line-height: 1.6;
}

.log-empty {
  color: #666;
  text-align: center;
  padding: 40px 0;
}

.log-line {
  display: flex;
  gap: 12px;
  white-space: pre-wrap;
  word-break: break-all;
}

.log-line:hover {
  background: rgba(255, 255, 255, 0.05);
}

.line-no {
  color: #666;
  min-width: 48px;
  text-align: right;
  user-select: none;
  flex-shrink: 0;
}

.line-text {
  color: #d4d4d4;
}
</style>

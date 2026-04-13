<template>
  <div class="process-page">
    <!-- Tab 切换 -->
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <el-tab-pane label="进程列表" name="process" />
      <el-tab-pane label="网络连接" name="network" />
    </el-tabs>

    <!-- 进程列表 -->
    <div v-show="activeTab === 'process'">
      <!-- 搜索栏 -->
      <div class="toolbar">
        <el-select
          v-model="psFilters"
          placeholder="状态过滤"
          clearable
          multiple
          collapse-tags
          collapse-tags-tooltip
          :max-collapse-tags="3"
          style="width: 280px"
          @change="refreshProcesses"
        >
          <el-option v-for="s in statusOptions" :key="s.value" :label="s.label" :value="s.value" />
        </el-select>
        <el-input
          v-model="psSearch.pid"
          placeholder="PID"
          clearable
          style="width: 120px"
          @keyup.enter="refreshProcesses"
          @clear="refreshProcesses"
        />
        <el-input
          v-model="psSearch.name"
          placeholder="进程名称"
          clearable
          style="width: 180px"
          @keyup.enter="refreshProcesses"
          @clear="refreshProcesses"
        />
        <el-input
          v-model="psSearch.username"
          placeholder="用户"
          clearable
          style="width: 140px"
          @keyup.enter="refreshProcesses"
          @clear="refreshProcesses"
        />
        <el-button :icon="Search" @click="refreshProcesses">搜索</el-button>
        <el-button :icon="Refresh" @click="resetPsSearch">重置</el-button>
      </div>

      <!-- 进程表格 -->
      <el-card shadow="hover" class="table-card">
        <el-table
          :data="filteredProcesses"
          v-loading="psLoading"
          stripe
          :default-sort="{ prop: 'cpuValue', order: 'descending' }"
          @sort-change="handlePsSort"
          max-height="calc(100vh - 280px)"
        >
          <el-table-column prop="PID" label="PID" width="90" sortable="custom" />
          <el-table-column prop="name" label="名称" min-width="200" show-overflow-tooltip />
          <el-table-column prop="PPID" label="PPID" width="90" />
          <el-table-column prop="numThreads" label="线程数" width="80" />
          <el-table-column prop="username" label="用户" width="120" show-overflow-tooltip />
          <el-table-column prop="cpuValue" label="CPU" width="100" sortable="custom">
            <template #default="{ row }">{{ row.cpuPercent }}</template>
          </el-table-column>
          <el-table-column prop="rssValue" label="内存" width="110" sortable="custom">
            <template #default="{ row }">{{ row.rss }}</template>
          </el-table-column>
          <el-table-column prop="numConnections" label="连接数" width="80" sortable="custom" />
          <el-table-column prop="status" label="状态" width="80">
            <template #default="{ row }">
              <span>{{ statusLabel(row.status) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="startTime" label="启动时间" width="170" />
          <el-table-column label="操作" width="150" fixed="right">
            <template #default="{ row }">
              <el-button type="primary" link size="small" @click="openDetail(row.PID)">
                详情
              </el-button>
              <el-button type="danger" link size="small" @click="confirmStop(row)">
                终止
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>

    <!-- 网络连接 -->
    <div v-show="activeTab === 'network'">
      <!-- 搜索栏 -->
      <div class="toolbar">
        <el-select
          v-model="netFilters"
          placeholder="状态过滤"
          clearable
          multiple
          collapse-tags
          collapse-tags-tooltip
          :max-collapse-tags="3"
          style="width: 320px"
          @change="refreshNet"
        >
          <el-option v-for="s in netStatusOptions" :key="s" :label="s" :value="s" />
        </el-select>
        <el-input
          v-model="netSearch.processID"
          placeholder="进程 ID"
          clearable
          style="width: 120px"
          @keyup.enter="refreshNet"
          @clear="refreshNet"
        />
        <el-input
          v-model="netSearch.processName"
          placeholder="进程名称"
          clearable
          style="width: 160px"
          @keyup.enter="refreshNet"
          @clear="refreshNet"
        />
        <el-input
          v-model="netSearch.port"
          placeholder="端口"
          clearable
          style="width: 120px"
          @keyup.enter="refreshNet"
          @clear="refreshNet"
        />
        <el-button :icon="Search" @click="refreshNet">搜索</el-button>
        <el-button :icon="Refresh" @click="resetNetSearch">重置</el-button>
      </div>

      <!-- 网络连接表格 -->
      <el-card shadow="hover" class="table-card">
        <el-table
          :data="filteredNet"
          v-loading="netLoading"
          stripe
          max-height="calc(100vh - 280px)"
        >
          <el-table-column prop="type" label="类型" width="100" />
          <el-table-column prop="PID" label="PID" width="90" sortable />
          <el-table-column prop="name" label="进程名" width="180" show-overflow-tooltip>
            <template #default="{ row }">
              <span class="clickable-text" @click="quickSearchByName(row.name)">{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="localaddr" label="本地地址" min-width="220" show-overflow-tooltip>
            <template #default="{ row }">
              <span>{{ row.localaddr }}</span>
              <el-button
                v-if="extractPort(row.localaddr)"
                :icon="Search"
                link
                size="small"
                style="margin-left: 4px"
                @click="quickSearchByPort(extractPort(row.localaddr))"
              />
            </template>
          </el-table-column>
          <el-table-column prop="remoteaddr" label="远程地址" min-width="220" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="140" />
        </el-table>
      </el-card>
    </div>

    <!-- 终止进程确认对话框 -->
    <el-dialog v-model="stopDialogVisible" title="终止进程" width="440px" :close-on-click-modal="false">
      <p>
        确定要终止进程
        <strong>{{ stopTarget?.name }}</strong>
        (PID: {{ stopTarget?.PID }}) 吗？
      </p>
      <template #footer>
        <el-button @click="stopDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="stopLoading" @click="doStop">确认终止</el-button>
      </template>
    </el-dialog>

    <!-- 进程详情抽屉 -->
    <el-drawer v-model="detailVisible" :title="detailData?.name ?? '进程详情'" size="680px" direction="rtl">
      <div v-loading="detailLoading">
        <el-tabs v-if="detailData" v-model="detailTab">
          <!-- 基本信息 -->
          <el-tab-pane label="基本信息" name="basic">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="进程名称">{{ detailData.name }}</el-descriptions-item>
              <el-descriptions-item label="状态">{{ statusLabel(detailData.status) }}</el-descriptions-item>
              <el-descriptions-item label="PID">{{ detailData.PID }}</el-descriptions-item>
              <el-descriptions-item label="PPID">{{ detailData.PPID }}</el-descriptions-item>
              <el-descriptions-item label="线程数">{{ detailData.numThreads }}</el-descriptions-item>
              <el-descriptions-item label="连接数">{{ detailData.numConnections }}</el-descriptions-item>
              <el-descriptions-item label="CPU">{{ detailData.cpuPercent }}</el-descriptions-item>
              <el-descriptions-item label="内存">{{ detailData.rss }}</el-descriptions-item>
              <el-descriptions-item label="磁盘读取">{{ detailData.diskRead }}</el-descriptions-item>
              <el-descriptions-item label="磁盘写入">{{ detailData.diskWrite }}</el-descriptions-item>
              <el-descriptions-item label="用户" :span="2">{{ detailData.username }}</el-descriptions-item>
              <el-descriptions-item label="启动时间" :span="2">{{ detailData.startTime }}</el-descriptions-item>
              <el-descriptions-item label="命令行" :span="2">
                <div class="cmd-line">{{ detailData.cmdLine }}</div>
              </el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <!-- 内存信息 -->
          <el-tab-pane label="内存信息" name="mem">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="RSS (物理内存)">{{ detailData.rss }}</el-descriptions-item>
              <el-descriptions-item label="PSS (比例共享)">{{ detailData.pss }}</el-descriptions-item>
              <el-descriptions-item label="USS (独占内存)">{{ detailData.uss }}</el-descriptions-item>
              <el-descriptions-item label="Swap (交换空间)">{{ detailData.swap }}</el-descriptions-item>
              <el-descriptions-item label="Shared (共享内存)">{{ detailData.shared }}</el-descriptions-item>
              <el-descriptions-item label="VMS (虚拟内存)">{{ detailData.vms }}</el-descriptions-item>
              <el-descriptions-item label="HWM (高水位)">{{ detailData.hwm }}</el-descriptions-item>
              <el-descriptions-item label="Data (数据段)">{{ detailData.data }}</el-descriptions-item>
              <el-descriptions-item label="Stack (栈)">{{ detailData.stack }}</el-descriptions-item>
              <el-descriptions-item label="Locked (锁定)">{{ detailData.locked }}</el-descriptions-item>
              <el-descriptions-item label="Text (代码段)">{{ detailData.text }}</el-descriptions-item>
              <el-descriptions-item label="Dirty (脏页)">{{ detailData.dirty }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <!-- 打开文件 -->
          <el-tab-pane label="打开文件" name="files">
            <el-table :data="detailData.openFiles" stripe max-height="500">
              <el-table-column prop="path" label="文件路径" show-overflow-tooltip />
              <el-table-column prop="fd" label="文件描述符" width="120" />
            </el-table>
            <el-empty v-if="!detailData.openFiles?.length" description="暂无数据" />
          </el-tab-pane>

          <!-- 环境变量 -->
          <el-tab-pane label="环境变量" name="env">
            <el-input
              type="textarea"
              :model-value="envText"
              readonly
              :autosize="{ minRows: 10, maxRows: 30 }"
              class="env-textarea"
            />
            <el-empty v-if="!detailData.envs?.length" description="暂无数据" />
          </el-tab-pane>

          <!-- 网络连接 -->
          <el-tab-pane label="网络连接" name="connects">
            <el-table :data="detailData.connects" stripe max-height="500">
              <el-table-column prop="localaddr" label="本地地址" min-width="200" show-overflow-tooltip />
              <el-table-column prop="remoteaddr" label="远程地址" min-width="200" show-overflow-tooltip />
              <el-table-column prop="status" label="状态" width="140" />
            </el-table>
            <el-empty v-if="!detailData.connects?.length" description="暂无数据" />
          </el-tab-pane>
        </el-tabs>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Search, Refresh } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import {
  listProcesses,
  getProcessDetail,
  stopProcess,
  listNetConnections,
} from '@/api/modules/process'
import type { ProcessInfo, ProcessDetail, NetConnection } from '@/api/interface/process'

// ======================== 通用 ========================

const activeTab = ref('process')
let pollingTimer: ReturnType<typeof setInterval> | null = null

const statusOptions = [
  { label: 'Running', value: 'running' },
  { label: 'Sleep', value: 'sleep' },
  { label: 'Stop', value: 'stop' },
  { label: 'Idle', value: 'idle' },
  { label: 'Wait', value: 'wait' },
  { label: 'Lock', value: 'lock' },
  { label: 'Zombie', value: 'zombie' },
]

const netStatusOptions = ['LISTEN', 'ESTABLISHED', 'TIME_WAIT', 'CLOSE_WAIT', 'NONE']

const statusMap: Record<string, string> = {
  running: '运行中',
  sleep: '睡眠',
  stop: '停止',
  idle: '空闲',
  wait: '等待',
  lock: '锁定',
  zombie: '僵尸',
}

function statusLabel(status: string): string {
  return statusMap[status] || status || '-'
}

// ======================== 进程列表 ========================

const psLoading = ref(false)
const psData = ref<ProcessInfo[]>([])
const psFilters = ref<string[]>([])
const psSearch = ref({ pid: '', name: '', username: '' })
const psSortProp = ref('cpuValue')
const psSortOrder = ref<'ascending' | 'descending'>('descending')

const filteredProcesses = computed(() => {
  let list = psData.value
  if (psFilters.value.length > 0) {
    list = list.filter((p) => psFilters.value.includes(p.status))
  }
  if (psSortProp.value) {
    const prop = psSortProp.value as keyof ProcessInfo
    const dir = psSortOrder.value === 'ascending' ? 1 : -1
    list = [...list].sort((a, b) => {
      const av = Number(a[prop]) || 0
      const bv = Number(b[prop]) || 0
      return (av - bv) * dir
    })
  }
  return list
})

async function refreshProcesses() {
  psLoading.value = psData.value.length === 0
  try {
    const pidNum = psSearch.value.pid ? Number(psSearch.value.pid) : undefined
    const res = await listProcesses({
      pid: pidNum,
      name: psSearch.value.name || undefined,
      username: psSearch.value.username || undefined,
    })
    psData.value = res.data?.data ?? []
  } catch {
    ElMessage.error('获取进程列表失败')
  } finally {
    psLoading.value = false
  }
}

function resetPsSearch() {
  psSearch.value = { pid: '', name: '', username: '' }
  psFilters.value = []
  refreshProcesses()
}

function handlePsSort({ prop, order }: { prop: string; order: string | null }) {
  psSortProp.value = prop
  psSortOrder.value = (order as 'ascending' | 'descending') || 'ascending'
}

// ======================== 网络连接 ========================

const netLoading = ref(false)
const netData = ref<NetConnection[]>([])
const netFilters = ref<string[]>(['LISTEN', 'ESTABLISHED'])
const netSearch = ref({ processID: '', processName: '', port: '' })

const filteredNet = computed(() => {
  if (netFilters.value.length === 0) return netData.value
  return netData.value.filter((n) => netFilters.value.includes(n.status))
})

async function refreshNet() {
  netLoading.value = netData.value.length === 0
  try {
    const res = await listNetConnections({
      processID: netSearch.value.processID ? Number(netSearch.value.processID) : undefined,
      processName: netSearch.value.processName || undefined,
      port: netSearch.value.port ? Number(netSearch.value.port) : undefined,
    })
    netData.value = res.data?.data ?? []
  } catch {
    ElMessage.error('获取网络连接失败')
  } finally {
    netLoading.value = false
  }
}

function resetNetSearch() {
  netSearch.value = { processID: '', processName: '', port: '' }
  netFilters.value = ['LISTEN', 'ESTABLISHED']
  refreshNet()
}

function quickSearchByName(name: string) {
  netSearch.value.processID = ''
  netSearch.value.processName = name
  netSearch.value.port = ''
  refreshNet()
}

function quickSearchByPort(port: number) {
  netSearch.value.processID = ''
  netSearch.value.processName = ''
  netSearch.value.port = String(port)
  refreshNet()
}

function extractPort(addr: string): number {
  const parts = addr.split(':')
  const portStr = parts[parts.length - 1]
  const port = Number(portStr)
  return port > 0 ? port : 0
}

// ======================== 终止进程 ========================

const stopDialogVisible = ref(false)
const stopLoading = ref(false)
const stopTarget = ref<ProcessInfo | null>(null)

function confirmStop(row: ProcessInfo) {
  stopTarget.value = row
  stopDialogVisible.value = true
}

async function doStop() {
  if (!stopTarget.value) return
  stopLoading.value = true
  try {
    await stopProcess({ PID: stopTarget.value.PID })
    ElMessage.success(`进程 ${stopTarget.value.name} 已终止`)
    stopDialogVisible.value = false
    refreshProcesses()
  } catch {
    ElMessage.error('终止进程失败')
  } finally {
    stopLoading.value = false
  }
}

// ======================== 进程详情 ========================

const detailVisible = ref(false)
const detailLoading = ref(false)
const detailData = ref<ProcessDetail | null>(null)
const detailTab = ref('basic')

const envText = computed(() => {
  return detailData.value?.envs?.join('\n') ?? ''
})

async function openDetail(pid: number) {
  detailVisible.value = true
  detailLoading.value = true
  detailTab.value = 'basic'
  detailData.value = null
  try {
    const res = await getProcessDetail(pid)
    detailData.value = res.data?.data ?? null
  } catch {
    ElMessage.error('获取进程详情失败')
  } finally {
    detailLoading.value = false
  }
}

// ======================== Tab 切换与轮询 ========================

function handleTabChange(tab: string) {
  stopPolling()
  if (tab === 'process') {
    refreshProcesses()
  } else {
    refreshNet()
  }
  startPolling()
}

function startPolling() {
  stopPolling()
  pollingTimer = setInterval(() => {
    if (activeTab.value === 'process') {
      refreshProcesses()
    } else {
      refreshNet()
    }
  }, 3000)
}

function stopPolling() {
  if (pollingTimer) {
    clearInterval(pollingTimer)
    pollingTimer = null
  }
}

onMounted(() => {
  refreshProcesses()
  startPolling()
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.process-page {
  min-height: calc(100vh - 60px);
  padding: 1.25rem;
  background: var(--bg-color);
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  margin-bottom: 16px;
}

.table-card {
  margin-bottom: 16px;
}

.clickable-text {
  color: var(--el-color-primary);
  cursor: pointer;
}
.clickable-text:hover {
  text-decoration: underline;
}

.cmd-line {
  word-break: break-all;
  white-space: pre-wrap;
  max-height: 120px;
  overflow-y: auto;
  font-family: monospace;
  font-size: 12px;
  line-height: 1.5;
}

.env-textarea :deep(.el-textarea__inner) {
  font-family: monospace;
  font-size: 12px;
  line-height: 1.6;
}
</style>

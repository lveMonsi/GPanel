<template>
  <div class="hosts-container">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-input
          v-model="searchInfo"
          placeholder="搜索主机名称、地址、用户"
          style="width: 250px"
          clearable
          @clear="handleSearch"
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-select
          v-model="selectedGroupID"
          placeholder="选择分组"
          style="width: 150px; margin-left: 10px"
          clearable
          @change="handleSearch"
        >
          <el-option label="全部分组" :value="0" />
          <el-option
            v-for="group in groups"
            :key="group.id"
            :label="group.name"
            :value="group.id"
          />
        </el-select>
      </div>
      <div class="toolbar-right">
        <el-button type="primary" @click="handleAddGroup">
          <el-icon><FolderAdd /></el-icon>
          添加分组
        </el-button>
        <el-button type="primary" @click="handleAddHost">
          <el-icon><Plus /></el-icon>
          添加主机
        </el-button>
        <el-button type="success" @click="handleExport">
          <el-icon><Download /></el-icon>
          导出
        </el-button>
        <el-button type="success" @click="handleImport">
          <el-icon><Upload /></el-icon>
          导入
        </el-button>
      </div>
    </div>

    <!-- 主机列表 -->
    <el-card shadow="hover" class="table-card">
      <el-table
        :data="hosts"
        v-loading="loading"
        stripe
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column prop="name" label="主机名称" min-width="150" />
        <el-table-column prop="groupName" label="所属分组" width="120" />
        <el-table-column prop="addr" label="主机地址" min-width="150" />
        <el-table-column prop="port" label="端口" width="80" />
        <el-table-column prop="user" label="用户名" width="100" />
        <el-table-column prop="authMode" label="认证方式" width="100">
          <template #default="{ row }">
            <el-tag :type="row.authMode === 'password' ? 'primary' : 'success'" size="small">
              {{ row.authMode === 'password' ? '密码' : '密钥' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="handleConnect(row)">
              <el-icon><Connection /></el-icon>
              连接
            </el-button>
            <el-button type="primary" link size="small" @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button type="danger" link size="small" @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadHosts"
          @current-change="loadHosts"
        />
      </div>
    </el-card>

    <!-- 批量操作栏 -->
    <div v-if="selectedHosts.length > 0" class="batch-actions">
      <span>已选择 {{ selectedHosts.length }} 项</span>
      <el-button type="primary" size="small" @click="handleMoveToGroup">
        移动到分组
      </el-button>
    </div>

    <!-- 主机操作对话框 -->
    <HostOperateDialog
      v-model:visible="hostDialogVisible"
      :host="currentHost"
      :groups="groups"
      @save="handleHostSave"
    />

    <!-- 分组操作对话框 -->
    <HostGroupOperateDialog
      v-model:visible="groupDialogVisible"
      :group="currentGroup"
      @save="handleGroupSave"
    />

    <!-- 移动到分组对话框 -->
    <el-dialog
      v-model="moveDialogVisible"
      title="移动到分组"
      width="400px"
    >
      <el-form label-width="80px">
        <el-form-item label="目标分组">
          <el-select v-model="targetGroupID" placeholder="请选择分组" style="width: 100%">
            <el-option
              v-for="group in groups"
              :key="group.id"
              :label="group.name"
              :value="group.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="moveDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmMove">确定</el-button>
      </template>
    </el-dialog>

    <!-- 导出确认对话框 -->
    <el-dialog
      v-model="exportDialogVisible"
      title="导出主机配置"
      width="450px"
    >
      <div class="export-options">
        <el-checkbox v-model="exportEncrypted">
          加密导出敏感信息（密码、密钥等）
        </el-checkbox>
        <div class="export-tips">
          <p v-if="exportEncrypted">
            <el-icon><InfoFilled /></el-icon>
            敏感信息会以加密形式存储，导入时自动解密
          </p>
          <p v-else>
            <el-icon><WarningFilled /></el-icon>
            敏感信息以明文存储，便于查看和跨平台迁移，但请注意安全
          </p>
        </div>
      </div>
      <template #footer>
        <el-button @click="exportDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfirmExport">确定导出</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search,
  FolderAdd,
  Plus,
  Connection,
  Edit,
  Delete,
  Download,
  Upload,
  InfoFilled,
  WarningFilled
} from '@element-plus/icons-vue'
import {
  listHosts,
  deleteHost,
  createHost,
  updateHost,
  moveHosts,
  listGroups,
  deleteGroup,
  createGroup,
  updateGroup,
  exportHosts,
  importHosts
} from '@/api/modules/host'
import type { HostInfo, HostGroupInfo, HostOperate, HostGroupOperate } from '@/api/interface/host'
import HostOperateDialog from '@/components/HostOperate.vue'
import HostGroupOperateDialog from '@/components/HostGroupOperate.vue'

const router = useRouter()

const loading = ref(false)
const searchInfo = ref('')
const selectedGroupID = ref(0)
const hosts = ref<HostInfo[]>([])
const groups = ref<HostGroupInfo[]>([])
const selectedHosts = ref<HostInfo[]>([])

const pagination = ref({
  page: 1,
  pageSize: 10,
  total: 0
})

const hostDialogVisible = ref(false)
const groupDialogVisible = ref(false)
const moveDialogVisible = ref(false)
const currentHost = ref<HostOperate>()
const currentGroup = ref<HostGroupOperate>()
const targetGroupID = ref(0)

const loadHosts = async () => {
  loading.value = true
  try {
    const res = await listHosts({
      page: pagination.value.page,
      pageSize: pagination.value.pageSize,
      groupID: selectedGroupID.value,
      info: searchInfo.value
    })
    hosts.value = res.data.data
    pagination.value.total = res.data.total
  } catch (error) {
    console.error('加载主机列表失败:', error)
  } finally {
    loading.value = false
  }
}

const loadGroups = async () => {
  try {
    const res = await listGroups({
      page: 1,
      pageSize: 100
    })
    groups.value = res.data.data
  } catch (error) {
    console.error('加载分组列表失败:', error)
  }
}

const handleSearch = () => {
  pagination.value.page = 1
  loadHosts()
}

const handleSelectionChange = (selection: HostInfo[]) => {
  selectedHosts.value = selection
}

const handleAddHost = () => {
  currentHost.value = undefined
  hostDialogVisible.value = true
}

const handleEdit = (host: HostInfo) => {
  currentHost.value = { ...host }
  hostDialogVisible.value = true
}

const handleDelete = async (host: HostInfo) => {
  try {
    await ElMessageBox.confirm(`确定要删除主机 "${host.name}" 吗？`, '确认删除', {
      type: 'warning'
    })
    await deleteHost(host.id)
    ElMessage.success('删除成功')
    loadHosts()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
    }
  }
}

const handleConnect = (host: HostInfo) => {
  router.push({
    path: '/terminal',
    query: { hostId: host.id }
  })
}

const handleHostSave = async (data: HostOperate) => {
  try {
    if (data.id) {
      await updateHost(data.id, data)
      ElMessage.success('更新成功')
    } else {
      await createHost(data)
      ElMessage.success('创建成功')
    }
    loadHosts()
  } catch (error) {
    console.error('保存失败:', error)
  }
}

const handleAddGroup = () => {
  currentGroup.value = undefined
  groupDialogVisible.value = true
}

const handleGroupSave = async (data: HostGroupOperate) => {
  try {
    if (data.id) {
      await updateGroup(data.id, data)
      ElMessage.success('更新成功')
    } else {
      await createGroup(data)
      ElMessage.success('创建成功')
    }
    loadGroups()
    loadHosts()
  } catch (error) {
    console.error('保存失败:', error)
  }
}

const handleMoveToGroup = () => {
  if (selectedHosts.value.length === 0) {
    ElMessage.warning('请先选择要移动的主机')
    return
  }
  targetGroupID.value = 0
  moveDialogVisible.value = true
}

const handleConfirmMove = async () => {
  if (targetGroupID.value === 0) {
    ElMessage.warning('请选择目标分组')
    return
  }
  try {
    await moveHosts({
      hostIDs: selectedHosts.value.map(h => h.id),
      groupID: targetGroupID.value
    })
    ElMessage.success('移动成功')
    moveDialogVisible.value = false
    loadHosts()
  } catch (error) {
    console.error('移动失败:', error)
  }
}

const exportDialogVisible = ref(false)
const exportEncrypted = ref(true)

const handleExport = () => {
  exportEncrypted.value = true
  exportDialogVisible.value = true
}

const handleConfirmExport = async () => {
  try {
    const res = await exportHosts(exportEncrypted.value)
    const dataStr = JSON.stringify(res.data, null, 2)
    const dataBlob = new Blob([dataStr], { type: 'application/json' })
    const url = URL.createObjectURL(dataBlob)
    const link = document.createElement('a')
    link.href = url
    link.download = `hosts_${new Date().toISOString().slice(0, 10)}.json`
    link.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
    exportDialogVisible.value = false
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败')
  }
}

const handleImport = () => {
  const input = document.createElement('input')
  input.type = 'file'
  input.accept = '.json'
  input.onchange = async (e: Event) => {
    const target = e.target as HTMLInputElement
    const file = target.files?.[0]
    if (!file) return

    try {
      const text = await file.text()
      const data = JSON.parse(text)
      const res = await importHosts(data)
      ElMessage.success(res.data.message)
      loadHosts()
    } catch (error) {
      console.error('导入失败:', error)
      ElMessage.error('导入失败，请检查文件格式')
    }
  }
  input.click()
}

onMounted(() => {
  loadHosts()
  loadGroups()
})
</script>

<style scoped>
.hosts-container {
  padding: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.toolbar-right {
  display: flex;
  gap: 10px;
}

.table-card {
  min-height: 400px;
}

.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.batch-actions {
  position: fixed;
  bottom: 30px;
  left: 50%;
  transform: translateX(-50%);
  background: var(--el-bg-color);
  padding: 12px 24px;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  display: flex;
  align-items: center;
  gap: 16px;
  z-index: 100;
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    transform: translateX(-50%) translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateX(-50%) translateY(0);
    opacity: 1;
  }
}

.export-options {
  padding: 10px 0;
}

.export-options .el-checkbox {
  font-size: 14px;
}

.export-tips {
  margin-top: 12px;
  padding: 10px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
}

.export-tips p {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.export-tips .el-icon {
  font-size: 14px;
}
</style>
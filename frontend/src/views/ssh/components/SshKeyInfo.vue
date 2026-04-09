<template>
  <el-dialog v-model="visible" title="密钥信息" width="980px">
    <div class="toolbar">
      <div>
        <el-button type="primary" @click="openCreate">新增</el-button>
        <el-button @click="handleSync" :loading="syncing">同步</el-button>
        <el-button type="danger" :disabled="selectedIds.length === 0" @click="handleBatchDelete">删除</el-button>
      </div>
      <el-button @click="loadData" :loading="loading">刷新</el-button>
    </div>

    <el-table :data="keys" v-loading="loading" border stripe @selection-change="handleSelectionChange">
      <el-table-column type="selection" width="55" />
      <el-table-column prop="name" label="名称" min-width="140" />
      <el-table-column prop="encryptionMode" label="加密方式" width="120" />
      <el-table-column prop="description" label="描述" min-width="180" show-overflow-tooltip />
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" size="small" @click="openEdit(row)">编辑</el-button>
          <el-button link type="primary" size="small" @click="openView(row)">查看</el-button>
          <el-button link type="danger" size="small" @click="handleDelete([row.id])">删除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @current-change="loadData"
        @size-change="loadData"
      />
    </div>

    <el-dialog v-model="viewVisible" title="查看密钥" width="760px" append-to-body>
      <el-descriptions :column="1" border>
        <el-descriptions-item label="私钥密码">
          <div class="view-actions">
            <span>{{ currentKey?.passPhrase || '-' }}</span>
            <el-button v-if="currentKey?.passPhrase" size="small" @click="copyText(currentKey.passPhrase)">复制</el-button>
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="公钥">
          <div class="view-actions">
            <el-button size="small" @click="copyText(currentKey?.publicKey || '')">复制</el-button>
            <el-button size="small" @click="downloadText(`${currentKey?.name || 'key'}.pub`, currentKey?.publicKey || '')">下载</el-button>
          </div>
        </el-descriptions-item>
        <el-descriptions-item label="私钥">
          <div class="view-actions">
            <el-button size="small" @click="copyText(currentKey?.privateKey || '')">复制</el-button>
            <el-button size="small" @click="downloadText(currentKey?.name || 'key', currentKey?.privateKey || '')">下载</el-button>
          </div>
        </el-descriptions-item>
      </el-descriptions>
    </el-dialog>

    <SshKeyOperate ref="operateRef" @saved="loadData" />
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { deleteSSHKeys, getSSHKeys, syncSSHKeys } from '@/api/modules/ssh';
import type { SSHKeyInfo } from '@/api/interface/ssh';
import SshKeyOperate from './SshKeyOperate.vue';

const visible = ref(false);
const viewVisible = ref(false);
const loading = ref(false);
const syncing = ref(false);
const keys = ref<SSHKeyInfo[]>([]);
const selectedIds = ref<number[]>([]);
const currentKey = ref<SSHKeyInfo>();
const operateRef = ref<InstanceType<typeof SshKeyOperate>>();
const pagination = reactive({
  page: 1,
  pageSize: 10,
  total: 0,
});

const open = async () => {
  visible.value = true;
  await loadData();
};

const loadData = async () => {
  loading.value = true;
  try {
    const res = await getSSHKeys({ page: pagination.page, pageSize: pagination.pageSize });
    keys.value = res.data?.items || [];
    pagination.total = res.data?.total || 0;
  } catch (error: any) {
    ElMessage.error(error.message || '加载密钥信息失败');
  } finally {
    loading.value = false;
  }
};

const handleSelectionChange = (rows: SSHKeyInfo[]) => {
  selectedIds.value = rows.map(item => item.id);
};

const openCreate = () => {
  operateRef.value?.open();
};

const openEdit = (row: SSHKeyInfo) => {
  operateRef.value?.open(row);
};

const openView = (row: SSHKeyInfo) => {
  currentKey.value = row;
  viewVisible.value = true;
};

const handleDelete = async (ids: number[]) => {
  try {
    await ElMessageBox.confirm('确认删除选中的密钥信息吗？', '确认删除', { type: 'warning' });
    await deleteSSHKeys({ ids });
    ElMessage.success('删除成功');
    if (selectedIds.value.length) {
      selectedIds.value = [];
    }
    await loadData();
  } catch (error: any) {
    if (error !== 'cancel' && error?.message) {
      ElMessage.error(error.message);
    }
  }
};

const handleBatchDelete = async () => {
  await handleDelete(selectedIds.value);
};

const handleSync = async () => {
  syncing.value = true;
  try {
    await syncSSHKeys();
    ElMessage.success('同步成功');
    await loadData();
  } catch (error: any) {
    ElMessage.error(error.message || '同步失败');
  } finally {
    syncing.value = false;
  }
};

const copyText = async (text: string) => {
  if (!text) return;
  try {
    await navigator.clipboard.writeText(text);
    ElMessage.success('复制成功');
  } catch {
    ElMessage.error('复制失败');
  }
};

const downloadText = (filename: string, content: string) => {
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' });
  const link = document.createElement('a');
  link.href = URL.createObjectURL(blob);
  link.download = filename;
  link.click();
  URL.revokeObjectURL(link.href);
};

defineExpose({ open });
</script>

<style scoped>
.toolbar {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16px;
}
.pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
}
.view-actions {
  display: flex;
  gap: 8px;
  align-items: center;
}
</style>

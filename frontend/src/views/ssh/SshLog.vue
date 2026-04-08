<template>
  <div class="ssh-log">
    <div class="toolbar">
      <el-select v-model="searchStatus" @change="loadLogs" style="width: 120px">
        <el-option label="全部" value="all" />
        <el-option label="成功" value="success" />
        <el-option label="失败" value="failed" />
      </el-select>
      <el-input
        v-model="searchText"
        placeholder="搜索IP或用户"
        clearable
        style="width: 200px; margin-left: 8px"
        @input="loadLogs"
      />
      <el-button :icon="Refresh" @click="loadLogs" style="margin-left: 8px">刷新</el-button>
    </div>

    <el-table :data="logs" v-loading="loading" border stripe>
      <el-table-column label="登录IP" prop="address" min-width="140" />
      <el-table-column label="端口" prop="port" width="80" />
      <el-table-column label="用户" prop="user" min-width="100" />
      <el-table-column label="认证方式" prop="authMode" width="100">
        <template #default="{ row }">
          {{ row.authMode === 'password' ? '密码' : '密钥' }}
        </template>
      </el-table-column>
      <el-table-column label="状态" prop="status" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
            {{ row.status === 'success' ? '成功' : '失败' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="时间" prop="date" min-width="160" />
    </el-table>

    <el-pagination
      v-model:current-page="pagination.page"
      v-model:page-size="pagination.pageSize"
      :total="pagination.total"
      :page-sizes="[20, 50, 100]"
      layout="total, sizes, prev, pager, next"
      @current-change="loadLogs"
      @size-change="loadLogs"
      style="margin-top: 16px; justify-content: flex-end"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { Refresh } from '@element-plus/icons-vue';
import { getSSHLogs } from '@/api/modules/ssh';
import type { SSHLogItem } from '@/api/interface/ssh';

const loading = ref(false);
const searchStatus = ref('all');
const searchText = ref('');
const logs = ref<SSHLogItem[]>([]);
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0,
});

const loadLogs = async () => {
  loading.value = true;
  try {
    const res = await getSSHLogs({
      page: pagination.page,
      pageSize: pagination.pageSize,
      status: searchStatus.value,
      info: searchText.value,
    });
    logs.value = res.data?.items || [];
    pagination.total = res.data?.total || 0;
  } finally {
    loading.value = false;
  }
};

onMounted(() => {
  loadLogs();
});
</script>

<style scoped>
.ssh-log {
  padding: 20px;
}
.toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}
</style>

<template>
  <div class="ssh-session">
    <div class="toolbar">
      <el-input
        v-model="searchUser"
        placeholder="搜索用户"
        clearable
        style="width: 200px"
        @input="filterData"
      />
      <el-button :icon="Refresh" @click="loadSessions" style="margin-left: 8px">刷新</el-button>
    </div>

    <el-table :data="filteredData" v-loading="loading" border stripe>
      <el-table-column label="用户" prop="username" min-width="100" />
      <el-table-column label="TTY" prop="terminal" min-width="80" />
      <el-table-column label="登录IP" prop="host" min-width="140" />
      <el-table-column label="登录时间" prop="loginTime" min-width="160" />
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button type="danger" size="small" link @click="handleDisconnect(row)">
            断开
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Refresh } from '@element-plus/icons-vue';
import { getSSHSessions, killSSHSession } from '@/api/modules/ssh';

interface SessionItem {
  pid: number;
  username: string;
  terminal: string;
  host: string;
  loginTime: string;
}

const loading = ref(false);
const searchUser = ref('');
const sessions = ref<SessionItem[]>([]);
let timer: ReturnType<typeof setInterval> | null = null;

const filteredData = computed(() =>
  searchUser.value
    ? sessions.value.filter((s) => s.username.includes(searchUser.value))
    : sessions.value
);

const filterData = () => {};

const loadSessions = async () => {
  loading.value = true;
  try {
    const res = await getSSHSessions();
    sessions.value = res.data || [];
  } catch (e) {
    // ignore
  } finally {
    loading.value = false;
  }
};

const handleDisconnect = async (row: SessionItem) => {
  try {
    await ElMessageBox.confirm(`确认断开用户 ${row.username} 的SSH连接吗?`, '确认操作', {
      type: 'warning',
    });
    await killSSHSession(row.pid);
    ElMessage.success('已断开连接');
    await loadSessions();
  } catch (e: any) {
    if (e !== 'cancel') ElMessage.error(e.message || '操作失败');
  }
};

onMounted(() => {
  loadSessions();
  timer = setInterval(loadSessions, 5000);
});

onUnmounted(() => {
  if (timer) clearInterval(timer);
});
</script>

<style scoped>
.ssh-session {
  padding: 20px;
}
.toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 16px;
}
</style>

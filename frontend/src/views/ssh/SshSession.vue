<template>
  <div class="ssh-session">
    <div class="toolbar">
      <el-input
        v-model="loginUser"
        placeholder="按用户名筛选"
        clearable
        style="width: 220px"
        @clear="fetchSessions"
        @keyup.enter="fetchSessions"
      >
        <template #append>
          <el-button :icon="Search" @click="fetchSessions" />
        </template>
      </el-input>
    </div>

    <el-table :data="sessions" v-loading="loading" border stripe>
      <el-table-column label="来源" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.source === 'websocket'" type="success" size="small">面板终端</el-tag>
          <el-tag v-else type="info" size="small">SSH</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="用户" prop="username" />
      <el-table-column label="TTY" prop="terminal" />
      <el-table-column label="登录IP" prop="host" />
      <el-table-column label="登录时间" prop="loginTime" min-width="120" />
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button type="danger" size="small" link @click="disconnect(row.pid)">
            断开
          </el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Search } from '@element-plus/icons-vue';
import { getSSHSessions, killSSHSession } from '@/api/modules/ssh';
import type { SSHSession } from '@/api/interface/ssh';

const loading = ref(false);
const loginUser = ref('');
const sessions = ref<SSHSession[]>([]);
let timer: ReturnType<typeof setInterval> | null = null;

const fetchSessions = async () => {
  loading.value = true;
  try {
    const res = await getSSHSessions(loginUser.value || undefined);
    sessions.value = res.data ?? [];
  } catch {
    // 静默处理，轮询场景下避免刷屏
  } finally {
    loading.value = false;
  }
};

const startPolling = () => {
  fetchSessions();
  timer = setInterval(fetchSessions, 3000);
};

const stopPolling = () => {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
};

const disconnect = (pid: number) => {
  ElMessageBox.confirm('是否断开此SSH连接？', '断开', {
    confirmButtonText: '确认',
    cancelButtonText: '取消',
    type: 'info',
  })
    .then(async () => {
      try {
        await killSSHSession(pid);
        ElMessage.success('操作成功');
      } catch (e: any) {
        ElMessage.error(e.message || '操作失败');
      }
    })
    .catch(() => {});
};

onMounted(() => {
  startPolling();
});

onUnmounted(() => {
  stopPolling();
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

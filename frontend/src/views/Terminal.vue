<template>
  <div class="terminal-page">
    <el-tabs
      v-model="activeTab"
      type="card"
      class="terminal-tabs"
      @tab-remove="handleTabRemove"
      @tab-change="handleTabChange"
    >
      <!-- 现有的终端标签页 -->
      <el-tab-pane
        v-for="tab in terminalTabs"
        :key="tab.id"
        :label="tab.title"
        :name="tab.id"
        :closable="true"
      >
        <template #label>
          <span class="tab-label">
            <el-icon
              v-if="tab.status === 'online'"
              :style="`color: ${getLatencyColor(tab.latency)}; margin-right: 4px;`"
            >
              <CircleCheck />
            </el-icon>
            <el-icon v-else style="color: #f56c6c; margin-right: 4px;">
              <CircleClose />
            </el-icon>
            <span>{{ tab.title }}</span>
          </span>
        </template>
        <div class="terminal-wrapper">
          <Terminal
            :ref="(el: any) => setTerminalRef(tab.id, el)"
            :mode="tab.mode"
            :ssh-config="tab.sshConfig"
            :cols="tab.cols"
            :rows="tab.rows"
            :font-size="fontSize"
            :theme="theme"
            @connected="handleConnected(tab.id)"
            @disconnected="handleDisconnected(tab.id, $event)"
            @error="handleError(tab.id, $event)"
          />
        </div>
      </el-tab-pane>

      <!-- 新建终端按钮 -->
      <el-tab-pane name="new" :closable="false">
        <template #label>
          <el-button
            type="primary"
            size="small"
            circle
            @click="showNewTerminalDialog = true"
          >
            <el-icon><Plus /></el-icon>
          </el-button>
        </template>
      </el-tab-pane>
    </el-tabs>

    <!-- 空状态 -->
    <el-empty
      v-if="terminalTabs.length === 0"
      description="暂无终端会话，点击右上角按钮新建"
      class="empty-state"
    />

    <!-- 新建终端对话框 -->
    <el-dialog
      v-model="showNewTerminalDialog"
      title="新建终端"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form :model="newTerminalForm" label-width="80px">
        <el-form-item label="终端类型">
          <el-radio-group v-model="newTerminalForm.mode">
            <el-radio value="local">本地终端</el-radio>
            <el-radio value="ssh">SSH 终端</el-radio>
          </el-radio-group>
        </el-form-item>

        <template v-if="newTerminalForm.mode === 'ssh'">
          <el-form-item label="主机地址" required>
            <el-input
              v-model="newTerminalForm.sshConfig.host"
              placeholder="请输入主机地址"
            />
          </el-form-item>
          <el-form-item label="端口" required>
            <el-input-number
              v-model="newTerminalForm.sshConfig.port"
              :min="1"
              :max="65535"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="用户名" required>
            <el-input
              v-model="newTerminalForm.sshConfig.user"
              placeholder="请输入用户名"
            />
          </el-form-item>
          <el-form-item label="密码" required>
            <el-input
              v-model="newTerminalForm.sshConfig.password"
              type="password"
              placeholder="请输入密码"
              show-password
            />
          </el-form-item>
        </template>

        <el-form-item label="列数">
          <el-input-number v-model="newTerminalForm.cols" :min="40" :max="200" />
        </el-form-item>
        <el-form-item label="行数">
          <el-input-number v-model="newTerminalForm.rows" :min="10" :max="100" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showNewTerminalDialog = false">取消</el-button>
        <el-button type="primary" @click="createTerminal">创建</el-button>
      </template>
    </el-dialog>

    <!-- 终端设置对话框 -->
    <el-dialog
      v-model="showSettingsDialog"
      title="终端设置"
      width="400px"
    >
      <el-form :model="settingsForm" label-width="80px">
        <el-form-item label="字体大小">
          <el-slider
            v-model="fontSize"
            :min="10"
            :max="24"
            :step="1"
            show-input
          />
        </el-form-item>
        <el-form-item label="主题">
          <el-radio-group v-model="theme">
            <el-radio value="dark">暗色</el-radio>
            <el-radio value="light">亮色</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button type="primary" @click="showSettingsDialog = false">确定</el-button>
      </template>
    </el-dialog>

    <!-- 设置按钮 -->
    <el-button
      class="settings-button"
      circle
      @click="showSettingsDialog = true"
    >
      <el-icon><Setting /></el-icon>
    </el-button>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, nextTick, onMounted, onBeforeUnmount } from 'vue';
import { ElMessage } from 'element-plus';
import { Plus, CircleCheck, CircleClose, Setting } from '@element-plus/icons-vue';
import Terminal from '@/components/Terminal.vue';
import type { SSHConfig } from '@/api/interface/terminal';

interface TerminalTab {
  id: string;
  title: string;
  mode: 'local' | 'ssh';
  sshConfig?: SSHConfig;
  cols: number;
  rows: number;
  status: 'online' | 'offline';
  latency: number;
}

const activeTab = ref('new');
const terminalTabs = ref<TerminalTab[]>([]);
const terminalRefs = ref<Map<string, any>>(new Map());
const showNewTerminalDialog = ref(false);
const showSettingsDialog = ref(false);
const fontSize = ref(14);
const theme = ref<'light' | 'dark'>('dark');

const newTerminalForm = reactive({
  mode: 'local' as 'local' | 'ssh',
  sshConfig: {
    host: '',
    port: 22,
    user: '',
    password: ''
  } as SSHConfig,
  cols: 80,
  rows: 24
});

const settingsForm = reactive({
  fontSize: 14,
  theme: 'dark' as 'light' | 'dark'
});

let tabCounter = 0;
let latencyTimer: ReturnType<typeof setInterval> | null = null;

const setTerminalRef = (id: string, el: any) => {
  if (el) {
    terminalRefs.value.set(id, el);
  }
};

const getLatencyColor = (latency: number): string => {
  if (latency < 100) return '#67c23a';
  if (latency < 300) return '#e6a23c';
  return '#f56c6c';
};

const createTerminal = () => {
  // 验证 SSH 配置
  if (newTerminalForm.mode === 'ssh') {
    const { host, port, user, password } = newTerminalForm.sshConfig;
    if (!host || !user || !password) {
      ElMessage.error('请填写完整的 SSH 连接信息');
      return;
    }
  }

  const id = `terminal-${tabCounter++}`;
  const title = newTerminalForm.mode === 'local'
    ? '本地终端'
    : `${newTerminalForm.sshConfig.user}@${newTerminalForm.sshConfig.host}`;

  terminalTabs.value.push({
    id,
    title,
    mode: newTerminalForm.mode,
    sshConfig: newTerminalForm.mode === 'ssh' ? { ...newTerminalForm.sshConfig } : undefined,
    cols: newTerminalForm.cols,
    rows: newTerminalForm.rows,
    status: 'offline',
    latency: 0
  });

  activeTab.value = id;
  showNewTerminalDialog.value = false;

  // 重置表单
  newTerminalForm.sshConfig = {
    host: '',
    port: 22,
    user: '',
    password: ''
  };
};

const handleTabRemove = (targetName: string) => {
  const terminal = terminalRefs.value.get(targetName);
  if (terminal) {
    terminal.onClose();
  }
  terminalRefs.value.delete(targetName);

  const index = terminalTabs.value.findIndex(t => t.id === targetName);
  if (index !== -1) {
    terminalTabs.value.splice(index, 1);
  }

  if (activeTab.value === targetName) {
    if (terminalTabs.value.length > 0) {
      activeTab.value = terminalTabs.value[terminalTabs.value.length - 1].id;
    } else {
      activeTab.value = 'new';
    }
  }
};

const handleTabChange = (name: string) => {
  // 标签切换处理
};

const handleConnected = (id: string) => {
  const tab = terminalTabs.value.find(t => t.id === id);
  if (tab) {
    tab.status = 'online';
    tab.latency = 0;
  }
};

const handleDisconnected = (id: string, event: Event) => {
  const tab = terminalTabs.value.find(t => t.id === id);
  if (tab) {
    tab.status = 'offline';
  }
};

const handleError = (id: string, error: Event) => {
  const tab = terminalTabs.value.find(t => t.id === id);
  if (tab) {
    tab.status = 'offline';
  }
};

const updateLatency = () => {
  terminalTabs.value.forEach(tab => {
    const terminal = terminalRefs.value.get(tab.id);
    if (terminal && terminal.isWsOpen()) {
      tab.latency = terminal.getLatency();
      tab.status = 'online';
    } else {
      tab.status = 'offline';
    }
  });
};

onMounted(() => {
  // 定时更新延迟
  latencyTimer = setInterval(updateLatency, 3000);
});

onBeforeUnmount(() => {
  if (latencyTimer) {
    clearInterval(latencyTimer);
  }

  // 关闭所有终端
  terminalTabs.value.forEach(tab => {
    const terminal = terminalRefs.value.get(tab.id);
    if (terminal) {
      terminal.onClose();
    }
  });
});
</script>

<style scoped>
.terminal-page {
  height: calc(100vh - 80px);
  display: flex;
  flex-direction: column;
  background-color: #1e1e1e;
  position: relative;
}

.terminal-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.terminal-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 8px 8px 0 8px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
}

.terminal-tabs :deep(.el-tabs__nav) {
  border: none;
}

.terminal-tabs :deep(.el-tabs__item) {
  background-color: #3d3d3d;
  color: #e0e0e0;
  border: 1px solid #3d3d3d;
  border-bottom: none;
  margin-right: 4px;
  transition: all 0.2s;
}

.terminal-tabs :deep(.el-tabs__item:hover) {
  background-color: #4d4d4d;
  color: #ffffff;
}

.terminal-tabs :deep(.el-tabs__item.is-active) {
  background-color: #1e1e1e;
  color: #ffffff;
  border-color: #4d4d4d;
  border-bottom-color: #1e1e1e;
}

.terminal-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: hidden;
  padding: 0;
}

.terminal-tabs :deep(.el-tab-pane) {
  height: 100%;
  overflow: hidden;
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 4px;
}

.terminal-wrapper {
  width: 100%;
  height: 100%;
  overflow: hidden;
}

.empty-state {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
}

.settings-button {
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 100;
}

:deep(.el-dialog) {
  background-color: #2d2d2d;
}

:deep(.el-dialog__header) {
  border-bottom: 1px solid #3d3d3d;
}

:deep(.el-dialog__title) {
  color: #e0e0e0;
}

:deep(.el-dialog__body) {
  color: #e0e0e0;
}

:deep(.el-form-item__label) {
  color: #e0e0e0;
}

:deep(.el-input__inner) {
  background-color: #3d3d3d;
  border-color: #4d4d4d;
  color: #e0e0e0;
}

:deep(.el-input__inner:focus) {
  border-color: #409eff;
}

:deep(.el-input-number) {
  background-color: #3d3d3d;
}

:deep(.el-input-number .el-input__inner) {
  background-color: #3d3d3d;
  color: #e0e0e0;
}

:deep(.el-slider__runway) {
  background-color: #3d3d3d;
}

:deep(.el-radio__label) {
  color: #e0e0e0;
}
</style>
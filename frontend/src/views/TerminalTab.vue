<template>
  <div class="terminal-page">
    <!-- 主机列表侧边栏 -->
    <div class="host-sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header">
        <span v-if="!sidebarCollapsed" class="sidebar-title">主机列表</span>
        <el-button
          :icon="sidebarCollapsed ? DArrowRight : DArrowLeft"
          circle
          size="small"
          @click="sidebarCollapsed = !sidebarCollapsed"
        />
      </div>
      <div v-if="!sidebarCollapsed" class="sidebar-content">
        <div
          v-for="group in hostTree"
          :key="group.id"
          class="host-group"
        >
          <div class="group-header" @click="toggleGroup(group.id)">
            <el-icon>
              <Folder v-if="!group.expanded" />
              <FolderOpened v-else />
            </el-icon>
            <span>{{ group.name }}</span>
          </div>
          <div v-show="group.expanded" class="host-list">
            <div
              v-for="host in group.children"
              :key="host.id"
              class="host-item"
              @click="connectHost(host)"
            >
              <el-icon><Monitor /></el-icon>
              <span>{{ host.name }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 终端区域 -->
    <div class="terminal-content">
      <!-- 工具栏 -->
      <div class="terminal-toolbar">
        <div class="toolbar-left">
          <el-button
            :icon="Plus"
            size="small"
            @click="showNewTerminalDialog = true"
          >
            新建终端
          </el-button>
          <el-button
            :icon="Close"
            size="small"
            :disabled="terminalTabs.length === 0"
            @click="closeCurrentTab"
          >
            关闭当前
          </el-button>
          <el-button
            :icon="FullScreen"
            size="small"
            :disabled="terminalTabs.length === 0"
            @click="toggleFullscreen"
          >
            全屏
          </el-button>
          <el-button
            :icon="Grid"
            size="small"
            :disabled="terminalTabs.length < 2"
            @click="toggleBatchMode"
            :type="batchMode ? 'primary' : 'default'"
          >
            批量输入
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-button
            :icon="List"
            size="small"
            @click="showQuickCommandPanel = true"
          >
            快速命令
          </el-button>
          <el-button
            :icon="Setting"
            size="small"
            @click="showSettingsDialog = true"
          >
            设置
          </el-button>
        </div>
      </div>

      <!-- 终端标签页 -->
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
              :font-size="terminalSettings.fontSize"
              :theme="theme"
              :line-height="terminalSettings.lineHeight"
              :letter-spacing="terminalSettings.letterSpacing"
              :cursor-blink="terminalSettings.cursorBlink"
              :cursor-style="terminalSettings.cursorStyle"
              :scrollback="terminalSettings.scrollback"
              :scroll-sensitivity="terminalSettings.scrollSensitivity"
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
        description="暂无终端会话，点击上方按钮新建或从左侧主机列表连接"
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
            <el-form-item label="认证方式" required>
              <el-radio-group v-model="newTerminalForm.sshConfig.authMode">
                <el-radio value="password">密码</el-radio>
                <el-radio value="key">密钥</el-radio>
              </el-radio-group>
            </el-form-item>
            <el-form-item
              v-if="newTerminalForm.sshConfig.authMode === 'password'"
              label="密码"
              required
            >
              <el-input
                v-model="newTerminalForm.sshConfig.password"
                type="password"
                placeholder="请输入密码"
                show-password
              />
            </el-form-item>
            <el-form-item
              v-if="newTerminalForm.sshConfig.authMode === 'key'"
              label="私钥"
              required
            >
              <el-input
                v-model="newTerminalForm.sshConfig.key"
                type="textarea"
                :rows="4"
                placeholder="请输入私钥内容"
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
              v-model="settingsForm.fontSize"
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

      <!-- 快速命令面板 -->
      <el-drawer
        v-model="showQuickCommandPanel"
        title="快速命令"
        direction="rtl"
        size="400px"
      >
        <el-input
          v-model="commandSearchKeyword"
          placeholder="搜索命令"
          clearable
          style="margin-bottom: 16px"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-empty v-if="filteredCommands.length === 0" description="暂无快速命令" />
        <div v-else class="command-list">
          <div
            v-for="cmd in filteredCommands"
            :key="cmd.id"
            class="command-item"
            @click="executeQuickCommand(cmd)"
          >
            <div class="command-name">{{ cmd.name }}</div>
            <div class="command-desc">{{ cmd.description || cmd.command }}</div>
          </div>
        </div>
      </el-drawer>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, nextTick, onMounted, onBeforeUnmount, computed } from 'vue';
import { useRoute } from 'vue-router';
import { ElMessage } from 'element-plus';
import {
  Plus,
  CircleCheck,
  CircleClose,
  Setting,
  DArrowLeft,
  DArrowRight,
  Folder,
  FolderOpened,
  Monitor,
  Close,
  FullScreen,
  Grid,
  List,
  Search
} from '@element-plus/icons-vue';
import Terminal from '@/components/Terminal.vue';
import { getHostTree, getHostForTerminal } from '@/api/modules/host';
import { getAllQuickCommands } from '@/api/modules/quick_command';
import type { SSHConfig } from '@/api/interface/terminal';
import type { HostTreeNode, HostConnInfo } from '@/api/interface/host';
import type { QuickCommand } from '@/api/interface/quick_command';
import screenfull from 'screenfull';

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

interface HostGroupWithChildren extends HostTreeNode {
  expanded?: boolean;
}

const route = useRoute();

const activeTab = ref('new');
const terminalTabs = ref<TerminalTab[]>([]);
const terminalRefs = ref<Map<string, any>>(new Map());
const showNewTerminalDialog = ref(false);
const showSettingsDialog = ref(false);
const showQuickCommandPanel = ref(false);
const fontSize = ref(14);
const theme = ref<'light' | 'dark'>('dark');
const sidebarCollapsed = ref(false);
const hostTree = ref<HostGroupWithChildren[]>([]);
const batchMode = ref(false);
const commandSearchKeyword = ref('');
const quickCommands = ref<QuickCommand[]>([]);

const terminalSettings = reactive({
  lineHeight: '1.2',
  letterSpacing: '1.2',
  fontSize: '14',
  cursorBlink: 'enable',
  cursorStyle: 'underline',
  scrollback: '1000',
  scrollSensitivity: '10'
});

const newTerminalForm = reactive({
  mode: 'local' as 'local' | 'ssh',
  sshConfig: {
    host: '',
    port: 22,
    user: '',
    password: '',
    authMode: 'password' as 'password' | 'key',
    key: ''
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
let reconnectTimer: Map<string, ReturnType<typeof setInterval>> = new Map();

const filteredCommands = computed(() => {
  if (!commandSearchKeyword.value) {
    return quickCommands.value;
  }
  const keyword = commandSearchKeyword.value.toLowerCase();
  return quickCommands.value.filter(cmd =>
    cmd.name.toLowerCase().includes(keyword) ||
    cmd.command.toLowerCase().includes(keyword) ||
    (cmd.description && cmd.description.toLowerCase().includes(keyword))
  );
});

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

const loadHostTree = async () => {
  try {
    const res = await getHostTree();
    hostTree.value = res.data.map(group => ({
      ...group,
      expanded: true
    }));
  } catch (error) {
    console.error('加载主机列表失败:', error);
  }
};

const loadQuickCommands = async () => {
  try {
    const res = await getAllQuickCommands();
    quickCommands.value = res.data;
  } catch (error) {
    console.error('加载快速命令失败:', error);
  }
};

const toggleGroup = (groupId: number) => {
  const group = hostTree.value.find(g => g.id === groupId);
  if (group) {
    group.expanded = !group.expanded;
  }
};

const connectHost = async (host: HostTreeNode) => {
  try {
    const res = await getHostForTerminal(host.id);
    const hostInfo = res.data;

    const id = `terminal-${tabCounter++}`;
    const title = `${hostInfo.user}@${hostInfo.addr}`;

    terminalTabs.value.push({
      id,
      title,
      mode: 'ssh',
      sshConfig: {
        host: hostInfo.addr,
        port: hostInfo.port,
        user: hostInfo.user,
        password: hostInfo.authMode === 'password' ? hostInfo.password : '',
        key: hostInfo.authMode === 'key' ? hostInfo.privateKey : '',
        authMode: hostInfo.authMode
      },
      cols: 80,
      rows: 24,
      status: 'offline',
      latency: 0
    });

    activeTab.value = id;
  } catch (error) {
    console.error('连接主机失败:', error);
    ElMessage.error('连接主机失败');
  }
};

const createTerminal = () => {
  // 验证 SSH 配置
  if (newTerminalForm.mode === 'ssh') {
    const { host, port, user, authMode } = newTerminalForm.sshConfig;
    if (!host || !user || !authMode) {
      ElMessage.error('请填写完整的 SSH 连接信息');
      return;
    }
    if (authMode === 'password' && !newTerminalForm.sshConfig.password) {
      ElMessage.error('请输入密码');
      return;
    }
    if (authMode === 'key' && !newTerminalForm.sshConfig.key) {
      ElMessage.error('请输入私钥');
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
    password: '',
    authMode: 'password',
    key: ''
  };
};

const closeCurrentTab = () => {
  if (activeTab.value !== 'new') {
    handleTabRemove(activeTab.value);
  }
};

const handleTabRemove = (targetName: string) => {
  // 清除重连定时器
  if (reconnectTimer.has(targetName)) {
    clearInterval(reconnectTimer.get(targetName)!);
    reconnectTimer.delete(targetName);
  }

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
    // 清除重连定时器
    if (reconnectTimer.has(id)) {
      clearInterval(reconnectTimer.get(id)!);
      reconnectTimer.delete(id);
    }
  }
};

const handleDisconnected = (id: string, event: Event) => {
  const tab = terminalTabs.value.find(t => t.id === id);
  if (tab) {
    tab.status = 'offline';
    // 启动自动重连
    startAutoReconnect(id);
  }
};

const handleError = (id: string, error: Event) => {
  const tab = terminalTabs.value.find(t => t.id === id);
  if (tab) {
    tab.status = 'offline';
  }
};

const startAutoReconnect = (id: string) => {
  // 如果已经有重连定时器，不重复创建
  if (reconnectTimer.has(id)) {
    return;
  }

  let retryCount = 0;
  const maxRetries = 5;

  const timer = setInterval(() => {
    retryCount++;
    const terminal = terminalRefs.value.get(id);
    if (terminal) {
      console.log(`尝试重连终端 ${id} (${retryCount}/${maxRetries})`);
      terminal.reconnect();
    }

    if (retryCount >= maxRetries) {
      clearInterval(timer);
      reconnectTimer.delete(id);
      ElMessage.warning(`终端 ${id} 重连失败，请手动重连`);
    }
  }, 3000);

  reconnectTimer.set(id, timer);
};

const toggleFullscreen = () => {
  if (screenfull.isEnabled) {
    screenfull.toggle();
  } else {
    ElMessage.warning('当前浏览器不支持全屏功能');
  }
};

const toggleBatchMode = () => {
  batchMode.value = !batchMode.value;
  if (batchMode.value) {
    ElMessage.info('批量输入模式已开启，输入将发送到所有在线终端');
  } else {
    ElMessage.info('批量输入模式已关闭');
  }
};

const executeQuickCommand = (cmd: QuickCommand) => {
  if (terminalTabs.length === 0) {
    ElMessage.warning('请先创建终端');
    return;
  }

  const command = cmd.command + '\r';
  const targetTabs = batchMode.value
    ? terminalTabs.value.filter(t => t.status === 'online')
    : [terminalTabs.value.find(t => t.id === activeTab.value)].filter(Boolean);

  if (targetTabs.length === 0) {
    ElMessage.warning('没有可用的在线终端');
    return;
  }

  targetTabs.forEach(tab => {
    const terminal = terminalRefs.value.get(tab.id);
    if (terminal && terminal.isWsOpen()) {
      terminal.sendMsg(command);
    }
  });

  showQuickCommandPanel.value = false;
  ElMessage.success(`命令已发送到 ${targetTabs.length} 个终端`);
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

// 从URL参数连接主机
const connectFromUrl = async () => {
  const hostId = route.query.hostId as string;
  if (hostId) {
    await loadHostTree();
    // 查找主机
    for (const group of hostTree.value) {
      const host = group.children?.find(h => h.id === parseInt(hostId));
      if (host) {
        await connectHost(host);
        break;
      }
    }
  }
};

const acceptParams = () => {
  // 页面加载时刷新数据
  loadHostTree();
  loadQuickCommands();
};

onMounted(async () => {
  await loadHostTree();
  await loadQuickCommands();
  await connectFromUrl();
  // 定时更新延迟
  latencyTimer = setInterval(updateLatency, 3000);
});

onBeforeUnmount(() => {
  if (latencyTimer) {
    clearInterval(latencyTimer);
  }

  // 清除所有重连定时器
  reconnectTimer.forEach(timer => clearInterval(timer));
  reconnectTimer.clear();

  // 关闭所有终端
  terminalTabs.value.forEach(tab => {
    const terminal = terminalRefs.value.get(tab.id);
    if (terminal) {
      terminal.onClose();
    }
  });
});

defineExpose({
  acceptParams
});
</script>

<style scoped>
.terminal-page {
  height: calc(100vh - 80px);
  display: flex;
  background-color: #1e1e1e;
  position: relative;
}

.host-sidebar {
  width: 250px;
  background-color: #2d2d2d;
  border-right: 1px solid #3d3d3d;
  display: flex;
  flex-direction: column;
  transition: width 0.3s;
}

.host-sidebar.collapsed {
  width: 50px;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #3d3d3d;
}

.sidebar-title {
  font-size: 14px;
  font-weight: 600;
  color: #e0e0e0;
}

.sidebar-content {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
}

.sidebar-content::-webkit-scrollbar {
  width: 6px;
}

.sidebar-content::-webkit-scrollbar-thumb {
  background-color: #4d4d4d;
  border-radius: 3px;
}

.host-group {
  margin-bottom: 4px;
}

.group-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  cursor: pointer;
  color: #e0e0e0;
  font-size: 13px;
  transition: background-color 0.2s;
}

.group-header:hover {
  background-color: #3d3d3d;
}

.host-list {
  padding-left: 24px;
}

.host-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  cursor: pointer;
  color: #b0b0b0;
  font-size: 12px;
  transition: all 0.2s;
}

.host-item:hover {
  background-color: #3d3d3d;
  color: #e0e0e0;
}

.terminal-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.terminal-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
}

.toolbar-left,
.toolbar-right {
  display: flex;
  gap: 8px;
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

.command-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.command-item {
  padding: 12px;
  background-color: #f5f5f5;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.command-item:hover {
  background-color: #e6f7ff;
}

.command-name {
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.command-desc {
  font-size: 12px;
  color: #666;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

:deep(.el-textarea__inner) {
  background-color: #3d3d3d;
  border-color: #4d4d4d;
  color: #e0e0e0;
}

:deep(.el-textarea__inner:focus) {
  border-color: #409eff;
}
</style>
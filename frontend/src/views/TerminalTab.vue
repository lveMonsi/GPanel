<template>
  <div class="terminal-page">
    <!-- 终端区域 -->
    <main class="terminal-main">
      <!-- 顶部工具栏 -->
      <div class="toolbar">
        <div class="toolbar-section-left">
          <el-button type="primary" :icon="Plus" size="small" @click="showNewTerminalDialog = true">
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
            :type="batchMode ? 'primary' : ''"
          >
            批量输入
          </el-button>
        </div>
        <div class="toolbar-section-right">
          <el-button
            :icon="Connection"
            size="small"
            @click="showHostListDialog = true"
          >
            主机列表
          </el-button>
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
            @click="openSettingsDialog"
          >
            设置
          </el-button>
        </div>
      </div>

      <!-- 终端标签页 -->
      <div class="terminal-tabs-container">
        <el-tabs
          v-model="activeTab"
          type="card"
          class="terminal-tabs"
          @tab-remove="handleTabRemove"
          @tab-change="handleTabChange"
        >
          <el-tab-pane
            v-for="tab in terminalTabs"
            :key="tab.id"
            :name="tab.id"
            :closable="true"
          >
            <template #label>
              <span class="tab-label">
                <el-icon
                  v-if="tab.status === 'online'"
                  :style="`color: ${getLatencyColor(tab.latency)}`"
                  class="status-icon"
                >
                  <CircleCheck />
                </el-icon>
                <el-icon v-else class="status-icon offline">
                  <CircleClose />
                </el-icon>
                <span class="tab-title">{{ tab.title }}</span>
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
        <div v-if="terminalTabs.length === 0" class="empty-state">
          <el-empty description="暂无终端会话，点击上方按钮新建或从主机列表连接" />
        </div>
      </div>

      <!-- 批量输入栏 -->
      <div class="batch-input-bar">
        <el-switch
          v-model="batchMode"
          active-text="批量输入"
          size="small"
          @change="handleBatchModeChange"
        />
        <el-input
          v-model="batchInputValue"
          placeholder="批量输入内容，回车发送到所有在线终端"
          :disabled="!batchMode"
          @keyup.enter="handleBatchInputSubmit"
          clearable
        >
          <template #append>
            <el-button :icon="Position" @click="handleBatchInputSubmit" :disabled="!batchMode" />
          </template>
        </el-input>
      </div>
    </main>

    <!-- 主机列表对话框 -->
    <el-dialog
      v-model="showHostListDialog"
      title="主机列表"
      width="600px"
      :close-on-click-modal="false"
    >
      <div class="host-list-dialog">
        <div
          v-for="group in hostTree"
          :key="group.id"
          class="host-group"
        >
          <div class="group-header" @click="toggleGroup(group.id)">
            <el-icon class="group-icon">
              <Folder v-if="!group.expanded" />
              <FolderOpened v-else />
            </el-icon>
            <span class="group-name">{{ group.name }}</span>
            <span class="group-count">{{ group.children?.length || 0 }}</span>
          </div>
          <div v-show="group.expanded" class="host-list">
            <div
              v-for="host in group.children"
              :key="host.id"
              class="host-item"
              @click="connectHost(host)"
            >
              <el-icon class="host-icon"><Monitor /></el-icon>
              <div class="host-info">
                <span class="host-name">{{ host.name }}</span>
                <span class="host-addr">{{ host.addr }}</span>
              </div>
              <el-button type="primary" size="small" :icon="Connection">连接</el-button>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 新建终端对话框 -->
    <el-dialog
      v-model="showNewTerminalDialog"
      title="新建终端"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form :model="newTerminalForm" label-width="90px">
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
      <el-form :model="settingsForm" label-width="100px">
        <el-form-item label="字体大小">
          <el-slider
            v-model="settingsForm.fontSize"
            :min="10"
            :max="24"
            :step="1"
            show-input
            @change="handleFontSizeChange"
          />
        </el-form-item>
        <el-form-item label="主题">
          <el-radio-group v-model="theme">
            <el-radio value="dark">暗色</el-radio>
            <el-radio value="light">亮色</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="最大重连次数">
          <el-input-number
            v-model="settingsForm.maxReconnectRetries"
            :min="0"
            :max="20"
            :step="1"
          />
          <div class="form-tip">设置为 0 表示不自动重连</div>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showSettingsDialog = false">取消</el-button>
        <el-button type="primary" @click="applySettings">确定</el-button>
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
  Folder,
  FolderOpened,
  Monitor,
  Close,
  FullScreen,
  Grid,
  List,
  Search,
  Connection,
  Position
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
const showHostListDialog = ref(false);
const fontSize = ref(14);
const theme = ref<'light' | 'dark'>('dark');
const hostTree = ref<HostGroupWithChildren[]>([]);
const batchMode = ref(false);
const batchInputValue = ref('');
const commandSearchKeyword = ref('');
const quickCommands = ref<QuickCommand[]>([]);

const terminalSettings = reactive({
  lineHeight: '1.2',
  letterSpacing: '1.2',
  fontSize: '14',
  cursorBlink: 'enable',
  cursorStyle: 'underline',
  scrollback: '1000',
  scrollSensitivity: '10',
  maxReconnectRetries: 5
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
  } as SSHConfig
});

const settingsForm = reactive({
  fontSize: 14,
  theme: 'dark' as 'light' | 'dark',
  maxReconnectRetries: 5
});

let tabCounter = 0;
let latencyTimer: ReturnType<typeof setInterval> | null = null;
let reconnectTimer: Map<string, ReturnType<typeof setInterval>> = new Map();
let reconnectCount: Map<string, number> = new Map();
let reconnectAbandoned: Map<string, boolean> = new Map(); // 标记已放弃重连的终端

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
    showHostListDialog.value = false;
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
    cols: 80,
    rows: 24,
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
  // 清除重连定时器和计数器
  if (reconnectTimer.has(targetName)) {
    clearInterval(reconnectTimer.get(targetName)!);
    reconnectTimer.delete(targetName);
  }
  reconnectCount.delete(targetName);
  reconnectAbandoned.delete(targetName); // 清除放弃重连的标记

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
    // 清除重连定时器和计数器
    if (reconnectTimer.has(id)) {
      clearInterval(reconnectTimer.get(id)!);
      reconnectTimer.delete(id);
      reconnectCount.delete(id);
    }
    // 清除放弃重连的标记
    reconnectAbandoned.delete(id);
  }
};

const handleDisconnected = (id: string, event: Event) => {
  const tab = terminalTabs.value.find(t => t.id === id);
  if (tab) {
    tab.status = 'offline';
    // 只在定时器不存在时启动自动重连
    if (!reconnectTimer.has(id)) {
      startAutoReconnect(id);
    }
  }
};

const handleError = (id: string, error: Event) => {
  const tab = terminalTabs.value.find(t => t.id === id);
  if (tab) {
    tab.status = 'offline';
  }
  // 在终端中显示错误信息（使用 write 确保消息能显示）
  const terminal = terminalRefs.value.get(id);
  if (terminal) {
    const errorMsg = error instanceof Error ? error.message : String(error);
    const errMsg = `\r\n[系统错误] ${errorMsg}\r\n`;
    const term = terminal.getTerm();
    if (term) {
      term.write('\x1b[31m' + errMsg + '\x1b[m');
    }
    console.error(`终端 ${id} 错误:`, error);
  }
};

const handleBatchModeChange = (value: boolean) => {
  if (value) {
    ElMessage.info('批量输入模式已开启，在下方输入框输入内容并回车，将发送到所有在线终端');
  } else {
    ElMessage.info('批量输入模式已关闭');
  }
};

const handleFontSizeChange = (value: number) => {
  terminalSettings.fontSize = value.toString();
};

const openSettingsDialog = () => {
  // 同步当前设置到表单
  settingsForm.fontSize = parseInt(terminalSettings.fontSize);
  settingsForm.theme = theme.value as 'light' | 'dark';
  settingsForm.maxReconnectRetries = terminalSettings.maxReconnectRetries;
  showSettingsDialog.value = true;
};

const applySettings = () => {
  terminalSettings.fontSize = settingsForm.fontSize.toString();
  theme.value = settingsForm.theme;
  terminalSettings.maxReconnectRetries = settingsForm.maxReconnectRetries;
  showSettingsDialog.value = false;
  ElMessage.success('设置已应用');
};

const handleBatchInputSubmit = () => {
  if (!batchMode.value) {
    return;
  }

  const input = batchInputValue.value.trim();
  if (!input) {
    ElMessage.warning('请输入内容');
    return;
  }

  const command = input + '\r';
  const onlineTabs = terminalTabs.value.filter(t => t.status === 'online');

  if (onlineTabs.length === 0) {
    ElMessage.warning('没有可用的在线终端');
    return;
  }

  onlineTabs.forEach(tab => {
    const terminal = terminalRefs.value.get(tab.id);
    if (terminal && terminal.isWsOpen()) {
      terminal.sendMsg(command);
    }
  });

  ElMessage.success(`命令已发送到 ${onlineTabs.length} 个终端`);
  batchInputValue.value = '';
};

const startAutoReconnect = (id: string) => {
  // 如果已经有重连定时器，不重复创建
  if (reconnectTimer.has(id)) {
    return;
  }

  // 如果该终端已经放弃重连，不再启动重连
  if (reconnectAbandoned.has(id)) {
    console.log(`[重连] 终端 ${id} 已放弃重连，跳过`);
    return;
  }

  const maxRetries = terminalSettings.maxReconnectRetries;
  if (maxRetries <= 0) {
    return;
  }

  // 初始化重连次数
  reconnectCount.set(id, 0);

  const timer = setInterval(() => {
    const currentRetry = (reconnectCount.get(id) || 0) + 1;
    reconnectCount.set(id, currentRetry);

    const terminal = terminalRefs.value.get(id);
    if (terminal) {
      console.log(`尝试重连终端 ${id} (${currentRetry}/${maxRetries})`);
      // 在终端中显示重连信息（使用 write 确保消息能显示）
      const reconnectMsg = `\r\n[系统] 正在尝试重连... (${currentRetry}/${maxRetries})\r\n`;
      const term = terminal.getTerm();
      if (term) {
        term.write('\x1b[33m' + reconnectMsg + '\x1b[m');
      }
      terminal.reconnect();
    }

    if (currentRetry >= maxRetries) {
      clearInterval(timer);
      reconnectTimer.delete(id);
      reconnectCount.delete(id);
      reconnectAbandoned.set(id, true); // 标记该终端已放弃重连
      // 在终端中显示失败信息
      const terminal = terminalRefs.value.get(id);
      if (terminal) {
        const failMsg = `\r\n[系统] 自动重连失败，已达到最大重试次数 (${maxRetries} 次)，请手动重连\r\n`;
        const term = terminal.getTerm();
        if (term) {
          term.write('\x1b[31m' + failMsg + '\x1b[m');
        }
      }
      ElMessage.warning(`终端重连失败，已达到最大重试次数`);
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
  const targetTabs = [terminalTabs.value.find(t => t.id === activeTab.value)].filter(Boolean);

  if (targetTabs.length === 0) {
    ElMessage.warning('没有选中的终端');
    return;
  }

  const targetTab = targetTabs[0];
  if (targetTab.status !== 'online') {
    ElMessage.warning('选中的终端未在线');
    return;
  }

  const terminal = terminalRefs.value.get(targetTab.id);
  if (terminal && terminal.isWsOpen()) {
    terminal.sendMsg(command);
    showQuickCommandPanel.value = false;
    ElMessage.success(`命令已发送到 ${targetTab.title}`);
  } else {
    ElMessage.warning('终端连接已断开');
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
  // 不要在这里关闭连接，让用户在离开终端页面时才关闭
  if (latencyTimer) {
    clearInterval(latencyTimer);
  }

  // 清除所有重连定时器
  reconnectTimer.forEach(timer => clearInterval(timer));
  reconnectTimer.clear();
});

// 暴露给父组件的方法
const cleanupTerminals = () => {
  // 关闭所有终端连接
  terminalTabs.value.forEach(tab => {
    const terminal = terminalRefs.value.get(tab.id);
    if (terminal) {
      terminal.onClose();
    }
  });
  terminalRefs.value.clear();
  terminalTabs.value = [];
};

defineExpose({
  acceptParams,
  cleanupTerminals
});
</script>

<style scoped>
.terminal-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: #f5f7fa;
  overflow: hidden;
}

/* 终端主区域 */
.terminal-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background-color: #1e1e1e;
  min-width: 0;
}

/* 工具栏 */
.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
  gap: 16px;
}

.toolbar-section-left,
.toolbar-section-right {
  display: flex;
  gap: 8px;
  align-items: center;
}

/* 终端标签页容器 */
.terminal-tabs-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
  background-color: #1e1e1e;
}

.terminal-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
  background-color: #1e1e1e;
}

.terminal-tabs :deep(.el-tabs__header) {
  margin: 0;
  padding: 8px 8px 0 8px;
  background-color: #2d2d2d;
  border-bottom: 1px solid #3d3d3d;
  flex-shrink: 0;
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
  height: 32px;
  line-height: 32px;
  font-size: 13px;
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
  background-color: #1e1e1e;
}

.terminal-tabs :deep(.el-tab-pane) {
  height: 100%;
  overflow: hidden;
  background-color: #1e1e1e;
}

.tab-label {
  display: flex;
  align-items: center;
  gap: 6px;
}

.status-icon {
  font-size: 14px;
  flex-shrink: 0;
}

.status-icon.offline {
  color: #f56c6c;
}

.tab-title {
  font-size: 13px;
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.terminal-wrapper {
  width: 100%;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.terminal-wrapper :deep(.terminal-container) {
  flex: 1;
  overflow: hidden;
}

/* 空状态 */
.empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: #1e1e1e;
  overflow: hidden;
}

.empty-state :deep(.el-empty) {
  --el-empty-description-margin-top: 16px;
}

.empty-state :deep(.el-empty__description) {
  color: #b0b0b0;
  font-size: 14px;
}

.empty-state :deep(.el-empty__description) {
  color: #b0b0b0;
  font-size: 14px;
}

/* 批量输入栏 */
.batch-input-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background-color: #2d2d2d;
  border-top: 1px solid #3d3d3d;
}

.batch-input-bar .el-switch {
  flex-shrink: 0;
  min-width: 80px;
}

.batch-input-bar .el-switch :deep(.el-switch__label) {
  color: #e0e0e0;
}

.batch-input-bar .el-switch :deep(.el-switch__label.is-active) {
  color: #409eff;
}

.batch-input-bar .el-input {
  flex: 1;
}

.batch-input-bar :deep(.el-input__wrapper) {
  background-color: #3d3d3d;
  box-shadow: none;
  padding: 0;
}

.batch-input-bar :deep(.el-input__inner) {
  background-color: transparent;
  border: none;
  color: #e0e0e0;
  box-shadow: none;
}

.batch-input-bar :deep(.el-input__inner:focus) {
  background-color: transparent;
  border: none;
  box-shadow: none;
}

.batch-input-bar :deep(.el-input__inner:disabled) {
  background-color: transparent;
  color: #666;
}

.batch-input-bar :deep(.el-input-group__append) {
  background-color: #3d3d3d;
  border: none;
  box-shadow: none;
}

.batch-input-bar :deep(.el-input-group__append .el-button) {
  background-color: transparent;
  border: none;
  color: #e0e0e0;
  box-shadow: none;
}

.batch-input-bar :deep(.el-input-group__append .el-button:hover) {
  background-color: #4d4d4d;
  color: #ffffff;
}

.batch-input-bar :deep(.el-input-group__append .el-button:disabled) {
  background-color: transparent;
  color: #666;
}

/* 主机列表对话框 */
.host-list-dialog {
  max-height: 500px;
  overflow-y: auto;
}

.host-group {
  margin-bottom: 8px;
}

.group-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  cursor: pointer;
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 600;
  transition: all 0.2s;
  background-color: var(--bg-color);
  border-radius: var(--radius-md);
  margin-bottom: 4px;
}

.group-header:hover {
  background-color: var(--border-color);
}

.group-icon {
  font-size: 18px;
  color: var(--primary-dark);
  flex-shrink: 0;
}

.group-name {
  flex: 1;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.group-count {
  font-size: 12px;
  color: var(--text-secondary);
  background: var(--primary-light);
  padding: 2px 8px;
  border-radius: 10px;
  font-weight: 500;
}

.host-list {
  padding: 0 8px 8px 8px;
}

.host-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  color: var(--text-secondary);
  font-size: 13px;
  transition: all 0.2s;
  border-radius: var(--radius-md);
  background-color: var(--card-bg);
  border: 1px solid var(--border-color);
  margin-bottom: 8px;
}

.host-item:hover {
  background-color: var(--primary-light);
  border-color: var(--primary);
  transform: translateX(4px);
}

.host-icon {
  font-size: 18px;
  color: var(--primary);
  flex-shrink: 0;
}

.host-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.host-name {
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.host-addr {
  font-size: 12px;
  color: var(--text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 快速命令列表 */
.command-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.command-item {
  padding: 12px;
  background-color: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
}

.command-item:hover {
  background-color: var(--primary-light);
  border-color: var(--primary);
  transform: translateX(-2px);
}

.command-name {
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 4px;
  font-size: 14px;
}

.command-desc {
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 对话框样式 */
:deep(.el-dialog) {
  background-color: var(--card-bg);
}

:deep(.el-dialog__header) {
  border-bottom: 1px solid var(--border-color);
  padding: 16px 20px;
}

:deep(.el-dialog__title) {
  color: var(--text-primary);
  font-size: 16px;
  font-weight: 600;
}

:deep(.el-dialog__body) {
  color: var(--text-primary);
  padding: 20px;
}

:deep(.el-dialog__footer) {
  border-top: 1px solid var(--border-color);
  padding: 16px 20px;
}

:deep(.el-form-item__label) {
  color: var(--text-primary);
  font-size: 14px;
  font-weight: 500;
}

:deep(.el-input__inner) {
  background-color: var(--bg-color);
  border-color: var(--border-color);
  color: var(--text-primary);
}

:deep(.el-input__inner:focus) {
  border-color: var(--primary);
}

:deep(.el-input-number) {
  background-color: var(--bg-color);
}

:deep(.el-input-number .el-input__inner) {
  background-color: var(--bg-color);
  color: var(--text-primary);
}

:deep(.el-slider__runway) {
  background-color: var(--border-color);
}

:deep(.el-radio__label) {
  color: var(--text-primary);
}

:deep(.el-textarea__inner) {
  background-color: var(--bg-color);
  border-color: var(--border-color);
  color: var(--text-primary);
}

:deep(.el-textarea__inner:focus) {
  border-color: var(--primary);
}

/* 抽屉样式 */
:deep(.el-drawer__header) {
  margin-bottom: 16px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

:deep(.el-drawer__title) {
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

:deep(.el-drawer__body) {
  padding: 0 20px 20px 20px;
}

/* 表单提示 */
.form-tip {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 4px;
}
</style>
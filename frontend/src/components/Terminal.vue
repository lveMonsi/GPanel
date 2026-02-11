<template>
  <div ref="terminalElement" class="terminal-container"></div>
</template>

<script setup lang="ts">
import { ref, watch, onBeforeUnmount, onMounted, computed } from 'vue';
import { Terminal } from '@xterm/xterm';
import '@xterm/xterm/css/xterm.css';
import { FitAddon } from '@xterm/addon-fit';
import { Base64 } from 'js-base64';
import type { TerminalConfig, SSHConfig, WsMsg } from '@/api/interface/terminal';
import { getLocalTerminalUrl, getSSHTerminalUrl } from '@/api/modules/terminal';

interface Props {
  mode: 'local' | 'ssh';
  sshConfig?: SSHConfig;
  cols?: number;
  rows?: number;
  fontSize?: number;
  theme?: 'light' | 'dark';
  lineHeight?: string;
  letterSpacing?: string;
  cursorBlink?: string;
  cursorStyle?: string;
  scrollback?: string;
  scrollSensitivity?: string;
}

const props = withDefaults(defineProps<Props>(), {
  cols: 80,
  rows: 24,
  fontSize: 14,
  theme: 'dark',
  lineHeight: '1.2',
  letterSpacing: '1.2',
  cursorBlink: 'enable',
  cursorStyle: 'underline',
  scrollback: '1000',
  scrollSensitivity: '10'
});

const emit = defineEmits<{
  connected: [];
  disconnected: [event: Event];
  error: [error: Event];
}>();

const terminalElement = ref<HTMLDivElement | null>(null);
const fitAddon = new FitAddon();
const termReady = ref(false);
const webSocketReady = ref(false);
const term = ref<Terminal | null>(null);
const terminalSocket = ref<WebSocket | null>(null);
const heartbeatTimer = ref<NodeJS.Timeout | null>(null);
const latency = ref(0);
const resizeObserver = ref<ResizeObserver | null>(null);

const terminalConfig = computed<TerminalConfig>(() => ({
  cols: props.cols,
  rows: props.rows,
  fontSize: props.fontSize,
  fontFamily: "Monaco, Menlo, Consolas, 'Courier New', monospace",
  theme: props.theme
}));

// 监听主题变化
watch(() => props.theme, (newTheme) => {
  if (term.value) {
    const background = newTheme === 'dark' ? '#1e1e1e' : '#ffffff';
    const foreground = newTheme === 'dark' ? '#ffffff' : '#000000';
    term.value.options.theme = {
      background,
      foreground,
      cursor: foreground,
      selection: newTheme === 'dark' ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.3)'
    };
  }
});

// 监听字体大小变化
watch(() => props.fontSize, (newFontSize) => {
  if (term.value) {
    term.value.options.fontSize = newFontSize;
    changeTerminalSize();
  }
});

// 监听行高变化
watch(() => props.lineHeight, (newLineHeight) => {
  if (term.value) {
    term.value.options.lineHeight = parseFloat(newLineHeight);
    changeTerminalSize();
  }
});

// 监听字间距变化
watch(() => props.letterSpacing, (newLetterSpacing) => {
  if (term.value) {
    term.value.options.letterSpacing = parseFloat(newLetterSpacing);
    changeTerminalSize();
  }
});

// 监听光标闪烁变化
watch(() => props.cursorBlink, (newCursorBlink) => {
  if (term.value) {
    term.value.options.cursorBlink = newCursorBlink === 'enable';
  }
});

// 监听光标样式变化
watch(() => props.cursorStyle, (newCursorStyle) => {
  if (term.value) {
    term.value.options.cursorStyle = newCursorStyle as any;
  }
});

// 监听滚动行数变化
watch(() => props.scrollback, (newScrollback) => {
  if (term.value) {
    term.value.options.scrollback = parseInt(newScrollback);
  }
});

// 监听滚动灵敏度变化
watch(() => props.scrollSensitivity, (newScrollSensitivity) => {
  if (term.value) {
    term.value.options.scrollSensitivity = parseInt(newScrollSensitivity);
  }
});

const readyWatcher = watch(
  () => webSocketReady.value && termReady.value,
  (ready) => {
    if (ready) {
      changeTerminalSize();
      readyWatcher();
    }
  }
);

const newTerm = () => {
  const background = props.theme === 'dark' ? '#1e1e1e' : '#ffffff';
  const foreground = props.theme === 'dark' ? '#ffffff' : '#000000';

  term.value = new Terminal({
    lineHeight: parseFloat(props.lineHeight),
    fontSize: props.fontSize,
    fontFamily: "Monaco, Menlo, Consolas, 'Courier New', monospace",
    theme: {
      background,
      foreground,
      cursor: foreground,
      selection: props.theme === 'dark' ? 'rgba(255, 255, 255, 0.3)' : 'rgba(0, 0, 0, 0.3)',
      black: '#000000',
      red: '#cd3131',
      green: '#0dbc79',
      yellow: '#e5e510',
      blue: '#2472c8',
      magenta: '#bc3fbc',
      cyan: '#11a8cd',
      white: '#e5e5e5',
      brightBlack: '#666666',
      brightRed: '#f14c4c',
      brightGreen: '#23d18b',
      brightYellow: '#f5f543',
      brightBlue: '#3b8eea',
      brightMagenta: '#d670d6',
      brightCyan: '#29b8db',
      brightWhite: '#ffffff'
    },
    cursorBlink: props.cursorBlink === 'enable',
    cursorStyle: props.cursorStyle as any,
    scrollback: parseInt(props.scrollback),
    scrollSensitivity: parseInt(props.scrollSensitivity)
  });
};

const initTerminal = (online: boolean = false): boolean => {
  newTerm();
  if (terminalElement.value && term.value) {
    term.value.open(terminalElement.value);
    term.value.loadAddon(fitAddon);
    window.addEventListener('resize', changeTerminalSize);
    if (online) {
      term.value.onData((data) => sendMsg(data));
    }
    termReady.value = true;
  }
  return termReady.value;
};

const initWebSocket = () => {
  let url = '';
  if (props.mode === 'local') {
    url = getLocalTerminalUrl(terminalConfig.value);
  } else if (props.mode === 'ssh' && props.sshConfig) {
    url = getSSHTerminalUrl(terminalConfig.value);
  }

  if (!url) {
    console.error('Invalid terminal mode or missing config');
    return;
  }

  console.log('Connecting to terminal:', url);
  terminalSocket.value = new WebSocket(url);
  terminalSocket.value.onopen = runRealTerminal;
  terminalSocket.value.onmessage = onWSReceive;
  terminalSocket.value.onclose = closeRealTerminal;
  terminalSocket.value.onerror = errorRealTerminal;

  // 启动心跳检测
  heartbeatTimer.value = setInterval(() => {
    if (isWsOpen()) {
      terminalSocket.value!.send(
        JSON.stringify({
          type: 'heartbeat',
          timestamp: new Date().getTime()
        } as WsMsg)
      );
    }
  }, 10000);
};

const runRealTerminal = () => {
  webSocketReady.value = true;

  // 如果是 SSH 模式，先发送连接消息（包含 SSH 凭证）
  if (props.mode === 'ssh' && props.sshConfig && terminalSocket.value) {
    const connectMsg = JSON.stringify({
      type: 'connect',
      data: Base64.encode(JSON.stringify(props.sshConfig))
    } as WsMsg);
    terminalSocket.value.send(connectMsg);
  }

  emit('connected');
};

const onWSReceive = (message: MessageEvent) => {
  const wsMsg: WsMsg = JSON.parse(message.data);
  switch (wsMsg.type) {
    case 'cmd': {
      if (term.value && term.value.element) {
        term.value.focus();
        if (wsMsg.data) {
          const receiveMsg = Base64.decode(wsMsg.data);
          term.value.write(receiveMsg);
        }
      }
      break;
    }
    case 'heartbeat': {
      latency.value = new Date().getTime() - (wsMsg.timestamp || 0);
      break;
    }
  }
};

const errorRealTerminal = (ex: Event) => {
  let message = (ex as any).message;
  if (!message) message = 'disconnected';
  if (term.value) {
    term.value.write(`\x1b[31m${message}\x1b[m\r\n`);
  }
  emit('error', ex);
};

const closeRealTerminal = (ev: CloseEvent) => {
  if (heartbeatTimer.value) {
    clearInterval(heartbeatTimer.value);
    heartbeatTimer.value = null;
  }
  webSocketReady.value = false;
  if (term.value) {
    term.value.write('\r\n\x1b[31mThe connection has been disconnected.\x1b[m\r\n');
    if (ev.reason) {
      term.value.write(ev.reason + '\r\n');
    }
  }
  emit('disconnected', ev);
};

const changeTerminalSize = () => {
  if (term.value) {
    fitAddon.fit();
    if (isWsOpen()) {
      const { cols, rows } = term.value;
      terminalSocket.value!.send(
        JSON.stringify({
          type: 'resize',
          cols: cols,
          rows: rows
        } as WsMsg)
      );
    }
  }
};

const isWsOpen = (): boolean => {
  const readyState = terminalSocket.value && terminalSocket.value.readyState;
  return readyState === WebSocket.OPEN;
};

const sendMsg = (data: string) => {
  if (isWsOpen()) {
    terminalSocket.value!.send(
      JSON.stringify({
        type: 'cmd',
        data: Base64.encode(data)
      } as WsMsg)
    );
  }
};

const onClose = (isKeepShow: boolean = false) => {
  // 清理事件监听器
  window.removeEventListener('resize', changeTerminalSize);

  // 清理心跳定时器
  if (heartbeatTimer.value) {
    try {
      clearInterval(heartbeatTimer.value);
    } catch {
      // 忽略错误
    } finally {
      heartbeatTimer.value = null;
    }
  }

  // 关闭 WebSocket 连接
  if (terminalSocket.value) {
    try {
      if (terminalSocket.value.readyState === WebSocket.OPEN ||
          terminalSocket.value.readyState === WebSocket.CONNECTING) {
        terminalSocket.value.close();
      }
    } catch {
      // 忽略错误
    } finally {
      terminalSocket.value = null;
    }
  }

  // 释放终端实例
  if (!isKeepShow && term.value) {
    try {
      term.value.dispose();
    } catch {
      // 忽略错误
    } finally {
      term.value = null;
    }
  }

  // 清理容器
  if (terminalElement.value) {
    terminalElement.value.innerHTML = '';
  }

  // 更新状态
  webSocketReady.value = false;
};

const reconnect = () => {
  onClose(true);
  webSocketReady.value = false;
  initWebSocket();
};

const getLatency = (): number => {
  return latency.value;
};

// 初始化
onMounted(() => {
  if (initTerminal(true)) {
    initWebSocket();
  }

  // 使用 ResizeObserver 监听容器大小变化
  resizeObserver.value = new ResizeObserver(() => {
    if (termReady.value && webSocketReady.value) {
      changeTerminalSize();
    }
  });

  if (terminalElement.value) {
    resizeObserver.value.observe(terminalElement.value);
  }
});

onBeforeUnmount(() => {
  onClose();
  resizeObserver.value?.disconnect();
});

// 暴露方法给父组件
defineExpose({
  onClose,
  reconnect,
  isWsOpen,
  sendMsg,
  getLatency
});
</script>

<style scoped>
.terminal-container {
  width: 100%;
  height: 100%;
}

:deep(.xterm) {
  padding: 5px !important;
  background-color: #1e1e1e !important;
}

:deep(.xterm .xterm-viewport) {
  background-color: #1e1e1e !important;
}

:deep(.xterm .xterm-screen) {
  background-color: #1e1e1e !important;
}

:deep(.xterm .xterm-rows) {
  background-color: #1e1e1e !important;
}
</style>
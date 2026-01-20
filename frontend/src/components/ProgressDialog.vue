<template>
  <el-dialog
    v-model="dialogVisible"
    title="操作进度"
    width="500px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <div class="progress-container">
      <div class="progress-info">
        <span class="progress-name">{{ progress?.name }}</span>
        <span class="progress-percent">{{ progress?.percent?.toFixed(2) }}%</span>
      </div>
      <el-progress
        :percentage="progress?.percent || 0"
        :status="progressStatus"
        :stroke-width="20"
      />
      <div class="progress-message">{{ progress?.message }}</div>
    </div>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue';
import { fileApi } from '@/api/modules/file';
import { getWebSocket, type WebSocketMessage, type ProgressInfo } from '@/utils/websocket';
import type { ProgressInfo as ApiProgressInfo } from '@/api/interface/file';

const props = defineProps<{
  visible: boolean;
  progressKey: string;
}>();

const emit = defineEmits<{
  close: [];
}>();

const dialogVisible = ref(false);
const progress = ref<ApiProgressInfo | null>(null);
let timer: number | null = null;
let wsHandler: ((message: WebSocketMessage) => void) | null = null;

const progressStatus = computed(() => {
  if (!progress.value) return undefined;
  switch (progress.value.status) {
    case 'completed':
      return 'success';
    case 'failed':
      return 'exception';
    default:
      return undefined;
  }
});

const handleWebSocketMessage = (message: WebSocketMessage) => {
  if (message.type === 'progress' && message.data.key === props.progressKey) {
    progress.value = message.data.progress as ApiProgressInfo;
    
    // 如果操作完成或失败，停止轮询
    if (progress.value.status === 'completed' || progress.value.status === 'failed') {
      if (timer) {
        clearInterval(timer);
        timer = null;
      }
    }
  }
};

const loadProgress = async () => {
  if (!props.progressKey) return;
  
  try {
    const response = await fileApi.getProgress(props.progressKey);
    if (response.data.code === 200) {
      progress.value = response.data.data;
      
      // 如果操作完成或失败，停止轮询
      if (progress.value.status === 'completed' || progress.value.status === 'failed') {
        if (timer) {
          clearInterval(timer);
          timer = null;
        }
      }
    }
  } catch (error: any) {
    console.error('Failed to load progress:', error);
  }
};

const startPolling = () => {
  if (timer) {
    clearInterval(timer);
  }
  timer = window.setInterval(loadProgress, 1000);
};

const stopPolling = () => {
  if (timer) {
    clearInterval(timer);
    timer = null;
  }
};

const handleClose = () => {
  stopPolling();
  emit('close');
};

watch(() => props.visible, (newVal) => {
  dialogVisible.value = newVal;
  if (newVal && props.progressKey) {
    // 尝试使用 WebSocket
    const ws = getWebSocket();
    if (ws && ws.isConnected()) {
      // 注册 WebSocket 消息处理器
      wsHandler = handleWebSocketMessage;
      ws.on('progress', wsHandler);
      
      // 初始加载一次进度
      loadProgress();
    } else {
      // WebSocket 不可用，使用轮询
      loadProgress();
      startPolling();
    }
  }
});

watch(dialogVisible, (newVal) => {
  if (!newVal) {
    stopPolling();
    
    // 取消 WebSocket 消息处理器
    const ws = getWebSocket();
    if (ws && wsHandler) {
      ws.off('progress', wsHandler);
      wsHandler = null;
    }
    
    emit('close');
  }
});

onUnmounted(() => {
  stopPolling();
  
  // 取消 WebSocket 消息处理器
  const ws = getWebSocket();
  if (ws && wsHandler) {
    ws.off('progress', wsHandler);
    wsHandler = null;
  }
});
</script>

<style scoped>
.progress-container {
  padding: 20px 0;
}

.progress-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.progress-name {
  font-size: 16px;
  font-weight: 500;
}

.progress-percent {
  font-size: 14px;
  color: #409eff;
}

.progress-message {
  margin-top: 12px;
  font-size: 14px;
  color: #606266;
  text-align: center;
}
</style>
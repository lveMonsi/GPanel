<template>
  <div class="firewall-install">
    <!-- 安装选项 -->
    <div class="install-options">
      <span class="label">选择防火墙类型：</span>
      <el-radio-group v-model="selectedType" :disabled="installing">
        <el-radio-button label="ufw">
          <el-tooltip content="Ubuntu/Debian 推荐使用" placement="top">
            <span>UFW</span>
          </el-tooltip>
        </el-radio-button>
        <el-radio-button label="iptables">
          <el-tooltip content="Linux 内核级别防火墙" placement="top">
            <span>iptables</span>
          </el-tooltip>
        </el-radio-button>
        <el-radio-button label="firewalld">
          <el-tooltip content="CentOS/RHEL 推荐使用" placement="top">
            <span>firewalld</span>
          </el-tooltip>
        </el-radio-button>
      </el-radio-group>
      <el-button
        type="primary"
        :loading="installing"
        :disabled="installing"
        @click="handleInstall"
        style="margin-left: 16px"
      >
        {{ installing ? '安装中...' : '一键安装' }}
      </el-button>
    </div>

    <!-- 安装进度 -->
    <transition name="fade">
      <div v-if="installing || completed || error" class="install-progress">
        <div class="progress-header">
          <span class="status-text">
            <el-icon v-if="installing" class="is-loading" style="margin-right: 8px">
              <Loading />
            </el-icon>
            <el-icon v-else-if="completed" style="color: #67c23a; margin-right: 8px">
              <CircleCheckFilled />
            </el-icon>
            <el-icon v-else-if="error" style="color: #f56c6c; margin-right: 8px">
              <CircleCloseFilled />
            </el-icon>
            {{ statusMessage }}
          </span>
          <el-button
            v-if="completed || error"
            type="primary"
            size="small"
            @click="reset"
          >
            {{ completed ? '完成' : '重试' }}
          </el-button>
        </div>

        <!-- 进度条 -->
        <el-progress
          :percentage="progress"
          :status="progressStatus"
          :stroke-width="12"
          style="margin: 16px 0"
        />

        <!-- 日志展开/收起 -->
        <div class="log-section">
          <div class="log-header" @click="logExpanded = !logExpanded">
            <span>
              <el-icon style="margin-right: 4px">
                <ArrowDown v-if="logExpanded" />
                <ArrowRight v-else />
              </el-icon>
              安装日志
            </span>
            <el-tag size="small" type="info">{{ logs.length }} 条</el-tag>
          </div>
          <transition name="slide">
            <div v-show="logExpanded" class="log-content">
              <div class="log-container" ref="logContainer">
                <div
                  v-for="(log, index) in logs"
                  :key="index"
                  class="log-line"
                  :class="getLogClass(log)"
                >
                  {{ log }}
                </div>
                <div v-if="logs.length === 0" class="log-empty">
                  等待安装日志...
                </div>
              </div>
            </div>
          </transition>
        </div>
      </div>
    </transition>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { Loading, CircleCheckFilled, CircleCloseFilled, ArrowDown, ArrowRight } from '@element-plus/icons-vue';
import { installFirewall } from '@/api/modules/firewall';
import type { InstallProgress } from '@/api/interface/firewall';

const emit = defineEmits<{
  (e: 'installed'): void;
  (e: 'cancel'): void;
}>();

const selectedType = ref<'ufw' | 'iptables' | 'firewalld'>('ufw');
const installing = ref(false);
const completed = ref(false);
const error = ref(false);
const progress = ref(0);
const statusMessage = ref('');
const logs = ref<string[]>([]);
const logExpanded = ref(true);
const logContainer = ref<HTMLElement | null>(null);

// 进度条状态
const progressStatus = computed(() => {
  if (error.value) return 'exception';
  if (completed.value) return 'success';
  return '';
});

// 获取日志样式类
const getLogClass = (log: string) => {
  if (log.startsWith('[ERROR]') || log.startsWith('[WARN]')) {
    return 'log-error';
  }
  if (log.startsWith('[INFO]')) {
    return 'log-info';
  }
  if (log.startsWith('[CMD]')) {
    return 'log-cmd';
  }
  return '';
};

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight;
    }
  });
};

// 监听日志变化，自动滚动
watch(logs, () => {
  scrollToBottom();
}, { deep: true });

// 处理安装
const handleInstall = () => {
  installing.value = true;
  completed.value = false;
  error.value = false;
  progress.value = 0;
  statusMessage.value = '正在连接服务...';
  logs.value = [];
  logExpanded.value = true;

  const ws = installFirewall(selectedType.value);

  ws.onopen = () => {
    logs.value.push('[INFO] WebSocket 连接成功，开始安装...');
  };

  ws.onmessage = (event) => {
    try {
      const data: InstallProgress = JSON.parse(event.data);
      
      // 处理不同类型的消息
      switch (data.type) {
        case 'progress':
          progress.value = data.progress;
          statusMessage.value = data.message;
          if (data.log) {
            logs.value.push(data.log);
          }
          break;
        case 'log':
          if (data.log) {
            logs.value.push(data.log);
          }
          break;
        case 'error':
          error.value = true;
          installing.value = false;
          statusMessage.value = data.message || '安装失败';
          if (data.log) {
            logs.value.push(data.log);
          }
          ws.close();
          ElMessage.error(data.message || '安装失败');
          break;
        case 'complete':
          completed.value = true;
          installing.value = false;
          progress.value = 100;
          statusMessage.value = data.message || '安装完成！';
          if (data.log) {
            logs.value.push(data.log);
          }
          ws.close();
          ElMessage.success('防火墙安装成功！');
          emit('installed');
          break;
      }
    } catch (e) {
      console.error('Failed to parse WebSocket message:', e);
    }
  };

  ws.onerror = (err) => {
    console.error('WebSocket error:', err);
    error.value = true;
    installing.value = false;
    statusMessage.value = '连接失败';
    logs.value.push('[ERROR] WebSocket 连接失败，请检查网络或服务状态');
    ElMessage.error('连接失败，请检查服务是否正常运行');
  };

  ws.onclose = () => {
    if (installing.value) {
      logs.value.push('[INFO] 连接已关闭');
    }
  };
};

// 重置状态
const reset = () => {
  if (completed.value) {
    emit('cancel');
  } else {
    installing.value = false;
    completed.value = false;
    error.value = false;
    progress.value = 0;
    statusMessage.value = '';
    logs.value = [];
  }
};
</script>

<style scoped>
.firewall-install {
  padding: 16px;
  background: #fff;
  border-radius: 8px;
}

.install-options {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.label {
  font-weight: 500;
  color: #606266;
}

.install-progress {
  margin-top: 20px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 8px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-text {
  display: flex;
  align-items: center;
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.log-section {
  margin-top: 12px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  overflow: hidden;
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  background: #fff;
  cursor: pointer;
  user-select: none;
  transition: background 0.2s;
}

.log-header:hover {
  background: #f5f7fa;
}

.log-content {
  background: #1e1e1e;
}

.log-container {
  max-height: 300px;
  overflow-y: auto;
  padding: 12px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.6;
}

.log-line {
  color: #d4d4d4;
  white-space: pre-wrap;
  word-break: break-all;
}

.log-line.log-error {
  color: #f56c6c;
}

.log-line.log-info {
  color: #409eff;
}

.log-line.log-cmd {
  color: #67c23a;
}

.log-empty {
  color: #909399;
  text-align: center;
  padding: 20px;
}

/* 过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease;
  overflow: hidden;
}

.slide-enter-from,
.slide-leave-to {
  max-height: 0;
  opacity: 0;
}

.slide-enter-to,
.slide-leave-from {
  max-height: 300px;
  opacity: 1;
}

/* 自定义滚动条 */
.log-container::-webkit-scrollbar {
  width: 6px;
}

.log-container::-webkit-scrollbar-track {
  background: #2d2d2d;
}

.log-container::-webkit-scrollbar-thumb {
  background: #555;
  border-radius: 3px;
}

.log-container::-webkit-scrollbar-thumb:hover {
  background: #666;
}
</style>

<template>
  <div class="firewall-uninstall">
    <!-- 卸载选项 -->
    <div class="uninstall-options">
      <div class="option-row">
        <span class="label">当前防火墙：</span>
        <el-tag type="info" size="large">{{ firewallName }}</el-tag>
      </div>
      
      <div class="option-row">
        <span class="label">卸载选项：</span>
        <div class="checkbox-group">
          <el-checkbox v-model="keepRules" :disabled="uninstalling">
            保留规则数据
          </el-checkbox>
          <el-tooltip content="卸载时备份防火墙规则，重新安装后可手动恢复" placement="top">
            <el-icon class="tip-icon"><QuestionFilled /></el-icon>
          </el-tooltip>
        </div>
        <div class="checkbox-group">
          <el-checkbox v-model="keepPolicies" :disabled="uninstalling">
            保留策略配置
          </el-checkbox>
          <el-tooltip content="保留防火墙配置��件，卸载后不会删除配置目录" placement="top">
            <el-icon class="tip-icon"><QuestionFilled /></el-icon>
          </el-tooltip>
        </div>
      </div>

      <div class="warning-text" v-if="!keepRules && !keepPolicies">
        <el-icon><WarningFilled /></el-icon>
        <span>完全卸载将清除所有规则和配置，重新安装后需要重新配置！</span>
      </div>

      <div class="button-row">
        <el-button @click="handleCancel">取消</el-button>
        <el-button
          type="danger"
          :loading="uninstalling"
          :disabled="uninstalling"
          @click="handleUninstall"
        >
          {{ uninstalling ? '卸载中...' : '确认卸载' }}
        </el-button>
      </div>
    </div>

    <!-- 卸载进度 -->
    <transition name="fade">
      <div v-if="uninstalling || completed || error" class="uninstall-progress">
        <div class="progress-header">
          <span class="status-text">
            <el-icon v-if="uninstalling" class="is-loading" style="margin-right: 8px">
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
              卸载日志
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
                  等待卸载日志...
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
import { ElMessage, ElMessageBox } from 'element-plus';
import { 
  Loading, 
  CircleCheckFilled, 
  CircleCloseFilled, 
  ArrowDown, 
  ArrowRight, 
  QuestionFilled,
  WarningFilled 
} from '@element-plus/icons-vue';
import { uninstallFirewall } from '@/api/modules/firewall';
import type { UninstallProgress } from '@/api/interface/firewall';

interface Props {
  firewallName: string;
  firewallType: 'ufw' | 'iptables' | 'firewalld';
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: 'uninstalled'): void;
  (e: 'cancel'): void;
}>();

const keepRules = ref(true);
const keepPolicies = ref(false);
const uninstalling = ref(false);
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

// 处理卸载
const handleUninstall = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要卸载 ${props.firewallName} 防火墙吗？此操作不可撤销！`,
      '确认卸载',
      {
        type: 'warning',
        confirmButtonText: '确认卸载',
        cancelButtonText: '取消',
      }
    );
  } catch {
    return;
  }

  uninstalling.value = true;
  completed.value = false;
  error.value = false;
  progress.value = 0;
  statusMessage.value = '正在连接服务...';
  logs.value = [];
  logExpanded.value = true;

  const ws = uninstallFirewall({
    type: props.firewallType,
    keepRules: keepRules.value,
    keepPolicies: keepPolicies.value
  });

  ws.onopen = () => {
    logs.value.push('[INFO] WebSocket 连接成功，开始卸载...');
  };

  ws.onmessage = (event) => {
    try {
      const data: UninstallProgress = JSON.parse(event.data);
      
      // 处理不同类型的消息
      switch (data.type) {
        case 'progress':
          progress.value = data.progress || 0;
          statusMessage.value = data.message || '';
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
          uninstalling.value = false;
          statusMessage.value = data.message || '卸载失败';
          if (data.log) {
            logs.value.push(data.log);
          }
          ws.close();
          ElMessage.error(data.message || '卸载失败');
          break;
        case 'complete':
          completed.value = true;
          uninstalling.value = false;
          progress.value = 100;
          statusMessage.value = data.message || '卸载完成！';
          if (data.log) {
            logs.value.push(data.log);
          }
          ws.close();
          ElMessage.success('防火墙卸载成功！');
          emit('uninstalled');
          break;
      }
    } catch (e) {
      console.error('Failed to parse WebSocket message:', e);
    }
  };

  ws.onerror = (err) => {
    console.error('WebSocket error:', err);
    error.value = true;
    uninstalling.value = false;
    statusMessage.value = '连接失败';
    logs.value.push('[ERROR] WebSocket 连接失败，请检查网络或服务状态');
    ElMessage.error('连接失败，请检查服务是否正常运行');
  };

  ws.onclose = () => {
    if (uninstalling.value) {
      logs.value.push('[INFO] 连接已关闭');
    }
  };
};

// 取消
const handleCancel = () => {
  emit('cancel');
};

// 重置状态
const reset = () => {
  if (completed.value) {
    emit('cancel');
  } else {
    uninstalling.value = false;
    completed.value = false;
    error.value = false;
    progress.value = 0;
    statusMessage.value = '';
    logs.value = [];
  }
};
</script>

<style scoped>
.firewall-uninstall {
  padding: 16px;
  background: #fff;
  border-radius: 8px;
}

.uninstall-options {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.option-row {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.label {
  font-weight: 500;
  color: #606266;
  min-width: 100px;
}

.checkbox-group {
  display: flex;
  align-items: center;
  gap: 4px;
}

.tip-icon {
  color: #909399;
  cursor: help;
}

.warning-text {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: #fef0f0;
  border-radius: 6px;
  color: #f56c6c;
  font-size: 13px;
}

.button-row {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 8px;
}

.uninstall-progress {
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

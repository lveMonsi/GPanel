<template>
  <el-dialog
    :model-value="visible"
    title="下载进度"
    width="460px"
    :close-on-click-modal="false"
    :close-on-press-escape="status !== 'downloading'"
    :show-close="status !== 'downloading'"
    :before-close="handleBeforeClose"
    @update:model-value="handleVisibleChange"
  >
    <div class="download-progress-dialog">
      <div class="download-progress-header">
        <span class="download-progress-label">当前文件</span>
        <span class="download-progress-name">{{ fileName }}</span>
      </div>
      <el-progress
        :percentage="percentage"
        :status="status === 'failed' ? 'exception' : status === 'completed' ? 'success' : undefined"
        :stroke-width="14"
      />
      <div class="download-progress-meta">
        <span>{{ message }}</span>
        <span>{{ transferText }}</span>
      </div>
    </div>
    <template #footer>
      <el-button v-if="status === 'failed'" @click="handleVisibleChange(false)">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import type { DownloadStatus } from '@/composables/useDownloadProgress';

defineProps<{
  visible: boolean;
  fileName: string;
  percentage: number;
  status: DownloadStatus;
  message: string;
  transferText: string;
}>();

const emit = defineEmits<{
  'update:visible': [visible: boolean];
}>();

const handleVisibleChange = (visible: boolean) => {
  emit('update:visible', visible);
};

const handleBeforeClose = (done: () => void) => {
  emit('update:visible', false);
  done();
};
</script>

<style scoped>
.download-progress-dialog {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.download-progress-header {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px 16px;
  background: #f5f7fa;
  border-radius: 10px;
}

.download-progress-label {
  font-size: 12px;
  color: #909399;
}

.download-progress-name {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  word-break: break-all;
}

.download-progress-meta {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  font-size: 13px;
  color: #606266;
}

@media (max-width: 768px) {
  .download-progress-meta {
    flex-direction: column;
  }
}
</style>

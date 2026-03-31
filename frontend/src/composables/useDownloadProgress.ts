import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { ElMessage } from 'element-plus';

import { fileApi } from '@/api/modules/file';

export type DownloadStatus = 'idle' | 'downloading' | 'completed' | 'failed';

interface DownloadTask {
  backendPath: string;
  fileName: string;
  size?: number;
}

const formatSize = (bytes: number): string => {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
};

const saveBlob = (blob: Blob, fileName: string) => {
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.setAttribute('download', fileName);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
};

export const useDownloadProgress = () => {
  const downloadDialogVisible = ref(false);
  const downloadStatus = ref<DownloadStatus>('idle');
  const downloadFileName = ref('');
  const downloadPercent = ref(0);
  const downloadLoaded = ref(0);
  const downloadTotal = ref(0);
  const downloadMessage = ref('准备下载...');
  let downloadDialogTimer: ReturnType<typeof window.setTimeout> | null = null;

  const downloadTransferText = computed(() => {
    if (downloadTotal.value > 0) {
      return `${formatSize(downloadLoaded.value)} / ${formatSize(downloadTotal.value)}`;
    }

    if (downloadLoaded.value > 0) {
      return formatSize(downloadLoaded.value);
    }

    return '等待开始...';
  });

  const clearDownloadDialogTimer = () => {
    if (downloadDialogTimer !== null) {
      window.clearTimeout(downloadDialogTimer);
      downloadDialogTimer = null;
    }
  };

  const resetDownloadState = () => {
    clearDownloadDialogTimer();
    downloadStatus.value = 'idle';
    downloadFileName.value = '';
    downloadPercent.value = 0;
    downloadLoaded.value = 0;
    downloadTotal.value = 0;
    downloadMessage.value = '准备下载...';
  };

  const handleDownloadDialogVisibleChange = (visible: boolean) => {
    if (!visible && downloadStatus.value === 'downloading') {
      return;
    }

    downloadDialogVisible.value = visible;
  };

  const downloadFile = async ({ backendPath, fileName, size = 0 }: DownloadTask) => {
    clearDownloadDialogTimer();
    downloadDialogVisible.value = true;
    downloadStatus.value = 'downloading';
    downloadFileName.value = fileName;
    downloadPercent.value = 0;
    downloadLoaded.value = 0;
    downloadTotal.value = size;
    downloadMessage.value = '正在准备下载...';

    try {
      const response = await fileApi.downloadFile(backendPath, (progressEvent) => {
        const loaded = progressEvent.loaded ?? 0;
        const total = progressEvent.total ?? size ?? 0;

        downloadLoaded.value = loaded;
        downloadTotal.value = total;
        downloadPercent.value = total > 0 ? Math.min(100, Math.round((loaded / total) * 100)) : 0;
        downloadMessage.value = total > 0
          ? '正在下载到浏览器...'
          : '正在下载，等待服务器返回大小信息...';
      });

      const blob = response.data instanceof Blob ? response.data : new Blob([response.data]);
      saveBlob(blob, fileName);

      const finalSize = downloadTotal.value || size || blob.size || downloadLoaded.value;
      downloadLoaded.value = finalSize;
      downloadTotal.value = finalSize;
      downloadPercent.value = 100;
      downloadStatus.value = 'completed';
      downloadMessage.value = '下载完成，已开始保存到本地';
      downloadDialogTimer = window.setTimeout(() => {
        downloadDialogVisible.value = false;
      }, 1200);
    } catch (error: any) {
      downloadStatus.value = 'failed';
      downloadMessage.value = error.message || '下载失败';
      ElMessage.error(error.message || '下载失败');
    }
  };

  watch(downloadDialogVisible, (visible) => {
    if (!visible && downloadStatus.value !== 'downloading') {
      resetDownloadState();
    }
  });

  onBeforeUnmount(() => {
    clearDownloadDialogTimer();
  });

  return {
    downloadDialogVisible,
    downloadStatus,
    downloadFileName,
    downloadPercent,
    downloadMessage,
    downloadTransferText,
    handleDownloadDialogVisibleChange,
    downloadFile,
  };
};

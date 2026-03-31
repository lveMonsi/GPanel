<template>
  <el-dialog
    v-model="dialogVisible"
    :title="`预览: ${fileName}`"
    width="80%"
    top="5vh"
    @close="handleClose"
  >
    <div v-loading="loading" class="preview-container">
      <!-- 图片预览 -->
      <div v-if="previewType === 'image'" class="image-preview">
        <img :src="previewContent" :alt="fileName" />
      </div>
      
      <!-- 文本预览 -->
      <div v-else-if="previewType === 'text'" class="text-preview">
        <pre>{{ previewContent }}</pre>
      </div>
      
      <!-- 视频预览 -->
      <div v-else-if="previewType === 'video'" class="video-preview">
        <video :src="previewContent" controls>
          您的浏览器不支持视频播放
        </video>
      </div>
      
      <!-- 音频预览 -->
      <div v-else-if="previewType === 'audio'" class="audio-preview">
        <audio :src="previewContent" controls>
          您的浏览器不支持音频播放
        </audio>
      </div>
      
      <!-- PDF 预览 -->
      <div v-else-if="previewType === 'pdf'" class="pdf-preview">
        <iframe :src="previewContent" frameborder="0"></iframe>
      </div>
      
      <!-- 不支持预览 -->
      <div v-else class="no-preview">
        <el-icon :size="64"><Document /></el-icon>
        <p>此文件类型不支持预览</p>
        <el-button type="primary" @click="handleDownloadFile">下载文件</el-button>
      </div>
    </div>
  </el-dialog>

  <DownloadProgressDialog
    :visible="downloadDialogVisible"
    :file-name="downloadFileName"
    :percentage="downloadPercent"
    :status="downloadStatus"
    :message="downloadMessage"
    :transfer-text="downloadTransferText"
    @update:visible="handleDownloadDialogVisibleChange"
  />
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { ElMessage } from 'element-plus';
import { Document } from '@element-plus/icons-vue';
import { fileApi } from '@/api/modules/file';
import type { FileInfo } from '@/api/interface/file';
import DownloadProgressDialog from '@/components/DownloadProgressDialog.vue';
import { useDownloadProgress } from '@/composables/useDownloadProgress';

const props = defineProps<{
  visible: boolean;
  file: FileInfo | null;
}>();

const emit = defineEmits<{
  close: [];
}>();

const dialogVisible = ref(false);
const loading = ref(false);
const previewContent = ref('');
const previewType = ref<'image' | 'text' | 'video' | 'audio' | 'pdf' | 'none'>('none');

const {
  downloadDialogVisible,
  downloadStatus,
  downloadFileName,
  downloadPercent,
  downloadMessage,
  downloadTransferText,
  handleDownloadDialogVisibleChange,
  downloadFile: startDownloadWithProgress,
} = useDownloadProgress();

const fileName = computed(() => props.file?.name || '');

const getPreviewType = (file: FileInfo): 'image' | 'text' | 'video' | 'audio' | 'pdf' | 'none' => {
  const ext = file.extension.toLowerCase();
  
  // 图片类型
  const imageExts = ['.jpg', '.jpeg', '.png', '.gif', '.bmp', '.webp', '.svg', '.ico'];
  if (imageExts.includes(ext)) {
    return 'image';
  }
  
  // 文本类型
  const textExts = ['.txt', '.md', '.json', '.xml', '.html', '.css', '.js', '.ts', '.py', '.go', '.java', '.c', '.cpp', '.h', '.sh', '.yaml', '.yml', '.toml', '.ini', '.conf', '.log'];
  if (textExts.includes(ext) || file.mimeType.startsWith('text/')) {
    return 'text';
  }
  
  // 视频类型
  const videoExts = ['.mp4', '.avi', '.mov', '.wmv', '.flv', '.mkv', '.webm'];
  if (videoExts.includes(ext) || file.mimeType.startsWith('video/')) {
    return 'video';
  }
  
  // 音频类型
  const audioExts = ['.mp3', '.wav', '.ogg', '.flac', '.aac', '.m4a'];
  if (audioExts.includes(ext) || file.mimeType.startsWith('audio/')) {
    return 'audio';
  }
  
  // PDF 类型
  if (ext === '.pdf' || file.mimeType === 'application/pdf') {
    return 'pdf';
  }
  
  return 'none';
};

const loadPreview = async () => {
  if (!props.file) return;
  
  loading.value = true;
  previewType.value = getPreviewType(props.file);
  
  try {
    const response = await fileApi.previewFile({
      path: props.file.path,
    });
    
    if (response.data.code === 200) {
      const data = response.data.data;
      
      if (previewType.value === 'image' || previewType.value === 'video' || previewType.value === 'audio') {
        // Base64 编码的媒体文件
        previewContent.value = data.content;
      } else if (previewType.value === 'text') {
        // 文本文件直接显示
        previewContent.value = data.content;
      } else if (previewType.value === 'pdf') {
        // PDF 文件使用 Blob URL
        const blob = base64ToBlob(data.content, 'application/pdf');
        previewContent.value = URL.createObjectURL(blob);
      }
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载预览失败');
  } finally {
    loading.value = false;
  }
};

const base64ToBlob = (base64: string, mimeType: string): Blob => {
  const byteCharacters = atob(base64);
  const byteNumbers = new Array(byteCharacters.length);
  
  for (let i = 0; i < byteCharacters.length; i++) {
    byteNumbers[i] = byteCharacters.charCodeAt(i);
  }
  
  const byteArray = new Uint8Array(byteNumbers);
  return new Blob([byteArray], { type: mimeType });
};

const handleDownloadFile = async () => {
  if (!props.file) return;

  await startDownloadWithProgress({
    backendPath: props.file.path,
    fileName: props.file.name,
    size: props.file.size,
  });
};

const handleClose = () => {
  // 清理 Blob URL
  if (previewType.value === 'pdf' && previewContent.value.startsWith('blob:')) {
    URL.revokeObjectURL(previewContent.value);
  }
  previewContent.value = '';
  emit('close');
};

watch(() => props.visible, (newVal) => {
  dialogVisible.value = newVal;
  if (newVal && props.file) {
    loadPreview();
  }
});

watch(dialogVisible, (newVal) => {
  if (!newVal) {
    handleClose();
  }
});
</script>

<style scoped>
.preview-container {
  min-height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.image-preview img {
  max-width: 100%;
  max-height: 600px;
  object-fit: contain;
}

.text-preview {
  width: 100%;
  max-height: 600px;
  overflow: auto;
}

.text-preview pre {
  margin: 0;
  padding: 20px;
  background-color: #f5f5f5;
  border-radius: 4px;
  white-space: pre-wrap;
  word-wrap: break-word;
  font-family: 'Courier New', Courier, monospace;
  font-size: 14px;
  line-height: 1.6;
}

.video-preview video,
.audio-preview audio {
  max-width: 100%;
  max-height: 600px;
}

.pdf-preview iframe {
  width: 100%;
  height: 600px;
  border: none;
}

.no-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 20px;
  color: #909399;
}

.no-preview p {
  margin: 0;
  font-size: 16px;
}
</style>

<template>
  <el-dialog
    v-model="dialogVisible"
    :title="title"
    width="80%"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <div v-if="loading" class="loading-container">
      <el-icon class="is-loading"><Loading /></el-icon>
      <span>加载中...</span>
    </div>
    <div v-else class="editor-container">
      <textarea
        v-model="content"
        class="file-editor"
        spellcheck="false"
      />
    </div>
    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { ElMessage } from 'element-plus';
import { Loading } from '@element-plus/icons-vue';
import { fileApi } from '@/api/modules/file';

const props = defineProps<{
  visible: boolean;
  filePath: string;
}>();

const emit = defineEmits<{
  close: [];
  save: [content: string];
}>();

const dialogVisible = ref(false);
const content = ref('');
const loading = ref(false);
const saving = ref(false);

const title = computed(() => {
  if (!props.filePath) return '文件编辑器';
  const parts = props.filePath.split('/');
  return `编辑文件: ${parts[parts.length - 1]}`;
});

const loadFileContent = async () => {
  if (!props.filePath) return;
  
  loading.value = true;
  content.value = '';
  
  try {
    const response = await fileApi.getFileContent({ path: props.filePath });
    if (response.data.code === 200) {
      content.value = response.data.data.content;
    } else {
      ElMessage.error(response.data.message);
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载文件内容失败');
  } finally {
    loading.value = false;
  }
};

const handleSave = () => {
  emit('save', content.value);
};

const handleClose = () => {
  emit('close');
};

watch(() => props.visible, (newVal) => {
  dialogVisible.value = newVal;
  if (newVal && props.filePath) {
    loadFileContent();
  }
});

watch(dialogVisible, (newVal) => {
  if (!newVal) {
    emit('close');
  }
});
</script>

<style scoped>
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 400px;
  gap: 12px;
  color: #909399;
}

.editor-container {
  height: 500px;
}

.file-editor {
  width: 100%;
  height: 100%;
  padding: 12px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  resize: none;
  outline: none;
  background: #fafafa;
}

.file-editor:focus {
  border-color: #409eff;
  background: white;
}
</style>
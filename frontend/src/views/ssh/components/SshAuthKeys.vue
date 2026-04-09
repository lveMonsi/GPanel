<template>
  <el-dialog v-model="visible" title="认证密钥" width="900px">
    <el-input
      v-model="content"
      type="textarea"
      :rows="18"
      placeholder="请输入 authorized_keys 内容"
    />
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSave">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { ElMessage } from 'element-plus';
import { getSSHFile, updateSSHFile } from '@/api/modules/ssh';

const visible = ref(false);
const loading = ref(false);
const content = ref('');

const open = async () => {
  visible.value = true;
  loading.value = true;
  try {
    const res = await getSSHFile('authKeys');
    content.value = res.data ?? '';
  } catch (error: any) {
    ElMessage.error(error.message || '加载认证密钥失败');
  } finally {
    loading.value = false;
  }
};

const handleSave = async () => {
  loading.value = true;
  try {
    await updateSSHFile('authKeys', content.value);
    ElMessage.success('认证密钥已保存');
    visible.value = false;
  } catch (error: any) {
    ElMessage.error(error.message || '保存认证密钥失败');
  } finally {
    loading.value = false;
  }
};

defineExpose({ open });
</script>

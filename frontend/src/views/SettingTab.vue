<template>
  <div class="setting-page">
    <el-card>
      <template #header>
        <span>终端设置</span>
      </template>

      <el-form :model="form" label-width="150px">
        <el-form-item label="行高">
          <el-input-number
            v-model="form.lineHeight"
            :min="1"
            :max="2"
            :step="0.1"
            :precision="1"
          />
        </el-form-item>
        <el-form-item label="字间距">
          <el-input-number
            v-model="form.letterSpacing"
            :min="0"
            :max="3"
            :step="0.5"
            :precision="1"
          />
        </el-form-item>
        <el-form-item label="字体大小">
          <el-input-number
            v-model="form.fontSize"
            :min="10"
            :max="24"
            :step="1"
          />
        </el-form-item>
        <el-form-item label="光标闪烁">
          <el-switch
            v-model="form.cursorBlink"
            active-value="enable"
            inactive-value="disable"
          />
        </el-form-item>
        <el-form-item label="光标样式">
          <el-select v-model="form.cursorStyle">
            <el-option label="块状" value="block" />
            <el-option label="下划线" value="underline" />
            <el-option label="竖线" value="bar" />
          </el-select>
        </el-form-item>
        <el-form-item label="滚动行数">
          <el-input-number
            v-model="form.scrollback"
            :min="0"
            :max="10000"
            :step="100"
          />
        </el-form-item>
        <el-form-item label="滚动灵敏度">
          <el-input-number
            v-model="form.scrollSensitivity"
            :min="0"
            :max="16"
            :step="1"
          />
        </el-form-item>
        <el-form-item>
          <el-button @click="resetToDefault">恢复默认</el-button>
          <el-button type="primary" @click="saveSettings">保存</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import {
  getTerminalSetting,
  updateTerminalSetting
} from '@/api/modules/terminal_setting';
import type { TerminalInfo, TerminalUpdate } from '@/api/interface/terminal_setting';

const form = reactive<TerminalInfo>({
  lineHeight: '1.2',
  letterSpacing: '1.2',
  fontSize: '14',
  cursorBlink: 'enable',
  cursorStyle: 'underline',
  scrollback: '1000',
  scrollSensitivity: '10'
});

const loadSettings = async () => {
  try {
    const res = await getTerminalSetting();
    Object.assign(form, res.data);
  } catch (error) {
    ElMessage.error('加载设置失败');
  }
};

const saveSettings = async () => {
  try {
    await updateTerminalSetting(form as TerminalUpdate);
    ElMessage.success('保存成功');
  } catch (error) {
    ElMessage.error('保存失败');
  }
};

const resetToDefault = () => {
  form.lineHeight = '1.2';
  form.letterSpacing = '1.2';
  form.fontSize = '14';
  form.cursorBlink = 'enable';
  form.cursorStyle = 'underline';
  form.scrollback = '1000';
  form.scrollSensitivity = '10';
};

const acceptParams = () => {
  loadSettings();
};

defineExpose({
  acceptParams
});

onMounted(() => {
  loadSettings();
});
</script>

<style scoped>
.setting-page {
  padding: 20px;
}
</style>
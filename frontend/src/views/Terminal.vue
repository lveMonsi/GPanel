<template>
  <div class="terminal-page">
    <el-card class="router-card">
      <el-radio-group v-model="activeTab" @change="handleTabChange">
        <el-radio-button value="terminal">终端</el-radio-button>
        <el-radio-button value="host">主机</el-radio-button>
        <el-radio-button value="command">快速命令</el-radio-button>
        <el-radio-button value="setting">设置</el-radio-button>
      </el-radio-group>
    </el-card>

    <div v-show="activeTab === 'terminal'">
      <TerminalTab ref="terminalTabRef" />
    </div>
    <div v-if="activeTab === 'host'">
      <HostTab ref="hostTabRef" />
    </div>
    <div v-if="activeTab === 'command'">
      <CommandTab ref="commandTabRef" />
    </div>
    <div v-if="activeTab === 'setting'">
      <SettingTab ref="settingTabRef" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import TerminalTab from './TerminalTab.vue';
import HostTab from './HostTab.vue';
import CommandTab from './CommandTab.vue';
import SettingTab from './SettingTab.vue';

const activeTab = ref('terminal');
const terminalTabRef = ref();
const hostTabRef = ref();
const commandTabRef = ref();
const settingTabRef = ref();

const handleTabChange = (tab: string) => {
  if (tab === 'host' && hostTabRef.value) {
    hostTabRef.value.acceptParams();
  }
  if (tab === 'command' && commandTabRef.value) {
    commandTabRef.value.acceptParams();
  }
  if (tab === 'terminal' && terminalTabRef.value) {
    terminalTabRef.value.acceptParams();
  }
  if (tab === 'setting' && settingTabRef.value) {
    settingTabRef.value.acceptParams();
  }
};

onMounted(() => {
  handleTabChange('terminal');
});
</script>

<style scoped>
.terminal-page {
  height: calc(100vh - 80px);
  display: flex;
  flex-direction: column;
}

.router-card {
  --el-card-padding: 0;
}

.router-card :deep(.el-radio-button__inner) {
  min-width: 100px;
  height: 100%;
  background-color: var(--el-fill-color-light);
  box-shadow: none !important;
  border: 2px solid transparent !important;
}

.router-card :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  color: var(--el-color-primary);
  border-color: var(--el-color-primary) !important;
  border-radius: 4px;
}
</style>
<template>
  <div class="terminal-page">
    <!-- 顶部标签切换 -->
    <div class="tabs-container">
      <el-radio-group v-model="activeTab" @change="handleTabChange" class="tab-radio-group">
        <el-radio-button value="terminal">
          <el-icon><Monitor /></el-icon>
          <span>终端</span>
        </el-radio-button>
        <el-radio-button value="host">
          <el-icon><Connection /></el-icon>
          <span>主机</span>
        </el-radio-button>
        <el-radio-button value="command">
          <el-icon><List /></el-icon>
          <span>快速命令</span>
        </el-radio-button>
        <el-radio-button value="setting">
          <el-icon><Setting /></el-icon>
          <span>设置</span>
        </el-radio-button>
      </el-radio-group>
    </div>

    <!-- 内容区域 -->
    <div class="content-area">
      <TerminalTab 
        v-show="activeTab === 'terminal'" 
        ref="terminalTabRef" 
      />
      <HostTab 
        v-show="activeTab === 'host'" 
        ref="hostTabRef" 
      />
      <CommandTab 
        v-show="activeTab === 'command'" 
        ref="commandTabRef" 
      />
      <SettingTab 
        v-show="activeTab === 'setting'" 
        ref="settingTabRef" 
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue';
import TerminalTab from './TerminalTab.vue';
import HostTab from './HostTab.vue';
import CommandTab from './CommandTab.vue';
import SettingTab from './SettingTab.vue';
import { Monitor, Connection, List, Setting } from '@element-plus/icons-vue';

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

onBeforeUnmount(() => {
  // 只在离开终端页面时才关闭所有连接
  if (terminalTabRef.value && terminalTabRef.value.cleanupTerminals) {
    terminalTabRef.value.cleanupTerminals();
  }
});
</script>

<style scoped>
.terminal-page {
  height: 100%;
  display: flex;
  flex-direction: column;
  background-color: var(--bg-color);
  padding: 0;
  overflow: hidden;
}

.tabs-container {
  padding: 16px 20px 0 20px;
  background-color: var(--bg-color);
}

.tab-radio-group {
  display: flex;
  gap: 4px;
}

.tab-radio-group :deep(.el-radio-button__inner) {
  min-width: 100px;
  padding: 10px 20px;
  background-color: var(--card-bg);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
  font-weight: 500;
  border-radius: var(--radius-md);
  box-shadow: none;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  gap: 6px;
}

.tab-radio-group :deep(.el-radio-button__inner:hover) {
  color: var(--primary);
  border-color: var(--primary);
}

.tab-radio-group :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background-color: var(--primary-light);
  border-color: var(--primary-dark);
  color: var(--text-primary);
  box-shadow: 0 1px 3px rgba(90, 184, 214, 0.24);
  font-weight: 600;
}

.tab-radio-group :deep(.el-radio-button:first-child .el-radio-button__inner) {
  border-radius: var(--radius-md);
}

.tab-radio-group :deep(.el-radio-button:last-child .el-radio-button__inner) {
  border-radius: var(--radius-md);
}

.content-area {
  flex: 1;
  padding: 16px 20px 20px 20px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* 过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>

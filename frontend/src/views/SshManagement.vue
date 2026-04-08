<template>
  <div class="ssh-management">
    <div class="nav-tabs">
      <div
        v-for="tab in tabs"
        :key="tab.name"
        :class="['tab-item', { active: activeTab === tab.name }]"
        @click="activeTab = tab.name"
      >
        {{ tab.label }}
      </div>
    </div>

    <div class="tab-content">
      <SshConfig v-if="activeTab === 'config'" />
      <SshSession v-else-if="activeTab === 'session'" />
      <SshLog v-else-if="activeTab === 'log'" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import SshConfig from './ssh/SshConfig.vue';
import SshSession from './ssh/SshSession.vue';
import SshLog from './ssh/SshLog.vue';

const activeTab = ref('config');
const tabs = [
  { name: 'config', label: '配置' },
  { name: 'session', label: '会话' },
  { name: 'log', label: '日志' },
];
</script>

<style scoped>
.ssh-management {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #fafafa;
}

.nav-tabs {
  display: flex;
  background: white;
  border-bottom: 1px solid #e4e7ed;
  padding: 0 20px;
}

.tab-item {
  padding: 16px 24px;
  cursor: pointer;
  color: #606266;
  font-size: 14px;
  border-bottom: 2px solid transparent;
  transition: all 0.3s;
}

.tab-item:hover {
  color: #409eff;
}

.tab-item.active {
  color: #409eff;
  border-bottom-color: #409eff;
}

.tab-content {
  flex: 1;
  overflow: auto;
}
</style>

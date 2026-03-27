<template>
  <div class="monitor">
    <el-tabs v-model="activeTab" class="monitor-tabs">
      <el-tab-pane label="监控报表" name="dashboard">
        <MonitorDashboard />
      </el-tab-pane>
      <el-tab-pane label="设置" name="settings">
        <MonitorSettings />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import MonitorDashboard from '@/components/MonitorDashboard.vue'
import MonitorSettings from '@/components/MonitorSettings.vue'

const activeTab = ref('dashboard')
</script>

<style scoped>
.monitor {
  padding: 1rem;
  background: var(--bg-color);
  min-height: calc(100vh - 60px);
}

.monitor-tabs {
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  padding: 1rem;
}
</style>

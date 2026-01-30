<template>
  <div class="firewall-container">
    <!-- Windows提示 -->
    <el-alert
      v-if="isWindows"
      type="warning"
      :closable="false"
      show-icon
      style="margin-bottom: 20px"
    >
      <template #title>
        <span>防火墙功能仅支持 Linux 系统</span>
      </template>
    </el-alert>

    <template v-else>
      <!-- 基础信息卡片 -->
      <el-card class="info-card" shadow="hover">
        <template #header>
          <div class="card-header">
            <span>防火墙状态</span>
            <div class="status-actions">
              <el-tag :type="baseInfo.isActive ? 'success' : 'danger'" size="large">
                {{ baseInfo.isActive ? '已启用' : '已禁用' }}
              </el-tag>
              <el-button-group style="margin-left: 20px">
                <el-button
                  type="primary"
                  :disabled="!baseInfo.isExist"
                  @click="handleOperate('start')"
                >
                  启动
                </el-button>
                <el-button
                  type="warning"
                  :disabled="!baseInfo.isExist"
                  @click="handleOperate('stop')"
                >
                  停止
                </el-button>
                <el-button
                  type="info"
                  :disabled="!baseInfo.isExist"
                  @click="handleOperate('restart')"
                >
                  重启
                </el-button>
              </el-button-group>
            </div>
          </div>
        </template>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="防火墙名称">
            {{ baseInfo.name }}
          </el-descriptions-item>
          <el-descriptions-item label="版本">
            {{ baseInfo.version }}
          </el-descriptions-item>
          <el-descriptions-item label="Ping防护">
            <el-switch
              v-model="pingEnabled"
              :loading="loading"
              @change="handlePingChange"
            />
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 标签页 -->
      <el-tabs v-model="activeTab" class="rule-tabs" @tab-change="handleTabChange">
        <el-tab-pane label="端口规则" name="port">
          <PortRules @refresh="loadRules" />
        </el-tab-pane>
        <el-tab-pane label="IP规则" name="address">
          <IPRules @refresh="loadRules" />
        </el-tab-pane>
        <el-tab-pane label="端口转发" name="forward">
          <ForwardRules @refresh="loadRules" />
        </el-tab-pane>
      </el-tabs>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { loadBaseInfo as loadBaseInfoApi, operateFirewall } from '@/api/modules/firewall';
import type { FirewallBaseInfo } from '@/api/interface/firewall';
import PortRules from '@/components/firewall/PortRules.vue';
import IPRules from '@/components/firewall/IPRules.vue';
import ForwardRules from '@/components/firewall/ForwardRules.vue';

const isWindows = ref(false);
const activeTab = ref('port');
const loading = ref(false);
const baseInfo = ref<FirewallBaseInfo>({
  name: '-',
  isExist: false,
  isActive: false,
  isInit: false,
  isBind: false,
  version: '-',
  pingStatus: 'disabled',
});

const pingEnabled = ref(false);

const loadBaseInfo = async () => {
  try {
    loading.value = true;
    const res = await loadBaseInfoApi();
    console.log('loadBaseInfo response:', res);
    if (res && res.data) {
      baseInfo.value = res.data;
      pingEnabled.value = res.data.pingStatus === 'enabled';
    }
  } catch (error) {
    console.error('加载防火墙信息失败:', error);
  } finally {
    loading.value = false;
  }
};

const handleOperate = async (operation: string) => {
  try {
    await ElMessageBox.confirm(`确认${operation === 'start' ? '启动' : operation === 'stop' ? '停止' : '重启'}防火墙吗?`, '确认操作', {
      type: 'warning',
    });
    loading.value = true;
    await operateFirewall({ operation });
    ElMessage.success('操作成功');
    await loadBaseInfo();
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('操作失败:', error);
      ElMessage.error(error.message || '操作失败');
    }
  } finally {
    loading.value = false;
  }
};

const handlePingChange = async (enabled: boolean) => {
  try {
    loading.value = true;
    const operation = enabled ? 'enableBanPing' : 'disableBanPing';
    await operateFirewall({ operation });
    ElMessage.success('设置成功');
    await loadBaseInfo();
  } catch (error: any) {
    console.error('设置失败:', error);
    ElMessage.error(error.message || '设置失败');
    await loadBaseInfo();
  } finally {
    loading.value = false;
  }
};

const handleTabChange = () => {
  loadBaseInfo();
};

const loadRules = () => {
  // 子组件调用刷新
};

// 检测操作系统
const checkOS = async () => {
  try {
    const response = await fetch('/api/v1/system/os');
    const data = await response.json();
    isWindows.value = data.os === 'windows';
  } catch (error) {
    console.error('检测操作系统失败:', error);
  }
};

onMounted(() => {
  checkOS();
  if (!isWindows.value) {
    loadBaseInfo();
  }
});
</script>

<style scoped>
.firewall-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 20px;
  background: #fafafa;
  overflow: hidden;
}

.info-card {
  margin-bottom: 20px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-actions {
  display: flex;
  align-items: center;
}

.rule-tabs {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: white;
  border-radius: 8px;
  padding: 16px;
}

.rule-tabs :deep(.el-tabs__header) {
  margin: 0 0 16px 0;
  padding: 0 16px;
}

.rule-tabs :deep(.el-tabs__content) {
  flex: 1;
  overflow: hidden;
}

.rule-tabs :deep(.el-tab-pane) {
  height: 100%;
  overflow: hidden;
}
</style>
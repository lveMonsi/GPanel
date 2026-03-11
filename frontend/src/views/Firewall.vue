<template>
  <div class="firewall-container">
    <!-- 初始加载状态 -->
    <div v-if="initialLoading" class="loading-container">
      <el-icon class="loading-icon" :size="48"><Loading /></el-icon>
      <p class="loading-text">正在加载防火墙信息...</p>
    </div>

    <!-- Windows提示 -->
    <el-alert
      v-else-if="isWindows"
      type="warning"
      :closable="false"
      show-icon
      style="margin-bottom: 20px"
    >
      <template #title>
        <span>防火墙功能仅支持 Linux 系统</span>
      </template>
    </el-alert>

    <!-- 未安装防火墙提示 -->
    <el-card v-else-if="!baseInfo.isExist" class="install-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <span>
            <el-icon style="color: #f56c6c; margin-right: 8px"><WarningFilled /></el-icon>
            系统未检测到防火墙管理工具
          </span>
          <el-button 
            v-if="!showInstall" 
            type="primary" 
            size="small"
            @click="showInstall = true"
          >
            一键安装
          </el-button>
        </div>
      </template>
      
      <template v-if="!showInstall">
        <div class="install-tips">
          <p style="margin-bottom: 12px">请安装防火墙管理工具后再使用此功能。推荐安装：</p>
          <ul style="margin: 0; padding-left: 20px">
            <li><strong>ufw</strong> (Ubuntu/Debian): <code>apt install ufw</code></li>
            <li><strong>firewalld</strong> (CentOS/RHEL): <code>yum install firewalld</code></li>
            <li><strong>iptables</strong>: <code>apt install iptables</code> 或 <code>yum install iptables</code></li>
          </ul>
        </div>
      </template>
      
      <template v-else>
        <FirewallInstall 
          @installed="handleInstalled"
          @cancel="showInstall = false"
        />
      </template>
    </el-card>

    <template v-else>
      <!-- 卸载组件 -->
      <el-card v-if="showUninstall" class="uninstall-card" shadow="hover" style="margin-bottom: 20px">
        <template #header>
          <div class="card-header">
            <span>
              <el-icon style="color: #f56c6c; margin-right: 8px"><WarningFilled /></el-icon>
              卸载防火墙
            </span>
          </div>
        </template>
        <FirewallUninstall
          :firewall-name="baseInfo.name"
          :firewall-type="getFirewallType()"
          @uninstalled="handleUninstalled"
          @cancel="showUninstall = false"
        />
      </el-card>

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
                <el-button
                  type="danger"
                  :disabled="!baseInfo.isExist"
                  @click="showUninstall = true"
                >
                  卸载
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
import { WarningFilled, Loading } from '@element-plus/icons-vue';
import { loadBaseInfo as loadBaseInfoApi, operateFirewall } from '@/api/modules/firewall';
import { getOSInfo } from '@/api/modules/system';
import type { FirewallBaseInfo } from '@/api/interface/firewall';
import PortRules from '@/components/firewall/PortRules.vue';
import IPRules from '@/components/firewall/IPRules.vue';
import ForwardRules from '@/components/firewall/ForwardRules.vue';
import FirewallInstall from '@/components/firewall/FirewallInstall.vue';
import FirewallUninstall from '@/components/firewall/FirewallUninstall.vue';

const isWindows = ref(false);
const activeTab = ref('port');
const loading = ref(false);
const initialLoading = ref(true);
const showInstall = ref(false);
const showUninstall = ref(false);
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
    initialLoading.value = false;
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
  const message = enabled
    ? '禁 ping 后将无法 ping 通服务器，是否继续？'
    : '解除禁 ping 后您的服务器可能会被黑客发现，是否继续？';

  try {
    await ElMessageBox.confirm(message, '是否禁 ping', {
      type: 'warning',
    });
    loading.value = true;
    const operation = enabled ? 'enableBanPing' : 'disableBanPing';
    await operateFirewall({ operation });
    ElMessage.success('设置成功');
    await loadBaseInfo();
  } catch (error: any) {
    if (error !== 'cancel') {
      console.error('设置失败:', error);
      ElMessage.error(error.message || '设置失败');
    }
    // 取消或失败时恢复开关状态
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

// 处理安装完成
const handleInstalled = async () => {
  showInstall.value = false;
  await loadBaseInfo();
};

// 处理卸载完成
const handleUninstalled = async () => {
  showUninstall.value = false;
  await loadBaseInfo();
};

// 获取防火墙类型
const getFirewallType = (): 'ufw' | 'iptables' | 'firewalld' => {
  const name = baseInfo.value.name.toLowerCase();
  if (name === 'ufw') return 'ufw';
  if (name === 'firewalld') return 'firewalld';
  return 'iptables';
};

// 检测操作系统
const checkOS = async () => {
  try {
    const res = await getOSInfo();
    if (res && res.data) {
      isWindows.value = res.data.os === 'windows';
    }
  } catch (error) {
    console.error('检测操作系统失败:', error);
  }
};

onMounted(async () => {
  await checkOS();
  if (!isWindows.value) {
    loadBaseInfo();
  } else {
    initialLoading.value = false;
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

.install-card {
  margin-bottom: 20px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.install-card .card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.uninstall-card {
  margin-bottom: 20px;
  border-radius: 8px;
  border: 1px solid #f56c6c;
}

.install-tips {
  color: #606266;
  line-height: 1.8;
}

.install-tips code {
  background: #f5f7fa;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  color: #303133;
}

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 300px;
}

.loading-icon {
  animation: spin 1s linear infinite;
  color: #409eff;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.loading-text {
  margin-top: 16px;
  font-size: 14px;
  color: #606266;
}
</style>
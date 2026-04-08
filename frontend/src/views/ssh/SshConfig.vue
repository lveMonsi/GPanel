<template>
  <div class="ssh-config">
    <el-card class="status-card" shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="status-info">
            <el-tag type="success" size="large">SSH</el-tag>
            <el-tag :type="sshInfo.isActive ? 'success' : 'danger'" size="large" style="margin-left: 12px">
              {{ sshInfo.isActive ? '运行中' : '已停止' }}
            </el-tag>
          </div>
          <div class="status-actions">
            <el-button type="primary" :disabled="sshInfo.isActive" @click="handleOperate('start')">
              启动
            </el-button>
            <el-button type="warning" :disabled="!sshInfo.isActive" @click="handleOperate('stop')">
              停止
            </el-button>
            <el-button type="info" @click="handleOperate('restart')">
              重启
            </el-button>
          </div>
        </div>
      </template>
      <el-descriptions :column="2" border>
        <el-descriptions-item label="SSH端口">
          {{ sshInfo.port }}
        </el-descriptions-item>
        <el-descriptions-item label="监听地址">
          {{ sshInfo.listenAddress || '0.0.0.0' }}
        </el-descriptions-item>
        <el-descriptions-item label="密码认证">
          <el-tag :type="sshInfo.passwordAuth === 'yes' ? 'success' : 'danger'" size="small">
            {{ sshInfo.passwordAuth === 'yes' ? '已启用' : '已禁用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="密钥认证">
          <el-tag :type="sshInfo.pubkeyAuth === 'yes' ? 'success' : 'danger'" size="small">
            {{ sshInfo.pubkeyAuth === 'yes' ? '已启用' : '已禁用' }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-card class="config-card" shadow="hover">
      <template #header>
        <span>SSH配置</span>
      </template>
      <el-form :model="sshInfo" label-width="120px">
        <el-form-item label="Root登录">
          <el-switch
            v-model="sshInfo.permitRootLogin"
            active-value="yes"
            inactive-value="no"
            @change="handleConfigChange('permitRootLogin', sshInfo.permitRootLogin)"
          />
          <span class="form-help">是否允许root用户通过SSH登录</span>
        </el-form-item>
        <el-form-item label="密码认证">
          <el-switch
            v-model="sshInfo.passwordAuth"
            active-value="yes"
            inactive-value="no"
            @change="handleConfigChange('passwordAuth', sshInfo.passwordAuth)"
          />
          <span class="form-help">是否允许使用密码进行SSH认证</span>
        </el-form-item>
        <el-form-item label="密钥认证">
          <el-switch
            v-model="sshInfo.pubkeyAuth"
            active-value="yes"
            inactive-value="no"
            @change="handleConfigChange('pubkeyAuth', sshInfo.pubkeyAuth)"
          />
          <span class="form-help">是否允许使用公钥进行SSH认证</span>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';

const sshInfo = ref({
  isActive: false,
  port: 22,
  listenAddress: '0.0.0.0',
  passwordAuth: 'yes',
  pubkeyAuth: 'yes',
  permitRootLogin: 'yes',
});

const loadSSHInfo = async () => {
  // TODO: 调用API获取SSH信息
  console.log('加载SSH信息');
};

const handleOperate = async (operation: string) => {
  const operationText = operation === 'start' ? '启动' : operation === 'stop' ? '停止' : '重启';
  try {
    await ElMessageBox.confirm(`确认${operationText}SSH服务吗?`, '确认操作', {
      type: 'warning',
    });
    // TODO: 调用API执行操作
    ElMessage.success('操作成功');
    await loadSSHInfo();
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '操作失败');
    }
  }
};

const handleConfigChange = async (key: string, value: string) => {
  try {
    await ElMessageBox.confirm(`确认修改SSH配置吗?`, '确认操作', {
      type: 'warning',
    });
    // TODO: 调用API修改配置
    ElMessage.success('配置已更新');
    await loadSSHInfo();
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '配置更新失败');
    }
    await loadSSHInfo();
  }
};

onMounted(() => {
  loadSSHInfo();
});
</script>

<style scoped>
.ssh-config {
  padding: 20px;
  background: #fafafa;
}

.status-card,
.config-card {
  margin-bottom: 20px;
  border-radius: 8px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.status-info {
  display: flex;
  align-items: center;
}

.status-actions {
  display: flex;
  gap: 8px;
}

.form-help {
  margin-left: 12px;
  font-size: 12px;
  color: #909399;
}
</style>

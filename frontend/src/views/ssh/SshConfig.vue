<template>
  <div class="ssh-config">
    <el-card shadow="hover">
      <template #header>
        <div class="card-header">
          <div class="status-info">
            <el-tag type="success" size="large">SSH</el-tag>
            <el-tag :type="sshInfo.isActive ? 'success' : 'danger'" size="large" style="margin-left: 12px">
              {{ sshInfo.isActive ? '运行中' : '已停止' }}
            </el-tag>
          </div>
          <div class="status-actions">
            <el-button type="primary" :disabled="sshInfo.isActive" @click="handleOperate('start')">启动</el-button>
            <el-button type="warning" :disabled="!sshInfo.isActive" @click="handleOperate('stop')">停止</el-button>
            <el-button type="info" @click="handleOperate('restart')">重启</el-button>
            <el-switch
              v-model="sshInfo.autoStart"
              active-text="开机自启"
              style="margin-left: 8px"
              @change="handleAutoStart"
            />
          </div>
        </div>
      </template>

      <el-form label-width="140px">
        <el-form-item label="连接端口">
          <span>{{ sshInfo.port }}</span>
          <el-button link type="primary" style="margin-left: 12px" @click="openDialog('port')">设置</el-button>
        </el-form-item>
        <el-form-item label="监听地址">
          <span>{{ sshInfo.listenAddress || '0.0.0.0' }}</span>
          <el-button link type="primary" style="margin-left: 12px" @click="openDialog('listenAddress')">设置</el-button>
        </el-form-item>
        <el-form-item label="Root用户登录">
          <el-tag :type="rootLoginTagType" size="small">{{ rootLoginLabel }}</el-tag>
          <el-button link type="primary" style="margin-left: 12px" @click="openDialog('permitRootLogin')">设置</el-button>
        </el-form-item>
        <el-form-item label="密码认证">
          <el-switch
            v-model="sshInfo.passwordAuth"
            active-value="yes"
            inactive-value="no"
            @change="handleSwitchChange('PasswordAuthentication', sshInfo.passwordAuth)"
          />
        </el-form-item>
        <el-form-item label="密钥认证">
          <el-switch
            v-model="sshInfo.pubkeyAuth"
            active-value="yes"
            inactive-value="no"
            @change="handleSwitchChange('PubkeyAuthentication', sshInfo.pubkeyAuth)"
          />
        </el-form-item>
        <el-form-item label="反向解析">
          <el-switch
            v-model="sshInfo.useDNS"
            active-value="yes"
            inactive-value="no"
            @change="handleSwitchChange('UseDNS', sshInfo.useDNS)"
          />
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 修改对话框 -->
    <el-dialog v-model="dialog.visible" :title="dialog.title" width="400px">
      <el-form>
        <el-form-item v-if="dialog.type === 'port'" label="端口">
          <el-input-number v-model="dialog.value" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>
        <el-form-item v-else-if="dialog.type === 'listenAddress'" label="监听地址">
          <el-input v-model="dialog.value" />
        </el-form-item>
        <el-form-item v-else-if="dialog.type === 'permitRootLogin'" label="Root登录">
          <el-select v-model="dialog.value" style="width: 100%">
            <el-option label="允许" value="yes" />
            <el-option label="禁止" value="no" />
            <el-option label="仅密钥登录" value="without-password" />
            <el-option label="仅强制命令" value="forced-commands-only" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.visible = false">取消</el-button>
        <el-button type="primary" @click="submitDialog">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { ElMessage } from 'element-plus';
import { getSSHInfo, operateSSH, updateSSHConfig } from '@/api/modules/ssh';
import type { SSHInfo } from '@/api/interface/ssh';

const sshInfo = ref<SSHInfo>({
  isActive: false,
  port: 22,
  listenAddress: '0.0.0.0',
  passwordAuth: 'yes',
  pubkeyAuth: 'yes',
  permitRootLogin: 'yes',
  autoStart: false,
  useDNS: 'no',
});

const dialog = ref({
  visible: false,
  type: '',
  title: '',
  value: '' as string | number,
});

const rootLoginMap: Record<string, { label: string; type: string }> = {
  yes: { label: '允许', type: 'success' },
  no: { label: '禁止', type: 'danger' },
  'without-password': { label: '仅密钥登录', type: 'warning' },
  'forced-commands-only': { label: '仅强制命令', type: 'info' },
};

const rootLoginLabel = computed(() => rootLoginMap[sshInfo.value.permitRootLogin]?.label || sshInfo.value.permitRootLogin);
const rootLoginTagType = computed(() => rootLoginMap[sshInfo.value.permitRootLogin]?.type || 'info');

const loadSSHInfo = async () => {
  try {
    const res = await getSSHInfo();
    if (res.data) Object.assign(sshInfo.value, res.data);
  } catch (e: any) {
    ElMessage.error(e.message || '获取SSH信息失败');
  }
};

const handleOperate = async (operation: string) => {
  try {
    await operateSSH(operation);
    ElMessage.success('操作成功');
    await loadSSHInfo();
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败');
  }
};

const handleAutoStart = async (val: boolean) => {
  try {
    await operateSSH(val ? 'enable' : 'disable');
    ElMessage.success('设置成功');
  } catch (e: any) {
    ElMessage.error(e.message || '设置失败');
    await loadSSHInfo();
  }
};

const handleSwitchChange = async (key: string, value: string) => {
  try {
    await updateSSHConfig(key, value);
    ElMessage.success('配置已更新');
  } catch (e: any) {
    ElMessage.error(e.message || '配置更新失败');
    await loadSSHInfo();
  }
};

const dialogKeyMap: Record<string, { title: string; configKey: string }> = {
  port: { title: '修改端口', configKey: 'Port' },
  listenAddress: { title: '修改监听地址', configKey: 'ListenAddress' },
  permitRootLogin: { title: '修改Root登录', configKey: 'PermitRootLogin' },
};

const openDialog = (type: string) => {
  dialog.value.type = type;
  dialog.value.title = dialogKeyMap[type].title;
  dialog.value.value = type === 'port' ? sshInfo.value.port : (sshInfo.value as any)[type === 'listenAddress' ? 'listenAddress' : 'permitRootLogin'];
  dialog.value.visible = true;
};

const submitDialog = async () => {
  const { type, value } = dialog.value;
  const { configKey } = dialogKeyMap[type];
  try {
    await updateSSHConfig(configKey, String(value));
    ElMessage.success('配置已更新');
    dialog.value.visible = false;
    await loadSSHInfo();
  } catch (e: any) {
    ElMessage.error(e.message || '配置更新失败');
  }
};

onMounted(() => {
  loadSSHInfo();
});
</script>

<style scoped>
.ssh-config {
  padding: 20px;
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
  align-items: center;
  gap: 8px;
}
</style>

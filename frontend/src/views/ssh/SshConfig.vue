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

      <div class="toolbar-actions">
        <el-radio-group v-model="configMode" @change="handleModeChange">
          <el-radio-button value="base">基础配置</el-radio-button>
          <el-radio-button value="all">全部配置</el-radio-button>
        </el-radio-group>
        <div>
          <el-button @click="openKeyInfo">密钥信息</el-button>
          <el-button @click="openAuthKeys">认证密钥</el-button>
        </div>
      </div>

      <template v-if="configMode === 'base'">
        <el-form label-width="140px">
          <el-form-item label="连接端口">
            <span>{{ sshInfo.port }}</span>
            <el-button link type="primary" style="margin-left: 12px" @click="openDialog('port')">设置</el-button>
            <span class="form-help">修改 SSH 服务端口。</span>
          </el-form-item>
          <el-form-item label="监听地址">
            <span>{{ sshInfo.listenAddress || '0.0.0.0' }}</span>
            <el-button link type="primary" style="margin-left: 12px" @click="openDialog('listenAddress')">设置</el-button>
            <span class="form-help">限制 SSH 监听的网卡地址。</span>
          </el-form-item>
          <el-form-item label="Root用户登录">
            <el-tag :type="rootLoginTagType" size="small">{{ rootLoginLabel }}</el-tag>
            <el-button link type="primary" style="margin-left: 12px" @click="openDialog('permitRootLogin')">设置</el-button>
            <span class="form-help">控制 Root 用户是否允许直接登录。</span>
          </el-form-item>
          <el-form-item label="密码认证">
            <el-switch
              v-model="sshInfo.passwordAuth"
              active-value="yes"
              inactive-value="no"
              @change="handleSwitchChange('PasswordAuthentication', sshInfo.passwordAuth)"
            />
            <span class="form-help">允许用户名 + 密码方式登录 SSH。</span>
          </el-form-item>
          <el-form-item label="密钥认证">
            <el-switch
              v-model="sshInfo.pubkeyAuth"
              active-value="yes"
              inactive-value="no"
              @change="handleSwitchChange('PubkeyAuthentication', sshInfo.pubkeyAuth)"
            />
            <span class="form-help">允许使用 SSH 密钥进行登录认证。</span>
          </el-form-item>
          <el-form-item label="反向解析">
            <el-switch
              v-model="sshInfo.useDNS"
              active-value="yes"
              inactive-value="no"
              @change="handleSwitchChange('UseDNS', sshInfo.useDNS)"
            />
            <span class="form-help">控制是否启用 DNS 反向解析。</span>
          </el-form-item>
        </el-form>
      </template>

      <template v-else>
        <el-input
          v-model="sshConfigContent"
          type="textarea"
          :rows="20"
          placeholder="请输入 sshd_config 内容"
        />
        <el-button type="primary" style="margin-top: 16px" :loading="fileLoading" @click="saveFullConfig">
          保存配置
        </el-button>
      </template>
    </el-card>

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

    <SshAuthKeys ref="authKeysRef" />
    <SshKeyInfo ref="keyInfoRef" />
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { getSSHFile, getSSHInfo, operateSSH, updateSSHConfig, updateSSHFile } from '@/api/modules/ssh';
import type { SSHInfo } from '@/api/interface/ssh';
import SshAuthKeys from './components/SshAuthKeys.vue';
import SshKeyInfo from './components/SshKeyInfo.vue';

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

const configMode = ref<'base' | 'all'>('base');
const fileLoading = ref(false);
const sshConfigContent = ref('');
const authKeysRef = ref<InstanceType<typeof SshAuthKeys>>();
const keyInfoRef = ref<InstanceType<typeof SshKeyInfo>>();

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

const loadFullConfig = async () => {
  fileLoading.value = true;
  try {
    const res = await getSSHFile('sshdConf');
    sshConfigContent.value = res.data || '';
  } catch (error: any) {
    ElMessage.error(error.message || '加载配置文件失败');
  } finally {
    fileLoading.value = false;
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

const handleModeChange = async (value: 'base' | 'all') => {
  if (value === 'all') {
    await loadFullConfig();
  }
};

const saveFullConfig = async () => {
  fileLoading.value = true;
  try {
    await updateSSHFile('sshdConf', sshConfigContent.value);
    ElMessage.success('配置文件已保存');
  } catch (error: any) {
    ElMessage.error(error.message || '保存配置文件失败');
  } finally {
    fileLoading.value = false;
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
  dialog.value.value = type === 'port'
    ? sshInfo.value.port
    : (sshInfo.value as any)[type === 'listenAddress' ? 'listenAddress' : 'permitRootLogin'];
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

const openAuthKeys = () => {
  authKeysRef.value?.open();
};

const openKeyInfo = () => {
  keyInfoRef.value?.open();
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
.toolbar-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.form-help {
  margin-left: 12px;
  color: #909399;
  font-size: 12px;
}
</style>

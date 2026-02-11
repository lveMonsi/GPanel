<template>
  <div class="host-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <el-button type="primary" @click="openCreateGroupDialog">
            <el-icon><Plus /></el-icon>
            新建分组
          </el-button>
          <el-button type="primary" plain @click="openCreateHostDialog">
            <el-icon><Plus /></el-icon>
            新建主机
          </el-button>
          <el-button type="primary" plain @click="exportHosts">
            <el-icon><Download /></el-icon>
            导出
          </el-button>
          <el-button type="primary" plain @click="importHosts">
            <el-icon><Upload /></el-icon>
            导入
          </el-button>
        </div>
      </template>

      <!-- 主机树 -->
      <div class="host-tree-container">
        <el-tree
          :data="hostTree"
          :props="treeProps"
          node-key="id"
          default-expand-all
          :expand-on-click-node="false"
        >
          <template #default="{ node, data }">
            <div class="tree-node">
              <span v-if="data.type === 'group'" class="node-icon">
                <el-icon><Folder /></el-icon>
              </span>
              <span v-else class="node-icon">
                <el-icon><Monitor /></el-icon>
              </span>
              <span class="node-label">{{ node.label }}</span>
              <span class="node-actions">
                <el-button
                  v-if="data.type === 'group'"
                  link
                  type="primary"
                  size="small"
                  @click.stop="openEditGroupDialog(data)"
                >
                  编辑
                </el-button>
                <el-button
                  v-if="data.type === 'group'"
                  link
                  type="danger"
                  size="small"
                  @click.stop="deleteGroup(data)"
                >
                  删除
                </el-button>
                <el-button
                  v-if="data.type === 'host'"
                  link
                  type="primary"
                  size="small"
                  @click.stop="openEditHostDialog(data)"
                >
                  编辑
                </el-button>
                <el-button
                  v-if="data.type === 'host'"
                  link
                  type="success"
                  size="small"
                  @click.stop="testHostConnection(data)"
                >
                  测试
                </el-button>
                <el-button
                  v-if="data.type === 'host'"
                  link
                  type="danger"
                  size="small"
                  @click.stop="deleteHost(data)"
                >
                  删除
                </el-button>
              </span>
            </div>
          </template>
        </el-tree>
      </div>
    </el-card>

    <!-- 分组对话框 -->
    <el-dialog
      v-model="groupDialogVisible"
      :title="groupDialogTitle"
      width="500px"
      @close="resetGroupForm"
    >
      <el-form :model="groupForm" :rules="groupRules" ref="groupFormRef" label-width="80px">
        <el-form-item label="分组名称" prop="name">
          <el-input v-model="groupForm.name" placeholder="请输入分组名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="groupForm.description" placeholder="请输入描述（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="groupDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitGroupForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 主机对话框 -->
    <el-dialog
      v-model="hostDialogVisible"
      :title="hostDialogTitle"
      width="600px"
      @close="resetHostForm"
    >
      <el-form :model="hostForm" :rules="hostRules" ref="hostFormRef" label-width="100px">
        <el-form-item label="所属分组" prop="groupID">
          <el-select v-model="hostForm.groupID" placeholder="请选择分组" style="width: 100%">
            <el-option
              v-for="group in groups"
              :key="group.id"
              :label="group.name"
              :value="group.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="主机名称" prop="name">
          <el-input v-model="hostForm.name" placeholder="请输入主机名称" />
        </el-form-item>
        <el-form-item label="主机地址" prop="addr">
          <el-input v-model="hostForm.addr" placeholder="请输入主机地址" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input-number v-model="hostForm.port" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>
        <el-form-item label="用户名" prop="user">
          <el-input v-model="hostForm.user" placeholder="请输入用户名" />
        </el-form-item>
        <el-form-item label="认证方式" prop="authMode">
          <el-radio-group v-model="hostForm.authMode">
            <el-radio value="password">密码</el-radio>
            <el-radio value="key">密钥</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item
          v-if="hostForm.authMode === 'password'"
          label="密码"
          prop="password"
        >
          <el-input
            v-model="hostForm.password"
            type="password"
            placeholder="请输入密码"
            show-password
          />
        </el-form-item>
        <el-form-item
          v-if="hostForm.authMode === 'key'"
          label="私钥"
          prop="privateKey"
        >
          <el-input
            v-model="hostForm.privateKey"
            type="textarea"
            :rows="4"
            placeholder="请输入私钥内容"
          />
        </el-form-item>
        <el-form-item
          v-if="hostForm.authMode === 'key'"
          label="私钥密码"
        >
          <el-input
            v-model="hostForm.passPhrase"
            type="password"
            placeholder="请输入私钥密码（可选）"
            show-password
          />
        </el-form-item>
        <el-form-item label="记住密码">
          <el-switch v-model="hostForm.rememberPassword" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="hostForm.description" placeholder="请输入描述（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="hostDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitHostForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 导入对话框 -->
    <el-dialog
      v-model="importDialogVisible"
      title="导入主机"
      width="500px"
    >
      <el-upload
        drag
        :auto-upload="false"
        :on-change="handleImportFile"
        accept=".json"
      >
        <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
        <div class="el-upload__text">
          拖拽文件到此处或 <em>点击上传</em>
        </div>
        <template #tip>
          <div class="el-upload__tip">
            请上传 JSON 格式的主机配置文件
          </div>
        </template>
      </el-upload>
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import {
  Plus,
  Delete,
  Download,
  Upload,
  UploadFilled,
  Folder,
  Monitor
} from '@element-plus/icons-vue';
import {
  createGroup,
  updateGroup,
  deleteGroup,
  listGroups
} from '@/api/modules/host';
import {
  getHostTree,
  createHost,
  updateHost,
  deleteHost,
  testHostConnection as testConnection,
  exportHosts as exportHostsAPI,
  importHosts as importHostsAPI
} from '@/api/modules/host';
import type { HostTreeNode, HostGroupInfo, HostGroupOperate, HostOperate } from '@/api/interface/host';

const hostTree = ref<HostTreeNode[]>([]);
const groups = ref<HostGroupInfo[]>([]);
const groupDialogVisible = ref(false);
const hostDialogVisible = ref(false);
const importDialogVisible = ref(false);
const groupDialogTitle = ref('');
const hostDialogTitle = ref('');
const groupFormRef = ref();
const hostFormRef = ref();

const treeProps = {
  children: 'children',
  label: 'name'
};

const groupForm = reactive<HostGroupOperate>({
  name: '',
  description: ''
});

const groupRules = {
  name: [{ required: true, message: '请输入分组名称', trigger: 'blur' }]
};

const hostForm = reactive<HostOperate>({
  groupID: 0,
  name: '',
  addr: '',
  port: 22,
  user: '',
  authMode: 'password',
  password: '',
  privateKey: '',
  passPhrase: '',
  rememberPassword: false,
  description: ''
});

const hostRules = {
  groupID: [{ required: true, message: '请选择所属分组', trigger: 'change' }],
  name: [{ required: true, message: '请输入主机名称', trigger: 'blur' }],
  addr: [{ required: true, message: '请输入主机地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  user: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  authMode: [{ required: true, message: '请选择认证方式', trigger: 'change' }],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ],
  privateKey: [
    { required: true, message: '请输入私钥', trigger: 'blur' }
  ]
};

const loadHostTree = async () => {
  try {
    const res = await getHostTree();
    hostTree.value = res.data;
  } catch (error) {
    ElMessage.error('加载主机列表失败');
  }
};

const loadGroups = async () => {
  try {
    const res = await listGroups({ page: 1, pageSize: 100 });
    groups.value = res.data.data;
  } catch (error) {
    ElMessage.error('加载分组列表失败');
  }
};

const openCreateGroupDialog = () => {
  groupDialogTitle.value = '新建分组';
  groupDialogVisible.value = true;
};

const openEditGroupDialog = (data: HostTreeNode) => {
  groupDialogTitle.value = '编辑分组';
  Object.assign(groupForm, {
    id: data.id,
    name: data.name,
    description: ''
  });
  groupDialogVisible.value = true;
};

const resetGroupForm = () => {
  groupFormRef.value?.resetFields();
  Object.assign(groupForm, {
    name: '',
    description: ''
  });
};

const submitGroupForm = async () => {
  await groupFormRef.value?.validate();
  try {
    if ((groupForm as any).id) {
      await updateGroup((groupForm as any).id, groupForm);
      ElMessage.success('更新成功');
    } else {
      await createGroup(groupForm);
      ElMessage.success('创建成功');
    }
    groupDialogVisible.value = false;
    loadHostTree();
    loadGroups();
  } catch (error) {
    ElMessage.error('操作失败');
  }
};

const deleteGroup = async (data: HostTreeNode) => {
  try {
    await ElMessageBox.confirm('确定要删除该分组吗？分组下的主机也会被删除。', '提示', {
      type: 'warning'
    });
    await deleteGroup(data.id);
    ElMessage.success('删除成功');
    loadHostTree();
    loadGroups();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败');
    }
  }
};

const openCreateHostDialog = () => {
  hostDialogTitle.value = '新建主机';
  if (groups.value.length > 0) {
    hostForm.groupID = groups.value[0].id;
  }
  hostDialogVisible.value = true;
};

const openEditHostDialog = (data: HostTreeNode) => {
  hostDialogTitle.value = '编辑主机';
  Object.assign(hostForm, {
    id: data.id,
    groupID: (data as any).groupID || 0,
    name: data.name,
    addr: (data as any).addr || '',
    port: (data as any).port || 22,
    user: (data as any).user || '',
    authMode: (data as any).authMode || 'password',
    password: '',
    privateKey: '',
    passPhrase: '',
    rememberPassword: false,
    description: ''
  });
  hostDialogVisible.value = true;
};

const resetHostForm = () => {
  hostFormRef.value?.resetFields();
  Object.assign(hostForm, {
    groupID: 0,
    name: '',
    addr: '',
    port: 22,
    user: '',
    authMode: 'password',
    password: '',
    privateKey: '',
    passPhrase: '',
    rememberPassword: false,
    description: ''
  });
};

const submitHostForm = async () => {
  await hostFormRef.value?.validate();
  try {
    if ((hostForm as any).id) {
      await updateHost((hostForm as any).id, hostForm);
      ElMessage.success('更新成功');
    } else {
      await createHost(hostForm);
      ElMessage.success('创建成功');
    }
    hostDialogVisible.value = false;
    loadHostTree();
  } catch (error) {
    ElMessage.error('操作失败');
  }
};

const testHostConnection = async (data: HostTreeNode) => {
  try {
    const hostData = data as any;
    const testData = {
      addr: hostData.addr,
      port: hostData.port,
      user: hostData.user,
      authMode: hostData.authMode,
      password: hostData.password || '',
      privateKey: hostData.privateKey || '',
      passPhrase: hostData.passPhrase || ''
    };
    const res = await testConnection(testData);
    if (res.data.success) {
      ElMessage.success('连接测试成功');
    } else {
      ElMessage.error(`连接测试失败: ${res.data.message}`);
    }
  } catch (error) {
    ElMessage.error('连接测试失败');
  }
};

const deleteHost = async (data: HostTreeNode) => {
  try {
    await ElMessageBox.confirm('确定要删除该主机吗？', '提示', {
      type: 'warning'
    });
    await deleteHost(data.id);
    ElMessage.success('删除成功');
    loadHostTree();
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败');
    }
  }
};

const exportHosts = async () => {
  try {
    const res = await exportHostsAPI();
    const dataStr = JSON.stringify(res.data, null, 2);
    const blob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'hosts.json';
    a.click();
    URL.revokeObjectURL(url);
    ElMessage.success('导出成功');
  } catch (error) {
    ElMessage.error('导出失败');
  }
};

const importHosts = () => {
  importDialogVisible.value = true;
};

const handleImportFile = async (file: any) => {
  const reader = new FileReader();
  reader.onload = async (e) => {
    try {
      const data = JSON.parse(e.target?.result as string);
      const res = await importHostsAPI(data);
      ElMessage.success(`导入成功: ${res.data.success} 个, 失败: ${res.data.fail} 个`);
      if (res.data.message) {
        ElMessage.warning(res.data.message);
      }
      importDialogVisible.value = false;
      loadHostTree();
    } catch (error) {
      ElMessage.error('导入失败，请检查文件格式');
    }
  };
  reader.readAsText(file.raw);
};

const acceptParams = () => {
  loadHostTree();
  loadGroups();
};

defineExpose({
  acceptParams
});

onMounted(() => {
  loadHostTree();
  loadGroups();
});
</script>

<style scoped>
.host-page {
  padding: 20px;
}

.card-header {
  display: flex;
  gap: 10px;
}

.host-tree-container {
  max-height: calc(100vh - 250px);
  overflow-y: auto;
}

.tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.node-icon {
  color: #409eff;
}

.node-label {
  flex: 1;
}

.node-actions {
  display: flex;
  gap: 4px;
  opacity: 0;
  transition: opacity 0.2s;
}

.el-tree-node__content:hover .node-actions {
  opacity: 1;
}
</style>
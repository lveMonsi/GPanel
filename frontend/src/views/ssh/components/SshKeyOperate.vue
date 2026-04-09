<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑密钥信息' : '新增密钥信息'"
    width="720px"
    :close-on-click-modal="false"
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
      <el-form-item label="名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入密钥名称" />
      </el-form-item>
      <el-form-item label="加密方式" prop="encryptionMode">
        <el-select v-model="form.encryptionMode" style="width: 100%">
          <el-option label="ED25519" value="ed25519" />
          <el-option label="ECDSA" value="ecdsa" />
          <el-option label="RSA" value="rsa" />
          <el-option label="DSA" value="dsa" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="!isEdit" label="创建方式" prop="mode">
        <el-radio-group v-model="form.mode">
          <el-radio value="generate">生成</el-radio>
          <el-radio value="input">输入</el-radio>
          <el-radio value="import">导入</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="私钥密码" prop="passPhrase">
        <el-input v-model="form.passPhrase" type="password" show-password placeholder="可选" />
      </el-form-item>
      <template v-if="form.mode === 'input' || isEdit">
        <el-form-item label="私钥" prop="privateKey">
          <el-input v-model="form.privateKey" type="textarea" :rows="6" placeholder="请输入私钥内容" />
        </el-form-item>
        <el-form-item label="公钥" prop="publicKey">
          <el-input v-model="form.publicKey" type="textarea" :rows="4" placeholder="请输入公钥内容" />
        </el-form-item>
      </template>
      <template v-else-if="form.mode === 'import'">
        <el-form-item label="私钥文件" prop="privateKey">
          <input type="file" @change="handlePrivateKeyFileChange" />
        </el-form-item>
        <el-form-item label="公钥文件" prop="publicKey">
          <input type="file" @change="handlePublicKeyFileChange" />
        </el-form-item>
      </template>
      <el-form-item label="描述" prop="description">
        <el-input v-model="form.description" type="textarea" :rows="2" placeholder="请输入描述" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { ElMessage } from 'element-plus';
import type { FormInstance, FormRules } from 'element-plus';
import { createSSHKey, updateSSHKey } from '@/api/modules/ssh';
import type { SSHKeyInfo, SSHKeyOperate } from '@/api/interface/ssh';

const emit = defineEmits<{ saved: [] }>();

const visible = ref(false);
const loading = ref(false);
const formRef = ref<FormInstance>();
const form = ref<SSHKeyOperate>({
  name: '',
  mode: 'generate',
  encryptionMode: 'ed25519',
  passPhrase: '',
  description: '',
  publicKey: '',
  privateKey: '',
});

const isEdit = computed(() => !!form.value.id);

const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  encryptionMode: [{ required: true, message: '请选择加密方式', trigger: 'change' }],
  privateKey: [{
    validator: (_rule, value, callback) => {
      if ((form.value.mode === 'input' || isEdit.value) && !value) callback(new Error('请输入私钥'));
      else callback();
    },
    trigger: 'blur',
  }],
  publicKey: [{
    validator: (_rule, value, callback) => {
      if ((form.value.mode === 'input' || isEdit.value) && !value) callback(new Error('请输入公钥'));
      else callback();
    },
    trigger: 'blur',
  }],
};

const open = (row?: SSHKeyInfo) => {
  if (row) {
    form.value = {
      id: row.id,
      name: row.name,
      mode: row.mode,
      encryptionMode: row.encryptionMode,
      passPhrase: row.passPhrase || '',
      description: row.description || '',
      publicKey: row.publicKey || '',
      privateKey: row.privateKey || '',
    };
  } else {
    form.value = {
      name: '',
      mode: 'generate',
      encryptionMode: 'ed25519',
      passPhrase: '',
      description: '',
      publicKey: '',
      privateKey: '',
    };
  }
  visible.value = true;
};

const handleFileChange = (event: Event, field: 'privateKey' | 'publicKey') => {
  const target = event.target as HTMLInputElement;
  const file = target.files?.[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = () => {
    form.value[field] = String(reader.result || '');
  };
  reader.readAsText(file);
};

const handlePrivateKeyFileChange = (event: Event) => {
  handleFileChange(event, 'privateKey');
};

const handlePublicKeyFileChange = (event: Event) => {
  handleFileChange(event, 'publicKey');
};

const handleSubmit = async () => {
  try {
    await formRef.value?.validate();
    loading.value = true;
    if (isEdit.value) {
      await updateSSHKey(form.value);
      ElMessage.success('密钥信息已更新');
    } else {
      await createSSHKey(form.value);
      ElMessage.success('密钥信息已创建');
    }
    visible.value = false;
    emit('saved');
  } catch (error: any) {
    if (error?.message) {
      ElMessage.error(error.message);
    }
  } finally {
    loading.value = false;
  }
};

defineExpose({ open });
</script>

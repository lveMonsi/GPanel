<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑主机' : '添加主机'"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="100px"
      label-position="right"
    >
      <el-form-item label="主机名称" prop="name">
        <el-input v-model="formData.name" placeholder="请输入主机名称" />
      </el-form-item>

      <el-form-item label="所属分组" prop="groupID">
        <el-select
          v-model="formData.groupID"
          placeholder="请选择分组"
          style="width: 100%"
        >
          <el-option
            v-for="group in groups"
            :key="group.id"
            :label="group.name"
            :value="group.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="主机地址" prop="addr">
        <el-input v-model="formData.addr" placeholder="请输入主机地址（IP或域名）" />
      </el-form-item>

      <el-form-item label="端口" prop="port">
        <el-input-number
          v-model="formData.port"
          :min="1"
          :max="65535"
          placeholder="请输入端口"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item label="用户名" prop="user">
        <el-input v-model="formData.user" placeholder="请输入用户名" />
      </el-form-item>

      <el-form-item label="认证方式" prop="authMode">
        <el-radio-group v-model="formData.authMode">
          <el-radio value="password">密码认证</el-radio>
          <el-radio value="key">密钥认证</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item v-if="formData.authMode === 'password'" label="密码" prop="password">
        <el-input
          v-model="formData.password"
          type="password"
          show-password
          placeholder="请输入密码"
        />
      </el-form-item>

      <el-form-item v-if="formData.authMode === 'key'" label="私钥" prop="privateKey">
        <el-input
          v-model="formData.privateKey"
          type="textarea"
          :rows="6"
          placeholder="请输入私钥内容"
        />
      </el-form-item>

      <el-form-item v-if="formData.authMode === 'key'" label="私钥密码" prop="passPhrase">
        <el-input
          v-model="formData.passPhrase"
          type="password"
          show-password
          placeholder="请输入私钥密码（如有）"
        />
      </el-form-item>

      <el-form-item label="记住密码" prop="rememberPassword">
        <el-switch v-model="formData.rememberPassword" />
      </el-form-item>

      <el-form-item label="描述" prop="description">
        <el-input
          v-model="formData.description"
          type="textarea"
          :rows="2"
          placeholder="请输入描述信息"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="handleTest" :loading="testing">测试连接</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">保存</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import type { HostOperate, HostGroupInfo } from '@/api/interface/host'
import { testHostConnection, getHostForTerminal } from '@/api/modules/host'

interface Props {
  visible: boolean
  host?: HostOperate
  groups: HostGroupInfo[]
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  host: undefined,
  groups: () => []
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'save': [data: HostOperate]
}>()

const dialogVisible = ref(props.visible)
const formRef = ref<FormInstance>()
const testing = ref(false)
const saving = ref(false)

const isEdit = ref(!!props.host?.id)

const formData = ref<HostOperate>({
  groupID: props.host?.groupID || (props.groups[0]?.id || 0),
  name: props.host?.name || '',
  addr: props.host?.addr || '',
  port: props.host?.port || 22,
  user: props.host?.user || '',
  authMode: props.host?.authMode || 'password',
  password: '',
  privateKey: '',
  passPhrase: '',
  rememberPassword: props.host?.rememberPassword || false,
  description: props.host?.description || ''
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入主机名称', trigger: 'blur' }
  ],
  groupID: [
    { required: true, message: '请选择分组', trigger: 'change' }
  ],
  addr: [
    { required: true, message: '请输入主机地址', trigger: 'blur' }
  ],
  port: [
    { required: true, message: '请输入端口', trigger: 'blur' }
  ],
  user: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  authMode: [
    { required: true, message: '请选择认证方式', trigger: 'change' }
  ],
  password: [
    {
      validator: (rule, value, callback) => {
        if (formData.value.authMode === 'password' && !value) {
          callback(new Error('请输入密码'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  privateKey: [
    {
      validator: (rule, value, callback) => {
        if (formData.value.authMode === 'key' && !value) {
          callback(new Error('请输入私钥'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

watch(() => props.visible, async (val) => {
  dialogVisible.value = val
  if (val) {
    isEdit.value = !!props.host?.id
    if (props.host?.id) {
      // 编辑模式：从后端获取完整的主机信息（包含解密后的密码/密钥）
      try {
        const res = await getHostForTerminal(props.host.id)
        const hostData = res.data
        formData.value = {
          id: hostData.id,
          groupID: hostData.groupID,
          name: hostData.name,
          addr: hostData.addr,
          port: hostData.port,
          user: hostData.user,
          authMode: hostData.authMode,
          password: hostData.password || '',
          privateKey: hostData.privateKey || '',
          passPhrase: hostData.passPhrase || '',
          rememberPassword: hostData.rememberPassword,
          description: hostData.description || ''
        }
      } catch (error) {
        console.error('获取主机信息失败:', error)
        ElMessage.error('获取主机信息失败')
        // 如果获取失败，使用传入的数据
        formData.value = {
          ...props.host,
          password: '',
          privateKey: '',
          passPhrase: ''
        }
      }
    } else if (props.host) {
      // 新增模式但有传入数据
      formData.value = {
        ...props.host,
        password: '',
        privateKey: '',
        passPhrase: ''
      }
    } else {
      // 全新新增
      formData.value = {
        groupID: props.groups[0]?.id || 0,
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
      }
    }
  }
})

watch(dialogVisible, (val) => {
  emit('update:visible', val)
})

const handleTest = async () => {
  try {
    await formRef.value?.validate()
    testing.value = true
    const res = await testHostConnection({
      addr: formData.value.addr,
      port: formData.value.port,
      user: formData.value.user,
      authMode: formData.value.authMode,
      password: formData.value.password,
      privateKey: formData.value.privateKey,
      passPhrase: formData.value.passPhrase
    })
    if (res.data.data?.success) {
      ElMessage.success('连接测试成功')
    } else {
      ElMessage.error(res.data.message)
    }
  } catch (error) {
    console.error('测试连接失败:', error)
  } finally {
    testing.value = false
  }
}

const handleSave = async () => {
  try {
    await formRef.value?.validate()
    saving.value = true
    emit('save', {
      ...formData.value,
      id: props.host?.id
    })
    dialogVisible.value = false
  } catch (error) {
    console.error('保存失败:', error)
  } finally {
    saving.value = false
  }
}

const handleClose = () => {
  formRef.value?.resetFields()
  dialogVisible.value = false
}
</script>

<style scoped>
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}
</style>

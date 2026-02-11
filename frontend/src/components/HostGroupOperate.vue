<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑分组' : '添加分组'"
    width="500px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="80px"
      label-position="right"
    >
      <el-form-item label="分组名称" prop="name">
        <el-input v-model="formData.name" placeholder="请输入分组名称" />
      </el-form-item>

      <el-form-item label="描述" prop="description">
        <el-input
          v-model="formData.description"
          type="textarea"
          :rows="3"
          placeholder="请输入描述信息"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import type { HostGroupOperate } from '@/api/interface/host'

interface Props {
  visible: boolean
  group?: HostGroupOperate
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  group: undefined
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'save': [data: HostGroupOperate]
}>()

const dialogVisible = ref(props.visible)
const formRef = ref<FormInstance>()

const isEdit = ref(!!props.group?.id)

const formData = ref<HostGroupOperate>({
  id: props.group?.id,
  name: props.group?.name || '',
  description: props.group?.description || ''
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入分组名称', trigger: 'blur' }
  ]
}

watch(() => props.visible, (val) => {
  dialogVisible.value = val
  if (val) {
    isEdit.value = !!props.group?.id
    if (props.group) {
      formData.value = { ...props.group }
    } else {
      formData.value = {
        name: '',
        description: ''
      }
    }
  }
})

watch(dialogVisible, (val) => {
  emit('update:visible', val)
})

const handleSave = async () => {
  try {
    await formRef.value?.validate()
    emit('save', formData.value)
    dialogVisible.value = false
  } catch (error) {
    console.error('保存失败:', error)
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
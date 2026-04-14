<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑计划任务' : '新建计划任务'"
    width="700px"
    @close="resetForm"
    destroy-on-close
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="任务名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入任务名称" />
      </el-form-item>

      <el-form-item label="任务类型" prop="type">
        <el-select v-model="form.type" :disabled="isEdit" placeholder="请选择任务类型" style="width: 100%">
          <el-option
            v-for="t in taskTypes"
            :key="t.value"
            :label="t.label"
            :value="t.value"
          />
        </el-select>
      </el-form-item>

      <!-- 调度配置 -->
      <CronScheduler
        v-model="form.spec"
        v-model:spec-custom="form.specCustom"
      />

      <!-- Shell 脚本 -->
      <el-form-item v-if="form.type === 'shell'" label="脚本内容" prop="script">
        <el-input
          v-model="form.script"
          type="textarea"
          :rows="10"
          placeholder="#!/bin/bash&#10;echo 'Hello World'"
          style="font-family: monospace"
        />
      </el-form-item>

      <!-- Curl URL -->
      <el-form-item v-if="form.type === 'curl'" label="请求地址" prop="url">
        <el-input
          v-model="form.url"
          type="textarea"
          :rows="3"
          placeholder="每行一个 URL，例如：&#10;https://example.com/api/ping&#10;https://example.com/api/health"
        />
      </el-form-item>

      <!-- 目录备份 -->
      <template v-if="form.type === 'directory'">
        <el-form-item label="源目录" prop="sourceDir">
          <el-input v-model="form.sourceDir" placeholder="例如: /home/data" />
        </el-form-item>
        <el-form-item label="排除规则">
          <el-input
            v-model="form.exclusionRules"
            type="textarea"
            :rows="3"
            placeholder="每行一个排除规则，例如：&#10;*.tmp&#10;node_modules"
          />
        </el-form-item>
        <el-form-item label="保留份数">
          <el-input-number v-model="form.retainCopies" :min="1" :max="999" />
        </el-form-item>
      </template>

      <!-- 磁盘清理 -->
      <el-form-item v-if="form.type === 'clean'">
        <el-alert type="info" :closable="false" show-icon>
          清理 /tmp 和 /var/tmp 下超过 24 小时的临时文件
        </el-alert>
      </el-form-item>

      <!-- 日志清理 -->
      <el-form-item v-if="form.type === 'cleanLog'">
        <el-alert type="info" :closable="false" show-icon>
          清理 /var/log 下超过 7 天的日志文件（.log 文件截断，.gz/.log.1 文件删除）
        </el-alert>
      </el-form-item>

      <!-- 高级选项 -->
      <el-divider content-position="left">高级选项</el-divider>

      <el-form-item label="重试次数">
        <el-input-number v-model="form.retryCount" :min="0" :max="10" />
        <span class="form-hint">失败后重试的次数，0 表示不重试</span>
      </el-form-item>

      <el-form-item label="超时时间">
        <el-input-number v-model="form.timeout" :min="0" :max="86400" />
        <span class="form-hint">秒，0 表示无限制</span>
      </el-form-item>

      <el-form-item label="忽略错误">
        <el-switch v-model="form.ignoreErr" />
        <span class="form-hint">开启后即使执行失败也记为成功</span>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" @click="submitForm" :loading="submitting">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import CronScheduler from './CronScheduler.vue'
import { createCronjob, updateCronjob } from '@/api/modules/cronjob'
import type { Cronjob, CronjobType } from '@/api/interface/cronjob'

const emit = defineEmits<{
  success: []
}>()

const visible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref()

const taskTypes: { value: CronjobType; label: string }[] = [
  { value: 'shell', label: 'Shell 脚本' },
  { value: 'curl', label: 'Curl 请求' },
  { value: 'directory', label: '目录备份' },
  { value: 'clean', label: '磁盘清理' },
  { value: 'cleanLog', label: '日志清理' }
]

const defaultForm = () => ({
  id: 0,
  name: '',
  type: 'shell' as CronjobType,
  spec: '30 1 * * 1',
  specCustom: false,
  script: '',
  url: '',
  sourceDir: '',
  exclusionRules: '',
  retainCopies: 5,
  retryCount: 0,
  timeout: 0,
  ignoreErr: false
})

const form = reactive(defaultForm())

const rules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择任务类型', trigger: 'change' }],
  script: [{ required: true, message: '请输入脚本内容', trigger: 'blur' }],
  url: [{ required: true, message: '请输入请求地址', trigger: 'blur' }],
  sourceDir: [{ required: true, message: '请输入源目录', trigger: 'blur' }]
}

const open = (row?: Cronjob) => {
  if (row) {
    isEdit.value = true
    Object.assign(form, {
      id: row.id,
      name: row.name,
      type: row.type,
      spec: row.spec,
      specCustom: row.specCustom,
      script: row.script,
      url: row.url,
      sourceDir: row.sourceDir,
      exclusionRules: row.exclusionRules,
      retainCopies: row.retainCopies,
      retryCount: row.retryCount,
      timeout: row.timeout,
      ignoreErr: row.ignoreErr
    })
  } else {
    isEdit.value = false
    Object.assign(form, defaultForm())
  }
  visible.value = true
}

const resetForm = () => {
  formRef.value?.resetFields()
  Object.assign(form, defaultForm())
}

const submitForm = async () => {
  await formRef.value?.validate()
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateCronjob({
        id: form.id,
        name: form.name,
        spec: form.spec,
        specCustom: form.specCustom,
        script: form.script,
        url: form.url,
        sourceDir: form.sourceDir,
        exclusionRules: form.exclusionRules,
        retainCopies: form.retainCopies,
        retryCount: form.retryCount,
        timeout: form.timeout,
        ignoreErr: form.ignoreErr
      })
      ElMessage.success('更新成功')
    } else {
      await createCronjob({
        name: form.name,
        type: form.type,
        spec: form.spec,
        specCustom: form.specCustom,
        script: form.script,
        url: form.url,
        sourceDir: form.sourceDir,
        exclusionRules: form.exclusionRules,
        retainCopies: form.retainCopies,
        retryCount: form.retryCount,
        timeout: form.timeout,
        ignoreErr: form.ignoreErr
      })
      ElMessage.success('创建成功')
    }
    visible.value = false
    emit('success')
  } catch {
    ElMessage.error('操作失败')
  } finally {
    submitting.value = false
  }
}

defineExpose({ open })
</script>

<style scoped>
.form-hint {
  margin-left: 12px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>

<template>
  <div v-if="visible" class="edit-modal-overlay" @click.self="handleCancel">
    <div class="edit-modal">
      <div class="edit-modal-header">
        <h3>{{ title }}</h3>
        <button class="btn-close" @click="handleCancel">
          <el-icon><Close /></el-icon>
        </button>
      </div>
      <div class="edit-modal-body">
        <div class="form-group">
          <label :for="fieldId">{{ title }}</label>
          <div v-if="fieldType === 'security'" class="input-with-button">
            <input
              :id="fieldId"
              v-model="inputValue"
              type="text"
              :placeholder="placeholder"
              ref="inputRef"
              @keyup.enter="handleSave"
            />
            <button type="button" class="btn-generate" @click="generateRandom">
              随机生成
            </button>
          </div>
          <input
            v-else
            :id="fieldId"
            v-model="inputValue"
            :type="fieldType"
            :placeholder="placeholder"
            ref="inputRef"
            @keyup.enter="handleSave"
          />
          <small v-if="fieldType === 'security'" class="hint">安全入口设置为空时，则取消安全入口</small>
        </div>
      </div>
      <div class="edit-modal-footer">
        <button class="btn btn-secondary" @click="handleCancel">取消</button>
        <button class="btn btn-reset" @click="handleReset">重置</button>
        <button class="btn btn-primary" @click="handleSave">保存</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { Close } from '@element-plus/icons-vue'

interface Props {
  visible: boolean
  title: string
  fieldName: string
  fieldValue: string
  fieldType: 'text' | 'password' | 'number' | 'security'
}

const props = withDefaults(defineProps<Props>(), {
  visible: false,
  title: '',
  fieldName: '',
  fieldValue: '',
  fieldType: 'text'
})

const emit = defineEmits<{
  'update:visible': [value: boolean]
  'save': [value: string]
}>()

const inputValue = ref(props.fieldValue)
const inputRef = ref<HTMLInputElement | null>(null)

const fieldId = ref(`edit-field-${Math.random().toString(36).substr(2, 9)}`)

const placeholder = ref('')
switch (props.fieldType) {
  case 'text':
  case 'security':
    placeholder.value = '请输入文本'
    break
  case 'password':
    placeholder.value = '请输入密码'
    break
  case 'number':
    placeholder.value = '请输入数字'
    break
}

const generateRandom = () => {
  const chars = 'abcdefghijklmnopqrstuvwxyz0123456789'
  let result = '/'
  for (let i = 0; i < 12; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  inputValue.value = result
}

const handleCancel = () => {
  inputValue.value = props.fieldValue
  emit('update:visible', false)
}

const handleSave = () => {
  emit('save', inputValue.value)
  inputValue.value = ''
}

const handleReset = () => {
  inputValue.value = props.fieldValue
}

// 监听 visible 变化，自动聚焦输入框
watch(() => props.visible, (newVal) => {
  if (newVal) {
    inputValue.value = props.fieldValue
    nextTick(() => {
      inputRef.value?.focus()
    })
  }
})
</script>

<style scoped>
.edit-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

.edit-modal {
  background: var(--card-bg);
  border-radius: var(--radius-md);
  width: 90%;
  max-width: 400px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    transform: translateY(20px);
    opacity: 0;
  }
  to {
    transform: translateY(0);
    opacity: 1;
  }
}

.edit-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--border-color);
}

.edit-modal-header h3 {
  margin: 0;
  font-size: 1rem;
  color: var(--text-primary);
  font-weight: 600;
}

.btn-close {
  padding: 0.3rem;
  background: transparent;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

.btn-close:hover {
  background: var(--bg-color);
  color: var(--text-primary);
}

.edit-modal-body {
  padding: 1.25rem;
}

.form-group {
  margin-bottom: 0;
}

.form-group label {
  display: block;
  margin-bottom: 0.4rem;
  font-weight: 500;
  color: var(--text-primary);
  font-size: 0.8rem;
}

.form-group input {
  width: 100%;
  padding: 0.5rem 0.6rem;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  font-size: 0.85rem;
  transition: all 0.2s;
  background: var(--bg-color);
}

.form-group input:focus {
  outline: none;
  border-color: var(--primary);
  background: var(--card-bg);
  box-shadow: 0 0 0 2px var(--primary-light);
}

.form-group .hint {
  display: block;
  margin-top: 0.3rem;
  font-size: 0.7rem;
  color: var(--text-secondary);
}

.edit-modal-footer {
  display: flex;
  gap: 0.5rem;
  padding: 1rem 1.25rem;
  border-top: 1px solid var(--border-color);
  justify-content: flex-end;
}

.btn {
  padding: 0.5rem 1rem;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-primary {
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
  color: white;
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 6px rgba(122, 201, 224, 0.3);
}

.btn-secondary {
  background: var(--bg-color);
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
}

.btn-secondary:hover {
  background: var(--border-color);
  color: var(--text-primary);
}

.btn-reset {
  background: #fff0f0;
  color: #e74c3c;
  border: 1px solid #fadbd8;
}

.btn-reset:hover {
  background: #fadbd8;
  border-color: #e74c3c;
}

.input-with-button {
  display: flex;
  gap: 0.4rem;
}

.input-with-button input {
  flex: 1;
}

.btn-generate {
  padding: 0.5rem 0.6rem;
  background: var(--bg-color);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--primary-dark);
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.75rem;
  white-space: nowrap;
}

.btn-generate:hover {
  background: var(--primary-light);
  border-color: var(--primary);
}

.btn-generate:active {
  transform: scale(0.95);
}
</style>
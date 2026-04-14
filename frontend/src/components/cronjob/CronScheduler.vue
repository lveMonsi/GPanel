<template>
  <div class="cron-scheduler">
    <el-form-item label="自定义表达式">
      <el-switch v-model="isCustom" @change="onCustomToggle" />
    </el-form-item>

    <template v-if="!isCustom">
      <el-form-item label="执行周期">
        <el-select v-model="specObj.specType" @change="onPresetChange" style="width: 160px">
          <el-option label="每月" value="perMonth" />
          <el-option label="每周" value="perWeek" />
          <el-option label="每天" value="perDay" />
          <el-option label="每小时" value="perHour" />
          <el-option label="每N天" value="perNDay" />
          <el-option label="每N小时" value="perNHour" />
          <el-option label="每N分钟" value="perNMinute" />
        </el-select>
      </el-form-item>

      <!-- 每月：日期 + 时间 -->
      <el-form-item v-if="specObj.specType === 'perMonth'" label="日期与时间">
        <el-input-number v-model="specObj.day" :min="1" :max="28" style="width: 100px" />
        <span class="unit">日</span>
        <el-input-number v-model="specObj.hour" :min="0" :max="23" style="width: 100px" />
        <span class="unit">时</span>
        <el-input-number v-model="specObj.minute" :min="0" :max="59" style="width: 100px" />
        <span class="unit">分</span>
      </el-form-item>

      <!-- 每周：星期几 + 时间 -->
      <el-form-item v-if="specObj.specType === 'perWeek'" label="星期与时间">
        <el-select v-model="specObj.week" style="width: 120px">
          <el-option label="周一" :value="1" />
          <el-option label="周二" :value="2" />
          <el-option label="周三" :value="3" />
          <el-option label="周四" :value="4" />
          <el-option label="周五" :value="5" />
          <el-option label="周六" :value="6" />
          <el-option label="周日" :value="0" />
        </el-select>
        <el-input-number v-model="specObj.hour" :min="0" :max="23" style="width: 100px" />
        <span class="unit">时</span>
        <el-input-number v-model="specObj.minute" :min="0" :max="59" style="width: 100px" />
        <span class="unit">分</span>
      </el-form-item>

      <!-- 每天：时间 -->
      <el-form-item v-if="specObj.specType === 'perDay'" label="执行时间">
        <el-input-number v-model="specObj.hour" :min="0" :max="23" style="width: 100px" />
        <span class="unit">时</span>
        <el-input-number v-model="specObj.minute" :min="0" :max="59" style="width: 100px" />
        <span class="unit">分</span>
      </el-form-item>

      <!-- 每小时：分钟 -->
      <el-form-item v-if="specObj.specType === 'perHour'" label="执行分钟">
        <el-input-number v-model="specObj.minute" :min="0" :max="59" style="width: 100px" />
        <span class="unit">分</span>
      </el-form-item>

      <!-- 每N天：间隔 + 时间 -->
      <el-form-item v-if="specObj.specType === 'perNDay'" label="间隔与时间">
        <span class="unit">每</span>
        <el-input-number v-model="specObj.day" :min="1" :max="30" style="width: 100px" />
        <span class="unit">天</span>
        <el-input-number v-model="specObj.hour" :min="0" :max="23" style="width: 100px" />
        <span class="unit">时</span>
        <el-input-number v-model="specObj.minute" :min="0" :max="59" style="width: 100px" />
        <span class="unit">分</span>
      </el-form-item>

      <!-- 每N小时：间隔 + 分钟 -->
      <el-form-item v-if="specObj.specType === 'perNHour'" label="间隔与分钟">
        <span class="unit">每</span>
        <el-input-number v-model="specObj.hour" :min="1" :max="23" style="width: 100px" />
        <span class="unit">小时</span>
        <el-input-number v-model="specObj.minute" :min="0" :max="59" style="width: 100px" />
        <span class="unit">分</span>
      </el-form-item>

      <!-- 每N分钟：间隔 -->
      <el-form-item v-if="specObj.specType === 'perNMinute'" label="间隔">
        <span class="unit">每</span>
        <el-input-number v-model="specObj.minute" :min="1" :max="59" style="width: 100px" />
        <span class="unit">分钟</span>
      </el-form-item>

      <el-form-item>
        <span class="cron-preview">{{ previewText }}</span>
      </el-form-item>
    </template>

    <template v-else>
      <el-form-item label="Cron 表达式">
        <el-input
          v-model="customSpec"
          placeholder="分 时 日 月 周 (例如: 30 2 * * *)"
          @input="onCustomInput"
        />
        <div class="cron-hint">5段标准格式：分钟(0-59) 小时(0-23) 日(1-31) 月(1-12) 星期(0-6)</div>
      </el-form-item>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch, computed } from 'vue'
import type { SpecObj } from '@/api/interface/cronjob'

const props = defineProps<{
  modelValue: string
  specCustom: boolean
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'update:specCustom': [value: boolean]
}>()

const isCustom = ref(props.specCustom)
const customSpec = ref(props.modelValue)

const specObj = reactive<SpecObj>({
  specType: 'perWeek',
  week: 1,
  day: 1,
  hour: 1,
  minute: 30
})

// 从 cron 表达式解析到预设对象
function parseSpec(spec: string) {
  if (!spec) return

  const parts = spec.split(' ')
  if (parts.length !== 5) return

  specObj.minute = parseInt(parts[0]) || 0

  if (parts[1] === '*') {
    specObj.specType = 'perHour'
    return
  }
  if (parts[1].startsWith('*/')) {
    specObj.specType = 'perNHour'
    specObj.hour = parseInt(parts[1].replace('*/', '')) || 1
    return
  }
  specObj.hour = parseInt(parts[1]) || 0

  if (parts[2].startsWith('*/')) {
    specObj.specType = 'perNDay'
    specObj.day = parseInt(parts[2].replace('*/', '')) || 1
    return
  }
  if (parts[2] !== '*') {
    specObj.specType = 'perMonth'
    specObj.day = parseInt(parts[2]) || 1
    return
  }
  if (parts[4] !== '*') {
    specObj.specType = 'perWeek'
    specObj.week = parseInt(parts[4]) || 0
    return
  }
  specObj.specType = 'perDay'
}

// 从预设对象生成 cron 表达式
function buildSpec(): string {
  switch (specObj.specType) {
    case 'perMonth':
      return `${specObj.minute} ${specObj.hour} ${specObj.day} * *`
    case 'perWeek':
      return `${specObj.minute} ${specObj.hour} * * ${specObj.week}`
    case 'perDay':
      return `${specObj.minute} ${specObj.hour} * * *`
    case 'perHour':
      return `${specObj.minute} * * * *`
    case 'perNDay':
      return `${specObj.minute} ${specObj.hour} */${specObj.day} * *`
    case 'perNHour':
      return `${specObj.minute} */${specObj.hour} * * *`
    case 'perNMinute':
      return `*/${specObj.minute} * * * *`
    default:
      return '30 1 * * 1'
  }
}

const weekLabels: Record<number, string> = {
  0: '周日', 1: '周一', 2: '周二', 3: '周三', 4: '周四', 5: '周五', 6: '周六'
}

const previewText = computed(() => {
  const pad = (n: number) => String(n).padStart(2, '0')
  switch (specObj.specType) {
    case 'perMonth':
      return `每月 ${specObj.day} 日 ${pad(specObj.hour)}:${pad(specObj.minute)} 执行`
    case 'perWeek':
      return `每${weekLabels[specObj.week] || '周一'} ${pad(specObj.hour)}:${pad(specObj.minute)} 执行`
    case 'perDay':
      return `每天 ${pad(specObj.hour)}:${pad(specObj.minute)} 执行`
    case 'perHour':
      return `每小时 第${pad(specObj.minute)}分 执行`
    case 'perNDay':
      return `每 ${specObj.day} 天 ${pad(specObj.hour)}:${pad(specObj.minute)} 执行`
    case 'perNHour':
      return `每 ${specObj.hour} 小时 第${pad(specObj.minute)}分 执行`
    case 'perNMinute':
      return `每 ${specObj.minute} 分钟 执行`
    default:
      return ''
  }
})

// 初始化
if (!props.specCustom && props.modelValue) {
  parseSpec(props.modelValue)
}

const onCustomToggle = (val: boolean | string | number) => {
  isCustom.value = val as boolean
  emit('update:specCustom', val as boolean)
  if (!val) {
    const spec = buildSpec()
    emit('update:modelValue', spec)
  } else {
    customSpec.value = props.modelValue
  }
}

const onCustomInput = () => {
  emit('update:modelValue', customSpec.value)
}

const onPresetChange = () => {
  // 重置子字段默认值
  specObj.week = 1
  specObj.day = 1
  specObj.hour = 1
  specObj.minute = 30
}

// 监听预设对象变化
watch(specObj, () => {
  if (!isCustom.value) {
    emit('update:modelValue', buildSpec())
  }
}, { deep: true })
</script>

<style scoped>
.cron-scheduler {
  width: 100%;
}

.unit {
  margin: 0 8px;
  color: var(--el-text-color-secondary);
}

.cron-preview {
  color: var(--el-color-primary);
  font-size: 13px;
}

.cron-hint {
  margin-top: 4px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>

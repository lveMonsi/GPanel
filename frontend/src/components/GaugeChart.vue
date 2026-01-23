<template>
  <div class="gauge-chart">
    <div class="gauge-container">
      <svg :width="size" :height="size * 0.7" viewBox="0 0 200 140">
        <!-- 背景圆弧 -->
        <path
          :d="backgroundArc"
          fill="none"
          :stroke="backgroundColor"
          :stroke-width="strokeWidth"
          stroke-linecap="round"
        />
        <!-- 进度圆弧 -->
        <path
          :d="progressArc"
          fill="none"
          :stroke="progressColor"
          :stroke-width="strokeWidth"
          stroke-linecap="round"
          class="progress-arc"
        />
      </svg>
      <div class="gauge-content">
        <div class="gauge-value">{{ displayValue }}</div>
        <div class="gauge-label">{{ label }}</div>
        <div v-if="subLabel" class="gauge-sublabel">{{ subLabel }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  value: number
  maxValue?: number
  label: string
  subLabel?: string
  size?: number
  strokeWidth?: number
  backgroundColor?: string
  progressColor?: string
  unit?: string
}

const props = withDefaults(defineProps<Props>(), {
  maxValue: 100,
  size: 140,
  strokeWidth: 12,
  backgroundColor: '#e5e7eb',
  unit: ''
})

const percentage = computed(() => {
  return Math.min((props.value / props.maxValue) * 100, 100)
})

const displayValue = computed(() => {
  if (props.unit) {
    return `${props.value.toFixed(2)}${props.unit}`
  }
  return props.value.toFixed(2)
})

const progressColor = computed(() => {
  if (props.progressColor) {
    return props.progressColor
  }
  const pct = percentage.value
  if (pct < 50) return '#10b981'
  if (pct < 75) return '#f59e0b'
  return '#ef4444'
})

const centerX = 100
const centerY = 100
const radius = 80
const startAngle = -180
const endAngle = 0

const polarToCartesian = (cx: number, cy: number, r: number, angleInDegrees: number) => {
  const angleInRadians = (angleInDegrees * Math.PI) / 180
  return {
    x: cx + r * Math.cos(angleInRadians),
    y: cy + r * Math.sin(angleInRadians)
  }
}

const describeArc = (startAngle: number, endAngle: number) => {
  const start = polarToCartesian(centerX, centerY, radius, endAngle)
  const end = polarToCartesian(centerX, centerY, radius, startAngle)
  const largeArcFlag = endAngle - startAngle <= 180 ? 0 : 1
  return [
    'M',
    start.x,
    start.y,
    'A',
    radius,
    radius,
    0,
    largeArcFlag,
    0,
    end.x,
    end.y
  ].join(' ')
}

const backgroundArc = computed(() => {
  return describeArc(startAngle, endAngle)
})

const progressArc = computed(() => {
  const currentAngle = startAngle + (endAngle - startAngle) * (percentage.value / 100)
  return describeArc(startAngle, currentAngle)
})
</script>

<style scoped>
.gauge-chart {
  display: flex;
  justify-content: center;
  align-items: center;
}

.gauge-container {
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
}

.gauge-content {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -40%);
  text-align: center;
}

.gauge-value {
  font-size: 1.4rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1;
  margin-bottom: 0.25rem;
}

.gauge-label {
  font-size: 0.7rem;
  color: var(--text-secondary);
  font-weight: 500;
  margin-bottom: 0.1rem;
}

.gauge-sublabel {
  font-size: 0.6rem;
  color: var(--text-secondary);
  font-weight: 400;
}

.progress-arc {
  transition: d 0.5s ease-in-out;
}

@media (max-width: 1024px) {
  .gauge-value {
    font-size: 1.2rem;
  }

  .gauge-label {
    font-size: 0.65rem;
  }

  .gauge-sublabel {
    font-size: 0.55rem;
  }
}

@media (max-width: 640px) {
  .gauge-value {
    font-size: 1.1rem;
  }

  .gauge-label {
    font-size: 0.6rem;
  }

  .gauge-sublabel {
    font-size: 0.5rem;
  }
}
</style>
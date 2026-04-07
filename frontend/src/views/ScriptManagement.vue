<template>
  <div class="placeholder-page">
    <div class="placeholder-card">
      <span class="placeholder-tag">即将开放</span>
      <h1 class="placeholder-title">脚本管理</h1>
      <p class="placeholder-description">
        当前为页面占位文件，后续可在这里补充脚本列表、分类管理、执行历史与参数配置等能力。
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.placeholder-page {
  min-height: calc(100vh - 60px);
  padding: 1.25rem;
  background: var(--bg-color);
}

.placeholder-card {
  max-width: 720px;
  padding: 2rem;
  background: var(--card-bg);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-sm);
}

.placeholder-tag {
  display: inline-flex;
  align-items: center;
  padding: 0.35rem 0.75rem;
  border-radius: 999px;
  background: rgba(64, 158, 255, 0.12);
  color: var(--primary);
  font-size: 0.8rem;
  font-weight: 500;
}

.placeholder-title {
  margin: 1rem 0 0.75rem;
  color: var(--text-primary);
  font-size: 1.75rem;
  font-weight: 600;
}

.placeholder-description {
  margin: 0;
  color: var(--text-secondary);
  font-size: 0.95rem;
  line-height: 1.7;
}
</style>

<template>
  <div class="not-found-container">
    <div class="not-found-card">
      <div class="icon-wrapper">
        <svg class="lock-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 2C9.243 2 7 4.243 7 7V10H6C4.89543 10 4 10.8954 4 12V20C4 21.1046 4.89543 22 6 22H18C19.1046 22 20 21.1046 20 20V12C20 10.8954 19.1046 10 18 10H17V7C17 4.243 14.757 2 12 2ZM9 7C9 5.34315 10.3431 4 12 4C13.6569 4 15 5.34315 15 7V10H9V7ZM6 12H18V20H6V12Z" fill="var(--primary-dark)"/>
        </svg>
      </div>
      <h1 class="title">暂时无法访问</h1>
      <p class="description">当前环境已经开启了安全入口登录</p>
      <p class="instruction">可在 SSH 终端输入以下命令来查看面板入口：</p>
      <div class="code-block">
        <code>gpctl user-info</code>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.not-found-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-color);
  padding: 20px;
}

.not-found-card {
  background: var(--card-bg);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  padding: 1.75rem 2rem;
  max-width: 420px;
  width: 100%;
  text-align: center;
  border: 1px solid var(--border-color);
}

.icon-wrapper {
  margin-bottom: 1rem;
}

.lock-icon {
  width: 48px;
  height: 48px;
  color: var(--primary-dark);
}

.title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 0.75rem 0;
  letter-spacing: -0.2px;
}

.description {
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.5;
  margin: 0 0 0.5rem 0;
}

.instruction {
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.5;
  margin: 0 0 0.75rem 0;
}

.code-block {
  display: inline-block;
  background: var(--bg-color);
  border-radius: var(--radius-sm);
  padding: 0.5rem 0.75rem;
  margin-top: 0.75rem;
  border: 1px solid var(--border-color);
}

.code-block code {
  font-family: 'Courier New', 'Consolas', monospace;
  font-size: 0.8rem;
  color: var(--text-primary);
  font-weight: 500;
}

@media (max-width: 480px) {
  .not-found-card {
    padding: 1.5rem 1.25rem;
  }

  .title {
    font-size: 0.95rem;
  }

  .description,
  .instruction {
    font-size: 0.8rem;
  }
}
</style>
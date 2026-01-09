<template>
  <div class="not-found-container">
    <div class="not-found-card">
      <div class="icon-wrapper">
        <svg class="lock-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M12 2C9.243 2 7 4.243 7 7V10H6C4.89543 10 4 10.8954 4 12V20C4 21.1046 4.89543 22 6 22H18C19.1046 22 20 21.1046 20 20V12C20 10.8954 19.1046 10 18 10H17V7C17 4.243 14.757 2 12 2ZM9 7C9 5.34315 10.3431 4 12 4C13.6569 4 15 5.34315 15 7V10H9V7ZM6 12H18V20H6V12Z" fill="#667eea"/>
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
  background: #F5F5F5;
  padding: 20px;
}

.not-found-card {
  background: #FFFFFF;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
  padding: 32px 40px;
  max-width: 500px;
  width: 100%;
  text-align: center;
}

.icon-wrapper {
  margin-bottom: 24px;
}

.lock-icon {
  width: 64px;
  height: 64px;
  color: #667eea;
}

.title {
  font-size: 22px;
  font-weight: 600;
  color: #333333;
  margin: 0 0 16px 0;
}

.description {
  font-size: 15px;
  color: #666666;
  line-height: 1.6;
  margin: 0 0 12px 0;
}

.instruction {
  font-size: 15px;
  color: #666666;
  line-height: 1.6;
  margin: 0 0 16px 0;
}

.code-block {
  display: inline-block;
  background: #EFEFEF;
  border-radius: 4px;
  padding: 8px 12px;
  margin-top: 16px;
}

.code-block code {
  font-family: 'Courier New', 'Consolas', monospace;
  font-size: 14px;
  color: #333333;
  font-weight: 500;
}

@media (max-width: 480px) {
  .not-found-card {
    padding: 24px 20px;
  }

  .title {
    font-size: 20px;
  }

  .description,
  .instruction {
    font-size: 14px;
  }
}
</style>
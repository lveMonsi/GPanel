<template>
  <div class="home">
    <div class="logo">
      <div class="logo-icon">G</div>
    </div>
    <h1 class="title">GPanel</h1>
    <p class="subtitle">Linux 服务器管理面板</p>
    <router-link to="/dashboard" class="enter-btn">进入仪表盘</router-link>
  </div>
</template>

<script setup lang="ts">
</script>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: linear-gradient(135deg, var(--primary-light) 0%, var(--primary) 50%, var(--primary-dark) 100%);
  color: white;
}

.logo {
  margin-bottom: 1.5rem;
}

.logo-icon {
  width: 64px;
  height: 64px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  font-weight: 700;
  background: linear-gradient(135deg, var(--primary) 0%, var(--primary-dark) 100%);
  color: white;
  box-shadow: var(--shadow-lg);
}

.title {
  font-size: 2.5rem;
  font-weight: 600;
  margin-bottom: 0.5rem;
  letter-spacing: -1px;
}

.subtitle {
  font-size: 0.95rem;
  margin-bottom: 2rem;
  opacity: 0.9;
  font-weight: 400;
}

.enter-btn {
  color: white;
  text-decoration: none;
  padding: 0.65rem 1.5rem;
  background: rgba(255, 255, 255, 0.2);
  border-radius: var(--radius-sm);
  transition: all 0.2s;
  font-size: 0.85rem;
  font-weight: 500;
  border: 1px solid rgba(255, 255, 255, 0.3);
}

.enter-btn:hover {
  background: rgba(255, 255, 255, 0.3);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}
</style>
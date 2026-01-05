<template>
  <div class="home">
    <h1>GPanel</h1>
    <p>Linux 服务器管理面板</p>
    <router-link to="/dashboard">进入仪表盘</router-link>
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
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
}

.home h1 {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.home p {
  font-size: 1.5rem;
  margin-bottom: 2rem;
}

.home a {
  color: white;
  text-decoration: none;
  padding: 12px 24px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 8px;
  transition: background 0.3s;
}

.home a:hover {
  background: rgba(255, 255, 255, 0.3);
}
</style>
<template>
  <div class="layout">
    <aside class="sidebar">
      <div class="sidebar-header">
        <div class="logo-icon">G</div>
        <h2>GPanel</h2>
      </div>
      <nav class="sidebar-nav">
        <router-link to="/dashboard" class="nav-item" active-class="active">
          <el-icon class="nav-icon"><DataBoard /></el-icon>
          <span class="nav-text">概览</span>
        </router-link>
        <router-link to="/settings" class="nav-item" active-class="active">
          <el-icon class="nav-icon"><Setting /></el-icon>
          <span class="nav-text">面板设置</span>
        </router-link>
        <a href="#" class="nav-item" @click.prevent="handleLogout">
          <el-icon class="nav-icon"><SwitchButton /></el-icon>
          <span class="nav-text">退出登录</span>
        </a>
      </nav>
    </aside>
    <main class="main-content">
      <router-view />
    </main>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { DataBoard, Setting, SwitchButton } from '@element-plus/icons-vue'

const router = useRouter()

const handleLogout = () => {
  localStorage.removeItem('token')
  router.push('/login')
}
</script>

<style scoped>
.layout {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: 200px;
  background: linear-gradient(180deg, var(--primary) 0%, var(--primary-dark) 100%);
  color: white;
  display: flex;
  flex-direction: column;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.05);
}

.sidebar-header {
  padding: 1.5rem 1.25rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.15);
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.logo-icon {
  width: 32px;
  height: 32px;
  background: rgba(255, 255, 255, 0.95);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.9rem;
  font-weight: 700;
  color: var(--primary-dark);
  flex-shrink: 0;
}

.sidebar-header h2 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  letter-spacing: -0.3px;
}

.sidebar-nav {
  flex: 1;
  padding: 1rem 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.65rem 1rem;
  color: white;
  text-decoration: none;
  transition: all 0.2s;
  border-radius: var(--radius-sm);
  margin: 0 0.5rem;
  font-size: 0.85rem;
  font-weight: 400;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.15);
}

.nav-item.active {
  background: rgba(255, 255, 255, 0.2);
  font-weight: 500;
}

.nav-icon {
  font-size: 1rem;
  opacity: 0.9;
}

.nav-text {
  font-size: 0.85rem;
}

.main-content {
  flex: 1;
  background-color: var(--bg-color);
  overflow-y: auto;
}
</style>
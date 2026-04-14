<template>
  <div class="layout">
    <aside :class="['sidebar', { collapsed: isSidebarCollapsed }]">
      <div class="sidebar-header">
        <div class="sidebar-brand">
          <div class="logo-icon">G</div>
          <h2>GPanel</h2>
        </div>
        <button
          type="button"
          class="sidebar-toggle"
          @click="toggleSidebar"
          :aria-label="isSidebarCollapsed ? '展开侧栏' : '收起侧栏'"
          :title="isSidebarCollapsed ? '展开侧栏' : '收起侧栏'"
        >
          <el-icon class="sidebar-toggle-icon"><DArrowLeft /></el-icon>
        </button>
      </div>
      <nav class="sidebar-nav">
        <router-link to="/dashboard" class="nav-item" active-class="active" title="概览">
          <el-icon class="nav-icon"><DataBoard /></el-icon>
          <span class="nav-text">概览</span>
        </router-link>
        <router-link to="/files" class="nav-item" active-class="active" title="文件管理">
          <el-icon class="nav-icon"><Folder /></el-icon>
          <span class="nav-text">文件管理</span>
        </router-link>
        <router-link to="/firewall" class="nav-item" active-class="active" title="防火墙">
          <el-icon class="nav-icon"><Lock /></el-icon>
          <span class="nav-text">防火墙</span>
        </router-link>
        <router-link to="/terminal" class="nav-item" active-class="active" title="终端">
          <el-icon class="nav-icon"><Monitor /></el-icon>
          <span class="nav-text">终端</span>
        </router-link>
        <router-link to="/ssh" class="nav-item" active-class="active" title="SSH管理">
          <el-icon class="nav-icon"><Connection /></el-icon>
          <span class="nav-text">SSH管理</span>
        </router-link>
        <router-link to="/process" class="nav-item" active-class="active" title="进程管理">
          <el-icon class="nav-icon"><Cpu /></el-icon>
          <span class="nav-text">进程管理</span>
        </router-link>
        <router-link to="/cronjob" class="nav-item" active-class="active" title="定时任务">
          <el-icon class="nav-icon"><Timer /></el-icon>
          <span class="nav-text">定时任务</span>
        </router-link>
        <router-link to="/logs" class="nav-item" active-class="active" title="日志管理">
          <el-icon class="nav-icon"><Document /></el-icon>
          <span class="nav-text">日志管理</span>
        </router-link>
        <router-link to="/monitor" class="nav-item" active-class="active" title="监控">
          <el-icon class="nav-icon"><Monitor /></el-icon>
          <span class="nav-text">监控</span>
        </router-link>
        <router-link to="/settings" class="nav-item" active-class="active" title="面板设置">
          <el-icon class="nav-icon"><Setting /></el-icon>
          <span class="nav-text">面板设置</span>
        </router-link>
        <a href="#" class="nav-item" title="退出登录" @click.prevent="handleLogout">
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
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessageBox } from 'element-plus'
import { DataBoard, Setting, SwitchButton, Folder, Lock, Monitor, Connection, Document, Timer, Cpu, DArrowLeft } from '@element-plus/icons-vue'

const router = useRouter()
const isSidebarCollapsed = ref(false)

const toggleSidebar = () => {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
}

const handleLogout = () => {
  ElMessageBox.confirm('确定要退出登录吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    localStorage.removeItem('token')
    router.push('/login')
  }).catch(() => {
    // 用户取消操作
  })
}
</script>

<style scoped>
.layout {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: 220px;
  background:
    linear-gradient(180deg, rgba(35, 63, 79, 0.16) 0%, rgba(35, 63, 79, 0.24) 100%),
    linear-gradient(180deg, var(--primary) 0%, var(--primary-dark) 100%);
  color: white;
  display: flex;
  flex-direction: column;
  box-shadow: 2px 0 18px rgba(10, 37, 64, 0.12);
  flex-shrink: 0;
  transition: width 0.25s ease, box-shadow 0.25s ease;
}

.sidebar.collapsed {
  width: 72px;
}

.sidebar-header {
  padding: 1.25rem 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.15);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.sidebar-brand {
  min-width: 0;
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
  box-shadow: 0 8px 18px rgba(16, 24, 40, 0.16);
}

.sidebar-header h2 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
  letter-spacing: -0.3px;
  white-space: nowrap;
  transition: opacity 0.2s ease, width 0.2s ease, margin 0.2s ease;
}

.sidebar-toggle {
  width: 30px;
  height: 30px;
  border: none;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.14);
  color: rgba(255, 255, 255, 0.95);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  transition: background 0.2s ease, transform 0.25s ease;
}

.sidebar-toggle:hover {
  background: rgba(255, 255, 255, 0.22);
}

.sidebar-toggle-icon {
  font-size: 0.9rem;
  transition: transform 0.25s ease;
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
  color: rgba(255, 255, 255, 0.96);
  text-decoration: none;
  transition: all 0.2s;
  border-radius: var(--radius-sm);
  margin: 0 0.5rem;
  font-size: 0.85rem;
  font-weight: 400;
  white-space: nowrap;
  overflow: hidden;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.18);
  transform: translateX(2px);
}

.nav-item.active {
  background: rgba(255, 255, 255, 0.24);
  font-weight: 500;
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.08);
}

.nav-icon {
  font-size: 1.05rem;
  opacity: 1;
  flex-shrink: 0;
}

.nav-text {
  font-size: 0.85rem;
  opacity: 1;
  transition: opacity 0.15s ease, width 0.15s ease;
}

.sidebar.collapsed .sidebar-header {
  padding: 1.25rem 0.75rem;
  flex-direction: column;
  justify-content: flex-start;
}

.sidebar.collapsed .sidebar-brand {
  width: 100%;
  justify-content: center;
}

.sidebar.collapsed .sidebar-header h2,
.sidebar.collapsed .nav-text {
  width: 0;
  opacity: 0;
  margin: 0;
}

.sidebar.collapsed .sidebar-toggle-icon {
  transform: rotate(180deg);
}

.sidebar.collapsed .sidebar-nav {
  padding: 1rem 0;
  align-items: center;
}

.sidebar.collapsed .nav-item {
  width: 48px;
  height: 44px;
  padding: 0;
  margin: 0;
  justify-content: center;
  border-radius: 14px;
}

.sidebar.collapsed .nav-item:hover {
  transform: none;
}

.sidebar.collapsed .sidebar-toggle {
  margin-top: 0.25rem;
}

.main-content {
  flex: 1;
  background-color: var(--bg-color);
  overflow-y: auto;
}
</style>

import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import Layout from '@/components/Layout.vue'
import Login from '@/views/Login.vue'
import Home from '@/views/Home.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/',
    name: 'Home',
    component: Home
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'DashboardMain',
        component: () => import('@/views/Dashboard.vue')
      }
    ]
  },
  {
    path: '/settings',
    name: 'Settings',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'SettingsMain',
        component: () => import('@/views/Settings.vue')
      }
    ]
  },
  {
    path: '/files',
    name: 'Files',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'FilesMain',
        component: () => import('@/views/Files.vue')
      }
    ]
  },
  {
    path: '/firewall',
    name: 'Firewall',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'FirewallMain',
        component: () => import('@/views/Firewall.vue')
      }
    ]
  },
  {
    path: '/hosts',
    name: 'Hosts',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'HostsMain',
        component: () => import('@/views/Hosts.vue')
      }
    ]
  },
  {
    path: '/terminal',
    name: 'Terminal',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'TerminalMain',
        component: () => import('@/views/Terminal.vue')
      }
    ]
  },
  {
    path: '/monitor',
    name: 'Monitor',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'MonitorMain',
        component: () => import('@/views/Monitor.vue')
      }
    ]
  },
  {
    path: '/ssh',
    name: 'SshManagement',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'SshManagementMain',
        component: () => import('@/views/SshManagement.vue')
      }
    ]
  },
  {
    path: '/process',
    name: 'ProcessManagement',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'ProcessManagementMain',
        component: () => import('@/views/ProcessManagement.vue')
      }
    ]
  },
  {
    path: '/scripts',
    name: 'ScriptManagement',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'ScriptManagementMain',
        component: () => import('@/views/ScriptManagement.vue')
      }
    ]
  },
  {
    path: '/logs',
    name: 'LogManagement',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'LogManagementMain',
        component: () => import('@/views/LogManagement.vue')
      }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 检查 sessionkey 是否存在
const checkSessionKey = (): boolean => {
  return document.cookie.includes('sessionkey=')
}

router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  const hasSessionKey = checkSessionKey()

  // 如果访问登录页面
  if (to.path === '/login') {
    // 如果有 token，重定向到 dashboard
    if (token) {
      next('/dashboard')
      return
    }
    // 没有 token，允许访问登录页面
    next()
    return
  }

  // 如果需要认证但没有 token
  if (to.meta.requiresAuth && !token) {
    next('/login')
    return
  }

  next()
})

export default router

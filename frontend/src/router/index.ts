import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import Layout from '@/components/Layout.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/Login.vue')
  },
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue')
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '/dashboard',
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
        path: '/settings',
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
        path: '/files',
        name: 'FilesMain',
        component: () => import('@/views/Files.vue')
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
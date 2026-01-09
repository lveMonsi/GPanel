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

  // 如果需要认证但没有 token
  if (to.meta.requiresAuth && !token) {
    next('/login')
    return
  }

  // 如果访问登录页面但有 token
  if (to.path === '/login' && token) {
    next('/dashboard')
    return
  }

  // 如果访问登录页面但没有 sessionkey 且有 token（说明 sessionkey 已过期）
  if (to.path === '/login' && !hasSessionKey && token) {
    alert('登录已失效，请重新从安全入口进入登录页面')
    localStorage.removeItem('token')
    // 不跳转，让后端返回404页面
    return
  }

  next()
})

export default router
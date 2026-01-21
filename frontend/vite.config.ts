import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  base: '/', // 使用根路径，配合后端处理
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      }
    },
    headers: {
      'Cache-Control': 'no-cache, no-store, must-revalidate',
      'Pragma': 'no-cache',
      'Expires': '0'
    }
  },
  build: {
    // 生产环境构建优化
    target: 'es2015',
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: false, // 暂时不禁用 console，以便调试
        drop_debugger: true
      }
    },
    chunkSizeWarningLimit: 1000, // 调整 chunk 大小警告限制为 1000KB
    rollupOptions: {
      output: {
        // 细粒度代码分割
        manualChunks(id) {
          // Vue 核心库
          if (id.includes('node_modules/vue/') || id.includes('node_modules/vue-router/') || id.includes('node_modules/pinia/')) {
            return 'vue-vendor'
          }
          // Element Plus 组件库
          if (id.includes('node_modules/element-plus/')) {
            return 'element-plus'
          }
          // Element Plus 图标库
          if (id.includes('node_modules/@element-plus/icons-vue/')) {
            return 'element-icons'
          }
          // HTTP 客户端
          if (id.includes('node_modules/axios/')) {
            return 'http-vendor'
          }
          // 其他第三方库
          if (id.includes('node_modules/')) {
            return 'vendor'
          }
        }
      }
    },
    // 资源内联限制
    assetsInlineLimit: 4096,
    // 源码映射（生产环境关闭）
    sourcemap: false
  }
})
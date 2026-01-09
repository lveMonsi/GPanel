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
    }
  },
  build: {
    // 生产环境构建优化
    target: 'es2015',
    minify: 'terser',
    terserOptions: {
      compress: {
        drop_console: true, // 生产环境移除 console
        drop_debugger: true
      }
    },
    rollupOptions: {
      output: {
        // 代码分割
        manualChunks: {
          'vue-vendor': ['vue', 'vue-router', 'pinia'],
          'http-vendor': ['axios']
        }
      }
    },
    // 资源内联限制
    assetsInlineLimit: 4096,
    // 源码映射（生产环境关闭）
    sourcemap: false
  }
})
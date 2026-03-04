import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

// Monaco Editor Worker 前缀路径
const monacoPrefix = 'monaco-editor/esm/vs'

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
    // 使用 esbuild 替代 terser，更快且内存占用更少
    minify: 'esbuild',
    chunkSizeWarningLimit: 2000, // Monaco Editor 较大，放宽限制
    rollupOptions: {
      output: {
        // 细粒度代码分割
        manualChunks(id) {
          // Monaco Editor Workers - 单独打包
          if (id.includes(`${monacoPrefix}/language/json/json.worker`)) {
            return 'monaco-json-worker'
          }
          if (id.includes(`${monacoPrefix}/language/css/css.worker`)) {
            return 'monaco-css-worker'
          }
          if (id.includes(`${monacoPrefix}/language/html/html.worker`)) {
            return 'monaco-html-worker'
          }
          if (id.includes(`${monacoPrefix}/language/typescript/ts.worker`)) {
            return 'monaco-ts-worker'
          }
          if (id.includes(`${monacoPrefix}/editor/editor.worker`)) {
            return 'monaco-editor-worker'
          }
          // Monaco Editor 核心模块拆分
          if (id.includes(`${monacoPrefix}/base/`)) {
            return 'monaco-base'
          }
          if (id.includes(`${monacoPrefix}/editor/`)) {
            return 'monaco-editor-core'
          }
          if (id.includes(`${monacoPrefix}/platform/`)) {
            return 'monaco-platform'
          }
          if (id.includes(`${monacoPrefix}/language/`)) {
            return 'monaco-language'
          }
          // Monaco Editor 其他部分
          if (id.includes('node_modules/monaco-editor/')) {
            return 'monaco-editor'
          }
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
    sourcemap: false,
    // 关闭 CSS 代码分割可减少构建内存
    cssCodeSplit: true,
  }
})
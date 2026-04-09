import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
    }),
    Components({
      resolvers: [ElementPlusResolver()],
    }),
  ],
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
          // Monaco Editor - 按需加载，不打入主包
          if (id.includes('node_modules/monaco-editor/')) {
            return 'monaco-editor'
          }
          // xterm.js - 终端组件按需加载
          if (id.includes('node_modules/@xterm/')) {
            return 'xterm'
          }
          // echarts - 图表库按需加载
          if (id.includes('node_modules/echarts/')) {
            return 'echarts'
          }
          // Vue 核心库 - 关键依赖，优先加载
          if (id.includes('node_modules/vue/') || id.includes('node_modules/vue-router/') || id.includes('node_modules/pinia/')) {
            return 'vue-vendor'
          }
          // Element Plus 组件库 - 拆分为独立包
          if (id.includes('node_modules/element-plus/')) {
            return 'element-plus'
          }
          // Element Plus 图标库 - 独立加载
          if (id.includes('node_modules/@element-plus/icons-vue/')) {
            return 'element-icons'
          }
          // HTTP 客户端 - 关键依赖
          if (id.includes('node_modules/axios/')) {
            return 'http-vendor'
          }
          // 其他第三方库
          if (id.includes('node_modules/')) {
            return 'vendor'
          }
        },
        // 优化输出文件名，添加 hash 以支持长期缓存
        entryFileNames: 'assets/[name]-[hash].js',
        chunkFileNames: 'assets/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]'
      }
    },
    // 资源内联限制
    assetsInlineLimit: 4096,
    // 源码映射（生产环境关闭）
    sourcemap: false,
    // CSS 代码分割
    cssCodeSplit: true,
    // 压缩选项
    minify: 'esbuild',
    // 清理输出目录
    emptyOutDir: true,
  },
  // 优化依赖预构建 - 移除 monaco-editor，按需加载
  optimizeDeps: {
    include: ['vue', 'vue-router', 'pinia', 'axios', 'element-plus'],
    exclude: ['monaco-editor', '@xterm/xterm', '@xterm/addon-fit', 'echarts'],
  },
})
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
      '@napi-rs/canvas': path.resolve(__dirname, 'src/shims/napi-canvas.ts'),
    },
  },
  server: {
    // 端口可配:FRONTEND_PORT=3001 npm run dev(避免与其他本地项目冲突,默认3000)
    port: Number(process.env.FRONTEND_PORT) || 3000,
    proxy: {
      '/api': {
        target: process.env.BACKEND_URL || 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'editor-heavy': ['@eigenpal/docx-editor-vue', '@tiptap/vue-3', '@tiptap/starter-kit'],
          'markdown-utils': ['marked', 'turndown'],
        },
      },
    },
    chunkSizeWarningLimit: 800,
  },
})


import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  build: {
    // 輸出到 dist/，供 Go go:embed 嵌入
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    // 開發時代理 API 請求到 Go 後端
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})

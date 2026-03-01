import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
      // Interpreter service (трассировка) - порт 8084
      '/api/snapshot': {
        target: 'http://localhost:8084',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      },
      // Auth service - порт 8083
      '/api/auth': {
        target: 'http://localhost:8083',
        changeOrigin: true
      },
      // Flowchart service - порт 8081
      '/api/flowchart': {
        target: 'http://localhost:8081',
        changeOrigin: true
      }
    }
  }
})

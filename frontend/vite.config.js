import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
      '/api/snapshot': {
        target: 'http://localhost:8084',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      },
      '/api/analyze': {
        target: 'http://localhost:8086',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, '')
      },
      '/api/auth': {
        target: 'http://localhost:8083',
        changeOrigin: true
      },
      '/api/flowchart': {
        target: 'http://localhost:8081',
        changeOrigin: true
      },

      '/api/metrics': {
        target: 'http://localhost:8085',
        changeOrigin: true
      }
    }
  }
})
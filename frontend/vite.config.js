import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// Общая функция для обработки ошибок прокси —
// при ECONNREFUSED вместо дефолтного 500 отдаём 503
// чтобы фронт мог понять: сервис недоступен, а не упал
function proxyErrorHandler(proxyName) {
  return (err, req, res) => {
    if (!res.headersSent) {
      res.writeHead(503, { 'Content-Type': 'text/plain' })
    }
    res.end(`${proxyName} unavailable`)
  }
}

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3000,
    proxy: {
      '/api/snapshot': {
        target: 'http://localhost:8084',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
        configure: (proxy) => {
          proxy.on('error', proxyErrorHandler('snapshot-service'))
        }
      },
      '/api/analyze': {
        target: 'http://localhost:8086',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/api/, ''),
        configure: (proxy) => {
          proxy.on('error', proxyErrorHandler('analyze-service'))
        }
      },
      '/api/auth': {
        target: 'http://localhost:8083',
        changeOrigin: true,
        configure: (proxy) => {
          proxy.on('error', proxyErrorHandler('auth-service'))
        }
      },
      '/api/flowchart': {
        target: 'http://localhost:8081',
        changeOrigin: true,
        configure: (proxy) => {
          proxy.on('error', proxyErrorHandler('flowchart-service'))
        }
      },
      '/api/metrics': {
        target: 'http://localhost:8085',
        changeOrigin: true,
        configure: (proxy) => {
          proxy.on('error', proxyErrorHandler('metrics-service'))
        }
      }
    }
  }
})

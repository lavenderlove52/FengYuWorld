import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  base: '/blog/',
  plugins: [
    vue(),
    tailwindcss(),
    // 自动将 /blog 重定向到 /blog/，避免 Vite 显示提示页
    {
      name: 'base-redirect',
      configureServer(server) {
        server.middlewares.use((req, res, next) => {
          if (req.url === '/blog') {
            res.writeHead(301, { Location: '/blog/' })
            res.end()
            return
          }
          next()
        })
      },
    },
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/blog/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/blog\/api/, '/api'),
        ws: true,
      },
      '/uploads': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})

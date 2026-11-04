import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

// The Go server embeds the build from internal/views/dist/app and serves
// the API under /api/. The dev server (npm run dev) forwards API
// requests to the Go server on port 8080.
// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
  ],
  build: {
    outDir: '../internal/views/dist/app',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
})

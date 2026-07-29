import { defineConfig } from 'vite'
import { fileURLToPath, URL } from 'node:url'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  optimizeDeps: { exclude: ['@skykin/ui'] },
  server: {
    port: 3001,
    fs: { allow: ['..'] },
    proxy: {
      '/api': 'http://localhost:8081',
    },
  },
})

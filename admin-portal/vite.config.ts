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
  // @skykin/ui is a workspace source package — let Vite transpile it, don't prebundle
  optimizeDeps: { exclude: ['@skykin/ui'] },
  server: {
    port: 3002,
    fs: { allow: ['..'] },
    proxy: {
      '/api': 'http://localhost:8081',
    },
  },
})

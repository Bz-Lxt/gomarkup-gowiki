import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'

export default defineConfig({
  plugins: [vue(), UnoCSS()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://127.0.0.1:27122',
      '/uploads': 'http://127.0.0.1:27122',
      '/ws': { target: 'http://127.0.0.1:27122', ws: true },
    },
  },
})

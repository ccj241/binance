import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/login': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/api-key': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/balance': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/orders': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/trades': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/prices': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/strategy': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/strategies': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/order': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/withdrawals': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/symbols': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
});
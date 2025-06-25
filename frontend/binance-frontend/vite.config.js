// frontend/binance-frontend/vite.config.js - 添加删除路由的代理配置

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
      '/register': {
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
      '/withdrawalhistory': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/symbols': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/cancel_order': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/batch_cancel_orders': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/toggle_strategy': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/delete_strategy': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
      '/admin': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
});
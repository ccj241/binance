import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',  // 允许外部访问
    port: 8080,       // 容器内部端口保持 8080
    proxy: {
      // 开发环境代理配置 - 指向后端容器
      '/login': {
        target: 'http://binance-backend:23337',  // 使用容器名称
        changeOrigin: true,
      },
      '/register': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/api-key': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/balance': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/orders': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/trades': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/prices': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/strategy': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/strategies': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/order': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/withdrawals': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/withdrawalhistory': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/auto_withdraw_rule': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/auto_withdraw_rules': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/withdraw_coins': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/symbols': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/cancel_order': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/batch_cancel_orders': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/toggle_strategy': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/delete_strategy': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/admin': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/dual-investment': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
      '/futures': {
        target: 'http://binance-backend:23337',
        changeOrigin: true,
      },
    },
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    sourcemap: false,
  }
});
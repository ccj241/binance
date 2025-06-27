import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',  // 添加这行，允许外部访问
    port: 8080,       // 添加这行，明确指定端口
    proxy: {
      '/login': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/register': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/api-key': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/balance': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/orders': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/trades': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/prices': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/strategy': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/strategies': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/order': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/withdrawals': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/withdrawalhistory': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      // 自动提币相关端点
      '/auto_withdraw_rule': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/auto_withdraw_rules': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/withdraw_coins': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/symbols': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/cancel_order': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/batch_cancel_orders': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/toggle_strategy': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/delete_strategy': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/admin': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
      '/dual-investment': {
        target: 'http://host.docker.internal:8081',
        changeOrigin: true,
      },
    },
  },
});
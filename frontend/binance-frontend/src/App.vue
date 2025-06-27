<template>
  <div class="app-container">
    <nav v-if="isLoggedIn && $route.path !== '/login' && $route.path !== '/register'">
      <div class="nav-container">
        <div class="nav-brand">
          <span class="brand-text">交易系统</span>
        </div>
        <div class="nav-links">
          <router-link to="/" class="nav-link">
            <span class="nav-icon">📊</span>
            <span class="nav-text">仪表盘</span>
          </router-link>

          <router-link to="/orders" class="nav-link">
            <span class="nav-icon">📋</span>
            <span class="nav-text">订单</span>
          </router-link>

          <router-link to="/strategies" class="nav-link">
            <span class="nav-icon">🎯</span>
            <span class="nav-text">策略</span>
          </router-link>

          <router-link to="/settings" class="nav-link">
            <span class="nav-icon">⚙️</span>
            <span class="nav-text">设置</span>
          </router-link>

          <router-link to="/dual-investment" class="nav-link">
            <span class="nav-icon">💰</span>
            <span class="nav-text">双币投资</span>
          </router-link>

          <router-link v-if="isAdmin" to="/admin" class="nav-link">
            <span class="nav-icon">👤</span>
            <span class="nav-text">管理</span>
          </router-link>
        </div>

        <div class="nav-user">
          <div class="user-info">
            <span class="username">{{ username }}</span>
            <span class="user-role" :class="isAdmin ? 'admin' : 'user'">
          {{ isAdmin ? '管理员' : '用户' }}
        </span>
          </div>
          <button @click="logout" class="logout-btn">
            退出
          </button>
        </div>
      </div>
    </nav>

    <main class="main-content">
      <router-view></router-view>
    </main>
  </div>
</template>
<script>
export default {
  computed: {
    isLoggedIn() {
      return !!localStorage.getItem('token');
    },
    isAdmin() {
      const token = localStorage.getItem('token');
      if (!token) return false;

      try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        return payload.is_admin === 'admin';
      } catch (e) {
        return false;
      }
    },
    username() {
      const token = localStorage.getItem('token');
      if (!token) return '';

      try {
        const payload = JSON.parse(atob(token.split('.')[1]));
        return payload.username || '';
      } catch (e) {
        return '';
      }
    }
  },
  methods: {
    logout() {
      localStorage.removeItem('token');
      this.$router.push('/login');
    },
  },
};
</script>
<style>
/* CSS 变量定义 */
:root {
  /* 颜色系统 */
  --color-primary: #2563eb;
  --color-primary-hover: #1d4ed8;
  --color-secondary: #64748b;
  --color-success: #10b981;
  --color-warning: #f59e0b;
  --color-danger: #ef4444;

  /* 中性色 */
  --color-bg: #ffffff;
  --color-bg-secondary: #f8fafc;
  --color-bg-tertiary: #f1f5f9;
  --color-border: #e2e8f0;
  --color-text-primary: #0f172a;
  --color-text-secondary: #475569;
  --color-text-tertiary: #94a3b8;

  /* 间距系统 */
  --spacing-xs: 0.5rem;
  --spacing-sm: 0.75rem;
  --spacing-md: 1rem;
  --spacing-lg: 1.5rem;
  --spacing-xl: 2rem;
  --spacing-2xl: 3rem;

  /* 字体 */
  --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;

  /* 圆角 */
  --radius-sm: 0.25rem;
  --radius-md: 0.375rem;
  --radius-lg: 0.5rem;

  /* 过渡 */
  --transition-fast: 150ms ease;
  --transition-normal: 200ms ease;
}

/* 全局样式重置 */
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

body {
  font-family: var(--font-sans);
  font-size: 14px;
  line-height: 1.5;
  color: var(--color-text-primary);
  background-color: var(--color-bg-secondary);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* 应用容器 */
.app-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

/* 导航栏 */
nav {
  background-color: var(--color-bg);
  border-bottom: 1px solid var(--color-border);
  position: sticky;
  top: 0;
  z-index: 100;
}

.nav-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0 var(--spacing-xl);
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.nav-brand {
  flex-shrink: 0;
}

.brand-text {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.nav-links {
  display: flex;
  gap: var(--spacing-xs);
  flex: 1;
  margin: 0 var(--spacing-xl);
}

.nav-link {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  padding: var(--spacing-sm) var(--spacing-md);
  text-decoration: none;
  color: var(--color-text-secondary);
  border-radius: var(--radius-md);
  transition: all var(--transition-normal);
  font-weight: 500;
}

.nav-link:hover {
  color: var(--color-text-primary);
  background-color: var(--color-bg-tertiary);
}

.nav-link.router-link-exact-active {
  color: var(--color-primary);
  background-color: #eff6ff;
}

.nav-icon {
  font-size: 1.125rem;
}

.nav-text {
  font-size: 0.875rem;
}

/* 用户信息 */
.nav-user {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  flex-shrink: 0;
}

.user-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.username {
  font-weight: 500;
  color: var(--color-text-primary);
}

.user-role {
  padding: 0.125rem var(--spacing-sm);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
  font-weight: 500;
}

.user-role.admin {
  background-color: #dbeafe;
  color: var(--color-primary);
}

.user-role.user {
  background-color: #f3f4f6;
  color: var(--color-text-secondary);
}

.logout-btn {
  padding: var(--spacing-xs) var(--spacing-md);
  background-color: transparent;
  color: var(--color-text-secondary);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
}

.logout-btn:hover {
  color: var(--color-danger);
  border-color: var(--color-danger);
  background-color: #fef2f2;
}

/* 主要内容区域 */
.main-content {
  flex: 1;
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  padding: var(--spacing-xl);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .nav-container {
    padding: 0 var(--spacing-md);
    height: auto;
    flex-direction: column;
    align-items: stretch;
    gap: var(--spacing-md);
    padding-top: var(--spacing-md);
    padding-bottom: var(--spacing-md);
  }

  .nav-brand {
    text-align: center;
  }

  .nav-links {
    margin: 0;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    scrollbar-width: none;
  }

  .nav-links::-webkit-scrollbar {
    display: none;
  }

  .nav-text {
    display: none;
  }

  .nav-user {
    justify-content: space-between;
  }

  .main-content {
    padding: var(--spacing-md);
  }
}

/* 通用组件样式 */
button {
  font-family: inherit;
  cursor: pointer;
}

input, select, textarea {
  font-family: inherit;
}

/* 表格基础样式 */
table {
  width: 100%;
  border-collapse: collapse;
  background-color: var(--color-bg);
}

th {
  background-color: var(--color-bg-tertiary);
  font-weight: 600;
  text-align: left;
  padding: var(--spacing-md);
  color: var(--color-text-primary);
  font-size: 0.875rem;
}

td {
  padding: var(--spacing-md);
  border-top: 1px solid var(--color-border);
  color: var(--color-text-secondary);
}

tr:hover td {
  background-color: var(--color-bg-secondary);
}
</style>

<template>
  <div class="app-container">
    <nav v-if="isLoggedIn && $route.path !== '/login' && $route.path !== '/register'">
      <div class="nav-links">
        <router-link to="/">Dashboard</router-link> |
        <router-link to="/orders">Orders</router-link> |
        <router-link to="/strategies">Strategies</router-link> |
        <router-link to="/settings">Settings</router-link>
        <span v-if="isAdmin"> | <router-link to="/admin">Admin</router-link></span>
      </div>
      <div class="user-info">
        <span class="username">{{ username }}</span>
        <span v-if="isAdmin" class="role-badge admin">管理员</span>
        <span v-else class="role-badge user">用户</span>
        <button @click="logout">退出登录</button>
      </div>
    </nav>
    <router-view></router-view>
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
        return payload.role === 'admin';
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
.app-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 20px;
}

nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 30px;
  padding: 15px 20px;
  background-color: #f5f5f5;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.nav-links a {
  margin: 0 10px;
  text-decoration: none;
  color: #4CAF50;
  font-weight: 500;
  transition: color 0.3s;
}

.nav-links a:hover {
  color: #45a049;
}

.nav-links a.router-link-exact-active {
  font-weight: bold;
  color: #2e7d32;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.username {
  font-weight: bold;
  color: #333;
}

.role-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  font-weight: bold;
}

.role-badge.admin {
  background-color: #2196f3;
  color: white;
}

.role-badge.user {
  background-color: #4caf50;
  color: white;
}

nav button {
  background-color: #ff4444;
  color: white;
  border: none;
  padding: 8px 16px;
  cursor: pointer;
  border-radius: 4px;
  font-size: 14px;
  transition: background-color 0.3s;
}

nav button:hover {
  background-color: #cc0000;
}

/* 全局样式 */
* {
  box-sizing: border-box;
}

body {
  margin: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
  background-color: #fafafa;
  color: #333;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  text-align: left;
  padding: 12px;
  border-bottom: 1px solid #ddd;
}

th {
  background-color: #f5f5f5;
  font-weight: bold;
}

button {
  cursor: pointer;
  transition: all 0.3s;
}

button:hover {
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
}

button:active {
  transform: translateY(0);
  box-shadow: none;
}
</style>
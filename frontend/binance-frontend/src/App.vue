<template>
  <div class="app-container">
    <nav v-if="$route.path !== '/login'">
      <router-link to="/">Dashboard</router-link> |
      <router-link to="/orders">Orders</router-link> |
      <router-link to="/strategies">Strategies</router-link> |
      <router-link to="/settings">Settings</router-link> |
      <button @click="logout">Logout</button>
    </nav>
    <router-view></router-view>
  </div>
</template>

<script>
export default {
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
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}
nav {
  margin-bottom: 20px;
}
nav a {
  margin-right: 10px;
  text-decoration: none;
  color: #4CAF50;
}
nav a.router-link-exact-active {
  font-weight: bold;
}
nav button {
  background-color: #ff4444;
  color: white;
  border: none;
  padding: 5px 10px;
  cursor: pointer;
}
nav button:hover {
  background-color: #cc0000;
}
</style>
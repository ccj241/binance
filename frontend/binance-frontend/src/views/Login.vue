<template>
  <div class="login-container">
    <h2>Login</h2>
    <div v-if="error" class="error">{{ error }}</div>
    <form @submit.prevent="login">
      <div>
        <label>Username:</label>
        <input v-model="username" type="text" required />
      </div>
      <div>
        <label>Password:</label>
        <input v-model="password" type="password" required />
      </div>
      <button type="submit">Login</button>
    </form>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  data() {
    return {
      username: 'testuser',
      password: 'testpass',
      error: '',
    };
  },
  methods: {
    async login() {
      try {
        const response = await axios.post('/login', {
          username: this.username,
          password: this.password,
        });
        localStorage.setItem('token', response.data.token);
        this.$router.push('/');
      } catch (err) {
        this.error = err.response?.data?.error || 'Login failed';
      }
    },
  },
};
</script>

<style scoped>
.login-container {
  max-width: 400px;
  margin: 0 auto;
  padding: 20px;
}
h2 {
  color: #333;
}
input {
  padding: 8px;
  margin: 5px;
  width: 100%;
  border: 1px solid #ddd;
  border-radius: 4px;
}
button {
  background-color: #4CAF50;
  color: white;
  padding: 8px;
  border: none;
  cursor: pointer;
}
button:hover {
  background-color: #45a049;
}
.error {
  color: red;
  margin: 10px 0;
}
</style>
<template>
  <div class="login-container">
    <h2>用户登录</h2>
    <div v-if="error" class="error">{{ error }}</div>
    <form @submit.prevent="login">
      <div>
        <label>用户名:</label>
        <input v-model="username" type="text" required />
      </div>
      <div>
        <label>密码:</label>
        <input v-model="password" type="password" required />
      </div>
      <button type="submit" :disabled="isLoading">
        {{ isLoading ? '登录中...' : '登录' }}
      </button>
    </form>
    <div class="register-link">
      还没有账号？<router-link to="/register">立即注册</router-link>
    </div>
    <div class="demo-info">
      <p>演示账号：</p>
      <p>管理员：admin / admin123</p>
      <p>测试用户：testuser / testpass</p>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  data() {
    return {
      username: '',
      password: '',
      error: '',
      isLoading: false,
    };
  },
  methods: {
    async login() {
      this.error = '';
      this.isLoading = true;

      try {
        const response = await axios.post('/login', {
          username: this.username,
          password: this.password,
        });

        // 保存token和用户角色
        localStorage.setItem('token', response.data.token);

        // 根据角色跳转到不同页面
        if (response.data.role === 'admin') {
          this.$router.push('/admin');
        } else {
          this.$router.push('/');
        }
      } catch (err) {
        if (err.response?.status === 403) {
          this.error = err.response.data.message || '账号状态异常';
        } else {
          this.error = err.response?.data?.message || '登录失败';
        }
      } finally {
        this.isLoading = false;
      }
    },
  },
};
</script>

<style scoped>
.login-container {
  max-width: 400px;
  margin: 50px auto;
  padding: 30px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #f9f9f9;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

h2 {
  color: #333;
  text-align: center;
  margin-bottom: 30px;
}

form > div {
  margin-bottom: 20px;
}

label {
  display: block;
  margin-bottom: 8px;
  color: #666;
  font-weight: bold;
}

input {
  padding: 10px;
  width: 100%;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  box-sizing: border-box;
}

input:focus {
  outline: none;
  border-color: #4CAF50;
  box-shadow: 0 0 0 2px rgba(76, 175, 80, 0.2);
}

button {
  background-color: #4CAF50;
  color: white;
  padding: 12px;
  width: 100%;
  border: none;
  border-radius: 4px;
  font-size: 16px;
  cursor: pointer;
  transition: background-color 0.3s;
}

button:hover {
  background-color: #45a049;
}

button:disabled {
  background-color: #cccccc;
  cursor: not-allowed;
}

.error {
  color: #f44336;
  margin: 15px 0;
  padding: 10px;
  background-color: #ffebee;
  border: 1px solid #ffcdd2;
  border-radius: 4px;
}

.register-link {
  text-align: center;
  margin-top: 20px;
  color: #666;
}

.register-link a {
  color: #4CAF50;
  text-decoration: none;
  font-weight: bold;
}

.register-link a:hover {
  text-decoration: underline;
}

.demo-info {
  margin-top: 30px;
  padding: 15px;
  background-color: #e8f5e9;
  border: 1px solid #c8e6c9;
  border-radius: 4px;
}

.demo-info p {
  margin: 5px 0;
  color: #2e7d32;
  font-size: 14px;
}

.demo-info p:first-child {
  font-weight: bold;
}
</style>
<template>
  <div class="register-container">
    <h2>用户注册</h2>
    <div v-if="error" class="error">{{ error }}</div>
    <div v-if="success" class="success">{{ success }}</div>
    <form @submit.prevent="register">
      <div>
        <label>用户名:</label>
        <input
            v-model="username"
            type="text"
            placeholder="3-50个字符"
            minlength="3"
            maxlength="50"
            required
        />
      </div>
      <div>
        <label>密码:</label>
        <input
            v-model="password"
            type="password"
            placeholder="至少6个字符"
            minlength="6"
            required
        />
      </div>
      <div>
        <label>确认密码:</label>
        <input
            v-model="confirmPassword"
            type="password"
            placeholder="请再次输入密码"
            required
        />
      </div>
      <button type="submit" :disabled="isLoading">
        {{ isLoading ? '注册中...' : '注册' }}
      </button>
    </form>
    <div class="login-link">
      已有账号？<router-link to="/login">立即登录</router-link>
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
      confirmPassword: '',
      error: '',
      success: '',
      isLoading: false,
    };
  },
  methods: {
    async register() {
      // 清除之前的消息
      this.error = '';
      this.success = '';

      // 验证密码是否一致
      if (this.password !== this.confirmPassword) {
        this.error = '两次输入的密码不一致';
        return;
      }

      // 验证用户名长度
      if (this.username.length < 3 || this.username.length > 50) {
        this.error = '用户名长度必须在3-50个字符之间';
        return;
      }

      // 验证密码长度
      if (this.password.length < 6) {
        this.error = '密码长度至少6个字符';
        return;
      }

      this.isLoading = true;
      try {
        const response = await axios.post('/register', {
          username: this.username,
          password: this.password,
        });

        this.success = response.data.message || '注册成功，请等待管理员审核后方可登录';

        // 清空表单
        this.username = '';
        this.password = '';
        this.confirmPassword = '';

        // 3秒后跳转到登录页
        setTimeout(() => {
          this.$router.push('/login');
        }, 3000);
      } catch (err) {
        this.error = err.response?.data?.message || '注册失败，请稍后重试';
      } finally {
        this.isLoading = false;
      }
    },
  },
};
</script>

<style scoped>
.register-container {
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

.success {
  color: #4CAF50;
  margin: 15px 0;
  padding: 10px;
  background-color: #e8f5e9;
  border: 1px solid #c8e6c9;
  border-radius: 4px;
}

.login-link {
  text-align: center;
  margin-top: 20px;
  color: #666;
}

.login-link a {
  color: #4CAF50;
  text-decoration: none;
  font-weight: bold;
}

.login-link a:hover {
  text-decoration: underline;
}
</style>
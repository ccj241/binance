<template>
  <div class="login-page">
    <div class="login-container">
      <!-- Logo区域 -->
      <div class="login-header">
        <div class="logo">
          <span class="logo-icon">📈</span>
          <h1 class="logo-text">xx交易系统xx</h1>
        </div>
        <p class="welcome-text">欢迎回来，请登录您的账户</p>
      </div>

      <!-- 登录表单 -->
      <div class="login-card">
        <form @submit.prevent="login" class="login-form">
          <!-- 错误提示 -->
          <transition name="fade">
            <div v-if="error" class="alert alert-error">
              <span class="alert-icon">⚠️</span>
              <span>{{ error }}</span>
            </div>
          </transition>

          <!-- 用户名输入 -->
          <div class="form-group">
            <label for="username" class="form-label">用户名</label>
            <div class="input-wrapper">
              <span class="input-icon">👤</span>
              <input
                  id="username"
                  v-model="username"
                  type="text"
                  class="form-input"
                  placeholder="请输入用户名"
                  required
                  :disabled="isLoading"
              />
            </div>
          </div>

          <!-- 密码输入 -->
          <div class="form-group">
            <label for="password" class="form-label">密码</label>
            <div class="input-wrapper">
              <span class="input-icon">🔒</span>
              <input
                  id="password"
                  v-model="password"
                  :type="showPassword ? 'text' : 'password'"
                  class="form-input"
                  placeholder="请输入密码"
                  required
                  :disabled="isLoading"
              />
              <button
                  type="button"
                  @click="showPassword = !showPassword"
                  class="toggle-password"
                  tabindex="-1"
              >
                <span>{{ showPassword ? '🙈' : '👁️' }}</span>
              </button>
            </div>
          </div>

          <!-- 记住我 -->
          <div class="form-options">
            <label class="checkbox-wrapper">
              <input
                  v-model="rememberMe"
                  type="checkbox"
                  class="checkbox-input"
              />
              <span class="checkbox-label">记住我</span>
            </label>
            <a href="#" class="forgot-link">忘记密码？</a>
          </div>

          <!-- 登录按钮 -->
          <button
              type="submit"
              class="submit-btn"
              :disabled="isLoading"
          >
            <span v-if="!isLoading" class="btn-content">
              <span>登录</span>
              <span class="btn-icon">→</span>
            </span>
            <span v-else class="btn-loading">
              <span class="spinner"></span>
              <span>登录中...</span>
            </span>
          </button>
        </form>

        <!-- 注册链接 -->
        <div class="register-link">
          <span>还没有账号？</span>
          <router-link to="/register" class="link">立即注册</router-link>
        </div>
      </div>

      <!-- 页脚信息 -->
      <div class="login-footer">
        <p>© 2024 交易系统. All rights reserved.</p>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'Login',
  data() {
    return {
      username: '',
      password: '',
      error: '',
      isLoading: false,
      showPassword: false,
      rememberMe: false,
    };
  },
  mounted() {
    // 检查是否有记住的用户名
    const savedUsername = localStorage.getItem('rememberedUsername');
    if (savedUsername) {
      this.username = savedUsername;
      this.rememberMe = true;
    }
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

        // 添加详细的调试日志
        console.log('完整响应对象:', response);
        console.log('响应状态:', response.status);
        console.log('响应头:', response.headers);
        console.log('响应数据:', response.data);
        console.log('响应数据类型:', typeof response.data);

        // 检查响应数据结构
        const responseData = response.data;

        // 尝试不同的数据结构
        let token = null;
        let role = null;

        // 情况1: 直接返回 {token: "xxx", role: "xxx"}
        if (responseData.token) {
          token = responseData.token;
          role = responseData.role;
        }
        // 情况2: 包装在 data 字段中 {data: {token: "xxx", role: "xxx"}}
        else if (responseData.data && responseData.data.token) {
          token = responseData.data.token;
          role = responseData.data.role;
        }
        // 情况3: 字符串响应（需要解析）
        else if (typeof responseData === 'string') {
          try {
            const parsed = JSON.parse(responseData);
            token = parsed.token;
            role = parsed.role;
          } catch (e) {
            console.error('解析字符串响应失败:', e);
          }
        }

        console.log('提取的 Token:', token);
        console.log('提取的 Role:', role);

        // 严格验证 token 存在且有效
        if (!token || token === 'undefined' || token === 'null' || token === '') {
          console.error('无法从响应中提取有效的 token');
          console.error('完整响应数据:', JSON.stringify(responseData, null, 2));
          this.error = '登录失败：服务器未返回有效的认证信息';
          return;
        }

        // 验证 token 格式（JWT 应该有3个部分）
        const tokenParts = token.split('.');
        if (tokenParts.length !== 3) {
          console.error('Token 格式错误，应该包含3个部分，实际:', tokenParts.length);
          this.error = '登录失败：服务器返回的认证信息格式错误';
          return;
        }

        // 保存token
        localStorage.setItem('token', token);
        console.log('Token 已保存到 localStorage');

        // 记住用户名
        if (this.rememberMe) {
          localStorage.setItem('rememberedUsername', this.username);
        } else {
          localStorage.removeItem('rememberedUsername');
        }

        // 根据角色跳转
        if (role === 'admin') {
          this.$router.push('/admin');
        } else {
          // 从查询参数获取重定向地址
          const redirect = this.$route.query.redirect || '/';
          this.$router.push(redirect);
        }
      } catch (err) {
        console.error('登录错误:', err);
        console.log('错误响应:', err.response);
        console.log('错误响应数据:', err.response?.data);

        if (err.response?.status === 403) {
          this.error = err.response.data.error || err.response.data.message || '账号状态异常，请联系管理员';
        } else if (err.response?.status === 401) {
          this.error = '用户名或密码错误';
        } else {
          this.error = err.response?.data?.error || err.response?.data?.message || '登录失败，请稍后重试';
        }
      } finally {
        this.isLoading = false;
      }
    },
  },
};
</script>

<style scoped>
/* 页面容器 */
.login-page {
  min-height: 100vh;
  background-color: var(--color-bg-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.login-container {
  width: 100%;
  max-width: 420px;
}

/* Logo区域 */
.login-header {
  text-align: center;
  margin-bottom: 2rem;
}

.logo {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  margin-bottom: 0.5rem;
}

.logo-icon {
  font-size: 2.5rem;
}

.logo-text {
  font-size: 1.875rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
}

.welcome-text {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}

/* 登录卡片 */
.login-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 2rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

/* 表单样式 */
.login-form {
  margin-bottom: 1.5rem;
}

/* 错误提示 */
.alert {
  padding: 0.75rem 1rem;
  border-radius: var(--radius-md);
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.alert-error {
  background-color: #fef2f2;
  border: 1px solid #fecaca;
  color: #dc2626;
}

.alert-icon {
  font-size: 1rem;
}

/* 表单组 */
.form-group {
  margin-bottom: 1.25rem;
}

.form-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
  margin-bottom: 0.5rem;
}

/* 输入框容器 */
.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 0.875rem;
  font-size: 1.125rem;
  color: var(--color-text-tertiary);
  pointer-events: none;
}

.form-input {
  width: 100%;
  padding: 0.75rem 1rem 0.75rem 2.75rem;
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-primary);
  font-size: 0.875rem;
  transition: all var(--transition-normal);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

.form-input:disabled {
  background-color: var(--color-bg-tertiary);
  cursor: not-allowed;
}

.form-input::placeholder {
  color: var(--color-text-tertiary);
}

/* 密码切换按钮 */
.toggle-password {
  position: absolute;
  right: 0.875rem;
  background: none;
  border: none;
  color: var(--color-text-tertiary);
  cursor: pointer;
  padding: 0.25rem;
  font-size: 1.125rem;
  transition: color var(--transition-fast);
}

.toggle-password:hover {
  color: var(--color-text-secondary);
}

/* 表单选项 */
.form-options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.checkbox-wrapper {
  display: flex;
  align-items: center;
  cursor: pointer;
}

.checkbox-input {
  width: 16px;
  height: 16px;
  margin-right: 0.5rem;
  cursor: pointer;
}

.checkbox-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  user-select: none;
}

.forgot-link {
  font-size: 0.875rem;
  color: var(--color-primary);
  text-decoration: none;
  transition: color var(--transition-fast);
}

.forgot-link:hover {
  color: var(--color-primary-hover);
  text-decoration: underline;
}

/* 提交按钮 */
.submit-btn {
  width: 100%;
  padding: 0.875rem;
  background-color: var(--color-primary);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
}

.submit-btn:hover:not(:disabled) {
  background-color: var(--color-primary-hover);
}

.submit-btn:disabled {
  background-color: var(--color-secondary);
  cursor: not-allowed;
}

.btn-content {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.btn-icon {
  font-size: 1.125rem;
}

.btn-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 注册链接 */
.register-link {
  text-align: center;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.link {
  color: var(--color-primary);
  text-decoration: none;
  font-weight: 500;
  transition: color var(--transition-fast);
}

.link:hover {
  color: var(--color-primary-hover);
  text-decoration: underline;
}

/* 页脚 */
.login-footer {
  text-align: center;
  margin-top: 2rem;
}

.login-footer p {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
  margin: 0;
}

/* 动画 */
.fade-enter-active,
.fade-leave-active {
  transition: all 0.3s ease;
}

.fade-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* 响应式设计 */
@media (max-width: 480px) {
  .login-card {
    padding: 1.5rem;
  }
}
</style>
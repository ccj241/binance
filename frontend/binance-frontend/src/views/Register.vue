<template>
  <div class="register-page">
    <div class="register-container">
      <!-- Logo区域 -->
      <div class="register-header">
        <div class="logo">
          <span class="logo-icon">📈</span>
          <h1 class="logo-text">交易系统</h1>
        </div>
        <p class="welcome-text">创建账户，开启您的交易之旅</p>
      </div>

      <!-- 注册表单 -->
      <div class="register-card">
        <form @submit.prevent="register" class="register-form">
          <!-- 错误提示 -->
          <transition name="fade">
            <div v-if="error" class="alert alert-error">
              <span class="alert-icon">⚠️</span>
              <span>{{ error }}</span>
            </div>
          </transition>

          <!-- 成功提示 -->
          <transition name="fade">
            <div v-if="success" class="alert alert-success">
              <span class="alert-icon">✅</span>
              <span>{{ success }}</span>
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
                  placeholder="3-50个字符"
                  minlength="3"
                  maxlength="50"
                  required
                  :disabled="isLoading"
                  @input="validateUsername"
              />
              <transition name="fade">
                <span v-if="usernameError" class="field-error">
                  {{ usernameError }}
                </span>
              </transition>
            </div>
            <p class="field-hint">用户名将作为您的登录凭证</p>
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
                  placeholder="至少6个字符"
                  minlength="6"
                  required
                  :disabled="isLoading"
                  @input="validatePassword"
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
            <!-- 密码强度指示器 -->
            <div v-if="password" class="password-strength">
              <div class="strength-bars">
                <div
                    class="strength-bar"
                    :class="{ active: passwordStrength >= 1 }"
                ></div>
                <div
                    class="strength-bar"
                    :class="{ active: passwordStrength >= 2 }"
                ></div>
                <div
                    class="strength-bar"
                    :class="{ active: passwordStrength >= 3 }"
                ></div>
                <div
                    class="strength-bar"
                    :class="{ active: passwordStrength >= 4 }"
                ></div>
              </div>
              <span class="strength-text">{{ passwordStrengthText }}</span>
            </div>
          </div>

          <!-- 确认密码输入 -->
          <div class="form-group">
            <label for="confirmPassword" class="form-label">确认密码</label>
            <div class="input-wrapper">
              <span class="input-icon">🔒</span>
              <input
                  id="confirmPassword"
                  v-model="confirmPassword"
                  :type="showConfirmPassword ? 'text' : 'password'"
                  class="form-input"
                  placeholder="请再次输入密码"
                  required
                  :disabled="isLoading"
                  @input="validatePasswordMatch"
              />
              <button
                  type="button"
                  @click="showConfirmPassword = !showConfirmPassword"
                  class="toggle-password"
                  tabindex="-1"
              >
                <span>{{ showConfirmPassword ? '🙈' : '👁️' }}</span>
              </button>
              <transition name="fade">
                <span v-if="passwordMatchError" class="field-error">
                  {{ passwordMatchError }}
                </span>
              </transition>
            </div>
          </div>

          <!-- 服务条款 -->
          <div class="form-group">
            <label class="checkbox-wrapper">
              <input
                  v-model="agreeTerms"
                  type="checkbox"
                  class="checkbox-input"
                  required
              />
              <span class="checkbox-label">
                我已阅读并同意
                <a href="#" class="link">服务条款</a>
                和
                <a href="#" class="link">隐私政策</a>
              </span>
            </label>
          </div>

          <!-- 注册按钮 -->
          <button
              type="submit"
              class="submit-btn"
              :disabled="isLoading || !isFormValid"
          >
            <span v-if="!isLoading" class="btn-content">
              <span>创建账户</span>
              <span class="btn-icon">→</span>
            </span>
            <span v-else class="btn-loading">
              <span class="spinner"></span>
              <span>注册中...</span>
            </span>
          </button>
        </form>

        <!-- 登录链接 -->
        <div class="login-link">
          <span>已有账号？</span>
          <router-link to="/login" class="link">立即登录</router-link>
        </div>

        <!-- 提示信息 -->
        <div class="info-box">
          <div class="info-icon">ℹ️</div>
          <div class="info-content">
            <p class="info-title">注册须知</p>
            <ul class="info-list">
              <li>注册成功后需要等待管理员审核</li>
              <li>审核通过后方可正常登录使用</li>
              <li>请确保填写真实有效的信息</li>
            </ul>
          </div>
        </div>
      </div>

      <!-- 页脚信息 -->
      <div class="register-footer">
        <p>© 2024 交易系统. All rights reserved.</p>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'Register',
  data() {
    return {
      username: '',
      password: '',
      confirmPassword: '',
      error: '',
      success: '',
      isLoading: false,
      showPassword: false,
      showConfirmPassword: false,
      agreeTerms: false,
      usernameError: '',
      passwordMatchError: '',
      redirectTimer: null,
    };
  },
  computed: {
    passwordStrength() {
      if (!this.password) return 0;

      let strength = 0;

      // 长度检查
      if (this.password.length >= 8) strength++;
      if (this.password.length >= 12) strength++;

      // 包含数字
      if (/\d/.test(this.password)) strength++;

      // 包含大小写字母和特殊字符
      if (/[a-z]/.test(this.password) &&
          /[A-Z]/.test(this.password) &&
          /[^a-zA-Z0-9]/.test(this.password)) {
        strength++;
      }

      return strength;
    },

    passwordStrengthText() {
      const strengthMap = {
        0: '太弱',
        1: '弱',
        2: '中等',
        3: '强',
        4: '很强'
      };
      return strengthMap[this.passwordStrength];
    },

    isFormValid() {
      return this.username.length >= 3 &&
          this.password.length >= 6 &&
          this.password === this.confirmPassword &&
          this.agreeTerms &&
          !this.usernameError &&
          !this.passwordMatchError;
    }
  },

  beforeUnmount() {
    if (this.redirectTimer) {
      clearTimeout(this.redirectTimer);
    }
  },

  methods: {
    validateUsername() {
      if (this.username.length > 0 && this.username.length < 3) {
        this.usernameError = '用户名至少需要3个字符';
      } else if (this.username.length > 50) {
        this.usernameError = '用户名不能超过50个字符';
      } else {
        this.usernameError = '';
      }
    },

    validatePassword() {
      // 可以在这里添加更多密码验证逻辑
      if (this.confirmPassword && this.password !== this.confirmPassword) {
        this.passwordMatchError = '两次输入的密码不一致';
      } else {
        this.passwordMatchError = '';
      }
    },

    validatePasswordMatch() {
      if (this.confirmPassword && this.password !== this.confirmPassword) {
        this.passwordMatchError = '两次输入的密码不一致';
      } else {
        this.passwordMatchError = '';
      }
    },

    async register() {
      // 清除之前的消息
      this.error = '';
      this.success = '';

      // 最终验证
      if (this.password !== this.confirmPassword) {
        this.error = '两次输入的密码不一致';
        return;
      }

      if (this.username.length < 3 || this.username.length > 50) {
        this.error = '用户名长度必须在3-50个字符之间';
        return;
      }

      if (this.password.length < 6) {
        this.error = '密码长度至少6个字符';
        return;
      }

      if (!this.agreeTerms) {
        this.error = '请先同意服务条款';
        return;
      }

      this.isLoading = true;

      try {
        const response = await axios.post('/register', {
          username: this.username,
          password: this.password,
        });

        this.success = response.data.message || '注册成功！请等待管理员审核后登录';

        // 清空表单
        this.username = '';
        this.password = '';
        this.confirmPassword = '';
        this.agreeTerms = false;

        // 3秒后跳转到登录页
        this.redirectTimer = setTimeout(() => {
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
/* 页面容器 */
.register-page {
  min-height: 100vh;
  background-color: var(--color-bg-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1rem;
}

.register-container {
  width: 100%;
  max-width: 460px;
}

/* Logo区域 */
.register-header {
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

/* 注册卡片 */
.register-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 2rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
}

/* 表单样式 */
.register-form {
  margin-bottom: 1.5rem;
}

/* 提示框 */
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

.alert-success {
  background-color: #f0fdf4;
  border: 1px solid #bbf7d0;
  color: #16a34a;
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
  z-index: 1;
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

/* 字段提示 */
.field-hint {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
  margin-top: 0.25rem;
}

.field-error {
  position: absolute;
  bottom: -1.25rem;
  left: 0;
  font-size: 0.75rem;
  color: var(--color-danger);
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
  z-index: 1;
}

.toggle-password:hover {
  color: var(--color-text-secondary);
}

/* 密码强度指示器 */
.password-strength {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 0.5rem;
}

.strength-bars {
  display: flex;
  gap: 0.25rem;
  flex: 1;
}

.strength-bar {
  height: 4px;
  flex: 1;
  background-color: var(--color-bg-tertiary);
  border-radius: 2px;
  transition: background-color var(--transition-normal);
}

.strength-bar.active:nth-child(1) {
  background-color: var(--color-danger);
}

.strength-bar.active:nth-child(2) {
  background-color: var(--color-warning);
}

.strength-bar.active:nth-child(3) {
  background-color: var(--color-warning);
}

.strength-bar.active:nth-child(4) {
  background-color: var(--color-success);
}

.strength-text {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  white-space: nowrap;
}

/* 复选框 */
.checkbox-wrapper {
  display: flex;
  align-items: flex-start;
  cursor: pointer;
  gap: 0.5rem;
}

.checkbox-input {
  width: 16px;
  height: 16px;
  margin-top: 0.125rem;
  cursor: pointer;
  flex-shrink: 0;
}

.checkbox-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  user-select: none;
  line-height: 1.5;
}

/* 链接样式 */
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
  margin-top: 1.5rem;
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

/* 登录链接 */
.login-link {
  text-align: center;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  margin-bottom: 1.5rem;
}

/* 信息框 */
.info-box {
  display: flex;
  gap: 0.75rem;
  padding: 1rem;
  background-color: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
}

.info-icon {
  font-size: 1.25rem;
  flex-shrink: 0;
}

.info-content {
  flex: 1;
}

.info-title {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
  margin: 0 0 0.5rem 0;
}

.info-list {
  margin: 0;
  padding-left: 1.25rem;
}

.info-list li {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

/* 页脚 */
.register-footer {
  text-align: center;
  margin-top: 2rem;
}

.register-footer p {
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
  .register-card {
    padding: 1.5rem;
  }

  .info-box {
    flex-direction: column;
    text-align: center;
  }
}
</style>
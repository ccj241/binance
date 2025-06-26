<template>
  <div class="settings-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">系统设置</h1>
      <p class="page-subtitle">管理您的API密钥和自动提币规则</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">
          <span>🔑</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ apiKey ? '已配置' : '未配置' }}</div>
          <div class="stat-label">API 密钥</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">
          <span>🔒</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ secretKey ? '已配置' : '未配置' }}</div>
          <div class="stat-label">Secret 密钥</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">
          <span>⚡</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ withdrawalRules.length }}</div>
          <div class="stat-label">提币规则</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">
          <span>✅</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ enabledRulesCount }}</div>
          <div class="stat-label">启用规则</div>
        </div>
      </div>
    </div>

    <!-- 消息提示 -->
    <transition name="toast">
      <div v-if="toastMessage" :class="['toast', toastType]">
        <span class="toast-icon">{{ toastType === 'success' ? '✓' : '×' }}</span>
        <span>{{ toastMessage }}</span>
      </div>
    </transition>

    <!-- API 密钥管理 -->
    <div class="settings-section">
      <div class="section-header">
        <h2 class="section-title">
          <span class="section-icon">🔑</span>
          API 密钥管理
        </h2>
        <button @click="toggleApiSection" class="toggle-btn">
          <span>{{ showApiSection ? '收起' : '展开' }}</span>
        </button>
      </div>

      <transition name="section-slide">
        <div v-if="showApiSection" class="section-content">
          <!-- 当前密钥状态 -->
          <div v-if="apiKey || secretKey" class="current-keys">
            <div class="key-display">
              <div class="key-card">
                <div class="key-header">
                  <div class="key-icon">🔑</div>
                  <div class="key-info">
                    <h4>API Key</h4>
                    <p class="key-status">已配置</p>
                  </div>
                </div>
                <div class="key-value">
                  <span class="masked-key">{{ maskKey(apiKey) }}</span>
                  <button @click="toggleKeyVisibility('api')" class="visibility-btn">
                    <span>{{ showApiKey ? '隐藏' : '显示' }}</span>
                  </button>
                </div>
                <div v-if="showApiKey" class="full-key">{{ apiKey }}</div>
              </div>

              <div class="key-card">
                <div class="key-header">
                  <div class="key-icon">🔒</div>
                  <div class="key-info">
                    <h4>Secret Key</h4>
                    <p class="key-status">已配置</p>
                  </div>
                </div>
                <div class="key-value">
                  <span class="masked-key">{{ maskKey(secretKey) }}</span>
                  <button @click="toggleKeyVisibility('secret')" class="visibility-btn">
                    <span>{{ showSecretKey ? '隐藏' : '显示' }}</span>
                  </button>
                </div>
                <div v-if="showSecretKey" class="full-key">{{ secretKey }}</div>
              </div>
            </div>

            <button @click="deleteAPIKey" class="action-btn delete">
              删除 API 密钥
            </button>
          </div>

          <div v-else class="no-keys">
            <div class="no-keys-icon">🔑</div>
            <p class="no-keys-text">尚未配置 API 密钥</p>
            <p class="no-keys-subtitle">请添加您的 Binance API 密钥以开始使用</p>
          </div>

          <!-- 添加新密钥 -->
          <div class="add-keys-section">
            <h3 class="subsection-title">添加新的 API 密钥</h3>

            <form @submit.prevent="saveAPIKey" class="key-form">
              <div class="form-grid">
                <div class="form-group">
                  <label>API Key</label>
                  <input
                      v-model="newAPIKey"
                      type="text"
                      placeholder="请输入您的 Binance API Key"
                      required
                  />
                </div>

                <div class="form-group">
                  <label>Secret Key</label>
                  <div class="password-input">
                    <input
                        v-model="newSecretKey"
                        :type="showNewSecretInput ? 'text' : 'password'"
                        placeholder="请输入您的 Secret Key"
                        required
                    />
                    <button
                        type="button"
                        @click="showNewSecretInput = !showNewSecretInput"
                        class="password-toggle"
                    >
                      {{ showNewSecretInput ? '隐藏' : '显示' }}
                    </button>
                  </div>
                </div>
              </div>

              <div class="form-actions">
                <button type="submit" class="action-btn primary">
                  保存 API 密钥
                </button>
                <button type="button" @click="resetApiForm" class="action-btn secondary">
                  重置表单
                </button>
              </div>
            </form>
          </div>
        </div>
      </transition>
    </div>

    <!-- 自动提币设置 -->
    <div class="settings-section">
      <div class="section-header">
        <h2 class="section-title">
          <span class="section-icon">⚡</span>
          自动提币设置
        </h2>
        <button @click="toggleWithdrawalSection" class="toggle-btn">
          <span>{{ showWithdrawalSection ? '收起' : '展开' }}</span>
        </button>
      </div>

      <transition name="section-slide">
        <div v-if="showWithdrawalSection" class="section-content">
          <!-- 添加提币规则 -->
          <div class="add-rule-section">
            <h3 class="subsection-title">添加提币规则</h3>

            <form @submit.prevent="createWithdrawalRule" class="rule-form">
              <div class="form-grid">
                <div class="form-group">
                  <label>币种</label>
                  <input
                      v-model="newWithdrawal.asset"
                      type="text"
                      placeholder="例如: BTC, ETH, USDT"
                      required
                  />
                </div>

                <div class="form-group">
                  <label>触发阈值</label>
                  <input
                      v-model.number="newWithdrawal.threshold"
                      type="number"
                      step="0.00000001"
                      placeholder="余额超过此数量时触发"
                      required
                  />
                </div>

                <div class="form-group">
                  <label>提币金额</label>
                  <input
                      v-model.number="newWithdrawal.amount"
                      type="number"
                      step="0.00000001"
                      min="0"
                      placeholder="每次提币数量（0表示提取最大可用金额）"
                      required
                  />
                  <small class="form-hint">设置为0将自动提取所有可用余额</small>
                </div>

                <div class="form-group">
                  <label>提币地址</label>
                  <input
                      v-model="newWithdrawal.address"
                      type="text"
                      placeholder="目标钱包地址"
                      required
                  />
                </div>
              </div>

              <!-- 规则说明 -->
              <div class="rule-description">
                <div class="description-card">
                  <div class="description-icon">💡</div>
                  <div class="description-content">
                    <h4>自动提币规则说明</h4>
                    <p>当您的 <strong>{{ newWithdrawal.asset || '[币种]' }}</strong> 余额超过 <strong>{{ newWithdrawal.threshold || '[阈值]' }}</strong> 时，系统将自动提取 <strong>{{ newWithdrawal.amount > 0 ? formatNumber(newWithdrawal.amount) : '最大可用金额' }}</strong> 到指定地址。</p>
                    <small>⚠️ 提示：请确保提币地址正确，提币操作无法撤回。</small>
                  </div>
                </div>
              </div>

              <div class="form-actions">
                <button type="submit" class="action-btn primary">
                  创建规则
                </button>
                <button type="button" @click="resetWithdrawalForm" class="action-btn secondary">
                  重置表单
                </button>
              </div>
            </form>
          </div>

          <!-- 现有规则列表 -->
          <div class="rules-list">
            <h3 class="subsection-title">现有规则 ({{ withdrawalRules.length }})</h3>

            <div v-if="withdrawalRules.length === 0" class="empty-state">
              <div class="empty-icon">⚡</div>
              <p class="empty-text">暂无提币规则</p>
              <p class="empty-subtitle">添加第一个自动提币规则以开始使用</p>
            </div>

            <div v-else class="rules-grid">
              <div v-for="rule in withdrawalRules" :key="rule.id" class="rule-card">
                <div class="rule-header">
                  <div class="rule-asset">
                    <div class="asset-icon">🪙</div>
                    <div class="asset-info">
                      <h4>{{ rule.asset }}</h4>
                      <span :class="['status-chip', rule.enabled ? 'enabled' : 'disabled']">
                        {{ rule.enabled ? '启用' : '禁用' }}
                      </span>
                    </div>
                  </div>
                  <div class="rule-id">ID: {{ rule.id }}</div>
                </div>

                <div class="rule-details">
                  <div class="detail-item">
                    <span class="detail-label">触发阈值</span>
                    <span class="detail-value">{{ formatNumber(rule.threshold) }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">提币金额</span>
                    <span class="detail-value">{{ rule.amount > 0 ? formatNumber(rule.amount) : '最大可用' }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">提币地址</span>
                    <span class="detail-value address">{{ formatAddress(rule.address) }}</span>
                  </div>
                </div>

                <div class="rule-actions">
                  <button @click="toggleRuleStatus(rule)" class="action-btn toggle">
                    {{ rule.enabled ? '禁用' : '启用' }}
                  </button>
                  <button @click="deleteWithdrawalRule(rule.id)" class="action-btn delete">
                    删除
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'Settings',
  data() {
    return {
      apiKey: '',
      secretKey: '',
      newAPIKey: '',
      newSecretKey: '',
      showApiKey: false,
      showSecretKey: false,
      showNewSecretInput: false,
      showApiSection: true,
      showWithdrawalSection: true,
      newWithdrawal: {
        asset: '',
        threshold: 0,
        amount: 0,
        address: '',
      },
      withdrawalRules: [],
      toastMessage: '',
      toastType: 'success'
    };
  },
  computed: {
    enabledRulesCount() {
      return this.withdrawalRules.filter(rule => rule.enabled).length;
    }
  },
  async mounted() {
    await this.fetchAPIKey();
    await this.fetchWithdrawalRules();
  },
  methods: {
    getAuthHeaders() {
      const token = localStorage.getItem('token');
      return {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json'
      };
    },

    showToast(message, type = 'success') {
      this.toastMessage = message;
      this.toastType = type;
      setTimeout(() => {
        this.toastMessage = '';
      }, 3000);
    },

    toggleApiSection() {
      this.showApiSection = !this.showApiSection;
    },

    toggleWithdrawalSection() {
      this.showWithdrawalSection = !this.showWithdrawalSection;
    },

    toggleKeyVisibility(type) {
      if (type === 'api') {
        this.showApiKey = !this.showApiKey;
      } else {
        this.showSecretKey = !this.showSecretKey;
      }
    },

    maskKey(key) {
      if (!key) return '';
      if (key.length <= 8) return '***';
      return key.substring(0, 4) + '***' + key.substring(key.length - 4);
    },

    formatNumber(num) {
      if (!num) return '0';
      return parseFloat(num).toFixed(8).replace(/\.?0+$/, '');
    },

    formatAddress(address) {
      if (!address) return '';
      if (address.length <= 16) return address;
      return address.substring(0, 8) + '...' + address.substring(address.length - 8);
    },

    resetApiForm() {
      this.newAPIKey = '';
      this.newSecretKey = '';
      this.showNewSecretInput = false;
    },

    resetWithdrawalForm() {
      this.newWithdrawal = {
        asset: '',
        threshold: 0,
        amount: 0,
        address: '',
      };
    },

    async fetchAPIKey() {
      try {
        const response = await axios.get('/api-key', {
          headers: this.getAuthHeaders(),
        });
        this.apiKey = response.data.apiKey || '';
        this.secretKey = response.data.secretKey || '';
      } catch (err) {
        console.error('fetchAPIKey error:', err);
        this.showToast(err.response?.data?.error || '获取 API 密钥失败', 'error');
      }
    },

    async saveAPIKey() {
      if (!this.newAPIKey.trim() || !this.newSecretKey.trim()) {
        this.showToast('请填写完整的 API 密钥信息', 'error');
        return;
      }

      try {
        const response = await axios.post(
            '/api-key',
            {
              apiKey: this.newAPIKey,
              apiSecret: this.newSecretKey,  // 修改这里：从 apiSecret 改为 secretKey
            },
            {
              headers: this.getAuthHeaders(),
            }
        );

        this.showToast(response.data.message || 'API 密钥保存成功');
        this.resetApiForm();
        await this.fetchAPIKey();
      } catch (err) {
        console.error('saveAPIKey error:', err);
        this.showToast(err.response?.data?.error || '保存 API 密钥失败', 'error');
      }
    },

    async deleteAPIKey() {
      if (!window.confirm('确定要删除 API 密钥吗？删除后将无法进行交易操作。')) {
        return;
      }

      try {
        const response = await axios.delete('/api-key/delete', {
          headers: this.getAuthHeaders(),
        });
        this.showToast(response.data.message || 'API 密钥删除成功');
        this.apiKey = '';
        this.secretKey = '';
        this.showApiKey = false;
        this.showSecretKey = false;
      } catch (err) {
        console.error('deleteAPIKey error:', err);
        this.showToast(err.response?.data?.error || '删除 API 密钥失败', 'error');
      }
    },

    async createWithdrawalRule() {
      const { asset, threshold, amount, address } = this.newWithdrawal;

      if (!asset.trim() || threshold <= 0 || amount < 0 || !address.trim()) {
        this.showToast('请填写所有必需字段，阈值必须大于0，金额不能为负数', 'error');
        return;
      }

      try {
        const response = await axios.post(
            '/withdrawals',
            {
              asset: asset.toUpperCase(),
              threshold: Number(threshold),
              amount: Number(amount),
              address: address,
              enabled: true,
            },
            {
              headers: this.getAuthHeaders(),
            }
        );
        this.showToast(response.data.message || '自动提币规则创建成功');
        this.resetWithdrawalForm();
        await this.fetchWithdrawalRules();
      } catch (err) {
        console.error('createWithdrawalRule error:', err);
        this.showToast(err.response?.data?.error || '创建提币规则失败', 'error');
      }
    },

    async fetchWithdrawalRules() {
      try {
        const response = await axios.get('/withdrawals', {
          headers: this.getAuthHeaders(),
        });
        this.withdrawalRules = response.data.rules || [];
      } catch (err) {
        console.error('fetchWithdrawalRules error:', err);
        this.showToast(err.response?.data?.error || '获取提币规则失败', 'error');
      }
    },

    async toggleRuleStatus(rule) {
      try {
        const response = await axios.put(
            `/withdrawals/${rule.id}`,
            {
              ...rule,
              enabled: !rule.enabled,
            },
            {
              headers: this.getAuthHeaders(),
            }
        );
        this.showToast(response.data.message || `规则已${!rule.enabled ? '启用' : '禁用'}`);
        await this.fetchWithdrawalRules();
      } catch (err) {
        console.error('toggleRuleStatus error:', err);
        this.showToast(err.response?.data?.error || '更新规则状态失败', 'error');
      }
    },

    async deleteWithdrawalRule(ruleId) {
      if (!window.confirm(`确定要删除提币规则 ID ${ruleId} 吗？`)) {
        return;
      }

      try {
        const response = await axios.delete(`/withdrawals/${ruleId}`, {
          headers: this.getAuthHeaders(),
        });
        this.showToast(response.data.message || '提币规则删除成功');
        await this.fetchWithdrawalRules();
      } catch (err) {
        console.error('deleteWithdrawalRule error:', err);
        this.showToast(err.response?.data?.error || '删除提币规则失败', 'error');
      }
    }
  },
};
</script>

<style scoped>
/* 容器 */
.settings-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0;
  background: #ffffff;
  min-height: 100vh;
}

/* 页面头部 */
.page-header {
  margin-bottom: 2rem;
}

.page-title {
  font-size: 1.875rem;
  font-weight: 600;
  color: #0f172a;
  margin: 0 0 0.5rem 0;
}

.page-subtitle {
  color: #64748b;
  font-size: 0.875rem;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.stat-icon {
  width: 48px;
  height: 48px;
  background: #f8fafc;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 0.25rem;
}

.stat-label {
  color: #64748b;
  font-size: 0.875rem;
}

/* Toast 消息 */
.toast {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1rem 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
  z-index: 1000;
}

.toast.success {
  border-color: #10b981;
}

.toast.error {
  border-color: #ef4444;
}

.toast-icon {
  font-size: 1.25rem;
}

.toast.success .toast-icon {
  color: #10b981;
}

.toast.error .toast-icon {
  color: #ef4444;
}

/* 动画 */
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from {
  transform: translateX(100%);
  opacity: 0;
}

.toast-leave-to {
  transform: translateY(100%);
  opacity: 0;
}

/* 设置区块 */
.settings-section {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  margin-bottom: 1.5rem;
  overflow: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1.125rem;
  font-weight: 600;
  color: #0f172a;
  margin: 0;
}

.section-icon {
  font-size: 1.25rem;
}

.toggle-btn {
  padding: 0.5rem 1rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  color: #475569;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-btn:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.section-content {
  padding: 1.5rem;
}

.section-slide-enter-active,
.section-slide-leave-active {
  transition: all 0.3s ease;
}

.section-slide-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.section-slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* 密钥显示 */
.current-keys {
  margin-bottom: 2rem;
}

.key-display {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.key-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1.25rem;
}

.key-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.key-icon {
  width: 40px;
  height: 40px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.key-info h4 {
  margin: 0 0 0.25rem 0;
  color: #0f172a;
  font-size: 1rem;
}

.key-status {
  margin: 0;
  color: #10b981;
  font-size: 0.875rem;
}

.key-value {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  padding: 0.75rem;
  border-radius: 0.375rem;
  margin-bottom: 0.5rem;
}

.masked-key {
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Mono', 'Courier New', monospace;
  color: #475569;
  font-size: 0.875rem;
}

.visibility-btn {
  padding: 0.25rem 0.75rem;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 0.25rem;
  color: #475569;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.visibility-btn:hover {
  background: #e2e8f0;
}

.full-key {
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Mono', 'Courier New', monospace;
  background: #f0fdf4;
  border: 1px solid #86efac;
  padding: 0.75rem;
  border-radius: 0.375rem;
  color: #16a34a;
  font-size: 0.8125rem;
  word-break: break-all;
}

/* 空状态 */
.no-keys,
.empty-state {
  text-align: center;
  padding: 3rem 2rem;
}

.no-keys-icon,
.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.3;
}

.no-keys-text,
.empty-text {
  color: #475569;
  font-size: 1.125rem;
  margin-bottom: 0.5rem;
}

.no-keys-subtitle,
.empty-subtitle {
  color: #94a3b8;
  font-size: 0.875rem;
}

/* 添加区域 */
.add-keys-section,
.add-rule-section {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.subsection-title {
  font-size: 1rem;
  font-weight: 600;
  color: #0f172a;
  margin: 0 0 1.5rem 0;
}

/* 表单 */
.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-weight: 500;
  color: #475569;
  font-size: 0.875rem;
}

.form-group input {
  padding: 0.625rem 0.875rem;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  color: #0f172a;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

.form-group input::placeholder {
  color: #94a3b8;
}

.password-input {
  position: relative;
}

.password-input input {
  width: 100%;
  padding-right: 4rem;
}

.password-toggle {
  position: absolute;
  right: 0.5rem;
  top: 50%;
  transform: translateY(-50%);
  padding: 0.25rem 0.75rem;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 0.25rem;
  color: #475569;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.password-toggle:hover {
  background: #e2e8f0;
}

.form-hint {
  color: #94a3b8;
  font-size: 0.75rem;
  margin-top: 0.25rem;
}

/* 规则说明 */
.rule-description {
  margin: 1.5rem 0;
}

.description-card {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  background: #fffbeb;
  border: 1px solid #fbbf24;
  border-radius: 0.5rem;
  padding: 1rem;
}

.description-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.description-content h4 {
  margin: 0 0 0.5rem 0;
  color: #0f172a;
  font-size: 0.875rem;
}

.description-content p {
  margin: 0 0 0.5rem 0;
  color: #475569;
  font-size: 0.875rem;
  line-height: 1.5;
}

.description-content small {
  color: #92400e;
  font-size: 0.75rem;
}

/* 操作按钮 */
.form-actions {
  display: flex;
  gap: 0.75rem;
}

.action-btn {
  padding: 0.625rem 1.25rem;
  border: 1px solid transparent;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  flex: 1;
}

.action-btn.primary {
  background: #2563eb;
  color: white;
}

.action-btn.primary:hover {
  background: #1d4ed8;
}

.action-btn.secondary {
  background: #ffffff;
  color: #475569;
  border-color: #e2e8f0;
}

.action-btn.secondary:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.action-btn.delete {
  background: #ffffff;
  color: #ef4444;
  border-color: #fecaca;
}

.action-btn.delete:hover {
  background: #fef2f2;
  border-color: #fca5a5;
}

.action-btn.toggle {
  background: #ffffff;
  color: #3b82f6;
  border-color: #bfdbfe;
}

.action-btn.toggle:hover {
  background: #eff6ff;
  border-color: #93c5fd;
}

/* 规则列表 */
.rules-list {
  margin-top: 2rem;
}

.rules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}

.rule-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1.5rem;
  transition: all 0.2s;
}

.rule-card:hover {
  border-color: #cbd5e1;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #f1f5f9;
}

.rule-asset {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.asset-icon {
  width: 40px;
  height: 40px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.asset-info h4 {
  margin: 0 0 0.25rem 0;
  color: #0f172a;
  font-size: 1.125rem;
  font-weight: 600;
}

.rule-id {
  color: #94a3b8;
  font-size: 0.75rem;
}

/* 状态标签 */
.status-chip {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-chip.enabled {
  background: #d1fae5;
  color: #065f46;
}

.status-chip.disabled {
  background: #f3f4f6;
  color: #6b7280;
}

/* 规则详情 */
.rule-details {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-label {
  color: #64748b;
  font-size: 0.875rem;
}

.detail-value {
  color: #0f172a;
  font-size: 0.875rem;
  font-weight: 500;
}

.detail-value.address {
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Mono', 'Courier New', monospace;
  color: #2563eb;
}

/* 规则操作 */
.rule-actions {
  display: flex;
  gap: 0.5rem;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr 1fr;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .key-display {
    grid-template-columns: 1fr;
  }

  .rules-grid {
    grid-template-columns: 1fr;
  }

  .form-actions {
    flex-direction: column;
  }

  .rule-actions {
    flex-direction: column;
  }

  .action-btn {
    width: 100%;
  }

  .toast {
    left: 1rem;
    right: 1rem;
    bottom: 1rem;
  }
}
</style>
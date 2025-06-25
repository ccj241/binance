<template>
  <div class="settings-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">
        <span class="gradient-text">系统设置</span>
      </h1>
      <p class="page-subtitle">管理您的API密钥和自动提币规则</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%)">
          <i>🔑</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ apiKey ? '已配置' : '未配置' }}</div>
          <div class="stat-label">API 密钥</div>
        </div>
        <div class="stat-bg"></div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)">
          <i>🔒</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ secretKey ? '已配置' : '未配置' }}</div>
          <div class="stat-label">Secret 密钥</div>
        </div>
        <div class="stat-bg"></div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #fa709a 0%, #fee140 100%)">
          <i>⚡</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ withdrawalRules.length }}</div>
          <div class="stat-label">提币规则</div>
        </div>
        <div class="stat-bg"></div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%)">
          <i>✅</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ enabledRulesCount }}</div>
          <div class="stat-label">启用规则</div>
        </div>
        <div class="stat-bg"></div>
      </div>
    </div>

    <!-- 消息提示 -->
    <transition name="toast">
      <div v-if="toastMessage" :class="['toast', toastType]">
        <i class="toast-icon">{{ toastType === 'success' ? '✅' : '❌' }}</i>
        <span>{{ toastMessage }}</span>
      </div>
    </transition>

    <!-- API 密钥管理 -->
    <div class="settings-section">
      <div class="section-header">
        <h2 class="section-title">
          <i class="section-icon">🔑</i>
          API 密钥管理
        </h2>
        <button @click="toggleApiSection" class="toggle-btn">
          <i>{{ showApiSection ? '🔽' : '▶️' }}</i>
          {{ showApiSection ? '收起' : '展开' }}
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
                    <i>{{ showApiKey ? '🙈' : '👁️' }}</i>
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
                    <i>{{ showSecretKey ? '🙈' : '👁️' }}</i>
                  </button>
                </div>
                <div v-if="showSecretKey" class="full-key">{{ secretKey }}</div>
              </div>
            </div>

            <button @click="deleteAPIKey" class="action-btn delete">
              <i>🗑️</i> 删除 API 密钥
            </button>
          </div>

          <div v-else class="no-keys">
            <div class="no-keys-icon">🔑</div>
            <p class="no-keys-text">尚未配置 API 密钥</p>
            <p class="no-keys-subtitle">请添加您的 Binance API 密钥以开始使用</p>
          </div>

          <!-- 添加新密钥 -->
          <div class="add-keys-section">
            <h3 class="subsection-title">
              <i>➕</i> 添加新的 API 密钥
            </h3>

            <form @submit.prevent="saveAPIKey" class="key-form">
              <div class="form-grid">
                <div class="form-group">
                  <label>API Key</label>
                  <div class="input-wrapper">
                    <i class="input-icon">🔑</i>
                    <input
                        v-model="newAPIKey"
                        type="text"
                        placeholder="请输入您的 Binance API Key"
                        required
                    />
                  </div>
                </div>

                <div class="form-group">
                  <label>Secret Key</label>
                  <div class="input-wrapper">
                    <i class="input-icon">🔒</i>
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
                      <i>{{ showNewSecretInput ? '🙈' : '👁️' }}</i>
                    </button>
                  </div>
                </div>
              </div>

              <div class="form-actions">
                <button type="submit" class="action-btn save">
                  <i>💾</i> 保存 API 密钥
                </button>
                <button type="button" @click="resetApiForm" class="action-btn secondary">
                  <i>🔄</i> 重置表单
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
          <i class="section-icon">⚡</i>
          自动提币设置
        </h2>
        <button @click="toggleWithdrawalSection" class="toggle-btn">
          <i>{{ showWithdrawalSection ? '🔽' : '▶️' }}</i>
          {{ showWithdrawalSection ? '收起' : '展开' }}
        </button>
      </div>

      <transition name="section-slide">
        <div v-if="showWithdrawalSection" class="section-content">
          <!-- 添加提币规则 -->
          <div class="add-rule-section">
            <h3 class="subsection-title">
              <i>➕</i> 添加提币规则
            </h3>

            <form @submit.prevent="createWithdrawalRule" class="rule-form">
              <div class="form-grid">
                <div class="form-group">
                  <label>币种</label>
                  <div class="input-wrapper">
                    <i class="input-icon">🪙</i>
                    <input
                        v-model="newWithdrawal.asset"
                        type="text"
                        placeholder="例如: BTC, ETH, USDT"
                        required
                    />
                  </div>
                </div>

                <div class="form-group">
                  <label>阈值 (数量)</label>
                  <div class="input-wrapper">
                    <i class="input-icon">📊</i>
                    <input
                        v-model.number="newWithdrawal.threshold"
                        type="number"
                        step="0.00000001"
                        placeholder="触发提币的最小数量"
                        required
                    />
                  </div>
                </div>

                <div class="form-group">
                  <label>提币地址</label>
                  <div class="input-wrapper">
                    <i class="input-icon">🏠</i>
                    <input
                        v-model="newWithdrawal.address"
                        type="text"
                        placeholder="目标钱包地址"
                        required
                    />
                  </div>
                </div>

                <div class="form-group">
                  <label>提币数量</label>
                  <div class="input-wrapper">
                    <i class="input-icon">💰</i>
                    <input
                        v-model.number="newWithdrawal.amount"
                        type="number"
                        step="0.00000001"
                        placeholder="每次提币的数量"
                        required
                    />
                  </div>
                </div>
              </div>

              <div class="form-actions">
                <button type="submit" class="action-btn create">
                  <i>🚀</i> 创建规则
                </button>
                <button type="button" @click="resetWithdrawalForm" class="action-btn secondary">
                  <i>🔄</i> 重置表单
                </button>
              </div>
            </form>
          </div>

          <!-- 现有规则列表 -->
          <div class="rules-list">
            <h3 class="subsection-title">
              <i>📋</i> 现有规则 ({{ withdrawalRules.length }})
            </h3>

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
                        <span class="status-dot"></span>
                        {{ rule.enabled ? '启用' : '禁用' }}
                      </span>
                    </div>
                  </div>
                  <div class="rule-id">ID: {{ rule.id }}</div>
                </div>

                <div class="rule-details">
                  <div class="detail-item">
                    <span class="detail-label">阈值</span>
                    <span class="detail-value">{{ formatNumber(rule.threshold) }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">提币数量</span>
                    <span class="detail-value">{{ formatNumber(rule.amount) }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">提币地址</span>
                    <span class="detail-value address">{{ formatAddress(rule.address) }}</span>
                  </div>
                </div>

                <div class="rule-actions">
                  <button @click="viewRuleDetails(rule)" class="action-btn view">
                    <i>👁️</i> 查看详情
                  </button>
                  <button @click="deleteWithdrawalRule(rule.id)" class="action-btn delete">
                    <i>🗑️</i> 删除规则
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </div>

    <!-- 规则详情弹窗 -->
    <div v-if="showRuleDetails" class="modal-overlay" @click="closeRuleDetails">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>提币规则详情</h3>
          <button @click="closeRuleDetails" class="close-btn">✕</button>
        </div>

        <div class="modal-body">
          <div class="detail-grid">
            <div class="detail-card">
              <div class="detail-label">规则ID</div>
              <div class="detail-value">{{ selectedRule.id }}</div>
            </div>
            <div class="detail-card">
              <div class="detail-label">币种</div>
              <div class="detail-value">{{ selectedRule.asset }}</div>
            </div>
            <div class="detail-card">
              <div class="detail-label">阈值</div>
              <div class="detail-value">{{ formatNumber(selectedRule.threshold) }}</div>
            </div>
            <div class="detail-card">
              <div class="detail-label">提币数量</div>
              <div class="detail-value">{{ formatNumber(selectedRule.amount) }}</div>
            </div>
            <div class="detail-card">
              <div class="detail-label">状态</div>
              <div class="detail-value">
                <span :class="['status-chip', selectedRule.enabled ? 'enabled' : 'disabled']">
                  <span class="status-dot"></span>
                  {{ selectedRule.enabled ? '启用' : '禁用' }}
                </span>
              </div>
            </div>
          </div>

          <div class="address-section">
            <h4>提币地址</h4>
            <div class="address-display">
              <span class="full-address">{{ selectedRule.address }}</span>
              <button @click="copyAddress" class="copy-btn">
                <i>📋</i> 复制地址
              </button>
            </div>
          </div>
        </div>
      </div>
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
        address: '',
        amount: 0,
      },
      withdrawalRules: [],
      toastMessage: '',
      toastType: 'success',
      showRuleDetails: false,
      selectedRule: {}
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
        address: '',
        amount: 0,
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
              apiSecret: this.newSecretKey,
            },
            {
              headers: this.getAuthHeaders(),
            }
        );

        this.showToast(response.data.message || 'API 密钥保存成功 🎉');
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
        this.showToast(response.data.message || 'API 密钥删除成功 🗑️');
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
      const { asset, threshold, address, amount } = this.newWithdrawal;

      if (!asset.trim() || threshold <= 0 || !address.trim() || amount <= 0) {
        this.showToast('请填写所有必需字段，且数量必须大于0', 'error');
        return;
      }

      try {
        const response = await axios.post(
            '/withdrawals',
            {
              asset: asset.toUpperCase(),
              threshold,
              address,
              amount,
              enabled: true,
            },
            {
              headers: this.getAuthHeaders(),
            }
        );
        this.showToast(response.data.message || '提币规则创建成功 🚀');
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

    async deleteWithdrawalRule(ruleId) {
      if (!window.confirm(`确定要删除提币规则 ID ${ruleId} 吗？`)) {
        return;
      }

      try {
        const response = await axios.delete(`/withdrawals/${ruleId}`, {
          headers: this.getAuthHeaders(),
        });
        this.showToast(response.data.message || '提币规则删除成功 🗑️');
        await this.fetchWithdrawalRules();
      } catch (err) {
        console.error('deleteWithdrawalRule error:', err);
        this.showToast(err.response?.data?.error || '删除提币规则失败', 'error');
      }
    },

    viewRuleDetails(rule) {
      this.selectedRule = rule;
      this.showRuleDetails = true;
    },

    closeRuleDetails() {
      this.showRuleDetails = false;
      this.selectedRule = {};
    },

    async copyAddress() {
      try {
        await navigator.clipboard.writeText(this.selectedRule.address);
        this.showToast('地址已复制到剪贴板 📋');
      } catch (err) {
        console.error('复制失败:', err);
        this.showToast('复制失败，请手动复制', 'error');
      }
    }
  },
};
</script>

<style scoped>
/* 全局样式 */
.settings-container {
  min-height: 100vh;
  background: #0f0f0f;
  color: #ffffff;
  padding: 2rem;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
}

/* 页面标题 */
.page-header {
  text-align: center;
  margin-bottom: 3rem;
}

.page-title {
  font-size: 3rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.gradient-text {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.page-subtitle {
  color: #666;
  font-size: 1.1rem;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 3rem;
}

.stat-card {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 2rem;
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-5px);
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  margin-bottom: 1rem;
}

.stat-content {
  position: relative;
  z-index: 1;
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.stat-label {
  color: #999;
  font-size: 0.9rem;
}

.stat-bg {
  position: absolute;
  top: -50%;
  right: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255,255,255,0.03) 0%, transparent 70%);
  transform: rotate(45deg);
}

/* Toast 消息 */
.toast {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  padding: 1rem 1.5rem;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 0.8rem;
  font-weight: 500;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
  z-index: 1000;
}

.toast.success {
  border-color: rgba(34, 197, 94, 0.3);
  background: rgba(34, 197, 94, 0.1);
}

.toast.error {
  border-color: rgba(239, 68, 68, 0.3);
  background: rgba(239, 68, 68, 0.1);
}

.toast-icon {
  font-size: 1.2rem;
}

.toast-enter-active, .toast-leave-active {
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

/* 设置区域 */
.settings-section {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 24px;
  padding: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  margin-bottom: 2rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 0.8rem;
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
}

.section-icon {
  font-size: 1.8rem;
}

.toggle-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1.5rem;
  background: rgba(255, 255, 255, 0.1);
  color: #ccc;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.toggle-btn:hover {
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
}

.section-slide-enter-active, .section-slide-leave-active {
  transition: all 0.3s ease;
}

.section-slide-enter-from {
  opacity: 0;
  transform: translateY(-20px);
}

.section-slide-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

.section-content {
  margin-top: 2rem;
}

/* 当前密钥显示 */
.current-keys {
  margin-bottom: 3rem;
}

.key-display {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.key-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.key-card:hover {
  background: rgba(255, 255, 255, 0.08);
  transform: translateY(-2px);
}

.key-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.key-icon {
  width: 50px;
  height: 50px;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
}

.key-info h4 {
  margin: 0 0 0.25rem 0;
  color: #fff;
  font-size: 1.1rem;
}

.key-status {
  margin: 0;
  color: #22c55e;
  font-size: 0.9rem;
}

.key-value {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(0, 0, 0, 0.2);
  padding: 0.8rem;
  border-radius: 8px;
  margin-bottom: 1rem;
}

.masked-key {
  font-family: 'Courier New', monospace;
  color: #ccc;
  font-size: 0.9rem;
}

.visibility-btn {
  background: none;
  border: none;
  color: #999;
  cursor: pointer;
  padding: 0.2rem;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.visibility-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

.full-key {
  font-family: 'Courier New', monospace;
  background: rgba(34, 197, 94, 0.1);
  border: 1px solid rgba(34, 197, 94, 0.3);
  padding: 0.8rem;
  border-radius: 8px;
  color: #22c55e;
  font-size: 0.85rem;
  word-break: break-all;
}

/* 无密钥状态 */
.no-keys {
  text-align: center;
  padding: 3rem 2rem;
  margin-bottom: 3rem;
}

.no-keys-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.3;
}

.no-keys-text {
  color: #666;
  font-size: 1.2rem;
  margin-bottom: 0.5rem;
}

.no-keys-subtitle {
  color: #999;
  font-size: 1rem;
}

/* 添加密钥区域 */
.add-keys-section, .add-rule-section {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 2rem;
  margin-bottom: 2rem;
}

.subsection-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0 0 1.5rem 0;
  color: #fff;
  font-size: 1.2rem;
  font-weight: 600;
}

/* 表单样式 */
.key-form, .rule-form {
  margin-top: 1.5rem;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-weight: 600;
  color: #ccc;
  font-size: 0.9rem;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 1rem;
  z-index: 1;
  font-size: 1.2rem;
}

.input-wrapper input {
  width: 100%;
  padding: 0.8rem 1rem 0.8rem 3rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 0.9rem;
  transition: all 0.3s ease;
}

.input-wrapper input:focus {
  outline: none;
  background: rgba(255, 255, 255, 0.08);
  border-color: #667eea;
}

.input-wrapper input::placeholder {
  color: #666;
}

.password-toggle {
  position: absolute;
  right: 1rem;
  background: none;
  border: none;
  color: #999;
  cursor: pointer;
  padding: 0.2rem;
  border-radius: 4px;
  transition: all 0.3s ease;
}

.password-toggle:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

/* 操作按钮 */
.form-actions {
  display: flex;
  gap: 1rem;
}

.action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.8rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-btn i {
  font-style: normal;
}

.action-btn.save, .action-btn.create {
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
  color: white;
  flex: 1;
}

.action-btn.save:hover, .action-btn.create:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(34, 197, 94, 0.4);
}

.action-btn.delete {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.action-btn.delete:hover {
  background: rgba(239, 68, 68, 0.2);
  transform: translateY(-2px);
}

.action-btn.view {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
  border: 1px solid rgba(59, 130, 246, 0.3);
  flex: 1;
}

.action-btn.view:hover {
  background: rgba(59, 130, 246, 0.2);
  transform: translateY(-2px);
}

.action-btn.secondary {
  background: rgba(108, 117, 125, 0.1);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
}

.action-btn.secondary:hover {
  background: rgba(108, 117, 125, 0.2);
}

/* 规则列表 */
.rules-list {
  margin-top: 2rem;
}

.empty-state {
  text-align: center;
  padding: 3rem 2rem;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.3;
}

.empty-text {
  color: #666;
  font-size: 1.2rem;
  margin-bottom: 0.5rem;
}

.empty-subtitle {
  color: #999;
  font-size: 1rem;
}

.rules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1.5rem;
  margin-top: 1.5rem;
}

.rule-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.rule-card:hover {
  background: rgba(255, 255, 255, 0.08);
  transform: translateY(-2px);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.rule-asset {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.asset-icon {
  width: 50px;
  height: 50px;
  border-radius: 12px;
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
}

.asset-info h4 {
  margin: 0 0 0.5rem 0;
  color: #fff;
  font-size: 1.2rem;
  font-weight: 600;
}

.rule-id {
  color: #666;
  font-size: 0.8rem;
}

/* 状态标签 */
.status-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.4rem 0.8rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
  position: relative;
  padding-left: 1.5rem;
}

.status-dot {
  position: absolute;
  left: 0.5rem;
  top: 50%;
  transform: translateY(-50%);
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-chip.enabled {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.status-chip.enabled .status-dot {
  background: #22c55e;
}

.status-chip.disabled {
  background: rgba(108, 117, 125, 0.2);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
}

.status-chip.disabled .status-dot {
  background: #94a3b8;
}

/* 规则详情 */
.rule-details {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-label {
  color: #666;
  font-size: 0.9rem;
  font-weight: 500;
}

.detail-value {
  color: #ccc;
  font-size: 0.9rem;
  font-weight: 500;
}

.detail-value.address {
  font-family: 'Courier New', monospace;
  color: #a78bfa;
}

/* 规则操作 */
.rule-actions {
  display: flex;
  gap: 0.5rem;
}

/* 弹窗样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(5px);
}

.modal-content {
  background: #1a1a1a;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 2rem 2rem 1rem 2rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.modal-header h3 {
  margin: 0;
  color: #fff;
  font-size: 1.5rem;
  font-weight: 600;
}

.close-btn {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.1);
  color: #ccc;
  border: none;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.modal-body {
  padding: 2rem;
}

/* 详情网格 */
.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.detail-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1rem;
}

.detail-card .detail-label {
  color: #999;
  font-size: 0.8rem;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.detail-card .detail-value {
  color: #fff;
  font-size: 1rem;
  font-weight: 600;
}

/* 地址显示 */
.address-section {
  margin-top: 2rem;
  padding-top: 2rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.address-section h4 {
  margin: 0 0 1rem 0;
  color: #fff;
  font-size: 1.1rem;
}

.address-display {
  display: flex;
  align-items: center;
  gap: 1rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 1rem;
}

.full-address {
  flex: 1;
  font-family: 'Courier New', monospace;
  color: #a78bfa;
  font-size: 0.9rem;
  word-break: break-all;
}

.copy-btn {
  background: rgba(167, 139, 250, 0.1);
  color: #a78bfa;
  border: 1px solid rgba(167, 139, 250, 0.3);
  padding: 0.5rem 1rem;
  border-radius: 6px;
  font-size: 0.8rem;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.copy-btn:hover {
  background: rgba(167, 139, 250, 0.2);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .settings-container {
    padding: 1rem;
  }

  .page-title {
    font-size: 2rem;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .stat-card {
    padding: 1.5rem;
  }

  .section-header {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
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

  .modal-content {
    width: 95%;
    max-height: 90vh;
  }

  .modal-header {
    padding: 1.5rem 1.5rem 1rem 1.5rem;
  }

  .modal-body {
    padding: 1.5rem;
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }

  .address-display {
    flex-direction: column;
    align-items: stretch;
  }

  .toast {
    left: 1rem;
    right: 1rem;
    bottom: 1rem;
  }
}
</style>
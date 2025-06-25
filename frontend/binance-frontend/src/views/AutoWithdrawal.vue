<template>
  <div class="auto-withdraw-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">
        <span class="gradient-text">自动提币管理</span>
      </h1>
      <p class="page-subtitle">设置自动提币规则，保护您的资产安全</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%)">
          <i>📋</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ rules.length }}</div>
          <div class="stat-label">总规则数</div>
        </div>
        <div class="stat-bg"></div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)">
          <i>✅</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ activeRulesCount }}</div>
          <div class="stat-label">活跃规则</div>
        </div>
        <div class="stat-bg"></div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #fa709a 0%, #fee140 100%)">
          <i>💰</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ totalWithdrawalsToday }}</div>
          <div class="stat-label">今日提币次数</div>
        </div>
        <div class="stat-bg"></div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%)">
          <i>📊</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ uniqueCoinsCount }}</div>
          <div class="stat-label">管理币种数</div>
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

    <!-- 创建规则区域 -->
    <div class="create-section">
      <div class="section-header">
        <h2 class="section-title">创建提币规则</h2>
        <button @click="toggleCreateForm" class="toggle-btn">
          <i>{{ showCreateForm ? '🔽' : '➕' }}</i>
          {{ showCreateForm ? '收起' : '创建规则' }}
        </button>
      </div>

      <transition name="form-slide">
        <div v-if="showCreateForm" class="create-form">
          <form @submit.prevent="createRule">
            <div class="form-grid">
              <div class="form-group">
                <label>币种</label>
                <select v-model="newRule.symbol" @change="onSymbolChange" required>
                  <option value="">选择币种</option>
                  <option v-for="coin in availableCoins" :key="coin" :value="coin">
                    {{ coin }}
                  </option>
                </select>
              </div>

              <div class="form-group">
                <label>提币网络</label>
                <select v-model="newRule.network" :disabled="!newRule.symbol" required>
                  <option value="">选择网络</option>
                  <option v-for="network in availableNetworks" :key="network" :value="network">
                    {{ network }}
                  </option>
                </select>
              </div>

              <div class="form-group">
                <label>提币地址</label>
                <input v-model="newRule.address"
                       type="text"
                       placeholder="输入提币地址"
                       required />
              </div>

              <div class="form-group">
                <label>最小提币金额</label>
                <input v-model.number="newRule.min_amount"
                       type="number"
                       step="0.00000001"
                       min="0"
                       placeholder="触发提币的最小金额"
                       required />
              </div>

              <div class="form-group">
                <label>启用状态</label>
                <div class="switch-container">
                  <label class="switch">
                    <input type="checkbox" v-model="newRule.enabled">
                    <span class="slider"></span>
                  </label>
                  <span class="switch-label">{{ newRule.enabled ? '启用' : '禁用' }}</span>
                </div>
              </div>
            </div>

            <div class="rule-description">
              <div class="description-card">
                <div class="description-icon">💡</div>
                <div class="description-content">
                  <h4>规则说明</h4>
                  <p>当您的{{ newRule.symbol || '[币种]' }}余额超过{{ newRule.min_amount || '[最小金额]' }}时，系统将自动将超出部分提币到指定地址。</p>
                  <small>提示：请确保提币地址正确，一旦提币将无法撤回。</small>
                </div>
              </div>
            </div>

            <div class="form-actions">
              <button type="submit" :disabled="isCreatingRule" class="create-btn">
                <i>{{ isCreatingRule ? '⏳' : '🚀' }}</i>
                {{ isCreatingRule ? '创建中...' : '创建规则' }}
              </button>
              <button type="button" @click="resetForm" class="reset-btn">
                <i>🔄</i> 重置表单
              </button>
            </div>
          </form>
        </div>
      </transition>
    </div>

    <!-- 规则列表 -->
    <div class="rules-section">
      <div class="section-header">
        <h2 class="section-title">提币规则列表</h2>
        <div class="search-box">
          <i class="search-icon">🔍</i>
          <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索币种或地址..."
              class="search-input"
          />
        </div>
      </div>

      <div v-if="filteredRules.length === 0" class="empty-state">
        <div class="empty-icon">📭</div>
        <p class="empty-text">暂无提币规则</p>
        <button @click="showCreateForm = true" class="empty-action">
          <i>➕</i> 创建第一个规则
        </button>
      </div>

      <div v-else class="rules-grid">
        <div v-for="rule in paginatedRules" :key="rule.id" class="rule-card">
          <div class="rule-header">
            <div class="rule-symbol">
              {{ rule.symbol }}
            </div>
            <div class="rule-status">
              <span :class="['status-chip', rule.enabled ? 'active' : 'inactive']">
                <span class="status-dot"></span>
                {{ rule.enabled ? '启用' : '禁用' }}
              </span>
            </div>
          </div>

          <div class="rule-info">
            <div class="info-item">
              <span class="info-label">网络</span>
              <span class="info-value network-chip">{{ rule.network }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">提币地址</span>
              <span class="info-value address" @click="copyAddress(rule.address)">
                {{ formatAddress(rule.address) }}
                <i class="copy-icon">📋</i>
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">最小金额</span>
              <span class="info-value amount">{{ rule.min_amount }} {{ rule.symbol }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">今日提币</span>
              <span class="info-value">{{ rule.withdrawals_today || 0 }} 次</span>
            </div>
            <div class="info-item">
              <span class="info-label">创建时间</span>
              <span class="info-value">{{ formatDate(rule.created_at) }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">最后执行</span>
              <span class="info-value">{{ rule.last_executed ? formatDate(rule.last_executed) : '从未' }}</span>
            </div>
          </div>

          <div class="rule-actions">
            <button @click="toggleRule(rule)"
                    :class="['action-btn', rule.enabled ? 'disable' : 'enable']">
              <i>{{ rule.enabled ? '⏸️' : '▶️' }}</i>
              {{ rule.enabled ? '禁用' : '启用' }}
            </button>

            <button @click="editRule(rule)" class="action-btn edit">
              <i>✏️</i> 编辑
            </button>

            <button @click="viewHistory(rule)" class="action-btn history">
              <i>📜</i> 历史记录
            </button>

            <button @click="deleteRule(rule.id)" class="action-btn delete">
              <i>🗑️</i> 删除
            </button>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination" v-if="filteredRules.length > pageSize">
        <button :disabled="currentPage === 1" @click="currentPage--" class="page-btn">
          <i>◀️</i> 上一页
        </button>
        <span class="page-info">第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
        <button :disabled="currentPage === totalPages" @click="currentPage++" class="page-btn">
          下一页 <i>▶️</i>
        </button>
      </div>
    </div>

    <!-- 编辑规则弹窗 -->
    <div v-if="showEditModal" class="modal-overlay" @click="closeEditModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>编辑提币规则</h3>
          <button @click="closeEditModal" class="close-btn">✕</button>
        </div>

        <div class="modal-body">
          <form @submit.prevent="updateRule">
            <div class="form-grid">
              <div class="form-group">
                <label>币种</label>
                <input :value="editingRule.symbol" disabled />
              </div>

              <div class="form-group">
                <label>网络</label>
                <input :value="editingRule.network" disabled />
              </div>

              <div class="form-group">
                <label>提币地址</label>
                <input v-model="editingRule.address" type="text" required />
              </div>

              <div class="form-group">
                <label>最小提币金额</label>
                <input v-model.number="editingRule.min_amount"
                       type="number"
                       step="0.00000001"
                       min="0"
                       required />
              </div>
            </div>

            <div class="modal-actions">
              <button type="submit" :disabled="isUpdatingRule" class="action-btn primary">
                <i>{{ isUpdatingRule ? '⏳' : '💾' }}</i>
                {{ isUpdatingRule ? '更新中...' : '更新规则' }}
              </button>
              <button type="button" @click="closeEditModal" class="action-btn secondary">
                <i>✕</i> 取消
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 历史记录弹窗 -->
    <div v-if="showHistoryModal" class="modal-overlay" @click="closeHistoryModal">
      <div class="modal-content large" @click.stop>
        <div class="modal-header">
          <h3>提币历史 - {{ selectedRule?.symbol }}</h3>
          <button @click="closeHistoryModal" class="close-btn">✕</button>
        </div>

        <div class="modal-body">
          <div v-if="withdrawHistory.length === 0" class="no-history">
            <div class="no-history-icon">📄</div>
            <p>暂无提币记录</p>
          </div>

          <div v-else class="history-table">
            <div class="table-header">
              <span>时间</span>
              <span>金额</span>
              <span>交易哈希</span>
              <span>状态</span>
            </div>
            <div v-for="record in withdrawHistory" :key="record.id" class="table-row">
              <span>{{ formatDateTime(record.created_at) }}</span>
              <span>{{ record.amount }} {{ record.symbol }}</span>
              <span class="tx-hash" @click="viewTransaction(record.tx_hash)">
                {{ formatTxHash(record.tx_hash) }}
                <i>🔗</i>
              </span>
              <span :class="['status-badge', record.status]">
                {{ getStatusText(record.status) }}
              </span>
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
  name: 'AutoWithdraw',
  data() {
    return {
      rules: [],
      newRule: {
        symbol: '',
        network: '',
        address: '',
        min_amount: '',
        enabled: true
      },
      editingRule: {},
      withdrawHistory: [],
      selectedRule: null,
      availableCoins: [],
      availableNetworks: [],
      networksBySymbol: {},
      currentPage: 1,
      pageSize: 9,
      searchQuery: '',
      showCreateForm: false,
      showEditModal: false,
      showHistoryModal: false,
      isCreatingRule: false,
      isUpdatingRule: false,
      toastMessage: '',
      toastType: 'success',
      totalWithdrawalsToday: 0
    };
  },
  computed: {
    filteredRules() {
      if (!this.searchQuery) return this.rules;

      const query = this.searchQuery.toLowerCase();
      return this.rules.filter(rule =>
          rule.symbol.toLowerCase().includes(query) ||
          rule.address.toLowerCase().includes(query) ||
          rule.network.toLowerCase().includes(query)
      );
    },
    paginatedRules() {
      const start = (this.currentPage - 1) * this.pageSize;
      const end = start + this.pageSize;
      return this.filteredRules.slice(start, end);
    },
    totalPages() {
      return Math.ceil(this.filteredRules.length / this.pageSize);
    },
    activeRulesCount() {
      return this.rules.filter(r => r.enabled).length;
    },
    uniqueCoinsCount() {
      return new Set(this.rules.map(r => r.symbol)).size;
    }
  },
  mounted() {
    this.fetchRules();
    this.fetchAvailableCoins();
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

    toggleCreateForm() {
      this.showCreateForm = !this.showCreateForm;
    },

    formatAddress(address) {
      if (!address) return '';
      if (address.length <= 16) return address;
      return `${address.slice(0, 8)}...${address.slice(-8)}`;
    },

    formatTxHash(hash) {
      if (!hash) return '';
      return `${hash.slice(0, 10)}...${hash.slice(-10)}`;
    },

    formatDate(dateString) {
      if (!dateString) return '从未';
      const date = new Date(dateString);
      const now = new Date();
      const diff = now - date;
      const days = Math.floor(diff / (1000 * 60 * 60 * 24));

      if (days === 0) return '今天';
      if (days === 1) return '昨天';
      if (days < 7) return `${days}天前`;
      if (days < 30) return `${Math.floor(days / 7)}周前`;

      return date.toLocaleDateString('zh-CN');
    },

    formatDateTime(dateString) {
      const date = new Date(dateString);
      return date.toLocaleString('zh-CN');
    },

    getStatusText(status) {
      const statusMap = {
        'pending': '处理中',
        'completed': '已完成',
        'failed': '失败',
        'cancelled': '已取消'
      };
      return statusMap[status] || status;
    },

    async copyAddress(address) {
      try {
        await navigator.clipboard.writeText(address);
        this.showToast('地址已复制到剪贴板');
      } catch (error) {
        this.showToast('复制失败，请手动复制', 'error');
      }
    },

    async fetchRules() {
      try {
        const response = await axios.get('/auto_withdraw_rules', {
          headers: this.getAuthHeaders()
        });
        this.rules = response.data.rules || [];
        this.totalWithdrawalsToday = response.data.total_withdrawals_today || 0;
      } catch (error) {
        console.error('获取规则失败:', error);
        this.showToast(error.response?.data?.error || '获取规则失败', 'error');
      }
    },

    async fetchAvailableCoins() {
      try {
        const response = await axios.get('/withdraw_coins', {
          headers: this.getAuthHeaders()
        });
        this.availableCoins = response.data.coins || [];
        this.networksBySymbol = response.data.networks_by_symbol || {};
      } catch (error) {
        console.error('获取可用币种失败:', error);
        this.showToast(error.response?.data?.error || '获取可用币种失败', 'error');
      }
    },

    onSymbolChange() {
      this.newRule.network = '';
      if (this.newRule.symbol && this.networksBySymbol[this.newRule.symbol]) {
        this.availableNetworks = this.networksBySymbol[this.newRule.symbol];
      } else {
        this.availableNetworks = [];
      }
    },

    async createRule() {
      try {
        // 验证必填字段
        if (!this.newRule.symbol || !this.newRule.network || !this.newRule.address || !this.newRule.min_amount) {
          this.showToast('请填写所有必填字段', 'error');
          return;
        }

        this.isCreatingRule = true;

        // 构建请求体，确保数据类型正确
        const ruleData = {
          symbol: this.newRule.symbol,
          network: this.newRule.network,
          address: this.newRule.address,
          min_amount: Number(this.newRule.min_amount),
          enabled: Boolean(this.newRule.enabled)
        };

        const response = await axios.post('/auto_withdraw_rule', ruleData, {
          headers: this.getAuthHeaders()
        });

        this.showToast('提币规则创建成功！', 'success');
        this.resetForm();
        this.showCreateForm = false;
        this.fetchRules();
      } catch (error) {
        console.error('创建规则失败:', error);
        this.showToast(error.response?.data?.error || '创建规则失败', 'error');
      } finally {
        this.isCreatingRule = false;
      }
    },

    async toggleRule(rule) {
      try {
        const response = await axios.put(`/auto_withdraw_rule/${rule.id}/toggle`, {}, {
          headers: this.getAuthHeaders()
        });

        this.showToast(response.data.message || '规则状态已更新');
        this.fetchRules();
      } catch (error) {
        console.error('切换规则状态失败:', error);
        this.showToast(error.response?.data?.error || '切换规则状态失败', 'error');
      }
    },

    editRule(rule) {
      this.editingRule = { ...rule };
      this.showEditModal = true;
    },

    closeEditModal() {
      this.showEditModal = false;
      this.editingRule = {};
    },

    async updateRule() {
      try {
        this.isUpdatingRule = true;

        const updateData = {
          address: this.editingRule.address,
          min_amount: Number(this.editingRule.min_amount)
        };

        const response = await axios.put(`/auto_withdraw_rule/${this.editingRule.id}`, updateData, {
          headers: this.getAuthHeaders()
        });

        this.showToast('规则更新成功！');
        this.closeEditModal();
        this.fetchRules();
      } catch (error) {
        console.error('更新规则失败:', error);
        this.showToast(error.response?.data?.error || '更新规则失败', 'error');
      } finally {
        this.isUpdatingRule = false;
      }
    },

    async viewHistory(rule) {
      try {
        this.selectedRule = rule;
        const response = await axios.get(`/auto_withdraw_rule/${rule.id}/history`, {
          headers: this.getAuthHeaders()
        });
        this.withdrawHistory = response.data.history || [];
        this.showHistoryModal = true;
      } catch (error) {
        console.error('获取历史记录失败:', error);
        this.showToast(error.response?.data?.error || '获取历史记录失败', 'error');
      }
    },

    closeHistoryModal() {
      this.showHistoryModal = false;
      this.withdrawHistory = [];
      this.selectedRule = null;
    },

    viewTransaction(txHash) {
      // 这里可以根据不同的网络打开不同的区块浏览器
      // 示例：打开 BSCScan
      window.open(`https://bscscan.com/tx/${txHash}`, '_blank');
    },

    async deleteRule(ruleId) {
      if (!window.confirm('确定要删除这个提币规则吗？删除后无法恢复。')) {
        return;
      }

      try {
        const response = await axios.delete(`/auto_withdraw_rule/${ruleId}`, {
          headers: this.getAuthHeaders()
        });

        this.showToast('规则删除成功！');
        this.fetchRules();
      } catch (error) {
        console.error('删除规则失败:', error);
        this.showToast(error.response?.data?.error || '删除规则失败', 'error');
      }
    },

    resetForm() {
      this.newRule = {
        symbol: '',
        network: '',
        address: '',
        min_amount: '',
        enabled: true
      };
      this.availableNetworks = [];
    }
  }
};
</script>

<style scoped>
/* 全局样式 */
.auto-withdraw-container {
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
  font-size: 2.5rem;
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

/* 创建规则区域 */
.create-section {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 24px;
  padding: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  margin-bottom: 3rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.section-title {
  font-size: 1.5rem;
  font-weight: 600;
}

.toggle-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.toggle-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
}

.form-slide-enter-active, .form-slide-leave-active {
  transition: all 0.3s ease;
}

.form-slide-enter-from {
  opacity: 0;
  transform: translateY(-20px);
}

.form-slide-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

.create-form {
  margin-top: 2rem;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
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

.form-group input, .form-group select {
  padding: 0.8rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 0.9rem;
  transition: all 0.3s ease;
}

.form-group input:focus, .form-group select:focus {
  outline: none;
  background: rgba(255, 255, 255, 0.08);
  border-color: #667eea;
}

.form-group input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.form-group input::placeholder {
  color: #666;
}

/* 开关样式 */
.switch-container {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 24px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(255, 255, 255, 0.1);
  transition: .4s;
  border-radius: 34px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background-color: white;
  transition: .4s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: #667eea;
}

input:checked + .slider:before {
  transform: translateX(26px);
}

.switch-label {
  color: #ccc;
  font-weight: 500;
}

/* 规则说明 */
.rule-description {
  margin: 2rem 0;
}

.description-card {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1.5rem;
}

.description-icon {
  font-size: 2rem;
  flex-shrink: 0;
}

.description-content h4 {
  margin: 0 0 0.5rem 0;
  color: #667eea;
  font-size: 1.1rem;
}

.description-content p {
  margin: 0 0 0.5rem 0;
  color: #ccc;
  line-height: 1.5;
}

.description-content small {
  color: #999;
  font-size: 0.85rem;
}

/* 表单操作 */
.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
}

.create-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem 2rem;
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.create-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(34, 197, 94, 0.4);
}

.create-btn:disabled {
  background: #6c757d;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.reset-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem 2rem;
  background: rgba(108, 117, 125, 0.1);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.reset-btn:hover {
  background: rgba(108, 117, 125, 0.2);
}

/* 规则列表区域 */
.rules-section {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 24px;
  padding: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

/* 搜索框 */
.search-box {
  position: relative;
  width: 300px;
}

.search-icon {
  position: absolute;
  left: 1rem;
  top: 50%;
  transform: translateY(-50%);
  font-size: 1.2rem;
}

.search-input {
  width: 100%;
  padding: 0.8rem 1rem 0.8rem 3rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  color: #fff;
  font-size: 0.9rem;
  transition: all 0.3s ease;
}

.search-input:focus {
  outline: none;
  background: rgba(255, 255, 255, 0.08);
  border-color: #667eea;
}

.search-input::placeholder {
  color: #666;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.3;
}

.empty-text {
  color: #666;
  font-size: 1.1rem;
  margin-bottom: 2rem;
}

.empty-action {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 2rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.empty-action:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
}

/* 规则卡片网格 */
.rules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 1.5rem;
  margin-top: 2rem;
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
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.rule-symbol {
  font-size: 1.3rem;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
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

.status-chip.active {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.status-chip.active .status-dot {
  background: #22c55e;
}

.status-chip.inactive {
  background: rgba(108, 117, 125, 0.2);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
}

.status-chip.inactive .status-dot {
  background: #94a3b8;
}

/* 规则信息 */
.rule-info {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.info-label {
  color: #666;
  font-size: 0.8rem;
  font-weight: 500;
}

.info-value {
  color: #ccc;
  font-size: 0.9rem;
  font-weight: 500;
}

.network-chip {
  display: inline-block;
  background: rgba(59, 130, 246, 0.2);
  color: #3b82f6;
  padding: 0.2rem 0.6rem;
  border-radius: 12px;
  font-size: 0.8rem;
}

.address {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  transition: color 0.3s ease;
}

.address:hover {
  color: #667eea;
}

.copy-icon {
  font-size: 0.8rem;
  opacity: 0.6;
}

.amount {
  color: #fbbf24;
  font-weight: 600;
}

/* 规则操作按钮 */
.rule-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.action-btn {
  flex: 1;
  min-width: 80px;
  padding: 0.6rem 0.8rem;
  border: none;
  border-radius: 8px;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
}

.action-btn i {
  font-style: normal;
  font-size: 0.9rem;
}

.action-btn.enable {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.action-btn.enable:hover {
  background: rgba(34, 197, 94, 0.2);
  transform: translateY(-1px);
}

.action-btn.disable {
  background: rgba(255, 193, 7, 0.1);
  color: #fbbf24;
  border: 1px solid rgba(255, 193, 7, 0.3);
}

.action-btn.disable:hover {
  background: rgba(255, 193, 7, 0.2);
  transform: translateY(-1px);
}

.action-btn.edit {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.action-btn.edit:hover {
  background: rgba(59, 130, 246, 0.2);
  transform: translateY(-1px);
}

.action-btn.history {
  background: rgba(139, 92, 246, 0.1);
  color: #8b5cf6;
  border: 1px solid rgba(139, 92, 246, 0.3);
}

.action-btn.history:hover {
  background: rgba(139, 92, 246, 0.2);
  transform: translateY(-1px);
}

.action-btn.delete {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.action-btn.delete:hover {
  background: rgba(239, 68, 68, 0.2);
  transform: translateY(-1px);
}

/* 分页 */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-top: 2rem;
}

.page-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1.2rem;
  background: rgba(255, 255, 255, 0.05);
  color: #ccc;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.page-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  color: #999;
  font-size: 0.9rem;
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

.modal-content.large {
  max-width: 900px;
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

/* 弹窗操作按钮 */
.modal-actions {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  margin-top: 2rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.action-btn.primary {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 0.8rem 1.5rem;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-btn.primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
}

.action-btn.primary:disabled {
  background: #6c757d;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.action-btn.secondary {
  background: rgba(108, 117, 125, 0.1);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
  padding: 0.8rem 1.5rem;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-btn.secondary:hover {
  background: rgba(108, 117, 125, 0.2);
}

/* 历史记录 */
.no-history {
  text-align: center;
  padding: 3rem;
  color: #666;
}

.no-history-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.history-table {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  overflow: hidden;
}

.table-header {
  display: grid;
  grid-template-columns: 1.5fr 1fr 2fr 1fr;
  gap: 1rem;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.05);
  font-weight: 600;
  color: #ccc;
  font-size: 0.9rem;
}

.table-row {
  display: grid;
  grid-template-columns: 1.5fr 1fr 2fr 1fr;
  gap: 1rem;
  padding: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  color: #ccc;
  font-size: 0.9rem;
  transition: all 0.3s ease;
}

.table-row:hover {
  background: rgba(255, 255, 255, 0.05);
}

.tx-hash {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  color: #667eea;
  transition: color 0.3s ease;
}

.tx-hash:hover {
  color: #764ba2;
}

.status-badge {
  padding: 0.2rem 0.5rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
  text-align: center;
}

.status-badge.pending {
  background: rgba(255, 193, 7, 0.2);
  color: #fbbf24;
}

.status-badge.completed {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

.status-badge.failed {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.status-badge.cancelled {
  background: rgba(108, 117, 125, 0.2);
  color: #94a3b8;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .auto-withdraw-container {
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

  .form-grid {
    grid-template-columns: 1fr;
  }

  .section-header {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }

  .search-box {
    width: 100%;
  }

  .rules-grid {
    grid-template-columns: 1fr;
  }

  .rule-info {
    grid-template-columns: 1fr;
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

  .table-header,
  .table-row {
    grid-template-columns: 1fr;
    gap: 0.5rem;
  }

  .table-header span,
  .table-row span {
    padding: 0.5rem 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .modal-actions {
    flex-direction: column;
  }

  .toast {
    left: 1rem;
    right: 1rem;
    bottom: 1rem;
  }
}
</style>
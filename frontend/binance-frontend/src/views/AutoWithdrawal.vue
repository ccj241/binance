<template>
  <div class="withdrawal-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1 class="page-title">自动提币管理</h1>
      <p class="page-description">设置自动提币规则，保护您的资产安全</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">
          <span>📋</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ rules.length }}</div>
          <div class="stat-label">总规则数</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon success">
          <span>✓</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ activeRulesCount }}</div>
          <div class="stat-label">活跃规则</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon primary">
          <span>💰</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ totalWithdrawalsToday }}</div>
          <div class="stat-label">今日提币</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon warning">
          <span>🪙</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ uniqueCoinsCount }}</div>
          <div class="stat-label">币种数量</div>
        </div>
      </div>
    </div>

    <!-- 创建规则 -->
    <div class="content-card">
      <div class="card-header">
        <h2 class="card-title">创建提币规则</h2>
        <button @click="showCreateForm = !showCreateForm" class="toggle-btn">
          {{ showCreateForm ? '收起' : '展开' }}
        </button>
      </div>

      <transition name="collapse">
        <div v-if="showCreateForm" class="card-body">
          <form @submit.prevent="createRule" class="rule-form">
            <div class="form-grid">
              <div class="form-group">
                <label class="form-label">币种</label>
                <select v-model="newRule.symbol" class="form-control" required>
                  <option value="">选择币种</option>
                  <option v-for="coin in availableCoins" :key="coin" :value="coin">
                    {{ coin }}
                  </option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">提币网络</label>
                <select v-model="newRule.network" class="form-control" :disabled="!newRule.symbol" required>
                  <option value="">选择网络</option>
                  <option v-for="network in availableNetworks" :key="network" :value="network">
                    {{ network }}
                  </option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">提币地址</label>
                <input
                    v-model="newRule.address"
                    type="text"
                    class="form-control"
                    placeholder="输入提币地址"
                    required
                />
              </div>

              <div class="form-group">
                <label class="form-label">最小提币金额</label>
                <input
                    v-model.number="newRule.min_amount"
                    type="number"
                    step="0.00000001"
                    min="0"
                    class="form-control"
                    placeholder="触发提币的最小金额"
                    required
                />
              </div>
            </div>

            <div class="form-footer">
              <div class="form-info">
                <span class="info-icon">💡</span>
                <span>当您的 {{ newRule.symbol || '[币种]' }} 余额超过 {{ newRule.min_amount || '[金额]' }} 时，系统将自动提币到指定地址</span>
              </div>

              <div class="form-actions">
                <button type="button" @click="resetForm" class="btn btn-outline">
                  重置
                </button>
                <button type="submit" class="btn btn-primary" :disabled="isCreatingRule">
                  {{ isCreatingRule ? '创建中...' : '创建规则' }}
                </button>
              </div>
            </div>
          </form>
        </div>
      </transition>
    </div>

    <!-- 规则列表 -->
    <div class="content-card">
      <div class="card-header">
        <h2 class="card-title">提币规则列表</h2>
        <input
            v-model="searchQuery"
            type="text"
            class="search-input"
            placeholder="搜索币种或地址..."
        />
      </div>

      <div class="card-body">
        <div v-if="filteredRules.length === 0" class="empty-state">
          <span class="empty-icon">📭</span>
          <p>暂无提币规则</p>
          <button @click="showCreateForm = true" class="btn btn-primary">
            添加第一个规则
          </button>
        </div>

        <div v-else class="rules-list">
          <div v-for="rule in paginatedRules" :key="rule.id" class="rule-item">
            <div class="rule-header">
              <div class="rule-info">
                <h3 class="rule-symbol">{{ rule.symbol }}</h3>
                <span :class="['status-badge', rule.enabled ? 'active' : 'inactive']">
                  {{ rule.enabled ? '启用' : '禁用' }}
                </span>
              </div>
              <span class="rule-id">ID: {{ rule.id }}</span>
            </div>

            <div class="rule-details">
              <div class="detail-item">
                <span class="detail-label">网络</span>
                <span class="detail-value">{{ rule.network }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">提币地址</span>
                <span class="detail-value address" @click="copyAddress(rule.address)">
                  {{ formatAddress(rule.address) }}
                  <span class="copy-icon">📋</span>
                </span>
              </div>
              <div class="detail-item">
                <span class="detail-label">最小金额</span>
                <span class="detail-value">{{ rule.min_amount }} {{ rule.symbol }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">今日提币</span>
                <span class="detail-value">{{ rule.withdrawals_today || 0 }} 次</span>
              </div>
            </div>

            <div class="rule-actions">
              <button
                  @click="toggleRule(rule)"
                  :class="['btn', 'btn-sm', rule.enabled ? 'btn-outline' : 'btn-success']"
              >
                {{ rule.enabled ? '禁用' : '启用' }}
              </button>
              <button @click="editRule(rule)" class="btn btn-sm btn-outline">
                编辑
              </button>
              <button @click="viewHistory(rule)" class="btn btn-sm btn-outline">
                历史
              </button>
              <button @click="deleteRule(rule.id)" class="btn btn-sm btn-danger">
                删除
              </button>
            </div>
          </div>
        </div>

        <!-- 分页 -->
        <div v-if="filteredRules.length > pageSize" class="pagination">
          <button
              :disabled="currentPage === 1"
              @click="currentPage--"
              class="page-btn"
          >
            上一页
          </button>
          <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
          <button
              :disabled="currentPage === totalPages"
              @click="currentPage++"
              class="page-btn"
          >
            下一页
          </button>
        </div>
      </div>
    </div>

    <!-- 编辑弹窗 -->
    <transition name="modal">
      <div v-if="showEditModal" class="modal-overlay" @click="closeEditModal">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">编辑提币规则</h3>
            <button @click="closeEditModal" class="modal-close">×</button>
          </div>

          <form @submit.prevent="updateRule" class="modal-body">
            <div class="form-group">
              <label class="form-label">币种</label>
              <input :value="editingRule.symbol" class="form-control" disabled />
            </div>

            <div class="form-group">
              <label class="form-label">网络</label>
              <input :value="editingRule.network" class="form-control" disabled />
            </div>

            <div class="form-group">
              <label class="form-label">提币地址</label>
              <input v-model="editingRule.address" class="form-control" required />
            </div>

            <div class="form-group">
              <label class="form-label">最小提币金额</label>
              <input
                  v-model.number="editingRule.min_amount"
                  type="number"
                  step="0.00000001"
                  min="0"
                  class="form-control"
                  required
              />
            </div>

            <div class="modal-footer">
              <button type="button" @click="closeEditModal" class="btn btn-outline">
                取消
              </button>
              <button type="submit" class="btn btn-primary" :disabled="isUpdatingRule">
                {{ isUpdatingRule ? '更新中...' : '更新规则' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </transition>

    <!-- 历史记录弹窗 -->
    <transition name="modal">
      <div v-if="showHistoryModal" class="modal-overlay" @click="closeHistoryModal">
        <div class="modal-content modal-lg" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">提币历史 - {{ selectedRule?.symbol }}</h3>
            <button @click="closeHistoryModal" class="modal-close">×</button>
          </div>

          <div class="modal-body">
            <div v-if="withdrawHistory.length === 0" class="empty-state">
              <span class="empty-icon">📄</span>
              <p>暂无提币记录</p>
            </div>

            <table v-else class="data-table">
              <thead>
              <tr>
                <th>时间</th>
                <th>金额</th>
                <th>交易哈希</th>
                <th>状态</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="record in withdrawHistory" :key="record.id">
                <td>{{ formatDateTime(record.created_at) }}</td>
                <td>{{ record.amount }} {{ record.symbol }}</td>
                <td>
                    <span class="tx-hash" @click="viewTransaction(record.tx_hash)">
                      {{ formatTxHash(record.tx_hash) }}
                      <span class="link-icon">🔗</span>
                    </span>
                </td>
                <td>
                    <span :class="['status-badge', record.status]">
                      {{ getStatusText(record.status) }}
                    </span>
                </td>
              </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </transition>

    <!-- Toast 消息 -->
    <transition name="toast">
      <div v-if="toastMessage" :class="['toast', toastType]">
        <span class="toast-icon">{{ toastType === 'success' ? '✓' : '×' }}</span>
        <span>{{ toastMessage }}</span>
      </div>
    </transition>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'AutoWithdrawal',
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
  watch: {
    'newRule.symbol'(newVal) {
      this.newRule.network = '';
      if (newVal && this.networksBySymbol[newVal]) {
        this.availableNetworks = this.networksBySymbol[newVal];
      } else {
        this.availableNetworks = [];
      }
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

    formatAddress(address) {
      if (!address) return '';
      if (address.length <= 16) return address;
      return `${address.slice(0, 8)}...${address.slice(-8)}`;
    },

    formatTxHash(hash) {
      if (!hash) return '';
      return `${hash.slice(0, 10)}...${hash.slice(-10)}`;
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
        this.showToast('地址已复制');
      } catch (error) {
        this.showToast('复制失败', 'error');
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
        this.showToast('获取规则失败', 'error');
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
        this.showToast('获取可用币种失败', 'error');
      }
    },

    async createRule() {
      if (!this.newRule.symbol || !this.newRule.network || !this.newRule.address || !this.newRule.min_amount) {
        this.showToast('请填写所有必填字段', 'error');
        return;
      }

      this.isCreatingRule = true;
      try {
        const ruleData = {
          symbol: this.newRule.symbol,
          network: this.newRule.network,
          address: this.newRule.address,
          min_amount: Number(this.newRule.min_amount),
          enabled: Boolean(this.newRule.enabled)
        };

        await axios.post('/auto_withdraw_rule', ruleData, {
          headers: this.getAuthHeaders()
        });

        this.showToast('提币规则创建成功');
        this.resetForm();
        this.showCreateForm = false;
        await this.fetchRules();
      } catch (error) {
        console.error('创建规则失败:', error);
        this.showToast(error.response?.data?.error || '创建规则失败', 'error');
      } finally {
        this.isCreatingRule = false;
      }
    },

    async toggleRule(rule) {
      try {
        await axios.put(`/auto_withdraw_rule/${rule.id}/toggle`, {}, {
          headers: this.getAuthHeaders()
        });

        this.showToast('规则状态已更新');
        await this.fetchRules();
      } catch (error) {
        console.error('切换规则状态失败:', error);
        this.showToast('切换规则状态失败', 'error');
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
      this.isUpdatingRule = true;
      try {
        const updateData = {
          address: this.editingRule.address,
          min_amount: Number(this.editingRule.min_amount)
        };

        await axios.put(`/auto_withdraw_rule/${this.editingRule.id}`, updateData, {
          headers: this.getAuthHeaders()
        });

        this.showToast('规则更新成功');
        this.closeEditModal();
        await this.fetchRules();
      } catch (error) {
        console.error('更新规则失败:', error);
        this.showToast('更新规则失败', 'error');
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
        this.showToast('获取历史记录失败', 'error');
      }
    },

    closeHistoryModal() {
      this.showHistoryModal = false;
      this.withdrawHistory = [];
      this.selectedRule = null;
    },

    viewTransaction(txHash) {
      window.open(`https://bscscan.com/tx/${txHash}`, '_blank');
    },

    async deleteRule(ruleId) {
      if (!confirm('确定要删除这个提币规则吗？')) {
        return;
      }

      try {
        await axios.delete(`/auto_withdraw_rule/${ruleId}`, {
          headers: this.getAuthHeaders()
        });

        this.showToast('规则删除成功');
        await this.fetchRules();
      } catch (error) {
        console.error('删除规则失败:', error);
        this.showToast('删除规则失败', 'error');
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
/* 页面容器 */
.withdrawal-container {
  max-width: 1200px;
  margin: 0 auto;
}

/* 页面头部 */
.page-header {
  margin-bottom: 2rem;
}

.page-title {
  font-size: 1.875rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.5rem 0;
}

.page-description {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  background-color: var(--color-bg-tertiary);
  color: var(--color-text-secondary);
}

.stat-icon.primary {
  background-color: #dbeafe;
  color: var(--color-primary);
}

.stat-icon.success {
  background-color: #d1fae5;
  color: var(--color-success);
}

.stat-icon.warning {
  background-color: #fef3c7;
  color: var(--color-warning);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--color-text-primary);
  line-height: 1;
  margin-bottom: 0.25rem;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

/* 内容卡片 */
.content-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  margin-bottom: 1.5rem;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--color-border);
}

.card-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.card-body {
  padding: 1.5rem;
}

/* 表单样式 */
.rule-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.form-control {
  padding: 0.625rem 0.875rem;
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-primary);
  font-size: 0.875rem;
  transition: all var(--transition-normal);
}

.form-control:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

.form-control:disabled {
  background-color: var(--color-bg-tertiary);
  cursor: not-allowed;
}

.form-footer {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-info {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.75rem;
  background-color: var(--color-bg-secondary);
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.info-icon {
  font-size: 1rem;
  flex-shrink: 0;
}

.form-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: flex-end;
}

/* 规则列表 */
.rules-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.rule-item {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  transition: all var(--transition-normal);
}

.rule-item:hover {
  background-color: var(--color-bg-secondary);
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.rule-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.rule-symbol {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.rule-id {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.rule-details {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 1rem;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.detail-value {
  font-size: 0.875rem;
  color: var(--color-text-primary);
  font-weight: 500;
}

.detail-value.address {
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.25rem;
  color: var(--color-primary);
}

.detail-value.address:hover {
  text-decoration: underline;
}

.copy-icon {
  font-size: 0.75rem;
  opacity: 0.7;
}

.rule-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

/* 按钮样式 */
.btn {
  padding: 0.5rem 1rem;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.75rem;
}

.btn-primary {
  background-color: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  background-color: var(--color-primary-hover);
}

.btn-primary:disabled {
  background-color: var(--color-secondary);
  cursor: not-allowed;
}

.btn-success {
  background-color: var(--color-success);
  color: white;
}

.btn-success:hover {
  background-color: #059669;
}

.btn-danger {
  background-color: var(--color-danger);
  color: white;
}

.btn-danger:hover {
  background-color: #dc2626;
}

.btn-outline {
  background-color: transparent;
  border-color: var(--color-border);
  color: var(--color-text-secondary);
}

.btn-outline:hover {
  background-color: var(--color-bg-tertiary);
  border-color: var(--color-text-tertiary);
}

.toggle-btn {
  padding: 0.375rem 0.75rem;
  background-color: transparent;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all var(--transition-normal);
}

.toggle-btn:hover {
  background-color: var(--color-bg-tertiary);
}

/* 搜索框 */
.search-input {
  padding: 0.5rem 1rem;
  background-color: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  color: var(--color-text-primary);
  width: 240px;
  transition: all var(--transition-normal);
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
  background-color: var(--color-bg);
}

/* 状态徽章 */
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-badge.active {
  background-color: #d1fae5;
  color: #065f46;
}

.status-badge.inactive {
  background-color: var(--color-bg-tertiary);
  color: var(--color-text-secondary);
}

.status-badge.pending {
  background-color: #fef3c7;
  color: #92400e;
}

.status-badge.completed {
  background-color: #d1fae5;
  color: #065f46;
}

.status-badge.failed {
  background-color: #fee2e2;
  color: #991b1b;
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
  padding: 0.5rem 1rem;
  background-color: transparent;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all var(--transition-normal);
}

.page-btn:hover:not(:disabled) {
  background-color: var(--color-bg-tertiary);
  border-color: var(--color-text-tertiary);
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}

/* 空状态 */
.empty-state {
  padding: 3rem 2rem;
  text-align: center;
  color: var(--color-text-tertiary);
}

.empty-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: 1rem;
  opacity: 0.5;
}

/* 弹窗 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--color-bg);
  border-radius: var(--radius-lg);
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.modal-lg {
  max-width: 800px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--color-border);
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.modal-close {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: transparent;
  border: none;
  border-radius: var(--radius-md);
  color: var(--color-text-tertiary);
  font-size: 1.5rem;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.modal-close:hover {
  background-color: var(--color-bg-tertiary);
  color: var(--color-text-primary);
}

.modal-body {
  padding: 1.5rem;
  overflow-y: auto;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--color-border);
}

/* 数据表格 */
.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  text-align: left;
  padding: 0.75rem;
  background-color: var(--color-bg-secondary);
  color: var(--color-text-secondary);
  font-weight: 600;
  font-size: 0.875rem;
  border-bottom: 1px solid var(--color-border);
}

.data-table td {
  padding: 0.75rem;
  border-bottom: 1px solid var(--color-border);
}

.tx-hash {
  cursor: pointer;
  color: var(--color-primary);
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.tx-hash:hover {
  text-decoration: underline;
}

.link-icon {
  font-size: 0.75rem;
  opacity: 0.7;
}

/* Toast 消息 */
.toast {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  font-weight: 500;
  z-index: 1000;
}

.toast.success {
  border-color: var(--color-success);
  color: var(--color-success);
}

.toast.error {
  border-color: var(--color-danger);
  color: var(--color-danger);
}

.toast-icon {
  font-size: 1.25rem;
}

/* 动画 */
.collapse-enter-active,
.collapse-leave-active {
  transition: all 0.3s ease;
}

.collapse-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.collapse-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

.modal-enter-active,
.modal-leave-active {
  transition: all 0.3s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .modal-content,
.modal-leave-to .modal-content {
  transform: scale(0.95);
}

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

/* 响应式设计 */
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr 1fr;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .rule-details {
    grid-template-columns: 1fr;
  }

  .search-input {
    width: 100%;
  }

  .modal-content {
    width: 95%;
  }

  .data-table {
    font-size: 0.875rem;
  }

  .data-table th,
  .data-table td {
    padding: 0.5rem;
  }
}
</style>
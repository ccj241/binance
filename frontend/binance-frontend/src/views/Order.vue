<template>
  <div class="order-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1 class="page-title">订单管理</h1>
      <p class="page-description">管理您的交易订单</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">
          <span>📊</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">总订单数</div>
          <div class="stat-value">{{ allOrdersCount }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon pending">
          <span>⏳</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">待处理</div>
          <div class="stat-value">{{ pendingOrdersCount }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon success">
          <span>✅</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">已成交</div>
          <div class="stat-value">{{ filledOrdersCount }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon cancelled">
          <span>❌</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">已取消</div>
          <div class="stat-value">{{ cancelledOrdersCount }}</div>
        </div>
      </div>
    </div>

    <!-- 筛选和操作区 -->
    <div class="filter-section">
      <div class="filter-tabs">
        <button
            v-for="tab in filterTabs"
            :key="tab.value"
            @click="filterStatus = tab.value"
            :class="['filter-tab', { active: filterStatus === tab.value }]"
        >
          <span class="tab-icon">{{ tab.icon }}</span>
          <span>{{ tab.label }}</span>
          <span class="tab-count">{{ getFilterCount(tab.value) }}</span>
        </button>
      </div>

      <div class="filter-controls">
        <div class="search-box">
          <span class="search-icon">🔍</span>
          <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索订单..."
              class="search-input"
          />
        </div>
        <button @click="fetchOrders" class="refresh-btn">
          <span>🔄</span>
          刷新
        </button>
      </div>
    </div>

    <!-- 批量操作 -->
    <transition name="slide">
      <div v-if="filterStatus === 'pending' && selectedOrders.length > 0" class="batch-actions">
        <div class="selected-info">
          <span class="checkbox-icon">☑️</span>
          <span>已选择 {{ selectedOrders.length }} 个订单</span>
        </div>
        <button @click="batchCancelOrders" class="btn btn-danger btn-sm">
          <span>🗑️</span>
          批量取消
        </button>
      </div>
    </transition>

    <!-- 订单列表 -->
    <div class="orders-section">
      <div v-if="filteredOrders.length === 0" class="empty-state">
        <div class="empty-icon">📋</div>
        <p class="empty-text">暂无{{ filterStatusText }}订单</p>
      </div>

      <div v-else class="orders-list">
        <div v-for="order in paginatedOrders" :key="order.id" class="order-card">
          <!-- 选择框 -->
          <div v-if="filterStatus === 'pending'" class="order-select">
            <input
                type="checkbox"
                :value="order.orderId || order.id"
                v-model="selectedOrders"
                class="checkbox"
            />
          </div>

          <!-- 订单头部 -->
          <div class="order-header">
            <div class="order-symbol">
              <h3>{{ order.symbol }}</h3>
              <span :class="['side-badge', order.side.toLowerCase()]">
                {{ order.side === 'BUY' ? '买入' : '卖出' }}
              </span>
            </div>
            <span :class="['status-badge', order.status]">
              {{ getStatusText(order.status) }}
            </span>
          </div>

          <!-- 订单详情 -->
          <div class="order-details">
            <div class="detail-row">
              <div class="detail-item">
                <span class="detail-label">订单ID</span>
                <span class="detail-value">{{ order.orderId || order.id }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">价格</span>
                <span class="detail-value highlight">{{ formatPrice(order.price) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">数量</span>
                <span class="detail-value">{{ formatQuantity(order.quantity) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">总值</span>
                <span class="detail-value">{{ formatPrice(order.price * order.quantity) }}</span>
              </div>
            </div>
          </div>

          <!-- 时间信息 -->
          <div class="order-time">
            <span class="time-icon">🕐</span>
            <span>创建于 {{ formatDate(order.createdAt) }}</span>
            <span class="time-separator">•</span>
            <span>更新于 {{ formatDate(order.updatedAt) }}</span>
          </div>

          <!-- 操作按钮 -->
          <div class="order-actions">
            <button
                v-if="order.status === 'pending'"
                @click="cancelOrder(order.orderId || order.id)"
                class="btn btn-outline btn-sm"
            >
              <span>❌</span>
              取消订单
            </button>
            <button
                @click="viewOrderDetails(order)"
                class="btn btn-outline btn-sm"
            >
              <span>👁️</span>
              查看详情
            </button>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="filteredOrders.length > pageSize" class="pagination">
        <button
            :disabled="currentPage === 1"
            @click="currentPage--"
            class="page-btn"
        >
          <span>←</span>
          上一页
        </button>
        <div class="page-info">
          <span>第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
          <span class="page-total">（共 {{ filteredOrders.length }} 条）</span>
        </div>
        <button
            :disabled="currentPage === totalPages"
            @click="currentPage++"
            class="page-btn"
        >
          下一页
          <span>→</span>
        </button>
      </div>
    </div>

    <!-- 创建订单 -->
    <div class="create-section">
      <div class="section-header">
        <h2 class="section-title">创建新订单</h2>
      </div>

      <form @submit.prevent="createOrder" class="order-form">
        <div class="form-grid">
          <div class="form-group">
            <label class="form-label">交易对</label>
            <input
                v-model="newOrder.symbol"
                placeholder="如: BTCUSDT"
                @input="newOrder.symbol = newOrder.symbol.toUpperCase()"
                class="form-control"
                required
            />
          </div>

          <div class="form-group">
            <label class="form-label">交易方向</label>
            <select v-model="newOrder.side" class="form-control" required>
              <option value="">选择方向</option>
              <option value="BUY">买入</option>
              <option value="SELL">卖出</option>
            </select>
          </div>

          <div class="form-group">
            <label class="form-label">价格</label>
            <input
                v-model.number="newOrder.price"
                type="number"
                step="0.00000001"
                placeholder="限价单价格"
                class="form-control"
                required
            />
          </div>

          <div class="form-group">
            <label class="form-label">数量</label>
            <input
                v-model.number="newOrder.quantity"
                type="number"
                step="0.00000001"
                placeholder="交易数量"
                class="form-control"
                required
            />
          </div>
        </div>

        <!-- 订单预览 -->
        <transition name="fade">
          <div v-if="newOrder.price > 0 && newOrder.quantity > 0" class="order-preview">
            <h3 class="preview-title">订单预览</h3>
            <div class="preview-content">
              <div class="preview-item">
                <span class="preview-label">订单总值</span>
                <span class="preview-value">
                  {{ (newOrder.price * newOrder.quantity).toFixed(8) }} {{ getQuoteCurrency() }}
                </span>
              </div>
              <div class="preview-item">
                <span class="preview-label">预估手续费</span>
                <span class="preview-value">
                  {{ ((newOrder.price * newOrder.quantity) * 0.001).toFixed(8) }} {{ getQuoteCurrency() }}
                </span>
              </div>
            </div>
          </div>
        </transition>

        <button
            type="submit"
            :disabled="isCreatingOrder || !isFormValid"
            class="submit-btn"
        >
          <span v-if="!isCreatingOrder">创建订单</span>
          <span v-else class="btn-loading">
            <span class="spinner"></span>
            创建中...
          </span>
        </button>
      </form>
    </div>

    <!-- 订单详情弹窗 -->
    <transition name="modal">
      <div v-if="showOrderDetails" class="modal-overlay" @click="closeOrderDetails">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">订单详情</h3>
            <button @click="closeOrderDetails" class="modal-close">×</button>
          </div>

          <div class="modal-body">
            <div class="detail-grid">
              <div class="detail-group">
                <label>订单ID</label>
                <span>{{ selectedOrderDetails.orderId }}</span>
              </div>
              <div class="detail-group">
                <label>交易对</label>
                <span>{{ selectedOrderDetails.symbol }}</span>
              </div>
              <div class="detail-group">
                <label>方向</label>
                <span :class="['side-badge', selectedOrderDetails.side?.toLowerCase()]">
                  {{ selectedOrderDetails.side === 'BUY' ? '买入' : '卖出' }}
                </span>
              </div>
              <div class="detail-group">
                <label>价格</label>
                <span>{{ formatPrice(selectedOrderDetails.price) }}</span>
              </div>
              <div class="detail-group">
                <label>数量</label>
                <span>{{ formatQuantity(selectedOrderDetails.quantity) }}</span>
              </div>
              <div class="detail-group">
                <label>状态</label>
                <span :class="['status-badge', selectedOrderDetails.status]">
                  {{ getStatusText(selectedOrderDetails.status) }}
                </span>
              </div>
              <div class="detail-group">
                <label>创建时间</label>
                <span>{{ new Date(selectedOrderDetails.createdAt).toLocaleString('zh-CN') }}</span>
              </div>
              <div class="detail-group">
                <label>更新时间</label>
                <span>{{ new Date(selectedOrderDetails.updatedAt).toLocaleString('zh-CN') }}</span>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeOrderDetails" class="btn btn-primary">
              关闭
            </button>
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
  name: 'Order',
  data() {
    return {
      orders: [],
      newOrder: {
        symbol: '',
        side: '',
        price: 0,
        quantity: 0
      },
      currentPage: 1,
      pageSize: 10,
      filterStatus: 'all',
      searchQuery: '',
      selectedOrders: [],
      isCreatingOrder: false,
      showOrderDetails: false,
      selectedOrderDetails: {},
      toastMessage: '',
      toastType: 'success',
      refreshInterval: null,
      filterTabs: [
        { value: 'all', label: '全部', icon: '📋' },
        { value: 'pending', label: '待处理', icon: '⏳' },
        { value: 'filled', label: '已成交', icon: '✅' },
        { value: 'cancelled', label: '已取消', icon: '❌' }
      ]
    };
  },
  computed: {
    filteredOrders() {
      let filtered = this.orders;

      // 状态筛选
      if (this.filterStatus !== 'all') {
        if (this.filterStatus === 'cancelled') {
          filtered = filtered.filter(order =>
              ['cancelled', 'expired', 'rejected'].includes(order.status)
          );
        } else {
          filtered = filtered.filter(order => order.status === this.filterStatus);
        }
      }

      // 搜索筛选
      if (this.searchQuery) {
        const query = this.searchQuery.toLowerCase();
        filtered = filtered.filter(order =>
            order.symbol.toLowerCase().includes(query) ||
            (order.orderId && order.orderId.toString().includes(query))
        );
      }

      return filtered;
    },

    paginatedOrders() {
      const start = (this.currentPage - 1) * this.pageSize;
      const end = start + this.pageSize;
      return this.filteredOrders.slice(start, end);
    },

    totalPages() {
      return Math.ceil(this.filteredOrders.length / this.pageSize);
    },

    filterStatusText() {
      const texts = {
        'all': '',
        'pending': '待处理',
        'filled': '已成交',
        'cancelled': '已取消'
      };
      return texts[this.filterStatus] || '';
    },

    allOrdersCount() {
      return this.orders.length;
    },

    pendingOrdersCount() {
      return this.orders.filter(o => o.status === 'pending').length;
    },

    filledOrdersCount() {
      return this.orders.filter(o => o.status === 'filled').length;
    },

    cancelledOrdersCount() {
      return this.orders.filter(o =>
          ['cancelled', 'expired', 'rejected'].includes(o.status)
      ).length;
    },

    isFormValid() {
      return this.newOrder.symbol &&
          this.newOrder.side &&
          this.newOrder.price > 0 &&
          this.newOrder.quantity > 0;
    }
  },

  watch: {
    filterStatus() {
      this.currentPage = 1;
      this.selectedOrders = [];
    },
    searchQuery() {
      this.currentPage = 1;
    }
  },

  mounted() {
    this.fetchOrders();
    this.refreshInterval = setInterval(this.fetchOrders, 30000);
  },

  beforeUnmount() {
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval);
    }
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

    formatPrice(price) {
      return parseFloat(price).toFixed(8).replace(/\.?0+$/, '');
    },

    formatQuantity(quantity) {
      return parseFloat(quantity).toFixed(8).replace(/\.?0+$/, '');
    },

    formatDate(dateString) {
      if (!dateString) return '-';
      const date = new Date(dateString);
      const now = new Date();
      const diff = now - date;
      const hours = Math.floor(diff / (1000 * 60 * 60));

      if (hours < 1) return '刚刚';
      if (hours < 24) return `${hours}小时前`;

      return date.toLocaleDateString('zh-CN', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      });
    },

    getStatusText(status) {
      const statusMap = {
        'pending': '待处理',
        'filled': '已成交',
        'cancelled': '已取消',
        'expired': '已过期',
        'rejected': '已拒绝'
      };
      return statusMap[status] || status;
    },

    getQuoteCurrency() {
      if (!this.newOrder.symbol) return '';
      if (this.newOrder.symbol.endsWith('USDT')) return 'USDT';
      if (this.newOrder.symbol.endsWith('BTC')) return 'BTC';
      if (this.newOrder.symbol.endsWith('ETH')) return 'ETH';
      if (this.newOrder.symbol.endsWith('BNB')) return 'BNB';
      return '';
    },

    getFilterCount(filterValue) {
      if (filterValue === 'all') return this.orders.length;
      if (filterValue === 'cancelled') {
        return this.orders.filter(o => ['cancelled', 'expired', 'rejected'].includes(o.status)).length;
      }
      return this.orders.filter(o => o.status === filterValue).length;
    },

    async fetchOrders() {
      try {
        const response = await axios.get('/orders', {
          headers: this.getAuthHeaders(),
        });
        this.orders = response.data.orders || [];
        this.orders.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
      } catch (error) {
        console.error('获取订单失败:', error);
        this.showToast(error.response?.data?.error || '获取订单失败', 'error');
      }
    },

    async createOrder() {
      if (!this.isFormValid) {
        this.showToast('请填写所有必需字段', 'error');
        return;
      }

      this.isCreatingOrder = true;
      try {
        const response = await axios.post('/order', this.newOrder, {
          headers: this.getAuthHeaders(),
        });

        this.showToast(response.data.message || '订单创建成功');
        this.newOrder = { symbol: '', side: '', price: 0, quantity: 0 };
        await this.fetchOrders();
      } catch (error) {
        console.error('创建订单失败:', error);
        this.showToast(error.response?.data?.error || '创建订单失败', 'error');
      } finally {
        this.isCreatingOrder = false;
      }
    },

    async cancelOrder(orderId) {
      if (!window.confirm('确定要取消这个订单吗？')) {
        return;
      }

      try {
        const response = await axios.post(`/cancel_order/${orderId}`, {}, {
          headers: this.getAuthHeaders(),
        });

        this.showToast(response.data.message || '订单取消成功');
        await this.fetchOrders();
      } catch (error) {
        console.error('取消订单失败:', error);
        this.showToast(error.response?.data?.error || '取消订单失败', 'error');
      }
    },

    async batchCancelOrders() {
      if (this.selectedOrders.length === 0) return;

      const count = this.selectedOrders.length;
      if (!window.confirm(`确定要批量取消 ${count} 个订单吗？`)) {
        return;
      }

      try {
        const response = await axios.post('/batch_cancel_orders', {
          orderIds: this.selectedOrders.map(id => parseInt(id))
        }, {
          headers: this.getAuthHeaders(),
        });

        const results = response.data.results;
        if (results && results.failed && results.failed.length === 0) {
          this.showToast(`成功取消 ${results.success.length} 个订单`);
        } else {
          this.showToast(response.data.message || '批量取消完成', results?.failed?.length > 0 ? 'error' : 'success');
        }
      } catch (error) {
        console.error('批量取消订单失败:', error);
        this.showToast(error.response?.data?.error || '批量取消订单失败', 'error');
      }

      this.selectedOrders = [];
      await this.fetchOrders();
    },

    viewOrderDetails(order) {
      this.selectedOrderDetails = order;
      this.showOrderDetails = true;
    },

    closeOrderDetails() {
      this.showOrderDetails = false;
      this.selectedOrderDetails = {};
    }
  }
};
</script>

<style scoped>
/* 页面容器 */
.order-container {
  max-width: 1400px;
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
  background: var(--color-bg-tertiary);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
}

.stat-icon.pending {
  background: #fef3c7;
  color: #f59e0b;
}

.stat-icon.success {
  background: #d1fae5;
  color: #10b981;
}

.stat-icon.cancelled {
  background: #fee2e2;
  color: #ef4444;
}

.stat-content {
  flex: 1;
}

.stat-label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
  margin-bottom: 0.25rem;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

/* 筛选区域 */
.filter-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
  flex-wrap: wrap;
}

.filter-tabs {
  display: flex;
  gap: 0.5rem;
}

.filter-tab {
  padding: 0.625rem 1rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-tab:hover {
  background: var(--color-bg-secondary);
}

.filter-tab.active {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: white;
}

.tab-icon {
  font-size: 1rem;
}

.tab-count {
  background: rgba(0, 0, 0, 0.1);
  padding: 0.125rem 0.375rem;
  border-radius: 10px;
  font-size: 0.75rem;
}

.filter-tab.active .tab-count {
  background: rgba(255, 255, 255, 0.2);
}

.filter-controls {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

/* 搜索框 */
.search-box {
  position: relative;
}

.search-icon {
  position: absolute;
  left: 0.875rem;
  top: 50%;
  transform: translateY(-50%);
  font-size: 1rem;
  color: var(--color-text-tertiary);
}

.search-input {
  padding: 0.625rem 0.875rem 0.625rem 2.5rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-primary);
  font-size: 0.875rem;
  width: 240px;
  transition: all var(--transition-normal);
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

.refresh-btn {
  padding: 0.625rem 1rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.refresh-btn:hover {
  background: var(--color-bg-secondary);
}

/* 批量操作 */
.batch-actions {
  background: #fef3c7;
  border: 1px solid #fbbf24;
  border-radius: var(--radius-md);
  padding: 0.75rem 1rem;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.selected-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: #92400e;
  font-size: 0.875rem;
  font-weight: 500;
}

.checkbox-icon {
  font-size: 1.125rem;
}

/* 订单列表 */
.orders-section {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  margin-bottom: 2rem;
}

.orders-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.order-card {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 1.25rem;
  transition: all var(--transition-normal);
  position: relative;
}

.order-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.order-select {
  position: absolute;
  top: 1.25rem;
  left: 1.25rem;
}

.checkbox {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-left: 2rem;
}

.order-symbol {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.order-symbol h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

/* 徽章样式 */
.side-badge {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.side-badge.buy {
  background: #d1fae5;
  color: #065f46;
}

.side-badge.sell {
  background: #fee2e2;
  color: #991b1b;
}

.status-badge {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-badge.pending {
  background: #fef3c7;
  color: #92400e;
}

.status-badge.filled {
  background: #d1fae5;
  color: #065f46;
}

.status-badge.cancelled,
.status-badge.expired,
.status-badge.rejected {
  background: #fee2e2;
  color: #991b1b;
}

/* 订单详情 */
.order-details {
  padding-left: 2rem;
  margin-bottom: 0.75rem;
}

.detail-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
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

.detail-value.highlight {
  color: var(--color-primary);
}

/* 时间信息 */
.order-time {
  padding-left: 2rem;
  margin-bottom: 0.75rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.time-icon {
  font-size: 0.875rem;
}

.time-separator {
  color: var(--color-border);
}

/* 操作按钮 */
.order-actions {
  padding-left: 2rem;
  display: flex;
  gap: 0.5rem;
}

/* 创建订单 */
.create-section {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
}

.section-header {
  margin-bottom: 1.5rem;
}

.section-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.order-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
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

/* 订单预览 */
.order-preview {
  background: var(--color-bg-secondary);
  border-radius: var(--radius-md);
  padding: 1rem;
}

.preview-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.75rem 0;
}

.preview-content {
  display: flex;
  gap: 2rem;
}

.preview-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.preview-label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.preview-value {
  font-size: 0.875rem;
  color: var(--color-text-primary);
  font-weight: 500;
}

.submit-btn {
  padding: 0.75rem 1.5rem;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
  align-self: flex-start;
}

.submit-btn:hover:not(:disabled) {
  background: var(--color-primary-hover);
}

.submit-btn:disabled {
  background: var(--color-secondary);
  cursor: not-allowed;
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
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-primary {
  background-color: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  background-color: var(--color-primary-hover);
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

.btn-danger {
  background-color: var(--color-danger);
  color: white;
}

.btn-danger:hover {
  background-color: #dc2626;
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.75rem;
}

.btn-loading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

/* 分页 */
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 1.5rem;
}

.page-btn {
  padding: 0.625rem 1rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.page-btn:hover:not(:disabled) {
  background: var(--color-bg-secondary);
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  text-align: center;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}

.page-total {
  color: var(--color-text-tertiary);
  font-size: 0.75rem;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 3rem 2rem;
  color: var(--color-text-tertiary);
}

.empty-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.empty-text {
  font-size: 1rem;
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

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.detail-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-group label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
  font-weight: 500;
}

.detail-group span {
  font-size: 0.875rem;
  color: var(--color-text-primary);
}

/* 加载动画 */
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
.slide-enter-active,
.slide-leave-active {
  transition: all 0.3s ease;
}

.slide-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

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

  .filter-section {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-tabs {
    width: 100%;
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .filter-controls {
    width: 100%;
  }

  .search-input {
    width: 100%;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .order-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
    padding-left: 1rem;
  }

  .order-details,
  .order-time,
  .order-actions {
    padding-left: 1rem;
  }

  .detail-row {
    grid-template-columns: 1fr 1fr;
  }

  .preview-content {
    flex-direction: column;
    gap: 1rem;
  }

  .pagination {
    flex-direction: column;
    gap: 1rem;
  }

  .modal-content {
    width: 95%;
  }
}
</style>
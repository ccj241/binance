<template>
  <div class="order-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">
        <span class="gradient-text">订单管理中心</span>
      </h1>
      <p class="page-subtitle">管理您的交易订单</p>
    </div>

    <!-- 筛选和操作区 -->
    <div class="filter-section">
      <div class="filter-tabs">
        <div class="tabs-wrapper">
          <button
              v-for="tab in filterTabs"
              :key="tab.value"
              @click="filterStatus = tab.value"
              :class="['filter-tab', { active: filterStatus === tab.value }]"
          >
            <span class="tab-icon">{{ tab.icon }}</span>
            <span class="tab-label">{{ tab.label }}</span>
            <span class="tab-count">{{ getFilterCount(tab.value) }}</span>
          </button>
          <div class="tab-indicator" :style="tabIndicatorStyle"></div>
        </div>
      </div>

      <!-- 批量操作 -->
      <div class="batch-actions" v-if="filterStatus === 'pending' && selectedOrders.length > 0">
        <div class="selected-info">
          <span class="selected-count">已选择 {{ selectedOrders.length }} 个订单</span>
        </div>
        <button @click="batchCancelOrders" class="batch-cancel-btn">
          <i>🗑️</i>
          批量取消
        </button>
      </div>
    </div>

    <!-- 订单统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card total">
        <div class="stat-icon">📊</div>
        <div class="stat-content">
          <div class="stat-value">{{ allOrdersCount }}</div>
          <div class="stat-label">总订单数</div>
        </div>
      </div>
      <div class="stat-card pending">
        <div class="stat-icon">⏳</div>
        <div class="stat-content">
          <div class="stat-value">{{ pendingOrdersCount }}</div>
          <div class="stat-label">待处理</div>
        </div>
      </div>
      <div class="stat-card filled">
        <div class="stat-icon">✅</div>
        <div class="stat-content">
          <div class="stat-value">{{ filledOrdersCount }}</div>
          <div class="stat-label">已成交</div>
        </div>
      </div>
      <div class="stat-card cancelled">
        <div class="stat-icon">❌</div>
        <div class="stat-content">
          <div class="stat-value">{{ cancelledOrdersCount }}</div>
          <div class="stat-label">已取消</div>
        </div>
      </div>
    </div>

    <!-- 订单列表 -->
    <div class="orders-section">
      <div class="section-header">
        <h2 class="section-title">订单列表</h2>
        <div class="controls">
          <div class="search-box">
            <i class="search-icon">🔍</i>
            <input
                v-model="searchQuery"
                type="text"
                placeholder="搜索订单..."
                class="search-input"
            >
          </div>
          <button @click="fetchOrders" class="refresh-btn">
            <i>🔄</i>
            刷新
          </button>
        </div>
      </div>

      <div v-if="filteredOrders.length === 0" class="empty-state">
        <div class="empty-icon">📋</div>
        <p class="empty-text">暂无{{ filterStatusText }}订单</p>
      </div>

      <div v-else class="orders-grid">
        <div v-for="order in paginatedOrders" :key="order.id" class="order-card">
          <!-- 选择框 -->
          <div v-if="filterStatus === 'pending'" class="order-select">
            <input type="checkbox"
                   :value="order.orderId || order.id"
                   v-model="selectedOrders" />
          </div>

          <!-- 订单头部 -->
          <div class="order-header">
            <div class="order-symbol">
              <span class="symbol-text">{{ order.symbol }}</span>
              <span :class="['side-badge', order.side.toLowerCase()]">
                {{ order.side === 'BUY' ? '买入' : '卖出' }}
              </span>
            </div>
            <div class="order-id">
              <span class="id-label">ID:</span>
              <span class="id-value">{{ order.orderId || order.id }}</span>
            </div>
          </div>

          <!-- 订单详情 -->
          <div class="order-details">
            <div class="detail-row">
              <div class="detail-item">
                <span class="detail-label">价格</span>
                <span class="detail-value price">{{ formatPrice(order.price) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">数量</span>
                <span class="detail-value">{{ formatQuantity(order.quantity) }}</span>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <span class="detail-label">状态</span>
                <span :class="['status-badge', order.status]">
                  {{ getStatusText(order.status) }}
                </span>
              </div>
              <div class="detail-item">
                <span class="detail-label">总值</span>
                <span class="detail-value total">{{ formatPrice(order.price * order.quantity) }}</span>
              </div>
            </div>
          </div>

          <!-- 时间信息 -->
          <div class="order-time">
            <div class="time-item">
              <span class="time-label">创建时间</span>
              <span class="time-value">{{ formatDate(order.createdAt) }}</span>
            </div>
            <div class="time-item">
              <span class="time-label">更新时间</span>
              <span class="time-value">{{ formatDate(order.updatedAt) }}</span>
            </div>
          </div>

          <!-- 操作按钮 -->
          <div class="order-actions">
            <button v-if="order.status === 'pending'"
                    @click="cancelOrder(order.orderId || order.id)"
                    class="action-btn cancel">
              <i>❌</i>
              取消订单
            </button>
            <button @click="viewOrderDetails(order)" class="action-btn view">
              <i>👁️</i>
              查看详情
            </button>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination" v-if="filteredOrders.length > pageSize">
        <button :disabled="currentPage === 1" @click="currentPage--" class="page-btn">
          <i>⬅️</i>
          上一页
        </button>
        <div class="page-info">
          <span>第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
          <span class="total-count">总计 {{ filteredOrders.length }} 条记录</span>
        </div>
        <button :disabled="currentPage === totalPages" @click="currentPage++" class="page-btn">
          下一页
          <i>➡️</i>
        </button>
      </div>
    </div>

    <!-- 创建订单表单 -->
    <div class="create-order-section">
      <h2 class="section-title">
        <span class="gradient-text">创建新订单</span>
      </h2>

      <div class="create-order-card">
        <form @submit.prevent="createOrder" class="order-form">
          <div class="form-grid">
            <div class="form-group">
              <label class="form-label">交易对</label>
              <input v-model="newOrder.symbol"
                     placeholder="如: BTCUSDT"
                     @input="newOrder.symbol = newOrder.symbol.toUpperCase()"
                     class="form-input"
                     required />
            </div>
            <div class="form-group">
              <label class="form-label">交易方向</label>
              <select v-model="newOrder.side" class="form-select" required>
                <option value="">选择方向</option>
                <option value="BUY">买入</option>
                <option value="SELL">卖出</option>
              </select>
            </div>
            <div class="form-group">
              <label class="form-label">价格</label>
              <input v-model.number="newOrder.price"
                     type="number"
                     step="0.00000001"
                     placeholder="限价单价格"
                     class="form-input"
                     required />
            </div>
            <div class="form-group">
              <label class="form-label">数量</label>
              <input v-model.number="newOrder.quantity"
                     type="number"
                     step="0.00000001"
                     placeholder="交易数量"
                     class="form-input"
                     required />
            </div>
          </div>

          <!-- 订单预览 -->
          <div class="order-preview" v-if="newOrder.price > 0 && newOrder.quantity > 0">
            <h3>订单预览</h3>
            <div class="preview-details">
              <div class="preview-item">
                <span>订单总值:</span>
                <span class="preview-value">{{ (newOrder.price * newOrder.quantity).toFixed(8) }} {{ getQuoteCurrency() }}</span>
              </div>
              <div class="preview-item">
                <span>预估手续费:</span>
                <span class="preview-value">{{ ((newOrder.price * newOrder.quantity) * 0.001).toFixed(8) }} {{ getQuoteCurrency() }}</span>
              </div>
            </div>
          </div>

          <button type="submit"
                  :disabled="isCreatingOrder || !isFormValid"
                  class="submit-btn">
            <span v-if="isCreatingOrder" class="loading-spinner">⏳</span>
            <span>{{ isCreatingOrder ? '创建中...' : '创建订单' }}</span>
          </button>
        </form>
      </div>
    </div>

    <!-- 订单详情弹窗 -->
    <div v-if="showOrderDetails" class="modal-overlay" @click="closeOrderDetails">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>订单详情</h3>
          <button @click="closeOrderDetails" class="close-btn">✕</button>
        </div>
        <div class="modal-body">
          <div class="detail-grid">
            <div class="detail-item">
              <label>订单ID:</label>
              <span>{{ selectedOrderDetails.orderId }}</span>
            </div>
            <div class="detail-item">
              <label>交易对:</label>
              <span>{{ selectedOrderDetails.symbol }}</span>
            </div>
            <div class="detail-item">
              <label>方向:</label>
              <span :class="['side-badge', selectedOrderDetails.side?.toLowerCase()]">
                {{ selectedOrderDetails.side === 'BUY' ? '买入' : '卖出' }}
              </span>
            </div>
            <div class="detail-item">
              <label>价格:</label>
              <span>{{ formatPrice(selectedOrderDetails.price) }}</span>
            </div>
            <div class="detail-item">
              <label>数量:</label>
              <span>{{ formatQuantity(selectedOrderDetails.quantity) }}</span>
            </div>
            <div class="detail-item">
              <label>状态:</label>
              <span :class="['status-badge', selectedOrderDetails.status]">
                {{ getStatusText(selectedOrderDetails.status) }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 消息提示 -->
    <transition name="toast">
      <div v-if="toastMessage" :class="['toast', toastType]">
        <i class="toast-icon">{{ toastType === 'success' ? '✅' : '❌' }}</i>
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
      pageSize: 12,
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
    },
    tabIndicatorStyle() {
      const index = this.filterTabs.findIndex(f => f.value === this.filterStatus);
      return {
        transform: `translateX(${index * 100}%)`
      };
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
/* 全局样式 */
.order-container {
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

/* 筛选标签 */
.filter-section {
  margin-bottom: 2rem;
}

.filter-tabs {
  margin-bottom: 1rem;
}

.tabs-wrapper {
  display: flex;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 16px;
  padding: 0.5rem;
  position: relative;
  gap: 0.5rem;
}

.filter-tab {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem;
  background: none;
  border: none;
  color: #999;
  cursor: pointer;
  border-radius: 12px;
  transition: all 0.3s ease;
  position: relative;
  z-index: 1;
}

.filter-tab:hover {
  color: #fff;
}

.filter-tab.active {
  color: #fff;
}

.tab-icon {
  font-size: 1.2rem;
}

.tab-count {
  background: rgba(255, 255, 255, 0.1);
  padding: 0.2rem 0.6rem;
  border-radius: 20px;
  font-size: 0.8rem;
}

.tab-indicator {
  position: absolute;
  top: 0.5rem;
  left: 0.5rem;
  width: calc(25% - 0.4rem);
  height: calc(100% - 1rem);
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  transition: transform 0.3s ease;
  z-index: 0;
}

/* 批量操作 */
.batch-actions {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  margin-top: 1rem;
}

.selected-info {
  color: #ccc;
}

.batch-cancel-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1.5rem;
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  color: white;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.batch-cancel-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(239, 68, 68, 0.4);
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  margin-bottom: 3rem;
}

.stat-card {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-5px);
  background: rgba(255, 255, 255, 0.08);
}

.stat-icon {
  font-size: 2rem;
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.1);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 0.25rem;
}

.stat-label {
  color: #999;
  font-size: 0.9rem;
}

/* 订单区域 */
.orders-section {
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

.controls {
  display: flex;
  gap: 1rem;
  align-items: center;
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

.refresh-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  color: #fff;
  cursor: pointer;
  transition: all 0.3s ease;
}

.refresh-btn:hover {
  background: rgba(255, 255, 255, 0.08);
}

/* 订单网格 */
.orders-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 1.5rem;
}

.order-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 1.5rem;
  transition: all 0.3s ease;
  position: relative;
}

.order-card:hover {
  background: rgba(255, 255, 255, 0.08);
  transform: translateY(-2px);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.order-select {
  position: absolute;
  top: 1rem;
  right: 1rem;
}

.order-select input[type="checkbox"] {
  width: 18px;
  height: 18px;
  cursor: pointer;
}

.order-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.order-symbol {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.symbol-text {
  font-size: 1.2rem;
  font-weight: 600;
}

.side-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
}

.side-badge.buy {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.side-badge.sell {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.order-id {
  color: #666;
  font-size: 0.9rem;
}

.id-label {
  color: #999;
}

.id-value {
  color: #ccc;
  font-family: monospace;
}

.order-details {
  margin-bottom: 1rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.75rem;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-label {
  color: #666;
  font-size: 0.8rem;
}

.detail-value {
  color: #ccc;
  font-weight: 500;
}

.detail-value.price {
  color: #fbbf24;
  font-family: monospace;
}

.detail-value.total {
  color: #22c55e;
  font-family: monospace;
}

.status-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-badge.pending {
  background: rgba(255, 193, 7, 0.2);
  color: #fbbf24;
  border: 1px solid rgba(255, 193, 7, 0.3);
}

.status-badge.filled {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.status-badge.cancelled,
.status-badge.expired,
.status-badge.rejected {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.order-time {
  display: flex;
  justify-content: space-between;
  margin-bottom: 1rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.time-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.time-label {
  color: #666;
  font-size: 0.8rem;
}

.time-value {
  color: #ccc;
  font-size: 0.9rem;
}

.order-actions {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  flex: 1;
  padding: 0.75rem;
  border: none;
  border-radius: 10px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.action-btn.cancel {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.action-btn.cancel:hover {
  background: rgba(239, 68, 68, 0.2);
}

.action-btn.view {
  background: rgba(108, 117, 125, 0.1);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
}

.action-btn.view:hover {
  background: rgba(108, 117, 125, 0.2);
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
}

/* 分页 */
.pagination {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 2rem;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 12px;
}

.page-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.5rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  color: #fff;
  cursor: pointer;
  transition: all 0.3s ease;
}

.page-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.08);
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.25rem;
  color: #ccc;
}

.total-count {
  font-size: 0.9rem;
  color: #666;
}

/* 创建订单区域 */
.create-order-section {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 24px;
  padding: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.create-order-card {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 16px;
  padding: 2rem;
}

.order-form {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  color: #ccc;
  font-weight: 500;
  font-size: 0.9rem;
}

.form-input,
.form-select {
  padding: 1rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  color: #fff;
  font-size: 1rem;
  transition: all 0.3s ease;
}

.form-input:focus,
.form-select:focus {
  outline: none;
  background: rgba(255, 255, 255, 0.08);
  border-color: #667eea;
  box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
}

.form-input::placeholder {
  color: #666;
}

.order-preview {
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  padding: 1.5rem;
}

.order-preview h3 {
  margin: 0 0 1rem 0;
  color: #ccc;
}

.preview-details {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.preview-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.preview-value {
  color: #22c55e;
  font-weight: 600;
  font-family: monospace;
}

.submit-btn {
  padding: 1rem 2rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  border-radius: 12px;
  color: white;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 10px 30px rgba(102, 126, 234, 0.3);
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
  transform: none;
}

.loading-spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* 弹窗样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(5px);
}

.modal-content {
  background: rgba(15, 15, 15, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.modal-header h3 {
  margin: 0;
  color: #fff;
}

.close-btn {
  background: none;
  border: none;
  color: #666;
  font-size: 1.5rem;
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 6px;
  transition: all 0.3s ease;
}

.close-btn:hover {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
}

.modal-body {
  padding: 1.5rem;
}

.detail-grid {
  display: grid;
  gap: 1rem;
}

.detail-grid .detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
}

.detail-grid .detail-item label {
  color: #666;
  font-weight: 500;
}

.detail-grid .detail-item span {
  color: #ccc;
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

/* 动画 */
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

/* 响应式设计 */
@media (max-width: 768px) {
  .order-container {
    padding: 1rem;
  }

  .page-title {
    font-size: 2rem;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .orders-grid {
    grid-template-columns: 1fr;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .section-header {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }

  .controls {
    flex-direction: column;
  }

  .search-box {
    width: 100%;
  }

  .pagination {
    flex-direction: column;
    gap: 1rem;
  }

  .tabs-wrapper {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .filter-tab {
    white-space: nowrap;
    min-width: 120px;
  }

  .tab-indicator {
    width: calc(25% - 0.4rem);
  }
}
</style>
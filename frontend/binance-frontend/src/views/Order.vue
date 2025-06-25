<template>
  <div class="order">
    <h1>订单管理</h1>

    <!-- 错误提示 -->
    <div v-if="errorMessage" class="error-message">
      {{ errorMessage }}
    </div>

    <!-- 成功提示 -->
    <div v-if="successMessage" class="success-message">
      {{ successMessage }}
    </div>

    <!-- 订单筛选 -->
    <div class="filter-section">
      <div class="filter-buttons">
        <button @click="filterStatus = 'all'" :class="{ active: filterStatus === 'all' }">
          全部订单 ({{ allOrdersCount }})
        </button>
        <button @click="filterStatus = 'pending'" :class="{ active: filterStatus === 'pending' }">
          待处理 ({{ pendingOrdersCount }})
        </button>
        <button @click="filterStatus = 'filled'" :class="{ active: filterStatus === 'filled' }">
          已成交 ({{ filledOrdersCount }})
        </button>
        <button @click="filterStatus = 'cancelled'" :class="{ active: filterStatus === 'cancelled' }">
          已取消 ({{ cancelledOrdersCount }})
        </button>
      </div>
      <div class="batch-actions" v-if="filterStatus === 'pending' && selectedOrders.length > 0">
        <button @click="batchCancelOrders" class="batch-cancel-btn">
          批量取消 ({{ selectedOrders.length }})
        </button>
      </div>
    </div>

    <!-- 订单列表 -->
    <div v-if="filteredOrders.length === 0" class="no-data">暂无{{ filterStatusText }}订单</div>
    <table v-else>
      <thead>
      <tr>
        <th v-if="filterStatus === 'pending'" class="checkbox-column">
          <input type="checkbox"
                 v-model="selectAll"
                 @change="toggleSelectAll" />
        </th>
        <th>订单ID</th>
        <th>交易对</th>
        <th>方向</th>
        <th>价格</th>
        <th>数量</th>
        <th>状态</th>
        <th>创建时间</th>
        <th>更新时间</th>
        <th>操作</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="order in paginatedOrders" :key="order.id">
        <td v-if="filterStatus === 'pending'" class="checkbox-column">
          <input type="checkbox"
                 :value="order.orderId || order.id"
                 v-model="selectedOrders" />
        </td>
        <td>{{ order.orderId || order.id }}</td>
        <td>{{ order.symbol }}</td>
        <td>
          <span :class="order.side === 'BUY' ? 'buy-side' : 'sell-side'">
            {{ order.side === 'BUY' ? '买入' : '卖出' }}
          </span>
        </td>
        <td>{{ formatPrice(order.price) }}</td>
        <td>{{ formatQuantity(order.quantity) }}</td>
        <td>
          <span :class="`status-${order.status}`">
            {{ getStatusText(order.status) }}
          </span>
        </td>
        <td>{{ formatDate(order.createdAt) }}</td>
        <td>{{ formatDate(order.updatedAt) }}</td>
        <td>
          <button v-if="order.status === 'pending'"
                  @click="cancelOrder(order.orderId || order.id)"
                  class="cancel-btn">
            取消
          </button>
          <span v-else class="no-action">-</span>
        </td>
      </tr>
      </tbody>
    </table>

    <div class="pagination" v-if="filteredOrders.length > pageSize">
      <button :disabled="currentPage === 1" @click="currentPage--">上一页</button>
      <span>第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
      <button :disabled="currentPage === totalPages" @click="currentPage++">下一页</button>
    </div>

    <!-- 创建订单表单 -->
    <section class="create-order">
      <h2>创建新订单</h2>
      <form @submit.prevent="createOrder">
        <div class="form-row">
          <div class="form-group">
            <label>交易对</label>
            <input v-model="newOrder.symbol"
                   placeholder="如: BTCUSDT"
                   @input="newOrder.symbol = newOrder.symbol.toUpperCase()"
                   required />
          </div>
          <div class="form-group">
            <label>交易方向</label>
            <select v-model="newOrder.side" required>
              <option value="">选择方向</option>
              <option value="BUY">买入</option>
              <option value="SELL">卖出</option>
            </select>
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>价格</label>
            <input v-model.number="newOrder.price"
                   type="number"
                   step="0.00000001"
                   placeholder="限价单价格"
                   required />
          </div>
          <div class="form-group">
            <label>数量</label>
            <input v-model.number="newOrder.quantity"
                   type="number"
                   step="0.00000001"
                   placeholder="交易数量"
                   required />
          </div>
        </div>
        <div class="order-summary" v-if="newOrder.price > 0 && newOrder.quantity > 0">
          <p>订单总值: {{ (newOrder.price * newOrder.quantity).toFixed(8) }} {{ getQuoteCurrency() }}</p>
        </div>
        <button type="submit" :disabled="isCreatingOrder || !isFormValid">
          {{ isCreatingOrder ? '创建中...' : '创建订单' }}
        </button>
      </form>
    </section>
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
      errorMessage: '',
      successMessage: '',
      isCreatingOrder: false,
      filterStatus: 'all',
      refreshInterval: null,
      selectedOrders: [],
      selectAll: false,
    };
  },
  computed: {
    filteredOrders() {
      if (this.filterStatus === 'all') {
        return this.orders;
      }
      if (this.filterStatus === 'cancelled') {
        return this.orders.filter(order =>
            ['cancelled', 'expired', 'rejected'].includes(order.status)
        );
      }
      return this.orders.filter(order => order.status === this.filterStatus);
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
      this.selectAll = false;
    }
  },
  mounted() {
    this.fetchOrders();
    // 每30秒刷新一次订单状态
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

    showMessage(message, isError = false) {
      if (isError) {
        this.errorMessage = message;
        this.successMessage = '';
      } else {
        this.successMessage = message;
        this.errorMessage = '';
      }

      setTimeout(() => {
        this.errorMessage = '';
        this.successMessage = '';
      }, 5000);
    },

    formatPrice(price) {
      return parseFloat(price).toFixed(8).replace(/\.?0+$/, '');
    },

    formatQuantity(quantity) {
      return parseFloat(quantity).toFixed(8).replace(/\.?0+$/, '');
    },

    formatDate(dateString) {
      if (!dateString) return '-';
      return new Date(dateString).toLocaleString('zh-CN');
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
      // 简单判断，假设最后4个字符是报价货币
      if (this.newOrder.symbol.endsWith('USDT')) return 'USDT';
      if (this.newOrder.symbol.endsWith('BTC')) return 'BTC';
      if (this.newOrder.symbol.endsWith('ETH')) return 'ETH';
      if (this.newOrder.symbol.endsWith('BNB')) return 'BNB';
      return '';
    },

    async fetchOrders() {
      try {
        const response = await axios.get('/orders', {
          headers: this.getAuthHeaders(),
        });
        this.orders = response.data.orders || [];

        // 按创建时间倒序排序
        this.orders.sort((a, b) =>
            new Date(b.createdAt) - new Date(a.createdAt)
        );
      } catch (error) {
        console.error('获取订单失败:', error);
        this.showMessage(error.response?.data?.error || '获取订单失败', true);
      }
    },

    async createOrder() {
      if (!this.isFormValid) {
        this.showMessage('请填写所有必需字段', true);
        return;
      }

      this.isCreatingOrder = true;
      try {
        const response = await axios.post('/order', this.newOrder, {
          headers: this.getAuthHeaders(),
        });

        this.showMessage(response.data.message || '订单创建成功');
        this.newOrder = { symbol: '', side: '', price: 0, quantity: 0 };
        await this.fetchOrders(); // 刷新订单列表
      } catch (error) {
        console.error('创建订单失败:', error);
        this.showMessage(error.response?.data?.error || '创建订单失败', true);
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

        this.showMessage(response.data.message || '订单取消成功');
        await this.fetchOrders(); // 刷新订单列表
      } catch (error) {
        console.error('取消订单失败:', error);
        this.showMessage(error.response?.data?.error || '取消订单失败', true);
      }
    },

    toggleSelectAll() {
      if (this.selectAll) {
        // 选择当前页面所有待处理订单
        this.selectedOrders = this.paginatedOrders
            .filter(order => order.status === 'pending')
            .map(order => order.orderId || order.id);
      } else {
        this.selectedOrders = [];
      }
    },

    async batchCancelOrders() {
      if (this.selectedOrders.length === 0) {
        return;
      }

      const count = this.selectedOrders.length;
      if (!window.confirm(`确定要批量取消 ${count} 个订单吗？`)) {
        return;
      }

      try {
        // 使用批量取消API
        const response = await axios.post('/batch_cancel_orders', {
          orderIds: this.selectedOrders.map(id => parseInt(id))
        }, {
          headers: this.getAuthHeaders(),
        });

        // 显示结果
        const results = response.data.results;
        if (results && results.failed && results.failed.length === 0) {
          this.showMessage(`成功取消 ${results.success.length} 个订单`);
        } else if (results) {
          let message = response.data.message;
          if (results.failed && results.failed.length > 0) {
            message += '\n失败详情：\n' + results.failed
                .map(e => `订单 ${e.orderId}: ${e.error}`)
                .join('\n');
          }
          this.showMessage(message, results.failed.length > 0);
        } else {
          this.showMessage(response.data.message || '批量取消完成');
        }
      } catch (error) {
        console.error('批量取消订单失败:', error);
        this.showMessage(error.response?.data?.error || '批量取消订单失败', true);
      }

      // 清空选择并刷新列表
      this.selectedOrders = [];
      this.selectAll = false;
      await this.fetchOrders();
    },
  },
};
</script>

<style scoped>
.order {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.filter-section {
  margin: 20px 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.filter-buttons {
  display: flex;
  gap: 10px;
}

.batch-actions {
  display: flex;
  gap: 10px;
}

.filter-section button {
  padding: 8px 16px;
  border: 1px solid #ddd;
  background: #f5f5f5;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.3s;
}

.filter-section button:hover {
  background: #e0e0e0;
}

.filter-section button.active {
  background: #007bff;
  color: white;
  border-color: #007bff;
}

.batch-cancel-btn {
  background-color: #dc3545 !important;
  color: white !important;
  border-color: #dc3545 !important;
}

.batch-cancel-btn:hover {
  background-color: #c82333 !important;
}

.checkbox-column {
  width: 50px;
  text-align: center;
}

.checkbox-column input[type="checkbox"] {
  cursor: pointer;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 10px;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

th, td {
  border: 1px solid #ddd;
  padding: 12px;
  text-align: left;
}

th {
  background-color: #f2f2f2;
  font-weight: bold;
}

tr:hover {
  background-color: #f5f5f5;
}

.create-order {
  margin-top: 40px;
  padding: 30px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #f9f9f9;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.form-row {
  display: flex;
  gap: 15px;
  margin-bottom: 15px;
}

.form-group {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.form-group label {
  font-weight: bold;
  margin-bottom: 5px;
  color: #333;
}

input, select, button {
  padding: 10px;
  font-size: 14px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

button {
  background-color: #007bff;
  color: white;
  border: none;
  cursor: pointer;
  transition: all 0.3s;
}

button:hover {
  background-color: #0056b3;
  transform: translateY(-1px);
  box-shadow: 0 2px 4px rgba(0,0,0,0.2);
}

button:disabled {
  background-color: #6c757d;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.cancel-btn {
  background-color: #dc3545;
  padding: 6px 12px;
  font-size: 12px;
}

.cancel-btn:hover {
  background-color: #c82333;
}

.no-action {
  color: #999;
  font-style: italic;
}

.no-data {
  color: #888;
  font-style: italic;
  padding: 40px;
  text-align: center;
  font-size: 18px;
}

.pagination {
  margin-top: 20px;
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: center;
}

.pagination span {
  font-size: 14px;
}

.error-message {
  background-color: #f8d7da;
  color: #721c24;
  padding: 15px;
  border: 1px solid #f5c6cb;
  border-radius: 4px;
  margin-bottom: 20px;
}

.success-message {
  background-color: #d4edda;
  color: #155724;
  padding: 15px;
  border: 1px solid #c3e6cb;
  border-radius: 4px;
  margin-bottom: 20px;
}

.order-summary {
  background-color: #e9ecef;
  padding: 10px;
  border-radius: 4px;
  margin-bottom: 15px;
}

.order-summary p {
  margin: 0;
  font-weight: bold;
  color: #495057;
}

input:focus, select:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0,123,255,0.25);
}

.buy-side {
  color: #28a745;
  font-weight: bold;
}

.sell-side {
  color: #dc3545;
  font-weight: bold;
}

.status-pending {
  color: #ffc107;
  font-weight: bold;
}

.status-filled {
  color: #28a745;
  font-weight: bold;
}

.status-cancelled,
.status-expired,
.status-rejected {
  color: #dc3545;
  font-weight: bold;
}
</style>
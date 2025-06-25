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

    <!-- 订单列表 -->
    <div v-if="orders.length === 0" class="no-data">无订单记录</div>
    <table v-else>
      <thead>
      <tr>
        <th>交易对</th>
        <th>方向</th>
        <th>价格</th>
        <th>数量</th>
        <th>状态</th>
        <th>时间</th>
        <th>操作</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="order in paginatedOrders" :key="order.id">
        <td>{{ order.symbol }}</td>
        <td>{{ order.side }}</td>
        <td>{{ order.price }}</td>
        <td>{{ order.quantity }}</td>
        <td>{{ order.status }}</td>
        <td>{{ new Date(order.createdAt).toLocaleString() }}</td>
        <td>
          <button v-if="order.status === 'pending'"
                  @click="cancelOrder(order.orderId)"
                  class="cancel-btn">
            取消
          </button>
        </td>
      </tr>
      </tbody>
    </table>

    <div class="pagination" v-if="orders.length > pageSize">
      <button :disabled="currentPage === 1" @click="currentPage--">上一页</button>
      <span>第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
      <button :disabled="currentPage === totalPages" @click="currentPage++">下一页</button>
    </div>

    <!-- 创建订单表单 -->
    <section class="create-order">
      <h2>创建新订单</h2>
      <form @submit.prevent="createOrder">
        <div class="form-row">
          <input v-model="newOrder.symbol" placeholder="交易对 (如 BTCUSDT)" required />
          <select v-model="newOrder.side" required>
            <option value="">选择方向</option>
            <option value="BUY">买入</option>
            <option value="SELL">卖出</option>
          </select>
        </div>
        <div class="form-row">
          <input v-model.number="newOrder.price" type="number" step="0.00000001" placeholder="价格" required />
          <input v-model.number="newOrder.quantity" type="number" step="0.00000001" placeholder="数量" required />
        </div>
        <button type="submit" :disabled="isCreatingOrder">
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
    };
  },
  computed: {
    paginatedOrders() {
      const start = (this.currentPage - 1) * this.pageSize;
      const end = start + this.pageSize;
      return this.orders.slice(start, end);
    },
    totalPages() {
      return Math.ceil(this.orders.length / this.pageSize);
    },
  },
  mounted() {
    this.fetchOrders();
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

    async fetchOrders() {
      try {
        const response = await axios.get('/orders', {
          headers: this.getAuthHeaders(),
        });
        this.orders = response.data.orders || [];
        this.currentPage = 1;
      } catch (error) {
        console.error('获取订单失败:', error);
        this.showMessage(error.response?.data?.error || '获取订单失败', true);
      }
    },

    async createOrder() {
      if (!this.newOrder.symbol || !this.newOrder.side ||
          this.newOrder.price <= 0 || this.newOrder.quantity <= 0) {
        this.showMessage('请填写所有必需字段，且价格和数量必须大于0', true);
        return;
      }

      this.isCreatingOrder = true;
      try {
        const response = await axios.post('/order', this.newOrder, {
          headers: this.getAuthHeaders(),
        });

        this.showMessage(response.data.message || '订单创建成功');
        this.newOrder = { symbol: '', side: '', price: 0, quantity: 0 };
        this.fetchOrders(); // 刷新订单列表
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
        this.fetchOrders(); // 刷新订单列表
      } catch (error) {
        console.error('取消订单失败:', error);
        this.showMessage(error.response?.data?.error || '取消订单失败', true);
      }
    },
  },
};
</script>

<style scoped>
.order {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 10px;
}

th, td {
  border: 1px solid #ddd;
  padding: 8px;
  text-align: left;
}

th {
  background-color: #f2f2f2;
}

.create-order {
  margin-top: 40px;
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #f9f9f9;
}

.form-row {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
}

input, select, button {
  padding: 8px;
  font-size: 14px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

button {
  background-color: #007bff;
  color: white;
  border: none;
  cursor: pointer;
}

button:hover {
  background-color: #0056b3;
}

button:disabled {
  background-color: #6c757d;
  cursor: not-allowed;
}

.cancel-btn {
  background-color: #dc3545;
  padding: 4px 8px;
  font-size: 12px;
}

.cancel-btn:hover {
  background-color: #c82333;
}

.no-data {
  color: #888;
  font-style: italic;
  padding: 20px;
  text-align: center;
}

.pagination {
  margin-top: 10px;
  display: flex;
  gap: 10px;
  align-items: center;
}

.pagination span {
  font-size: 14px;
}

.error-message {
  background-color: #f8d7da;
  color: #721c24;
  padding: 12px;
  border: 1px solid #f5c6cb;
  border-radius: 4px;
  margin-bottom: 20px;
}

.success-message {
  background-color: #d4edda;
  color: #155724;
  padding: 12px;
  border: 1px solid #c3e6cb;
  border-radius: 4px;
  margin-bottom: 20px;
}

input:focus, select:focus {
  outline: none;
  border-color: #007bff;
}

select {
  background-color: white;
}
</style>
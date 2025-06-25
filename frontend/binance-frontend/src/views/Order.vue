<template>
  <div class="order">
    <h1>订单管理</h1>
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
      </tr>
      </tbody>
    </table>
    <div class="pagination" v-if="orders.length > pageSize">
      <button :disabled="currentPage === 1" @click="currentPage--">上一页</button>
      <span>第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
      <button :disabled="currentPage === totalPages" @click="currentPage++">下一页</button>
    </div>
    <form @submit.prevent="createOrder">
      <input v-model="newOrder.symbol" placeholder="交易对" required />
      <input v-model="newOrder.side" placeholder="方向 (BUY/SELL)" required />
      <input v-model.number="newOrder.price" type="number" placeholder="价格" required />
      <input v-model.number="newOrder.quantity" type="number" placeholder="数量" required />
      <button type="submit">创建订单</button>
    </form>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'Order',
  data() {
    return {
      orders: [],
      newOrder: { symbol: '', side: '', price: 0, quantity: 0 },
      currentPage: 1,
      pageSize: 10,
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
    async fetchOrders() {
      try {
        const response = await axios.get('/orders', {
          headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
        });
        this.orders = response.data.orders || [];
        this.currentPage = 1;
      } catch (error) {
        console.error('获取订单失败:', error);
      }
    },
    async createOrder() {
      try {
        await axios.post('/order', this.newOrder, {
          headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
        });
        this.newOrder = { symbol: '', side: '', price: 0, quantity: 0 };
        this.fetchOrders();
      } catch (error) {
        console.error('创建订单失败:', error);
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

form {
  display: flex;
  gap: 10px;
  margin-top: 10px;
}

input, button {
  padding: 8px;
  font-size: 14px;
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

.no-data {
  color: #888;
  font-style: italic;
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
</style>
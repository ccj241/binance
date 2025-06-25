<template>
  <div class="strategy">
    <h1>策略管理</h1>
    <div v-if="strategies.length === 0" class="no-data">无策略记录</div>
    <table v-else>
      <thead>
      <tr>
        <th>交易对</th>
        <th>类型</th>
        <th>方向</th>
        <th>价格</th>
        <th>数量</th>
        <th>状态</th>
        <th>启用</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="strategy in paginatedStrategies" :key="strategy.id">
        <td>{{ strategy.symbol }}</td>
        <td>{{ strategy.strategyType }}</td>
        <td>{{ strategy.side }}</td>
        <td>{{ strategy.price }}</td>
        <td>{{ strategy.totalQuantity }}</td>
        <td>{{ strategy.status }}</td>
        <td>{{ strategy.enabled ? '是' : '否' }}</td>
      </tr>
      </tbody>
    </table>
    <div class="pagination" v-if="strategies.length > pageSize">
      <button :disabled="currentPage === 1" @click="currentPage--">上一页</button>
      <span>第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
      <button :disabled="currentPage === totalPages" @click="currentPage++">下一页</button>
    </div>
    <form @submit.prevent="createStrategy">
      <input v-model="newStrategy.symbol" placeholder="交易对" required />
      <input v-model="newStrategy.strategyType" placeholder="策略类型 (simple/iceberg/custom)" required />
      <input v-model="newStrategy.side" placeholder="方向 (BUY/SELL)" required />
      <input v-model.number="newStrategy.price" type="number" placeholder="价格" required />
      <input v-model.number="newStrategy.totalQuantity" type="number" placeholder="总数量" required />
      <button type="submit">创建策略</button>
    </form>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'Strategy',
  data() {
    return {
      strategies: [],
      newStrategy: { symbol: '', strategyType: '', side: '', price: 0, totalQuantity: 0 },
      currentPage: 1,
      pageSize: 10,
    };
  },
  computed: {
    paginatedStrategies() {
      const start = (this.currentPage - 1) * this.pageSize;
      const end = start + this.pageSize;
      return this.strategies.slice(start, end);
    },
    totalPages() {
      return Math.ceil(this.strategies.length / this.pageSize);
    },
  },
  mounted() {
    this.fetchStrategies();
  },
  methods: {
    async fetchStrategies() {
      try {
        const response = await axios.get('/strategies', {
          headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
        });
        this.strategies = response.data.strategies || [];
        this.currentPage = 1;
      } catch (error) {
        console.error('获取策略失败:', error);
      }
    },
    async createStrategy() {
      try {
        await axios.post('/strategy', this.newStrategy, {
          headers: { Authorization: `Bearer ${localStorage.getItem('token')}` },
        });
        this.newStrategy = { symbol: '', strategyType: '', side: '', price: 0, totalQuantity: 0 };
        this.fetchStrategies();
      } catch (error) {
        console.error('创建策略失败:', error);
      }
    },
  },
};
</script>

<style scoped>
.strategy {
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
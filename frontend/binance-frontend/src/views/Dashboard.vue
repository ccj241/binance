<template>
  <div class="dashboard">
    <h1>仪表盘</h1>

    <!-- 实时价格 -->
    <section>
      <h2>实时价格</h2>
      <div v-if="Object.keys(prices).length === 0" class="no-data">未添加交易对</div>
      <table v-else>
        <thead>
        <tr>
          <th>交易对</th>
          <th>价格</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="(price, symbol) in prices" :key="symbol">
          <td>{{ symbol }}</td>
          <td>{{ price.toFixed(2) }}</td>
        </tr>
        </tbody>
      </table>
      <form @submit.prevent="addSymbol">
        <input v-model="newSymbol" placeholder="输入交易对 (如 SOLUSDT)" required />
        <button type="submit">添加交易对</button>
      </form>
    </section>

    <!-- 余额 -->
    <section>
      <h2>账户余额</h2>
      <div v-if="balances.length === 0" class="no-data">无可用余额</div>
      <table v-else>
        <thead>
        <tr>
          <th>资产</th>
          <th>可用</th>
          <th>锁定</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="balance in balances" :key="balance.asset">
          <td>{{ balance.asset }}</td>
          <td>{{ balance.free }}</td>
          <td>{{ balance.locked }}</td>
        </tr>
        </tbody>
      </table>
    </section>

    <!-- 交易记录 -->
    <section>
      <h2>交易记录</h2>
      <div v-if="trades.length === 0" class="no-data">无交易记录</div>
      <table v-else>
        <thead>
        <tr>
          <th>交易对</th>
          <th>价格</th>
          <th>数量</th>
          <th>时间</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="trade in paginatedTrades" :key="trade.id">
          <td>{{ trade.symbol }}</td>
          <td>{{ trade.price }}</td>
          <td>{{ trade.qty }}</td>
          <td>{{ new Date(trade.time).toLocaleString() }}</td>
        </tr>
        </tbody>
      </table>
      <div class="pagination" v-if="trades.length > pageSize">
        <button :disabled="currentPage === 1" @click="currentPage--">上一页</button>
        <span>第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
        <button :disabled="currentPage === totalPages" @click="currentPage++">下一页</button>
      </div>
    </section>

    <!-- 取款历史 -->
    <section>
      <h2>取款历史</h2>
      <div v-if="withdrawalHistory.length === 0" class="no-data">无取款历史</div>
      <table v-else>
        <thead>
        <tr>
          <th>资产</th>
          <th>金额</th>
          <th>地址</th>
          <th>取款 ID</th>
          <th>状态</th>
          <th>时间</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="record in withdrawalHistory" :key="record.id">
          <td>{{ record.asset }}</td>
          <td>{{ record.amount }}</td>
          <td>{{ record.address }}</td>
          <td>{{ record.withdrawalId }}</td>
          <td>{{ record.status }}</td>
          <td>{{ new Date(record.createdAt).toLocaleString() }}</td>
        </tr>
        </tbody>
      </table>
    </section>

    <!-- 策略 -->
    <section>
      <h2>交易策略</h2>
      <div v-if="strategies.length === 0" class="no-data">无可用策略</div>
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
        <tr v-for="strategy in strategies" :key="strategy.id">
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
    </section>

    <!-- 错误提示 -->
    <div v-if="errorMessage" class="error-message">
      {{ errorMessage }}
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'Dashboard',
  data() {
    return {
      prices: {},
      balances: [],
      trades: [],
      withdrawalHistory: [],
      strategies: [],
      newSymbol: '',
      currentPage: 1,
      pageSize: 10,
      errorMessage: '',
      priceInterval: null,
    };
  },
  computed: {
    paginatedTrades() {
      const start = (this.currentPage - 1) * this.pageSize;
      const end = start + this.pageSize;
      return this.trades.slice(start, end);
    },
    totalPages() {
      return Math.ceil(this.trades.length / this.pageSize);
    },
  },
  mounted() {
    this.fetchPrices();
    this.fetchBalances();
    this.fetchTrades();
    this.fetchWithdrawalHistory();
    this.fetchStrategies();
    this.priceInterval = setInterval(this.fetchPrices, 10000);
  },
  beforeUnmount() {
    if (this.priceInterval) {
      clearInterval(this.priceInterval);
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

    handleApiError(error, defaultMessage = '操作失败') {
      console.error(defaultMessage + ':', error);
      this.errorMessage = error.response?.data?.error || error.response?.data?.message || defaultMessage;
      setTimeout(() => {
        this.errorMessage = '';
      }, 5000);
    },

    async fetchPrices() {
      try {
        const response = await axios.get('/prices', {
          headers: this.getAuthHeaders(),
        });
        this.prices = response.data.prices || {};
      } catch (error) {
        this.handleApiError(error, '获取价格失败');
      }
    },

    async fetchBalances() {
      try {
        const response = await axios.get('/balance', {
          headers: this.getAuthHeaders(),
        });
        this.balances = response.data.balances || [];
      } catch (error) {
        this.handleApiError(error, '获取余额失败');
      }
    },

    async fetchTrades() {
      try {
        const response = await axios.get('/trades', {
          headers: this.getAuthHeaders(),
        });
        this.trades = response.data.trades || [];
        this.currentPage = 1;
      } catch (error) {
        this.handleApiError(error, '获取交易记录失败');
      }
    },

    async fetchWithdrawalHistory() {
      try {
        const response = await axios.get('/withdrawalhistory', {
          headers: this.getAuthHeaders(),
        });
        this.withdrawalHistory = response.data.history || [];
      } catch (error) {
        this.handleApiError(error, '获取取款历史失败');
      }
    },

    async fetchStrategies() {
      try {
        const response = await axios.get('/strategies', {
          headers: this.getAuthHeaders(),
        });
        this.strategies = response.data.strategies || [];
      } catch (error) {
        this.handleApiError(error, '获取策略失败');
      }
    },

    async addSymbol() {
      if (!this.newSymbol.trim()) {
        this.errorMessage = '请输入有效的交易对';
        return;
      }

      try {
        await axios.post('/symbols',
            { symbol: this.newSymbol.toUpperCase() },
            { headers: this.getAuthHeaders() }
        );
        this.newSymbol = '';
        this.fetchPrices();
        this.errorMessage = '';
      } catch (error) {
        this.handleApiError(error, '添加交易对失败');
      }
    },
  },
};
</script>

<style scoped>
.dashboard {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

section {
  margin-bottom: 40px;
}

h1, h2, h3 {
  color: #333;
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
  border-radius: 4px;
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
  margin-top: 20px;
}

input {
  border: 1px solid #ddd;
  border-radius: 4px;
}

input:focus {
  outline: none;
  border-color: #007bff;
}
</style>
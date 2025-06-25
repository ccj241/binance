<template>
  <div class="strategy">
    <h1>策略管理</h1>

    <!-- 错误提示 -->
    <div v-if="errorMessage" class="error-message">
      {{ errorMessage }}
    </div>

    <!-- 成功提示 -->
    <div v-if="successMessage" class="success-message">
      {{ successMessage }}
    </div>

    <!-- 策略列表 -->
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
        <th>操作</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="strategy in paginatedStrategies" :key="strategy.id">
        <td>{{ strategy.symbol }}</td>
        <td>{{ getStrategyTypeText(strategy.strategyType) }}</td>
        <td>{{ strategy.side === 'BUY' ? '买入' : '卖出' }}</td>
        <td>{{ strategy.price }}</td>
        <td>{{ strategy.totalQuantity }}</td>
        <td>{{ getStatusText(strategy.status) }}</td>
        <td>{{ strategy.enabled ? '是' : '否' }}</td>
        <td>
          <button @click="toggleStrategy(strategy.id)" class="toggle-btn">
            {{ strategy.enabled ? '禁用' : '启用' }}
          </button>
          <button @click="deleteStrategy(strategy.id)" class="delete-btn">
            删除
          </button>
        </td>
      </tr>
      </tbody>
    </table>

    <div class="pagination" v-if="strategies.length > pageSize">
      <button :disabled="currentPage === 1" @click="currentPage--">上一页</button>
      <span>第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
      <button :disabled="currentPage === totalPages" @click="currentPage++">下一页</button>
    </div>

    <!-- 创建策略表单 -->
    <section class="create-strategy">
      <h2>创建新策略</h2>
      <form @submit.prevent="createStrategy">
        <div class="form-row">
          <input v-model="newStrategy.symbol" placeholder="交易对 (如 BTCUSDT)" required />
          <select v-model="newStrategy.strategyType" required>
            <option value="">选择策略类型</option>
            <option value="simple">简单策略</option>
            <option value="iceberg">冰山策略</option>
            <option value="custom">自定义策略</option>
          </select>
        </div>
        <div class="form-row">
          <select v-model="newStrategy.side" required>
            <option value="">选择方向</option>
            <option value="BUY">买入</option>
            <option value="SELL">卖出</option>
          </select>
          <input v-model.number="newStrategy.price" type="number" step="0.00000001" placeholder="触发价格" required />
          <input v-model.number="newStrategy.totalQuantity" type="number" step="0.00000001" placeholder="总数量" required />
        </div>

        <!-- 自定义策略的额外配置 -->
        <div v-if="newStrategy.strategyType === 'custom'" class="custom-config">
          <h3>自定义配置</h3>
          <div class="form-row">
            <input v-model="buyQuantitiesInput" placeholder="买入数量比例 (如: 0.3,0.3,0.2,0.2)" />
            <input v-model="buyDepthLevelsInput" placeholder="买入深度级别 (如: 1,3,5,7)" />
          </div>
          <div class="form-row">
            <input v-model="sellQuantitiesInput" placeholder="卖出数量比例 (如: 0.3,0.3,0.2,0.2)" />
            <input v-model="sellDepthLevelsInput" placeholder="卖出深度级别 (如: 1,3,5,7)" />
          </div>
        </div>

        <button type="submit" :disabled="isCreatingStrategy">
          {{ isCreatingStrategy ? '创建中...' : '创建策略' }}
        </button>
      </form>
    </section>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'Strategy',
  data() {
    return {
      strategies: [],
      newStrategy: {
        symbol: '',
        strategyType: '',
        side: '',
        price: 0,
        totalQuantity: 0
      },
      buyQuantitiesInput: '',
      buyDepthLevelsInput: '',
      sellQuantitiesInput: '',
      sellDepthLevelsInput: '',
      currentPage: 1,
      pageSize: 10,
      errorMessage: '',
      successMessage: '',
      isCreatingStrategy: false,
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

    getStrategyTypeText(type) {
      const types = {
        'simple': '简单策略',
        'iceberg': '冰山策略',
        'custom': '自定义策略'
      };
      return types[type] || type;
    },

    getStatusText(status) {
      const statuses = {
        'active': '活跃',
        'inactive': '非活跃',
        'completed': '已完成',
        'cancelled': '已取消'
      };
      return statuses[status] || status;
    },

    async fetchStrategies() {
      try {
        const response = await axios.get('/strategies', {
          headers: this.getAuthHeaders(),
        });
        this.strategies = response.data.strategies || [];
        this.currentPage = 1;
      } catch (error) {
        console.error('获取策略失败:', error);
        this.showMessage(error.response?.data?.error || '获取策略失败', true);
      }
    },

    async createStrategy() {
      if (!this.newStrategy.symbol || !this.newStrategy.strategyType ||
          !this.newStrategy.side || this.newStrategy.price <= 0 ||
          this.newStrategy.totalQuantity <= 0) {
        this.showMessage('请填写所有必需字段，且价格和数量必须大于0', true);
        return;
      }

      this.isCreatingStrategy = true;
      try {
        const strategyData = { ...this.newStrategy };

        // 如果是自定义策略，添加自定义配置
        if (this.newStrategy.strategyType === 'custom') {
          if (this.newStrategy.side === 'BUY') {
            if (!this.buyQuantitiesInput || !this.buyDepthLevelsInput) {
              this.showMessage('买入策略需要填写买入数量比例和深度级别', true);
              return;
            }
            strategyData.buyQuantities = this.buyQuantitiesInput.split(',').map(v => parseFloat(v.trim()));
            strategyData.buyDepthLevels = this.buyDepthLevelsInput.split(',').map(v => parseInt(v.trim()));
          } else {
            if (!this.sellQuantitiesInput || !this.sellDepthLevelsInput) {
              this.showMessage('卖出策略需要填写卖出数量比例和深度级别', true);
              return;
            }
            strategyData.sellQuantities = this.sellQuantitiesInput.split(',').map(v => parseFloat(v.trim()));
            strategyData.sellDepthLevels = this.sellDepthLevelsInput.split(',').map(v => parseInt(v.trim()));
          }
        }

        const response = await axios.post('/strategy', strategyData, {
          headers: this.getAuthHeaders(),
        });

        this.showMessage(response.data.message || '策略创建成功');
        this.resetForm();
        this.fetchStrategies(); // 刷新策略列表
      } catch (error) {
        console.error('创建策略失败:', error);
        this.showMessage(error.response?.data?.error || '创建策略失败', true);
      } finally {
        this.isCreatingStrategy = false;
      }
    },

    async toggleStrategy(strategyId) {
      try {
        const response = await axios.post('/toggle_strategy', { id: strategyId }, {
          headers: this.getAuthHeaders(),
        });

        this.showMessage(response.data.message || '策略状态切换成功');
        this.fetchStrategies(); // 刷新策略列表
      } catch (error) {
        console.error('切换策略失败:', error);
        this.showMessage(error.response?.data?.error || '切换策略失败', true);
      }
    },

    async deleteStrategy(strategyId) {
      if (!window.confirm('确定要删除这个策略吗？')) {
        return;
      }

      try {
        const response = await axios.post('/delete_strategy', { id: strategyId }, {
          headers: this.getAuthHeaders(),
        });

        this.showMessage(response.data.message || '策略删除成功');
        this.fetchStrategies(); // 刷新策略列表
      } catch (error) {
        console.error('删除策略失败:', error);
        this.showMessage(error.response?.data?.error || '删除策略失败', true);
      }
    },

    resetForm() {
      this.newStrategy = { symbol: '', strategyType: '', side: '', price: 0, totalQuantity: 0 };
      this.buyQuantitiesInput = '';
      this.buyDepthLevelsInput = '';
      this.sellQuantitiesInput = '';
      this.sellDepthLevelsInput = '';
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

.create-strategy {
  margin-top: 40px;
  padding: 20px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #f9f9f9;
}

.custom-config {
  margin-top: 20px;
  padding: 15px;
  border: 1px solid #ccc;
  border-radius: 4px;
  background-color: #fff;
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

.toggle-btn {
  background-color: #28a745;
  padding: 4px 8px;
  font-size: 12px;
  margin-right: 5px;
}

.toggle-btn:hover {
  background-color: #218838;
}

.delete-btn {
  background-color: #dc3545;
  padding: 4px 8px;
  font-size: 12px;
}

.delete-btn:hover {
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

h3 {
  margin-top: 0;
  color: #333;
}
</style>
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
        <th>基准价格</th>
        <th>总数量</th>
        <th>状态</th>
        <th>启用</th>
        <th>执行状态</th>
        <th>创建时间</th>
        <th>操作</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="strategy in paginatedStrategies" :key="strategy.id">
        <td>{{ strategy.symbol }}</td>
        <td>{{ getStrategyTypeText(strategy.strategyType) }}</td>
        <td>
          <span :class="strategy.side === 'BUY' ? 'buy-side' : 'sell-side'">
            {{ strategy.side === 'BUY' ? '买入' : '卖出' }}
          </span>
        </td>
        <td>{{ formatPrice(strategy.price) }}</td>
        <td>{{ strategy.totalQuantity }}</td>
        <td>
          <span :class="`status-${strategy.status}`">
            {{ getStatusText(strategy.status) }}
          </span>
        </td>
        <td>
          <span :class="strategy.enabled ? 'enabled' : 'disabled'">
            {{ strategy.enabled ? '启用' : '禁用' }}
          </span>
        </td>
        <td>
          <span :class="strategy.pendingBatch ? 'executing' : 'idle'">
            {{ strategy.pendingBatch ? '执行中' : '空闲' }}
          </span>
        </td>
        <td>{{ formatDate(strategy.createdAt) }}</td>
        <td>
          <button @click="toggleStrategy(strategy)"
                  :class="strategy.enabled ? 'disable-btn' : 'enable-btn'">
            {{ strategy.enabled ? '禁用' : '启用' }}
          </button>
          <button @click="viewStrategyDetails(strategy)" class="view-btn">
            查看
          </button>
          <button @click="viewStrategyStats(strategy)" class="stats-btn">
            统计
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

    <!-- 策略详情弹窗 -->
    <div v-if="showDetails" class="modal-overlay" @click="closeDetails">
      <div class="modal-content" @click.stop>
        <h3>策略详情</h3>
        <div class="detail-item">
          <label>策略ID:</label>
          <span>{{ selectedStrategy.id }}</span>
        </div>
        <div class="detail-item">
          <label>交易对:</label>
          <span>{{ selectedStrategy.symbol }}</span>
        </div>
        <div class="detail-item">
          <label>策略类型:</label>
          <span>{{ getStrategyTypeText(selectedStrategy.strategyType) }}</span>
        </div>
        <div class="detail-item">
          <label>方向:</label>
          <span>{{ selectedStrategy.side === 'BUY' ? '买入' : '卖出' }}</span>
        </div>
        <div class="detail-item">
          <label>基准价格:</label>
          <span>{{ formatPrice(selectedStrategy.price) }}</span>
        </div>
        <div class="detail-item">
          <label>总数量:</label>
          <span>{{ selectedStrategy.totalQuantity }}</span>
        </div>
        <div v-if="selectedStrategy.buyQuantities && selectedStrategy.buyQuantities.length > 0" class="detail-item">
          <label>买入配置:</label>
          <div>
            <p>数量分配: {{ selectedStrategy.buyQuantities.join(', ') }}</p>
            <p v-if="selectedStrategy.strategyType !== 'custom'">深度级别: {{ selectedStrategy.buyDepthLevels ? selectedStrategy.buyDepthLevels.join(', ') : '' }}</p>
            <p v-if="selectedStrategy.strategyType === 'custom' && selectedStrategy.buyBasisPoints">
              万分比: {{ selectedStrategy.buyBasisPoints.map(bp => bp > 0 ? '+' + bp : bp).join(', ') }}bp
            </p>
          </div>
        </div>
        <div v-if="selectedStrategy.sellQuantities && selectedStrategy.sellQuantities.length > 0" class="detail-item">
          <label>卖出配置:</label>
          <div>
            <p>数量分配: {{ selectedStrategy.sellQuantities.join(', ') }}</p>
            <p v-if="selectedStrategy.strategyType !== 'custom'">深度级别: {{ selectedStrategy.sellDepthLevels ? selectedStrategy.sellDepthLevels.join(', ') : '' }}</p>
            <p v-if="selectedStrategy.strategyType === 'custom' && selectedStrategy.sellBasisPoints">
              万分比: {{ selectedStrategy.sellBasisPoints.map(bp => bp > 0 ? '+' + bp : bp).join(', ') }}bp
            </p>
          </div>
        </div>
        <button @click="closeDetails" class="close-btn">关闭</button>
      </div>
    </div>

    <!-- 策略统计弹窗 -->
    <div v-if="showStats" class="modal-overlay" @click="closeStats">
      <div class="modal-content large" @click.stop>
        <h3>策略统计 - {{ statsData.strategy?.symbol }}</h3>

        <div class="stats-grid">
          <div class="stat-card">
            <h4>总订单数</h4>
            <p class="stat-value">{{ statsData.stats?.totalOrders || 0 }}</p>
          </div>
          <div class="stat-card">
            <h4>待处理订单</h4>
            <p class="stat-value pending">{{ statsData.stats?.pendingOrders || 0 }}</p>
          </div>
          <div class="stat-card">
            <h4>已成交订单</h4>
            <p class="stat-value success">{{ statsData.stats?.filledOrders || 0 }}</p>
          </div>
          <div class="stat-card">
            <h4>已取消订单</h4>
            <p class="stat-value cancelled">{{ statsData.stats?.cancelledOrders || 0 }}</p>
          </div>
          <div class="stat-card">
            <h4>总交易额</h4>
            <p class="stat-value">{{ formatVolume(statsData.stats?.totalVolume || 0) }}</p>
          </div>
          <div class="stat-card">
            <h4>已成交额</h4>
            <p class="stat-value success">{{ formatVolume(statsData.stats?.filledVolume || 0) }}</p>
          </div>
        </div>

        <div class="recent-orders">
          <h4>最近订单</h4>
          <table v-if="statsData.recentOrders && statsData.recentOrders.length > 0">
            <thead>
            <tr>
              <th>订单ID</th>
              <th>方向</th>
              <th>价格</th>
              <th>数量</th>
              <th>状态</th>
              <th>创建时间</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="order in statsData.recentOrders" :key="order.id">
              <td>{{ order.orderId }}</td>
              <td>
                  <span :class="order.side === 'BUY' ? 'buy-side' : 'sell-side'">
                    {{ order.side === 'BUY' ? '买入' : '卖出' }}
                  </span>
              </td>
              <td>{{ formatPrice(order.price) }}</td>
              <td>{{ formatQuantity(order.quantity) }}</td>
              <td>
                  <span :class="`status-${order.status}`">
                    {{ getOrderStatusText(order.status) }}
                  </span>
              </td>
              <td>{{ formatDate(order.createdAt) }}</td>
            </tr>
            </tbody>
          </table>
          <p v-else class="no-orders">暂无订单记录</p>
        </div>

        <div class="modal-actions">
          <button @click="viewAllStrategyOrders" class="view-all-btn">查看所有订单</button>
          <button @click="closeStats" class="close-btn">关闭</button>
        </div>
      </div>
    </div>

    <!-- 创建策略表单 -->
    <section class="create-strategy">
      <h2>创建新策略</h2>
      <form @submit.prevent="createStrategy">
        <div class="form-row">
          <div class="form-group">
            <label>交易对</label>
            <select v-model="newStrategy.symbol" required>
              <option value="">选择交易对</option>
              <option v-for="symbol in availableSymbols" :key="symbol" :value="symbol">
                {{ symbol }}
              </option>
            </select>
          </div>
          <div class="form-group">
            <label>策略类型</label>
            <select v-model="newStrategy.strategyType" @change="onStrategyTypeChange" required>
              <option value="">选择策略类型</option>
              <option value="simple">简单策略</option>
              <option value="iceberg">冰山策略</option>
              <option value="custom">自定义策略</option>
            </select>
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>交易方向</label>
            <select v-model="newStrategy.side" @change="onSideChange" required>
              <option value="">选择方向</option>
              <option value="BUY">买入</option>
              <option value="SELL">卖出</option>
            </select>
          </div>
          <div class="form-group">
            <label>基准价格</label>
            <input v-model.number="newStrategy.price"
                   type="number"
                   step="0.00000001"
                   placeholder="基准价格"
                   required />
          </div>
          <div class="form-group">
            <label>总数量</label>
            <input v-model.number="newStrategy.totalQuantity"
                   type="number"
                   step="0.00000001"
                   placeholder="交易总数量"
                   required />
          </div>
        </div>

        <!-- 策略说明 -->
        <div class="strategy-description">
          <p v-if="newStrategy.strategyType === 'simple'">
            <strong>简单策略：</strong>当价格达到触发条件时，以基准价格一次性下单全部数量。
          </p>
          <p v-if="newStrategy.strategyType === 'iceberg'">
            <strong>冰山策略：</strong>将订单分成多个小订单，基于基准价格按固定万分比在不同价格层级分批下单。
            <br>默认万分比：买单[0, -1, -3, -5, -7]，卖单[0, 1, 3, 5, 7]
          </p>
          <p v-if="newStrategy.strategyType === 'custom'">
            <strong>自定义策略：</strong>基于基准价格，按自定义万分比计算各档位价格进行分批下单。
            <br><strong>万分比说明：</strong>正数表示高于基准价格，负数表示低于基准价格。例如：+50表示基准价格+0.5%，-30表示基准价格-0.3%
          </p>
        </div>

        <!-- 自定义策略的额外配置 -->
        <div v-if="newStrategy.strategyType === 'custom'" class="custom-config">
          <h3>自定义配置</h3>
          <p class="config-hint">配置每个订单的数量比例和相对于基准价格的万分比偏移</p>

          <div v-if="newStrategy.side === 'BUY'" class="config-section">
            <h4>买入配置</h4>
            <div class="form-row">
              <div class="form-group">
                <label>数量比例</label>
                <input v-model="buyQuantitiesInput"
                       placeholder="如: 0.3,0.3,0.2,0.2"
                       @blur="validateQuantities('buy')" />
                <small>每个值表示占总数量的比例，总和应为1</small>
              </div>
              <div class="form-group">
                <label>万分比偏移</label>
                <input v-model="buyBasisPointsInput"
                       placeholder="如: 0,-10,-20,-30" />
                <small>相对于基准价格的万分比偏移（负数表示更低价格）</small>
              </div>
            </div>
          </div>

          <div v-if="newStrategy.side === 'SELL'" class="config-section">
            <h4>卖出配置</h4>
            <div class="form-row">
              <div class="form-group">
                <label>数量比例</label>
                <input v-model="sellQuantitiesInput"
                       placeholder="如: 0.3,0.3,0.2,0.2"
                       @blur="validateQuantities('sell')" />
                <small>每个值表示占总数量的比例，总和应为1</small>
              </div>
              <div class="form-group">
                <label>万分比偏移</label>
                <input v-model="sellBasisPointsInput"
                       placeholder="如: 0,10,20,30" />
                <small>相对于基准价格的万分比偏移（正数表示更高价格）</small>
              </div>
            </div>
          </div>

          <div v-if="quantityWarning" class="warning-message">
            {{ quantityWarning }}
          </div>
        </div>

        <!-- 预览订单分配 -->
        <div v-if="orderPreview.length > 0" class="order-preview">
          <h3>订单预览</h3>
          <table>
            <thead>
            <tr>
              <th>订单</th>
              <th>数量</th>
              <th>占比</th>
              <th v-if="newStrategy.strategyType === 'custom'">万分比</th>
              <th>预估价格</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="(order, index) in orderPreview" :key="index">
              <td>订单 {{ index + 1 }}</td>
              <td>{{ order.quantity.toFixed(8) }}</td>
              <td>{{ (order.ratio * 100).toFixed(1) }}%</td>
              <td v-if="newStrategy.strategyType === 'custom'">
                {{ order.basisPoint > 0 ? '+' : '' }}{{ order.basisPoint }}bp
              </td>
              <td>{{ order.price.toFixed(8) }}</td>
            </tr>
            </tbody>
          </table>
        </div>

        <button type="submit" :disabled="isCreatingStrategy || !isFormValid">
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
      buyBasisPointsInput: '',
      sellQuantitiesInput: '',
      sellBasisPointsInput: '',
      currentPage: 1,
      pageSize: 10,
      errorMessage: '',
      successMessage: '',
      isCreatingStrategy: false,
      showDetails: false,
      selectedStrategy: {},
      quantityWarning: '',
      orderPreview: [],
      availableSymbols: [],
      showStats: false,
      statsData: {
        stats: {},
        recentOrders: [],
        strategy: {}
      }
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
    isFormValid() {
      if (!this.newStrategy.symbol || !this.newStrategy.strategyType ||
          !this.newStrategy.side || this.newStrategy.price <= 0 ||
          this.newStrategy.totalQuantity <= 0) {
        return false;
      }

      if (this.newStrategy.strategyType === 'custom') {
        if (this.newStrategy.side === 'BUY') {
          return this.buyQuantitiesInput && this.buyBasisPointsInput && !this.quantityWarning;
        } else {
          return this.sellQuantitiesInput && this.sellBasisPointsInput && !this.quantityWarning;
        }
      }

      return true;
    }
  },
  watch: {
    'newStrategy.strategyType': function(newVal) {
      this.updateOrderPreview();
    },
    'newStrategy.side': function(newVal) {
      this.updateOrderPreview();
    },
    'newStrategy.totalQuantity': function(newVal) {
      this.updateOrderPreview();
    },
    'newStrategy.price': function(newVal) {
      this.updateOrderPreview();
    },
    buyQuantitiesInput: function() {
      this.updateOrderPreview();
    },
    sellQuantitiesInput: function() {
      this.updateOrderPreview();
    },
    buyBasisPointsInput: function() {
      this.updateOrderPreview();
    },
    sellBasisPointsInput: function() {
      this.updateOrderPreview();
    }
  },
  mounted() {
    this.fetchStrategies();
    this.fetchSymbols();
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

    formatVolume(volume) {
      if (volume >= 1000000) {
        return (volume / 1000000).toFixed(2) + 'M';
      } else if (volume >= 1000) {
        return (volume / 1000).toFixed(2) + 'K';
      }
      return volume.toFixed(2);
    },

    formatDate(dateString) {
      return new Date(dateString).toLocaleString('zh-CN');
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

    getOrderStatusText(status) {
      const statusMap = {
        'pending': '待处理',
        'filled': '已成交',
        'cancelled': '已取消',
        'expired': '已过期',
        'rejected': '已拒绝'
      };
      return statusMap[status] || status;
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

    async fetchSymbols() {
      try {
        const response = await axios.get('/symbols', {
          headers: this.getAuthHeaders(),
        });
        this.availableSymbols = response.data.symbols || [];

        if (this.availableSymbols.length === 0) {
          this.showMessage('请先在仪表盘中添加交易对', true);
        }
      } catch (error) {
        console.error('获取交易对失败:', error);
        this.showMessage(error.response?.data?.error || '获取交易对失败', true);
      }
    },

    onStrategyTypeChange() {
      if (this.newStrategy.strategyType !== 'custom') {
        this.buyQuantitiesInput = '';
        this.buyBasisPointsInput = '';
        this.sellQuantitiesInput = '';
        this.sellBasisPointsInput = '';
      } else {
        if (this.newStrategy.side === 'BUY' && !this.buyQuantitiesInput) {
          this.buyQuantitiesInput = '0.3,0.3,0.2,0.2';
          this.buyBasisPointsInput = '0,-10,-20,-30';
        } else if (this.newStrategy.side === 'SELL' && !this.sellQuantitiesInput) {
          this.sellQuantitiesInput = '0.3,0.3,0.2,0.2';
          this.sellBasisPointsInput = '0,10,20,30';
        }
      }
      this.updateOrderPreview();
    },

    onSideChange() {
      if (this.newStrategy.strategyType === 'custom') {
        if (this.newStrategy.side === 'BUY') {
          this.buyQuantitiesInput = '0.3,0.3,0.2,0.2';
          this.buyBasisPointsInput = '0,-10,-20,-30';
          this.sellQuantitiesInput = '';
          this.sellBasisPointsInput = '';
        } else {
          this.sellQuantitiesInput = '0.3,0.3,0.2,0.2';
          this.sellBasisPointsInput = '0,10,20,30';
          this.buyQuantitiesInput = '';
          this.buyBasisPointsInput = '';
        }
      }
      this.updateOrderPreview();
    },

    validateQuantities(side) {
      let input = side === 'buy' ? this.buyQuantitiesInput : this.sellQuantitiesInput;
      if (!input) {
        this.quantityWarning = '';
        return;
      }

      const quantities = input.split(',').map(v => parseFloat(v.trim())).filter(v => !isNaN(v));
      const sum = quantities.reduce((acc, val) => acc + val, 0);

      if (Math.abs(sum - 1.0) > 0.01) {
        this.quantityWarning = `数量比例总和为 ${sum.toFixed(2)}，应该为 1.0`;
      } else {
        this.quantityWarning = '';
      }
    },

    updateOrderPreview() {
      this.orderPreview = [];

      if (!this.newStrategy.totalQuantity || this.newStrategy.totalQuantity <= 0 || !this.newStrategy.price) {
        return;
      }

      let quantities = [];
      let basisPoints = [];

      if (this.newStrategy.strategyType === 'simple') {
        quantities = [1.0];
        basisPoints = [0];
      } else if (this.newStrategy.strategyType === 'iceberg') {
        quantities = [0.35, 0.25, 0.2, 0.1, 0.1];
        if (this.newStrategy.side === 'SELL') {
          basisPoints = [0, 1, 3, 5, 7];
        } else {
          basisPoints = [0, -1, -3, -5, -7];
        }
      } else if (this.newStrategy.strategyType === 'custom') {
        if (this.newStrategy.side === 'BUY' && this.buyQuantitiesInput && this.buyBasisPointsInput) {
          quantities = this.buyQuantitiesInput.split(',').map(v => parseFloat(v.trim())).filter(v => !isNaN(v));
          basisPoints = this.buyBasisPointsInput.split(',').map(v => parseFloat(v.trim())).filter(v => !isNaN(v));
        } else if (this.newStrategy.side === 'SELL' && this.sellQuantitiesInput && this.sellBasisPointsInput) {
          quantities = this.sellQuantitiesInput.split(',').map(v => parseFloat(v.trim())).filter(v => !isNaN(v));
          basisPoints = this.sellBasisPointsInput.split(',').map(v => parseFloat(v.trim())).filter(v => !isNaN(v));
        }
      }

      for (let i = 0; i < quantities.length && i < basisPoints.length; i++) {
        const multiplier = 1 + (basisPoints[i] / 10000);
        const price = this.newStrategy.price * multiplier;

        this.orderPreview.push({
          quantity: this.newStrategy.totalQuantity * quantities[i],
          ratio: quantities[i],
          basisPoint: basisPoints[i],
          price: price
        });
      }
    },

    async createStrategy() {
      if (!this.isFormValid) {
        this.showMessage('请填写所有必需字段', true);
        return;
      }

      if (!this.availableSymbols.includes(this.newStrategy.symbol)) {
        this.showMessage('请选择有效的交易对', true);
        return;
      }

      this.isCreatingStrategy = true;
      try {
        const strategyData = { ...this.newStrategy };

        if (this.newStrategy.strategyType === 'custom') {
          if (this.newStrategy.side === 'BUY') {
            strategyData.buyQuantities = this.buyQuantitiesInput.split(',').map(v => parseFloat(v.trim()));
            strategyData.buyBasisPoints = this.buyBasisPointsInput.split(',').map(v => parseFloat(v.trim()));
          } else {
            strategyData.sellQuantities = this.sellQuantitiesInput.split(',').map(v => parseFloat(v.trim()));
            strategyData.sellBasisPoints = this.sellBasisPointsInput.split(',').map(v => parseFloat(v.trim()));
          }
        }

        const response = await axios.post('/strategy', strategyData, {
          headers: this.getAuthHeaders(),
        });

        this.showMessage(response.data.message || '策略创建成功');
        this.resetForm();
        this.fetchStrategies();
      } catch (error) {
        console.error('创建策略失败:', error);
        this.showMessage(error.response?.data?.error || '创建策略失败', true);
      } finally {
        this.isCreatingStrategy = false;
      }
    },

    async toggleStrategy(strategy) {
      try {
        const response = await axios.post('/toggle_strategy', { id: strategy.id }, {
          headers: this.getAuthHeaders(),
        });

        this.showMessage(response.data.message || '策略状态切换成功');
        this.fetchStrategies();
      } catch (error) {
        console.error('切换策略失败:', error);
        this.showMessage(error.response?.data?.error || '切换策略失败', true);
      }
    },

    viewStrategyDetails(strategy) {
      this.selectedStrategy = strategy;
      this.showDetails = true;
    },

    closeDetails() {
      this.showDetails = false;
      this.selectedStrategy = {};
    },

    async viewStrategyStats(strategy) {
      try {
        const response = await axios.get(`/strategy/${strategy.id}/stats`, {
          headers: this.getAuthHeaders(),
        });

        this.statsData = response.data;
        this.showStats = true;
      } catch (error) {
        console.error('获取策略统计失败:', error);
        this.showMessage(error.response?.data?.error || '获取策略统计失败', true);
      }
    },

    closeStats() {
      this.showStats = false;
      this.statsData = {
        stats: {},
        recentOrders: [],
        strategy: {}
      };
    },

    async viewAllStrategyOrders() {
      this.$router.push({
        path: '/orders',
        query: { strategyId: this.statsData.strategy.id }
      });
    },

    async deleteStrategy(strategyId) {
      if (!window.confirm('确定要删除这个策略吗？删除后无法恢复。')) {
        return;
      }

      try {
        const response = await axios.post('/delete_strategy', { id: strategyId }, {
          headers: this.getAuthHeaders(),
        });

        this.showMessage(response.data.message || '策略删除成功');
        this.fetchStrategies();
      } catch (error) {
        console.error('删除策略失败:', error);
        this.showMessage(error.response?.data?.error || '删除策略失败', true);
      }
    },

    resetForm() {
      this.newStrategy = {
        symbol: '',
        strategyType: '',
        side: '',
        price: 0,
        totalQuantity: 0
      };
      this.buyQuantitiesInput = '';
      this.buyBasisPointsInput = '';
      this.sellQuantitiesInput = '';
      this.sellBasisPointsInput = '';
      this.quantityWarning = '';
      this.orderPreview = [];
    },
  },
};
</script>

<style scoped>
.strategy {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
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

.create-strategy {
  margin-top: 40px;
  padding: 30px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background-color: #f9f9f9;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
}

.custom-config {
  margin-top: 20px;
  padding: 20px;
  border: 1px solid #ccc;
  border-radius: 4px;
  background-color: #fff;
}

.config-section {
  margin-bottom: 20px;
}

.config-hint {
  color: #666;
  font-size: 14px;
  margin-bottom: 15px;
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

.form-group small {
  color: #666;
  font-size: 12px;
  margin-top: 5px;
}

input, select, button {
  padding: 10px;
  font-size: 14px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

input:focus, select:focus {
  outline: none;
  border-color: #007bff;
  box-shadow: 0 0 0 2px rgba(0,123,255,0.25);
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

.enable-btn {
  background-color: #28a745;
  padding: 6px 12px;
  font-size: 12px;
  margin-right: 5px;
}

.enable-btn:hover {
  background-color: #218838;
}

.disable-btn {
  background-color: #ffc107;
  padding: 6px 12px;
  font-size: 12px;
  margin-right: 5px;
}

.disable-btn:hover {
  background-color: #e0a800;
}

.view-btn {
  background-color: #17a2b8;
  padding: 6px 12px;
  font-size: 12px;
  margin-right: 5px;
}

.view-btn:hover {
  background-color: #138496;
}

.stats-btn {
  background-color: #6c757d;
  padding: 6px 12px;
  font-size: 12px;
  margin-right: 5px;
}

.stats-btn:hover {
  background-color: #5a6268;
}

.delete-btn {
  background-color: #dc3545;
  padding: 6px 12px;
  font-size: 12px;
}

.delete-btn:hover {
  background-color: #c82333;
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

.warning-message {
  background-color: #fff3cd;
  color: #856404;
  padding: 10px;
  border: 1px solid #ffeaa7;
  border-radius: 4px;
  margin-top: 10px;
}

.strategy-description {
  background-color: #e9ecef;
  padding: 15px;
  border-radius: 4px;
  margin: 15px 0;
}

.strategy-description p {
  margin: 0;
  color: #495057;
}

.strategy-description strong {
  color: #007bff;
}

.order-preview {
  margin-top: 20px;
  padding: 15px;
  background-color: #fff;
  border: 1px solid #dee2e6;
  border-radius: 4px;
}

.order-preview h3 {
  margin-top: 0;
  color: #333;
}

.order-preview table {
  box-shadow: none;
}

.buy-side {
  color: #28a745;
  font-weight: bold;
}

.sell-side {
  color: #dc3545;
  font-weight: bold;
}

.enabled {
  color: #28a745;
  font-weight: bold;
}

.disabled {
  color: #6c757d;
}

.executing {
  color: #ffc107;
  font-weight: bold;
}

.idle {
  color: #6c757d;
}

.status-active {
  color: #28a745;
  font-weight: bold;
}

.status-inactive {
  color: #6c757d;
}

.status-completed {
  color: #007bff;
}

.status-cancelled {
  color: #dc3545;
}

.status-pending {
  color: #ffc107;
  font-weight: bold;
}

.status-filled {
  color: #28a745;
  font-weight: bold;
}

/* 弹窗样式 */
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
  background-color: white;
  padding: 30px;
  border-radius: 8px;
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
}

.modal-content.large {
  max-width: 900px;
}

.modal-content h3 {
  margin-top: 0;
  color: #333;
  border-bottom: 2px solid #007bff;
  padding-bottom: 10px;
}

.detail-item {
  margin: 15px 0;
  display: flex;
  align-items: flex-start;
}

.detail-item label {
  font-weight: bold;
  min-width: 120px;
  color: #666;
}

.detail-item span, .detail-item div {
  color: #333;
}

.close-btn {
  background-color: #6c757d;
  margin-top: 20px;
}

.close-btn:hover {
  background-color: #5a6268;
}

/* 统计相关样式 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 20px;
  margin: 20px 0;
}

.stat-card {
  background-color: #f8f9fa;
  padding: 20px;
  border-radius: 8px;
  text-align: center;
  border: 1px solid #dee2e6;
}

.stat-card h4 {
  margin: 0 0 10px 0;
  color: #666;
  font-size: 14px;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #333;
}

.stat-value.pending {
  color: #ffc107;
}

.stat-value.success {
  color: #28a745;
}

.stat-value.cancelled {
  color: #dc3545;
}

.recent-orders {
  margin-top: 30px;
}

.recent-orders h4 {
  margin-bottom: 15px;
  color: #333;
}

.recent-orders table {
  width: 100%;
  box-shadow: none;
  font-size: 14px;
}

.no-orders {
  text-align: center;
  color: #666;
  padding: 20px;
  font-style: italic;
}

.modal-actions {
  display: flex;
  justify-content: space-between;
  margin-top: 20px;
}

.view-all-btn {
  background-color: #007bff;
}

.view-all-btn:hover {
  background-color: #0056b3;
}
</style>
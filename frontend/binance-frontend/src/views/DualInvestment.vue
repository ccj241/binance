<template>
  <div class="dual-investment-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1 class="page-title">双币投资</h1>
      <p class="page-description">高收益结构化理财产品</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-label">总投资金额</span>
          <span class="stat-icon">💰</span>
        </div>
        <div class="stat-value">{{ formatCurrency(stats.totalInvested) }}</div>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-label">总盈亏</span>
          <span class="stat-icon">📈</span>
        </div>
        <div class="stat-value" :class="stats.totalPnL >= 0 ? 'positive' : 'negative'">
          {{ stats.totalPnL >= 0 ? '+' : '' }}{{ formatCurrency(Math.abs(stats.totalPnL)) }}
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-label">胜率</span>
          <span class="stat-icon">🎯</span>
        </div>
        <div class="stat-value">{{ stats.winRate?.toFixed(1) || 0 }}%</div>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-label">活跃订单</span>
          <span class="stat-icon">⚡</span>
        </div>
        <div class="stat-value">{{ stats.activeOrders || 0 }}</div>
      </div>
    </div>

    <!-- Tab 导航 -->
    <div class="tab-nav">
      <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          :class="['tab-btn', { active: activeTab === tab.key }]"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- 产品市场 -->
    <div v-show="activeTab === 'market'" class="tab-content">
      <div class="content-card">
        <div class="card-header">
          <h2 class="card-title">产品市场</h2>
          <div class="filters">
            <select v-model="filters.symbol" class="filter-select">
              <option value="">所有交易对</option>
              <option value="BTCUSDT">BTC/USDT</option>
              <option value="ETHUSDT">ETH/USDT</option>
              <option value="BNBUSDT">BNB/USDT</option>
              <option value="SOLUSDT">SOL/USDT</option>
            </select>
            <select v-model="filters.direction" class="filter-select">
              <option value="">所有方向</option>
              <option value="UP">低买(看涨)</option>
              <option value="DOWN">高卖(看跌)</option>
            </select>
            <button @click="fetchProducts" class="btn btn-primary">
              搜索
            </button>
          </div>
        </div>

        <div class="card-body">
          <div v-if="loadingProducts" class="loading-state">
            <div class="spinner"></div>
            <p>加载产品中...</p>
          </div>

          <div v-else-if="products.length === 0" class="empty-state">
            <span class="empty-icon">📦</span>
            <p>暂无可投资产品</p>
          </div>

          <div v-else class="products-grid">
            <div v-for="product in products" :key="product.id" class="product-card">
              <div class="product-header">
                <h3 class="product-symbol">{{ product.symbol }}</h3>
                <span :class="['direction-badge', product.direction.toLowerCase()]">
                  {{ product.direction === 'UP' ? '低买' : '高卖' }}
                </span>
              </div>

              <div class="product-info">
                <div class="info-row highlight">
                  <span class="label">年化收益率</span>
                  <span class="value apy">{{ product.apy.toFixed(2) }}%</span>
                </div>
                <div class="info-row">
                  <span class="label">执行价格</span>
                  <span class="value">{{ formatPrice(product.strikePrice) }}</span>
                </div>
                <div class="info-row">
                  <span class="label">当前价格</span>
                  <span class="value">{{ formatPrice(product.currentPrice) }}</span>
                </div>
                <div class="info-row">
                  <span class="label">投资期限</span>
                  <span class="value">{{ product.duration }}天</span>
                </div>
                <div class="info-row">
                  <span class="label">投资范围</span>
                  <span class="value">{{ product.minAmount }} - {{ product.maxAmount }}</span>
                </div>
              </div>

              <div class="product-actions">
                <button @click="showInvestModal(product)" class="btn btn-primary btn-block">
                  立即投资
                </button>
                <button @click="showSimulateModal(product)" class="btn btn-outline btn-block">
                  收益计算
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 我的策略 -->
    <div v-show="activeTab === 'strategies'" class="tab-content">
      <div class="content-card">
        <div class="card-header">
          <h2 class="card-title">我的策略</h2>
          <button @click="showStrategyModal()" class="btn btn-primary">
            创建策略
          </button>
        </div>

        <div class="card-body">
          <div v-if="strategies.length === 0" class="empty-state">
            <span class="empty-icon">🎯</span>
            <p>暂无投资策略</p>
            <button @click="showStrategyModal()" class="btn btn-primary">
              创建第一个策略
            </button>
          </div>

          <div v-else class="strategies-list">
            <div v-for="strategy in strategies" :key="strategy.id" class="strategy-item">
              <div class="strategy-header">
                <h3 class="strategy-name">{{ strategy.strategyName }}</h3>
                <label class="toggle-switch">
                  <input
                      type="checkbox"
                      :checked="strategy.enabled"
                      @change="toggleStrategy(strategy)"
                  />
                  <span class="toggle-slider"></span>
                </label>
              </div>

              <div class="strategy-info">
                <div class="info-item">
                  <span class="label">策略类型</span>
                  <span class="value">{{ getStrategyTypeText(strategy.strategyType) }}</span>
                </div>
                <div class="info-item">
                  <span class="label">交易对</span>
                  <span class="value">{{ strategy.baseAsset }}/{{ strategy.quoteAsset }}</span>
                </div>
                <div class="info-item">
                  <span class="label">方向偏好</span>
                  <span class="value">{{ getDirectionText(strategy.directionPreference) }}</span>
                </div>
                <div class="info-item">
                  <span class="label">目标年化</span>
                  <span class="value">{{ strategy.targetApyMin }}% - {{ strategy.targetApyMax }}%</span>
                </div>
                <div class="info-item">
                  <span class="label">已投资/限额</span>
                  <span class="value">
                    {{ formatCurrency(strategy.currentInvested) }} /
                    {{ formatCurrency(strategy.totalInvestmentLimit) }}
                  </span>
                </div>
              </div>

              <div class="strategy-actions">
                <button @click="editStrategy(strategy)" class="btn btn-sm btn-outline">
                  编辑
                </button>
                <button @click="deleteStrategy(strategy)" class="btn btn-sm btn-danger">
                  删除
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 我的订单 -->
    <div v-show="activeTab === 'orders'" class="tab-content">
      <div class="content-card">
        <div class="card-header">
          <h2 class="card-title">我的订单</h2>
          <select v-model="orderFilter" class="filter-select">
            <option value="">所有订单</option>
            <option value="active">活跃订单</option>
            <option value="settled">已结算</option>
          </select>
        </div>

        <div class="card-body">
          <div v-if="orders.length === 0" class="empty-state">
            <span class="empty-icon">📋</span>
            <p>暂无订单记录</p>
          </div>

          <div v-else class="table-wrapper">
            <table class="data-table">
              <thead>
              <tr>
                <th>订单ID</th>
                <th>交易对</th>
                <th>方向</th>
                <th>投资金额</th>
                <th>执行价格</th>
                <th>年化收益</th>
                <th>状态</th>
                <th>结算时间</th>
                <th>盈亏</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="order in filteredOrders" :key="order.id">
                <td>{{ order.orderId }}</td>
                <td>{{ order.symbol }}</td>
                <td>
                    <span :class="['direction-badge', order.direction.toLowerCase()]">
                      {{ order.direction === 'UP' ? '低买' : '高卖' }}
                    </span>
                </td>
                <td>{{ formatCurrency(order.investAmount) }} {{ order.investAsset }}</td>
                <td>{{ formatPrice(order.strikePrice) }}</td>
                <td>{{ order.apy.toFixed(2) }}%</td>
                <td>
                    <span :class="['status-badge', order.status]">
                      {{ getStatusText(order.status) }}
                    </span>
                </td>
                <td>{{ formatDate(order.settlementTime) }}</td>
                <td>
                    <span v-if="order.status === 'settled'"
                          :class="order.pnl >= 0 ? 'positive' : 'negative'">
                      {{ order.pnl >= 0 ? '+' : '' }}{{ formatCurrency(Math.abs(order.pnl)) }}
                    </span>
                  <span v-else>-</span>
                </td>
              </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>

    <!-- 投资弹窗 -->
    <transition name="modal">
      <div v-if="showInvestDialog" class="modal-overlay" @click="closeInvestModal">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">投资产品</h3>
            <button @click="closeInvestModal" class="modal-close">×</button>
          </div>

          <div class="modal-body">
            <div class="invest-info">
              <h4>{{ selectedProduct.symbol }} - {{ selectedProduct.direction === 'UP' ? '低买(看涨)' : '高卖(看跌)' }}</h4>
              <div class="info-grid">
                <div class="info-item">
                  <span class="label">年化收益率</span>
                  <span class="value highlight">{{ selectedProduct.apy?.toFixed(2) }}%</span>
                </div>
                <div class="info-item">
                  <span class="label">执行价格</span>
                  <span class="value">{{ formatPrice(selectedProduct.strikePrice) }}</span>
                </div>
                <div class="info-item">
                  <span class="label">投资期限</span>
                  <span class="value">{{ selectedProduct.duration }}天</span>
                </div>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">投资金额</label>
              <input
                  v-model.number="investAmount"
                  type="number"
                  class="form-control"
                  :min="selectedProduct.minAmount"
                  :max="selectedProduct.maxAmount"
                  :placeholder="`${selectedProduct.minAmount} - ${selectedProduct.maxAmount}`"
              />
            </div>

            <div class="form-group">
              <label class="form-label">关联策略（可选）</label>
              <select v-model="investStrategyId" class="form-control">
                <option :value="null">不关联策略</option>
                <option v-for="s in strategies" :key="s.id" :value="s.id">
                  {{ s.strategyName }}
                </option>
              </select>
            </div>

            <div class="risk-warning">
              <span class="warning-icon">⚠️</span>
              <p>风险提示：双币投资产品不保本，到期可能以其他币种结算</p>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeInvestModal" class="btn btn-outline">取消</button>
            <button @click="confirmInvest" class="btn btn-primary" :disabled="!isInvestValid">
              确认投资
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- 策略弹窗 -->
    <transition name="modal">
      <div v-if="showStrategyDialog" class="modal-overlay" @click="closeStrategyModal">
        <div class="modal-content modal-lg" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">{{ editingStrategy ? '编辑策略' : '创建策略' }}</h3>
            <button @click="closeStrategyModal" class="modal-close">×</button>
          </div>

          <form @submit.prevent="saveStrategy" class="modal-body">
            <div class="form-grid">
              <div class="form-group">
                <label class="form-label">策略名称</label>
                <input v-model="strategyForm.strategyName" class="form-control" required />
              </div>

              <div class="form-group">
                <label class="form-label">策略类型</label>
                <select v-model="strategyForm.strategyType" class="form-control" required>
                  <option value="single">单次投资</option>
                  <option value="auto_reinvest">自动复投</option>
                  <option value="ladder">梯度投资</option>
                  <option value="price_trigger">价格触发</option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">基础资产</label>
                <select v-model="strategyForm.baseAsset" class="form-control" required>
                  <option value="BTC">BTC</option>
                  <option value="ETH">ETH</option>
                  <option value="BNB">BNB</option>
                  <option value="SOL">SOL</option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">计价资产</label>
                <select v-model="strategyForm.quoteAsset" class="form-control" required>
                  <option value="USDT">USDT</option>
                  <option value="BUSD">BUSD</option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">方向偏好</label>
                <select v-model="strategyForm.directionPreference" class="form-control" required>
                  <option value="UP">只做低买(看涨)</option>
                  <option value="DOWN">只做高卖(看跌)</option>
                  <option value="BOTH">双向都做</option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">目标年化范围 (%)</label>
                <div class="input-group">
                  <input
                      v-model.number="strategyForm.targetApyMin"
                      type="number"
                      class="form-control"
                      min="0"
                      placeholder="最小"
                      required
                  />
                  <span class="input-separator">-</span>
                  <input
                      v-model.number="strategyForm.targetApyMax"
                      type="number"
                      class="form-control"
                      min="0"
                      placeholder="最大"
                      required
                  />
                </div>
              </div>

              <div class="form-group">
                <label class="form-label">单笔最大金额</label>
                <input v-model.number="strategyForm.maxSingleAmount" type="number" class="form-control" min="0" required />
              </div>

              <div class="form-group">
                <label class="form-label">总投资限额</label>
                <input v-model.number="strategyForm.totalInvestmentLimit" type="number" class="form-control" min="0" required />
              </div>
            </div>

            <div class="modal-footer">
              <button type="button" @click="closeStrategyModal" class="btn btn-outline">取消</button>
              <button type="submit" class="btn btn-primary">
                {{ editingStrategy ? '保存修改' : '创建策略' }}
              </button>
            </div>
          </form>
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
  name: 'DualInvestment',
  data() {
    return {
      activeTab: 'market',
      tabs: [
        { key: 'market', label: '产品市场' },
        { key: 'strategies', label: '我的策略' },
        { key: 'orders', label: '我的订单' }
      ],

      // 产品相关
      products: [],
      loadingProducts: false,
      filters: {
        symbol: '',
        direction: '',
        minApy: null
      },

      // 策略相关
      strategies: [],
      showStrategyDialog: false,
      editingStrategy: null,
      strategyForm: {
        strategyName: '',
        strategyType: 'single',
        baseAsset: 'BTC',
        quoteAsset: 'USDT',
        directionPreference: 'BOTH',
        targetApyMin: 5,
        targetApyMax: 50,
        maxSingleAmount: 1000,
        totalInvestmentLimit: 10000,
        maxStrikePriceOffset: 10,
        minDuration: 1,
        maxDuration: 30,
        autoReinvest: false
      },

      // 订单相关
      orders: [],
      orderFilter: '',

      // 统计信息
      stats: {
        totalInvested: 0,
        totalPnL: 0,
        winRate: 0,
        activeOrders: 0
      },

      // 投资弹窗
      showInvestDialog: false,
      selectedProduct: {},
      investAmount: 0,
      investStrategyId: null,

      // Toast
      toastMessage: '',
      toastType: 'success'
    };
  },

  computed: {
    filteredOrders() {
      if (!this.orderFilter) return this.orders;
      return this.orders.filter(order => order.status === this.orderFilter);
    },

    isInvestValid() {
      return this.investAmount >= (this.selectedProduct.minAmount || 0) &&
          this.investAmount <= (this.selectedProduct.maxAmount || Infinity);
    }
  },

  mounted() {
    this.fetchProducts();
    this.fetchStrategies();
    this.fetchOrders();
    this.fetchStats();
  },

  methods: {
    // 通用方法
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

    formatCurrency(amount) {
      return new Intl.NumberFormat('zh-CN', {
        style: 'currency',
        currency: 'USD',
        minimumFractionDigits: 2
      }).format(amount || 0);
    },

    formatPrice(price) {
      return parseFloat(price || 0).toFixed(2);
    },

    formatDate(dateString) {
      if (!dateString) return '-';
      return new Date(dateString).toLocaleDateString('zh-CN');
    },

    getStrategyTypeText(type) {
      const map = {
        'single': '单次投资',
        'auto_reinvest': '自动复投',
        'ladder': '梯度投资',
        'price_trigger': '价格触发'
      };
      return map[type] || type;
    },

    getDirectionText(direction) {
      const map = {
        'UP': '低买(看涨)',
        'DOWN': '高卖(看跌)',
        'BOTH': '双向'
      };
      return map[direction] || direction;
    },

    getStatusText(status) {
      const map = {
        'pending': '待处理',
        'active': '进行中',
        'settled': '已结算',
        'cancelled': '已取消'
      };
      return map[status] || status;
    },

    // 产品相关方法
    async fetchProducts() {
      this.loadingProducts = true;
      try {
        const params = new URLSearchParams();
        if (this.filters.symbol) params.append('symbol', this.filters.symbol);
        if (this.filters.direction) params.append('direction', this.filters.direction);
        if (this.filters.minApy) params.append('minApy', this.filters.minApy);

        const response = await axios.get(`/dual-investment/products?${params}`, {
          headers: this.getAuthHeaders()
        });

        this.products = response.data.products || [];
      } catch (error) {
        console.error('获取产品失败:', error);
        this.showToast('获取产品失败', 'error');
      } finally {
        this.loadingProducts = false;
      }
    },

    showInvestModal(product) {
      this.selectedProduct = product;
      this.investAmount = product.minAmount;
      this.investStrategyId = null;
      this.showInvestDialog = true;
    },

    closeInvestModal() {
      this.showInvestDialog = false;
      this.selectedProduct = {};
      this.investAmount = 0;
    },

    async confirmInvest() {
      if (!this.isInvestValid) {
        this.showToast('请输入有效的投资金额', 'error');
        return;
      }

      try {
        const response = await axios.post('/dual-investment/orders', {
          productId: this.selectedProduct.id,
          investAmount: this.investAmount,
          strategyId: this.investStrategyId
        }, {
          headers: this.getAuthHeaders()
        });

        this.showToast('投资成功！');
        this.closeInvestModal();
        this.fetchOrders();
        this.fetchStats();
      } catch (error) {
        console.error('投资失败:', error);
        this.showToast(error.response?.data?.error || '投资失败', 'error');
      }
    },

    // 策略相关方法
    async fetchStrategies() {
      try {
        const response = await axios.get('/dual-investment/strategies', {
          headers: this.getAuthHeaders()
        });
        this.strategies = response.data.strategies || [];
      } catch (error) {
        console.error('获取策略失败:', error);
        this.showToast('获取策略失败', 'error');
      }
    },

    showStrategyModal(strategy = null) {
      if (strategy) {
        this.editingStrategy = strategy;
        Object.assign(this.strategyForm, strategy);
      } else {
        this.editingStrategy = null;
        this.resetStrategyForm();
      }
      this.showStrategyDialog = true;
    },

    closeStrategyModal() {
      this.showStrategyDialog = false;
      this.editingStrategy = null;
      this.resetStrategyForm();
    },

    resetStrategyForm() {
      this.strategyForm = {
        strategyName: '',
        strategyType: 'single',
        baseAsset: 'BTC',
        quoteAsset: 'USDT',
        directionPreference: 'BOTH',
        targetApyMin: 5,
        targetApyMax: 50,
        maxSingleAmount: 1000,
        totalInvestmentLimit: 10000,
        maxStrikePriceOffset: 10,
        minDuration: 1,
        maxDuration: 30,
        autoReinvest: false
      };
    },

    async saveStrategy() {
      try {
        if (this.editingStrategy) {
          await axios.put(`/dual-investment/strategies/${this.editingStrategy.id}`,
              this.strategyForm, {
                headers: this.getAuthHeaders()
              });
          this.showToast('策略更新成功！');
        } else {
          await axios.post('/dual-investment/strategies', this.strategyForm, {
            headers: this.getAuthHeaders()
          });
          this.showToast('策略创建成功！');
        }

        this.closeStrategyModal();
        this.fetchStrategies();
      } catch (error) {
        console.error('保存策略失败:', error);
        this.showToast(error.response?.data?.error || '保存策略失败', 'error');
      }
    },

    editStrategy(strategy) {
      this.showStrategyModal(strategy);
    },

    async toggleStrategy(strategy) {
      try {
        await axios.put(`/dual-investment/strategies/${strategy.id}`, {
          enabled: !strategy.enabled
        }, {
          headers: this.getAuthHeaders()
        });

        strategy.enabled = !strategy.enabled;
        this.showToast(`策略已${strategy.enabled ? '启用' : '禁用'}`);
      } catch (error) {
        console.error('切换策略状态失败:', error);
        this.showToast('操作失败', 'error');
      }
    },

    async deleteStrategy(strategy) {
      if (!confirm(`确定要删除策略"${strategy.strategyName}"吗？`)) {
        return;
      }

      try {
        await axios.delete(`/dual-investment/strategies/${strategy.id}`, {
          headers: this.getAuthHeaders()
        });

        this.showToast('策略删除成功');
        this.fetchStrategies();
      } catch (error) {
        console.error('删除策略失败:', error);
        this.showToast(error.response?.data?.error || '删除策略失败', 'error');
      }
    },

    // 订单相关方法
    async fetchOrders() {
      try {
        const response = await axios.get('/dual-investment/orders', {
          headers: this.getAuthHeaders()
        });
        this.orders = response.data.orders || [];
      } catch (error) {
        console.error('获取订单失败:', error);
        this.showToast('获取订单失败', 'error');
      }
    },

    // 统计相关方法
    async fetchStats() {
      try {
        const response = await axios.get('/dual-investment/stats', {
          headers: this.getAuthHeaders()
        });
        this.stats = response.data.stats || {};
      } catch (error) {
        console.error('获取统计信息失败:', error);
      }
    },

    showSimulateModal(product) {
      // 简化处理，直接显示提示
      this.showToast('收益计算功能即将上线', 'info');
    }
  }
};
</script>

<style scoped>
/* 页面容器 */
.dual-investment-container {
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
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
}

.stat-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.stat-icon {
  font-size: 1.25rem;
  opacity: 0.7;
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.stat-value.positive {
  color: var(--color-success);
}

.stat-value.negative {
  color: var(--color-danger);
}

/* Tab 导航 */
.tab-nav {
  display: flex;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
  border-bottom: 1px solid var(--color-border);
  padding-bottom: -1px;
}

.tab-btn {
  padding: 0.75rem 1.5rem;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
}

.tab-btn:hover {
  color: var(--color-text-primary);
}

.tab-btn.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
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

/* 过滤器 */
.filters {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.filter-select {
  padding: 0.5rem 1rem;
  background-color: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-primary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all var(--transition-normal);
}

.filter-select:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

/* 产品网格 */
.products-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
}

.product-card {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  transition: all var(--transition-normal);
}

.product-card:hover {
  background-color: var(--color-bg-secondary);
}

.product-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.product-symbol {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.direction-badge {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.direction-badge.up {
  background-color: #d1fae5;
  color: #065f46;
}

.direction-badge.down {
  background-color: #fee2e2;
  color: #991b1b;
}

.product-info {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-row.highlight {
  padding: 0.5rem 0;
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
}

.info-row .label {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}

.info-row .value {
  color: var(--color-text-primary);
  font-weight: 500;
  font-size: 0.875rem;
}

.info-row .value.apy {
  color: var(--color-warning);
  font-size: 1.125rem;
  font-weight: 600;
}

.product-actions {
  display: flex;
  gap: 0.75rem;
}

/* 策略列表 */
.strategies-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.strategy-item {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  transition: all var(--transition-normal);
}

.strategy-item:hover {
  background-color: var(--color-bg-secondary);
}

.strategy-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.strategy-name {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

/* 开关样式 */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 44px;
  height: 24px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--color-border);
  transition: .4s;
  border-radius: 24px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 4px;
  bottom: 4px;
  background-color: white;
  transition: .4s;
  border-radius: 50%;
}

input:checked + .toggle-slider {
  background-color: var(--color-primary);
}

input:checked + .toggle-slider:before {
  transform: translateX(20px);
}

.strategy-info {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.info-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-item .label {
  color: var(--color-text-secondary);
  font-size: 0.75rem;
}

.info-item .value {
  color: var(--color-text-primary);
  font-size: 0.875rem;
  font-weight: 500;
}

.strategy-actions {
  display: flex;
  gap: 0.5rem;
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

.btn-block {
  width: 100%;
}

/* 表格样式 */
.table-wrapper {
  overflow-x: auto;
}

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
  white-space: nowrap;
}

.data-table td {
  padding: 0.75rem;
  border-top: 1px solid var(--color-border);
  font-size: 0.875rem;
}

.data-table tbody tr:hover {
  background-color: var(--color-bg-secondary);
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

.status-badge.pending {
  background-color: #fef3c7;
  color: #92400e;
}

.status-badge.active {
  background-color: #dbeafe;
  color: #1e40af;
}

.status-badge.settled {
  background-color: #d1fae5;
  color: #065f46;
}

.status-badge.cancelled {
  background-color: #f3f4f6;
  color: #6b7280;
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

/* 投资信息 */
.invest-info {
  background-color: var(--color-bg-secondary);
  border-radius: var(--radius-md);
  padding: 1rem;
  margin-bottom: 1.5rem;
}

.invest-info h4 {
  margin: 0 0 1rem 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
}

.info-grid .info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-grid .label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.info-grid .value {
  font-size: 0.875rem;
  color: var(--color-text-primary);
  font-weight: 500;
}

.info-grid .value.highlight {
  color: var(--color-warning);
  font-size: 1rem;
  font-weight: 600;
}

/* 表单样式 */
.form-group {
  margin-bottom: 1rem;
}

.form-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
  margin-bottom: 0.5rem;
}

.form-control {
  width: 100%;
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

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.input-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.input-separator {
  color: var(--color-text-tertiary);
  font-weight: 500;
}

/* 风险提示 */
.risk-warning {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.75rem;
  background-color: #fef3c7;
  border: 1px solid #fbbf24;
  border-radius: var(--radius-md);
  margin-top: 1rem;
}

.warning-icon {
  font-size: 1.25rem;
  flex-shrink: 0;
}

.risk-warning p {
  margin: 0;
  color: #92400e;
  font-size: 0.875rem;
  line-height: 1.5;
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

/* 加载状态 */
.loading-state {
  text-align: center;
  padding: 3rem 2rem;
  color: var(--color-text-tertiary);
}

.spinner {
  width: 40px;
  height: 40px;
  margin: 0 auto 1rem;
  border: 3px solid var(--color-border);
  border-top-color: var(--color-primary);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
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

  .tab-nav {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .filters {
    flex-wrap: wrap;
  }

  .products-grid {
    grid-template-columns: 1fr;
  }

  .strategy-info {
    grid-template-columns: 1fr;
  }

  .info-grid {
    grid-template-columns: 1fr;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .data-table {
    font-size: 0.75rem;
  }

  .data-table th,
  .data-table td {
    padding: 0.5rem;
  }

  .modal-content {
    width: 95%;
  }
}
</style>
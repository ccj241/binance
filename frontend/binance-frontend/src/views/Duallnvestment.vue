<template>
  <div class="dual-investment-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">
        <span class="gradient-text">双币投资</span>
      </h1>
      <p class="page-subtitle">高收益结构化理财产品</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%)">
          <i>💰</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ formatCurrency(stats.totalInvested) }}</div>
          <div class="stat-label">总投资金额</div>
        </div>
        <div class="stat-bg"></div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)">
          <i>📈</i>
        </div>
        <div class="stat-content">
          <div class="stat-value" :class="stats.totalPnL >= 0 ? 'positive' : 'negative'">
            {{ formatCurrency(stats.totalPnL) }}
          </div>
          <div class="stat-label">总盈亏</div>
        </div>
        <div class="stat-bg"></div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #fa709a 0%, #fee140 100%)">
          <i>🎯</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.winRate?.toFixed(1) || 0 }}%</div>
          <div class="stat-label">胜率</div>
        </div>
        <div class="stat-bg"></div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%)">
          <i>⚡</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.activeOrders || 0 }}</div>
          <div class="stat-label">活跃订单</div>
        </div>
        <div class="stat-bg"></div>
      </div>
    </div>

    <!-- Tab 切换 -->
    <div class="tabs">
      <button
          v-for="tab in tabs"
          :key="tab.key"
          @click="activeTab = tab.key"
          :class="['tab-btn', { active: activeTab === tab.key }]"
      >
        <i>{{ tab.icon }}</i>
        {{ tab.label }}
      </button>
    </div>

    <!-- 产品市场 -->
    <div v-if="activeTab === 'market'" class="section">
      <div class="section-header">
        <h2 class="section-title">产品市场</h2>
        <div class="filters">
          <select v-model="filters.symbol" class="filter-select">
            <option value="">所有交易对</option>
            <option value="BTCUSDT">BTC/USDT</option>
            <option value="ETHUSDT">ETH/USDT</option>
            <option value="BNBUSDT">BNB/USDT</option>
          </select>
          <select v-model="filters.direction" class="filter-select">
            <option value="">所有方向</option>
            <option value="UP">看涨</option>
            <option value="DOWN">看跌</option>
          </select>
          <input
              v-model.number="filters.minApy"
              type="number"
              placeholder="最低年化 %"
              class="filter-input"
          />
          <button @click="fetchProducts" class="filter-btn">
            <i>🔍</i> 搜索
          </button>
        </div>
      </div>

      <div v-if="loadingProducts" class="loading">
        <div class="loading-spinner"></div>
        <p>加载产品中...</p>
      </div>

      <div v-else-if="products.length === 0" class="empty-state">
        <div class="empty-icon">📦</div>
        <p>暂无可投资产品</p>
      </div>

      <div v-else class="products-grid">
        <div v-for="product in products" :key="product.id" class="product-card">
          <div class="product-header">
            <div class="product-symbol">{{ product.symbol }}</div>
            <div :class="['product-direction', product.direction.toLowerCase()]">
              <i>{{ product.direction === 'UP' ? '📈' : '📉' }}</i>
              {{ product.direction === 'UP' ? '看涨' : '看跌' }}
            </div>
          </div>

          <div class="product-info">
            <div class="info-row">
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
              <span class="label">价格偏离</span>
              <span class="value">
                {{ ((product.strikePrice - product.currentPrice) / product.currentPrice * 100).toFixed(2) }}%
              </span>
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
            <button @click="showInvestModal(product)" class="invest-btn">
              <i>💸</i> 立即投资
            </button>
            <button @click="showSimulateModal(product)" class="simulate-btn">
              <i>🧮</i> 收益计算
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 我的策略 -->
    <div v-if="activeTab === 'strategies'" class="section">
      <div class="section-header">
        <h2 class="section-title">我的策略</h2>
        <button @click="showStrategyModal()" class="create-btn">
          <i>➕</i> 创建策略
        </button>
      </div>

      <div v-if="strategies.length === 0" class="empty-state">
        <div class="empty-icon">🎯</div>
        <p>暂无投资策略</p>
        <button @click="showStrategyModal()" class="empty-action">
          创建第一个策略
        </button>
      </div>

      <div v-else class="strategies-grid">
        <div v-for="strategy in strategies" :key="strategy.id" class="strategy-card">
          <div class="strategy-header">
            <h3>{{ strategy.strategyName }}</h3>
            <div class="strategy-status">
              <label class="switch">
                <input
                    type="checkbox"
                    :checked="strategy.enabled"
                    @change="toggleStrategy(strategy)"
                />
                <span class="slider"></span>
              </label>
            </div>
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
              <span class="value">
                {{ strategy.targetApyMin }}% - {{ strategy.targetApyMax }}%
              </span>
            </div>
            <div class="info-item">
              <span class="label">已投资/限额</span>
              <span class="value">
                {{ formatCurrency(strategy.currentInvested) }} /
                {{ formatCurrency(strategy.totalInvestmentLimit) }}
              </span>
            </div>
            <div class="info-item">
              <span class="label">最后执行</span>
              <span class="value">
                {{ strategy.lastExecutedAt ? formatDate(strategy.lastExecutedAt) : '未执行' }}
              </span>
            </div>
          </div>

          <div class="strategy-actions">
            <button @click="editStrategy(strategy)" class="action-btn edit">
              <i>✏️</i> 编辑
            </button>
            <button @click="deleteStrategy(strategy)" class="action-btn delete">
              <i>🗑️</i> 删除
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 我的订单 -->
    <div v-if="activeTab === 'orders'" class="section">
      <div class="section-header">
        <h2 class="section-title">我的订单</h2>
        <select v-model="orderFilter" class="filter-select">
          <option value="">所有订单</option>
          <option value="active">活跃订单</option>
          <option value="settled">已结算</option>
        </select>
      </div>

      <div v-if="orders.length === 0" class="empty-state">
        <div class="empty-icon">📋</div>
        <p>暂无订单记录</p>
      </div>

      <div v-else class="orders-table">
        <table>
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
                  {{ order.direction === 'UP' ? '看涨' : '看跌' }}
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
                  {{ formatCurrency(order.pnl) }}
                  ({{ order.pnlPercent?.toFixed(2) }}%)
                </span>
              <span v-else>-</span>
            </td>
          </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 投资弹窗 -->
    <div v-if="showInvestDialog" class="modal-overlay" @click="closeInvestModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>投资产品</h3>
          <button @click="closeInvestModal" class="close-btn">✕</button>
        </div>

        <div class="modal-body">
          <div class="invest-product-info">
            <h4>{{ selectedProduct.symbol }} - {{ selectedProduct.direction === 'UP' ? '看涨' : '看跌' }}</h4>
            <div class="product-details">
              <div class="detail-row">
                <span>年化收益率：</span>
                <span class="highlight">{{ selectedProduct.apy.toFixed(2) }}%</span>
              </div>
              <div class="detail-row">
                <span>执行价格：</span>
                <span>{{ formatPrice(selectedProduct.strikePrice) }}</span>
              </div>
              <div class="detail-row">
                <span>投资期限：</span>
                <span>{{ selectedProduct.duration }}天</span>
              </div>
            </div>
          </div>

          <div class="invest-form">
            <div class="form-group">
              <label>投资金额</label>
              <input
                  v-model.number="investAmount"
                  type="number"
                  :min="selectedProduct.minAmount"
                  :max="selectedProduct.maxAmount"
                  :placeholder="`${selectedProduct.minAmount} - ${selectedProduct.maxAmount}`"
              />
            </div>

            <div class="form-group">
              <label>关联策略（可选）</label>
              <select v-model="investStrategyId">
                <option :value="null">不关联策略</option>
                <option v-for="s in strategies" :key="s.id" :value="s.id">
                  {{ s.strategyName }}
                </option>
              </select>
            </div>

            <div class="risk-warning">
              <i>⚠️</i>
              <p>风险提示：双币投资产品不保本，到期可能以其他币种结算</p>
            </div>
          </div>
        </div>

        <div class="modal-footer">
          <button @click="confirmInvest" class="confirm-btn" :disabled="!isInvestValid">
            确认投资
          </button>
          <button @click="closeInvestModal" class="cancel-btn">取消</button>
        </div>
      </div>
    </div>

    <!-- 策略弹窗 -->
    <div v-if="showStrategyDialog" class="modal-overlay" @click="closeStrategyModal">
      <div class="modal-content large" @click.stop>
        <div class="modal-header">
          <h3>{{ editingStrategy ? '编辑策略' : '创建策略' }}</h3>
          <button @click="closeStrategyModal" class="close-btn">✕</button>
        </div>

        <div class="modal-body">
          <form @submit.prevent="saveStrategy" class="strategy-form">
            <div class="form-grid">
              <div class="form-group">
                <label>策略名称</label>
                <input v-model="strategyForm.strategyName" type="text" required />
              </div>

              <div class="form-group">
                <label>策略类型</label>
                <select v-model="strategyForm.strategyType" required>
                  <option value="single">单次投资</option>
                  <option value="auto_reinvest">自动复投</option>
                  <option value="ladder">梯度投资</option>
                  <option value="price_trigger">价格触发</option>
                </select>
              </div>

              <div class="form-group">
                <label>基础资产</label>
                <select v-model="strategyForm.baseAsset" required>
                  <option value="BTC">BTC</option>
                  <option value="ETH">ETH</option>
                  <option value="BNB">BNB</option>
                </select>
              </div>

              <div class="form-group">
                <label>计价资产</label>
                <select v-model="strategyForm.quoteAsset" required>
                  <option value="USDT">USDT</option>
                  <option value="BUSD">BUSD</option>
                </select>
              </div>

              <div class="form-group">
                <label>方向偏好</label>
                <select v-model="strategyForm.directionPreference" required>
                  <option value="UP">只做看涨</option>
                  <option value="DOWN">只做看跌</option>
                  <option value="BOTH">双向都做</option>
                </select>
              </div>

              <div class="form-group">
                <label>目标年化范围 (%)</label>
                <div class="input-group">
                  <input
                      v-model.number="strategyForm.targetApyMin"
                      type="number"
                      min="0"
                      placeholder="最小"
                      required
                  />
                  <span>-</span>
                  <input
                      v-model.number="strategyForm.targetApyMax"
                      type="number"
                      min="0"
                      placeholder="最大"
                      required
                  />
                </div>
              </div>

              <div class="form-group">
                <label>单笔最大金额</label>
                <input v-model.number="strategyForm.maxSingleAmount" type="number" min="0" required />
              </div>

              <div class="form-group">
                <label>总投资限额</label>
                <input v-model.number="strategyForm.totalInvestmentLimit" type="number" min="0" required />
              </div>

              <div class="form-group">
                <label>最大执行价格偏离度 (%)</label>
                <input v-model.number="strategyForm.maxStrikePriceOffset" type="number" min="0" max="100" />
              </div>

              <div class="form-group">
                <label>投资期限范围（天）</label>
                <div class="input-group">
                  <input
                      v-model.number="strategyForm.minDuration"
                      type="number"
                      min="1"
                      placeholder="最小"
                  />
                  <span>-</span>
                  <input
                      v-model.number="strategyForm.maxDuration"
                      type="number"
                      min="1"
                      placeholder="最大"
                  />
                </div>
              </div>

              <div class="form-group full-width">
                <label>
                  <input v-model="strategyForm.autoReinvest" type="checkbox" />
                  自动复投
                </label>
              </div>
            </div>

            <!-- 价格触发策略参数 -->
            <div v-if="strategyForm.strategyType === 'price_trigger'" class="additional-params">
              <h4>价格触发参数</h4>
              <div class="form-grid">
                <div class="form-group">
                  <label>触发价格</label>
                  <input v-model.number="strategyForm.triggerPrice" type="number" min="0" required />
                </div>
                <div class="form-group">
                  <label>触发类型</label>
                  <select v-model="strategyForm.triggerType" required>
                    <option value="above">高于</option>
                    <option value="below">低于</option>
                  </select>
                </div>
              </div>
            </div>

            <!-- 梯度策略参数 -->
            <div v-if="strategyForm.strategyType === 'ladder'" class="additional-params">
              <h4>梯度投资参数</h4>
              <div class="form-grid">
                <div class="form-group">
                  <label>梯度层数</label>
                  <input v-model.number="strategyForm.ladderSteps" type="number" min="1" max="10" required />
                </div>
                <div class="form-group">
                  <label>每层价格间隔 (%)</label>
                  <input v-model.number="strategyForm.ladderStepPercent" type="number" min="0.1" max="10" step="0.1" required />
                </div>
              </div>
            </div>
          </form>
        </div>

        <div class="modal-footer">
          <button @click="saveStrategy" class="confirm-btn">
            {{ editingStrategy ? '保存修改' : '创建策略' }}
          </button>
          <button @click="closeStrategyModal" class="cancel-btn">取消</button>
        </div>
      </div>
    </div>

    <!-- 收益模拟弹窗 -->
    <div v-if="showSimulateDialog" class="modal-overlay" @click="closeSimulateModal">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>收益模拟计算</h3>
          <button @click="closeSimulateModal" class="close-btn">✕</button>
        </div>

        <div class="modal-body">
          <div class="simulate-form">
            <div class="form-group">
              <label>投资金额</label>
              <input v-model.number="simulateForm.investAmount" type="number" min="0" />
            </div>
            <button @click="runSimulation" class="simulate-btn">
              <i>🧮</i> 计算收益
            </button>
          </div>

          <div v-if="simulationResult" class="simulation-result">
            <h4>模拟结果</h4>

            <div class="result-scenario">
              <h5>情况1：价格未触及执行价</h5>
              <div class="result-info">
                <p>结算币种：{{ simulationResult.noTouch.settlementAsset }}</p>
                <p>结算金额：{{ formatCurrency(simulationResult.noTouch.settlementAmount) }}</p>
                <p>收益：{{ formatCurrency(simulationResult.noTouch.profit) }}
                  ({{ simulationResult.noTouch.profitPercent.toFixed(2) }}%)</p>
                <p class="description">{{ simulationResult.noTouch.description }}</p>
              </div>
            </div>

            <div class="result-scenario">
              <h5>情况2：价格触及执行价</h5>
              <div class="result-info">
                <p>结算币种：{{ simulationResult.touched.settlementAsset }}</p>
                <p>结算金额：{{ formatCurrency(simulationResult.touched.settlementAmount) }}</p>
                <p class="description">{{ simulationResult.touched.description }}</p>
              </div>
            </div>

            <div class="risk-tips">
              <h5>风险提示</h5>
              <ul>
                <li v-for="(risk, index) in simulationResult.risks" :key="index">
                  {{ risk }}
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Toast 消息 -->
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
  name: 'DualInvestment',
  data() {
    return {
      activeTab: 'market',
      tabs: [
        { key: 'market', label: '产品市场', icon: '🏪' },
        { key: 'strategies', label: '我的策略', icon: '🎯' },
        { key: 'orders', label: '我的订单', icon: '📋' }
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
        maxPositionRatio: 20,
        autoReinvest: false,
        triggerPrice: 0,
        triggerType: 'above',
        ladderSteps: 5,
        ladderStepPercent: 1
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

      // 模拟弹窗
      showSimulateDialog: false,
      selectedSimulateProduct: {},
      simulateForm: {
        investAmount: 1000
      },
      simulationResult: null,

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
      return this.investAmount >= this.selectedProduct.minAmount &&
          this.investAmount <= this.selectedProduct.maxAmount;
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
      return new Date(dateString).toLocaleString('zh-CN');
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
        'UP': '看涨',
        'DOWN': '看跌',
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
        maxPositionRatio: 20,
        autoReinvest: false,
        triggerPrice: 0,
        triggerType: 'above',
        ladderSteps: 5,
        ladderStepPercent: 1
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

    // 模拟相关方法
    showSimulateModal(product) {
      this.selectedSimulateProduct = product;
      this.simulateForm.investAmount = product.minAmount;
      this.simulationResult = null;
      this.showSimulateDialog = true;
    },

    closeSimulateModal() {
      this.showSimulateDialog = false;
      this.selectedSimulateProduct = {};
      this.simulationResult = null;
    },

    async runSimulation() {
      try {
        const response = await axios.post('/dual-investment/simulate', {
          investAmount: this.simulateForm.investAmount,
          strikePrice: this.selectedSimulateProduct.strikePrice,
          currentPrice: this.selectedSimulateProduct.currentPrice,
          apy: this.selectedSimulateProduct.apy,
          duration: this.selectedSimulateProduct.duration,
          direction: this.selectedSimulateProduct.direction,
          investAsset: this.selectedSimulateProduct.baseAsset
        }, {
          headers: this.getAuthHeaders()
        });

        this.simulationResult = response.data.simulation;
      } catch (error) {
        console.error('模拟计算失败:', error);
        this.showToast('模拟计算失败', 'error');
      }
    }
  }
};
</script>

<style scoped>
/* 容器样式 */
.dual-investment-container {
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

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 3rem;
}

.stat-card {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 2rem;
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-5px);
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  margin-bottom: 1rem;
}

.stat-content {
  position: relative;
  z-index: 1;
}

.stat-value {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.stat-value.positive {
  color: #22c55e;
}

.stat-value.negative {
  color: #ef4444;
}

.stat-label {
  color: #999;
  font-size: 0.9rem;
}

.stat-bg {
  position: absolute;
  top: -50%;
  right: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255,255,255,0.03) 0%, transparent 70%);
  transform: rotate(45deg);
}

/* Tab 切换 */
.tabs {
  display: flex;
  gap: 1rem;
  margin-bottom: 2rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding-bottom: 1rem;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1.5rem;
  background: none;
  border: none;
  color: #666;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
}

.tab-btn:hover {
  color: #fff;
}

.tab-btn.active {
  color: #667eea;
}

.tab-btn.active::after {
  content: '';
  position: absolute;
  bottom: -1rem;
  left: 0;
  right: 0;
  height: 2px;
  background: #667eea;
}

.tab-btn i {
  font-style: normal;
  font-size: 1.2rem;
}

/* Section 样式 */
.section {
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
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

/* 过滤器样式 */
.filters {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.filter-select,
.filter-input {
  padding: 0.8rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 0.9rem;
  transition: all 0.3s ease;
}

.filter-select:focus,
.filter-input:focus {
  outline: none;
  background: rgba(255, 255, 255, 0.08);
  border-color: #667eea;
}

.filter-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.filter-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
}

/* 产品网格 */
.products-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1.5rem;
}

.product-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.product-card:hover {
  background: rgba(255, 255, 255, 0.08);
  transform: translateY(-2px);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.product-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.product-symbol {
  font-size: 1.2rem;
  font-weight: 700;
  color: #fff;
}

.product-direction {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.4rem 0.8rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
}

.product-direction.up {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.product-direction.down {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.product-info {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
  margin-bottom: 1.5rem;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-row .label {
  color: #999;
  font-size: 0.9rem;
}

.info-row .value {
  color: #fff;
  font-weight: 500;
}

.info-row .value.apy {
  color: #fbbf24;
  font-size: 1.1rem;
  font-weight: 700;
}

.product-actions {
  display: flex;
  gap: 0.5rem;
}

.invest-btn,
.simulate-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.8rem;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.invest-btn {
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
  color: white;
}

.invest-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(34, 197, 94, 0.4);
}

.simulate-btn {
  background: rgba(139, 92, 246, 0.1);
  color: #8b5cf6;
  border: 1px solid rgba(139, 92, 246, 0.3);
}

.simulate-btn:hover {
  background: rgba(139, 92, 246, 0.2);
}

/* 策略网格 */
.strategies-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 1.5rem;
}

.strategy-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.strategy-card:hover {
  background: rgba(255, 255, 255, 0.08);
  transform: translateY(-2px);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.strategy-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.strategy-header h3 {
  margin: 0;
  font-size: 1.1rem;
  color: #fff;
}

/* 开关样式 */
.switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 24px;
}

.switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(255, 255, 255, 0.1);
  transition: .4s;
  border-radius: 24px;
}

.slider:before {
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

input:checked + .slider {
  background-color: #667eea;
}

input:checked + .slider:before {
  transform: translateX(26px);
}

.strategy-info {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.info-item .label {
  color: #999;
  font-size: 0.8rem;
}

.info-item .value {
  color: #fff;
  font-weight: 500;
}

.strategy-actions {
  display: flex;
  gap: 0.5rem;
}

.action-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
  padding: 0.6rem 0.8rem;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-btn.edit {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.action-btn.edit:hover {
  background: rgba(59, 130, 246, 0.2);
}

.action-btn.delete {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.action-btn.delete:hover {
  background: rgba(239, 68, 68, 0.2);
}

/* 订单表格 */
.orders-table {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  overflow: hidden;
}

.orders-table table {
  width: 100%;
  border-collapse: collapse;
}

.orders-table th {
  background: rgba(255, 255, 255, 0.05);
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  color: #ccc;
  font-size: 0.9rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.orders-table td {
  padding: 1rem;
  color: #ccc;
  font-size: 0.9rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
}

.orders-table tr:hover td {
  background: rgba(255, 255, 255, 0.03);
}

.direction-badge {
  padding: 0.2rem 0.5rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
}

.direction-badge.up {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

.direction-badge.down {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.status-badge {
  padding: 0.2rem 0.5rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-badge.active {
  background: rgba(59, 130, 246, 0.2);
  color: #3b82f6;
}

.status-badge.settled {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

.status-badge.cancelled {
  background: rgba(108, 117, 125, 0.2);
  color: #94a3b8;
}

.positive {
  color: #22c55e;
}

.negative {
  color: #ef4444;
}

/* 创建按钮 */
.create-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.create-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  color: #666;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.3;
}

.empty-action {
  margin-top: 1rem;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.empty-action:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
}

/* 加载状态 */
.loading {
  text-align: center;
  padding: 4rem;
  color: #666;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  margin: 0 auto 1rem;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: #667eea;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* 弹窗样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(5px);
}

.modal-content {
  background: #1a1a1a;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  max-width: 500px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
}

.modal-content.large {
  max-width: 800px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 2rem 2rem 1rem 2rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.modal-header h3 {
  margin: 0;
  color: #fff;
  font-size: 1.3rem;
  font-weight: 600;
}

.close-btn {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.1);
  color: #ccc;
  border: none;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  color: #fff;
}

.modal-body {
  padding: 2rem;
}

.modal-footer {
  padding: 1rem 2rem 2rem 2rem;
  display: flex;
  gap: 1rem;
  justify-content: flex-end;
}

/* 投资弹窗 */
.invest-product-info {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1.5rem;
  margin-bottom: 2rem;
}

.invest-product-info h4 {
  margin: 0 0 1rem 0;
  color: #fff;
  font-size: 1.1rem;
}

.product-details {
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  color: #ccc;
}

.detail-row .highlight {
  color: #fbbf24;
  font-weight: 600;
  font-size: 1.1rem;
}

.invest-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group label {
  font-weight: 600;
  color: #ccc;
  font-size: 0.9rem;
}

.form-group input,
.form-group select {
  padding: 0.8rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 0.9rem;
  transition: all 0.3s ease;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  background: rgba(255, 255, 255, 0.08);
  border-color: #667eea;
}

.risk-warning {
  background: rgba(255, 193, 7, 0.1);
  border: 1px solid rgba(255, 193, 7, 0.3);
  border-radius: 12px;
  padding: 1rem;
  display: flex;
  gap: 0.8rem;
  align-items: flex-start;
}

.risk-warning i {
  font-style: normal;
  font-size: 1.2rem;
  color: #fbbf24;
  flex-shrink: 0;
}

.risk-warning p {
  margin: 0;
  color: #fbbf24;
  font-size: 0.9rem;
  line-height: 1.4;
}

/* 策略表单 */
.strategy-form {
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
}

.form-group.full-width {
  grid-column: 1 / -1;
}

.input-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.input-group input {
  flex: 1;
}

.input-group span {
  color: #999;
}

.additional-params {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1.5rem;
}

.additional-params h4 {
  margin: 0 0 1rem 0;
  color: #fff;
  font-size: 1rem;
}

/* 模拟弹窗 */
.simulate-form {
  display: flex;
  gap: 1rem;
  align-items: flex-end;
  margin-bottom: 2rem;
}

.simulate-form .form-group {
  flex: 1;
}

.simulation-result {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.simulation-result h4 {
  margin: 0 0 1rem 0;
  color: #fff;
  font-size: 1.1rem;
}

.result-scenario {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1.5rem;
}

.result-scenario h5 {
  margin: 0 0 1rem 0;
  color: #667eea;
  font-size: 1rem;
}

.result-info p {
  margin: 0.5rem 0;
  color: #ccc;
  line-height: 1.5;
}

.result-info .description {
  color: #999;
  font-style: italic;
  font-size: 0.9rem;
}

.risk-tips {
  background: rgba(239, 68, 68, 0.1);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 12px;
  padding: 1.5rem;
}

.risk-tips h5 {
  margin: 0 0 1rem 0;
  color: #ef4444;
  font-size: 1rem;
}

.risk-tips ul {
  margin: 0;
  padding-left: 1.5rem;
}

.risk-tips li {
  color: #f87171;
  margin: 0.5rem 0;
  line-height: 1.5;
}

/* 按钮样式 */
.confirm-btn {
  padding: 0.8rem 2rem;
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.confirm-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(34, 197, 94, 0.4);
}

.confirm-btn:disabled {
  background: #6c757d;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.cancel-btn {
  padding: 0.8rem 2rem;
  background: rgba(108, 117, 125, 0.1);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.cancel-btn:hover {
  background: rgba(108, 117, 125, 0.2);
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
  z-index: 2000;
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
  font-style: normal;
  font-size: 1.2rem;
}

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
  .dual-investment-container {
    padding: 1rem;
  }

  .page-title {
    font-size: 2rem;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .tabs {
    flex-wrap: wrap;
  }

  .filters {
    flex-wrap: wrap;
  }

  .products-grid,
  .strategies-grid {
    grid-template-columns: 1fr;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .orders-table {
    overflow-x: auto;
  }

  .orders-table table {
    min-width: 800px;
  }

  .modal-content {
    width: 95%;
    max-height: 90vh;
  }

  .modal-header,
  .modal-body,
  .modal-footer {
    padding: 1.5rem;
  }

  .toast {
    left: 1rem;
    right: 1rem;
    bottom: 1rem;
  }
}
</style>
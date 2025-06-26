<template>
  <div class="dual-investment-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1 class="page-title">双币投资</h1>
      <p class="page-description">高收益结构化理财产品，把握市场机会</p>
    </div>
    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">
          <span>💰</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">总投资金额</div>
          <div class="stat-value">{{ formatCurrency(stats.totalInvested) }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">
          <span>📈</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">累计收益</div>
          <div class="stat-value" :class="stats.totalPnL >= 0 ? 'positive' : 'negative'">
            {{ stats.totalPnL >= 0 ? '+' : '' }}{{ formatCurrency(Math.abs(stats.totalPnL)) }}
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">
          <span>🎯</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">胜率</div>
          <div class="stat-value">{{ stats.winRate?.toFixed(1) || 0 }}%</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">
          <span>⏳</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">进行中</div>
          <div class="stat-value">{{ stats.activeOrders || 0 }}</div>
        </div>
      </div>
    </div>

    <!-- Tab 导航 -->
    <div class="tab-container">
      <div class="tab-nav">
        <button
            v-for="tab in tabs"
            :key="tab.key"
            @click="activeTab = tab.key"
            :class="['tab-btn', { active: activeTab === tab.key }]"
        >
          <span class="tab-icon">{{ tab.icon }}</span>
          <span>{{ tab.label }}</span>
        </button>
      </div>
    </div>

    <!-- 产品选择 -->
    <div v-show="activeTab === 'market'" class="content-section">
      <!-- 步骤1：选择交易对 -->
      <div v-if="!selectedSymbol" class="symbol-selection">
        <div class="section-header">
          <h2 class="section-title">选择交易对</h2>
          <p class="section-desc">请选择您想要投资的交易对</p>
        </div>

        <div class="symbol-grid">
          <div
              v-for="symbol in availableSymbols"
              :key="symbol.symbol"
              @click="selectSymbol(symbol)"
              class="symbol-card"
          >
            <div class="symbol-header">
              <div class="coin-icon">{{ symbol.icon }}</div>
              <h3 class="symbol-name">{{ symbol.displaySymbol }}</h3>
            </div>
            <div class="symbol-info">
              <div class="info-item">
                <span class="label">当前价格</span>
                <span class="value">{{ formatPrice(symbol.currentPrice) }}</span>
              </div>
              <div class="info-item">
                <span class="label">24h涨跌</span>
                <span class="value" :class="symbol.change24h >= 0 ? 'positive' : 'negative'">
              {{ symbol.change24h >= 0 ? '+' : '' }}{{ symbol.change24h.toFixed(2) }}%
            </span>
              </div>
              <div class="info-item">
                <span class="label">可用产品</span>
                <span class="value">{{ symbol.productCount }} 个</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 步骤2：选择产品 -->
      <div v-else class="product-selection">
        <div class="section-header">
          <div class="header-content">
            <h2 class="section-title">{{ selectedSymbol.displaySymbol }} 投资产品</h2>
            <p class="section-desc">选择合适的到期日和执行价格</p>
          </div>
          <button @click="selectedSymbol = null" class="back-btn">
            <span>←</span> 返回选择交易对
          </button>
        </div>

        <!-- 产品筛选 -->
        <div class="filter-bar">
          <div class="filter-group">
            <label>方向</label>
            <select v-model="productFilter.direction" class="filter-select">
              <option value="">全部</option>
              <option value="UP">看涨</option>
              <option value="DOWN">看跌</option>
            </select>
          </div>
          <div class="filter-group">
            <label>期限</label>
            <select v-model="productFilter.duration" class="filter-select">
              <option value="">全部</option>
              <option value="7">7天</option>
              <option value="14">14天</option>
              <option value="30">30天</option>
            </select>
          </div>
          <div class="filter-group">
            <label>最低年化</label>
            <input
                v-model.number="productFilter.minApy"
                type="number"
                placeholder="如：20"
                class="filter-input"
            />
          </div>
        </div>

        <!-- 产品列表 -->
        <div v-if="loadingProducts" class="loading-state">
          <div class="spinner"></div>
          <p>加载产品中...</p>
        </div>

        <div v-else-if="filteredProducts.length === 0" class="empty-state">
          <span class="empty-icon">📭</span>
          <p>暂无符合条件的产品</p>
        </div>

        <div v-else class="products-list">
          <div
              v-for="product in filteredProducts"
              :key="product.id"
              class="product-card"
          >
            <div class="product-header">
              <div class="product-badge" :class="product.direction.toLowerCase()">
                {{ product.direction === 'UP' ? '看涨' : '看跌' }}
              </div>
              <div class="product-apy">
                <span class="apy-value">{{ product.apy.toFixed(2) }}%</span>
                <span class="apy-label">年化收益</span>
              </div>
            </div>

            <div class="product-details">
              <div class="detail-item">
                <span class="label">执行价格</span>
                <span class="value">{{ formatPrice(product.strikePrice) }}</span>
              </div>
              <div class="detail-item">
                <span class="label">到期日</span>
                <span class="value">{{ formatDate(product.settlementTime) }}</span>
              </div>
              <div class="detail-item">
                <span class="label">期限</span>
                <span class="value">{{ product.duration }}天</span>
              </div>
              <div class="detail-item">
                <span class="label">投资范围</span>
                <span class="value">{{ product.minAmount }} - {{ product.maxAmount }}</span>
              </div>
            </div>

            <button @click="showInvestModal(product)" class="invest-btn">
              立即投资
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 我的策略 -->
    <div v-show="activeTab === 'strategies'" class="content-section">
      <div class="section-header">
        <h2 class="section-title">我的策略</h2>
        <button @click="showStrategyModal()" class="btn btn-primary">
          <span>+</span> 创建策略
        </button>
      </div>

      <div v-if="strategies.length === 0" class="empty-state">
        <span class="empty-icon">📋</span>
        <p>暂无投资策略</p>
        <button @click="showStrategyModal()" class="btn btn-primary">
          创建第一个策略
        </button>
      </div>

      <div v-else class="strategies-grid">
        <div v-for="strategy in strategies" :key="strategy.id" class="strategy-card">
          <div class="strategy-header">
            <h3 class="strategy-name">{{ strategy.strategyName }}</h3>
            <div class="toggle-switch">
              <input
                  type="checkbox"
                  :id="`strategy-${strategy.id}`"
                  :checked="strategy.enabled"
                  @change="toggleStrategy(strategy)"
              />
              <label :for="`strategy-${strategy.id}`"></label>
            </div>
          </div>

          <div class="strategy-info">
            <div class="info-row">
              <span class="label">策略类型</span>
              <span class="value">{{ getStrategyTypeText(strategy.strategyType) }}</span>
            </div>
            <div class="info-row">
              <span class="label">交易对</span>
              <span class="value">{{ strategy.baseAsset }}/{{ strategy.quoteAsset }}</span>
            </div>
            <div class="info-row">
              <span class="label">方向偏好</span>
              <span class="value">{{ getDirectionText(strategy.directionPreference) }}</span>
            </div>
            <div class="info-row" v-if="strategy.basePrice > 0">
              <span class="label">基准价格</span>
              <span class="value">{{ formatPrice(strategy.basePrice) }}</span>
            </div>
            <div class="info-row">
              <span class="label">目标年化</span>
              <span class="value">{{ strategy.targetApyMin }}% - {{ strategy.targetApyMax }}%</span>
            </div>
            <div class="info-row" v-if="strategy.strategyType === 'ladder'">
              <span class="label">梯度设置</span>
              <span class="value">{{ strategy.ladderSteps }}层深度</span>
            </div>
            <div class="info-row">
              <span class="label">投资进度</span>
              <div class="progress-bar">
                <div
                    class="progress-fill"
                    :style="{width: `${(strategy.currentInvested / strategy.totalInvestmentLimit) * 100}%`}"
                ></div>
              </div>
              <span class="value">
            {{ formatCurrency(strategy.currentInvested) }} / {{ formatCurrency(strategy.totalInvestmentLimit) }}
          </span>
            </div>
          </div>

          <div class="strategy-actions">
            <button @click="editStrategy(strategy)" class="btn btn-sm btn-outline">编辑</button>
            <button @click="viewStrategyStats(strategy)" class="btn btn-sm btn-outline">统计</button>
            <button @click="deleteStrategy(strategy)" class="btn btn-sm btn-danger">删除</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 我的订单 -->
    <div v-show="activeTab === 'orders'" class="content-section">
      <div class="section-header">
        <h2 class="section-title">我的订单</h2>
        <select v-model="orderFilter" class="filter-select">
          <option value="">全部订单</option>
          <option value="active">进行中</option>
          <option value="settled">已结算</option>
        </select>
      </div>

      <div v-if="orders.length === 0" class="empty-state">
        <span class="empty-icon">📄</span>
        <p>暂无订单记录</p>
      </div>

      <div v-else class="table-container">
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
            <th>到期时间</th>
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
            <td>{{ formatCurrency(order.investAmount) }}</td>
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

    <!-- 投资弹窗 -->
    <transition name="modal">
      <div v-if="showInvestDialog" class="modal-overlay" @click="closeInvestModal">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">确认投资</h3>
            <button @click="closeInvestModal" class="modal-close">×</button>
          </div>

          <div class="modal-body">
            <div class="product-summary">
              <h4>{{ selectedProduct.symbol }}</h4>
              <div class="summary-grid">
                <div class="summary-item">
                  <span class="label">方向</span>
                  <span class="value">{{ selectedProduct.direction === 'UP' ? '看涨' : '看跌' }}</span>
                </div>
                <div class="summary-item">
                  <span class="label">年化收益率</span>
                  <span class="value highlight">{{ selectedProduct.apy?.toFixed(2) }}%</span>
                </div>
                <div class="summary-item">
                  <span class="label">执行价格</span>
                  <span class="value">{{ formatPrice(selectedProduct.strikePrice) }}</span>
                </div>
                <div class="summary-item">
                  <span class="label">到期日</span>
                  <span class="value">{{ formatDate(selectedProduct.settlementTime) }}</span>
                </div>
              </div>
            </div>

            <div class="form-group">
              <label class="form-label">投资金额</label>
              <div class="input-group">
                <input
                    v-model.number="investAmount"
                    type="number"
                    class="form-control"
                    :min="selectedProduct.minAmount"
                    :max="selectedProduct.maxAmount"
                    :placeholder="`${selectedProduct.minAmount} - ${selectedProduct.maxAmount}`"
                />
                <span class="input-suffix">{{ selectedProduct.baseAsset }}</span>
              </div>
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
              <div>
                <p>风险提示</p>
                <p class="warning-text">双币投资产品不保本，到期时可能以另一种资产结算</p>
              </div>
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
                  <option value="UP">看涨(低买)</option>
                  <option value="DOWN">看跌(高卖)</option>
                  <option value="BOTH">双向</option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">基准价格</label>
                <input
                    v-model.number="strategyForm.basePrice"
                    type="number"
                    class="form-control"
                    min="0"
                    step="0.01"
                    placeholder="留空则使用当前价格"
                />
                <small class="form-hint">看涨时只在价格低于基准时投资，看跌时只在价格高于基准时投资</small>
              </div>

              <div class="form-group">
                <label class="form-label">目标年化范围 (%)</label>
                <div class="input-range">
                  <input
                      v-model.number="strategyForm.targetApyMin"
                      type="number"
                      class="form-control"
                      min="0"
                      placeholder="最小"
                      required
                  />
                  <span class="range-separator">-</span>
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

              <!-- 梯度策略参数 -->
              <template v-if="strategyForm.strategyType === 'ladder'">
                <div class="form-group">
                  <label class="form-label">梯度深度层数</label>
                  <input
                      v-model.number="strategyForm.ladderSteps"
                      type="number"
                      class="form-control"
                      min="1"
                      max="10"
                      required
                  />
                  <small class="form-hint">策略将自动选择币安提供的不同价格深度的产品进行投资</small>
                </div>
              </template>

              <!-- 价格触发参数 -->
              <template v-if="strategyForm.strategyType === 'price_trigger'">
                <div class="form-group">
                  <label class="form-label">触发价格</label>
                  <input
                      v-model.number="strategyForm.triggerPrice"
                      type="number"
                      class="form-control"
                      min="0"
                      required
                  />
                </div>
                <div class="form-group">
                  <label class="form-label">触发类型</label>
                  <select v-model="strategyForm.triggerType" class="form-control" required>
                    <option value="above">价格高于</option>
                    <option value="below">价格低于</option>
                  </select>
                </div>
              </template>

              <!-- 自动复投参数 -->
              <template v-if="strategyForm.strategyType === 'auto_reinvest'">
                <div class="form-group full-width">
                  <label class="checkbox-label">
                    <input
                        v-model="strategyForm.autoReinvest"
                        type="checkbox"
                    />
                    <span>启用自动复投</span>
                  </label>
                </div>
              </template>
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
        { key: 'market', label: '产品市场', icon: '🏪' },
        { key: 'strategies', label: '我的策略', icon: '🎯' },
        { key: 'orders', label: '我的订单', icon: '📋' }
      ],

      // 产品相关
      availableSymbols: [
        { symbol: 'BTCUSDT', displaySymbol: 'BTC/USDT', icon: '₿', currentPrice: 45000, change24h: 2.5, productCount: 0 },
        { symbol: 'ETHUSDT', displaySymbol: 'ETH/USDT', icon: 'Ξ', currentPrice: 3000, change24h: -1.2, productCount: 0 },
        { symbol: 'BNBUSDT', displaySymbol: 'BNB/USDT', icon: '🔸', currentPrice: 350, change24h: 0.8, productCount: 0 },
        { symbol: 'SOLUSDT', displaySymbol: 'SOL/USDT', icon: '◎', currentPrice: 120, change24h: 5.3, productCount: 0 }
      ],
      selectedSymbol: null,
      products: [],
      loadingProducts: false,
      productFilter: {
        direction: '',
        duration: '',
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
        autoReinvest: false,
        basePrice: null,
        triggerPrice: null,
        triggerType: 'above',
        ladderSteps: 5
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
    filteredProducts() {
      if (!this.products.length) return [];

      return this.products.filter(product => {
        if (this.productFilter.direction && product.direction !== this.productFilter.direction) {
          return false;
        }
        if (this.productFilter.duration && product.duration !== parseInt(this.productFilter.duration)) {
          return false;
        }
        if (this.productFilter.minApy && product.apy < this.productFilter.minApy) {
          return false;
        }
        return true;
      });
    },

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
    this.fetchStrategies();
    this.fetchOrders();
    this.fetchStats();
    this.updateSymbolPrices();
    this.fetchSymbolProductCounts();
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
        'UP': '看涨(低买)',
        'DOWN': '看跌(高卖)',
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

    // 更新交易对价格
    async updateSymbolPrices() {
      try {
        const response = await axios.get('/prices', {
          headers: this.getAuthHeaders()
        });

        const prices = response.data.prices || {};

        // 更新可用交易对的价格
        this.availableSymbols.forEach(symbol => {
          if (prices[symbol.symbol]) {
            symbol.currentPrice = prices[symbol.symbol];
          }
        });
      } catch (error) {
        console.error('获取价格失败:', error);
      }
    },

    // 获取每个交易对的产品数量
    async fetchSymbolProductCounts() {
      try {
        for (let symbol of this.availableSymbols) {
          const response = await axios.get(`/dual-investment/products?symbol=${symbol.symbol}`, {
            headers: this.getAuthHeaders()
          });
          symbol.productCount = response.data.products?.length || 0;
        }
      } catch (error) {
        console.error('获取产品数量失败:', error);
      }
    },

    // 产品相关方法
    selectSymbol(symbol) {
      this.selectedSymbol = symbol;
      this.fetchProducts(symbol.symbol);
    },

    async fetchProducts(symbol) {
      this.loadingProducts = true;
      try {
        const params = new URLSearchParams();
        params.append('symbol', symbol);

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
        autoReinvest: false,
        basePrice: null,
        triggerPrice: null,
        triggerType: 'above',
        ladderSteps: 5
      };
    },

    async saveStrategy() {
      try {
        // 根据策略类型清理不必要的参数
        const formData = { ...this.strategyForm };

        if (formData.strategyType !== 'ladder') {
          delete formData.ladderSteps;
        }

        if (formData.strategyType !== 'price_trigger') {
          delete formData.triggerPrice;
          delete formData.triggerType;
        }

        if (this.editingStrategy) {
          await axios.put(`/dual-investment/strategies/${this.editingStrategy.id}`,
              formData, {
                headers: this.getAuthHeaders()
              });
          this.showToast('策略更新成功！');
        } else {
          await axios.post('/dual-investment/strategies', formData, {
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

    async viewStrategyStats(strategy) {
      // 可以跳转到详细统计页面或显示弹窗
      this.showToast('功能开发中...', 'info');
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
    }
  }
};
</script>
<style scoped>
/* 页面容器 */
.dual-investment-container {
  max-width: 1400px;
  margin: 0 auto;
  padding: 0;
}

/* 页面头部 */
.page-header {
  text-align: center;
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
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.stat-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.stat-icon {
  width: 48px;
  height: 48px;
  background: var(--color-bg-tertiary);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
}

.stat-content {
  flex: 1;
}

.stat-label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
  margin-bottom: 0.25rem;
}

.stat-value {
  font-size: 1.5rem;
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
.tab-container {
  margin-bottom: 2rem;
}

.tab-nav {
  display: flex;
  gap: 0.5rem;
  border-bottom: 1px solid var(--color-border);
}

.tab-btn {
  padding: 0.75rem 1.5rem;
  background: none;
  border: none;
  border-bottom: 2px solid transparent;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.tab-btn:hover {
  color: var(--color-text-primary);
}

.tab-btn.active {
  color: var(--color-primary);
  border-bottom-color: var(--color-primary);
}

.tab-icon {
  font-size: 1rem;
}

/* 内容区域 */
.content-section {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 2rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.header-content {
  flex: 1;
}

.section-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.25rem 0;
}

.section-desc {
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}

/* 交易对选择 */
.symbol-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.symbol-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  cursor: pointer;
  transition: all var(--transition-normal);
}

.symbol-card:hover {
  border-color: var(--color-primary);
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.symbol-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1rem;
}

.coin-icon {
  width: 48px;
  height: 48px;
  background: var(--color-bg-tertiary);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
}

.symbol-name {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.symbol-info {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.info-item {
  display: flex;
  justify-content: space-between;
  font-size: 0.875rem;
}

.info-item .label {
  color: var(--color-text-secondary);
}

.info-item .value {
  color: var(--color-text-primary);
  font-weight: 500;
}

/* 返回按钮 */
.back-btn {
  padding: 0.5rem 1rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.back-btn:hover {
  background: var(--color-bg-secondary);
  color: var(--color-text-primary);
}

/* 筛选栏 */
.filter-bar {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
  padding: 1rem;
  background: var(--color-bg-secondary);
  border-radius: var(--radius-md);
}

.filter-group {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.filter-group label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.filter-select,
.filter-input {
  padding: 0.5rem 0.75rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-primary);
  font-size: 0.875rem;
  min-width: 120px;
}

.filter-select:focus,
.filter-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

/* 产品列表 */
.products-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1rem;
}

.product-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  transition: all var(--transition-normal);
}

.product-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
}

.product-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.product-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.product-badge.up {
  background: rgba(34, 197, 94, 0.1);
  color: var(--color-success);
}

.product-badge.down {
  background: rgba(239, 68, 68, 0.1);
  color: var(--color-danger);
}

.product-apy {
  text-align: right;
}

.apy-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--color-warning);
  display: block;
}

.apy-label {
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

.product-details {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-item .label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.detail-item .value {
  font-size: 0.875rem;
  color: var(--color-text-primary);
  font-weight: 500;
}

.invest-btn {
  width: 100%;
  padding: 0.75rem;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
}

.invest-btn:hover {
  background: var(--color-primary-hover);
}

/* 策略卡片 */
.strategies-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1rem;
}

.strategy-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
}

.strategy-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.strategy-name {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

/* 开关按钮 */
.toggle-switch {
  position: relative;
}

.toggle-switch input {
  display: none;
}

.toggle-switch label {
  display: block;
  width: 44px;
  height: 24px;
  background: var(--color-border);
  border-radius: 24px;
  cursor: pointer;
  transition: background var(--transition-normal);
  position: relative;
}

.toggle-switch label::after {
  content: '';
  position: absolute;
  top: 2px;
  left: 2px;
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  transition: transform var(--transition-normal);
}

.toggle-switch input:checked + label {
  background: var(--color-primary);
}

.toggle-switch input:checked + label::after {
  transform: translateX(20px);
}

.strategy-info {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.875rem;
}

.info-row .label {
  color: var(--color-text-secondary);
}

.info-row .value {
  color: var(--color-text-primary);
  font-weight: 500;
}

/* 进度条 */
.progress-bar {
  flex: 1;
  height: 4px;
  background: var(--color-bg-tertiary);
  border-radius: 2px;
  margin: 0 1rem;
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: var(--color-primary);
  border-radius: 2px;
  transition: width var(--transition-normal);
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
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
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

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.75rem;
}

/* 表格 */
.table-container {
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

/* 方向徽章 */
.direction-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.direction-badge.up {
  background: rgba(34, 197, 94, 0.1);
  color: var(--color-success);
}

.direction-badge.down {
  background: rgba(239, 68, 68, 0.1);
  color: var(--color-danger);
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
  background: rgba(251, 191, 36, 0.1);
  color: var(--color-warning);
}

.status-badge.active {
  background: rgba(37, 99, 235, 0.1);
  color: var(--color-primary);
}

.status-badge.settled {
  background: rgba(34, 197, 94, 0.1);
  color: var(--color-success);
}

.status-badge.cancelled {
  background: var(--color-bg-tertiary);
  color: var(--color-text-secondary);
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

/* 产品摘要 */
.product-summary {
  background: var(--color-bg-secondary);
  border-radius: var(--radius-md);
  padding: 1rem;
  margin-bottom: 1.5rem;
}

.product-summary h4 {
  margin: 0 0 1rem 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.summary-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.summary-item .label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.summary-item .value {
  font-size: 0.875rem;
  color: var(--color-text-primary);
  font-weight: 500;
}

.summary-item .value.highlight {
  color: var(--color-warning);
  font-size: 1.125rem;
  font-weight: 600;
}

/* 表单 */
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
}

.input-group .form-control {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}
.input-suffix {
  padding: 0.625rem 0.875rem;
  background: var(--color-bg-tertiary);
  border: 1px solid var(--color-border);
  border-left: 0;
  border-radius: 0 var(--radius-md) var(--radius-md) 0;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}
.input-range {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}
.range-separator {
  color: var(--color-text-tertiary);
}
/* 风险提示 */
.risk-warning {
  display: flex;
  gap: 0.75rem;
  padding: 0.75rem;
  background: rgba(251, 191, 36, 0.1);
  border: 1px solid rgba(251, 191, 36, 0.3);
  border-radius: var(--radius-md);
  margin-top: 1rem;
}
.warning-icon {
  font-size: 1.25rem;
  flex-shrink: 0;
}
.risk-warning p {
  margin: 0;
  font-size: 0.875rem;
}
.risk-warning p:first-child {
  font-weight: 500;
  color: var(--color-warning);
}
.warning-text {
  color: var(--color-text-secondary);
  font-size: 0.75rem;
}
/* 额外的表单样式 */
.form-hint {
  display: block;
  margin-top: 0.25rem;
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}
.full-width {
  grid-column: 1 / -1;
}
.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  font-weight: 500;
}
.checkbox-label input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
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
.toast.info {
  border-color: var(--color-primary);
  color: var(--color-primary);
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
  .symbol-grid {
    grid-template-columns: 1fr;
  }
  .filter-bar {
    flex-wrap: wrap;
  }
  .products-list {
    grid-template-columns: 1fr;
  }
  .strategies-grid {
    grid-template-columns: 1fr;
  }
  .product-details {
    grid-template-columns: 1fr;
  }
  .summary-grid {
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
  .tab-nav {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
    scrollbar-width: none;
  }
  .tab-nav::-webkit-scrollbar {
    display: none;
  }
  .stat-card {
    padding: 1rem;
  }
  .stat-icon {
    width: 40px;
    height: 40px;
    font-size: 1.25rem;
  }
  .stat-value {
    font-size: 1.25rem;
  }
}
@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  .content-section {
    padding: 1rem;
  }
  .modal-body {
    padding: 1rem;
  }
  .section-title {
    font-size: 1.125rem;
  }
  .strategy-actions {
    flex-wrap: wrap;
  }
  .btn-sm {
    flex: 1;
    min-width: 80px;
  }
}
</style>
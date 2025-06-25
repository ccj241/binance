<template>
  <div class="strategy-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">
        <span class="gradient-text">策略管理中心</span>
      </h1>
      <p class="page-subtitle">创建和管理您的交易策略</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #667eea 0%, #764ba2 100%)">
          <i>📊</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ strategies.length }}</div>
          <div class="stat-label">总策略数</div>
        </div>
        <div class="stat-bg"></div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)">
          <i>✅</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ activeStrategiesCount }}</div>
          <div class="stat-label">活跃策略</div>
        </div>
        <div class="stat-bg"></div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #fa709a 0%, #fee140 100%)">
          <i>⚡</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ executingStrategiesCount }}</div>
          <div class="stat-label">执行中策略</div>
        </div>
        <div class="stat-bg"></div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%)">
          <i>🎯</i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ completedStrategiesCount }}</div>
          <div class="stat-label">已完成策略</div>
        </div>
        <div class="stat-bg"></div>
      </div>
    </div>

    <!-- 消息提示 -->
    <transition name="toast">
      <div v-if="toastMessage" :class="['toast', toastType]">
        <i class="toast-icon">{{ toastType === 'success' ? '✅' : '❌' }}</i>
        <span>{{ toastMessage }}</span>
      </div>
    </transition>

    <!-- 创建策略区域 -->
    <div class="create-section">
      <div class="section-header">
        <h2 class="section-title">创建新策略</h2>
        <button @click="toggleCreateForm" class="toggle-btn">
          <i>{{ showCreateForm ? '🔽' : '➕' }}</i>
          {{ showCreateForm ? '收起' : '创建策略' }}
        </button>
      </div>

      <transition name="form-slide">
        <div v-if="showCreateForm" class="create-form">
          <form @submit.prevent="createStrategy">
            <div class="form-grid">
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
            <div v-if="newStrategy.strategyType" class="strategy-description">
              <div v-if="newStrategy.strategyType === 'simple'" class="description-card">
                <div class="description-icon">🎯</div>
                <div class="description-content">
                  <h4>简单策略</h4>
                  <p>当价格达到触发条件时，以基准价格一次性下单全部数量。</p>
                </div>
              </div>

              <div v-if="newStrategy.strategyType === 'iceberg'" class="description-card">
                <div class="description-icon">🧊</div>
                <div class="description-content">
                  <h4>冰山策略</h4>
                  <p>将订单分成多个小订单，基于基准价格按固定万分比在不同价格层级分批下单。</p>
                  <small>默认万分比：买单[0, -1, -3, -5, -7]，卖单[0, 1, 3, 5, 7]</small>
                </div>
              </div>

              <div v-if="newStrategy.strategyType === 'custom'" class="description-card">
                <div class="description-icon">⚙️</div>
                <div class="description-content">
                  <h4>自定义策略</h4>
                  <p>基于基准价格，按自定义万分比计算各档位价格进行分批下单。</p>
                  <small>万分比说明：正数表示高于基准价格，负数表示低于基准价格。例如：+50表示基准价格+0.5%，-30表示基准价格-0.3%</small>
                </div>
              </div>
            </div>

            <!-- 自定义策略配置 -->
            <div v-if="newStrategy.strategyType === 'custom'" class="custom-config">
              <h3>📝 自定义配置</h3>

              <div v-if="newStrategy.side === 'BUY'" class="config-section">
                <h4>🟢 买入配置</h4>
                <div class="config-grid">
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
                <h4>🔴 卖出配置</h4>
                <div class="config-grid">
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
                <i>⚠️</i> {{ quantityWarning }}
              </div>
            </div>

            <!-- 订单预览 -->
            <div v-if="orderPreview.length > 0" class="order-preview">
              <h3>📋 订单预览</h3>
              <div class="preview-grid">
                <div v-for="(order, index) in orderPreview" :key="index" class="preview-card">
                  <div class="preview-header">
                    <span class="order-number">订单 {{ index + 1 }}</span>
                    <span class="order-ratio">{{ (order.ratio * 100).toFixed(1) }}%</span>
                  </div>
                  <div class="preview-details">
                    <div class="detail-row">
                      <span class="label">数量:</span>
                      <span class="value">{{ order.quantity.toFixed(8) }}</span>
                    </div>
                    <div v-if="newStrategy.strategyType === 'custom'" class="detail-row">
                      <span class="label">万分比:</span>
                      <span class="value">{{ order.basisPoint > 0 ? '+' : '' }}{{ order.basisPoint }}bp</span>
                    </div>
                    <div class="detail-row">
                      <span class="label">预估价格:</span>
                      <span class="value">{{ order.price.toFixed(8) }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="form-actions">
              <button type="submit" :disabled="isCreatingStrategy || !isFormValid" class="create-btn">
                <i>{{ isCreatingStrategy ? '⏳' : '🚀' }}</i>
                {{ isCreatingStrategy ? '创建中...' : '创建策略' }}
              </button>
              <button type="button" @click="resetForm" class="reset-btn">
                <i>🔄</i> 重置表单
              </button>
            </div>
          </form>
        </div>
      </transition>
    </div>

    <!-- 策略列表 -->
    <div class="strategies-section">
      <div class="section-header">
        <h2 class="section-title">策略列表</h2>
        <div class="search-box">
          <i class="search-icon">🔍</i>
          <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索策略..."
              class="search-input"
          />
        </div>
      </div>

      <div v-if="filteredStrategies.length === 0" class="empty-state">
        <div class="empty-icon">📊</div>
        <p class="empty-text">暂无策略记录</p>
        <button @click="showCreateForm = true" class="empty-action">
          <i>➕</i> 创建第一个策略
        </button>
      </div>

      <div v-else class="strategies-grid">
        <div v-for="strategy in paginatedStrategies" :key="strategy.id" class="strategy-card">
          <div class="strategy-header">
            <div class="strategy-symbol">
              {{ strategy.symbol }}
            </div>
            <div class="strategy-status">
              <span :class="['status-chip', strategy.status]">
                <span class="status-dot"></span>
                {{ getStatusText(strategy.status) }}
              </span>
            </div>
          </div>

          <div class="strategy-meta">
            <div class="meta-item">
              <span class="meta-label">类型</span>
              <span class="meta-value">{{ getStrategyTypeText(strategy.strategyType) }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">方向</span>
              <span :class="['side-chip', strategy.side.toLowerCase()]">
                <i>{{ strategy.side === 'BUY' ? '📈' : '📉' }}</i>
                {{ strategy.side === 'BUY' ? '买入' : '卖出' }}
              </span>
            </div>
            <div class="meta-item">
              <span class="meta-label">基准价格</span>
              <span class="meta-value">{{ formatPrice(strategy.price) }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">总数量</span>
              <span class="meta-value">{{ strategy.totalQuantity }}</span>
            </div>
            <div class="meta-item">
              <span class="meta-label">启用状态</span>
              <span :class="['enable-chip', strategy.enabled ? 'enabled' : 'disabled']">
                <i>{{ strategy.enabled ? '✅' : '❌' }}</i>
                {{ strategy.enabled ? '启用' : '禁用' }}
              </span>
            </div>
            <div class="meta-item">
              <span class="meta-label">执行状态</span>
              <span :class="['exec-chip', strategy.pendingBatch ? 'executing' : 'idle']">
                <i>{{ strategy.pendingBatch ? '⚡' : '💤' }}</i>
                {{ strategy.pendingBatch ? '执行中' : '空闲' }}
              </span>
            </div>
            <div class="meta-item">
              <span class="meta-label">创建时间</span>
              <span class="meta-value">{{ formatDate(strategy.createdAt) }}</span>
            </div>
          </div>

          <div class="strategy-actions">
            <button
                @click="toggleStrategy(strategy)"
                :class="['action-btn', strategy.enabled ? 'disable' : 'enable']"
            >
              <i>{{ strategy.enabled ? '⏸️' : '▶️' }}</i>
              {{ strategy.enabled ? '禁用' : '启用' }}
            </button>

            <button @click="viewStrategyDetails(strategy)" class="action-btn view">
              <i>👁️</i> 查看详情
            </button>

            <button @click="viewStrategyStats(strategy)" class="action-btn stats">
              <i>📊</i> 统计信息
            </button>

            <button @click="deleteStrategy(strategy.id)" class="action-btn delete">
              <i>🗑️</i> 删除
            </button>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div class="pagination" v-if="strategies.length > pageSize">
        <button :disabled="currentPage === 1" @click="currentPage--" class="page-btn">
          <i>◀️</i> 上一页
        </button>
        <span class="page-info">第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
        <button :disabled="currentPage === totalPages" @click="currentPage++" class="page-btn">
          下一页 <i>▶️</i>
        </button>
      </div>
    </div>

    <!-- 策略详情弹窗 -->
    <div v-if="showDetails" class="modal-overlay" @click="closeDetails">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>策略详情</h3>
          <button @click="closeDetails" class="close-btn">✕</button>
        </div>

        <div class="modal-body">
          <div class="detail-grid">
            <div class="detail-card">
              <div class="detail-label">策略ID</div>
              <div class="detail-value">{{ selectedStrategy.id }}</div>
            </div>
            <div class="detail-card">
              <div class="detail-label">交易对</div>
              <div class="detail-value">{{ selectedStrategy.symbol }}</div>
            </div>
            <div class="detail-card">
              <div class="detail-label">策略类型</div>
              <div class="detail-value">{{ getStrategyTypeText(selectedStrategy.strategyType) }}</div>
            </div>
            <div class="detail-card">
              <div class="detail-label">方向</div>
              <div class="detail-value">{{ selectedStrategy.side === 'BUY' ? '买入' : '卖出' }}</div>
            </div>
            <div class="detail-card">
              <div class="detail-label">基准价格</div>
              <div class="detail-value">{{ formatPrice(selectedStrategy.price) }}</div>
            </div>
            <div class="detail-card">
              <div class="detail-label">总数量</div>
              <div class="detail-value">{{ selectedStrategy.totalQuantity }}</div>
            </div>
          </div>

          <div v-if="selectedStrategy.buyQuantities && selectedStrategy.buyQuantities.length > 0" class="config-display">
            <h4>🟢 买入配置</h4>
            <div class="config-info">
              <p><strong>数量分配:</strong> {{ selectedStrategy.buyQuantities.join(', ') }}</p>
              <p v-if="selectedStrategy.strategyType !== 'custom'">
                <strong>深度级别:</strong> {{ selectedStrategy.buyDepthLevels ? selectedStrategy.buyDepthLevels.join(', ') : '' }}
              </p>
              <p v-if="selectedStrategy.strategyType === 'custom' && selectedStrategy.buyBasisPoints">
                <strong>万分比:</strong> {{ selectedStrategy.buyBasisPoints.map(bp => bp > 0 ? '+' + bp : bp).join(', ') }}bp
              </p>
            </div>
          </div>

          <div v-if="selectedStrategy.sellQuantities && selectedStrategy.sellQuantities.length > 0" class="config-display">
            <h4>🔴 卖出配置</h4>
            <div class="config-info">
              <p><strong>数量分配:</strong> {{ selectedStrategy.sellQuantities.join(', ') }}</p>
              <p v-if="selectedStrategy.strategyType !== 'custom'">
                <strong>深度级别:</strong> {{ selectedStrategy.sellDepthLevels ? selectedStrategy.sellDepthLevels.join(', ') : '' }}
              </p>
              <p v-if="selectedStrategy.strategyType === 'custom' && selectedStrategy.sellBasisPoints">
                <strong>万分比:</strong> {{ selectedStrategy.sellBasisPoints.map(bp => bp > 0 ? '+' + bp : bp).join(', ') }}bp
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 策略统计弹窗 -->
    <div v-if="showStats" class="modal-overlay" @click="closeStats">
      <div class="modal-content large" @click.stop>
        <div class="modal-header">
          <h3>策略统计 - {{ statsData.strategy?.symbol }}</h3>
          <button @click="closeStats" class="close-btn">✕</button>
        </div>

        <div class="modal-body">
          <div class="stats-overview">
            <div class="overview-card">
              <div class="overview-icon">📊</div>
              <div class="overview-content">
                <div class="overview-value">{{ statsData.stats?.totalOrders || 0 }}</div>
                <div class="overview-label">总订单数</div>
              </div>
            </div>
            <div class="overview-card">
              <div class="overview-icon">⏳</div>
              <div class="overview-content">
                <div class="overview-value pending">{{ statsData.stats?.pendingOrders || 0 }}</div>
                <div class="overview-label">待处理订单</div>
              </div>
            </div>
            <div class="overview-card">
              <div class="overview-icon">✅</div>
              <div class="overview-content">
                <div class="overview-value success">{{ statsData.stats?.filledOrders || 0 }}</div>
                <div class="overview-label">已成交订单</div>
              </div>
            </div>
            <div class="overview-card">
              <div class="overview-icon">❌</div>
              <div class="overview-content">
                <div class="overview-value cancelled">{{ statsData.stats?.cancelledOrders || 0 }}</div>
                <div class="overview-label">已取消订单</div>
              </div>
            </div>
            <div class="overview-card">
              <div class="overview-icon">💰</div>
              <div class="overview-content">
                <div class="overview-value">{{ formatVolume(statsData.stats?.totalVolume || 0) }}</div>
                <div class="overview-label">总交易额</div>
              </div>
            </div>
            <div class="overview-card">
              <div class="overview-icon">🎯</div>
              <div class="overview-content">
                <div class="overview-value success">{{ formatVolume(statsData.stats?.filledVolume || 0) }}</div>
                <div class="overview-label">已成交额</div>
              </div>
            </div>
          </div>

          <div class="recent-orders">
            <h4>最近订单</h4>
            <div v-if="statsData.recentOrders && statsData.recentOrders.length > 0" class="orders-table">
              <div class="table-header">
                <span>订单ID</span>
                <span>方向</span>
                <span>价格</span>
                <span>数量</span>
                <span>状态</span>
                <span>创建时间</span>
              </div>
              <div v-for="order in statsData.recentOrders" :key="order.id" class="table-row">
                <span>{{ order.orderId }}</span>
                <span :class="['side-badge', order.side.toLowerCase()]">
                  {{ order.side === 'BUY' ? '买入' : '卖出' }}
                </span>
                <span>{{ formatPrice(order.price) }}</span>
                <span>{{ formatQuantity(order.quantity) }}</span>
                <span :class="['status-badge', order.status]">
                  {{ getOrderStatusText(order.status) }}
                </span>
                <span>{{ formatDate(order.createdAt) }}</span>
              </div>
            </div>
            <div v-else class="no-orders">
              <div class="no-orders-icon">📄</div>
              <p>暂无订单记录</p>
            </div>
          </div>

          <div class="modal-actions">
            <button @click="viewAllStrategyOrders" class="action-btn view-all">
              <i>📋</i> 查看所有订单
            </button>
            <button @click="closeStats" class="action-btn secondary">
              <i>✕</i> 关闭
            </button>
          </div>
        </div>
      </div>
    </div>
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
      pageSize: 9,
      searchQuery: '',
      isCreatingStrategy: false,
      showDetails: false,
      showStats: false,
      showCreateForm: false,
      selectedStrategy: {},
      quantityWarning: '',
      orderPreview: [],
      availableSymbols: [],
      statsData: {
        stats: {},
        recentOrders: [],
        strategy: {}
      },
      toastMessage: '',
      toastType: 'success'
    };
  },
  computed: {
    filteredStrategies() {
      if (!this.searchQuery) return this.strategies;

      return this.strategies.filter(strategy =>
          strategy.symbol.toLowerCase().includes(this.searchQuery.toLowerCase()) ||
          this.getStrategyTypeText(strategy.strategyType).includes(this.searchQuery)
      );
    },
    paginatedStrategies() {
      const start = (this.currentPage - 1) * this.pageSize;
      const end = start + this.pageSize;
      return this.filteredStrategies.slice(start, end);
    },
    totalPages() {
      return Math.ceil(this.filteredStrategies.length / this.pageSize);
    },
    activeStrategiesCount() {
      return this.strategies.filter(s => s.enabled && s.status === 'active').length;
    },
    executingStrategiesCount() {
      return this.strategies.filter(s => s.pendingBatch).length;
    },
    completedStrategiesCount() {
      return this.strategies.filter(s => s.status === 'completed').length;
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

    showToast(message, type = 'success') {
      this.toastMessage = message;
      this.toastType = type;
      setTimeout(() => {
        this.toastMessage = '';
      }, 3000);
    },

    toggleCreateForm() {
      this.showCreateForm = !this.showCreateForm;
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
      const date = new Date(dateString);
      const now = new Date();
      const diff = now - date;
      const days = Math.floor(diff / (1000 * 60 * 60 * 24));

      if (days === 0) return '今天';
      if (days === 1) return '昨天';
      if (days < 7) return `${days}天前`;
      if (days < 30) return `${Math.floor(days / 7)}周前`;
      if (days < 365) return `${Math.floor(days / 30)}个月前`;

      return date.toLocaleDateString('zh-CN');
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
        this.showToast(error.response?.data?.error || '获取策略失败', 'error');
      }
    },

    async fetchSymbols() {
      try {
        const response = await axios.get('/symbols', {
          headers: this.getAuthHeaders(),
        });
        this.availableSymbols = response.data.symbols || [];

        if (this.availableSymbols.length === 0) {
          this.showToast('请先在仪表盘中添加交易对', 'error');
        }
      } catch (error) {
        console.error('获取交易对失败:', error);
        this.showToast(error.response?.data?.error || '获取交易对失败', 'error');
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
        this.showToast('请填写所有必需字段', 'error');
        return;
      }

      if (!this.availableSymbols.includes(this.newStrategy.symbol)) {
        this.showToast('请选择有效的交易对', 'error');
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

        this.showToast(response.data.message || '策略创建成功 🎉');
        this.resetForm();
        this.showCreateForm = false;
        this.fetchStrategies();
      } catch (error) {
        console.error('创建策略失败:', error);
        this.showToast(error.response?.data?.error || '创建策略失败', 'error');
      } finally {
        this.isCreatingStrategy = false;
      }
    },

    async toggleStrategy(strategy) {
      try {
        const response = await axios.post('/toggle_strategy', { id: strategy.id }, {
          headers: this.getAuthHeaders(),
        });

        this.showToast(response.data.message || '策略状态切换成功 ✅');
        this.fetchStrategies();
      } catch (error) {
        console.error('切换策略失败:', error);
        this.showToast(error.response?.data?.error || '切换策略失败', 'error');
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
        this.showToast(error.response?.data?.error || '获取策略统计失败', 'error');
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

        this.showToast(response.data.message || '策略删除成功 🗑️');
        this.fetchStrategies();
      } catch (error) {
        console.error('删除策略失败:', error);
        this.showToast(error.response?.data?.error || '删除策略失败', 'error');
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
/* 全局样式 */
.strategy-container {
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
  z-index: 1000;
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

/* 创建策略区域 */
.create-section {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 24px;
  padding: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
  margin-bottom: 3rem;
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

.toggle-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.toggle-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
}

.form-slide-enter-active, .form-slide-leave-active {
  transition: all 0.3s ease;
}

.form-slide-enter-from {
  opacity: 0;
  transform: translateY(-20px);
}

.form-slide-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}

.create-form {
  margin-top: 2rem;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  margin-bottom: 2rem;
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

.form-group input, .form-group select {
  padding: 0.8rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  color: #fff;
  font-size: 0.9rem;
  transition: all 0.3s ease;
}

.form-group input:focus, .form-group select:focus {
  outline: none;
  background: rgba(255, 255, 255, 0.08);
  border-color: #667eea;
}

.form-group input::placeholder {
  color: #666;
}

.form-group small {
  color: #999;
  font-size: 0.8rem;
}

/* 策略说明 */
.strategy-description {
  margin: 2rem 0;
}

.description-card {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1.5rem;
}

.description-icon {
  font-size: 2rem;
  flex-shrink: 0;
}

.description-content h4 {
  margin: 0 0 0.5rem 0;
  color: #667eea;
  font-size: 1.1rem;
}

.description-content p {
  margin: 0 0 0.5rem 0;
  color: #ccc;
  line-height: 1.5;
}

.description-content small {
  color: #999;
  font-size: 0.85rem;
}

/* 自定义配置 */
.custom-config {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 2rem;
  margin: 2rem 0;
}

.custom-config h3 {
  margin: 0 0 1.5rem 0;
  color: #fff;
  font-size: 1.2rem;
}

.config-section {
  margin-bottom: 2rem;
}

.config-section h4 {
  margin: 0 0 1rem 0;
  color: #ccc;
  font-size: 1rem;
}

.config-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

.warning-message {
  background: rgba(255, 193, 7, 0.1);
  border: 1px solid rgba(255, 193, 7, 0.3);
  color: #fbbf24;
  padding: 1rem;
  border-radius: 8px;
  margin-top: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

/* 订单预览 */
.order-preview {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 2rem;
  margin: 2rem 0;
}

.order-preview h3 {
  margin: 0 0 1.5rem 0;
  color: #fff;
  font-size: 1.2rem;
}

.preview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1rem;
}

.preview-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1rem;
  transition: all 0.3s ease;
}

.preview-card:hover {
  background: rgba(255, 255, 255, 0.08);
  transform: translateY(-2px);
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.order-number {
  font-weight: 600;
  color: #667eea;
}

.order-ratio {
  background: rgba(102, 126, 234, 0.2);
  color: #a78bfa;
  padding: 0.2rem 0.5rem;
  border-radius: 20px;
  font-size: 0.8rem;
}

.preview-details {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.detail-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-row .label {
  color: #999;
  font-size: 0.8rem;
}

.detail-row .value {
  color: #ccc;
  font-size: 0.8rem;
  font-weight: 500;
}

/* 表单操作 */
.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 2rem;
}

.create-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem 2rem;
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.create-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(34, 197, 94, 0.4);
}

.create-btn:disabled {
  background: #6c757d;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.reset-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem 2rem;
  background: rgba(108, 117, 125, 0.1);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
  border-radius: 12px;
  font-size: 1rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.reset-btn:hover {
  background: rgba(108, 117, 125, 0.2);
}

/* 策略列表区域 */
.strategies-section {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 24px;
  padding: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

/* 搜索框 */
.search-box {
  position: relative;
  width: 300px;
}

.search-icon {
  position: absolute;
  left: 1rem;
  top: 50%;
  transform: translateY(-50%);
  font-size: 1.2rem;
}

.search-input {
  width: 100%;
  padding: 0.8rem 1rem 0.8rem 3rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  color: #fff;
  font-size: 0.9rem;
  transition: all 0.3s ease;
}

.search-input:focus {
  outline: none;
  background: rgba(255, 255, 255, 0.08);
  border-color: #667eea;
}

.search-input::placeholder {
  color: #666;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.3;
}

.empty-text {
  color: #666;
  font-size: 1.1rem;
  margin-bottom: 2rem;
}

.empty-action {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem 2rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  border-radius: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.empty-action:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
}

/* 策略卡片网格 */
.strategies-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
  gap: 1.5rem;
  margin-top: 2rem;
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
  padding-bottom: 1rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.strategy-symbol {
  font-size: 1.3rem;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* 状态和标签样式 */
.status-chip, .side-chip, .enable-chip, .exec-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.4rem 0.8rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
}

.status-chip {
  position: relative;
  padding-left: 1.5rem;
}

.status-dot {
  position: absolute;
  left: 0.5rem;
  top: 50%;
  transform: translateY(-50%);
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-chip.active {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.status-chip.active .status-dot {
  background: #22c55e;
}

.status-chip.inactive {
  background: rgba(108, 117, 125, 0.2);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
}

.status-chip.inactive .status-dot {
  background: #94a3b8;
}

.status-chip.completed {
  background: rgba(59, 130, 246, 0.2);
  color: #3b82f6;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.status-chip.completed .status-dot {
  background: #3b82f6;
}

.status-chip.cancelled {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.status-chip.cancelled .status-dot {
  background: #ef4444;
}

.side-chip.buy {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.side-chip.sell {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.enable-chip.enabled {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.enable-chip.disabled {
  background: rgba(108, 117, 125, 0.2);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
}

.exec-chip.executing {
  background: rgba(255, 193, 7, 0.2);
  color: #fbbf24;
  border: 1px solid rgba(255, 193, 7, 0.3);
}

.exec-chip.idle {
  background: rgba(108, 117, 125, 0.2);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
}

/* 策略元信息 */
.strategy-meta {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.meta-label {
  color: #666;
  font-size: 0.8rem;
  font-weight: 500;
}

.meta-value {
  color: #ccc;
  font-size: 0.9rem;
  font-weight: 500;
}

/* 策略操作按钮 */
.strategy-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.action-btn {
  flex: 1;
  min-width: 100px;
  padding: 0.6rem 0.8rem;
  border: none;
  border-radius: 8px;
  font-size: 0.8rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.3rem;
}

.action-btn i {
  font-style: normal;
  font-size: 0.9rem;
}

.action-btn.enable {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.action-btn.enable:hover {
  background: rgba(34, 197, 94, 0.2);
  transform: translateY(-1px);
}

.action-btn.disable {
  background: rgba(255, 193, 7, 0.1);
  color: #fbbf24;
  border: 1px solid rgba(255, 193, 7, 0.3);
}

.action-btn.disable:hover {
  background: rgba(255, 193, 7, 0.2);
  transform: translateY(-1px);
}

.action-btn.view {
  background: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.action-btn.view:hover {
  background: rgba(59, 130, 246, 0.2);
  transform: translateY(-1px);
}

.action-btn.stats {
  background: rgba(139, 92, 246, 0.1);
  color: #8b5cf6;
  border: 1px solid rgba(139, 92, 246, 0.3);
}

.action-btn.stats:hover {
  background: rgba(139, 92, 246, 0.2);
  transform: translateY(-1px);
}

.action-btn.delete {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.action-btn.delete:hover {
  background: rgba(239, 68, 68, 0.2);
  transform: translateY(-1px);
}

/* 分页 */
.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-top: 2rem;
}

.page-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.8rem 1.2rem;
  background: rgba(255, 255, 255, 0.05);
  color: #ccc;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.page-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  color: #999;
  font-size: 0.9rem;
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
  max-width: 600px;
  width: 90%;
  max-height: 80vh;
  overflow-y: auto;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.5);
}

.modal-content.large {
  max-width: 900px;
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
  font-size: 1.5rem;
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

/* 详情网格 */
.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.detail-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1rem;
}

.detail-label {
  color: #999;
  font-size: 0.8rem;
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.detail-value {
  color: #fff;
  font-size: 1rem;
  font-weight: 600;
}

/* 配置显示 */
.config-display {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1.5rem;
  margin-bottom: 1rem;
}

.config-display h4 {
  margin: 0 0 1rem 0;
  color: #fff;
  font-size: 1.1rem;
}

.config-info p {
  margin: 0.5rem 0;
  color: #ccc;
  line-height: 1.5;
}

.config-info strong {
  color: #667eea;
}

/* 统计概览 */
.stats-overview {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 1rem;
  margin-bottom: 2rem;
}

.overview-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  padding: 1rem;
  text-align: center;
}

.overview-icon {
  font-size: 1.5rem;
  margin-bottom: 0.5rem;
}

.overview-value {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 0.3rem;
}

.overview-value.pending {
  color: #fbbf24;
}

.overview-value.success {
  color: #22c55e;
}

.overview-value.cancelled {
  color: #ef4444;
}

.overview-label {
  color: #999;
  font-size: 0.8rem;
}

/* 最近订单 */
.recent-orders {
  margin-top: 2rem;
}

.recent-orders h4 {
  margin: 0 0 1rem 0;
  color: #fff;
  font-size: 1.2rem;
}

.orders-table {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  overflow: hidden;
}

.table-header {
  display: grid;
  grid-template-columns: 1fr 0.8fr 1fr 1fr 0.8fr 1.2fr;
  gap: 1rem;
  padding: 1rem;
  background: rgba(255, 255, 255, 0.05);
  font-weight: 600;
  color: #ccc;
  font-size: 0.9rem;
}

.table-row {
  display: grid;
  grid-template-columns: 1fr 0.8fr 1fr 1fr 0.8fr 1.2fr;
  gap: 1rem;
  padding: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  color: #ccc;
  font-size: 0.9rem;
  transition: all 0.3s ease;
}

.table-row:hover {
  background: rgba(255, 255, 255, 0.05);
}

.side-badge {
  padding: 0.2rem 0.5rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
  text-align: center;
}

.side-badge.buy {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

.side-badge.sell {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.status-badge {
  padding: 0.2rem 0.5rem;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: 500;
  text-align: center;
}

.status-badge.pending {
  background: rgba(255, 193, 7, 0.2);
  color: #fbbf24;
}

.status-badge.filled {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

.status-badge.cancelled {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.status-badge.expired {
  background: rgba(108, 117, 125, 0.2);
  color: #94a3b8;
}

.status-badge.rejected {
  background: rgba(239, 68, 68, 0.2);
  color: #ef4444;
}

.no-orders {
  text-align: center;
  padding: 3rem;
  color: #666;
}

.no-orders-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.5;
}

/* 弹窗操作按钮 */
.modal-actions {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  margin-top: 2rem;
  padding-top: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.action-btn.view-all {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 0.8rem 1.5rem;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-btn.view-all:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
}

.action-btn.secondary {
  background: rgba(108, 117, 125, 0.1);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
  padding: 0.8rem 1.5rem;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.action-btn.secondary:hover {
  background: rgba(108, 117, 125, 0.2);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .strategy-container {
    padding: 1rem;
  }

  .page-title {
    font-size: 2rem;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .stat-card {
    padding: 1.5rem;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .config-grid {
    grid-template-columns: 1fr;
  }

  .preview-grid {
    grid-template-columns: 1fr;
  }

  .section-header {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }

  .search-box {
    width: 100%;
  }

  .strategies-grid {
    grid-template-columns: 1fr;
  }

  .strategy-meta {
    grid-template-columns: 1fr;
  }

  .strategy-actions {
    flex-direction: column;
  }

  .action-btn {
    width: 100%;
  }

  .modal-content {
    width: 95%;
    max-height: 90vh;
  }

  .modal-header {
    padding: 1.5rem 1.5rem 1rem 1.5rem;
  }

  .modal-body {
    padding: 1.5rem;
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }

  .stats-overview {
    grid-template-columns: repeat(2, 1fr);
  }

  .table-header,
  .table-row {
    grid-template-columns: 1fr;
    gap: 0.5rem;
  }

  .table-header span,
  .table-row span {
    padding: 0.5rem 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }

  .modal-actions {
    flex-direction: column;
  }

  .toast {
    left: 1rem;
    right: 1rem;
    bottom: 1rem;
  }
}
</style>
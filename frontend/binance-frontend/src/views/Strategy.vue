<template>
  <div class="strategy-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1 class="page-title">策略管理</h1>
      <p class="page-description">创建和管理您的交易策略</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">
          <span>📊</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">总策略数</div>
          <div class="stat-value">{{ strategies.length }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon active">
          <span>✅</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">活跃策略</div>
          <div class="stat-value">{{ activeStrategiesCount }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon executing">
          <span>⚡</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">执行中</div>
          <div class="stat-value">{{ executingStrategiesCount }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon completed">
          <span>🎯</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">已完成</div>
          <div class="stat-value">{{ completedStrategiesCount }}</div>
        </div>
      </div>
    </div>

    <!-- 创建策略 -->
    <div class="create-section">
      <div class="section-header">
        <h2 class="section-title">创建新策略</h2>
        <button @click="toggleCreateForm" class="toggle-btn">
          <span>{{ showCreateForm ? '收起' : '展开' }}</span>
          <span class="toggle-icon">{{ showCreateForm ? '▲' : '▼' }}</span>
        </button>
      </div>

      <transition name="collapse">
        <div v-if="showCreateForm" class="create-form-wrapper">
          <form @submit.prevent="createStrategy" class="strategy-form">
            <div class="form-grid">
              <div class="form-group">
                <label class="form-label">交易对</label>
                <select v-model="newStrategy.symbol" class="form-control" required>
                  <option value="">选择交易对</option>
                  <option v-for="symbol in availableSymbols" :key="symbol" :value="symbol">
                    {{ symbol }}
                  </option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">策略类型</label>
                <select v-model="newStrategy.strategyType" @change="onStrategyTypeChange" class="form-control" required>
                  <option value="">选择策略类型</option>
                  <option value="simple">简单策略</option>
                  <option value="iceberg">冰山策略</option>
                  <option value="custom">自定义策略</option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">交易方向</label>
                <select v-model="newStrategy.side" @change="onSideChange" class="form-control" required>
                  <option value="">选择方向</option>
                  <option value="BUY">买入</option>
                  <option value="SELL">卖出</option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">基准价格</label>
                <input
                    v-model.number="newStrategy.price"
                    type="number"
                    step="0.00000001"
                    placeholder="基准价格"
                    class="form-control"
                    required
                />
              </div>

              <div class="form-group">
                <label class="form-label">总数量</label>
                <input
                    v-model.number="newStrategy.totalQuantity"
                    type="number"
                    step="0.00000001"
                    placeholder="交易总数量"
                    class="form-control"
                    required
                />
              </div>

              <div class="form-group">
                <label class="form-label">订单取消时间（分钟）</label>
                <input
                    v-model.number="newStrategy.cancelAfterMinutes"
                    type="number"
                    min="1"
                    max="10080"
                    placeholder="默认120分钟"
                    class="form-control"
                    @blur="validateCancelTime"
                />
                <p class="form-hint">订单将在指定时间后自动取消（1-10080分钟）</p>
              </div>
            </div>

            <!-- 策略说明 -->
            <div v-if="newStrategy.strategyType" class="strategy-info">
              <div v-if="newStrategy.strategyType === 'simple'" class="info-card">
                <div class="info-icon">🎯</div>
                <div class="info-content">
                  <h4>简单策略</h4>
                  <p>当价格达到触发条件时，以基准价格一次性下单全部数量。适合快速建仓或平仓。</p>
                </div>
              </div>

              <div v-if="newStrategy.strategyType === 'iceberg'" class="info-card">
                <div class="info-icon">🧊</div>
                <div class="info-content">
                  <h4>冰山策略</h4>
                  <p>将订单分成多个小订单，按照预设的价格层级分批下单，避免大单对市场的冲击。</p>
                  <p class="info-detail">默认分层：买单 [0%, -1%, -3%, -5%, -7%]，卖单 [0%, +1%, +3%, +5%, +7%]</p>
                </div>
              </div>

              <div v-if="newStrategy.strategyType === 'custom'" class="info-card">
                <div class="info-icon">⚙️</div>
                <div class="info-content">
                  <h4>自定义策略</h4>
                  <p>根据您的需求自定义价格层级和数量分配，实现更灵活的交易策略。</p>
                </div>
              </div>
            </div>

            <!-- 自定义策略配置 -->
            <transition name="fade">
              <div v-if="newStrategy.strategyType === 'custom'" class="custom-config">
                <h3 class="config-title">自定义配置</h3>

                <div v-if="newStrategy.side === 'BUY'" class="config-section">
                  <h4 class="config-subtitle">
                    <span class="config-icon">📈</span>
                    买入配置
                  </h4>
                  <div class="config-grid">
                    <div class="form-group">
                      <label class="form-label">数量比例</label>
                      <input
                          v-model="buyQuantitiesInput"
                          placeholder="如: 0.3,0.3,0.2,0.2"
                          class="form-control"
                          @blur="validateQuantities('buy')"
                      />
                      <p class="form-hint">每档占总数量的比例，总和应为1</p>
                    </div>
                    <div class="form-group">
                      <label class="form-label">价格偏移（万分比）</label>
                      <input
                          v-model="buyBasisPointsInput"
                          placeholder="如: 0,-10,-20,-30"
                          class="form-control"
                      />
                      <p class="form-hint">负数表示低于基准价格</p>
                    </div>
                  </div>
                </div>

                <div v-if="newStrategy.side === 'SELL'" class="config-section">
                  <h4 class="config-subtitle">
                    <span class="config-icon">📉</span>
                    卖出配置
                  </h4>
                  <div class="config-grid">
                    <div class="form-group">
                      <label class="form-label">数量比例</label>
                      <input
                          v-model="sellQuantitiesInput"
                          placeholder="如: 0.3,0.3,0.2,0.2"
                          class="form-control"
                          @blur="validateQuantities('sell')"
                      />
                      <p class="form-hint">每档占总数量的比例，总和应为1</p>
                    </div>
                    <div class="form-group">
                      <label class="form-label">价格偏移（万分比）</label>
                      <input
                          v-model="sellBasisPointsInput"
                          placeholder="如: 0,10,20,30"
                          class="form-control"
                      />
                      <p class="form-hint">正数表示高于基准价格</p>
                    </div>
                  </div>
                </div>

                <transition name="fade">
                  <div v-if="quantityWarning" class="warning-message">
                    <span class="warning-icon">⚠️</span>
                    <span>{{ quantityWarning }}</span>
                  </div>
                </transition>
              </div>
            </transition>

            <!-- 订单预览 -->
            <transition name="fade">
              <div v-if="orderPreview.length > 0" class="order-preview">
                <h3 class="preview-title">订单预览</h3>
                <div class="preview-info">
                  <span class="info-icon">⏰</span>
                  <span>订单将在 {{ formatCancelTime(newStrategy.cancelAfterMinutes || 120) }} 后自动取消</span>
                </div>
                <div class="preview-grid">
                  <div v-for="(order, index) in orderPreview" :key="index" class="preview-card">
                    <div class="preview-header">
                      <span class="order-number">订单 {{ index + 1 }}</span>
                      <span class="order-percent">{{ (order.ratio * 100).toFixed(1) }}%</span>
                    </div>
                    <div class="preview-details">
                      <div class="preview-item">
                        <span class="label">数量</span>
                        <span class="value">{{ order.quantity.toFixed(8) }}</span>
                      </div>
                      <div v-if="newStrategy.strategyType === 'custom'" class="preview-item">
                        <span class="label">价格偏移</span>
                        <span class="value">{{ order.basisPoint > 0 ? '+' : '' }}{{ order.basisPoint }}bp</span>
                      </div>
                      <div class="preview-item">
                        <span class="label">预估价格</span>
                        <span class="value">{{ order.price.toFixed(8) }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </transition>

            <div class="form-actions">
              <button type="submit" :disabled="isCreatingStrategy || !isFormValid" class="btn btn-primary">
                <span v-if="!isCreatingStrategy">创建策略</span>
                <span v-else class="btn-loading">
                  <span class="spinner"></span>
                  创建中...
                </span>
              </button>
              <button type="button" @click="resetForm" class="btn btn-outline">
                重置表单
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
          <span class="search-icon">🔍</span>
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
        <button @click="showCreateForm = true" class="btn btn-primary">
          创建第一个策略
        </button>
      </div>

      <div v-else class="strategies-grid">
        <div v-for="strategy in paginatedStrategies" :key="strategy.id" class="strategy-card">
          <div class="strategy-header">
            <div class="strategy-title">
              <h3>{{ strategy.symbol }}</h3>
              <span :class="['type-badge', strategy.strategyType]">
                {{ getStrategyTypeText(strategy.strategyType) }}
              </span>
            </div>
            <span :class="['status-chip', strategy.status]">
              {{ getStatusText(strategy.status) }}
            </span>
          </div>

          <div class="strategy-meta">
            <div class="meta-item">
              <span class="meta-label">方向</span>
              <span :class="['side-badge', strategy.side.toLowerCase()]">
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
              <span class="meta-label">取消时间</span>
              <span class="meta-value">{{ formatCancelTime(strategy.cancelAfterMinutes || 120) }}</span>
            </div>
          </div>

          <div class="strategy-status">
            <div class="status-item">
              <span class="status-label">启用状态</span>
              <label class="toggle-switch">
                <input
                    type="checkbox"
                    :checked="strategy.enabled"
                    @change="toggleStrategy(strategy)"
                />
                <span class="toggle-slider"></span>
              </label>
            </div>
            <div class="status-item">
              <span class="status-label">执行状态</span>
              <span :class="['exec-badge', strategy.pendingBatch ? 'executing' : 'idle']">
                {{ strategy.pendingBatch ? '执行中' : '空闲' }}
              </span>
            </div>
          </div>

          <div class="strategy-time">
            <span class="time-icon">🕐</span>
            <span>创建于 {{ formatDate(strategy.createdAt) }}</span>
          </div>

          <div class="strategy-actions">
            <button @click="viewStrategyDetails(strategy)" class="btn btn-outline btn-sm">
              <span>👁️</span>
              详情
            </button>
            <button @click="viewStrategyStats(strategy)" class="btn btn-outline btn-sm">
              <span>📊</span>
              统计
            </button>
            <button @click="deleteStrategy(strategy.id)" class="btn btn-danger btn-sm">
              <span>🗑️</span>
              删除
            </button>
          </div>
        </div>
      </div>

      <!-- 分页 -->
      <div v-if="strategies.length > pageSize" class="pagination">
        <button :disabled="currentPage === 1" @click="currentPage--" class="page-btn">
          <span>←</span> 上一页
        </button>
        <span class="page-info">第 {{ currentPage }} 页 / 共 {{ totalPages }} 页</span>
        <button :disabled="currentPage === totalPages" @click="currentPage++" class="page-btn">
          下一页 <span>→</span>
        </button>
      </div>
    </div>

    <!-- 策略详情弹窗 -->
    <transition name="modal">
      <div v-if="showDetails" class="modal-overlay" @click="closeDetails">
        <div class="modal-content modal-lg" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">策略详情</h3>
            <button @click="closeDetails" class="modal-close">×</button>
          </div>

          <div class="modal-body">
            <div class="detail-section">
              <h4 class="detail-title">基本信息</h4>
              <div class="detail-grid">
                <div class="detail-item">
                  <label>策略ID</label>
                  <span>{{ selectedStrategy.id }}</span>
                </div>
                <div class="detail-item">
                  <label>交易对</label>
                  <span>{{ selectedStrategy.symbol }}</span>
                </div>
                <div class="detail-item">
                  <label>策略类型</label>
                  <span>{{ getStrategyTypeText(selectedStrategy.strategyType) }}</span>
                </div>
                <div class="detail-item">
                  <label>方向</label>
                  <span>{{ selectedStrategy.side === 'BUY' ? '买入' : '卖出' }}</span>
                </div>
                <div class="detail-item">
                  <label>基准价格</label>
                  <span>{{ formatPrice(selectedStrategy.price) }}</span>
                </div>
                <div class="detail-item">
                  <label>总数量</label>
                  <span>{{ selectedStrategy.totalQuantity }}</span>
                </div>
                <div class="detail-item">
                  <label>订单取消时间</label>
                  <span>{{ formatCancelTime(selectedStrategy.cancelAfterMinutes || 120) }}</span>
                </div>
                <div class="detail-item">
                  <label>创建时间</label>
                  <span>{{ new Date(selectedStrategy.createdAt).toLocaleString('zh-CN') }}</span>
                </div>
              </div>
            </div>

            <div v-if="selectedStrategy.buyQuantities && selectedStrategy.buyQuantities.length > 0" class="detail-section">
              <h4 class="detail-title">买入配置</h4>
              <div class="config-display">
                <p><strong>数量分配：</strong>{{ selectedStrategy.buyQuantities.join(', ') }}</p>
                <p v-if="selectedStrategy.strategyType === 'custom' && selectedStrategy.buyBasisPoints">
                  <strong>价格偏移：</strong>{{ selectedStrategy.buyBasisPoints.map(bp => bp > 0 ? '+' + bp : bp).join(', ') }}bp
                </p>
              </div>
            </div>

            <div v-if="selectedStrategy.sellQuantities && selectedStrategy.sellQuantities.length > 0" class="detail-section">
              <h4 class="detail-title">卖出配置</h4>
              <div class="config-display">
                <p><strong>数量分配：</strong>{{ selectedStrategy.sellQuantities.join(', ') }}</p>
                <p v-if="selectedStrategy.strategyType === 'custom' && selectedStrategy.sellBasisPoints">
                  <strong>价格偏移：</strong>{{ selectedStrategy.sellBasisPoints.map(bp => bp > 0 ? '+' + bp : bp).join(', ') }}bp
                </p>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeDetails" class="btn btn-primary">关闭</button>
          </div>
        </div>
      </div>
    </transition>

    <!-- 策略统计弹窗 -->
    <transition name="modal">
      <div v-if="showStats" class="modal-overlay" @click="closeStats">
        <div class="modal-content modal-lg" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">策略统计 - {{ statsData.strategy?.symbol }}</h3>
            <button @click="closeStats" class="modal-close">×</button>
          </div>

          <div class="modal-body">
            <div class="stats-overview">
              <div class="overview-card">
                <div class="overview-icon">📊</div>
                <div class="overview-content">
                  <div class="overview-label">总订单数</div>
                  <div class="overview-value">{{ statsData.stats?.totalOrders || 0 }}</div>
                </div>
              </div>
              <div class="overview-card">
                <div class="overview-icon pending">⏳</div>
                <div class="overview-content">
                  <div class="overview-label">待处理</div>
                  <div class="overview-value">{{ statsData.stats?.pendingOrders || 0 }}</div>
                </div>
              </div>
              <div class="overview-card">
                <div class="overview-icon success">✅</div>
                <div class="overview-content">
                  <div class="overview-label">已成交</div>
                  <div class="overview-value">{{ statsData.stats?.filledOrders || 0 }}</div>
                </div>
              </div>
              <div class="overview-card">
                <div class="overview-icon cancelled">❌</div>
                <div class="overview-content">
                  <div class="overview-label">已取消</div>
                  <div class="overview-value">{{ statsData.stats?.cancelledOrders || 0 }}</div>
                </div>
              </div>
            </div>

            <div class="recent-orders">
              <h4 class="section-title">最近订单</h4>
              <div v-if="statsData.recentOrders && statsData.recentOrders.length > 0" class="orders-table">
                <table class="data-table">
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
                        <span :class="['side-badge', order.side.toLowerCase()]">
                          {{ order.side === 'BUY' ? '买入' : '卖出' }}
                        </span>
                    </td>
                    <td>{{ formatPrice(order.price) }}</td>
                    <td>{{ formatQuantity(order.quantity) }}</td>
                    <td>
                        <span :class="['status-badge', order.status]">
                          {{ getOrderStatusText(order.status) }}
                        </span>
                    </td>
                    <td>{{ formatDate(order.createdAt) }}</td>
                  </tr>
                  </tbody>
                </table>
              </div>
              <div v-else class="no-orders">
                <span class="no-orders-icon">📄</span>
                <p>暂无订单记录</p>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="viewAllStrategyOrders" class="btn btn-primary">
              查看所有订单
            </button>
            <button @click="closeStats" class="btn btn-outline">
              关闭
            </button>
          </div>
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
  name: 'Strategy',
  data() {
    return {
      strategies: [],
      newStrategy: {
        symbol: '',
        strategyType: '',
        side: '',
        price: 0,
        totalQuantity: 0,
        cancelAfterMinutes: 120
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

      const cancelTime = this.newStrategy.cancelAfterMinutes || 120;
      if (cancelTime < 1 || cancelTime > 10080) {
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

    formatCancelTime(minutes) {
      if (!minutes) minutes = 120;

      if (minutes < 60) {
        return `${minutes}分钟`;
      } else if (minutes < 1440) {
        const hours = Math.floor(minutes / 60);
        const mins = minutes % 60;
        return mins > 0 ? `${hours}小时${mins}分钟` : `${hours}小时`;
      } else {
        const days = Math.floor(minutes / 1440);
        const hours = Math.floor((minutes % 1440) / 60);
        if (hours > 0) {
          return `${days}天${hours}小时`;
        }
        return `${days}天`;
      }
    },

    validateCancelTime() {
      if (!this.newStrategy.cancelAfterMinutes) {
        this.newStrategy.cancelAfterMinutes = 120;
        return;
      }

      if (this.newStrategy.cancelAfterMinutes < 1) {
        this.newStrategy.cancelAfterMinutes = 1;
        this.showToast('取消时间最少为1分钟', 'error');
      } else if (this.newStrategy.cancelAfterMinutes > 10080) {
        this.newStrategy.cancelAfterMinutes = 10080;
        this.showToast('取消时间最多为7天（10080分钟）', 'error');
      }
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

        if (!strategyData.cancelAfterMinutes) {
          strategyData.cancelAfterMinutes = 120;
        }

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

        this.showToast(response.data.message || '策略创建成功');
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

        this.showToast(response.data.message || '策略状态切换成功');
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

        this.showToast(response.data.message || '策略删除成功');
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
        totalQuantity: 0,
        cancelAfterMinutes: 120
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
/* 页面容器 */
.strategy-container {
  max-width: 1400px;
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

.stat-icon.active {
  background: #d1fae5;
  color: #10b981;
}

.stat-icon.executing {
  background: #fef3c7;
  color: #f59e0b;
}

.stat-icon.completed {
  background: #dbeafe;
  color: #3b82f6;
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

/* 创建策略区域 */
.create-section {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  margin-bottom: 2rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid var(--color-border);
}

.section-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.toggle-btn {
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

.toggle-btn:hover {
  background: var(--color-bg-secondary);
}

.toggle-icon {
  font-size: 0.75rem;
}

.create-form-wrapper {
  padding: 1.5rem;
}

.strategy-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.form-control {
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

.form-hint {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

/* 策略说明 */
.strategy-info {
  margin-top: 1rem;
}

.info-card {
  display: flex;
  gap: 1rem;
  padding: 1rem;
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  margin-bottom: 1rem;
}

.info-icon {
  font-size: 2rem;
  flex-shrink: 0;
}

.info-content {
  flex: 1;
}

.info-content h4 {
  margin: 0 0 0.5rem 0;
  color: var(--color-text-primary);
  font-size: 1rem;
  font-weight: 600;
}

.info-content p {
  margin: 0 0 0.5rem 0;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  line-height: 1.5;
}

.info-detail {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

/* 自定义配置 */
.custom-config {
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 1.5rem;
}

.config-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 1rem 0;
}

.config-section {
  margin-bottom: 1.5rem;
}

.config-section:last-child {
  margin-bottom: 0;
}

.config-subtitle {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1rem;
  font-weight: 500;
  color: var(--color-text-primary);
  margin: 0 0 1rem 0;
}

.config-icon {
  font-size: 1.125rem;
}

.config-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.warning-message {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  background: #fef3c7;
  border: 1px solid #fbbf24;
  border-radius: var(--radius-md);
  color: #92400e;
  font-size: 0.875rem;
  margin-top: 1rem;
}

.warning-icon {
  font-size: 1rem;
}

/* 订单预览 */
.order-preview {
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 1.5rem;
}

.preview-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 1rem 0;
}

.preview-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  background: #dbeafe;
  border: 1px solid #3b82f6;
  border-radius: var(--radius-md);
  color: #1e40af;
  font-size: 0.875rem;
  margin-bottom: 1rem;
}

.preview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1rem;
}

.preview-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 1rem;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  padding-bottom: 0.75rem;
  border-bottom: 1px solid var(--color-border);
}

.order-number {
  font-weight: 600;
  color: var(--color-text-primary);
}

.order-percent {
  background: var(--color-primary);
  color: white;
  padding: 0.125rem 0.5rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.preview-details {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.preview-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.875rem;
}

.preview-item .label {
  color: var(--color-text-tertiary);
}

.preview-item .value {
  color: var(--color-text-primary);
  font-weight: 500;
}

/* 表单操作 */
.form-actions {
  display: flex;
  gap: 1rem;
  margin-top: 1.5rem;
}

/* 策略列表 */
.strategies-section {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
}

/* 搜索框 */
.search-box {
  position: relative;
}

.search-icon {
  position: absolute;
  left: 0.875rem;
  top: 50%;
  transform: translateY(-50%);
  font-size: 1rem;
  color: var(--color-text-tertiary);
}

.search-input {
  padding: 0.625rem 0.875rem 0.625rem 2.5rem;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-primary);
  font-size: 0.875rem;
  width: 240px;
  transition: all var(--transition-normal);
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

/* 策略网格 */
.strategies-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1rem;
  margin-top: 1.5rem;
}

.strategy-card {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 1.25rem;
  transition: all var(--transition-normal);
}

.strategy-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.strategy-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.strategy-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.strategy-title h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

/* 徽章样式 */
.type-badge {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
  background: var(--color-bg-tertiary);
  color: var(--color-text-secondary);
}

.type-badge.simple {
  background: #e0e7ff;
  color: #4338ca;
}

.type-badge.iceberg {
  background: #dbeafe;
  color: #1e40af;
}

.type-badge.custom {
  background: #f3e8ff;
  color: #6b21a8;
}

.status-chip {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-chip.active {
  background: #d1fae5;
  color: #065f46;
}

.status-chip.inactive {
  background: var(--color-bg-tertiary);
  color: var(--color-text-secondary);
}

.status-chip.completed {
  background: #dbeafe;
  color: #1e40af;
}

.status-chip.cancelled {
  background: #fee2e2;
  color: #991b1b;
}

.side-badge {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.side-badge.buy {
  background: #d1fae5;
  color: #065f46;
}

.side-badge.sell {
  background: #fee2e2;
  color: #991b1b;
}

/* 策略元信息 */
.strategy-meta {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.meta-label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.meta-value {
  font-size: 0.875rem;
  color: var(--color-text-primary);
  font-weight: 500;
}

/* 策略状态 */
.strategy-status {
  display: flex;
  gap: 1.5rem;
  padding: 0.75rem 0;
  border-top: 1px solid var(--color-border);
  border-bottom: 1px solid var(--color-border);
  margin-bottom: 0.75rem;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.status-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
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

.exec-badge {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.exec-badge.executing {
  background: #fef3c7;
  color: #92400e;
}

.exec-badge.idle {
  background: var(--color-bg-tertiary);
  color: var(--color-text-secondary);
}

/* 时间信息 */
.strategy-time {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
  margin-bottom: 0.75rem;
}

.time-icon {
  font-size: 0.875rem;
}

/* 策略操作 */
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

  .btn-loading {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  /* 分页 */
  .pagination {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 1rem;
    margin-top: 1.5rem;
  }

  .page-btn {
    padding: 0.625rem 1rem;
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

  .page-btn:hover:not(:disabled) {
    background: var(--color-bg-secondary);
  }

  .page-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .page-info {
    color: var(--color-text-secondary);
    font-size: 0.875rem;
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

  .empty-text {
    font-size: 1rem;
    margin-bottom: 1.5rem;
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
    max-width: 600px;
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

  /* 详情部分 */
  .detail-section {
    margin-bottom: 2rem;
  }

  .detail-section:last-child {
    margin-bottom: 0;
  }

  .detail-title {
    font-size: 1rem;
    font-weight: 600;
    color: var(--color-text-primary);
    margin: 0 0 1rem 0;
  }

  .detail-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }

  .detail-item {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .detail-item label {
    font-size: 0.75rem;
    color: var(--color-text-tertiary);
    font-weight: 500;
  }

  .detail-item span {
    font-size: 0.875rem;
    color: var(--color-text-primary);
  }

  .config-display {
    background: var(--color-bg-secondary);
    border-radius: var(--radius-md);
    padding: 1rem;
  }

  .config-display p {
    margin: 0 0 0.5rem 0;
    font-size: 0.875rem;
    color: var(--color-text-secondary);
  }

  .config-display p:last-child {
    margin-bottom: 0;
  }

  .config-display strong {
    color: var(--color-text-primary);
  }

  /* 统计概览 */
  .stats-overview {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 1rem;
    margin-bottom: 2rem;
  }

  .overview-card {
    background: var(--color-bg-secondary);
    border: 1px solid var(--color-border);
    border-radius: var(--radius-md);
    padding: 1rem;
    text-align: center;
  }

  .overview-icon {
    width: 40px;
    height: 40px;
    background: var(--color-bg-tertiary);
    border-radius: var(--radius-md);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.25rem;
    margin: 0 auto 0.5rem;
  }

  .overview-icon.pending {
    background: #fef3c7;
    color: #f59e0b;
  }

  .overview-icon.success {
    background: #d1fae5;
    color: #10b981;
  }

  .overview-icon.cancelled {
    background: #fee2e2;
    color: #ef4444;
  }

  .overview-content {
    text-align: center;
  }

  .overview-label {
    font-size: 0.75rem;
    color: var(--color-text-tertiary);
    margin-bottom: 0.25rem;
  }

  .overview-value {
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--color-text-primary);
  }

  /* 最近订单 */
  .recent-orders {
    margin-top: 2rem;
  }

  /* 数据表格 */
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
    border-bottom: 1px solid var(--color-border);
  }

  .data-table td {
    padding: 0.75rem;
    border-bottom: 1px solid var(--color-border);
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
    background: #fef3c7;
    color: #92400e;
  }

  .status-badge.filled {
    background: #d1fae5;
    color: #065f46;
  }

  .status-badge.cancelled,
  .status-badge.expired,
  .status-badge.rejected {
    background: #fee2e2;
    color: #991b1b;
  }

  .no-orders {
    text-align: center;
    padding: 3rem 2rem;
    color: var(--color-text-tertiary);
  }

  .no-orders-icon {
    font-size: 3rem;
    display: block;
    margin-bottom: 1rem;
    opacity: 0.5;
  }

  .no-orders p {
    margin: 0;
  }

  /* 加载动画 */
  .spinner {
    width: 16px;
    height: 16px;
    border: 2px solid rgba(255, 255, 255, 0.3);
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.6s linear infinite;
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
  .collapse-enter-active,
  .collapse-leave-active {
    transition: all 0.3s ease;
  }

  .collapse-enter-from {
    opacity: 0;
    transform: translateY(-10px);
  }

  .collapse-leave-to {
    opacity: 0;
    transform: translateY(-10px);
  }

  .fade-enter-active,
  .fade-leave-active {
    transition: all 0.3s ease;
  }

  .fade-enter-from {
    opacity: 0;
    transform: translateY(-10px);
  }

  .fade-leave-to {
    opacity: 0;
  }

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

    .form-grid {
      grid-template-columns: 1fr;
    }

    .config-grid {
      grid-template-columns: 1fr;
    }

    .preview-grid {
      grid-template-columns: 1fr;
    }

    .strategies-grid {
      grid-template-columns: 1fr;
    }

    .strategy-meta {
      grid-template-columns: 1fr;
    }

    .strategy-status {
      flex-direction: column;
      gap: 0.75rem;
    }

    .strategy-actions {
      flex-wrap: wrap;
    }

    .search-input {
      width: 100%;
    }

    .section-header {
      flex-direction: column;
      gap: 1rem;
      align-items: stretch;
    }

    .modal-content {
      width: 95%;
    }

    .detail-grid {
      grid-template-columns: 1fr;
    }

    .stats-overview {
      grid-template-columns: 1fr 1fr;
    }

    .data-table {
      font-size: 0.75rem;
    }

    .data-table th,
    .data-table td {
      padding: 0.5rem;
    }
  }
</style>
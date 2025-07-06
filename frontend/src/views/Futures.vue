<template>
  <div class="futures-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1 class="page-title">永续期货策略</h1>
      <p class="page-description">管理您的永续期货交易策略</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">
          <span>📊</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">活跃策略</div>
          <div class="stat-value">{{ stats.activeStrategies }}</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon success">
          <span>💰</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">总盈亏</div>
          <div class="stat-value" :class="stats.totalPnl >= 0 ? 'profit' : 'loss'">
            {{ formatCurrency(stats.totalPnl) }}
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">
          <span>📈</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">胜率</div>
          <div class="stat-value">{{ stats.winRate.toFixed(2) }}%</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon pending">
          <span>🎯</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">活跃持仓</div>
          <div class="stat-value">{{ stats.activePositions }}</div>
        </div>
      </div>
    </div>

    <!-- 策略列表 -->
    <div class="strategies-section">
      <div class="section-header">
        <h2 class="section-title">策略列表</h2>
        <button @click="showCreateModal = true" class="btn btn-primary">
          <span>➕</span>
          创建策略
        </button>
      </div>

      <div v-if="strategies.length === 0" class="empty-state">
        <div class="empty-icon">🎯</div>
        <p class="empty-text">暂无期货策略</p>
        <button @click="showCreateModal = true" class="btn btn-primary">
          创建第一个策略
        </button>
      </div>

      <div v-else class="strategies-list">
        <div v-for="strategy in strategies" :key="strategy.id" class="strategy-card">
          <!-- 策略头部 -->
          <div class="strategy-header">
            <div class="strategy-info">
              <h3>{{ strategy.strategyName }}</h3>
              <div class="strategy-badges">
                <span :class="['side-badge', strategy.side.toLowerCase()]">
                  {{ strategy.side === 'LONG' ? '做多' : '做空' }}
                </span>
                <span class="leverage-badge">
                  {{ strategy.leverage }}X
                </span>
                <span :class="['status-badge', getStatusClass(strategy.status)]">
                  {{ getStatusText(strategy.status) }}
                </span>
              </div>
            </div>
            <div class="strategy-toggle">
              <label class="switch">
                <input
                    type="checkbox"
                    :checked="strategy.enabled"
                    @change="toggleStrategy(strategy)"
                    :disabled="strategy.status !== 'waiting' && strategy.status !== 'cancelled'"
                />
                <span class="slider"></span>
              </label>
            </div>
          </div>

          <!-- 策略详情 -->
          <div class="strategy-details">
            <div class="detail-grid">
              <div class="detail-item">
                <span class="detail-label">交易对</span>
                <span class="detail-value">{{ strategy.symbol }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">触发价格</span>
                <span class="detail-value highlight">{{ formatPrice(strategy.basePrice) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">开仓价格</span>
                <span class="detail-value">{{ formatPrice(strategy.entryPrice) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">数量</span>
                <span class="detail-value">{{ formatQuantity(strategy.quantity) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">止盈价格</span>
                <span class="detail-value success">
                  {{ formatPrice(strategy.takeProfitPrice) }}
                  <span class="percentage">(+{{ strategy.takeProfitRate }}‰)</span>
                </span>
              </div>
              <div class="detail-item">
                <span class="detail-label">止损价格</span>
                <span class="detail-value danger">
                  {{ strategy.stopLossPrice > 0 ? formatPrice(strategy.stopLossPrice) : '未设置' }}
                  <span v-if="strategy.stopLossRate > 0" class="percentage">
                    (-{{ strategy.stopLossRate }}‰)
                  </span>
                </span>
              </div>
            </div>
          </div>

          <!-- 时间信息 -->
          <div class="strategy-time">
            <span class="time-icon">🕐</span>
            <span>创建于 {{ formatDate(strategy.createdAt) }}</span>
            <span v-if="strategy.triggeredAt" class="time-separator">•</span>
            <span v-if="strategy.triggeredAt">触发于 {{ formatDate(strategy.triggeredAt) }}</span>
          </div>

          <!-- 操作按钮 -->
          <div class="strategy-actions">
            <button
                v-if="strategy.status === 'waiting' || strategy.status === 'cancelled'"
                @click="editStrategy(strategy)"
                class="btn btn-outline btn-sm"
            >
              <span>✏️</span>
              编辑
            </button>
            <button
                @click="viewOrders(strategy)"
                class="btn btn-outline btn-sm"
            >
              <span>📋</span>
              订单
            </button>
            <button
                @click="viewPositions(strategy)"
                class="btn btn-outline btn-sm"
            >
              <span>📊</span>
              持仓
            </button>
            <button
                @click="deleteStrategy(strategy)"
                class="btn btn-outline btn-sm danger"
            >
              <span>🗑️</span>
              删除
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 持仓列表 -->
    <div class="positions-section">
      <div class="section-header">
        <h2 class="section-title">当前持仓</h2>
        <button @click="fetchPositions" class="refresh-btn">
          <span>🔄</span>
          刷新
        </button>
      </div>

      <div v-if="positions.length === 0" class="empty-state">
        <div class="empty-icon">📊</div>
        <p class="empty-text">暂无活跃持仓</p>
      </div>

      <div v-else class="positions-list">
        <div v-for="position in positions" :key="position.id" class="position-card">
          <div class="position-header">
            <div class="position-info">
              <h3>{{ position.symbol }}</h3>
              <span :class="['side-badge', position.positionSide.toLowerCase()]">
                {{ position.positionSide === 'LONG' ? '多头' : '空头' }}
              </span>
              <span class="leverage-badge">{{ position.leverage }}X</span>
            </div>
            <span :class="['pnl-value', position.unrealizedPnl >= 0 ? 'profit' : 'loss']">
              {{ position.unrealizedPnl >= 0 ? '+' : '' }}{{ formatCurrency(position.unrealizedPnl) }}
            </span>
          </div>

          <div class="position-details">
            <div class="detail-grid">
              <div class="detail-item">
                <span class="detail-label">开仓价格</span>
                <span class="detail-value">{{ formatPrice(position.entryPrice) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">标记价格</span>
                <span class="detail-value highlight">{{ formatPrice(position.markPrice) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">数量</span>
                <span class="detail-value">{{ formatQuantity(position.quantity) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">保证金</span>
                <span class="detail-value">{{ formatCurrency(position.isolatedMargin) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">强平价格</span>
                <span class="detail-value danger">{{ formatPrice(position.liquidationPrice) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">开仓时间</span>
                <span class="detail-value">{{ formatDate(position.openedAt) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 创建/编辑策略弹窗 -->
    <transition name="modal">
      <div v-if="showCreateModal" class="modal-overlay" @click="closeCreateModal">
        <div class="modal-content" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">{{ editingStrategy ? '编辑策略' : '创建策略' }}</h3>
            <button @click="closeCreateModal" class="modal-close">×</button>
          </div>

          <form @submit.prevent="submitStrategy" class="modal-body">
            <div class="form-grid">
              <div class="form-group full-width">
                <label class="form-label">策略名称</label>
                <input
                    v-model="strategyForm.strategyName"
                    type="text"
                    placeholder="输入策略名称"
                    class="form-control"
                    required
                />
              </div>

              <div class="form-group">
                <label class="form-label">交易对</label>
                <select v-model="strategyForm.symbol" class="form-control" :disabled="editingStrategy" required>
                  <option value="">选择交易对</option>
                  <option v-for="symbol in availableSymbols" :key="symbol" :value="symbol">
                    {{ symbol }}
                  </option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">方向</label>
                <select v-model="strategyForm.side" class="form-control" :disabled="editingStrategy" required>
                  <option value="">选择方向</option>
                  <option value="LONG">做多</option>
                  <option value="SHORT">做空</option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">触发价格</label>
                <input
                    v-model.number="strategyForm.basePrice"
                    type="number"
                    step="0.00000001"
                    placeholder="价格达到此值时触发"
                    class="form-control"
                    @input="calculateEntryPrice"
                    required
                />
              </div>

              <div class="form-group">
                <label class="form-label">
                  开仓价格浮动 (‰)
                  <span class="form-hint">
                    {{ strategyForm.side === 'LONG' ? '向下浮动' : '向上浮动' }}
                  </span>
                </label>
                <input
                    v-model.number="strategyForm.entryPriceFloat"
                    type="number"
                    step="0.1"
                    min="0"
                    placeholder="千分之几"
                    class="form-control"
                    @input="calculateEntryPrice"
                    required
                />
                <div class="calculated-price" v-if="strategyForm.entryPrice > 0">
                  计算后价格: {{ formatPrice(strategyForm.entryPrice) }}
                </div>
              </div>

              <div class="form-group">
                <label class="form-label">杠杆倍数</label>
                <select
                    v-model.number="strategyForm.leverage"
                    class="form-control leverage-select"
                    :class="getLeverageClass(strategyForm.leverage)"
                    required
                >
                  <option value="">选择杠杆</option>
                  <option v-for="i in 20" :key="i" :value="i">{{ i }}X</option>
                </select>
              </div>

              <div class="form-group">
                <label class="form-label">开仓数量 (USDT)</label>
                <input
                    v-model.number="strategyForm.quantity"
                    type="number"
                    step="0.001"
                    placeholder="投入的USDT数量"
                    class="form-control"
                    required
                />
                <span class="form-hint">请输入USDT数量</span>
              </div>

              <div class="form-group">
                <label class="form-label">止盈千分比 (‰)</label>
                <input
                    v-model.number="strategyForm.takeProfitRate"
                    type="number"
                    step="0.1"
                    min="0.1"
                    placeholder="扣除手续费后的净利润千分比"
                    class="form-control"
                    required
                />
              </div>

              <div class="form-group">
                <label class="form-label">止损千分比 (‰) <span class="optional">可选</span></label>
                <input
                    v-model.number="strategyForm.stopLossRate"
                    type="number"
                    step="0.1"
                    min="0"
                    placeholder="0 表示不设置止损"
                    class="form-control"
                />
              </div>

              <div class="form-group">
                <label class="form-label">保证金模式</label>
                <select v-model="strategyForm.marginType" class="form-control">
                  <option value="CROSSED">全仓</option>
                  <option value="ISOLATED">逐仓</option>
                </select>
              </div>
            </div>

            <!-- 策略预览 -->
            <div v-if="strategyForm.entryPrice > 0 && strategyForm.quantity > 0" class="strategy-preview">
              <h4 class="preview-title">策略预览</h4>
              <div class="preview-grid">
                <div class="preview-item">
                  <span class="preview-label">开仓价值</span>
                  <span class="preview-value">
                    {{ formatCurrency(strategyForm.quantity) }} USDT
                  </span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">合约数量</span>
                  <span class="preview-value">
                    {{ calculateContractQuantity() }} {{ getContractUnit() }}
                  </span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">所需保证金</span>
                  <span class="preview-value">
                    {{ formatCurrency(strategyForm.quantity / (strategyForm.leverage || 1)) }}
                  </span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">预计止盈价</span>
                  <span class="preview-value success">
                    {{ calculateTakeProfitPrice() }}
                  </span>
                </div>
                <div v-if="strategyForm.stopLossRate > 0" class="preview-item">
                  <span class="preview-label">预计止损价</span>
                  <span class="preview-value danger">
                    {{ calculateStopLossPrice() }}
                  </span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">开仓手续费</span>
                  <span class="preview-value">
                    {{ formatCurrency(calculateOpenFee()) }} (0.04%)
                  </span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">平仓手续费</span>
                  <span class="preview-value">
                    {{ formatCurrency(calculateCloseFee()) }} (0.04%)
                  </span>
                </div>
                <div class="preview-item full-width">
                  <span class="preview-label">总手续费</span>
                  <span class="preview-value">
                    {{ formatCurrency(calculateTotalFee()) }}
                  </span>
                </div>
              </div>
            </div>
          </form>

          <div class="modal-footer">
            <button @click="closeCreateModal" class="btn btn-outline">
              取消
            </button>
            <button
                @click="submitStrategy"
                :disabled="isSubmitting"
                class="btn btn-primary"
            >
              <span v-if="!isSubmitting">{{ editingStrategy ? '更新' : '创建' }}</span>
              <span v-else class="btn-loading">
                <span class="spinner"></span>
                {{ editingStrategy ? '更新中...' : '创建中...' }}
              </span>
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- 订单列表弹窗 -->
    <transition name="modal">
      <div v-if="showOrdersModal" class="modal-overlay" @click="closeOrdersModal">
        <div class="modal-content modal-large" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">策略订单 - {{ selectedStrategy?.strategyName }}</h3>
            <button @click="closeOrdersModal" class="modal-close">×</button>
          </div>

          <div class="modal-body">
            <div v-if="strategyOrders.length === 0" class="empty-state">
              <p>暂无订单</p>
            </div>
            <div v-else class="orders-table">
              <table>
                <thead>
                <tr>
                  <th>订单ID</th>
                  <th>类型</th>
                  <th>方向</th>
                  <th>价格</th>
                  <th>数量</th>
                  <th>状态</th>
                  <th>用途</th>
                  <th>时间</th>
                </tr>
                </thead>
                <tbody>
                <tr v-for="order in strategyOrders" :key="order.id">
                  <td>{{ order.orderId }}</td>
                  <td>{{ order.type }}</td>
                  <td>
                      <span :class="['side-badge', order.side.toLowerCase()]">
                        {{ order.side }}
                      </span>
                  </td>
                  <td>{{ formatPrice(order.price) }}</td>
                  <td>{{ formatQuantity(order.quantity) }}</td>
                  <td>
                      <span :class="['status-badge', order.status.toLowerCase()]">
                        {{ order.status }}
                      </span>
                  </td>
                  <td>{{ getOrderPurposeText(order.orderPurpose) }}</td>
                  <td>{{ formatDate(order.createdAt) }}</td>
                </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeOrdersModal" class="btn btn-primary">
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
  name: 'Futures',
  data() {
    return {
      strategies: [],
      positions: [],
      availableSymbols: [], // 可用交易对列表
      stats: {
        activeStrategies: 0,
        totalPnl: 0,
        winRate: 0,
        activePositions: 0,
        totalTrades: 0,
        winTrades: 0,
        lossTrades: 0,
        totalCommission: 0,
        netPnl: 0,
        averagePnl: 0,
        maxWin: 0,
        maxLoss: 0
      },
      showCreateModal: false,
      showOrdersModal: false,
      editingStrategy: null,
      selectedStrategy: null,
      strategyOrders: [],
      strategyForm: {
        strategyName: '',
        symbol: '',
        side: '',
        basePrice: 0,
        entryPrice: 0,
        entryPriceFloat: 0, // 新增：开仓价格浮动千分比
        leverage: 1,
        quantity: 0,
        takeProfitRate: 0,
        stopLossRate: 0,
        marginType: 'CROSSED' // 默认改为全仓
      },
      isSubmitting: false,
      toastMessage: '',
      toastType: 'success',
      refreshInterval: null
    };
  },

  mounted() {
    this.fetchSymbols();
    this.fetchStrategies();
    this.fetchPositions();
    this.fetchStats();

    // 定时刷新
    this.refreshInterval = setInterval(() => {
      this.fetchPositions();
      this.fetchStats();
    }, 30000);
  },

  beforeUnmount() {
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval);
    }
  },

  methods: {
    async fetchSymbols() {
      try {
        const response = await axios.get('/symbols');
        this.availableSymbols = response.data.symbols
            .filter(s => s.endsWith('USDT'))
            .sort();
      } catch (error) {
        console.error('获取交易对失败:', error);
      }
    },

    async fetchStrategies() {
      try {
        const response = await axios.get('/futures/strategies');
        this.strategies = response.data.strategies || [];
      } catch (error) {
        console.error('获取策略列表失败:', error);
        this.showToast('获取策略列表失败', 'error');
      }
    },

    async fetchPositions() {
      try {
        const response = await axios.get('/futures/positions?status=open');
        this.positions = response.data.positions || [];
      } catch (error) {
        console.error('获取持仓列表失败:', error);
      }
    },

    async fetchStats() {
      try {
        const response = await axios.get('/futures/stats');
        this.stats = response.data.stats || this.stats;
      } catch (error) {
        console.error('获取统计信息失败:', error);
      }
    },

    async toggleStrategy(strategy) {
      try {
        const response = await axios.put(`/futures/strategies/${strategy.id}`, {
          enabled: !strategy.enabled
        });

        this.showToast('策略状态更新成功');
        await this.fetchStrategies();
      } catch (error) {
        console.error('更新策略状态失败:', error);
        this.showToast(error.response?.data?.error || '更新失败', 'error');
      }
    },

    async submitStrategy() {
      if (this.isSubmitting) return;

      // 计算实际的合约数量
      const contractQuantity = this.calculateContractQuantity();

      const submitData = {
        ...this.strategyForm,
        quantity: parseFloat(contractQuantity), // 转换为合约数量
        takeProfitRate: this.strategyForm.takeProfitRate / 10, // 千分比转换为百分比
        stopLossRate: this.strategyForm.stopLossRate / 10 // 千分比转换为百分比
      };

      this.isSubmitting = true;
      try {
        if (this.editingStrategy) {
          // 更新策略
          await axios.put(`/futures/strategies/${this.editingStrategy.id}`, submitData);
          this.showToast('策略更新成功');
        } else {
          // 创建策略
          await axios.post('/futures/strategies', submitData);
          this.showToast('策略创建成功');
        }

        this.closeCreateModal();
        await this.fetchStrategies();
      } catch (error) {
        console.error('提交策略失败:', error);
        this.showToast(error.response?.data?.error || '提交失败', 'error');
      } finally {
        this.isSubmitting = false;
      }
    },

    async deleteStrategy(strategy) {
      if (!window.confirm(`确定要删除策略"${strategy.strategyName}"吗？`)) {
        return;
      }

      try {
        await axios.delete(`/futures/strategies/${strategy.id}`);
        this.showToast('策略删除成功');
        await this.fetchStrategies();
      } catch (error) {
        console.error('删除策略失败:', error);
        this.showToast(error.response?.data?.error || '删除失败', 'error');
      }
    },

    async viewOrders(strategy) {
      this.selectedStrategy = strategy;
      try {
        const response = await axios.get('/futures/orders', {
          params: { strategyId: strategy.id }
        });
        this.strategyOrders = response.data.orders || [];
        this.showOrdersModal = true;
      } catch (error) {
        console.error('获取订单失败:', error);
        this.showToast('获取订单失败', 'error');
      }
    },

    viewPositions(strategy) {
      // 滚动到持仓部分
      const positionsSection = document.querySelector('.positions-section');
      if (positionsSection) {
        positionsSection.scrollIntoView({ behavior: 'smooth' });
      }
    },

    editStrategy(strategy) {
      this.editingStrategy = strategy;
      this.strategyForm = {
        strategyName: strategy.strategyName,
        symbol: strategy.symbol,
        side: strategy.side,
        basePrice: strategy.basePrice,
        entryPrice: strategy.entryPrice,
        entryPriceFloat: 0, // 编辑时需要重新计算
        leverage: strategy.leverage,
        quantity: strategy.quantity,
        takeProfitRate: strategy.takeProfitRate * 10, // 百分比转换为千分比
        stopLossRate: (strategy.stopLossRate || 0) * 10, // 百分比转换为千分比
        marginType: strategy.marginType
      };
      this.showCreateModal = true;
    },

    closeCreateModal() {
      this.showCreateModal = false;
      this.editingStrategy = null;
      this.resetForm();
    },

    closeOrdersModal() {
      this.showOrdersModal = false;
      this.selectedStrategy = null;
      this.strategyOrders = [];
    },

    resetForm() {
      this.strategyForm = {
        strategyName: '',
        symbol: '',
        side: '',
        basePrice: 0,
        entryPrice: 0,
        entryPriceFloat: 0,
        leverage: 1,
        quantity: 0,
        takeProfitRate: 0,
        stopLossRate: 0,
        marginType: 'CROSSED'
      };
    },

    // 根据触发价格和浮动千分比计算开仓价格
    calculateEntryPrice() {
      const { basePrice, entryPriceFloat, side } = this.strategyForm;
      if (!basePrice || !entryPriceFloat) {
        this.strategyForm.entryPrice = basePrice || 0;
        return;
      }

      const floatRate = entryPriceFloat / 1000; // 千分比转小数
      if (side === 'LONG') {
        // 做多：向下浮动
        this.strategyForm.entryPrice = basePrice * (1 - floatRate);
      } else if (side === 'SHORT') {
        // 做空：向上浮动
        this.strategyForm.entryPrice = basePrice * (1 + floatRate);
      }
    },

    // 计算合约数量
    calculateContractQuantity() {
      const { quantity, entryPrice } = this.strategyForm;
      if (!quantity || !entryPrice) return '0';
      return (quantity / entryPrice).toFixed(8).replace(/\.?0+$/, '');
    },

    // 获取合约单位
    getContractUnit() {
      const { symbol } = this.strategyForm;
      if (!symbol) return '';
      return symbol.replace('USDT', '');
    },

    // 计算开仓手续费
    calculateOpenFee() {
      const { quantity } = this.strategyForm;
      return quantity * 0.0004; // 0.04%
    },

    // 计算平仓手续费
    calculateCloseFee() {
      const { quantity, takeProfitRate, side } = this.strategyForm;
      if (!quantity || !takeProfitRate) return 0;

      // 计算平仓价值
      const profitRate = takeProfitRate / 1000;
      let closeValue;
      if (side === 'LONG') {
        closeValue = quantity * (1 + profitRate);
      } else {
        closeValue = quantity * (1 - profitRate);
      }

      return closeValue * 0.0004; // 0.04%
    },

    // 计算总手续费
    calculateTotalFee() {
      return this.calculateOpenFee() + this.calculateCloseFee();
    },

    calculateTakeProfitPrice() {
      const { entryPrice, takeProfitRate, side } = this.strategyForm;
      if (!entryPrice || !takeProfitRate) return '-';

      const feeRate = 0.0004 * 2; // 开仓+平仓手续费
      const profitRate = takeProfitRate / 1000; // 千分比转小数

      if (side === 'LONG') {
        return this.formatPrice(entryPrice * (1 + profitRate + feeRate));
      } else {
        return this.formatPrice(entryPrice * (1 - profitRate - feeRate));
      }
    },

    calculateStopLossPrice() {
      const { entryPrice, stopLossRate, side } = this.strategyForm;
      if (!entryPrice || !stopLossRate) return '-';

      const lossRate = stopLossRate / 1000; // 千分比转小数

      if (side === 'LONG') {
        return this.formatPrice(entryPrice * (1 - lossRate));
      } else {
        return this.formatPrice(entryPrice * (1 + lossRate));
      }
    },

    // 获取杠杆样式类
    getLeverageClass(leverage) {
      if (leverage >= 1 && leverage <= 5) {
        return 'leverage-low';
      } else if (leverage >= 6 && leverage <= 20) {
        return 'leverage-high';
      }
      return '';
    },

    getStatusClass(status) {
      const statusMap = {
        'waiting': 'waiting',
        'triggered': 'triggered',
        'position_opened': 'active',
        'completed': 'completed',
        'cancelled': 'cancelled'
      };
      return statusMap[status] || status;
    },

    getStatusText(status) {
      const statusMap = {
        'waiting': '等待触发',
        'triggered': '已触发',
        'position_opened': '持仓中',
        'completed': '已完成',
        'cancelled': '已取消'
      };
      return statusMap[status] || status;
    },

    getOrderPurposeText(purpose) {
      const purposeMap = {
        'entry': '开仓',
        'take_profit': '止盈',
        'stop_loss': '止损'
      };
      return purposeMap[purpose] || purpose;
    },

    formatPrice(price) {
      return parseFloat(price).toFixed(8).replace(/\.?0+$/, '');
    },

    formatQuantity(quantity) {
      return parseFloat(quantity).toFixed(8).replace(/\.?0+$/, '');
    },

    formatCurrency(amount) {
      return new Intl.NumberFormat('zh-CN', {
        style: 'currency',
        currency: 'USD',
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }).format(amount || 0);
    },

    formatDate(dateString) {
      if (!dateString) return '-';
      const date = new Date(dateString);
      const now = new Date();
      const diff = now - date;
      const hours = Math.floor(diff / (1000 * 60 * 60));

      if (hours < 1) return '刚刚';
      if (hours < 24) return `${hours}小时前`;

      return date.toLocaleDateString('zh-CN', {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
      });
    },

    showToast(message, type = 'success') {
      this.toastMessage = message;
      this.toastType = type;
      setTimeout(() => {
        this.toastMessage = '';
      }, 3000);
    }
  }
};
</script>

<style scoped>
/* 复用原有样式，添加新样式 */
.futures-container {
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

.stat-icon.pending {
  background: #fef3c7;
  color: #f59e0b;
}

.stat-icon.success {
  background: #d1fae5;
  color: #10b981;
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

.stat-value.profit {
  color: var(--color-success);
}

.stat-value.loss {
  color: var(--color-danger);
}

/* 策略和持仓区域 */
.strategies-section,
.positions-section {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: 1.5rem;
  margin-bottom: 2rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.section-title {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

/* 策略卡片 */
.strategies-list,
.positions-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.strategy-card,
.position-card {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: 1.25rem;
  transition: all var(--transition-normal);
}

.strategy-card:hover,
.position-card:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
}

.strategy-header,
.position-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.strategy-info,
.position-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.strategy-info h3,
.position-info h3 {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.strategy-badges {
  display: flex;
  gap: 0.5rem;
}

/* 开关组件 */
.switch {
  position: relative;
  display: inline-block;
  width: 48px;
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
  background-color: #ccc;
  transition: 0.4s;
  border-radius: 34px;
}

.slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 4px;
  bottom: 4px;
  background-color: white;
  transition: 0.4s;
  border-radius: 50%;
}

input:checked + .slider {
  background-color: var(--color-success);
}

input:checked + .slider:before {
  transform: translateX(24px);
}

input:disabled + .slider {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 详情网格 */
.strategy-details,
.position-details {
  margin-bottom: 1rem;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
}

.detail-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.detail-label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.detail-value {
  font-size: 0.875rem;
  color: var(--color-text-primary);
  font-weight: 500;
}

.detail-value.highlight {
  color: var(--color-primary);
}

.detail-value.success {
  color: var(--color-success);
}

.detail-value.danger {
  color: var(--color-danger);
}

.percentage {
  font-size: 0.75rem;
  opacity: 0.8;
}

/* 时间信息 */
.strategy-time {
  margin-bottom: 0.75rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.time-icon {
  font-size: 0.875rem;
}

.time-separator {
  color: var(--color-border);
}

/* 操作按钮 */
.strategy-actions {
  display: flex;
  gap: 0.5rem;
}

/* 徽章样式 */
.side-badge {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.side-badge.long {
  background: #d1fae5;
  color: #065f46;
}

.side-badge.short {
  background: #fee2e2;
  color: #991b1b;
}

.side-badge.buy {
  background: #d1fae5;
  color: #065f46;
}

.side-badge.sell {
  background: #fee2e2;
  color: #991b1b;
}

.leverage-badge {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
  background: #e0e7ff;
  color: #3730a3;
}

.status-badge {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-badge.waiting {
  background: #f3f4f6;
  color: #4b5563;
}

.status-badge.triggered {
  background: #fef3c7;
  color: #92400e;
}

.status-badge.active {
  background: #dbeafe;
  color: #1e40af;
}

.status-badge.completed {
  background: #d1fae5;
  color: #065f46;
}

.status-badge.cancelled {
  background: #fee2e2;
  color: #991b1b;
}

.status-badge.new {
  background: #dbeafe;
  color: #1e40af;
}

.status-badge.filled {
  background: #d1fae5;
  color: #065f46;
}

.status-badge.canceled {
  background: #fee2e2;
  color: #991b1b;
}

/* 盈亏值 */
.pnl-value {
  font-size: 1.125rem;
  font-weight: 600;
}

.pnl-value.profit {
  color: var(--color-success);
}

.pnl-value.loss {
  color: var(--color-danger);
}

/* 表单样式 */
.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group.full-width {
  grid-column: 1 / -1;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-primary);
}

.form-label .optional {
  color: var(--color-text-tertiary);
  font-weight: 400;
  font-size: 0.75rem;
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

.form-control:disabled {
  background-color: var(--color-bg-tertiary);
  cursor: not-allowed;
}

/* 新增：表单提示 */
.form-hint {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
  margin-top: 0.25rem;
}

/* 新增：计算后价格显示 */
.calculated-price {
  font-size: 0.75rem;
  color: var(--color-primary);
  margin-top: 0.25rem;
  font-weight: 500;
}

/* 新增：杠杆选择样式 */
.leverage-select.leverage-low {
  color: var(--color-success);
}

.leverage-select.leverage-high {
  color: var(--color-danger);
}

/* 策略预览 */
.strategy-preview {
  margin-top: 1.5rem;
  padding: 1rem;
  background: var(--color-bg-secondary);
  border-radius: var(--radius-md);
}

.preview-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.75rem 0;
}

.preview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
}

.preview-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.preview-item.full-width {
  grid-column: 1 / -1;
}

.preview-label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.preview-value {
  font-size: 0.875rem;
  color: var(--color-text-primary);
  font-weight: 500;
}

.preview-value.success {
  color: var(--color-success);
}

.preview-value.danger {
  color: var(--color-danger);
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

.btn-outline {
  background-color: transparent;
  border-color: var(--color-border);
  color: var(--color-text-secondary);
}

.btn-outline:hover {
  background-color: var(--color-bg-tertiary);
  border-color: var(--color-text-tertiary);
}

.btn-outline.danger:hover {
  background-color: #fee2e2;
  border-color: var(--color-danger);
  color: var(--color-danger);
}

.btn-sm {
  padding: 0.375rem 0.75rem;
  font-size: 0.75rem;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-loading {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.refresh-btn {
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

.refresh-btn:hover {
  background: var(--color-bg-secondary);
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
  margin-bottom: 1rem;
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

.modal-content.modal-large {
  max-width: 900px;
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

/* 订单表格 */
.orders-table {
  overflow-x: auto;
}

.orders-table table {
  width: 100%;
  border-collapse: collapse;
}

.orders-table th {
  background-color: var(--color-bg-tertiary);
  font-weight: 600;
  text-align: left;
  padding: 0.75rem;
  color: var(--color-text-primary);
  font-size: 0.75rem;
  white-space: nowrap;
}

.orders-table td {
  padding: 0.75rem;
  border-top: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  font-size: 0.875rem;
}

.orders-table tr:hover td {
  background-color: var(--color-bg-secondary);
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

  .strategy-header,
  .position-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.5rem;
  }

  .detail-grid {
    grid-template-columns: 1fr 1fr;
  }

  .strategy-actions {
    flex-wrap: wrap;
  }

  .modal-content {
    width: 95%;
  }

  .orders-table {
    font-size: 0.75rem;
  }

  .orders-table th,
  .orders-table td {
    padding: 0.5rem;
  }
}
</style>
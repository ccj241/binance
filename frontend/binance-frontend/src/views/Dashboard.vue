<template>
  <div class="dashboard-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1 class="page-title">交易仪表盘</h1>
      <p class="page-description">实时监控您的交易数据和资产状况</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-label">总资产价值</span>
          <span class="stat-icon">💰</span>
        </div>
        <div class="stat-value">${{ formatCurrency(totalAssetValue) }}</div>
        <div class="stat-change positive">
          <span class="change-icon">↑</span>
          <span>+12.5%</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-label">今日盈亏</span>
          <span class="stat-icon">📈</span>
        </div>
        <div class="stat-value">${{ formatCurrency(todayPnL) }}</div>
        <div class="stat-change" :class="todayPnL >= 0 ? 'positive' : 'negative'">
          <span class="change-icon">{{ todayPnL >= 0 ? '↑' : '↓' }}</span>
          <span>{{ todayPnL >= 0 ? '+' : '' }}{{ ((todayPnL / totalAssetValue) * 100).toFixed(2) }}%</span>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-label">活跃交易</span>
          <span class="stat-icon">🔄</span>
        </div>
        <div class="stat-value">{{ activeTradesCount }}</div>
        <div class="stat-subtitle">{{ pendingOrdersCount }} 待处理订单</div>
      </div>

      <div class="stat-card">
        <div class="stat-header">
          <span class="stat-label">24h 交易量</span>
          <span class="stat-icon">⚡</span>
        </div>
        <div class="stat-value">{{ formatVolume(volume24h) }}</div>
        <div class="stat-subtitle">{{ tradesCount24h }} 笔交易</div>
      </div>
    </div>

    <!-- 价格监控 -->
    <div class="content-card">
      <div class="card-header">
        <h2 class="card-title">实时价格监控</h2>
        <button @click="openAddSymbolModal" class="btn btn-primary">
          <span class="btn-icon">+</span>
          添加交易对
        </button>
      </div>

      <div class="card-body">
        <div v-if="Object.keys(prices).length === 0" class="empty-state">
          <span class="empty-icon">📉</span>
          <p>还未添加任何交易对</p>
          <button @click="openAddSymbolModal" class="btn btn-primary">
            添加第一个交易对
          </button>
        </div>

        <div v-else class="price-grid">
          <div v-for="(price, symbol) in prices" :key="symbol" class="price-card">
            <div class="price-header">
              <h3 class="symbol-name">{{ symbol }}</h3>
              <button @click="confirmDeleteSymbol(symbol)" class="delete-btn" title="删除交易对">
                ×
              </button>
            </div>
            <div class="price-info">
              <div class="current-price">${{ formatPrice(price) }}</div>
              <div class="price-change" :class="getPriceChangeClass(symbol)">
                <span class="change-arrow">{{ getPriceChangeIcon(symbol) }}</span>
                <span>{{ Math.abs(getPriceChangePercent(symbol)) }}%</span>
              </div>
            </div>
            <div class="price-chart-placeholder"></div>
          </div>
        </div>
      </div>
    </div>

    <!-- 账户余额 -->
    <div class="content-card">
      <div class="card-header">
        <h2 class="card-title">账户余额</h2>
        <button @click="fetchBalances" class="btn btn-outline" :disabled="isLoadingBalances">
          <span class="btn-icon" :class="{ 'spinning': isLoadingBalances }">⟳</span>
          {{ isLoadingBalances ? '刷新中...' : '刷新' }}
        </button>
      </div>

      <div class="card-body">
        <div v-if="isLoadingBalances && balances.length === 0" class="loading-state">
          <div class="spinner"></div>
          <p>加载余额中...</p>
        </div>

        <div v-else-if="balances.length === 0" class="empty-state">
          <span class="empty-icon">💳</span>
          <p>暂无余额信息</p>
        </div>

        <div v-else class="balance-grid">
          <div v-for="balance in filteredBalances" :key="balance.asset" class="balance-card">
            <div class="balance-header">
              <div class="coin-info">
                <div class="coin-icon">{{ balance.asset.charAt(0) }}</div>
                <span class="coin-name">{{ balance.asset }}</span>
              </div>
              <div class="balance-value">
                ≈ ${{ formatCurrency(getBalanceValue(balance)) }}
              </div>
            </div>
            <div class="balance-details">
              <div class="balance-item">
                <span class="label">可用</span>
                <span class="value">{{ formatBalance(balance.free) }}</span>
              </div>
              <div class="balance-item">
                <span class="label">锁定</span>
                <span class="value">{{ formatBalance(balance.locked) }}</span>
              </div>
              <div class="balance-item">
                <span class="label">总计</span>
                <span class="value total">{{ formatBalance(parseFloat(balance.free) + parseFloat(balance.locked)) }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 最近交易 -->
    <div class="content-card">
      <div class="card-header">
        <h2 class="card-title">最近交易记录</h2>
        <select v-model="tradeFilter" class="filter-select">
          <option value="all">全部</option>
          <option value="buy">买入</option>
          <option value="sell">卖出</option>
        </select>
      </div>

      <div class="card-body">
        <div v-if="isLoadingTrades && trades.length === 0" class="loading-state">
          <div class="spinner"></div>
          <p>加载交易记录中...</p>
        </div>

        <div v-else-if="filteredTrades.length === 0" class="empty-state">
          <span class="empty-icon">📋</span>
          <p>暂无交易记录</p>
        </div>

        <div v-else class="table-container">
          <table class="data-table">
            <thead>
            <tr>
              <th>时间</th>
              <th>交易对</th>
              <th>方向</th>
              <th>价格</th>
              <th>数量</th>
              <th>总额</th>
              <th>状态</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="trade in paginatedTrades" :key="trade.id">
              <td>{{ formatTradeTime(trade.time) }}</td>
              <td class="symbol-cell">{{ trade.symbol }}</td>
              <td>
                <span :class="['trade-side', trade.side.toLowerCase()]">
                  {{ trade.side === 'BUY' ? '买入' : '卖出' }}
                </span>
              </td>
              <td>${{ formatPrice(trade.price) }}</td>
              <td>{{ formatQuantity(trade.qty) }}</td>
              <td class="amount-cell">${{ formatCurrency(trade.price * trade.qty) }}</td>
              <td>
                <span class="status-badge success">已完成</span>
              </td>
            </tr>
            </tbody>
          </table>

          <div v-if="filteredTrades.length > pageSize" class="pagination">
            <button :disabled="currentPage === 1" @click="currentPage--" class="page-btn">
              上一页
            </button>
            <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
            <button :disabled="currentPage === totalPages" @click="currentPage++" class="page-btn">
              下一页
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 添加交易对弹窗 -->
    <transition name="modal">
      <div v-if="showAddSymbolModal" class="modal-overlay" @click.self="closeAddSymbolModal">
        <div class="modal-content">
          <div class="modal-header">
            <h3 class="modal-title">添加交易对</h3>
            <button @click="closeAddSymbolModal" class="modal-close">×</button>
          </div>

          <div class="modal-body">
            <div class="form-group">
              <label class="form-label">交易对名称</label>
              <input
                  v-model="newSymbol"
                  @keyup.enter="addSymbol"
                  placeholder="输入交易对 (如 BTCUSDT)"
                  class="form-control"
                  ref="symbolInput"
                  :disabled="isAddingSymbol"
              />
              <p class="form-hint">请输入完整的交易对名称，如 BTCUSDT、ETHUSDT 等</p>
            </div>

            <div class="popular-symbols">
              <p class="popular-title">热门交易对</p>
              <div class="symbol-chips">
                <button
                    v-for="symbol in popularSymbols"
                    :key="symbol"
                    @click="selectPopularSymbol(symbol)"
                    class="symbol-chip"
                    :disabled="isAddingSymbol"
                >
                  {{ symbol }}
                </button>
              </div>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="closeAddSymbolModal" class="btn btn-outline" :disabled="isAddingSymbol">
              取消
            </button>
            <button @click="addSymbol" :disabled="!newSymbol.trim() || isAddingSymbol" class="btn btn-primary">
              {{ isAddingSymbol ? '添加中...' : '确认添加' }}
            </button>
          </div>
        </div>
      </div>
    </transition>

    <!-- 删除确认弹窗 -->
    <transition name="modal">
      <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="cancelDeleteSymbol">
        <div class="modal-content modal-sm">
          <div class="modal-header">
            <h3 class="modal-title">确认删除</h3>
            <button @click="cancelDeleteSymbol" class="modal-close">×</button>
          </div>

          <div class="modal-body">
            <div class="confirm-message">
              <span class="warning-icon">⚠️</span>
              <p>确定要删除交易对 <strong>{{ symbolToDelete }}</strong> 吗？</p>
              <p class="warning-text">删除后将停止价格监控</p>
            </div>
          </div>

          <div class="modal-footer">
            <button @click="cancelDeleteSymbol" class="btn btn-outline" :disabled="isDeletingSymbol">
              取消
            </button>
            <button @click="deleteSymbol" class="btn btn-danger" :disabled="isDeletingSymbol">
              {{ isDeletingSymbol ? '删除中...' : '确认删除' }}
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
export default {
  name: 'Dashboard',
  data() {
    return {
      // 价格相关
      prices: {},
      priceHistory: {},
      newSymbol: '',
      showAddSymbolModal: false,
      isAddingSymbol: false,
      popularSymbols: ['BTCUSDT', 'ETHUSDT', 'BNBUSDT', 'SOLUSDT', 'ADAUSDT'],

      // 余额相关
      balances: [],
      isLoadingBalances: false,

      // 交易相关
      trades: [],
      tradeFilter: 'all',
      currentPage: 1,
      pageSize: 10,
      isLoadingTrades: false,

      // 统计数据
      totalAssetValue: 50000,
      todayPnL: 1250.50,
      activeTradesCount: 5,
      pendingOrdersCount: 3,
      volume24h: 125000,
      tradesCount24h: 42,

      // UI 状态
      showDeleteConfirm: false,
      symbolToDelete: '',
      isDeletingSymbol: false,
      toastMessage: '',
      toastType: 'success',
      priceInterval: null,
    };
  },
  computed: {
    filteredBalances() {
      return this.balances.filter(b => (parseFloat(b.free) + parseFloat(b.locked)) > 0.00001);
    },

    filteredTrades() {
      if (this.tradeFilter === 'all') return this.trades;
      return this.trades.filter(t =>
          this.tradeFilter === 'buy' ? t.side === 'BUY' : t.side === 'SELL'
      );
    },

    paginatedTrades() {
      const start = (this.currentPage - 1) * this.pageSize;
      const end = start + this.pageSize;
      return this.filteredTrades.slice(start, end);
    },

    totalPages() {
      return Math.ceil(this.filteredTrades.length / this.pageSize);
    },
  },
  mounted() {
    console.log('Dashboard 组件已挂载');
    console.log('$axios 是否存在:', !!this.$axios);
    this.initDashboard();
  },
  beforeUnmount() {
    console.log('Dashboard 组件即将卸载');
    if (this.priceInterval) {
      clearInterval(this.priceInterval);
    }
  },
  methods: {
    async initDashboard() {
      try {
        console.log('开始初始化 Dashboard...');
        // 暂时注释掉 API 调用，先确保页面能显示
        // await Promise.all([
        //   this.fetchPrices(),
        //   this.fetchBalances(),
        //   this.fetchTrades(),
        // ]);

        // 启动价格更新定时器
        // this.priceInterval = setInterval(this.fetchPrices, 5000);
      } catch (error) {
        console.error('初始化仪表盘失败:', error);
        this.showToast('初始化仪表盘失败', 'error');
      }
    },

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

    formatCurrency(value) {
      return new Intl.NumberFormat('en-US', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
      }).format(value || 0);
    },

    formatPrice(price) {
      const numPrice = parseFloat(price);
      if (numPrice > 1000) return numPrice.toFixed(2);
      if (numPrice > 1) return numPrice.toFixed(4);
      return numPrice.toFixed(8);
    },

    formatQuantity(qty) {
      return parseFloat(qty).toFixed(8).replace(/\.?0+$/, '');
    },

    formatBalance(balance) {
      const numBalance = parseFloat(balance);
      if (numBalance === 0) return '0';
      if (numBalance < 0.00001) return '< 0.00001';
      return this.formatQuantity(numBalance);
    },

    formatVolume(volume) {
      if (volume >= 1000000) return (volume / 1000000).toFixed(2) + 'M';
      if (volume >= 1000) return (volume / 1000).toFixed(2) + 'K';
      return volume.toFixed(2);
    },

    formatTradeTime(timestamp) {
      const date = new Date(timestamp);
      return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      });
    },

    getBalanceValue(balance) {
      // 这里应该根据实时价格计算
      const mockPrices = {
        'BTC': 45000,
        'ETH': 3000,
        'BNB': 350,
        'USDT': 1,
      };
      const price = mockPrices[balance.asset] || 0;
      return (parseFloat(balance.free) + parseFloat(balance.locked)) * price;
    },

    getPriceChangeClass(symbol) {
      // 模拟价格变化
      return Math.random() > 0.5 ? 'positive' : 'negative';
    },

    getPriceChangeIcon(symbol) {
      const isPositive = this.getPriceChangeClass(symbol) === 'positive';
      return isPositive ? '↑' : '↓';
    },

    getPriceChangePercent(symbol) {
      // 模拟价格变化百分比
      return (Math.random() * 10 - 5).toFixed(2);
    },

    async fetchPrices() {
      try {
        const response = await this.$axios.get('/prices', {
          headers: this.getAuthHeaders(),
        });
        this.prices = response.data.prices || {};
      } catch (error) {
        console.error('获取价格失败:', error);
      }
    },

    async fetchBalances() {
      this.isLoadingBalances = true;
      try {
        const response = await this.$axios.get('/balance', {
          headers: this.getAuthHeaders(),
        });
        this.balances = response.data.balances || [];
      } catch (error) {
        console.error('获取余额失败:', error);
        this.showToast('获取余额失败', 'error');
      } finally {
        this.isLoadingBalances = false;
      }
    },

    async fetchTrades() {
      this.isLoadingTrades = true;
      try {
        const response = await this.$axios.get('/trades', {
          headers: this.getAuthHeaders(),
        });
        this.trades = response.data.trades || [];
        this.currentPage = 1;
      } catch (error) {
        console.error('获取交易记录失败:', error);
        this.showToast('获取交易记录失败', 'error');
      } finally {
        this.isLoadingTrades = false;
      }
    },

    openAddSymbolModal() {
      this.showAddSymbolModal = true;
      this.newSymbol = '';
      this.$nextTick(() => {
        if (this.$refs.symbolInput) {
          this.$refs.symbolInput.focus();
        }
      });
    },

    closeAddSymbolModal() {
      this.showAddSymbolModal = false;
      this.newSymbol = '';
      this.isAddingSymbol = false;
    },

    selectPopularSymbol(symbol) {
      this.newSymbol = symbol;
      this.addSymbol();
    },

    async addSymbol() {
      const symbol = this.newSymbol.trim().toUpperCase();

      if (!symbol) {
        this.showToast('请输入有效的交易对', 'error');
        return;
      }

      if (this.prices[symbol]) {
        this.showToast('该交易对已存在', 'error');
        return;
      }

      this.isAddingSymbol = true;
      try {
        const response = await this.$axios.post('/symbols',
            { symbol: symbol },
            { headers: this.getAuthHeaders() }
        );

        this.showToast('交易对添加成功');
        this.closeAddSymbolModal();
        await this.fetchPrices();
      } catch (error) {
        console.error('添加交易对失败:', error);
        this.showToast('添加交易对失败', 'error');
      } finally {
        this.isAddingSymbol = false;
      }
    },

    confirmDeleteSymbol(symbol) {
      this.symbolToDelete = symbol;
      this.showDeleteConfirm = true;
    },

    cancelDeleteSymbol() {
      this.showDeleteConfirm = false;
      this.symbolToDelete = '';
      this.isDeletingSymbol = false;
    },

    async deleteSymbol() {
      if (!this.symbolToDelete) return;

      this.isDeletingSymbol = true;
      try {
        const response = await this.$axios.delete('/symbols', {
          data: { symbol: this.symbolToDelete },
          headers: this.getAuthHeaders()
        });

        this.showToast('交易对删除成功');
        delete this.prices[this.symbolToDelete];
        this.cancelDeleteSymbol();
      } catch (error) {
        console.error('删除交易对失败:', error);
        this.showToast(error.response?.data?.error || '删除交易对失败', 'error');
        this.isDeletingSymbol = false;
      }
    },
  },
};
</script>

<style scoped>
/* 页面容器 */
.dashboard-container {
  width: 100%;
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
  color: #0f172a;
  margin: 0 0 0.5rem 0;
}

.page-description {
  color: #64748b;
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
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
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
  color: #64748b;
  font-weight: 500;
}

.stat-icon {
  font-size: 1.25rem;
  opacity: 0.7;
}

.stat-value {
  font-size: 1.75rem;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 0.5rem;
}

.stat-change {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.875rem;
  font-weight: 500;
}

.stat-change.positive {
  color: #10b981;
}

.stat-change.negative {
  color: #ef4444;
}

.change-icon {
  font-size: 0.75rem;
}

.stat-subtitle {
  font-size: 0.875rem;
  color: #94a3b8;
}

/* 内容卡片 */
.content-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  margin-bottom: 1.5rem;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.card-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #0f172a;
  margin: 0;
}

.card-body {
  padding: 1.5rem;
}

/* 价格网格 */
.price-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1rem;
}

.price-card {
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  padding: 1.25rem;
  transition: all 200ms ease;
}

.price-card:hover {
  background-color: #f8fafc;
}

.price-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.symbol-name {
  font-size: 1rem;
  font-weight: 600;
  color: #0f172a;
  margin: 0;
}

.delete-btn {
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: transparent;
  border: 1px solid #e2e8f0;
  border-radius: 0.25rem;
  color: #94a3b8;
  font-size: 1.25rem;
  cursor: pointer;
  transition: all 150ms ease;
}

.delete-btn:hover {
  color: #ef4444;
  border-color: #ef4444;
  background-color: rgba(239, 68, 68, 0.05);
}

.price-info {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 1rem;
}

.current-price {
  font-size: 1.5rem;
  font-weight: 600;
  color: #0f172a;
}

.price-change {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.875rem;
  font-weight: 500;
}

.price-change.positive {
  color: #10b981;
}

.price-change.negative {
  color: #ef4444;
}

.change-arrow {
  font-size: 0.75rem;
}

.price-chart-placeholder {
  height: 60px;
  background: #f8fafc;
  border-radius: 0.25rem;
}

/* 余额网格 */
.balance-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1rem;
}

.balance-card {
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  padding: 1.25rem;
  transition: all 200ms ease;
}

.balance-card:hover {
  background-color: #f8fafc;
}

.balance-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.coin-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.coin-icon {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #2563eb;
  color: white;
  border-radius: 50%;
  font-weight: 600;
  font-size: 0.875rem;
}

.coin-name {
  font-weight: 600;
  color: #0f172a;
}

.balance-value {
  font-size: 0.875rem;
  color: #64748b;
  font-weight: 500;
}

.balance-details {
  display: flex;
  gap: 1.5rem;
}

.balance-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.balance-item .label {
  font-size: 0.75rem;
  color: #94a3b8;
}

.balance-item .value {
  font-size: 0.875rem;
  color: #64748b;
  font-weight: 500;
}

.balance-item .value.total {
  color: #0f172a;
  font-weight: 600;
}

/* 按钮样式 */
.btn {
  padding: 0.5rem 1rem;
  border: 1px solid transparent;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 150ms ease;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-primary {
  background-color: #2563eb;
  color: white;
}

.btn-primary:hover {
  background-color: #1d4ed8;
}

.btn-primary:disabled {
  background-color: #64748b;
  cursor: not-allowed;
}

.btn-outline {
  background-color: transparent;
  border-color: #e2e8f0;
  color: #64748b;
}

.btn-outline:hover {
  background-color: #f1f5f9;
  border-color: #94a3b8;
}

.btn-outline:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-danger {
  background-color: #ef4444;
  color: white;
}

.btn-danger:hover {
  background-color: #dc2626;
}

.btn-danger:disabled {
  background-color: #64748b;
  cursor: not-allowed;
}

.btn-icon {
  font-size: 1rem;
}

.spinning {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* 筛选下拉框 */
.filter-select {
  padding: 0.5rem 1rem;
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  color: #0f172a;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 200ms ease;
}

.filter-select:focus {
  outline: none;
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

/* 表格容器 */
.table-container {
  overflow-x: auto;
}

/* 数据表格 */
.data-table {
  width: 100%;
  border-collapse: collapse;
  background-color: #ffffff;
}

.data-table th {
  text-align: left;
  padding: 0.75rem 1rem;
  background-color: #f8fafc;
  color: #64748b;
  font-weight: 600;
  font-size: 0.875rem;
  border-bottom: 1px solid #e2e8f0;
}

.data-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #e2e8f0;
  font-size: 0.875rem;
  color: #475569;
}

.data-table tbody tr:hover {
  background-color: #f8fafc;
}

.symbol-cell {
  font-weight: 600;
  color: #0f172a;
}

.trade-side {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.trade-side.buy {
  background-color: #d1fae5;
  color: #065f46;
}

.trade-side.sell {
  background-color: #fee2e2;
  color: #991b1b;
}

.amount-cell {
  font-weight: 600;
  color: #0f172a;
}

.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-badge.success {
  background-color: #d1fae5;
  color: #065f46;
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
  padding: 0.5rem 1rem;
  background-color: transparent;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  color: #64748b;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 200ms ease;
}

.page-btn:hover:not(:disabled) {
  background-color: #f1f5f9;
  border-color: #94a3b8;
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  color: #64748b;
  font-size: 0.875rem;
}

/* 加载状态 */
.loading-state {
  text-align: center;
  padding: 3rem 2rem;
  color: #94a3b8;
}

.spinner {
  width: 40px;
  height: 40px;
  margin: 0 auto 1rem;
  border: 3px solid #e2e8f0;
  border-top-color: #2563eb;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 3rem 2rem;
  color: #94a3b8;
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
  background: #ffffff;
  border-radius: 0.5rem;
  width: 90%;
  max-width: 500px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1);
}

.modal-sm {
  max-width: 400px;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.modal-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: #0f172a;
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
  border-radius: 0.375rem;
  color: #94a3b8;
  font-size: 1.5rem;
  cursor: pointer;
  transition: all 150ms ease;
}

.modal-close:hover {
  background-color: #f1f5f9;
  color: #0f172a;
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
  border-top: 1px solid #e2e8f0;
}

/* 表单样式 */
.form-group {
  margin-bottom: 1.5rem;
}

.form-label {
  display: block;
  font-size: 0.875rem;
  font-weight: 500;
  color: #0f172a;
  margin-bottom: 0.5rem;
}

.form-control {
  width: 100%;
  padding: 0.625rem 0.875rem;
  background-color: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  color: #0f172a;
  font-size: 0.875rem;
  transition: all 200ms ease;
}

.form-control:focus {
  outline: none;
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

.form-control:disabled {
  background-color: #f1f5f9;
  cursor: not-allowed;
}

.form-hint {
  font-size: 0.75rem;
  color: #94a3b8;
  margin-top: 0.25rem;
}

.popular-symbols {
  margin-top: 1.5rem;
}

.popular-title {
  font-size: 0.875rem;
  font-weight: 500;
  color: #64748b;
  margin-bottom: 0.75rem;
}

.symbol-chips {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.symbol-chip {
  padding: 0.375rem 0.875rem;
  background-color: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 9999px;
  color: #64748b;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 150ms ease;
}

.symbol-chip:hover:not(:disabled) {
  background-color: #2563eb;
  border-color: #2563eb;
  color: white;
}

.symbol-chip:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 确认消息 */
.confirm-message {
  text-align: center;
  padding: 1rem 0;
}

.warning-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: 1rem;
}

.confirm-message p {
  margin: 0.5rem 0;
  color: #0f172a;
}

.warning-text {
  font-size: 0.875rem;
  color: #64748b;
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
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1);
  font-weight: 500;
  z-index: 1000;
}

.toast.success {
  border-color: #10b981;
  color: #10b981;
}

.toast.error {
  border-color: #ef4444;
  color: #ef4444;
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

  .price-grid,
  .balance-grid {
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
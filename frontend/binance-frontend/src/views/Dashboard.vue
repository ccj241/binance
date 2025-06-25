<template>
  <div class="dashboard-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">
        <span class="gradient-text">交易仪表盘</span>
      </h1>
      <p class="page-subtitle">实时监控您的交易数据</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-overview">
      <div class="stat-card">
        <div class="stat-icon">
          <i>💰</i>
        </div>
        <div class="stat-content">
          <h3>总资产价值</h3>
          <p class="stat-value">{{ formatCurrency(totalAssetValue) }}</p>
          <p class="stat-change positive">+12.5%</p>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">
          <i>📈</i>
        </div>
        <div class="stat-content">
          <h3>今日盈亏</h3>
          <p class="stat-value">{{ formatCurrency(todayPnL) }}</p>
          <p class="stat-change" :class="todayPnL >= 0 ? 'positive' : 'negative'">
            {{ todayPnL >= 0 ? '+' : '' }}{{ ((todayPnL / totalAssetValue) * 100).toFixed(2) }}%
          </p>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">
          <i>🔄</i>
        </div>
        <div class="stat-content">
          <h3>活跃交易</h3>
          <p class="stat-value">{{ activeTradesCount }}</p>
          <p class="stat-subtitle">{{ pendingOrdersCount }} 待处理订单</p>
        </div>
      </div>
      <div class="stat-card">
        <div class="stat-icon">
          <i>⚡</i>
        </div>
        <div class="stat-content">
          <h3>24h 交易量</h3>
          <p class="stat-value">{{ formatVolume(volume24h) }}</p>
          <p class="stat-subtitle">{{ tradesCount24h }} 笔交易</p>
        </div>
      </div>
    </div>

    <!-- 实时价格监控 -->
    <section class="price-section">
      <div class="section-header">
        <h2 class="section-title">
          <i>📊</i> 实时价格监控
        </h2>
        <div class="section-actions">
          <button @click="showAddSymbolModal = true" class="add-btn">
            <i>+</i> 添加交易对
          </button>
        </div>
      </div>

      <div v-if="Object.keys(prices).length === 0" class="empty-state">
        <div class="empty-icon">📉</div>
        <p>还未添加任何交易对</p>
        <button @click="showAddSymbolModal = true" class="primary-btn">
          添加第一个交易对
        </button>
      </div>

      <div v-else class="price-grid">
        <div v-for="(price, symbol) in prices" :key="symbol" class="price-card">
          <div class="price-header">
            <h3>{{ symbol }}</h3>
            <button @click="confirmDeleteSymbol(symbol)" class="delete-btn">
              <i>×</i>
            </button>
          </div>
          <div class="price-content">
            <p class="current-price">${{ formatPrice(price) }}</p>
            <p class="price-change" :class="getPriceChangeClass(symbol)">
              <i :class="getPriceChangeIcon(symbol)"></i>
              {{ getPriceChangePercent(symbol) }}%
            </p>
          </div>
          <div class="price-chart">
            <div class="mini-chart" :id="`chart-${symbol}`"></div>
          </div>
        </div>
      </div>
    </section>

    <!-- 账户余额 -->
    <section class="balance-section">
      <div class="section-header">
        <h2 class="section-title">
          <i>💼</i> 账户余额
        </h2>
        <div class="section-actions">
          <button @click="fetchBalances" class="refresh-btn">
            <i>🔄</i> 刷新
          </button>
        </div>
      </div>

      <div v-if="balances.length === 0" class="empty-state">
        <div class="empty-icon">💳</div>
        <p>暂无余额信息</p>
      </div>

      <div v-else class="balance-grid">
        <div v-for="balance in filteredBalances" :key="balance.asset" class="balance-card">
          <div class="balance-header">
            <img :src="getCoinIcon(balance.asset)" :alt="balance.asset" class="coin-icon">
            <h4>{{ balance.asset }}</h4>
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
              <span class="value total">{{ formatBalance(balance.free + balance.locked) }}</span>
            </div>
          </div>
          <div class="balance-value">
            ≈ ${{ formatCurrency(getBalanceValue(balance)) }}
          </div>
        </div>
      </div>
    </section>

    <!-- 最近交易 -->
    <section class="trades-section">
      <div class="section-header">
        <h2 class="section-title">
          <i>📜</i> 最近交易记录
        </h2>
        <div class="section-actions">
          <select v-model="tradeFilter" class="filter-select">
            <option value="all">全部</option>
            <option value="buy">买入</option>
            <option value="sell">卖出</option>
          </select>
        </div>
      </div>

      <div v-if="filteredTrades.length === 0" class="empty-state">
        <div class="empty-icon">📋</div>
        <p>暂无交易记录</p>
      </div>

      <div v-else class="trades-table-wrapper">
        <table class="modern-table">
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
          <tr v-for="trade in paginatedTrades" :key="trade.id" class="table-row">
            <td>{{ formatTradeTime(trade.time) }}</td>
            <td class="symbol-cell">{{ trade.symbol }}</td>
            <td>
                <span :class="['trade-side', trade.side.toLowerCase()]">
                  {{ trade.side === 'BUY' ? '买入' : '卖出' }}
                </span>
            </td>
            <td>${{ formatPrice(trade.price) }}</td>
            <td>{{ formatQuantity(trade.qty) }}</td>
            <td class="total-cell">${{ formatCurrency(trade.price * trade.qty) }}</td>
            <td>
              <span class="status-badge success">已完成</span>
            </td>
          </tr>
          </tbody>
        </table>
      </div>

      <div v-if="filteredTrades.length > pageSize" class="pagination">
        <button :disabled="currentPage === 1" @click="currentPage--" class="page-btn">
          <i>←</i> 上一页
        </button>
        <span class="page-info">{{ currentPage }} / {{ totalPages }}</span>
        <button :disabled="currentPage === totalPages" @click="currentPage++" class="page-btn">
          下一页 <i>→</i>
        </button>
      </div>
    </section>

    <!-- 添加交易对弹窗 -->
    <div v-if="showAddSymbolModal" class="modal-overlay" @click="showAddSymbolModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>添加交易对</h3>
          <button @click="showAddSymbolModal = false" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <input
              v-model="newSymbol"
              @keyup.enter="addSymbol"
              placeholder="输入交易对 (如 BTCUSDT)"
              class="modal-input"
              ref="symbolInput"
          />
          <div class="popular-symbols">
            <p>热门交易对：</p>
            <div class="symbol-chips">
              <button
                  v-for="symbol in popularSymbols"
                  :key="symbol"
                  @click="newSymbol = symbol; addSymbol()"
                  class="symbol-chip"
              >
                {{ symbol }}
              </button>
            </div>
          </div>
        </div>
        <div class="modal-footer">
          <button @click="showAddSymbolModal = false" class="cancel-btn">取消</button>
          <button @click="addSymbol" :disabled="!newSymbol || isAddingSymbol" class="confirm-btn">
            {{ isAddingSymbol ? '添加中...' : '确认添加' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认弹窗 -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click="cancelDeleteSymbol">
      <div class="modal-content danger" @click.stop>
        <div class="modal-header">
          <h3>确认删除</h3>
          <button @click="cancelDeleteSymbol" class="close-btn">×</button>
        </div>
        <div class="modal-body">
          <div class="warning-icon">⚠️</div>
          <p>确定要删除交易对 <strong>{{ symbolToDelete }}</strong> 吗？</p>
          <p class="warning-text">删除后将停止价格监控，相关的策略和订单数据不会被删除。</p>
        </div>
        <div class="modal-footer">
          <button @click="cancelDeleteSymbol" class="cancel-btn">取消</button>
          <button @click="deleteSymbol" class="danger-btn" :disabled="isDeletingSymbol">
            {{ isDeletingSymbol ? '删除中...' : '确认删除' }}
          </button>
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

      // 交易相关
      trades: [],
      tradeFilter: 'all',
      currentPage: 1,
      pageSize: 10,

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
      return this.balances.filter(b => (b.free + b.locked) > 0.00001);
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
    this.initDashboard();
  },
  beforeUnmount() {
    if (this.priceInterval) {
      clearInterval(this.priceInterval);
    }
  },
  methods: {
    async initDashboard() {
      await Promise.all([
        this.fetchPrices(),
        this.fetchBalances(),
        this.fetchTrades(),
      ]);

      // 启动价格更新定时器
      this.priceInterval = setInterval(this.fetchPrices, 5000);
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
      }).format(value);
    },

    formatPrice(price) {
      if (price > 1000) return price.toFixed(2);
      if (price > 1) return price.toFixed(4);
      return price.toFixed(8);
    },

    formatQuantity(qty) {
      return parseFloat(qty).toFixed(8).replace(/\.?0+$/, '');
    },

    formatBalance(balance) {
      if (balance === 0) return '0';
      if (balance < 0.00001) return '< 0.00001';
      return this.formatQuantity(balance);
    },

    formatVolume(volume) {
      if (volume >= 1000000) return (volume / 1000000).toFixed(2) + 'M';
      if (volume >= 1000) return (volume / 1000).toFixed(2) + 'K';
      return volume.toFixed(2);
    },

    formatTradeTime(timestamp) {
      const date = new Date(timestamp);
      const now = new Date();
      const diff = now - date;

      if (diff < 60000) return '刚刚';
      if (diff < 3600000) return Math.floor(diff / 60000) + '分钟前';
      if (diff < 86400000) return Math.floor(diff / 3600000) + '小时前';

      return date.toLocaleDateString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
      });
    },

    getCoinIcon(asset) {
      // 这里可以返回真实的币种图标URL
      return `https://cryptoicons.org/api/icon/${asset.toLowerCase()}/50`;
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
      return (balance.free + balance.locked) * price;
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
        const response = await axios.get('/prices', {
          headers: this.getAuthHeaders(),
        });
        this.prices = response.data.prices || {};

        // 更新价格历史
        Object.entries(this.prices).forEach(([symbol, price]) => {
          if (!this.priceHistory[symbol]) {
            this.priceHistory[symbol] = [];
          }
          this.priceHistory[symbol].push({
            time: new Date(),
            price: price
          });
          // 保留最近50个数据点
          if (this.priceHistory[symbol].length > 50) {
            this.priceHistory[symbol].shift();
          }
        });
      } catch (error) {
        console.error('获取价格失败:', error);
      }
    },

    async fetchBalances() {
      try {
        const response = await axios.get('/balance', {
          headers: this.getAuthHeaders(),
        });
        this.balances = response.data.balances || [];
      } catch (error) {
        console.error('获取余额失败:', error);
        this.showToast(error.response?.data?.error || '获取余额失败', 'error');
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
        console.error('获取交易记录失败:', error);
        this.showToast(error.response?.data?.error || '获取交易记录失败', 'error');
      }
    },

    async addSymbol() {
      if (!this.newSymbol.trim()) {
        this.showToast('请输入有效的交易对', 'error');
        return;
      }

      this.isAddingSymbol = true;
      try {
        const response = await axios.post('/symbols',
            { symbol: this.newSymbol.toUpperCase() },
            { headers: this.getAuthHeaders() }
        );

        this.showToast('交易对添加成功');
        this.newSymbol = '';
        this.showAddSymbolModal = false;
        await this.fetchPrices();
      } catch (error) {
        console.error('添加交易对失败:', error);
        this.showToast(error.response?.data?.error || '添加交易对失败', 'error');
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
        const response = await axios.delete('/symbols', {
          data: { symbol: this.symbolToDelete },
          headers: this.getAuthHeaders()
        });

        this.showToast('交易对删除成功');
        delete this.prices[this.symbolToDelete];
        delete this.priceHistory[this.symbolToDelete];
        this.cancelDeleteSymbol();
        await this.fetchPrices();
      } catch (error) {
        console.error('删除交易对失败:', error);
        this.showToast(error.response?.data?.error || '删除交易对失败', 'error');
      } finally {
        this.isDeletingSymbol = false;
      }
    },
  },
};
</script>

<style scoped>
/* 容器样式 */
.dashboard-container {
  min-height: 100vh;
  background: linear-gradient(135deg, #0f0f0f 0%, #1a1a1a 100%);
  color: #ffffff;
  padding: 2rem;
}

/* 页面标题 */
.page-header {
  text-align: center;
  margin-bottom: 3rem;
  animation: fadeInDown 0.6s ease-out;
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
  color: #888;
  font-size: 1.1rem;
}

/* 统计卡片 */
.stats-overview {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 1.5rem;
  margin-bottom: 3rem;
}

.stat-card {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  padding: 2rem;
  display: flex;
  align-items: center;
  gap: 1.5rem;
  transition: all 0.3s ease;
  animation: fadeInUp 0.6s ease-out;
}

.stat-card:hover {
  transform: translateY(-5px);
  background: rgba(255, 255, 255, 0.08);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.stat-icon {
  width: 60px;
  height: 60px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.8rem;
}

.stat-content h3 {
  font-size: 0.9rem;
  color: #888;
  margin-bottom: 0.5rem;
  font-weight: 400;
}

.stat-value {
  font-size: 1.8rem;
  font-weight: 700;
  margin-bottom: 0.25rem;
}

.stat-change {
  font-size: 0.9rem;
  font-weight: 500;
}

.stat-change.positive {
  color: #22c55e;
}

.stat-change.negative {
  color: #ef4444;
}

.stat-subtitle {
  font-size: 0.85rem;
  color: #666;
}

/* Section 样式 */
section {
  margin-bottom: 3rem;
  animation: fadeIn 0.8s ease-out;
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
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.section-actions {
  display: flex;
  gap: 1rem;
}

/* 价格卡片 */
.price-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 1.5rem;
}

.price-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.price-card:hover {
  transform: translateY(-3px);
  background: rgba(255, 255, 255, 0.08);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
}

.price-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.price-header h3 {
  font-size: 1.2rem;
  font-weight: 600;
}

.delete-btn {
  background: none;
  border: none;
  color: #ef4444;
  font-size: 1.5rem;
  cursor: pointer;
  width: 30px;
  height: 30px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.delete-btn:hover {
  background: rgba(239, 68, 68, 0.1);
}

.current-price {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.price-change {
  font-size: 1.1rem;
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.price-change.positive {
  color: #22c55e;
}

.price-change.negative {
  color: #ef4444;
}

.mini-chart {
  height: 60px;
  margin-top: 1rem;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
}

/* 余额卡片 */
.balance-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 1.5rem;
}

.balance-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.balance-card:hover {
  transform: translateY(-3px);
  background: rgba(255, 255, 255, 0.08);
}

.balance-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.coin-icon {
  width: 32px;
  height: 32px;
  border-radius: 50%;
}

.balance-header h4 {
  font-size: 1.1rem;
  font-weight: 600;
}

.balance-details {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.balance-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.balance-item .label {
  color: #888;
  font-size: 0.9rem;
}

.balance-item .value {
  font-weight: 500;
}

.balance-item .value.total {
  font-weight: 700;
  color: #667eea;
}

.balance-value {
  font-size: 0.9rem;
  color: #888;
  padding-top: 0.75rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

/* 表格样式 */
.trades-table-wrapper {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 16px;
  overflow: hidden;
}

.modern-table {
  width: 100%;
  border-collapse: collapse;
}

.modern-table th {
  background: rgba(255, 255, 255, 0.05);
  padding: 1rem;
  text-align: left;
  font-weight: 600;
  color: #888;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.modern-table td {
  padding: 1rem;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
}

.table-row {
  transition: all 0.3s ease;
}

.table-row:hover {
  background: rgba(255, 255, 255, 0.05);
}

.symbol-cell {
  font-weight: 600;
}

.trade-side {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.85rem;
  font-weight: 500;
}

.trade-side.buy {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
}

.trade-side.sell {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
}

.total-cell {
  font-weight: 600;
}

.status-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.85rem;
  font-weight: 500;
}

.status-badge.success {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
}

/* 按钮样式 */
.add-btn, .refresh-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.add-btn:hover, .refresh-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
}

.primary-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  border: none;
  padding: 1rem 2rem;
  border-radius: 12px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.primary-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
}

/* 筛选下拉框 */
.filter-select {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: white;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.filter-select:hover {
  background: rgba(255, 255, 255, 0.08);
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
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.page-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.08);
}

.page-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.page-info {
  color: #888;
  font-size: 0.9rem;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  color: #888;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.3;
}

/* 弹窗样式 */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  backdrop-filter: blur(5px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  animation: fadeIn 0.3s ease-out;
}

.modal-content {
  background: #1a1a1a;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  width: 90%;
  max-width: 500px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
  animation: slideUp 0.3s ease-out;
}

.modal-content.danger {
  border-color: rgba(239, 68, 68, 0.3);
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 2rem;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.modal-header h3 {
  font-size: 1.5rem;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  color: #888;
  font-size: 2rem;
  cursor: pointer;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.05);
  color: white;
}

.modal-body {
  padding: 2rem;
}

.modal-input {
  width: 100%;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: white;
  padding: 1rem;
  border-radius: 12px;
  font-size: 1rem;
  transition: all 0.3s ease;
}

.modal-input:focus {
  outline: none;
  background: rgba(255, 255, 255, 0.08);
  border-color: #667eea;
}

.popular-symbols {
  margin-top: 1.5rem;
}

.popular-symbols p {
  color: #888;
  font-size: 0.9rem;
  margin-bottom: 1rem;
}

.symbol-chips {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.symbol-chip {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  font-size: 0.9rem;
}

.symbol-chip:hover {
  background: rgba(102, 126, 234, 0.2);
  border-color: #667eea;
}

.warning-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.warning-text {
  background: rgba(255, 193, 7, 0.1);
  border: 1px solid rgba(255, 193, 7, 0.3);
  color: #fbbf24;
  padding: 1rem;
  border-radius: 8px;
  font-size: 0.9rem;
  margin-top: 1rem;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  padding: 2rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
}

.cancel-btn {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.cancel-btn:hover {
  background: rgba(255, 255, 255, 0.08);
}

.confirm-btn {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.confirm-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(102, 126, 234, 0.4);
}

.confirm-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.danger-btn {
  background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
  border: none;
  color: white;
  padding: 0.75rem 1.5rem;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.danger-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(239, 68, 68, 0.4);
}

.danger-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
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
  font-size: 1.2rem;
}

/* 动画 */
@keyframes fadeIn {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes fadeInDown {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(50px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
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
  .dashboard-container {
    padding: 1rem;
  }

  .page-title {
    font-size: 2rem;
  }

  .stats-overview {
    grid-template-columns: 1fr;
  }

  .price-grid {
    grid-template-columns: 1fr;
  }

  .balance-grid {
    grid-template-columns: 1fr;
  }

  .section-header {
    flex-direction: column;
    gap: 1rem;
    align-items: flex-start;
  }

  .trades-table-wrapper {
    overflow-x: auto;
  }

  .modern-table {
    min-width: 600px;
  }
}
</style>
<template>
  <div class="admin-container">
    <h2 class="admin-title">管理员面板</h2>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-value">{{ totalValue.toFixed(2) }} USDT</div>
        <div class="stat-label">总资产价值</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ totalProfit > 0 ? '+' : '' }}{{ totalProfit.toFixed(2) }} USDT</div>
        <div class="stat-label">总收益</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ totalQuantity.toFixed(4) }} BTC</div>
        <div class="stat-label">总持仓量</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ profitRate.toFixed(2) }}%</div>
        <div class="stat-label">收益率</div>
      </div>
    </div>

    <!-- 标签页 -->
    <div class="tabs">
      <button
          v-for="tab in tabs"
          :key="tab"
          @click="activeTab = tab"
          :class="['tab-button', { active: activeTab === tab }]"
      >
        {{ tabLabels[tab] }}
      </button>
    </div>

    <!-- 持仓管理 -->
    <div v-if="activeTab === 'positions'" class="section">
      <h3 class="section-title">持仓管理</h3>
      <div class="table-container">
        <table class="data-table">
          <thead>
          <tr>
            <th>交易对</th>
            <th>数量</th>
            <th>平均成本</th>
            <th>当前价格</th>
            <th>总价值</th>
            <th>盈亏</th>
            <th>盈亏率</th>
            <th>操作</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="position in positions" :key="position.id">
            <td>{{ position.symbol }}</td>
            <td>{{ position.quantity }}</td>
            <td>${{ position.avgCost }}</td>
            <td>${{ position.currentPrice }}</td>
            <td>${{ position.totalValue }}</td>
            <td :class="position.profit > 0 ? 'profit' : 'loss'">
              {{ position.profit > 0 ? '+' : '' }}${{ position.profit }}
            </td>
            <td :class="position.profitRate > 0 ? 'profit' : 'loss'">
              {{ position.profitRate > 0 ? '+' : '' }}{{ position.profitRate }}%
            </td>
            <td>
              <button @click="editPosition(position)" class="btn-edit">编辑</button>
              <button @click="deletePosition(position.id)" class="btn-delete">删除</button>
            </td>
          </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 交易记录 -->
    <div v-if="activeTab === 'trades'" class="section">
      <h3 class="section-title">交易记录</h3>
      <div class="table-container">
        <table class="data-table">
          <thead>
          <tr>
            <th>时间</th>
            <th>交易对</th>
            <th>类型</th>
            <th>价格</th>
            <th>数量</th>
            <th>总额</th>
            <th>状态</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="trade in trades" :key="trade.id">
            <td>{{ new Date(trade.time).toLocaleString() }}</td>
            <td>{{ trade.symbol }}</td>
            <td :class="trade.type === 'BUY' ? 'buy' : 'sell'">{{ trade.type }}</td>
            <td>${{ trade.price }}</td>
            <td>{{ trade.quantity }}</td>
            <td>${{ trade.total }}</td>
            <td>{{ trade.status }}</td>
          </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 设置 -->
    <div v-if="activeTab === 'settings'" class="section">
      <h3 class="section-title">系统设置</h3>
      <div class="settings-form">
        <div class="form-group">
          <label>API Key</label>
          <input v-model="settings.apiKey" type="text" placeholder="输入 Binance API Key">
        </div>
        <div class="form-group">
          <label>API Secret</label>
          <input v-model="settings.apiSecret" type="password" placeholder="输入 Binance API Secret">
        </div>
        <div class="form-group">
          <label>刷新间隔（秒）</label>
          <input v-model.number="settings.refreshInterval" type="number" min="5" max="300">
        </div>
        <button @click="saveSettings" class="btn-save">保存设置</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import axios from 'axios'

// 状态管理
const activeTab = ref('positions')
const tabs = ['positions', 'trades', 'settings']
const tabLabels = {
  positions: '持仓管理',
  trades: '交易记录',
  settings: '系统设置'
}

// 数据
const positions = ref([])
const trades = ref([])
const settings = ref({
  apiKey: '',
  apiSecret: '',
  refreshInterval: 30
})

// 计算属性
const totalValue = computed(() => {
  return positions.value.reduce((sum, pos) => sum + pos.totalValue, 0)
})

const totalProfit = computed(() => {
  return positions.value.reduce((sum, pos) => sum + pos.profit, 0)
})

const totalQuantity = computed(() => {
  return positions.value.reduce((sum, pos) => sum + pos.quantity, 0)
})

const profitRate = computed(() => {
  const totalCost = positions.value.reduce((sum, pos) => sum + (pos.quantity * pos.avgCost), 0)
  return totalCost > 0 ? (totalProfit.value / totalCost) * 100 : 0
})

// 方法
const loadPositions = async () => {
  try {
    const response = await axios.get('/api/positions')
    positions.value = response.data
  } catch (error) {
    console.error('加载持仓失败:', error)
  }
}

const loadTrades = async () => {
  try {
    const response = await axios.get('/api/trades')
    trades.value = response.data
  } catch (error) {
    console.error('加载交易记录失败:', error)
  }
}

const editPosition = (position) => {
  // 实现编辑逻辑
  console.log('编辑持仓:', position)
}

const deletePosition = async (id) => {
  if (confirm('确定要删除这个持仓吗？')) {
    try {
      await axios.delete(`/api/positions/${id}`)
      await loadPositions()
    } catch (error) {
      console.error('删除持仓失败:', error)
    }
  }
}

const saveSettings = async () => {
  try {
    await axios.post('/api/settings', settings.value)
    alert('设置保存成功')
  } catch (error) {
    console.error('保存设置失败:', error)
    alert('保存设置失败')
  }
}

// 生命周期
onMounted(() => {
  loadPositions()
  loadTrades()
})
</script>

<style scoped>
.admin-container {
  padding: 20px;
  max-width: 1400px;
  margin: 0 auto;
}

.admin-title {
  color: #f0b90b;
  font-size: 28px;
  margin-bottom: 30px;
  text-align: center;
}

/* 统计卡片 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 20px;
  margin-bottom: 40px;
}

.stat-card {
  background: #1e2329;
  border-radius: 10px;
  padding: 25px;
  text-align: center;
  border: 1px solid #2b3139;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #f0b90b;
  margin-bottom: 10px;
}

.stat-label {
  color: #848e9c;
  font-size: 14px;
}

/* 标签页 */
.tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 30px;
  border-bottom: 2px solid #2b3139;
}

.tab-button {
  background: none;
  border: none;
  color: #848e9c;
  padding: 12px 24px;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.3s ease;
  position: relative;
}

.tab-button:hover {
  color: #f0b90b;
}

.tab-button.active {
  color: #f0b90b;
}

.tab-button.active::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  right: 0;
  height: 2px;
  background: #f0b90b;
}

/* 区块样式 */
.section {
  background: #1e2329;
  border-radius: 10px;
  padding: 30px;
  border: 1px solid #2b3139;
}

.section-title {
  color: #fff;
  font-size: 20px;
  margin-bottom: 20px;
}

/* 表格样式 */
.table-container {
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th,
.data-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #2b3139;
}

.data-table th {
  background: #0b0e11;
  color: #848e9c;
  font-weight: 500;
  font-size: 14px;
}

.data-table td {
  color: #eaecef;
  font-size: 14px;
}

.data-table tr:hover {
  background: #252930;
}

/* 盈亏样式 */
.profit {
  color: #0ecb81;
}

.loss {
  color: #f6465d;
}

.buy {
  color: #0ecb81;
}

.sell {
  color: #f6465d;
}

/* 按钮样式 */
.btn-edit,
.btn-delete {
  padding: 6px 12px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  margin-right: 5px;
  transition: all 0.3s ease;
}

.btn-edit {
  background: #2b3139;
  color: #f0b90b;
}

.btn-edit:hover {
  background: #3b4149;
}

.btn-delete {
  background: #2b3139;
  color: #f6465d;
}

.btn-delete:hover {
  background: #3b4149;
}

/* 设置表单 */
.settings-form {
  max-width: 600px;
}

.form-group {
  margin-bottom: 20px;
}

.form-group label {
  display: block;
  color: #848e9c;
  margin-bottom: 8px;
  font-size: 14px;
}

.form-group input {
  width: 100%;
  padding: 10px;
  background: #0b0e11;
  border: 1px solid #2b3139;
  border-radius: 4px;
  color: #eaecef;
  font-size: 14px;
}

.form-group input:focus {
  outline: none;
  border-color: #f0b90b;
}

.btn-save {
  background: #f0b90b;
  color: #0b0e11;
  border: none;
  padding: 12px 30px;
  border-radius: 4px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
}

.btn-save:hover {
  background: #d4a10a;
}
</style>
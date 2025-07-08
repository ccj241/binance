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

      <!-- 新增：USDT余额卡片 -->
      <div class="stat-card">
        <div class="stat-icon">
          <span>💵</span>
        </div>
        <div class="stat-content">
          <div class="stat-label">USDT可用余额</div>
          <div class="stat-value">{{ formatCurrency(availableBalance) }}</div>
        </div>
      </div>
    </div>

    <!-- 策略列表 -->
    <div class="strategies-section">
      <div class="section-header">
        <h2 class="section-title">策略列表</h2>
        <button @click="showCreateModal = true" class="btn btn-primary" @click.prevent="openCreateModal">
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
                <span v-if="strategy.strategyType === 'iceberg'" class="type-badge">
                  冰山
                </span>
                <span v-if="strategy.strategyType === 'slow_iceberg'" class="type-badge slow">
  慢冰山
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
                <span class="detail-label">开仓价格浮动</span>
                <span class="detail-value">{{ strategy.entryPriceFloat }}‱</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">数量</span>
                <span class="detail-value">{{ formatQuantity(strategy.quantity) }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">止盈价格</span>
                <span class="detail-value success">
                  {{ formatPrice(strategy.takeProfitPrice) }}
                  <span class="percentage">(+{{ strategy.takeProfitRate }}%)</span>
                </span>
              </div>
              <div class="detail-item">
                <span class="detail-label">止损价格</span>
                <span class="detail-value danger">
                  {{ strategy.stopLossPrice > 0 ? formatPrice(strategy.stopLossPrice) : '未设置' }}
                  <span v-if="strategy.stopLossRate > 0" class="percentage">
                    (-{{ strategy.stopLossRate }}%)
                  </span>
                </span>
              </div>
            </div>

            <!-- 冰山策略详情 -->
            <div v-if="(strategy.strategyType === 'iceberg' || strategy.strategyType === 'slow_iceberg') && strategy.icebergQuantities" class="iceberg-details">
              <div class="iceberg-title">{{ strategy.strategyType === 'slow_iceberg' ? '慢冰山' : '冰山' }}策略配置</div>
              <div class="iceberg-info-grid">
                <div class="iceberg-info-item">
                  <span class="info-label">层数</span>
                  <span class="info-value">{{ strategy.icebergLevels }}层</span>
                </div>
                <div class="iceberg-info-item">
                  <span class="info-label">数量分配</span>
                  <span class="info-value">{{ formatIcebergQuantities(strategy.icebergQuantities) }}</span>
                </div>
                <div class="iceberg-info-item">
                  <span class="info-label">价格间隔</span>
                  <span class="info-value">{{ formatIcebergPriceGaps(strategy.icebergPriceGaps, strategy.strategyType) }}</span>
                </div>
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
      <div v-if="showCreateModal" class="modal-overlay">
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
                    placeholder="留空自动生成"
                    class="form-control"
                    @input="isAutoGeneratedName = false"
                />
              </div>

              <div class="form-group">
                <label class="form-label">策略类型</label>
                <select v-model="strategyForm.strategyType" class="form-control" @change="onStrategyTypeChange" required>
                  <option value="simple">简单策略</option>
                  <option value="iceberg">冰山策略</option>
                  <option value="slow_iceberg">慢冰山策略</option>
                </select>
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
                <select
                    v-model="strategyForm.side"
                    class="form-control"
                    :disabled="editingStrategy"
                    required
                    @change="onSideChange"
                >
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
                    required
                    @input="generateStrategyName"
                />
              </div>

              <div class="form-group">
                <label class="form-label">
                  开仓价格浮动 (‱)
                  <span class="form-hint">
      相对于买卖1价的浮动万分比，用于避免吃单
    </span>
                </label>
                <input
                    v-model.number="strategyForm.entryPriceFloat"
                    type="number"
                    step="1"
                    min="0"
                    placeholder="0"
                    class="form-control"
                    @input="generateStrategyName"
                />
                <div class="calculated-price-hint" v-if="strategyForm.basePrice > 0">
                  <span v-if="!strategyForm.entryPriceFloat || strategyForm.entryPriceFloat === 0">
                    将按买卖1价挂单（可能吃单）
                  </span>
                  <span v-else-if="strategyForm.side === 'LONG'">
                    挂单价 = 卖1价 × {{ (1 - strategyForm.entryPriceFloat / 10000).toFixed(4) }}
                    <span class="price-example" v-if="strategyForm.basePrice">
                      ≈ {{ calculateEstimatedEntryPrice() }}
                    </span>
                  </span>
                  <span v-else-if="strategyForm.side === 'SHORT'">
                    挂单价 = 买1价 × {{ (1 + strategyForm.entryPriceFloat / 10000).toFixed(4) }}
                    <span class="price-example" v-if="strategyForm.basePrice">
                      ≈ {{ calculateEstimatedEntryPrice() }}
                    </span>
                  </span>
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
                <label class="form-label">开仓本金 (USDT)</label>
                <input
                    v-model.number="strategyForm.quantity"
                    type="number"
                    step="0.001"
                    placeholder="投入的本金数量"
                    class="form-control"
                    :class="{ 'warning': strategyForm.quantity > availableBalance }"
                    required
                />
                <span class="form-hint" :class="{ 'hint-warning': strategyForm.quantity > availableBalance }">
    可用余额: {{ availableBalance.toFixed(2) }} USDT
    <span v-if="strategyForm.quantity > availableBalance" class="warning-text">
      (余额不足，请确保创建策略前充值)
    </span>
  </span>
              </div>

              <div class="form-group">
                <label class="form-label">止盈万分比 (‱)</label>
                <input
                    v-model.number="strategyForm.takeProfitRate"
                    type="number"
                    step="1"
                    min="1"
                    placeholder="扣除手续费后的净利润万分比"
                    class="form-control"
                    required
                    @input="generateStrategyName"
                />
              </div>

              <div class="form-group">
                <label class="form-label">止损万分比 (‱) <span class="optional">可选</span></label>
                <input
                    v-model.number="strategyForm.stopLossRate"
                    type="number"
                    step="1"
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
            <!-- 冰山策略配置 -->
            <template v-if="strategyForm.strategyType === 'iceberg' || strategyForm.strategyType === 'slow_iceberg'">
              <div class="iceberg-config-section">
                <h4 class="config-title">{{ strategyForm.strategyType === 'slow_iceberg' ? '慢冰山' : '冰山' }}策略配置</h4>

                <div class="form-grid">
                  <div class="form-group">
                    <label class="form-label">
                      冰山层数
                      <span class="form-hint">将订单分为几层</span>
                    </label>
                    <select v-model.number="strategyForm.icebergLevels" class="form-control" @change="updateIcebergDefaults">
                      <option :value="2">2层</option>
                      <option :value="3">3层</option>
                      <option :value="4">4层</option>
                      <option :value="5">5层</option>
                      <option :value="6">6层</option>
                    </select>
                  </div>
                </div>

                <div class="form-group full-width">
                  <label class="form-label">
                    各层数量分配
                    <span class="form-hint">每层占总数量的比例，总和必须为1</span>
                  </label>
                  <div class="iceberg-layers">
                    <div v-for="(quantity, index) in strategyForm.icebergQuantities.slice(0, strategyForm.icebergLevels)" :key="'q' + index" class="iceberg-layer">
                      <span class="layer-label">第{{ index + 1 }}层</span>
                      <input
                          v-model.number="strategyForm.icebergQuantities[index]"
                          type="number"
                          step="0.01"
                          min="0.01"
                          max="1"
                          class="form-control"
                          placeholder="比例"
                          @input="validateIcebergSum"
                      />
                      <span class="layer-percent">{{ (strategyForm.icebergQuantities[index] * 100).toFixed(0) }}%</span>
                    </div>
                  </div>
                  <div v-if="icebergSumError" class="form-error">
                    数量总和必须为1，当前总和: {{ icebergSum.toFixed(2) }}
                  </div>
                </div>

                <div class="form-group full-width">
                  <label class="form-label">
                    各层价格间隔 (‱)
                    <span class="form-hint">
          {{ strategyForm.side === 'LONG' ? '负值表示低于基准价格' : '正值表示高于基准价格' }}
          <span v-if="strategyForm.strategyType === 'slow_iceberg'" class="slow-iceberg-hint">
            （慢冰山策略每层将基于成交时的最新买卖1价计算）
          </span>
        </span>
                  </label>
                  <div class="iceberg-layers">
                    <div v-for="(gap, index) in strategyForm.icebergPriceGaps.slice(0, strategyForm.icebergLevels)" :key="'g' + index" class="iceberg-layer">
                      <span class="layer-label">第{{ index + 1 }}层</span>
                      <input
                          v-model.number="strategyForm.icebergPriceGaps[index]"
                          type="number"
                          step="1"
                          class="form-control"
                          placeholder="万分比"
                      />
                      <span class="layer-percent">
                        {{ strategyForm.icebergPriceGaps[index] > 0 ? '+' : '' }}{{ strategyForm.icebergPriceGaps[index] }}‱
                      </span>
                    </div>
                  </div>
                  <div class="form-hint" style="margin-top: 0.5rem;">
                    <span v-if="strategyForm.side === 'LONG'">
                      做多时使用负值可以在价格下跌时分批建仓，获得更好的平均成本
                    </span>
                    <span v-else-if="strategyForm.side === 'SHORT'">
                      做空时使用正值可以在价格上涨时分批建仓，获得更好的平均成本
                    </span>
                  </div>
                </div>
              </div>
            </template>

            <!-- 策略预览 -->
            <div v-if="strategyForm.basePrice > 0 && strategyForm.quantity > 0" class="strategy-preview">
              <h4 class="preview-title">策略预览</h4>
              <div class="preview-grid">
                <div class="preview-item">
                  <span class="preview-label">投入本金</span>
                  <span class="preview-value">
                    {{ formatCurrency(strategyForm.quantity) }} USDT
                  </span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">开仓价值</span>
                  <span class="preview-value">
                    {{ formatCurrency(strategyForm.quantity * (strategyForm.leverage || 1)) }} USDT
                  </span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">预计合约数量</span>
                  <span class="preview-value">
                    ~{{ calculateEstimatedContractQuantity() }} {{ getContractUnit() }}
                  </span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">所需保证金</span>
                  <span class="preview-value">
                    {{ formatCurrency(strategyForm.quantity) }} USDT
                  </span>
                </div>
                <div class="preview-item">
                  <span class="preview-label">预估开仓价格</span>
                  <span class="preview-value highlight">
                    {{ calculateEstimatedEntryPrice() }}
                    <span class="percentage" v-if="strategyForm.entryPriceFloat > 0">
                      ({{ strategyForm.side === 'LONG' ? '-' : '+' }}{{ strategyForm.entryPriceFloat }}‱)
                    </span>
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
                <div class="preview-item">
                  <span class="preview-label">总手续费</span>
                  <span class="preview-value">
                    {{ formatCurrency(calculateTotalFee()) }}
                  </span>
                </div>
                <div class="preview-item full-width">
                  <span class="preview-label">预计净收益</span>
                  <span class="preview-value" :class="calculateNetProfit() >= 0 ? 'success' : 'danger'">
                    {{ calculateNetProfit() >= 0 ? '+' : '' }}{{ formatCurrency(calculateNetProfit()) }}
                    ({{ calculateNetProfitRate() }}%)
                  </span>
                </div>
              </div>

              <!-- 冰山策略预览 -->
              <div v-if="strategyForm.strategyType === 'iceberg' || strategyForm.strategyType === 'slow_iceberg'" class="iceberg-preview">
                <h5>{{ strategyForm.strategyType === 'slow_iceberg' ? '慢冰山' : '冰山' }}分层预览</h5>
                <div class="iceberg-preview-layers">
                  <div v-for="(_, index) in strategyForm.icebergQuantities.slice(0, strategyForm.icebergLevels)"
                       :key="'preview' + index"
                       class="iceberg-preview-layer">
                    <span class="preview-layer-label">第{{ index + 1 }}层</span>
                    <span class="preview-layer-info">
                      数量: {{ (strategyForm.icebergQuantities[index] * strategyForm.quantity).toFixed(3) }} USDT
                      ({{ (strategyForm.icebergQuantities[index] * 100).toFixed(0) }}%)
                    </span>
                    <span class="preview-layer-price">
                      价格: {{ calculateIcebergLayerPrice(index) }}
                      <span class="price-diff">
                        ({{ strategyForm.icebergPriceGaps[index] > 0 ? '+' : '' }}{{ strategyForm.icebergPriceGaps[index] }}‱)
                      </span>
                    </span>
                  </div>
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
                :disabled="isSubmitting || ((strategyForm.strategyType === 'iceberg' || strategyForm.strategyType === 'slow_iceberg') && icebergSumError)"
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
      <div v-if="showOrdersModal" class="modal-overlay">
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
        <span class="toast-icon">{{ toastType === 'success' ? '✓' : toastType === 'warning' ? '⚠' : '×' }}</span>
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
      availableBalance: 0, // 可用USDT余额
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
        strategyType: 'simple',
        basePrice: 0,
        entryPrice: 0,
        entryPriceFloat: 0,
        leverage: 1,
        quantity: 0,
        takeProfitRate: 0,
        stopLossRate: 0,
        marginType: 'CROSSED',
        icebergLevels: 5,
        icebergQuantities: [0.35, 0.25, 0.2, 0.1, 0.1],
        icebergPriceGaps: [0, -1, -2, -3, -4], // 默认做多的价格间隔（万分比）
      },
      isSubmitting: false,
      toastMessage: '',
      toastType: 'success',
      refreshInterval: null,
      isAutoGeneratedName: false,
      icebergSumError: false,
    };
  },

  computed: {
    icebergSum() {
      return this.strategyForm.icebergQuantities
          .slice(0, this.strategyForm.icebergLevels)
          .reduce((a, b) => a + b, 0);
    }
  },

  mounted() {
    this.fetchSymbols();
    this.fetchStrategies();
    this.fetchPositions();
    this.fetchStats();
    this.fetchBalance(); // 获取余额

    // 定时刷新
    this.refreshInterval = setInterval(() => {
      this.fetchPositions();
      this.fetchStats();
      this.fetchBalance(); // 定时刷新余额
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

    async fetchBalance() {
      try {
        const response = await axios.get('/futures/balance');
        this.availableBalance = response.data.availableBalance || 0;
      } catch (error) {
        console.error('获取期货账户余额失败:', error);
        this.availableBalance = 0;
      }
    },

    async submitStrategy() {
      if (this.isSubmitting) return;

      // 余额不足时只提醒，不阻止创建
      if (this.strategyForm.quantity > this.availableBalance) {
        this.showToast(`警告：USDT余额不足，当前可用: ${this.availableBalance.toFixed(2)} USDT，请确保在策略触发前充值！`, 'warning');
      }

      // 验证冰山策略配置
      if ((this.strategyForm.strategyType === 'iceberg' || this.strategyForm.strategyType === 'slow_iceberg') && this.icebergSumError) {
        this.showToast('冰山策略数量分配总和必须为1', 'error');
        return;
      }

      // 如果没有填写策略名称，使用自动生成的名称
      if (!this.strategyForm.strategyName) {
        this.generateStrategyName();
      }

      this.isSubmitting = true;
      try {
        if (this.editingStrategy) {
          // 更新策略时，只发送允许更新的字段
          const updateData = {
            strategyName: this.strategyForm.strategyName,
            enabled: this.editingStrategy.enabled,
            basePrice: this.strategyForm.basePrice,
            entryPriceFloat: this.strategyForm.entryPriceFloat || 0,  // 确保0值被正确处理
            quantity: this.strategyForm.quantity,
            takeProfitRate: this.strategyForm.takeProfitRate,
            stopLossRate: this.strategyForm.stopLossRate || 0,  // 确保0值被正确处理
          };

          // 如果是冰山策略，添加冰山配置
          if (this.strategyForm.strategyType === 'iceberg' || this.strategyForm.strategyType === 'slow_iceberg') {
            updateData.icebergLevels = this.strategyForm.icebergLevels;
            updateData.icebergQuantities = this.strategyForm.icebergQuantities.slice(0, this.strategyForm.icebergLevels);
            updateData.icebergPriceGaps = this.strategyForm.icebergPriceGaps.slice(0, this.strategyForm.icebergLevels);
          }

          await axios.put(`/futures/strategies/${this.editingStrategy.id}`, updateData);
          this.showToast('策略更新成功');
        } else {
          // 创建策略 - 确保数值字段正确处理
          const submitData = {
            strategyName: this.strategyForm.strategyName,
            symbol: this.strategyForm.symbol,
            side: this.strategyForm.side,
            strategyType: this.strategyForm.strategyType,
            basePrice: parseFloat(this.strategyForm.basePrice) || 0,
            entryPriceFloat: parseFloat(this.strategyForm.entryPriceFloat) || 0,  // 确保转换为数字
            leverage: parseInt(this.strategyForm.leverage) || 1,
            quantity: parseFloat(this.strategyForm.quantity) || 0,
            takeProfitRate: parseFloat(this.strategyForm.takeProfitRate) || 0,
            stopLossRate: parseFloat(this.strategyForm.stopLossRate) || 0,  // 确保转换为数字
            marginType: this.strategyForm.marginType,
          };

          // 如果是冰山策略，添加冰山配置
          if (submitData.strategyType === 'iceberg' || submitData.strategyType === 'slow_iceberg') {
            submitData.icebergLevels = this.strategyForm.icebergLevels;
            submitData.icebergQuantities = this.strategyForm.icebergQuantities.slice(0, this.strategyForm.icebergLevels);
            submitData.icebergPriceGaps = this.strategyForm.icebergPriceGaps.slice(0, this.strategyForm.icebergLevels);
          }

          console.log('提交的策略数据:', submitData);  // 添加调试日志

          await axios.post('/futures/strategies', submitData);
          this.showToast('策略创建成功');
        }

        this.closeCreateModal();
        await this.fetchStrategies();
        await this.fetchBalance();
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

    openCreateModal() {
      this.showCreateModal = true;
      this.fetchBalance(); // 获取最新余额
    },

    editStrategy(strategy) {
      this.editingStrategy = strategy;

// 解析冰山策略配置
      let icebergQuantities = [0.35, 0.25, 0.2, 0.1, 0.1];
      let icebergPriceGaps = strategy.side === 'LONG' ? [0, -1, -2, -3, -4] : [0, 1,2 , 3, 4];

      if (strategy.icebergQuantities) {
        const quantities = strategy.icebergQuantities.split(',').map(q => parseFloat(q.trim()));
        if (quantities.length > 0) {
          icebergQuantities = quantities;
        }
      }

      if (strategy.icebergPriceGaps) {
        const gaps = strategy.icebergPriceGaps.split(',').map(g => parseFloat(g.trim()));
        if (gaps.length > 0) {
          icebergPriceGaps = gaps;
        }
      }

      this.strategyForm = {
        strategyName: strategy.strategyName,
        symbol: strategy.symbol,
        side: strategy.side,
        strategyType: strategy.strategyType || 'simple',
        basePrice: strategy.basePrice,
        entryPrice: strategy.entryPrice,
        entryPriceFloat: strategy.entryPriceFloat || 0,
        leverage: strategy.leverage,
        quantity: strategy.quantity,
        takeProfitRate: strategy.takeProfitRate,
        stopLossRate: strategy.stopLossRate || 0,
        marginType: strategy.marginType,
        icebergLevels: strategy.icebergLevels || 5,
        icebergQuantities: icebergQuantities,
        icebergPriceGaps: icebergPriceGaps,
      };

      this.showCreateModal = true;
      this.fetchBalance();
    },

    closeCreateModal() {
      this.showCreateModal = false;
      this.editingStrategy = null;
      this.resetForm();
      this.fetchBalance();
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
        strategyType: 'simple',
        basePrice: 0,
        entryPrice: 0,
        entryPriceFloat: 0,
        leverage: 1,
        quantity: 0,
        takeProfitRate: 0,
        stopLossRate: 0,
        marginType: 'CROSSED',
        icebergLevels: 5,
        icebergQuantities: [0.35, 0.25, 0.2, 0.1, 0.1],
        icebergPriceGaps: [0, -1, -2, -3, -4],
      };
      this.isAutoGeneratedName = false;
      this.icebergSumError = false;
    },

// 策略类型改变
    onStrategyTypeChange() {
      if (this.strategyForm.strategyType === 'iceberg' || this.strategyForm.strategyType === 'slow_iceberg') {
        // 切换到冰山策略时，确保有正确的默认值
        this.updateIcebergDefaults();
      }
    },

// 方向改变时更新冰山策略默认值
    onSideChange() {
      this.generateStrategyName();

// 如果是冰山策略，更新价格间隔的默认值
      if (this.strategyForm.strategyType === 'iceberg' || this.strategyForm.strategyType === 'slow_iceberg') {
        this.updateIcebergDefaults();
      }
    },

// 更新冰山策略默认值
    updateIcebergDefaults() {
      const levels = this.strategyForm.icebergLevels;
      const side = this.strategyForm.side;

// 根据层数和方向设置默认值（万分比）
      const defaultConfigs = {
        2: {
          quantities: [0.6, 0.4],
          gapsLong: [0, -3],
          gapsShort: [0, 3]
        },
        3: {
          quantities: [0.5, 0.3, 0.2],
          gapsLong: [0, -2, -5],
          gapsShort: [0, 2 ,5]
        },
        4: {
          quantities: [0.4, 0.3, 0.2, 0.1],
          gapsLong: [0, -1, -3, -5],
          gapsShort: [0, 1, 3, 5]
        },
        5: {
          quantities: [0.3, 0.25, 0.2, 0.15, 0.1],
          gapsLong: [0, -1, -2, -3 ,-4],
          gapsShort: [0, 1, 2, 3, 4]
        },
        6: {
          quantities: [0.25, 0.25, 0.2, 0.1, 0.1, 0.1],
          gapsLong: [0, -1, -2, -3, -4, -5],
          gapsShort: [0, 1, 2, 3, 4, 5]
        }
      };

      if (defaultConfigs[levels]) {
        this.strategyForm.icebergQuantities = [...defaultConfigs[levels].quantities];

// 根据方向选择价格间隔
        if (side === 'LONG') {
          this.strategyForm.icebergPriceGaps = [...defaultConfigs[levels].gapsLong];
        } else if (side === 'SHORT') {
          this.strategyForm.icebergPriceGaps = [...defaultConfigs[levels].gapsShort];
        }
      }

// 验证数量总和
      this.validateIcebergSum();
    },

// 验证冰山策略数量总和
    validateIcebergSum() {
      const sum = this.strategyForm.icebergQuantities
          .slice(0, this.strategyForm.icebergLevels)
          .reduce((a, b) => a + b, 0);
      this.icebergSumError = Math.abs(sum - 1) > 0.001;
    },

// 计算预估开仓价格
    calculateEstimatedEntryPrice() {
      const { basePrice, entryPriceFloat, side } = this.strategyForm;
      if (!basePrice) return '-';

      let estimatedPrice = basePrice;

      // 如果设置了浮动
      if (entryPriceFloat > 0) {
        if (side === 'LONG') {
          // 做多时，开仓价格会低于触发价格
          estimatedPrice = basePrice * (1 - entryPriceFloat / 10000);
        } else if (side === 'SHORT') {
          // 做空时，开仓价格会高于触发价格
          estimatedPrice = basePrice * (1 + entryPriceFloat / 10000);
        }
      }

      return this.formatPrice(estimatedPrice);
    },

// 计算冰山策略每层价格
    calculateIcebergLayerPrice(index) {
      const { basePrice, icebergPriceGaps, side, entryPriceFloat } = this.strategyForm;
      if (!basePrice || !icebergPriceGaps[index] === undefined) return '-';

      // 计算基准价格（考虑浮动）
      let adjustedBasePrice = basePrice;
      if (index === 0 && entryPriceFloat > 0) {
        // 第一层应用开仓价格浮动
        if (side === 'LONG') {
          adjustedBasePrice = basePrice * (1 - entryPriceFloat / 10000);
        } else if (side === 'SHORT') {
          adjustedBasePrice = basePrice * (1 + entryPriceFloat / 10000);
        }
      }

      const layerPrice = adjustedBasePrice * (1 + icebergPriceGaps[index] / 10000); // 万分比
      return this.formatPrice(layerPrice);
    },

// 格式化冰山策略显示
    formatIcebergQuantities(quantitiesStr) {
      if (!quantitiesStr) return '-';
      const quantities = quantitiesStr.split(',').map(q => parseFloat(q.trim()));
      return quantities.map(q => `${(q * 100).toFixed(0)}%`).join(', ');
    },
// 格式化冰山价格间隔显示（添加对慢冰山的特殊处理）
    formatIcebergPriceGaps(gapsStr, strategyType) {
      if (!gapsStr) return '-';
      const gaps = gapsStr.split(',').map(g => parseFloat(g.trim()));
      const formatted = gaps.map(g => `${g > 0 ? '+' : ''}${g}‱`).join(', ');

      // 如果是慢冰山策略，添加特殊标记
      if (strategyType === 'slow_iceberg') {
        return formatted + ' (动态)';
      }
      return formatted;
    },

// 自动生成策略名称
    generateStrategyName() {
      const { basePrice, side, takeProfitRate } = this.strategyForm;

// 如果用户已经手动输入了名称，不覆盖
      if (this.strategyForm.strategyName && !this.isAutoGeneratedName) {
        return;
      }

      if (basePrice && side && takeProfitRate) {
        const takeProfitPrice = this.calculateTakeProfitPrice();
        const sideText = side === 'LONG' ? '做多' : '做空';

// 格式化价格，去掉小数点后多余的0
        const formattedBasePrice = parseFloat(basePrice).toString();
        const formattedTPPrice = takeProfitPrice !== '-' ?
            parseFloat(takeProfitPrice).toString() : '';

        if (formattedTPPrice) {
          this.strategyForm.strategyName = `${formattedBasePrice}${sideText}${formattedTPPrice}平仓`;
          this.isAutoGeneratedName = true;
        }
      }
    },

// 计算预估合约数量（基于预估开仓价格）
    calculateEstimatedContractQuantity() {
      const { quantity, basePrice, leverage, side, entryPriceFloat } = this.strategyForm;
      if (!quantity || !basePrice) return '0';

      // 计算预估开仓价格
      let estimatedEntryPrice = basePrice;
      if (entryPriceFloat > 0) {
        if (side === 'LONG') {
          estimatedEntryPrice = basePrice * (1 - entryPriceFloat / 10000);
        } else if (side === 'SHORT') {
          estimatedEntryPrice = basePrice * (1 + entryPriceFloat / 10000);
        }
      }

      // 使用本金乘以杠杆计算实际开仓价值，再除以预估开仓价格得到合约数量
      const totalValue = quantity * (leverage || 1);
      return (totalValue / estimatedEntryPrice).toFixed(8).replace(/\.?0+$/, '');
    },

// 获取合约单位
    getContractUnit() {
      const { symbol } = this.strategyForm;
      if (!symbol) return '';
      return symbol.replace('USDT', '');
    },

// 计算开仓手续费
    calculateOpenFee() {
      const { quantity, leverage } = this.strategyForm;
      const totalValue = quantity * (leverage || 1); // 实际开仓价值
      return totalValue * 0.0004; // 0.04%
    },

// 计算平仓手续费
    calculateCloseFee() {
      const { quantity, takeProfitRate, side, leverage } = this.strategyForm;
      if (!quantity || !takeProfitRate) return 0;

      const totalValue = quantity * (leverage || 1);
      const profitRate = takeProfitRate / 10000; // 万分比转小数
      let closeValue;

      if (side === 'LONG') {
        closeValue = totalValue * (1 + profitRate);
      } else {
        closeValue = totalValue * (1 - profitRate);
      }

      return closeValue * 0.0004; // 0.04%
    },

// 计算总手续费
    calculateTotalFee() {
      return this.calculateOpenFee() + this.calculateCloseFee();
    },

// 计算净收益
    calculateNetProfit() {
      const { quantity, takeProfitRate, leverage } = this.strategyForm;
      if (!quantity || !takeProfitRate) return 0;

      // 止盈率是基于本金的收益率（扣除手续费后）
      const netProfitRate = takeProfitRate / 10000; // 万分比转小数

      // 净收益 = 本金 × 净收益率
      const netProfit = quantity * netProfitRate;

      return netProfit;
    },

// 计算毛利润（用于显示）
    calculateGrossProfit() {
      const { quantity, takeProfitRate, leverage } = this.strategyForm;
      if (!quantity || !takeProfitRate) return 0;

      // 净收益
      const netProfit = this.calculateNetProfit();

      // 毛利润 = 净收益 + 手续费
      return netProfit + this.calculateTotalFee();
    },

// 计算实际的价格变动率
    calculateActualPriceChangeRate() {
      const { quantity, takeProfitRate, leverage } = this.strategyForm;
      if (!quantity || !takeProfitRate || !leverage) return 0;

      // 毛利润
      const grossProfit = this.calculateGrossProfit();

      // 价格变动率 = 毛利润 / 开仓价值
      const totalValue = quantity * leverage;
      return (grossProfit / totalValue) * 10000; // 转换为万分比
    },

// 计算净收益率
    calculateNetProfitRate() {
      const { quantity } = this.strategyForm;
      if (!quantity) return '0.00';

      const netProfit = this.calculateNetProfit();
      const rate = (netProfit / quantity) * 100;
      return rate.toFixed(2);
    },

    calculateTakeProfitPrice() {
      const { basePrice, takeProfitRate, side, entryPriceFloat } = this.strategyForm;
      if (!basePrice || !takeProfitRate) return '-';

      // 使用预估的开仓价格来计算止盈价
      let entryPrice = basePrice;
      if (entryPriceFloat > 0) {
        if (side === 'LONG') {
          entryPrice = basePrice * (1 - entryPriceFloat / 10000);
        } else if (side === 'SHORT') {
          entryPrice = basePrice * (1 + entryPriceFloat / 10000);
        }
      }

      const feeRate = 0.0004 * 2; // 开仓+平仓手续费
      const profitRate = takeProfitRate / 10000; // 万分比转小数

      // 基于预估开仓价格计算
      if (side === 'LONG') {
        return this.formatPrice(entryPrice * (1 + profitRate + feeRate));
      } else {
        return this.formatPrice(entryPrice * (1 - profitRate - feeRate));
      }
    },

    calculateStopLossPrice() {
      const { basePrice, stopLossRate, side, entryPriceFloat } = this.strategyForm;
      if (!basePrice || !stopLossRate) return '-';

      // 使用预估的开仓价格来计算止损价
      let entryPrice = basePrice;
      if (entryPriceFloat > 0) {
        if (side === 'LONG') {
          entryPrice = basePrice * (1 - entryPriceFloat / 10000);
        } else if (side === 'SHORT') {
          entryPrice = basePrice * (1 + entryPriceFloat / 10000);
        }
      }

      const lossRate = stopLossRate / 10000; // 万分比转小数

      // 基于预估开仓价格计算
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
/* CSS 变量定义 */
:root {
  /* 颜色系统 */
  --color-primary: #2563eb;
  --color-primary-hover: #1d4ed8;
  --color-secondary: #64748b;
  --color-success: #10b981;
  --color-warning: #f59e0b;
  --color-danger: #ef4444;

  /* 中性色 */
  --color-bg: #ffffff;
  --color-bg-secondary: #f8fafc;
  --color-bg-tertiary: #f1f5f9;
  --color-border: #e2e8f0;
  --color-text-primary: #0f172a;
  --color-text-secondary: #475569;
  --color-text-tertiary: #94a3b8;

  /* 间距系统 */
  --spacing-xs: 0.5rem;
  --spacing-sm: 0.75rem;
  --spacing-md: 1rem;
  --spacing-lg: 1.5rem;
  --spacing-xl: 2rem;
  --spacing-2xl: 3rem;

  /* 字体 */
  --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;

  /* 圆角 */
  --radius-sm: 0.25rem;
  --radius-md: 0.375rem;
  --radius-lg: 0.5rem;

  /* 过渡 */
  --transition-fast: 150ms ease;
  --transition-normal: 200ms ease;
}

/* 新增警告样式 */
.form-control.warning {
  border-color: var(--color-warning);
}

.form-hint.hint-warning {
  color: var(--color-warning);
}

.warning-text {
  font-weight: 600;
  color: var(--color-warning);
}

.toast.warning {
  border-color: var(--color-warning);
  color: var(--color-warning);
}

/* 策略类型徽章 */
.type-badge {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
  background: #f3e8ff;
  color: #7c3aed;
}

.type-badge.slow {
  background: #fef3c7;
  color: #92400e;
}

/* 慢冰山提示 */
.slow-iceberg-hint {
  display: block;
  margin-top: 0.25rem;
  color: var(--color-warning);
  font-weight: 500;
}

/* 冰山策略详情 */
.iceberg-details {
  margin-top: 1rem;
  padding: 1rem;
  background: var(--color-bg-secondary);
  border-radius: var(--radius-md);
}

.iceberg-title {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: 0.75rem;
}

.iceberg-info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 0.75rem;
}

.iceberg-info-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.info-label {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

.info-value {
  font-size: 0.875rem;
  color: var(--color-text-primary);
  font-weight: 500;
}

/* 冰山策略配置部分 */
.iceberg-config-section {
  margin-top: 1.5rem;
  padding: 1.5rem;
  background: var(--color-bg-secondary);
  border-radius: var(--radius-lg);
}

.config-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 1rem 0;
}

.iceberg-layers {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-top: 0.5rem;
}

.iceberg-layer {
  display: grid;
  grid-template-columns: 80px 1fr 60px;
  gap: 0.75rem;
  align-items: center;
}

.layer-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  font-weight: 500;
}

.layer-percent {
  font-size: 0.875rem;
  color: var(--color-text-tertiary);
  text-align: right;
  font-weight: 500;
}

.form-error {
  color: var(--color-danger);
  font-size: 0.75rem;
  margin-top: 0.5rem;
}

/* 冰山策略预览 */
.iceberg-preview {
  margin-top: 1rem;
  padding: 1rem;
  background: var(--color-bg-tertiary);
  border-radius: var(--radius-md);
}
.iceberg-preview h5 {
  font-size: 0.875rem;
  font-weight: 600;
  margin: 0 0 0.75rem 0;
  color: var(--color-text-primary);
}

.iceberg-preview-layers {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.iceberg-preview-layer {
  display: grid;
  grid-template-columns: 60px 1fr 1fr;
  gap: 0.75rem;
  padding: 0.5rem;
  background: var(--color-bg);
  border-radius: var(--radius-sm);
  font-size: 0.75rem;
}

.preview-layer-label {
  font-weight: 500;
  color: var(--color-text-secondary);
}

.preview-layer-info,
.preview-layer-price {
  color: var(--color-text-primary);
}

.price-diff {
  color: var(--color-text-tertiary);
  font-size: 0.7rem;
}

/* 计算后价格提示 */
.calculated-price-hint {
  font-size: 0.75rem;
  color: var(--color-primary);
  margin-top: 0.25rem;
  font-weight: 500;
}

.price-example {
  color: var(--color-text-primary);
  font-weight: 600;
  margin-left: 0.5rem;
}

.form-control.error {
  border-color: var(--color-danger);
}

.form-hint.hint-error {
  color: var(--color-danger);
}

.error-text {
  font-weight: 600;
}

/* 预览项目全宽 */
.preview-item.full-width {
  grid-column: 1 / -1;
  padding-top: 0.5rem;
  border-top: 1px solid var(--color-border);
  margin-top: 0.5rem;
}

/* 其他样式保持原样... */
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

/* 表单提示 */
.form-hint {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
  margin-top: 0.25rem;
}

/* 杠杆选择样式 */
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
  max-width: 700px;
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

.toast.warning {
  border-color: var(--color-warning);
  color: var(--color-warning);
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

  .iceberg-layer {
    grid-template-columns: 60px 1fr 50px;
    gap: 0.5rem;
  }

  .iceberg-preview-layer {
    grid-template-columns: 50px 1fr;
    gap: 0.5rem;
  }

  .preview-layer-price {
    grid-column: 2;
    margin-top: 0.25rem;
  }
}
</style>
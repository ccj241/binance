<template>
  <div class="settings-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">系统设置</h1>
      <p class="page-subtitle">管理您的API密钥和自动提币规则</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">
          <span>🔑</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ apiKey ? '已配置' : '未配置' }}</div>
          <div class="stat-label">API 密钥</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">
          <span>🔒</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ secretKey ? '已配置' : '未配置' }}</div>
          <div class="stat-label">Secret 密钥</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">
          <span>⚡</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ withdrawalRules.length }}</div>
          <div class="stat-label">提币规则</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon">
          <span>✅</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ enabledRulesCount }}</div>
          <div class="stat-label">启用规则</div>
        </div>
      </div>
    </div>

    <!-- 消息提示 -->
    <transition name="toast">
      <div v-if="toastMessage" :class="['toast', toastType]">
        <span class="toast-icon">{{ toastType === 'success' ? '✓' : '×' }}</span>
        <span>{{ toastMessage }}</span>
      </div>
    </transition>

    <!-- API 密钥管理 -->
    <div class="settings-section">
      <div class="section-header">
        <h2 class="section-title">
          <span class="section-icon">🔑</span>
          API 密钥管理
        </h2>
        <button @click="toggleApiSection" class="toggle-btn">
          <span>{{ showApiSection ? '收起' : '展开' }}</span>
        </button>
      </div>

      <transition name="section-slide">
        <div v-if="showApiSection" class="section-content">
          <!-- 当前密钥状态 -->
          <div v-if="apiKey || secretKey" class="current-keys">
            <div class="key-display">
              <div class="key-card">
                <div class="key-header">
                  <div class="key-icon">🔑</div>
                  <div class="key-info">
                    <h4>API Key</h4>
                    <p class="key-status">已配置</p>
                  </div>
                </div>
                <div class="key-value">
                  <span class="masked-key">{{ maskKey(apiKey) }}</span>
                  <button @click="toggleKeyVisibility('api')" class="visibility-btn">
                    <span>{{ showApiKey ? '隐藏' : '显示' }}</span>
                  </button>
                </div>
                <div v-if="showApiKey" class="full-key">{{ apiKey }}</div>
              </div>

              <div class="key-card">
                <div class="key-header">
                  <div class="key-icon">🔒</div>
                  <div class="key-info">
                    <h4>Secret Key</h4>
                    <p class="key-status">已配置</p>
                  </div>
                </div>
                <div class="key-value">
                  <span class="masked-key">{{ maskKey(secretKey) }}</span>
                  <button @click="toggleKeyVisibility('secret')" class="visibility-btn">
                    <span>{{ showSecretKey ? '隐藏' : '显示' }}</span>
                  </button>
                </div>
                <div v-if="showSecretKey" class="full-key">{{ secretKey }}</div>
              </div>
            </div>

            <button @click="deleteAPIKey" class="action-btn delete">
              删除 API 密钥
            </button>
          </div>

          <div v-else class="no-keys">
            <div class="no-keys-icon">🔑</div>
            <p class="no-keys-text">尚未配置 API 密钥</p>
            <p class="no-keys-subtitle">请添加您的 Binance API 密钥以开始使用</p>
          </div>

          <!-- 添加新密钥 -->
          <div class="add-keys-section">
            <h3 class="subsection-title">添加新的 API 密钥</h3>

            <form @submit.prevent="saveAPIKey" class="key-form">
              <div class="form-grid">
                <div class="form-group">
                  <label>API Key</label>
                  <input
                      v-model="newAPIKey"
                      type="text"
                      placeholder="请输入您的 Binance API Key"
                      required
                  />
                </div>

                <div class="form-group">
                  <label>Secret Key</label>
                  <div class="password-input">
                    <input
                        v-model="newSecretKey"
                        :type="showNewSecretInput ? 'text' : 'password'"
                        placeholder="请输入您的 Secret Key"
                        required
                    />
                    <button
                        type="button"
                        @click="showNewSecretInput = !showNewSecretInput"
                        class="password-toggle"
                    >
                      {{ showNewSecretInput ? '隐藏' : '显示' }}
                    </button>
                  </div>
                </div>
              </div>

              <div class="form-actions">
                <button type="submit" class="action-btn primary">
                  保存 API 密钥
                </button>
                <button type="button" @click="resetApiForm" class="action-btn secondary">
                  重置表单
                </button>
              </div>
            </form>
          </div>
        </div>
      </transition>
    </div>

    <!-- 自动提币设置 -->
    <div class="settings-section">
      <div class="section-header">
        <h2 class="section-title">
          <span class="section-icon">⚡</span>
          自动提币设置
        </h2>
        <button @click="toggleWithdrawalSection" class="toggle-btn">
          <span>{{ showWithdrawalSection ? '收起' : '展开' }}</span>
        </button>
      </div>

      <transition name="section-slide">
        <div v-if="showWithdrawalSection" class="section-content">
          <!-- 添加提币规则 -->
          <div class="add-rule-section">
            <h3 class="subsection-title">添加提币规则</h3>

            <form @submit.prevent="createWithdrawalRule" class="rule-form">
              <div class="form-grid">
                <div class="form-group">
                  <label>币种</label>
                  <select v-model="newWithdrawal.asset" @change="onAssetChange" required>
                    <option value="">请选择币种</option>
                    <option v-for="asset in supportedAssets" :key="asset" :value="asset">
                      {{ asset }}
                    </option>
                  </select>
                </div>

                <div class="form-group">
                  <label>网络</label>
                  <select v-model="newWithdrawal.network" @change="onNetworkChange" required :disabled="!newWithdrawal.asset">
                    <option value="">{{ newWithdrawal.asset ? '请选择网络' : '请先选择币种' }}</option>
                    <option v-for="network in availableNetworks" :key="network.value" :value="network.value">
                      {{ network.label }} {{ network.fee ? `(手续费: ${network.fee})` : '' }}
                    </option>
                  </select>
                  <small v-if="selectedNetworkInfo" class="form-hint network-info">
                    最小提币金额: {{ selectedNetworkInfo.minAmount }}，网络手续费: {{ selectedNetworkInfo.fee }}
                  </small>
                </div>

                <div class="form-group">
                  <label>触发阈值</label>
                  <input
                      v-model.number="newWithdrawal.threshold"
                      type="number"
                      step="0.00000001"
                      placeholder="余额超过此数量时触发"
                      required
                  />
                  <small v-if="selectedNetworkInfo" class="form-hint">
                    建议设置大于最小提币金额 {{ selectedNetworkInfo.minAmount }}
                  </small>
                </div>

                <div class="form-group">
                  <label>提币金额</label>
                  <input
                      v-model.number="newWithdrawal.amount"
                      type="number"
                      step="0.00000001"
                      min="0"
                      placeholder="每次提币数量（0表示提取最大可用金额）"
                      required
                  />
                  <small class="form-hint">设置为0将自动提取所有可用余额</small>
                </div>

                <div class="form-group form-group-wide">
                  <label>提币地址</label>
                  <input
                      v-model="newWithdrawal.address"
                      type="text"
                      placeholder="目标钱包地址"
                      required
                  />
                  <small v-if="newWithdrawal.network" class="form-hint">
                    请确保地址与所选网络 ({{ getNetworkDisplayName(newWithdrawal.network) }}) 兼容
                  </small>
                </div>
              </div>

              <!-- 规则说明 -->
              <div v-if="newWithdrawal.asset && newWithdrawal.network" class="rule-description">
                <div class="description-card">
                  <div class="description-icon">💡</div>
                  <div class="description-content">
                    <h4>自动提币规则预览</h4>
                    <p>
                      当您的 <strong>{{ newWithdrawal.asset }}</strong> 余额超过
                      <strong>{{ newWithdrawal.threshold || '[阈值]' }}</strong> 时，
                      系统将通过 <strong>{{ getNetworkDisplayName(newWithdrawal.network) }}</strong> 网络自动提取
                      <strong>{{ newWithdrawal.amount > 0 ? formatNumber(newWithdrawal.amount) : '最大可用金额' }}</strong> 到指定地址。
                    </p>
                    <div v-if="selectedNetworkInfo" class="fee-info">
                      <small>⚠️ 网络手续费: {{ selectedNetworkInfo.fee }}，实际到账金额会扣除手续费</small>
                    </div>
                    <div class="warning-info">
                      <small>⚠️ 提示：请确保提币地址正确，提币操作无法撤回</small>
                    </div>
                  </div>
                </div>
              </div>

              <div class="form-actions">
                <button type="submit" class="action-btn primary" :disabled="!isFormValid">
                  创建规则
                </button>
                <button type="button" @click="resetWithdrawalForm" class="action-btn secondary">
                  重置表单
                </button>
              </div>
            </form>
          </div>

          <!-- 现有规则列表 -->
          <div class="rules-list">
            <h3 class="subsection-title">现有规则 ({{ withdrawalRules.length }})</h3>

            <div v-if="withdrawalRules.length === 0" class="empty-state">
              <div class="empty-icon">⚡</div>
              <p class="empty-text">暂无提币规则</p>
              <p class="empty-subtitle">添加第一个自动提币规则以开始使用</p>
            </div>

            <div v-else class="rules-grid">
              <div v-for="rule in withdrawalRules" :key="rule.id" class="rule-card">
                <div class="rule-header">
                  <div class="rule-asset">
                    <div class="asset-icon">🪙</div>
                    <div class="asset-info">
                      <h4>{{ rule.asset }}</h4>
                      <div class="asset-meta">
                        <span class="network-chip">{{ getNetworkDisplayName(rule.network) }}</span>
                        <span :class="['status-chip', rule.enabled ? 'enabled' : 'disabled']">
                          {{ rule.enabled ? '启用' : '禁用' }}
                        </span>
                      </div>
                    </div>
                  </div>
                  <div class="rule-id">ID: {{ rule.id }}</div>
                </div>

                <div class="rule-details">
                  <div class="detail-item">
                    <span class="detail-label">触发阈值</span>
                    <span class="detail-value">{{ formatNumber(rule.threshold) }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">提币金额</span>
                    <span class="detail-value">{{ rule.amount > 0 ? formatNumber(rule.amount) : '最大可用' }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">网络</span>
                    <span class="detail-value network">{{ getNetworkDisplayName(rule.network) }}</span>
                  </div>
                  <div class="detail-item">
                    <span class="detail-label">提币地址</span>
                    <span class="detail-value address">{{ formatAddress(rule.address) }}</span>
                  </div>
                </div>

                <div class="rule-actions">
                  <button @click="toggleRuleStatus(rule)" class="action-btn toggle">
                    {{ rule.enabled ? '禁用' : '启用' }}
                  </button>
                  <button @click="deleteWithdrawalRule(rule.id)" class="action-btn delete">
                    删除
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  name: 'Settings',
  data() {
    return {
      apiKey: '',
      secretKey: '',
      newAPIKey: '',
      newSecretKey: '',
      showApiKey: false,
      showSecretKey: false,
      showNewSecretInput: false,
      showApiSection: true,
      showWithdrawalSection: true,
      newWithdrawal: {
        asset: '',
        network: '',
        threshold: 0,
        amount: 0,
        address: '',
      },
      withdrawalRules: [],
      toastMessage: '',
      toastType: 'success',

      // 网络相关数据
      supportedAssets: [
        'BTC', 'ETH', 'USDT', 'USDC', 'BNB', 'ADA', 'DOT', 'SOL', 'MATIC', 'AVAX',
        'TRX', 'LTC', 'BCH', 'XRP', 'DOGE', 'SHIB', 'UNI', 'LINK', 'ATOM', 'FTM',
        'NEAR', 'ALGO', 'VET', 'ICP', 'THETA', 'FIL', 'XTZ', 'EOS', 'AAVE', 'MKR',
        'COMP', 'YFI', 'SNX', 'CRV', 'SUSHI', '1INCH', 'BAT', 'ZRX', 'ENJ', 'MANA',
        'SAND', 'AXS', 'GALA', 'CHZ'
      ],

      // 币种网络映射
      assetNetworks: {
        'BTC': [
          { value: 'BTC', label: 'BTC (原生网络)', fee: '0.0005 BTC', minAmount: '0.001 BTC' },
          { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.0000035 BTC', minAmount: '0.0001 BTC' }
        ],
        'ETH': [
          { value: 'ERC20', label: 'ERC20 (以太坊)', fee: '0.005 ETH', minAmount: '0.01 ETH' },
          { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.0002 ETH', minAmount: '0.001 ETH' },
          { value: 'ARBITRUM', label: 'Arbitrum One', fee: '0.0001 ETH', minAmount: '0.001 ETH' },
          { value: 'POLYGON', label: 'Polygon', fee: '0.0001 ETH', minAmount: '0.001 ETH' }
        ],
        'USDT': [
          { value: 'ERC20', label: 'ERC20 (以太坊)', fee: '25 USDT', minAmount: '10 USDT' },
          { value: 'TRC20', label: 'TRC20 (TRON)', fee: '1 USDT', minAmount: '1 USDT' },
          { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.8 USDT', minAmount: '1 USDT' },
          { value: 'POLYGON', label: 'Polygon', fee: '0.8 USDT', minAmount: '1 USDT' },
          { value: 'ARBITRUM', label: 'Arbitrum One', fee: '0.8 USDT', minAmount: '1 USDT' },
          { value: 'OPTIMISM', label: 'Optimism', fee: '0.8 USDT', minAmount: '1 USDT' }
        ],
        'USDC': [
          { value: 'ERC20', label: 'ERC20 (以太坊)', fee: '25 USDC', minAmount: '10 USDC' },
          { value: 'TRC20', label: 'TRC20 (TRON)', fee: '1 USDC', minAmount: '1 USDC' },
          { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.8 USDC', minAmount: '1 USDC' },
          { value: 'POLYGON', label: 'Polygon', fee: '0.8 USDC', minAmount: '1 USDC' },
          { value: 'ARBITRUM', label: 'Arbitrum One', fee: '0.1 USDC', minAmount: '1 USDC' }
        ],
        'BNB': [
          { value: 'BEP20', label: 'BEP20 (BSC)', fee: '0.005 BNB', minAmount: '0.01 BNB' },
          { value: 'BEP2', label: 'BEP2 (币安链)', fee: '0.00075 BNB', minAmount: '0.01 BNB' }
        ],
        'ADA': [
          { value: 'ADA', label: 'Cardano', fee: '1 ADA', minAmount: '1 ADA' }
        ],
        'DOT': [
          { value: 'DOT', label: 'Polkadot', fee: '0.1 DOT', minAmount: '1 DOT' }
        ],
        'SOL': [
          { value: 'SOL', label: 'Solana', fee: '0.01 SOL', minAmount: '0.01 SOL' }
        ],
        'MATIC': [
          { value: 'POLYGON', label: 'Polygon', fee: '0.01 MATIC', minAmount: '0.1 MATIC' },
          { value: 'ERC20', label: 'ERC20 (以太坊)', fee: '15 MATIC', minAmount: '10 MATIC' }
        ],
        'AVAX': [
          { value: 'AVAXC', label: 'Avalanche C-Chain', fee: '0.005 AVAX', minAmount: '0.01 AVAX' }
        ],
        'TRX': [
          { value: 'TRC20', label: 'TRON', fee: '1 TRX', minAmount: '1 TRX' }
        ]
      }
    };
  },
  computed: {
    enabledRulesCount() {
      return this.withdrawalRules.filter(rule => rule.enabled).length;
    },

    availableNetworks() {
      if (!this.newWithdrawal.asset) return [];
      return this.assetNetworks[this.newWithdrawal.asset] || [];
    },

    selectedNetworkInfo() {
      if (!this.newWithdrawal.asset || !this.newWithdrawal.network) return null;
      const networks = this.assetNetworks[this.newWithdrawal.asset] || [];
      return networks.find(network => network.value === this.newWithdrawal.network);
    },

    isFormValid() {
      return this.newWithdrawal.asset &&
          this.newWithdrawal.network &&
          this.newWithdrawal.threshold > 0 &&
          this.newWithdrawal.amount >= 0 &&
          this.newWithdrawal.address.trim();
    }
  },
  async mounted() {
    await this.fetchAPIKey();
    await this.fetchWithdrawalRules();
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

    toggleApiSection() {
      this.showApiSection = !this.showApiSection;
    },

    toggleWithdrawalSection() {
      this.showWithdrawalSection = !this.showWithdrawalSection;
    },

    toggleKeyVisibility(type) {
      if (type === 'api') {
        this.showApiKey = !this.showApiKey;
      } else {
        this.showSecretKey = !this.showSecretKey;
      }
    },

    maskKey(key) {
      if (!key) return '';
      if (key.length <= 8) return '***';
      return key.substring(0, 4) + '***' + key.substring(key.length - 4);
    },

    formatNumber(num) {
      if (!num) return '0';
      return parseFloat(num).toFixed(8).replace(/\.?0+$/, '');
    },

    formatAddress(address) {
      if (!address) return '';
      if (address.length <= 16) return address;
      return address.substring(0, 8) + '...' + address.substring(address.length - 8);
    },

    getNetworkDisplayName(networkValue) {
      // 查找所有币种的网络配置，找到对应的显示名称
      for (const assetNetworks of Object.values(this.assetNetworks)) {
        const network = assetNetworks.find(n => n.value === networkValue);
        if (network) {
          return network.label;
        }
      }
      return networkValue;
    },

    onAssetChange() {
      // 重置网络选择
      this.newWithdrawal.network = '';
    },

    onNetworkChange() {
      // 当网络改变时，可以添加额外的验证或提示
      if (this.selectedNetworkInfo) {
        console.log('选择的网络信息:', this.selectedNetworkInfo);
      }
    },

    resetApiForm() {
      this.newAPIKey = '';
      this.newSecretKey = '';
      this.showNewSecretInput = false;
    },

    resetWithdrawalForm() {
      this.newWithdrawal = {
        asset: '',
        network: '',
        threshold: 0,
        amount: 0,
        address: '',
      };
    },

    async fetchAPIKey() {
      try {
        const response = await axios.get('/api-key', {
          headers: this.getAuthHeaders(),
        });
        this.apiKey = response.data.apiKey || '';
        this.secretKey = response.data.secretKey || '';
      } catch (err) {
        console.error('fetchAPIKey error:', err);
        this.showToast(err.response?.data?.error || '获取 API 密钥失败', 'error');
      }
    },

    async saveAPIKey() {
      if (!this.newAPIKey.trim() || !this.newSecretKey.trim()) {
        this.showToast('请填写完整的 API 密钥信息', 'error');
        return;
      }

      // 验证API密钥格式（币安API密钥通常是64个字符）
      if (this.newAPIKey.length !== 64) {
        this.showToast('API Key 长度应为 64 个字符', 'error');
        return;
      }

      if (this.newSecretKey.length !== 64) {
        this.showToast('Secret Key 长度应为 64 个字符', 'error');
        return;
      }

      try {
        console.log('保存API密钥请求:', {
          apiKeyLength: this.newAPIKey.length,
          secretKeyLength: this.newSecretKey.length
        });

        const response = await axios.post(
            '/api-key',
            {
              apiKey: this.newAPIKey,
              apiSecret: this.newSecretKey,
            },
            {
              headers: this.getAuthHeaders(),
            }
        );

        console.log('保存API密钥响应:', response.data);

        this.showToast(response.data.message || 'API 密钥保存成功');
        this.resetApiForm();

        // 等待一下再获取，确保数据库操作完成
        setTimeout(async () => {
          await this.fetchAPIKey();

          // 验证是否真的保存成功
          if (!this.apiKey || !this.secretKey) {
            this.showToast('API密钥保存可能失败，请检查', 'error');
          }
        }, 500);

      } catch (err) {
        console.error('saveAPIKey error:', err);
        const errorMsg = err.response?.data?.details || err.response?.data?.error || '保存 API 密钥失败';
        this.showToast(errorMsg, 'error');
      }
    },

    async deleteAPIKey() {
      if (!window.confirm('确定要删除 API 密钥吗？删除后将无法进行交易操作。')) {
        return;
      }

      try {
        const response = await axios.delete('/api-key/delete', {
          headers: this.getAuthHeaders(),
        });
        this.showToast(response.data.message || 'API 密钥删除成功');
        this.apiKey = '';
        this.secretKey = '';
        this.showApiKey = false;
        this.showSecretKey = false;
      } catch (err) {
        console.error('deleteAPIKey error:', err);
        this.showToast(err.response?.data?.error || '删除 API 密钥失败', 'error');
      }
    },

    async createWithdrawalRule() {
      const { asset, network, threshold, amount, address } = this.newWithdrawal;

      if (!asset.trim() || !network.trim() || threshold <= 0 || amount < 0 || !address.trim()) {
        this.showToast('请填写所有必需字段，阈值必须大于0，金额不能为负数', 'error');
        return;
      }

      // 验证网络兼容性
      const availableNetworks = this.assetNetworks[asset] || [];
      const isValidNetwork = availableNetworks.some(n => n.value === network);
      if (!isValidNetwork) {
        this.showToast(`币种 ${asset} 不支持 ${network} 网络`, 'error');
        return;
      }

      // 验证最小提币金额
      if (this.selectedNetworkInfo && amount > 0) {
        const minAmount = parseFloat(this.selectedNetworkInfo.minAmount.split(' ')[0]);
        if (amount < minAmount) {
          this.showToast(`提币金额不能小于最小提币金额 ${this.selectedNetworkInfo.minAmount}`, 'error');
          return;
        }
      }

      try {
        const response = await axios.post(
            '/withdrawals',
            {
              asset: asset.toUpperCase(),
              network: network,
              threshold: Number(threshold),
              amount: Number(amount),
              address: address,
              enabled: true,
            },
            {
              headers: this.getAuthHeaders(),
            }
        );
        this.showToast(response.data.message || '自动提币规则创建成功');
        this.resetWithdrawalForm();
        await this.fetchWithdrawalRules();
      } catch (err) {
        console.error('createWithdrawalRule error:', err);
        this.showToast(err.response?.data?.error || '创建提币规则失败', 'error');
      }
    },

    async fetchWithdrawalRules() {
      try {
        const response = await axios.get('/withdrawals', {
          headers: this.getAuthHeaders(),
        });
        this.withdrawalRules = response.data.rules || [];
      } catch (err) {
        console.error('fetchWithdrawalRules error:', err);
        this.showToast(err.response?.data?.error || '获取提币规则失败', 'error');
      }
    },

    async toggleRuleStatus(rule) {
      try {
        const response = await axios.put(
            `/withdrawals/${rule.id}`,
            {
              ...rule,
              enabled: !rule.enabled,
            },
            {
              headers: this.getAuthHeaders(),
            }
        );
        this.showToast(response.data.message || `规则已${!rule.enabled ? '启用' : '禁用'}`);
        await this.fetchWithdrawalRules();
      } catch (err) {
        console.error('toggleRuleStatus error:', err);
        this.showToast(err.response?.data?.error || '更新规则状态失败', 'error');
      }
    },

    async deleteWithdrawalRule(ruleId) {
      if (!window.confirm(`确定要删除提币规则 ID ${ruleId} 吗？`)) {
        return;
      }

      try {
        const response = await axios.delete(`/withdrawals/${ruleId}`, {
          headers: this.getAuthHeaders(),
        });
        this.showToast(response.data.message || '提币规则删除成功');
        await this.fetchWithdrawalRules();
      } catch (err) {
        console.error('deleteWithdrawalRule error:', err);
        this.showToast(err.response?.data?.error || '删除提币规则失败', 'error');
      }
    }
  },
};
</script>

<style scoped>
/* 容器 */
.settings-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0;
  background: #ffffff;
  min-height: 100vh;
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

.page-subtitle {
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
  display: flex;
  align-items: center;
  gap: 1rem;
}

.stat-icon {
  width: 48px;
  height: 48px;
  background: #f8fafc;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 0.25rem;
}

.stat-label {
  color: #64748b;
  font-size: 0.875rem;
}

/* Toast 消息 */
.toast {
  position: fixed;
  bottom: 2rem;
  right: 2rem;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1rem 1.5rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
  z-index: 1000;
}

.toast.success {
  border-color: #10b981;
}

.toast.error {
  border-color: #ef4444;
}

.toast-icon {
  font-size: 1.25rem;
}

.toast.success .toast-icon {
  color: #10b981;
}

.toast.error .toast-icon {
  color: #ef4444;
}

/* 动画 */
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

/* 设置区块 */
.settings-section {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  margin-bottom: 1.5rem;
  overflow: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1.125rem;
  font-weight: 600;
  color: #0f172a;
  margin: 0;
}

.section-icon {
  font-size: 1.25rem;
}

.toggle-btn {
  padding: 0.5rem 1rem;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  color: #475569;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.toggle-btn:hover {
  background: #f1f5f9;
  border-color: #cbd5e1;
}

.section-content {
  padding: 1.5rem;
}

.section-slide-enter-active,
.section-slide-leave-active {
  transition: all 0.3s ease;
}

.section-slide-enter-from {
  opacity: 0;
  transform: translateY(-10px);
}

.section-slide-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}

/* 密钥显示 */
.current-keys {
  margin-bottom: 2rem;
}

.key-display {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.key-card {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1.25rem;
}

.key-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.key-icon {
  width: 40px;
  height: 40px;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.key-info h4 {
  margin: 0 0 0.25rem 0;
  color: #0f172a;
  font-size: 1rem;
}

.key-status {
  margin: 0;
  color: #10b981;
  font-size: 0.875rem;
}

.key-value {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  padding: 0.75rem;
  border-radius: 0.375rem;
  margin-bottom: 0.5rem;
}

.masked-key {
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Mono', 'Courier New', monospace;
  color: #475569;
  font-size: 0.875rem;
}

.visibility-btn {
  padding: 0.25rem 0.75rem;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 0.25rem;
  color: #475569;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.visibility-btn:hover {
  background: #e2e8f0;
}

.full-key {
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Mono', 'Courier New', monospace;
  background: #f0fdf4;
  border: 1px solid #86efac;
  padding: 0.75rem;
  border-radius: 0.375rem;
  color: #16a34a;
  font-size: 0.8125rem;
  word-break: break-all;
}

/* 空状态 */
.no-keys,
.empty-state {
  text-align: center;
  padding: 3rem 2rem;
}

.no-keys-icon,
.empty-icon {
  font-size: 3rem;
  margin-bottom: 1rem;
  opacity: 0.3;
}

.no-keys-text,
.empty-text {
  color: #475569;
  font-size: 1.125rem;
  margin-bottom: 0.5rem;
}

.no-keys-subtitle,
.empty-subtitle {
  color: #94a3b8;
  font-size: 0.875rem;
}

/* 添加区域 */
.add-keys-section,
.add-rule-section {
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.subsection-title {
  font-size: 1rem;
  font-weight: 600;
  color: #0f172a;
  margin: 0 0 1.5rem 0;
}

/* 表单 */
.form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 1.5rem;
  margin-bottom: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.form-group.form-group-wide {
  grid-column: 1 / -1;
}

.form-group label {
  font-weight: 500;
  color: #475569;
  font-size: 0.875rem;
}

.form-group input,
.form-group select {
  padding: 0.625rem 0.875rem;
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  color: #0f172a;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.form-group input:focus,
.form-group select:focus {
  outline: none;
  border-color: #2563eb;
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
}

.form-group input::placeholder {
  color: #94a3b8;
}

.form-group select:disabled {
  background: #f8fafc;
  color: #94a3b8;
  cursor: not-allowed;
}

.password-input {
  position: relative;
}

.password-input input {
  width: 100%;
  padding-right: 4rem;
}

.password-toggle {
  position: absolute;
  right: 0.5rem;
  top: 50%;
  transform: translateY(-50%);
  padding: 0.25rem 0.75rem;
  background: #f1f5f9;
  border: 1px solid #e2e8f0;
  border-radius: 0.25rem;
  color: #475569;
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s;
}

.password-toggle:hover {
  background: #e2e8f0;
}

.form-hint {
  color: #94a3b8;
  font-size: 0.75rem;
  margin-top: 0.25rem;
}

.form-hint.network-info {
  color: #059669;
  font-weight: 500;
}

/* 规则说明 */
.rule-description {
  margin: 1.5rem 0;
}

.description-card {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  background: #fffbeb;
  border: 1px solid #fbbf24;
  border-radius: 0.5rem;
  padding: 1rem;
}

.description-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.description-content h4 {
  margin: 0 0 0.5rem 0;
  color: #0f172a;
  font-size: 0.875rem;
}

.description-content p {
  margin: 0 0 0.5rem 0;
  color: #475569;
  font-size: 0.875rem;
  line-height: 1.5;
}

.fee-info {
  margin: 0.5rem 0;
}

.fee-info small {
  color: #059669;
  font-size: 0.75rem;
}

.warning-info small {
  color: #92400e;
  font-size: 0.75rem;
}

/* 操作按钮 */
.form-actions {
  display: flex;
  gap: 0.75rem;
}

.action-btn {
  padding: 0.625rem 1.25rem;
  border: 1px solid transparent;
  border-radius: 0.375rem;
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
  flex: 1;
}

.action-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.action-btn.primary {
  background: #2563eb;
  color: white;
}

.action-btn.primary:hover:not(:disabled) {
  background: #1d4ed8;
}

.action-btn.secondary {
  background: #ffffff;
  color: #475569;
  border-color: #e2e8f0;
}

.action-btn.secondary:hover {
  background: #f8fafc;
  border-color: #cbd5e1;
}

.action-btn.delete {
  background: #ffffff;
  color: #ef4444;
  border-color: #fecaca;
}

.action-btn.delete:hover {
  background: #fef2f2;
  border-color: #fca5a5;
}

.action-btn.toggle {
  background: #ffffff;
  color: #3b82f6;
  border-color: #bfdbfe;
}

.action-btn.toggle:hover {
  background: #eff6ff;
  border-color: #93c5fd;
}

/* 规则列表 */
.rules-list {
  margin-top: 2rem;
}

.rules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1rem;
  margin-top: 1rem;
}

.rule-card {
  background: #ffffff;
  border: 1px solid #e2e8f0;
  border-radius: 0.5rem;
  padding: 1.5rem;
  transition: all 0.2s;
}

.rule-card:hover {
  border-color: #cbd5e1;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
}

.rule-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
  padding-bottom: 1rem;
  border-bottom: 1px solid #f1f5f9;
}

.rule-asset {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.asset-icon {
  width: 40px;
  height: 40px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
}

.asset-info h4 {
  margin: 0 0 0.5rem 0;
  color: #0f172a;
  font-size: 1.125rem;
  font-weight: 600;
}

.asset-meta {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.rule-id {
  color: #94a3b8;
  font-size: 0.75rem;
}

/* 网络标签 */
.network-chip {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.5rem;
  background: #e0e7ff;
  color: #3730a3;
  border-radius: 0.25rem;
  font-size: 0.75rem;
  font-weight: 500;
}

/* 状态标签 */
.status-chip {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-chip.enabled {
  background: #d1fae5;
  color: #065f46;
}

.status-chip.disabled {
  background: #f3f4f6;
  color: #6b7280;
}

/* 规则详情 */
.rule-details {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  margin-bottom: 1rem;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-label {
  color: #64748b;
  font-size: 0.875rem;
}

.detail-value {
  color: #0f172a;
  font-size: 0.875rem;
  font-weight: 500;
}

.detail-value.address {
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Fira Mono', 'Courier New', monospace;
  color: #2563eb;
}

.detail-value.network {
  color: #3730a3;
  font-weight: 600;
}

/* 规则操作 */
.rule-actions {
  display: flex;
  gap: 0.5rem;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr 1fr;
  }

  .form-grid {
    grid-template-columns: 1fr;
  }

  .key-display {
    grid-template-columns: 1fr;
  }

  .rules-grid {
    grid-template-columns: 1fr;
  }

  .form-actions {
    flex-direction: column;
  }

  .rule-actions {
    flex-direction: column;
  }

  .action-btn {
    width: 100%;
  }

  .toast {
    left: 1rem;
    right: 1rem;
    bottom: 1rem;
  }

  .asset-meta {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.25rem;
  }
}
</style>
<template>
  <div class="settings-container">
    <h2>设置</h2>
    <div v-if="error" class="error">{{ error }}</div>

    <!-- API Key Management -->
    <section>
      <h3>API 密钥管理</h3>
      <div v-if="apiKey || secretKey">
        <p>API Key: {{ apiKey }}</p>
        <p>Secret Key: {{ secretKey }}</p>
        <button class="delete" @click="deleteAPIKey">删除 API 密钥</button>
      </div>
      <p v-else>未设置 API 密钥</p>

      <h4>添加新的 API 密钥</h4>
      <div v-if="apiKeySuccess" class="success">{{ apiKeySuccess }}</div>
      <div v-if="apiKeyError" class="error">{{ apiKeyError }}</div>
      <form @submit.prevent="saveAPIKey">
        <div>
          <label>API Key:</label>
          <input v-model="newAPIKey" type="text" required />
        </div>
        <div>
          <label>Secret Key:</label>
          <input v-model="newSecretKey" type="text" required />
        </div>
        <button type="submit">保存 API 密钥</button>
      </form>
    </section>

    <!-- Automatic Withdrawal Settings -->
    <section>
      <h3>自动提币设置</h3>
      <h4>添加提币规则</h4>
      <div v-if="withdrawalSuccess" class="success">{{ withdrawalSuccess }}</div>
      <div v-if="withdrawalError" class="error">{{ withdrawalError }}</div>
      <form @submit.prevent="createWithdrawalRule">
        <div>
          <label>币种:</label>
          <input v-model="newWithdrawal.asset" type="text" placeholder="例如: BTC" required />
        </div>
        <div>
          <label>阈值 (数量):</label>
          <input v-model.number="newWithdrawal.threshold" type="number" step="0.00000001" required />
        </div>
        <div>
          <label>提币地址:</label>
          <input v-model="newWithdrawal.address" type="text" required />
        </div>
        <div>
          <label>提币数量:</label>
          <input v-model.number="newWithdrawal.amount" type="number" step="0.00000001" required />
        </div>
        <button type="submit">添加规则</button>
      </form>

      <h4>现有规则</h4>
      <div v-if="withdrawalRules.length">
        <table>
          <thead>
          <tr>
            <th>ID</th>
            <th>币种</th>
            <th>阈值</th>
            <th>地址</th>
            <th>数量</th>
            <th>状态</th>
            <th>操作</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="rule in withdrawalRules" :key="rule.id">
            <td>{{ rule.id }}</td>
            <td>{{ rule.asset }}</td>
            <td>{{ rule.threshold }}</td>
            <td>{{ rule.address }}</td>
            <td>{{ rule.amount }}</td>
            <td>{{ rule.enabled ? '启用' : '禁用' }}</td>
            <td>
              <button class="delete" @click="deleteWithdrawalRule(rule.id)">删除</button>
            </td>
          </tr>
          </tbody>
        </table>
      </div>
      <p v-else>未设置提币规则</p>
    </section>
  </div>
</template>

<script>
import axios from 'axios';

export default {
  data() {
    return {
      apiKey: '',
      secretKey: '',
      newAPIKey: '',
      newSecretKey: '',
      apiKeySuccess: '',
      apiKeyError: '',
      error: '',
      newWithdrawal: {
        asset: '',
        threshold: 0,
        address: '',
        amount: 0,
      },
      withdrawalRules: [],
      withdrawalSuccess: '',
      withdrawalError: '',
    };
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

    async fetchAPIKey() {
      try {
        const response = await axios.get('/api-key', {
          headers: this.getAuthHeaders(),
        });
        this.apiKey = response.data.apiKey || '';
        this.secretKey = response.data.secretKey || '';
        this.error = '';
      } catch (err) {
        this.error = err.response?.data?.error || '获取 API 密钥失败';
        console.error('fetchAPIKey error:', err);
      }
    },

    async saveAPIKey() {
      try {
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

        this.apiKeySuccess = response.data.message;
        this.apiKeyError = '';
        this.newAPIKey = '';
        this.newSecretKey = '';
        await this.fetchAPIKey();
      } catch (err) {
        console.error('saveAPIKey error:', err.response?.data, err.response?.status);
        this.apiKeyError = err.response?.data?.error || '保存 API 密钥失败';
        this.apiKeySuccess = '';
      }
    },

    async deleteAPIKey() {
      if (!window.confirm('确定要删除 API 密钥吗？')) {
        return;
      }
      try {
        const response = await axios.delete('/api-key/delete', {
          headers: this.getAuthHeaders(),
        });
        this.apiKeySuccess = response.data.message;
        this.apiKeyError = '';
        this.apiKey = '';
        this.secretKey = '';
      } catch (err) {
        this.apiKeyError = err.response?.data?.error || '删除 API 密钥失败';
        this.apiKeySuccess = '';
        console.error('deleteAPIKey error:', err);
      }
    },

    async createWithdrawalRule() {
      if (!this.newWithdrawal.asset || this.newWithdrawal.threshold <= 0 ||
          !this.newWithdrawal.address || this.newWithdrawal.amount <= 0) {
        this.withdrawalError = '请填写所有必需字段，且数量必须大于0';
        return;
      }
      try {
        const response = await axios.post(
            '/withdrawals',
            {
              asset: this.newWithdrawal.asset,
              threshold: this.newWithdrawal.threshold,
              address: this.newWithdrawal.address,
              amount: this.newWithdrawal.amount,
              enabled: true,
            },
            {
              headers: this.getAuthHeaders(),
            }
        );
        this.withdrawalSuccess = response.data.message;
        this.withdrawalError = '';
        this.newWithdrawal = { asset: '', threshold: 0, address: '', amount: 0 };
        await this.fetchWithdrawalRules();
      } catch (err) {
        this.withdrawalError = err.response?.data?.error || '创建提币规则失败';
        this.withdrawalSuccess = '';
        console.error('createWithdrawalRule error:', err);
      }
    },

    async fetchWithdrawalRules() {
      try {
        const response = await axios.get('/withdrawals', {
          headers: this.getAuthHeaders(),
        });
        this.withdrawalRules = response.data.rules || [];
        this.error = '';
      } catch (err) {
        this.error = err.response?.data?.error || '获取提币规则失败';
        console.error('fetchWithdrawalRules error:', err);
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
        this.withdrawalSuccess = response.data.message;
        this.withdrawalError = '';
        await this.fetchWithdrawalRules();
      } catch (err) {
        this.withdrawalError = err.response?.data?.error || '删除提币规则失败';
        this.withdrawalSuccess = '';
        console.error('deleteWithdrawalRule error:', err);
      }
    },
  },
};
</script>

<style scoped>
.settings-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

section {
  margin-bottom: 30px;
}

h2, h3, h4 {
  color: #333;
}

input, button {
  padding: 8px;
  margin: 5px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

button {
  background-color: #4CAF50;
  color: white;
  cursor: pointer;
}

button:hover {
  background-color: #45a049;
}

button.delete {
  background-color: #ff4444;
}

button.delete:hover {
  background-color: #cc0000;
}

table {
  width: 100%;
  border-collapse: collapse;
  margin-top: 10px;
}

th, td {
  padding: 10px;
  text-align: left;
  border: 1px solid #ddd;
}

th {
  background-color: #f2f2f2;
}

.error {
  color: red;
  margin: 10px 0;
}

.success {
  color: green;
  margin: 10px 0;
}

form div {
  margin-bottom: 10px;
}

label {
  display: inline-block;
  width: 120px;
  font-weight: bold;
}
</style>
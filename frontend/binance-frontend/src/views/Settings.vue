<template>
  <div class="settings-container">
    <h2>Settings</h2>
    <div v-if="error" class="error">{{ error }}</div>

    <!-- API Key Management -->
    <section>
      <h3>API Key</h3>
      <div v-if="apiKey || secretKey">
        <p>API Key: {{ apiKey }}</p>
        <p>Secret Key: {{ secretKey }}</p>
        <button class="delete" @click="deleteAPIKey">Delete API Key</button>
      </div>
      <p v-else>No API Key set</p>
      <h4>Add New API Key</h4>
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
        <button type="submit">Save API Key</button>
      </form>
    </section>

    <!-- Automatic Withdrawal Settings -->
    <section>
      <h3>Automatic Withdrawals</h3>
      <h4>Add Withdrawal Rule</h4>
      <div v-if="withdrawalSuccess" class="success">{{ withdrawalSuccess }}</div>
      <div v-if="withdrawalError" class="error">{{ withdrawalError }}</div>
      <form @submit.prevent="createWithdrawalRule">
        <div>
          <label>Coin:</label>
          <input v-model="newWithdrawal.coin" type="text" placeholder="e.g., BTC" required />
        </div>
        <div>
          <label>Threshold (Amount):</label>
          <input v-model.number="newWithdrawal.threshold" type="number" step="0.00000001" required />
        </div>
        <div>
          <label>Withdrawal Address:</label>
          <input v-model="newWithdrawal.address" type="text" required />
        </div>
        <button type="submit">Add Rule</button>
      </form>
      <h4>Existing Rules</h4>
      <div v-if="withdrawalRules.length">
        <table>
          <thead>
          <tr>
            <th>ID</th>
            <th>Coin</th>
            <th>Threshold</th>
            <th>Address</th>
            <th>Action</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="rule in withdrawalRules" :key="rule.id">
            <td>{{ rule.id }}</td>
            <td>{{ rule.coin }}</td>
            <td>{{ rule.threshold }}</td>
            <td>{{ rule.address }}</td>
            <td>
              <button class="delete" @click="deleteWithdrawalRule(rule.id)">Delete</button>
            </td>
          </tr>
          </tbody>
        </table>
      </div>
      <p v-else>No withdrawal rules set</p>
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
        coin: '',
        threshold: 0,
        address: '',
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
    async fetchAPIKey() {
      try {
        const token = localStorage.getItem('token');
        const response = await axios.get('/api-key', {
          headers: { Authorization: `Bearer ${token}` },
        });
        this.apiKey = response.data.apiKey || '';
        this.secretKey = response.data.secretKey || '';
        this.error = '';
      } catch (err) {
        this.error = err.response?.data?.error || 'Failed to fetch API Key';
        console.error('fetchAPIKey error:', err);
      }
    },
    async saveAPIKey() {
      try {
        const token = localStorage.getItem('token');
        console.log('Token:', token);
        console.log('Request body:', {
          apiKey: this.newAPIKey,
          apiSecret: this.newSecretKey, // 改为 apiSecret
        });
        const response = await axios.post(
            '/api-key',
            {
              apiKey: this.newAPIKey,
              apiSecret: this.newSecretKey,
            },
            {
              headers: { Authorization: `Bearer ${token}` },
            }
        );
        console.log('Response:', response.data);
        this.apiKeySuccess = response.data.message;
        this.apiKeyError = '';
        this.newAPIKey = '';
        this.newSecretKey = '';
        await this.fetchAPIKey();
      } catch (err) {
        console.error('saveAPIKey error:', err.response?.data, err.response?.status);
        this.apiKeyError = err.response?.data?.error || 'Failed to save API Key';
        this.apiKeySuccess = '';
      }
    },
    async deleteAPIKey() {
      if (!window.confirm('Are you sure you want to delete your API Key?')) {
        return;
      }
      try {
        const token = localStorage.getItem('token');
        const response = await axios.delete('/api-key/delete', {
          headers: { Authorization: `Bearer ${token}` },
        });
        this.apiKeySuccess = response.data.message;
        this.apiKeyError = '';
        this.apiKey = '';
        this.secretKey = '';
      } catch (err) {
        this.apiKeyError = err.response?.data?.error || 'Failed to delete API Key';
        this.apiKeySuccess = '';
        console.error('deleteAPIKey error:', err);
      }
    },
    async createWithdrawalRule() {
      if (!this.newWithdrawal.coin || this.newWithdrawal.threshold <= 0 || !this.newWithdrawal.address) {
        this.withdrawalError = 'Invalid coin, threshold, or address';
        return;
      }
      try {
        const token = localStorage.getItem('token');
        const response = await axios.post(
            '/withdrawals',
            {
              coin: this.newWithdrawal.coin,
              threshold: this.newWithdrawal.threshold,
              address: this.newWithdrawal.address,
            },
            {
              headers: { Authorization: `Bearer ${token}` },
            }
        );
        this.withdrawalSuccess = response.data.message;
        this.withdrawalError = '';
        this.newWithdrawal = { coin: '', threshold: 0, address: '' };
        await this.fetchWithdrawalRules();
      } catch (err) {
        this.withdrawalError = err.response?.data?.error || 'Failed to create withdrawal rule';
        this.withdrawalSuccess = '';
        console.error('createWithdrawalRule error:', err);
      }
    },
    async fetchWithdrawalRules() {
      try {
        const token = localStorage.getItem('token');
        const response = await axios.get('/withdrawals', {
          headers: { Authorization: `Bearer ${token}` },
        });
        this.withdrawalRules = response.data.rules || [];
        this.error = '';
      } catch (err) {
        this.error = err.response?.data?.error || 'Failed to fetch withdrawal rules';
        console.error('fetchWithdrawalRules error:', err);
      }
    },
    async deleteWithdrawalRule(ruleId) {
      if (!window.confirm(`Are you sure you want to delete withdrawal rule ID ${ruleId}?`)) {
        return;
      }
      try {
        const token = localStorage.getItem('token');
        const response = await axios.delete(`/withdrawals/${ruleId}`, {
          headers: { Authorization: `Bearer ${token}` },
        });
        this.withdrawalSuccess = response.data.message;
        this.withdrawalError = '';
        await this.fetchWithdrawalRules();
      } catch (err) {
        this.withdrawalError = err.response?.data?.error || 'Failed to delete withdrawal rule';
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
</style>
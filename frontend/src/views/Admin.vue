<template>
  <div class="admin-container">
    <!-- 页面头部 -->
    <div class="page-header">
      <h1 class="page-title">用户管理</h1>
      <p class="page-description">管理系统用户账号和权限</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon">
          <span>👥</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.totalUsers }}</div>
          <div class="stat-label">总用户数</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon warning">
          <span>⏳</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.pendingUsers }}</div>
          <div class="stat-label">待审核</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon success">
          <span>✓</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.activeUsers }}</div>
          <div class="stat-label">活跃用户</div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon danger">
          <span>×</span>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats.disabledUsers }}</div>
          <div class="stat-label">已禁用</div>
        </div>
      </div>
    </div>

    <!-- 用户列表 -->
    <div class="content-card">
      <!-- 筛选栏 -->
      <div class="filter-bar">
        <div class="filter-tabs">
          <button
              v-for="filter in filters"
              :key="filter.value"
              @click="currentFilter = filter.value"
              :class="['filter-tab', { active: currentFilter === filter.value }]"
          >
            {{ filter.label }}
            <span class="filter-count">{{ getFilterCount(filter.value) }}</span>
          </button>
        </div>

        <div class="search-box">
          <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索用户名..."
              class="search-input"
          />
          <span class="search-icon">🔍</span>
        </div>
      </div>

      <!-- 用户表格 -->
      <div class="table-container">
        <table class="data-table">
          <thead>
          <tr>
            <th>用户信息</th>
            <th>角色</th>
            <th>状态</th>
            <th>注册时间</th>
            <th>操作</th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="user in filteredAndSearchedUsers" :key="user.id">
            <td>
              <div class="user-info">
                <div class="user-avatar">
                  {{ user.username.charAt(0).toUpperCase() }}
                </div>
                <div class="user-details">
                  <div class="user-name">{{ user.username }}</div>
                  <div class="user-id">ID: {{ user.id }}</div>
                </div>
              </div>
            </td>
            <td>
                <span :class="['role-badge', user.role]">
                  {{ user.role === 'admin' ? '管理员' : '普通用户' }}
                </span>
            </td>
            <td>
                <span :class="['status-badge', user.status]">
                  {{ getStatusText(user.status) }}
                </span>
            </td>
            <td class="text-muted">
              {{ formatDate(user.createdAt) }}
            </td>
            <td>
              <div class="action-buttons">
                <!-- 审核通过 -->
                <button
                    v-if="user.status === 'pending'"
                    @click="approveUser(user.id)"
                    class="btn btn-sm btn-primary"
                    title="审核通过"
                >
                  通过
                </button>

                <!-- 启用/禁用 -->
                <button
                    v-if="user.status !== 'pending' && user.id !== currentUserId"
                    @click="toggleUserStatus(user)"
                    :class="['btn', 'btn-sm', user.status === 'active' ? 'btn-outline' : 'btn-success']"
                    :title="user.status === 'active' ? '禁用账号' : '启用账号'"
                >
                  {{ user.status === 'active' ? '禁用' : '启用' }}
                </button>

                <!-- 设置管理员 -->
                <button
                    v-if="user.status === 'active' && user.id !== currentUserId"
                    @click="user.role === 'admin' ? removeAdmin(user) : setAsAdmin(user)"
                    class="btn btn-sm btn-outline"
                    :title="user.role === 'admin' ? '取消管理员' : '设为管理员'"
                >
                  {{ user.role === 'admin' ? '取消管理' : '设为管理' }}
                </button>
              </div>
            </td>
          </tr>
          </tbody>
        </table>

        <!-- 空状态 -->
        <div v-if="filteredAndSearchedUsers.length === 0" class="empty-state">
          <span class="empty-icon">🔍</span>
          <p>没有找到符合条件的用户</p>
        </div>
      </div>
    </div>

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
  name: 'Admin',
  data() {
    return {
      users: [],
      stats: {
        totalUsers: 0,
        pendingUsers: 0,
        activeUsers: 0,
        disabledUsers: 0,
        adminUsers: 0
      },
      filters: [
        { value: 'all', label: '全部' },
        { value: 'pending', label: '待审核' },
        { value: 'active', label: '活跃' },
        { value: 'disabled', label: '已禁用' },
        { value: 'admin', label: '管理员' }
      ],
      currentFilter: 'all',
      searchQuery: '',
      toastMessage: '',
      toastType: 'success',
      currentUserId: null
    };
  },
  computed: {
    filteredAndSearchedUsers() {
      let filtered = this.users;

      // 状态筛选
      if (this.currentFilter !== 'all') {
        if (this.currentFilter === 'admin') {
          filtered = filtered.filter(user => user.role === 'admin');
        } else {
          filtered = filtered.filter(user => user.status === this.currentFilter);
        }
      }

      // 搜索筛选
      if (this.searchQuery) {
        const query = this.searchQuery.toLowerCase();
        filtered = filtered.filter(user =>
            user.username.toLowerCase().includes(query)
        );
      }

      return filtered;
    }
  },
  mounted() {
    this.getCurrentUserId();
    this.fetchUsers();
    this.fetchStats();
  },
  methods: {
    getCurrentUserId() {
      const token = localStorage.getItem('token');
      if (token) {
        try {
          const payload = JSON.parse(atob(token.split('.')[1]));
          this.currentUserId = payload.user_id;
        } catch (e) {
          console.error('解析token失败:', e);
        }
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

    formatDate(dateString) {
      const date = new Date(dateString);
      return date.toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit'
      });
    },

    getStatusText(status) {
      const statusMap = {
        'pending': '待审核',
        'active': '活跃',
        'disabled': '已禁用'
      };
      return statusMap[status] || status;
    },

    getFilterCount(filterValue) {
      if (filterValue === 'all') return this.users.length;
      if (filterValue === 'admin') return this.users.filter(u => u.role === 'admin').length;
      return this.users.filter(u => u.status === filterValue).length;
    },

    async fetchUsers() {
      try {
        const response = await axios.get('/admin/users', {
          headers: this.getAuthHeaders()
        });
        this.users = response.data.users || [];
      } catch (error) {
        console.error('获取用户列表失败:', error);
        this.showToast(error.response?.data?.error || '获取用户列表失败', 'error');
      }
    },

    async fetchStats() {
      try {
        const response = await axios.get('/admin/users/stats', {
          headers: this.getAuthHeaders()
        });
        this.stats = response.data;
      } catch (error) {
        console.error('获取统计信息失败:', error);
      }
    },

    async approveUser(userId) {
      try {
        await axios.post('/admin/users/approve',
            { userId },
            { headers: this.getAuthHeaders() }
        );

        this.showToast('用户审核通过');
        await this.fetchUsers();
        await this.fetchStats();
      } catch (error) {
        console.error('审核用户失败:', error);
        this.showToast(error.response?.data?.error || '审核用户失败', 'error');
      }
    },

    async toggleUserStatus(user) {
      const newStatus = user.status === 'active' ? 'disabled' : 'active';
      const action = user.status === 'active' ? '禁用' : '启用';

      try {
        await axios.put('/admin/users/status',
            { userId: user.id, status: newStatus },
            { headers: this.getAuthHeaders() }
        );

        this.showToast(`用户已${action}`);
        await this.fetchUsers();
        await this.fetchStats();
      } catch (error) {
        console.error('更新用户状态失败:', error);
        this.showToast(error.response?.data?.error || '更新用户状态失败', 'error');
      }
    },

    async setAsAdmin(user) {
      try {
        await axios.put('/admin/users/role',
            { userId: user.id, role: 'admin' },
            { headers: this.getAuthHeaders() }
        );

        this.showToast('已设为管理员');
        await this.fetchUsers();
        await this.fetchStats();
      } catch (error) {
        console.error('设置管理员失败:', error);
        this.showToast(error.response?.data?.error || '设置管理员失败', 'error');
      }
    },

    async removeAdmin(user) {
      try {
        await axios.put('/admin/users/role',
            { userId: user.id, role: 'user' },
            { headers: this.getAuthHeaders() }
        );

        this.showToast('已取消管理员权限');
        await this.fetchUsers();
        await this.fetchStats();
      } catch (error) {
        console.error('取消管理员权限失败:', error);
        this.showToast(error.response?.data?.error || '取消管理员权限失败', 'error');
      }
    }
  }
};
</script>

<style scoped>
/* 页面容器 */
.admin-container {
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
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  background-color: var(--color-bg-tertiary);
  color: var(--color-text-secondary);
}

.stat-icon.warning {
  background-color: #fef3c7;
  color: var(--color-warning);
}

.stat-icon.success {
  background-color: #d1fae5;
  color: var(--color-success);
}

.stat-icon.danger {
  background-color: #fee2e2;
  color: var(--color-danger);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--color-text-primary);
  line-height: 1;
  margin-bottom: 0.25rem;
}

.stat-label {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

/* 内容卡片 */
.content-card {
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

/* 筛选栏 */
.filter-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--color-border);
  gap: 1rem;
}

.filter-tabs {
  display: flex;
  gap: 0.5rem;
}

.filter-tab {
  padding: 0.5rem 1rem;
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-normal);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-tab:hover {
  background-color: var(--color-bg-tertiary);
}

.filter-tab.active {
  background-color: var(--color-primary);
  color: white;
}

.filter-count {
  background: rgba(0, 0, 0, 0.1);
  padding: 0.125rem 0.5rem;
  border-radius: 10px;
  font-size: 0.75rem;
}

.filter-tab.active .filter-count {
  background: rgba(255, 255, 255, 0.2);
}

/* 搜索框 */
.search-box {
  position: relative;
  width: 240px;
}

.search-input {
  width: 100%;
  padding: 0.5rem 2.5rem 0.5rem 1rem;
  background-color: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  color: var(--color-text-primary);
  transition: all var(--transition-normal);
}

.search-input:focus {
  outline: none;
  border-color: var(--color-primary);
  background-color: var(--color-bg);
}

.search-icon {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-text-tertiary);
}

/* 表格容器 */
.table-container {
  overflow-x: auto;
}

/* 数据表格 */
.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  text-align: left;
  padding: 1rem 1.5rem;
  background-color: var(--color-bg-secondary);
  color: var(--color-text-secondary);
  font-weight: 600;
  font-size: 0.875rem;
  white-space: nowrap;
}

.data-table td {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--color-border);
  vertical-align: middle;
}

.data-table tbody tr:hover {
  background-color: var(--color-bg-secondary);
}

/* 用户信息 */
.user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.user-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background-color: var(--color-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 1rem;
}

.user-details {
  display: flex;
  flex-direction: column;
}

.user-name {
  font-weight: 500;
  color: var(--color-text-primary);
}

.user-id {
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

/* 角色和状态徽章 */
.role-badge,
.status-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.75rem;
  font-weight: 500;
  white-space: nowrap;
}

.role-badge.admin {
  background-color: #dbeafe;
  color: var(--color-primary);
}

.role-badge.user {
  background-color: var(--color-bg-tertiary);
  color: var(--color-text-secondary);
}

.status-badge.pending {
  background-color: #fef3c7;
  color: #92400e;
}

.status-badge.active {
  background-color: #d1fae5;
  color: #065f46;
}

.status-badge.disabled {
  background-color: #fee2e2;
  color: #991b1b;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: 0.5rem;
}

.btn {
  padding: 0.375rem 0.75rem;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: all var(--transition-fast);
  white-space: nowrap;
}

.btn-sm {
  padding: 0.25rem 0.625rem;
  font-size: 0.75rem;
}

.btn-primary {
  background-color: var(--color-primary);
  color: white;
}

.btn-primary:hover {
  background-color: var(--color-primary-hover);
}

.btn-success {
  background-color: var(--color-success);
  color: white;
}

.btn-success:hover {
  background-color: #059669;
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

/* 空状态 */
.empty-state {
  padding: 4rem 2rem;
  text-align: center;
  color: var(--color-text-tertiary);
}

.empty-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: 1rem;
  opacity: 0.5;
}

/* 其他样式 */
.text-muted {
  color: var(--color-text-tertiary);
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

/* 过渡动画 */
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

  .filter-bar {
    flex-direction: column;
    align-items: stretch;
  }

  .filter-tabs {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .search-box {
    width: 100%;
  }

  .data-table {
    font-size: 0.875rem;
  }

  .data-table th,
  .data-table td {
    padding: 0.75rem;
  }

  .action-buttons {
    flex-direction: column;
  }

  .btn {
    width: 100%;
  }
}
</style>
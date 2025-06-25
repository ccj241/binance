<template>
  <div class="admin-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">
        <span class="gradient-text">用户管理中心</span>
      </h1>
      <p class="page-subtitle">管理和监控系统用户</p>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card" v-for="stat in statsConfig" :key="stat.key">
        <div class="stat-icon" :style="{ background: stat.color }">
          <i :class="stat.icon"></i>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ stats[stat.key] }}</div>
          <div class="stat-label">{{ stat.label }}</div>
        </div>
        <div class="stat-bg"></div>
      </div>
    </div>

    <!-- 筛选标签 -->
    <div class="filter-tabs">
      <div class="tabs-wrapper">
        <button
            v-for="filter in filters"
            :key="filter.value"
            @click="currentFilter = filter.value"
            :class="['filter-tab', { active: currentFilter === filter.value }]"
        >
          <span class="tab-icon">{{ filter.icon }}</span>
          <span class="tab-label">{{ filter.label }}</span>
          <span class="tab-count">{{ getFilterCount(filter.value) }}</span>
        </button>
        <div class="tab-indicator" :style="tabIndicatorStyle"></div>
      </div>
    </div>

    <!-- 用户列表 -->
    <div class="users-section">
      <div class="section-header">
        <h2 class="section-title">用户列表</h2>
        <div class="search-box">
          <i class="search-icon">🔍</i>
          <input
              v-model="searchQuery"
              type="text"
              placeholder="搜索用户名..."
              class="search-input"
          >
        </div>
      </div>

      <div class="users-grid">
        <div v-for="user in filteredAndSearchedUsers" :key="user.id" class="user-card">
          <div class="user-header">
            <div class="user-avatar">
              {{ user.username.charAt(0).toUpperCase() }}
            </div>
            <div class="user-info">
              <h3 class="user-name">{{ user.username }}</h3>
              <p class="user-id">ID: {{ user.id }}</p>
            </div>
          </div>

          <div class="user-meta">
            <div class="meta-item">
              <span class="meta-label">角色</span>
              <span :class="['role-chip', user.role]">
                <i :class="user.role === 'admin' ? '👑' : '👤'"></i>
                {{ user.role === 'admin' ? '管理员' : '普通用户' }}
              </span>
            </div>
            <div class="meta-item">
              <span class="meta-label">状态</span>
              <span :class="['status-chip', user.status]">
                <span class="status-dot"></span>
                {{ getStatusText(user.status) }}
              </span>
            </div>
            <div class="meta-item">
              <span class="meta-label">注册时间</span>
              <span class="meta-value">{{ formatDate(user.createdAt) }}</span>
            </div>
          </div>

          <div class="user-actions">
            <!-- 审核通过按钮 -->
            <button
                v-if="user.status === 'pending'"
                @click="approveUser(user.id)"
                class="action-btn approve"
            >
              <i>✓</i> 审核通过
            </button>

            <!-- 启用/禁用按钮 -->
            <button
                v-if="user.status !== 'pending' && user.id !== currentUserId"
                @click="toggleUserStatus(user)"
                :class="['action-btn', user.status === 'active' ? 'disable' : 'enable']"
            >
              <i>{{ user.status === 'active' ? '🚫' : '✓' }}</i>
              {{ user.status === 'active' ? '禁用账号' : '启用账号' }}
            </button>

            <!-- 管理员权限按钮 -->
            <button
                v-if="user.status === 'active' && user.id !== currentUserId"
                @click="user.role === 'admin' ? removeAdmin(user) : setAsAdmin(user)"
                :class="['action-btn', user.role === 'admin' ? 'remove-admin' : 'set-admin']"
            >
              <i>{{ user.role === 'admin' ? '👤' : '👑' }}</i>
              {{ user.role === 'admin' ? '取消管理员' : '设为管理员' }}
            </button>
          </div>
        </div>
      </div>

      <div v-if="filteredAndSearchedUsers.length === 0" class="empty-state">
        <div class="empty-icon">🔍</div>
        <p class="empty-text">没有找到符合条件的用户</p>
      </div>
    </div>

    <!-- 消息提示 -->
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
      statsConfig: [
        { key: 'totalUsers', label: '总用户数', icon: '👥', color: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' },
        { key: 'pendingUsers', label: '待审核', icon: '⏳', color: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)' },
        { key: 'activeUsers', label: '活跃用户', icon: '✅', color: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)' },
        { key: 'disabledUsers', label: '已禁用', icon: '🚫', color: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)' }
      ],
      currentFilter: 'all',
      filters: [
        { value: 'all', label: '全部', icon: '📋' },
        { value: 'pending', label: '待审核', icon: '⏳' },
        { value: 'active', label: '活跃', icon: '✅' },
        { value: 'disabled', label: '已禁用', icon: '🚫' },
        { value: 'admin', label: '管理员', icon: '👑' }
      ],
      searchQuery: '',
      toastMessage: '',
      toastType: 'success',
      currentUserId: null
    };
  },
  computed: {
    filteredUsers() {
      if (this.currentFilter === 'all') {
        return this.users;
      }
      if (this.currentFilter === 'admin') {
        return this.users.filter(user => user.role === 'admin');
      }
      return this.users.filter(user => user.status === this.currentFilter);
    },
    filteredAndSearchedUsers() {
      const filtered = this.filteredUsers;
      if (!this.searchQuery) return filtered;

      return filtered.filter(user =>
          user.username.toLowerCase().includes(this.searchQuery.toLowerCase())
      );
    },
    tabIndicatorStyle() {
      const index = this.filters.findIndex(f => f.value === this.currentFilter);
      return {
        transform: `translateX(${index * 100}%)`
      };
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
        this.showToast(error.response?.data?.error || '获取统计信息失败', 'error');
      }
    },

    async approveUser(userId) {
      try {
        const response = await axios.post('/admin/users/approve',
            { userId },
            { headers: this.getAuthHeaders() }
        );

        this.showToast('用户审核通过 ✅');
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
        const response = await axios.put('/admin/users/status',
            { userId: user.id, status: newStatus },
            { headers: this.getAuthHeaders() }
        );

        this.showToast(`用户已${action} ${user.status === 'active' ? '🚫' : '✅'}`);
        await this.fetchUsers();
        await this.fetchStats();
      } catch (error) {
        console.error('更新用户状态失败:', error);
        this.showToast(error.response?.data?.error || '更新用户状态失败', 'error');
      }
    },

    async setAsAdmin(user) {
      try {
        const response = await axios.put('/admin/users/role',
            { userId: user.id, role: 'admin' },
            { headers: this.getAuthHeaders() }
        );

        this.showToast('已设为管理员 👑');
        await this.fetchUsers();
        await this.fetchStats();
      } catch (error) {
        console.error('设置管理员失败:', error);
        this.showToast(error.response?.data?.error || '设置管理员失败', 'error');
      }
    },

    async removeAdmin(user) {
      try {
        const response = await axios.put('/admin/users/role',
            { userId: user.id, role: 'user' },
            { headers: this.getAuthHeaders() }
        );

        this.showToast('已取消管理员权限 👤');
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
/* 全局样式 */
.admin-container {
  min-height: 100vh;
  background: #0f0f0f;
  color: #ffffff;
  padding: 2rem;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
}

/* 页面标题 */
.page-header {
  text-align: center;
  margin-bottom: 3rem;
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
  color: #666;
  font-size: 1.1rem;
}

/* 统计卡片 */
.stats-grid {
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
  position: relative;
  overflow: hidden;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-5px);
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
}

.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  margin-bottom: 1rem;
}

.stat-content {
  position: relative;
  z-index: 1;
}

.stat-value {
  font-size: 2.5rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.stat-label {
  color: #999;
  font-size: 0.9rem;
}

.stat-bg {
  position: absolute;
  top: -50%;
  right: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255,255,255,0.03) 0%, transparent 70%);
  transform: rotate(45deg);
}

/* 筛选标签 */
.filter-tabs {
  margin-bottom: 2rem;
}

.tabs-wrapper {
  display: flex;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 16px;
  padding: 0.5rem;
  position: relative;
  gap: 0.5rem;
}

.filter-tab {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1rem;
  background: none;
  border: none;
  color: #999;
  cursor: pointer;
  border-radius: 12px;
  transition: all 0.3s ease;
  position: relative;
  z-index: 1;
}

.filter-tab:hover {
  color: #fff;
}

.filter-tab.active {
  color: #fff;
}

.tab-icon {
  font-size: 1.2rem;
}

.tab-count {
  background: rgba(255, 255, 255, 0.1);
  padding: 0.2rem 0.6rem;
  border-radius: 20px;
  font-size: 0.8rem;
}

.tab-indicator {
  position: absolute;
  top: 0.5rem;
  left: 0.5rem;
  width: calc(20% - 0.4rem);
  height: calc(100% - 1rem);
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  transition: transform 0.3s ease;
  z-index: 0;
}

/* 用户列表区域 */
.users-section {
  background: rgba(255, 255, 255, 0.03);
  border-radius: 24px;
  padding: 2rem;
  border: 1px solid rgba(255, 255, 255, 0.1);
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
}

/* 搜索框 */
.search-box {
  position: relative;
  width: 300px;
}

.search-icon {
  position: absolute;
  left: 1rem;
  top: 50%;
  transform: translateY(-50%);
  font-size: 1.2rem;
}

.search-input {
  width: 100%;
  padding: 0.8rem 1rem 0.8rem 3rem;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  color: #fff;
  font-size: 0.9rem;
  transition: all 0.3s ease;
}

.search-input:focus {
  outline: none;
  background: rgba(255, 255, 255, 0.08);
  border-color: #667eea;
}

.search-input::placeholder {
  color: #666;
}

/* 用户卡片网格 */
.users-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
  gap: 1.5rem;
}

.user-card {
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 1.5rem;
  transition: all 0.3s ease;
}

.user-card:hover {
  background: rgba(255, 255, 255, 0.08);
  transform: translateY(-2px);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.user-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.user-avatar {
  width: 50px;
  height: 50px;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: 600;
}

.user-info {
  flex: 1;
}

.user-name {
  font-size: 1.2rem;
  font-weight: 600;
  margin-bottom: 0.25rem;
}

.user-id {
  color: #666;
  font-size: 0.9rem;
}

/* 用户元信息 */
.user-meta {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.meta-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.meta-label {
  color: #666;
  font-size: 0.9rem;
}

.meta-value {
  color: #ccc;
  font-size: 0.9rem;
}

/* 角色和状态标签 */
.role-chip, .status-chip {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.4rem 0.8rem;
  border-radius: 20px;
  font-size: 0.85rem;
  font-weight: 500;
}

.role-chip.admin {
  background: rgba(111, 66, 193, 0.2);
  color: #a78bfa;
  border: 1px solid rgba(111, 66, 193, 0.3);
}

.role-chip.user {
  background: rgba(108, 117, 125, 0.2);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
}

.status-chip {
  position: relative;
  padding-left: 1.5rem;
}

.status-dot {
  position: absolute;
  left: 0.5rem;
  top: 50%;
  transform: translateY(-50%);
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.status-chip.pending {
  background: rgba(255, 193, 7, 0.2);
  color: #fbbf24;
  border: 1px solid rgba(255, 193, 7, 0.3);
}

.status-chip.pending .status-dot {
  background: #fbbf24;
  animation: pulse 2s infinite;
}

.status-chip.active {
  background: rgba(40, 167, 69, 0.2);
  color: #22c55e;
  border: 1px solid rgba(40, 167, 69, 0.3);
}

.status-chip.active .status-dot {
  background: #22c55e;
}

.status-chip.disabled {
  background: rgba(220, 53, 69, 0.2);
  color: #ef4444;
  border: 1px solid rgba(220, 53, 69, 0.3);
}

.status-chip.disabled .status-dot {
  background: #ef4444;
}

/* 操作按钮 */
.user-actions {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.action-btn {
  flex: 1;
  min-width: 120px;
  padding: 0.8rem 1rem;
  border: none;
  border-radius: 10px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.action-btn i {
  font-style: normal;
}

.action-btn.approve {
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%);
  color: white;
}

.action-btn.approve:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 20px rgba(34, 197, 94, 0.4);
}

.action-btn.enable {
  background: rgba(34, 197, 94, 0.1);
  color: #22c55e;
  border: 1px solid rgba(34, 197, 94, 0.3);
}

.action-btn.enable:hover {
  background: rgba(34, 197, 94, 0.2);
}

.action-btn.disable {
  background: rgba(239, 68, 68, 0.1);
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.action-btn.disable:hover {
  background: rgba(239, 68, 68, 0.2);
}

.action-btn.set-admin {
  background: rgba(111, 66, 193, 0.1);
  color: #a78bfa;
  border: 1px solid rgba(111, 66, 193, 0.3);
}

.action-btn.set-admin:hover {
  background: rgba(111, 66, 193, 0.2);
}

.action-btn.remove-admin {
  background: rgba(108, 117, 125, 0.1);
  color: #94a3b8;
  border: 1px solid rgba(108, 117, 125, 0.3);
}

.action-btn.remove-admin:hover {
  background: rgba(108, 117, 125, 0.2);
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 4rem 2rem;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
  opacity: 0.3;
}

.empty-text {
  color: #666;
  font-size: 1.1rem;
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
  z-index: 1000;
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
@keyframes pulse {
  0% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
  100% {
    opacity: 1;
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
  .admin-container {
    padding: 1rem;
  }

  .page-title {
    font-size: 2rem;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 1rem;
  }

  .stat-card {
    padding: 1.5rem;
  }

  .tabs-wrapper {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .filter-tab {
    white-space: nowrap;
    min-width: 120px;
  }

  .section-header {
    flex-direction: column;
    gap: 1rem;
    align-items: stretch;
  }

  .search-box {
    width: 100%;
  }

  .users-grid {
    grid-template-columns: 1fr;
  }

  .user-actions {
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
}

/* 暗色模式优化 */
@media (prefers-color-scheme: light) {
  .admin-container {
    background: #f9fafb;
    color: #111827;
  }

  .stat-card,
  .filter-tabs .tabs-wrapper,
  .users-section,
  .user-card,
  .search-input {
    background: rgba(0, 0, 0, 0.03);
    border-color: rgba(0, 0, 0, 0.1);
  }

  .stat-card:hover,
  .user-card:hover {
    background: rgba(0, 0, 0, 0.05);
  }

  .page-subtitle,
  .stat-label,
  .meta-label,
  .user-id {
    color: #6b7280;
  }

  .filter-tab {
    color: #6b7280;
  }

  .filter-tab:hover,
  .filter-tab.active {
    color: #111827;
  }

  .search-input {
    color: #111827;
  }

  .search-input::placeholder {
    color: #9ca3af;
  }

  .toast {
    background: rgba(0, 0, 0, 0.9);
    color: white;
  }
}
</style>
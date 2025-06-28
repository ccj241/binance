//test
import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import axios from 'axios';
// 设置axios默认配置
//axios.defaults.baseURL = 'http://localhost:8081';
axios.defaults.baseURL = '';
axios.defaults.timeout = 10000;
axios.defaults.headers.common['Content-Type'] = 'application/json';
// 添加请求拦截器
// 添加请求拦截器
axios.interceptors.request.use(
    config => {
        const token = localStorage.getItem('token');

        // 验证 token 有效性
        if (token && token !== 'undefined' && token !== 'null' && token !== '') {
            // 额外验证 token 格式
            const parts = token.split('.');
            if (parts.length === 3) {
                config.headers.Authorization = `Bearer ${token}`;
            } else {
                console.error('检测到无效的 token 格式，清理中...');
                localStorage.removeItem('token');
            }
        }

        // 确保所有请求都有正确的 Content-Type
        if (!config.headers['Content-Type']) {
            config.headers['Content-Type'] = 'application/json';
        }

        // 开发环境下打印请求信息
        if (process.env.NODE_ENV === 'development') {
            console.log(`📤 ${config.method?.toUpperCase()} ${config.url}`);
        }

        return config;
    },
    error => {
        console.error('❌ 请求错误:', error);
        return Promise.reject(error);
    }
);
// 添加响应拦截器
axios.interceptors.response.use(
    response => {
// 开发环境下打印响应信息
        if (process.env.NODE_ENV === 'development') {
            console.log('📥 ${response.config.url} - ${response.status}');
        }
        return response;
    },
    error => {
        // 统一错误处理
        if (error.response) {
            const { status, data } = error.response;

            switch (status) {
                case 401:
                    // Token过期或无效
                    console.error('🔐 认证失败');
                    localStorage.removeItem('token');

                    // 避免重复跳转
                    if (router.currentRoute.value.path !== '/login') {
                        router.push('/login');
                    }
                    break;

                case 403:
                    console.error('🚫 没有权限访问该资源');
                    break;

                case 404:
                    console.error('🔍 请求的资源不存在');
                    break;

                case 422:
                    console.error('⚠️ 请求参数验证失败');
                    break;

                case 500:
                    console.error('💥 服务器内部错误');
                    break;

                default:
                    console.error(`❌ 请求失败: ${data?.error || data?.message || '未知错误'}`);
            }
        } else if (error.request) {
            console.error('🌐 网络错误，请检查网络连接');
        } else {
            console.error('⚠️ 请求配置错误:', error.message);
        }

        return Promise.reject(error);
    }
);
// 创建Vue应用实例
const app = createApp(App);
// 全局错误处理
app.config.errorHandler = (err, instance, info) => {
    console.error('Vue Error:', err, info);
};
// 全局属性配置
app.config.globalProperties.$axios = axios;
// 注册全局方法
app.config.globalProperties.$formatNumber = (num) => {
    if (!num) return '0';
    return new Intl.NumberFormat('zh-CN').format(num);
};
app.config.globalProperties.$formatCurrency = (amount, currency = 'USD') => {
    return new Intl.NumberFormat('zh-CN', {
        style: 'currency',
        currency: currency,
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
    }).format(amount || 0);
};
app.config.globalProperties.$formatDate = (dateString) => {
    if (!dateString) return '-';
    const date = new Date(dateString);
    const now = new Date();
    const diff = now - date;
// 时间差转换
    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (seconds < 60) return '刚刚';
    if (minutes < 60) return `${minutes}分钟前`;
    if (hours < 24) return `${hours}小时前`;
    if (days < 7) return `${days}天前`;

// 超过7天显示具体日期
    return date.toLocaleDateString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
};
// 使用路由
app.use(router);
// 挂载应用
app.mount('#app');
// 开发环境提示
if (process.env.NODE_ENV === 'development') {
    console.log('🚀 应用已启动 - 开发模式');
} else {
    console.log('🚀 应用已启动 - 生产模式');
}
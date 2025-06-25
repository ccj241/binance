import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import axios from 'axios';

// 设置axios默认配置
axios.defaults.baseURL = 'http://localhost:8081';
axios.defaults.timeout = 10000;
axios.defaults.headers.common['Content-Type'] = 'application/json';

// 添加请求拦截器
axios.interceptors.request.use(
    config => {
        const token = localStorage.getItem('token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        // 确保所有请求都有正确的 Content-Type
        if (!config.headers['Content-Type']) {
            config.headers['Content-Type'] = 'application/json';
        }
        console.log('发送请求:', config.method, config.url);
        return config;
    },
    error => {
        console.error('请求错误:', error);
        return Promise.reject(error);
    }
);

// 添加响应拦截器
axios.interceptors.response.use(
    response => {
        console.log('收到响应:', response.config.url, response.status);
        return response;
    },
    error => {
        console.error('响应错误:', error.response?.status, error.response?.data);

        if (error.response) {
            switch (error.response.status) {
                case 401:
                    // Token过期或无效
                    console.log('认证失败，跳转到登录页');
                    localStorage.removeItem('token');
                    router.push('/login');
                    break;
                case 403:
                    console.error('没有权限访问该资源');
                    break;
                case 404:
                    console.error('请求的资源不存在');
                    break;
                case 500:
                    console.error('服务器内部错误');
                    break;
                default:
                    console.error('请求失败:', error.response.data?.error || '未知错误');
            }
        } else if (error.request) {
            console.error('网络错误，请检查网络连接');
        }

        return Promise.reject(error);
    }
);

// 将axios挂载到全局，方便组件中使用
const app = createApp(App);
app.config.globalProperties.$axios = axios;
app.use(router);
app.mount('#app');
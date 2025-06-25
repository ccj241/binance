import { createApp } from 'vue';
import App from './App.vue';
import router from './router';
import axios from 'axios';

// 设置axios默认配置
axios.defaults.baseURL = 'http://localhost:8081';
axios.defaults.timeout = 10000;

// 添加请求拦截器
axios.interceptors.request.use(
    config => {
        const token = localStorage.getItem('token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
    },
    error => {
        return Promise.reject(error);
    }
);

// 添加响应拦截器
axios.interceptors.response.use(
    response => {
        return response;
    },
    error => {
        if (error.response && error.response.status === 401) {
            // Token过期或无效，清除token并跳转到登录页
            localStorage.removeItem('token');
            router.push('/login');
        }
        return Promise.reject(error);
    }
);

const app = createApp(App);
app.use(router);
app.mount('#app');